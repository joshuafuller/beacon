// Package responder manages mDNS service registration and response logic.
//
// This package implements the internal service registry for tracking registered
// mDNS services and building responses to queries per RFC 6762 §6.
package responder

import (
	"fmt"
	"sync"
)

// Registry manages registered mDNS services with thread-safe concurrent access.
//
// RFC 6762 §6: Responding to Queries
//
// The registry maintains a thread-safe map of registered services, enabling:
//  1. Concurrent service registration/unregistration (write operations)
//  2. Concurrent service lookups during query response (read operations)
//  3. Service type enumeration for RFC 6763 §9 service discovery
//
// Thread-safety is critical because:
//  - Query handler goroutine continuously reads registry for response matching
//  - Application goroutines concurrently register/unregister services
//  - Multiple services can be registered/queried simultaneously
//
// Design Decision:
//   - R006: sync.RWMutex enables multiple concurrent readers (Get, List)
//     while serializing writers (Register, Remove)
//   - Map key is InstanceName (unique per service instance)
//   - Duplicate detection prevents name collisions
//
// Functional Requirements:
//   - FR-203: Multi-service support with concurrent access
//   - FR-204: Fast service lookup for query response
//   - FR-027: Service type enumeration for RFC 6763 §9
//   - T013: Thread-safe registry implementation
//
// Example:
//
//	registry := NewRegistry()
//
//	// Register service (write lock)
//	service := &Service{InstanceName: "My Printer", ServiceType: "_ipp._tcp.local", Port: 631}
//	if err := registry.Register(service); err != nil {
//	    return err
//	}
//
//	// Query service (read lock - concurrent with other reads)
//	if svc, found := registry.Get("My Printer"); found {
//	    fmt.Printf("Found service on port %d\n", svc.Port)
//	}
//
//	// List all service types (read lock)
//	types := registry.ListServiceTypes() // ["_ipp._tcp.local", "_http._tcp.local"]
type Registry struct {
	mu       sync.RWMutex
	services map[string]*Service
}

// NewRegistry creates a new service registry.
//
// Returns:
//   - *Registry: An empty registry ready for service registration
//
// T013: Initialize Registry with map and RWMutex
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]*Service),
	}
}

// Register adds a service to the registry.
//
// Parameters:
//   - service: The service to register
//
// Returns:
//   - error: Error if service with same InstanceName already exists
//
// Thread-safe: Uses write lock (RWMutex.Lock)
//
// T013: Implement Register with duplicate detection
func (r *Registry) Register(service *Service) error {
	if service == nil {
		return fmt.Errorf("cannot register nil service")
	}

	if service.InstanceName == "" {
		return fmt.Errorf("service InstanceName cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate
	if _, exists := r.services[service.InstanceName]; exists {
		return fmt.Errorf("service with InstanceName %q already registered", service.InstanceName)
	}

	r.services[service.InstanceName] = service
	return nil
}

// Get retrieves a service by instance name.
//
// Parameters:
//   - instanceName: The service instance name to look up
//
// Returns:
//   - *Service: The service if found, nil otherwise
//   - bool: true if service exists, false otherwise
//
// Thread-safe: Uses read lock (RWMutex.RLock) - allows concurrent reads
//
// T013: Implement Get with RLock for concurrent reads
func (r *Registry) Get(instanceName string) (*Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[instanceName]
	return service, exists
}

// Remove removes a service from the registry.
//
// Parameters:
//   - instanceName: The service instance name to remove
//
// Returns:
//   - error: Error if service not found
//
// Thread-safe: Uses write lock (RWMutex.Lock)
//
// T013: Implement Remove with error on not found
func (r *Registry) Remove(instanceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[instanceName]; !exists {
		return fmt.Errorf("service with InstanceName %q not found", instanceName)
	}

	delete(r.services, instanceName)
	return nil
}

// List returns all registered service instance names.
//
// Returns:
//   - []string: List of instance names
//
// Thread-safe: Uses read lock (RWMutex.RLock)
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}

// ListServiceTypes returns all unique registered service types.
//
// This method supports RFC 6763 §9 service enumeration by listing unique service types
// (e.g., "_http._tcp.local", "_ssh._tcp.local") rather than service instances.
//
// Returns:
//   - []string: List of unique service types (no duplicates)
//
// Thread-safe: Uses read lock (RWMutex.RLock) - allows concurrent reads
//
// RFC 6763 §9: "_services._dns-sd._udp.local" PTR query returns unique service types
// FR-027: System MUST respond with list of all registered service types
// T107: Implement service type enumeration support
func (r *Registry) ListServiceTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use map to deduplicate service types
	// Multiple services can have the same type (e.g., 3 web servers all "_http._tcp.local")
	typeMap := make(map[string]bool)
	for _, service := range r.services {
		typeMap[service.ServiceType] = true
	}

	// Convert map keys to slice
	types := make([]string, 0, len(typeMap))
	for serviceType := range typeMap {
		types = append(types, serviceType)
	}

	return types
}

// Service represents a registered mDNS service.
//
// This is the minimal implementation for T013 (Registry tests).
// Full implementation will be in responder/service.go (T031).
//
// T031: This will be moved to service.go with full validation
type Service struct {
	InstanceName string
	ServiceType  string
	Port         int
	TXT          map[string]string
}
