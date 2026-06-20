package responder

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/records"
	"github.com/joshuafuller/beacon/internal/responder"
	"github.com/joshuafuller/beacon/internal/security"
	"github.com/joshuafuller/beacon/internal/state"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Responder manages mDNS service registration and response per RFC 6762.
//
// Interface-Specific Addressing (RFC 6762 §15):
// The responder automatically detects which network interface received each query
// and responds with ONLY the IP address valid on that interface. This ensures
// clients can connect to the correct IP when the host has multiple network interfaces.
//
// Example: Host with WiFi (10.0.0.50) and Ethernet (192.168.1.100):
//   - Query on WiFi → Response contains 10.0.0.50
//   - Query on Ethernet → Response contains 192.168.1.100
//
// Graceful Degradation:
// If interface information is unavailable (e.g., on Windows or older kernels),
// the responder falls back to advertising the default interface IP.
//
// The implementation is split across several files in this package:
//   - responder.go      lifecycle scaffolding (struct, New, Close) and IP/dedup helpers
//   - lifecycle.go      service management (Register, Unregister, Get, Update)
//   - query_handler.go  incoming-query processing (RFC 6762 §6)
//   - testhooks.go      test-only observation/injection hooks (see file header)
//
// T035: Responder struct
// T080: Added query handler goroutine support
// T082: Added interface-specific addressing documentation
type Responder struct {
	ctx              context.Context
	transport        transport.Transport
	registry         *responder.Registry
	hostname         string
	queryHandlerWg   sync.WaitGroup             // Synchronize query handler goroutine shutdown
	responseBuilder  *responder.ResponseBuilder // RFC 6762 §6 response construction
	recordSet        *records.RecordSet         // Per-record rate limiting tracker
	rateLimiter      *security.RateLimiter      // Per-source-IP rate limiting (FR-026)
	queryHandlerDone chan struct{}              // Signal query handler shutdown

	// Test-only state. These fields exist solely to support black-box contract
	// tests (see testhooks.go); they are not part of the responder's runtime
	// behavior. Production code paths never read them except where guarded.
	injectConflict       bool              // Inject conflict during probing
	lastMachine          *state.Machine    // Last state machine used for registration
	onProbeCallback      func()            // Callback for probe events
	onAnnounceCallback   func()            // Callback for announce events
	lastAnnouncedRecords []*ResourceRecord // Last record set announced
}

// New creates a new mDNS responder.
//
// T036: Responder.New() implementation
// T080: Start query handler goroutine
func New(ctx context.Context, opts ...Option) (*Responder, error) {
	// Get system hostname if not provided
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	hostname = hostname + ".local"

	// Create transport
	t, err := transport.NewUDPv4Transport()
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	r := &Responder{
		ctx:              ctx,
		transport:        t,
		registry:         responder.NewRegistry(),
		hostname:         hostname,
		responseBuilder:  responder.NewResponseBuilder(),
		recordSet:        records.NewRecordSet(),
		rateLimiter:      security.NewRateLimiter(100, 60*time.Second, 10000),
		queryHandlerDone: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Start query handler goroutine (T080)
	r.queryHandlerWg.Add(1)
	go r.runQueryHandler()

	return r, nil
}

// Close closes the responder and unregisters all services per FR-015.
//
// Process:
//  1. Stop query handler goroutine
//  2. Unregister all services (sends goodbye packets)
//  3. Close transport
//
// Returns:
//   - error: transport close error
//
// T043: Implement Close()
// T080: Stop query handler
func (r *Responder) Close() error {
	// Stop query handler goroutine (T080)
	close(r.queryHandlerDone)

	// Unregister all services (sends goodbye packets)
	services := r.registry.List()
	for _, instanceName := range services {
		// Ignore errors - service may have been manually unregistered
		_ = r.Unregister(instanceName)
	}

	// Close transport - this also unblocks the query handler goroutine's
	// Receive() call so it can observe the queryHandlerDone signal and exit.
	var closeErr error
	if r.transport != nil {
		closeErr = r.transport.Close()
	}

	// Wait for query handler goroutine to finish after transport is closed.
	// The goroutine will exit once Receive() returns an error from the closed
	// transport and it checks queryHandlerDone.
	r.queryHandlerWg.Wait()

	return closeErr
}

// ResourceRecord is a type alias for records.ResourceRecord.
//
// This alias allows contract tests to reference ResourceRecord without importing
// the internal records package directly, maintaining clean architecture boundaries.
//
// The underlying type contains DNS resource record fields:
//   - Name: Domain name (e.g., "myservice._http._tcp.local")
//   - Type: Record type (A, PTR, SRV, TXT per RFC 1035)
//   - Class: Record class (IN for Internet)
//   - TTL: Time-to-live in seconds
//   - Data: Record-specific data (IP address, target name, etc.)
//   - CacheFlush: Cache-flush bit per RFC 6762 §10.2
//
// US2 GREEN: Contract test support for validating resource records
type ResourceRecord = records.ResourceRecord

// buildServiceInfo assembles a records.ServiceInfo from individual service
// fields. Shared by Register, Unregister, and UpdateService so the record-set
// inputs are constructed in exactly one place.
func buildServiceInfo(instanceName, serviceType, hostname string, port uint16, ipv4 []byte, txt map[string]string) *records.ServiceInfo {
	return &records.ServiceInfo{
		InstanceName: instanceName,
		ServiceType:  serviceType,
		Hostname:     hostname,
		Port:         port,
		IPv4Address:  ipv4,
		TXTRecords:   txt,
	}
}

// toInternalService converts a public Service to the internal registry type.
func toInternalService(s *Service) *responder.Service {
	return &responder.Service{
		InstanceName: s.InstanceName,
		ServiceType:  s.ServiceType,
		Port:         s.Port,
		TXT:          s.TXTRecords,
	}
}

// fromInternalService converts an internal registry Service to the public type.
func fromInternalService(s *responder.Service) *Service {
	return &Service{
		InstanceName: s.InstanceName,
		ServiceType:  s.ServiceType,
		Port:         s.Port,
		TXTRecords:   s.TXT,
	}
}

// getLocalIPv4 gets the first non-loopback IPv4 address from any interface.
//
// DEPRECATED for query response building: Use getIPv4ForInterface(interfaceIndex) instead
// to comply with RFC 6762 §15 (interface-specific addressing).
//
// Still used for:
//   - Service registration (choosing default interface for A record)
//   - Graceful degradation when interfaceIndex=0 (control messages unavailable)
//
// Returns:
//   - []byte: IPv4 address (4 bytes)
//   - error: if no suitable address found
//
// T037: Marked as deprecated for response building (007-interface-specific-addressing)
func getLocalIPv4() ([]byte, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4, nil
			}
		}
	}

	return nil, fmt.Errorf("no non-loopback IPv4 address found")
}

// getIPv4ForInterface returns the IPv4 address assigned to the specified network interface.
//
// RFC 6762 §15 "Responding to Address Queries" (lines 1020-1024):
//
//	When a Multicast DNS responder sends a Multicast DNS response message
//	containing its own address records, it MUST include all addresses
//	that are valid on the interface on which it is sending the message,
//	and MUST NOT include addresses that are not valid on that interface.
//
// This function enables RFC compliance by looking up the interface-specific IP address
// for building mDNS responses that contain ONLY the address valid on the receiving interface.
//
// 007-interface-specific-addressing: T014-T020 implementation
//
// Parameters:
//   - ifIndex: Network interface index (from Transport.Receive or ipv4.ControlMessage.IfIndex)
//
// Returns:
//   - []byte: IPv4 address (4 bytes) in network byte order
//   - error: NetworkError if interface not found, ValidationError if no IPv4 address
//
// Edge Cases:
//   - Interface not found (removed/down) → NetworkError
//   - Interface has no IPv4 address (IPv6-only) → ValidationError
//   - Interface has multiple IPs → returns first IPv4 (consistent behavior)
//
// Example:
//
//	ipv4, err := getIPv4ForInterface(2)  // Look up interface index 2 (e.g., wlan0)
//	if err != nil {
//	    // Handle error: skip response or fall back to getLocalIPv4()
//	}
//	// Use ipv4 in A record for mDNS response
func getIPv4ForInterface(ifIndex int) ([]byte, error) {
	// T015: Look up interface by index
	iface, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		// T018: Interface not found (removed, invalid index, etc.)
		return nil, &errors.NetworkError{
			Operation: "lookup interface",
			Err:       err,
			Details:   fmt.Sprintf("interface index %d not found", ifIndex),
		}
	}

	// T016: Get all addresses for this interface
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, &errors.NetworkError{
			Operation: "get interface addresses",
			Err:       err,
			Details:   fmt.Sprintf("failed to get addresses for %s", iface.Name),
		}
	}

	// T017: Filter for first IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4, nil
			}
		}
	}

	// T019: No IPv4 found on this interface
	return nil, &errors.ValidationError{
		Field:   "interface",
		Value:   iface.Name,
		Message: "no IPv4 address found on interface",
	}
}
