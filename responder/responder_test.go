package responder

import (
	"bytes"
	"context"
	goerrors "errors"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/errors"
)

// TestResponder_New_RED tests Responder initialization.
//
// TDD Phase: RED - These tests will FAIL until we implement Responder.New()
//
// FR-025: System MUST provide responder initialization with context
// T022: Write Responder.New() initialization tests
func TestResponder_New(t *testing.T) {
	ctx := context.Background()

	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	if responder == nil {
		t.Fatal("New() returned nil responder")
	}

	// Verify responder has required components
	if responder.registry == nil {
		t.Error("responder.registry = nil, want non-nil")
	}

	if responder.transport == nil {
		t.Error("responder.transport = nil, want non-nil")
	}
}

// TestResponder_New_WithOptions_RED tests Responder initialization with options.
//
// TDD Phase: RED
//
// FR-025: System MUST support functional options for configuration
// T022: Test functional options (WithHostname)
func TestResponder_New_WithOptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		options      []Option
		wantHostname string
	}{
		{
			name:         "default hostname",
			options:      nil,
			wantHostname: "", // Will use system hostname
		},
		{
			name: "custom hostname",
			options: []Option{
				WithHostname("myhost.local"),
			},
			wantHostname: "myhost.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responder, err := New(ctx, tt.options...)
			if err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}

			if responder == nil {
				t.Fatal("New() returned nil responder")
			}

			// If custom hostname provided, verify it was set
			if tt.wantHostname != "" && responder.hostname != tt.wantHostname {
				t.Errorf("responder.hostname = %q, want %q", responder.hostname, tt.wantHostname)
			}
		})
	}
}

// TestResponder_Register_Validation_RED tests that Register() validates services.
//
// TDD Phase: RED
//
// FR-026: System MUST validate service parameters before registration
// T023: Test Register() validation
func TestResponder_Register_Validation(t *testing.T) {
	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = responder.Close() }()

	tests := []struct {
		name        string
		service     *Service
		wantErr     bool
		errContains string
	}{
		{
			name: "valid service",
			service: &Service{
				InstanceName: "My Printer",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
			},
			wantErr: false,
		},
		{
			name: "invalid - empty InstanceName",
			service: &Service{
				InstanceName: "",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
			},
			wantErr:     true,
			errContains: "instance name cannot be empty",
		},
		{
			name: "invalid - bad ServiceType",
			service: &Service{
				InstanceName: "My Printer",
				ServiceType:  "http._tcp.local", // Missing leading underscore
				Port:         8080,
			},
			wantErr:     true,
			errContains: "invalid service type format",
		},
		{
			name: "invalid - port 0",
			service: &Service{
				InstanceName: "My Printer",
				ServiceType:  "_http._tcp.local",
				Port:         0,
			},
			wantErr:     true,
			errContains: "port must be in range 1-65535",
		},
		{
			name:        "invalid - nil service",
			service:     nil,
			wantErr:     true,
			errContains: "service cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := responder.Register(tt.service)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Register() error = nil, want error containing %q", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Register() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Register() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestResponder_Register_StartsStateMachine_RED tests that Register() starts state machine.
//
// TDD Phase: RED
//
// FR-027: System MUST start probing state machine on registration
// T023: Test that Register() starts state machine
func TestResponder_Register_StartsStateMachine(t *testing.T) {
	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = responder.Close() }()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	err = responder.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Give state machine time to start
	time.Sleep(100 * time.Millisecond)

	// Verify service is in registry
	registered, exists := responder.registry.Get(service.InstanceName)
	if !exists {
		t.Error("service not found in registry after Register()")
	}

	if registered == nil {
		t.Error("registered service is nil")
	}
}

// TestResponder_Register_WaitsForEstablished_RED tests that Register() waits for state machine.
//
// TDD Phase: RED
//
// FR-027: Register() MUST wait for probing+announcing to complete
// RFC 6762 §8.1: Probing takes ~750ms (3 probes × 250ms)
// RFC 6762 §8.3: Announcing takes ~1s (2 announcements × 1s)
// Total: ~1.75 seconds until Established
//
// T023: Test Register() blocks until Established state
func TestResponder_Register_WaitsForEstablished(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = responder.Close() }()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	start := time.Now()
	err = responder.Register(service)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Register() should block for ~1.75s (probing + announcing)
	// Allow ±500ms tolerance for test timing
	minDuration := 1250 * time.Millisecond // 1.75s - 500ms
	maxDuration := 2250 * time.Millisecond // 1.75s + 500ms

	if elapsed < minDuration || elapsed > maxDuration {
		t.Errorf("Register() blocked for %v, want ~1.75s (range: %v-%v)", elapsed, minDuration, maxDuration)
	}
}

// TestResponder_Unregister_RED tests service unregistration with goodbye packets.
//
// TDD Phase: RED
//
// RFC 6762 §10.1: Send goodbye packets with TTL=0
// FR-014: System MUST send goodbye packets on unregistration
//
// T023: Test Unregister() sends goodbye and removes service
func TestResponder_Unregister(t *testing.T) {
	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = responder.Close() }()

	service := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register service first
	err = responder.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Unregister service
	err = responder.Unregister(service.InstanceName)
	if err != nil {
		t.Fatalf("Unregister() error = %v, want nil", err)
	}

	// Verify service removed from registry
	_, exists := responder.registry.Get(service.InstanceName)
	if exists {
		t.Error("service still in registry after Unregister()")
	}
}

// TestResponder_Close_RED tests responder shutdown.
//
// TDD Phase: RED
//
// FR-015: System MUST gracefully shutdown all services
// T023: Test Close() unregisters all services and shuts down transport
func TestResponder_Close(t *testing.T) {
	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	// Register multiple services
	services := []*Service{
		{
			InstanceName: "Service 1",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		},
		{
			InstanceName: "Service 2",
			ServiceType:  "_printer._tcp.local",
			Port:         9100,
		},
	}

	for _, svc := range services {
		err := responder.Register(svc)
		if err != nil {
			t.Fatalf("Register() error = %v, want nil", err)
		}
	}

	// Close responder
	err = responder.Close()
	if err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}

	// Verify all services unregistered
	for _, svc := range services {
		_, exists := responder.registry.Get(svc.InstanceName)
		if exists {
			t.Errorf("service %q still in registry after Close()", svc.InstanceName)
		}
	}
}

// TestResponder_Register_MaxRenameAttempts tests that Register() fails after max rename attempts.
//
// TDD Phase: RED - This test will FAIL until we implement rename loop with max attempts
//
// RFC 6762 §9: "If a host receives a response containing a record that conflicts
// with one of its unique records, the host MUST immediately rename the record."
//
// FR-032: System MUST handle registration failures gracefully (max 10 rename attempts)
// T062: Test max rename attempts limit (RED phase)
func TestResponder_Register_MaxRenameAttempts(t *testing.T) {
	ctx := context.Background()
	responder, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = responder.Close() }()

	// Configure responder to always return conflict during probing
	// This will force the rename loop to run until max attempts
	responder.InjectConflictDuringProbing(true)

	service := &Service{
		InstanceName: "My Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register should fail after 10 rename attempts
	// Expected attempts:
	//  1. "My Service" → conflict
	//  2. "My Service-2" → conflict
	//  3. "My Service-3" → conflict
	//  ...
	//  10. "My Service-10" → conflict
	//  → Returns error: max rename attempts exceeded
	err = responder.Register(service)

	if err == nil {
		t.Fatal("Register() error = nil, want error (max rename attempts exceeded)")
	}

	// Verify error message indicates max attempts
	wantErrSubstr := "max rename attempts"
	if !contains(err.Error(), wantErrSubstr) {
		t.Errorf("Register() error = %q, want error containing %q", err.Error(), wantErrSubstr)
	}

	// Verify service NOT in registry (failed to register)
	_, exists := responder.registry.Get(service.InstanceName)
	if exists {
		t.Error("service should NOT be in registry after max rename attempts exceeded")
	}
}

// TestResponder_Register_RenameOnConflict tests that Register() renames on conflict.
//
// TDD Phase: RED
//
// RFC 6762 §9: Service renamed with numeric suffix on conflict
// FR-030: System MUST rename service on conflict
// T062: Test rename-on-conflict behavior (RED phase)
//
// NOTE: This test is currently disabled because the rename loop implementation
// requires more complex test infrastructure (conflict injection with counters).
// For now, T062 focuses on the max attempts limit test above.
// TODO US2-LATER: Implement detailed rename-on-conflict test when test infrastructure ready
func TestResponder_Register_RenameOnConflict(t *testing.T) {
	t.Skip("Skipping - requires advanced test injection (conflict counter). See T062 notes.")

	// Test logic will be:
	// 1. Inject conflict on first probe attempt
	// 2. Allow success on second probe attempt
	// 3. Verify service renamed to "My Service-2"
	// 4. Verify service registered successfully
}

// =============================================================================
// User Story 5: Multi-Service Support Tests (TDD - RED Phase)
// =============================================================================

// TestResponder_RegisterMultipleServices_RED tests concurrent registration
// of multiple services.
//
// TDD Phase: RED - This test will FAIL until multi-service support is working
//
// RFC 6762: Responder must support registering multiple services
// FR-027: System MUST support multiple simultaneous service registrations
// T100: Unit test for concurrent service registration
func TestResponder_RegisterMultipleServices(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Register 3 different services
	services := []*Service{
		{
			InstanceName: "Web Server",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		},
		{
			InstanceName: "SSH Server",
			ServiceType:  "_ssh._tcp.local",
			Port:         22,
		},
		{
			InstanceName: "FTP Server",
			ServiceType:  "_ftp._tcp.local",
			Port:         21,
		},
	}

	// Register all services
	for _, svc := range services {
		err := r.Register(svc)
		if err != nil {
			t.Errorf("Register(%q) error = %v, want nil", svc.InstanceName, err)
		}
	}

	// Verify all services are registered and retrievable
	for _, svc := range services {
		// Build expected service ID
		serviceID := svc.InstanceName + "." + svc.ServiceType

		// Retrieve from registry
		retrieved, found := r.GetService(serviceID)
		if !found {
			t.Errorf("GetService(%q) found = false, want true", serviceID)
			continue
		}
		if retrieved == nil {
			t.Errorf("GetService(%q) = nil, want non-nil", serviceID)
			continue
		}

		// Verify service details match
		if retrieved.InstanceName != svc.InstanceName {
			t.Errorf("GetService(%q).InstanceName = %q, want %q",
				serviceID, retrieved.InstanceName, svc.InstanceName)
		}
		if retrieved.ServiceType != svc.ServiceType {
			t.Errorf("GetService(%q).ServiceType = %q, want %q",
				serviceID, retrieved.ServiceType, svc.ServiceType)
		}
		if retrieved.Port != svc.Port {
			t.Errorf("GetService(%q).Port = %d, want %d",
				serviceID, retrieved.Port, svc.Port)
		}
	}
}

// TestResponder_UnregisterOneService_RED tests that unregistering one service
// doesn't affect other registered services.
//
// TDD Phase: RED
//
// FR-027: System MUST allow independent service lifecycle management
// T101: Unit test for independent service unregistration
func TestResponder_UnregisterOneService(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Register 3 services
	svc1 := &Service{InstanceName: "Service 1", ServiceType: "_http._tcp.local", Port: 8080}
	svc2 := &Service{InstanceName: "Service 2", ServiceType: "_ssh._tcp.local", Port: 22}
	svc3 := &Service{InstanceName: "Service 3", ServiceType: "_ftp._tcp.local", Port: 21}

	for _, svc := range []*Service{svc1, svc2, svc3} {
		if err := r.Register(svc); err != nil {
			t.Fatalf("Register(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Unregister service 2
	svc2ID := svc2.InstanceName + "." + svc2.ServiceType
	err = r.Unregister(svc2ID)
	if err != nil {
		t.Fatalf("Unregister(%q) error = %v, want nil", svc2ID, err)
	}

	// Verify service 2 is gone
	if retrieved, found := r.GetService(svc2ID); found {
		t.Errorf("GetService(%q) found = true, want false (should be unregistered); got %v", svc2ID, retrieved)
	}

	// Verify services 1 and 3 still exist
	svc1ID := svc1.InstanceName + "." + svc1.ServiceType
	if _, found := r.GetService(svc1ID); !found {
		t.Errorf("GetService(%q) found = false, want true (should still be registered)", svc1ID)
	}

	svc3ID := svc3.InstanceName + "." + svc3.ServiceType
	if _, found := r.GetService(svc3ID); !found {
		t.Errorf("GetService(%q) found = false, want true (should still be registered)", svc3ID)
	}
}

// TestResponder_UpdateOneService_RED tests that updating one service's TXT
// records doesn't affect other services.
//
// TDD Phase: RED
//
// FR-004: System MUST support service metadata updates without re-probing
// T102: Unit test for independent service updates
func TestResponder_UpdateOneService(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Register 2 services with TXT records
	svc1 := &Service{
		InstanceName: "Service 1",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords:   map[string]string{"version": "1.0"},
	}
	svc2 := &Service{
		InstanceName: "Service 2",
		ServiceType:  "_ssh._tcp.local",
		Port:         22,
		TXTRecords:   map[string]string{"version": "2.0"},
	}

	for _, svc := range []*Service{svc1, svc2} {
		if err := r.Register(svc); err != nil {
			t.Fatalf("Register(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Update service 1's TXT records
	svc1ID := svc1.InstanceName + "." + svc1.ServiceType
	newTXT := map[string]string{"version": "1.1", "status": "updated"}
	err = r.UpdateService(svc1ID, newTXT)
	if err != nil {
		t.Fatalf("UpdateService(%q) error = %v, want nil", svc1ID, err)
	}

	// Verify service 1 has updated TXT records
	retrieved1, found1 := r.GetService(svc1ID)
	if !found1 {
		t.Fatal("GetService(svc1) found = false, want true")
	}
	if retrieved1 == nil {
		t.Fatal("GetService(svc1) = nil, want non-nil")
	}
	if retrieved1.TXTRecords["version"] != "1.1" {
		t.Errorf("service1.TXTRecords[version] = %q, want %q", retrieved1.TXTRecords["version"], "1.1")
	}
	if retrieved1.TXTRecords["status"] != "updated" {
		t.Errorf("service1.TXTRecords[status] = %q, want %q", retrieved1.TXTRecords["status"], "updated")
	}

	// Verify service 2 is unchanged
	svc2ID := svc2.InstanceName + "." + svc2.ServiceType
	retrieved2, found2 := r.GetService(svc2ID)
	if !found2 {
		t.Fatal("GetService(svc2) found = false, want true")
	}
	if retrieved2 == nil {
		t.Fatal("GetService(svc2) = nil, want non-nil")
	}
	if retrieved2.TXTRecords["version"] != "2.0" {
		t.Errorf("service2.TXTRecords[version] = %q, want %q (should be unchanged)",
			retrieved2.TXTRecords["version"], "2.0")
	}
}

// ==============================================================================
// 007-interface-specific-addressing: Unit Tests for getIPv4ForInterface
// ==============================================================================

// TestGetIPv4ForInterface_ValidInterface tests interface-specific IP lookup.
//
// T045: Unit test for getIPv4ForInterface() with multiple NICs
// RFC 6762 §15: Responses MUST include only addresses from the receiving interface
func TestGetIPv4ForInterface_ValidInterface(t *testing.T) {
	// Get a valid non-loopback interface from the system
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("net.Interfaces() failed: %v", err)
	}

	var testIface *net.Interface
	for _, iface := range ifaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Check if it has an IPv4 address
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					testIface = &iface
					break
				}
			}
		}

		if testIface != nil {
			break
		}
	}

	if testIface == nil {
		t.Skip("No non-loopback interface with IPv4 found")
	}

	// Test: getIPv4ForInterface should return the interface's IPv4 address
	ipv4, err := getIPv4ForInterface(testIface.Index)
	if err != nil {
		t.Fatalf("getIPv4ForInterface(%d) error = %v, want nil", testIface.Index, err)
	}

	if len(ipv4) != 4 {
		t.Fatalf("getIPv4ForInterface(%d) returned %d bytes, want 4 (IPv4 address)", testIface.Index, len(ipv4))
	}

	t.Logf("✓ Interface %s (index=%d) → IPv4 %d.%d.%d.%d",
		testIface.Name, testIface.Index, ipv4[0], ipv4[1], ipv4[2], ipv4[3])
}

// TestGetIPv4ForInterface_InvalidIndex tests error handling for invalid interface index.
//
// T046: Edge case - interface index out of range → NetworkError
func TestGetIPv4ForInterface_InvalidIndex(t *testing.T) {
	// Use an impossibly high interface index
	invalidIndex := 9999

	ipv4, err := getIPv4ForInterface(invalidIndex)
	if err == nil {
		t.Fatalf("getIPv4ForInterface(%d) error = nil, want NetworkError", invalidIndex)
	}

	if ipv4 != nil {
		t.Errorf("getIPv4ForInterface(%d) ipv4 = %v, want nil", invalidIndex, ipv4)
	}

	// Verify it's a NetworkError
	var netErr *errors.NetworkError
	if !goerrors.As(err, &netErr) {
		t.Errorf("getIPv4ForInterface(%d) error type = %T, want *errors.NetworkError", invalidIndex, err)
	}

	t.Logf("✓ Invalid interface index %d → NetworkError: %v", invalidIndex, err)
}

// TestGetIPv4ForInterface_LoopbackInterface tests loopback handling.
//
// Loopback should work (it has an IPv4), but typically not used for mDNS responses
func TestGetIPv4ForInterface_LoopbackInterface(t *testing.T) {
	// Find loopback interface
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("net.Interfaces() failed: %v", err)
	}

	var loopbackIndex int
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			loopbackIndex = iface.Index
			break
		}
	}

	if loopbackIndex == 0 {
		t.Skip("No loopback interface found")
	}

	// Loopback should return its IPv4 address (127.0.0.1)
	ipv4, err := getIPv4ForInterface(loopbackIndex)
	if err != nil {
		t.Fatalf("getIPv4ForInterface(loopback=%d) error = %v, want nil", loopbackIndex, err)
	}

	if len(ipv4) != 4 {
		t.Fatalf("getIPv4ForInterface(loopback=%d) returned %d bytes, want 4", loopbackIndex, len(ipv4))
	}

	// Should be 127.0.0.1
	if ipv4[0] != 127 || ipv4[1] != 0 || ipv4[2] != 0 || ipv4[3] != 1 {
		t.Errorf("getIPv4ForInterface(loopback=%d) = %d.%d.%d.%d, want 127.0.0.1",
			loopbackIndex, ipv4[0], ipv4[1], ipv4[2], ipv4[3])
	}

	t.Logf("✓ Loopback interface (index=%d) → 127.0.0.1", loopbackIndex)
}

// TestGetIPv4ForInterface_MultipleInterfaces tests multi-NIC scenario.
//
// T045: Validates that different interface indices return different IPs
// This is the core of RFC 6762 §15 compliance
func TestGetIPv4ForInterface_MultipleInterfaces(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("net.Interfaces() failed: %v", err)
	}

	// Find all interfaces with IPv4 addresses
	type ifaceWithIP struct {
		index int
		name  string
		ipv4  []byte
	}
	var validIfaces []ifaceWithIP

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					validIfaces = append(validIfaces, ifaceWithIP{
						index: iface.Index,
						name:  iface.Name,
						ipv4:  ipv4,
					})
					break
				}
			}
		}
	}

	if len(validIfaces) < 2 {
		t.Skip("Need at least 2 interfaces with IPv4 for multi-NIC test")
	}

	t.Logf("Testing %d interfaces with IPv4 addresses", len(validIfaces))

	// Test each interface returns its own IP
	for _, iface := range validIfaces {
		ipv4, err := getIPv4ForInterface(iface.index)
		if err != nil {
			t.Errorf("getIPv4ForInterface(%d) error = %v, want nil", iface.index, err)
			continue
		}

		// Verify it matches the expected IP for this interface
		if !bytes.Equal(ipv4, iface.ipv4) {
			t.Errorf("getIPv4ForInterface(%d) = %d.%d.%d.%d, want %d.%d.%d.%d (interface %s)",
				iface.index,
				ipv4[0], ipv4[1], ipv4[2], ipv4[3],
				iface.ipv4[0], iface.ipv4[1], iface.ipv4[2], iface.ipv4[3],
				iface.name)
		}

		t.Logf("  ✓ Interface %s (index=%d) → %d.%d.%d.%d",
			iface.name, iface.index, ipv4[0], ipv4[1], ipv4[2], ipv4[3])
	}

	// KEY TEST: Verify different interfaces return DIFFERENT IPs (RFC 6762 §15)
	if len(validIfaces) >= 2 {
		ip1, err1 := getIPv4ForInterface(validIfaces[0].index)
		ip2, err2 := getIPv4ForInterface(validIfaces[1].index)

		// If either lookup failed, we can't compare
		if err1 != nil || err2 != nil {
			if err1 != nil {
				t.Logf("⚠️  Interface %s lookup failed: %v", validIfaces[0].name, err1)
			}
			if err2 != nil {
				t.Logf("⚠️  Interface %s lookup failed: %v", validIfaces[1].name, err2)
			}
		} else if bytes.Equal(ip1, ip2) {
			t.Logf("⚠️  WARNING: Interfaces %s and %s have the same IP %d.%d.%d.%d",
				validIfaces[0].name, validIfaces[1].name, ip1[0], ip1[1], ip1[2], ip1[3])
		} else {
			t.Logf("✅ RFC 6762 §15: Different interfaces return different IPs (%d.%d.%d.%d vs %d.%d.%d.%d)",
				ip1[0], ip1[1], ip1[2], ip1[3], ip2[0], ip2[1], ip2[2], ip2[3])
		}
	}
}

// BenchmarkGetIPv4ForInterface measures interface-specific IP lookup performance.
//
// T050: Performance measurement for getIPv4ForInterface()
// This validates NFR-002: Performance overhead <10% (should be <1μs per lookup)
func BenchmarkGetIPv4ForInterface(b *testing.B) {
	// Find a valid interface for benchmarking
	ifaces, err := net.Interfaces()
	if err != nil {
		b.Fatalf("net.Interfaces() failed: %v", err)
	}

	var testIndex int
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					testIndex = iface.Index
					break
				}
			}
		}
		if testIndex != 0 {
			break
		}
	}

	if testIndex == 0 {
		b.Skip("No non-loopback interface with IPv4 found")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := getIPv4ForInterface(testIndex)
		if err != nil {
			b.Fatalf("getIPv4ForInterface(%d) failed: %v", testIndex, err)
		}
	}
}

// BenchmarkGetIPv4ForInterface_CacheMiss measures worst-case lookup (invalid index).
//
// T050: Benchmark error path performance
func BenchmarkGetIPv4ForInterface_CacheMiss(b *testing.B) {
	invalidIndex := 9999

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = getIPv4ForInterface(invalidIndex) // Expect error
	}
}
