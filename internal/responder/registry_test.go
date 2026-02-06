package responder

import (
	"sync"
	"testing"
)

// TestRegistry_Register_RED tests service registration.
//
// TDD Phase: RED - These tests will FAIL until we implement Registry
//
// R006 Decision: Use sync.RWMutex for concurrent access
// T013: Implement Registry with sync.RWMutex
func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	err := registry.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Verify service was registered
	got, exists := registry.Get(service.InstanceName)
	if !exists {
		t.Fatal("Get() returned exists=false after Register()")
	}

	if got.InstanceName != service.InstanceName {
		t.Errorf("Get().InstanceName = %q, want %q", got.InstanceName, service.InstanceName)
	}
}

// TestRegistry_Register_Duplicate_RED tests duplicate registration handling.
//
// TDD Phase: RED
//
// T013: Registry must detect duplicate registrations
func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// First registration should succeed
	err := registry.Register(service)
	if err != nil {
		t.Fatalf("First Register() error = %v, want nil", err)
	}

	// Second registration with same InstanceName should fail
	err = registry.Register(service)
	if err == nil {
		t.Error("Duplicate Register() error = nil, want error")
	}
}

// TestRegistry_Get_NotFound_RED tests retrieving non-existent service.
//
// TDD Phase: RED
//
// T013: Registry must return exists=false for non-existent services
func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, exists := registry.Get("non-existent")
	if exists {
		t.Error("Get(non-existent) exists=true, want false")
	}
}

// TestRegistry_Remove_RED tests service removal.
//
// TDD Phase: RED
//
// T013: Registry must support removing services
func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register service
	err := registry.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Remove service
	err = registry.Remove(service.InstanceName)
	if err != nil {
		t.Fatalf("Remove() error = %v, want nil", err)
	}

	// Verify service was removed
	_, exists := registry.Get(service.InstanceName)
	if exists {
		t.Error("Get() exists=true after Remove(), want false")
	}
}

// TestRegistry_Remove_NotFound_RED tests removing non-existent service.
//
// TDD Phase: RED
//
// T013: Registry should return error when removing non-existent service
func TestRegistry_Remove_NotFound(t *testing.T) {
	registry := NewRegistry()

	err := registry.Remove("non-existent")
	if err == nil {
		t.Error("Remove(non-existent) error = nil, want error")
	}
}

// TestRegistry_ConcurrentAccess_RED tests concurrent registration and retrieval.
//
// TDD Phase: RED
//
// R006 Decision: sync.RWMutex for concurrent safety
// T013: Registry must be thread-safe
//
// This test will be run with `go test -race` to detect data races.
func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			service := &Service{
				InstanceName: formatInstanceName("Service", id),
				ServiceType:  "_http._tcp.local",
				Port:         uint16(8080 + id),
			}

			err := registry.Register(service)
			if err != nil {
				t.Errorf("Concurrent Register() error = %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all services were registered
	// Concurrent reads while other goroutines might still be writing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			instanceName := formatInstanceName("Service", id)
			_, exists := registry.Get(instanceName)
			if !exists {
				t.Errorf("Get(%q) exists=false, want true", instanceName)
			}
		}(i)
	}

	wg.Wait()
}

// TestRegistry_ConcurrentReadWrite_RED tests concurrent reads and writes.
//
// TDD Phase: RED
//
// R006 Decision: RWMutex allows multiple concurrent readers
// T013: Registry must support concurrent reads with single writer
func TestRegistry_ConcurrentReadWrite(_ *testing.T) {
	registry := NewRegistry()

	// Pre-populate registry
	for i := 0; i < 10; i++ {
		service := &Service{
			InstanceName: formatInstanceName("Service", i),
			ServiceType:  "_http._tcp.local",
			Port:         uint16(8080 + i),
		}
		_ = registry.Register(service)
	}

	var wg sync.WaitGroup

	// Start many concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)   // nosemgrep: beacon-waitgroup-missing-done
		go func() { // defer wg.Done() is present on next line (first statement in goroutine)
			defer wg.Done()
			for j := 0; j < 100; j++ {
				instanceName := formatInstanceName("Service", j%10)
				registry.Get(instanceName)
			}
		}()
	}

	// Start a few concurrent writers
	for i := 10; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service := &Service{
				InstanceName: formatInstanceName("Service", id),
				ServiceType:  "_http._tcp.local",
				Port:         uint16(8080 + id),
			}
			_ = registry.Register(service)
		}(i)
	}

	wg.Wait()
}

// TestRegistry_ListServiceTypes tests retrieving unique service types.
//
// TDD Phase: RED - This test will FAIL until we implement ListServiceTypes()
//
// RFC 6763 §9: Service enumeration requires listing unique service types
// FR-027: System MUST respond to "_services._dns-sd._udp.local" with unique service types
// T107: Implement service type enumeration support
func TestRegistry_ListServiceTypes(t *testing.T) {
	registry := NewRegistry()

	// Register 3 services with DIFFERENT types
	services := []*Service{
		{InstanceName: "Web1", ServiceType: "_http._tcp.local", Port: 8080},
		{InstanceName: "SSH1", ServiceType: "_ssh._tcp.local", Port: 22},
		{InstanceName: "FTP1", ServiceType: "_ftp._tcp.local", Port: 21},
	}

	for _, svc := range services {
		if err := registry.Register(svc); err != nil {
			t.Fatalf("Register(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Get unique service types
	types := registry.ListServiceTypes()

	// Should have exactly 3 unique types
	if len(types) != 3 {
		t.Errorf("ListServiceTypes() count = %d, want 3", len(types))
	}

	// Verify all expected types are present
	expectedTypes := map[string]bool{
		"_http._tcp.local": false,
		"_ssh._tcp.local":  false,
		"_ftp._tcp.local":  false,
	}

	for _, serviceType := range types {
		if _, expected := expectedTypes[serviceType]; expected {
			expectedTypes[serviceType] = true
		} else {
			t.Errorf("ListServiceTypes() returned unexpected type %q", serviceType)
		}
	}

	// Check all expected types were found
	for serviceType, found := range expectedTypes {
		if !found {
			t.Errorf("ListServiceTypes() missing expected type %q", serviceType)
		}
	}
}

// TestRegistry_ListServiceTypes_Duplicates tests deduplication of service types.
//
// TDD Phase: RED
//
// RFC 6763 §9: Service enumeration lists unique types, not instances
// T107: Verify duplicate service types appear only once
func TestRegistry_ListServiceTypes_Duplicates(t *testing.T) {
	registry := NewRegistry()

	// Register 3 services with SAME type
	services := []*Service{
		{InstanceName: "Web1", ServiceType: "_http._tcp.local", Port: 8080},
		{InstanceName: "Web2", ServiceType: "_http._tcp.local", Port: 8081},
		{InstanceName: "Web3", ServiceType: "_http._tcp.local", Port: 8082},
	}

	for _, svc := range services {
		if err := registry.Register(svc); err != nil {
			t.Fatalf("Register(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Get unique service types
	types := registry.ListServiceTypes()

	// Should have exactly 1 unique type (not 3)
	if len(types) != 1 {
		t.Errorf("ListServiceTypes() count = %d, want 1 (unique types only)", len(types))
	}

	if len(types) > 0 && types[0] != "_http._tcp.local" {
		t.Errorf("ListServiceTypes()[0] = %q, want %q", types[0], "_http._tcp.local")
	}
}

// TestRegistry_ListServiceTypes_Empty tests empty registry behavior.
//
// TDD Phase: RED
//
// T107: Empty registry should return empty slice
func TestRegistry_ListServiceTypes_Empty(t *testing.T) {
	registry := NewRegistry()

	// Get unique service types from empty registry
	types := registry.ListServiceTypes()

	// Should return empty slice (not nil)
	if types == nil {
		t.Error("ListServiceTypes() = nil, want empty slice")
	}

	if len(types) != 0 {
		t.Errorf("ListServiceTypes() count = %d, want 0 (empty registry)", len(types))
	}
}

// Note: Service type is now implemented in registry.go (T013 GREEN phase)

// formatInstanceName creates a test instance name.
func formatInstanceName(prefix string, id int) string {
	return prefix + "-" + string(rune('0'+id))
}
