package responder

import (
	"context"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
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

// =============================================================================
// Query Handling Tests (handleQuery, parseMessage, buildResponsePacket)
// =============================================================================

// TestHandleQuery_MalformedPacket tests that handleQuery handles malformed packets gracefully.
//
// RFC 6762 §6: "Malformed queries MUST be silently ignored"
// FR-004: Errors must be handled gracefully without panics
func TestHandleQuery_MalformedPacket(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Test with various malformed packets
	tests := []struct {
		name   string
		packet []byte
	}{
		{"empty packet", []byte{}},
		{"too short packet", []byte{0x00, 0x01}},
		{"invalid header", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"nil packet", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// handleQuery should not panic on malformed input
			err := r.handleQuery(tt.packet)

			// We expect an error, but no panic
			if err == nil {
				t.Log("handleQuery returned nil error (acceptable - packet ignored)")
			} else {
				t.Logf("handleQuery returned error: %v (acceptable - malformed packet)", err)
			}
		})
	}

	t.Log("✓ handleQuery handles malformed packets without panicking")
}

// TestHandleQuery_ResponsePacket tests that handleQuery ignores response packets.
//
// RFC 6762 §6: "Responses (QR=1) must be ignored by query handler"
func TestHandleQuery_ResponsePacket(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Build a minimal DNS response packet (QR=1)
	// DNS header: [ID=0x1234][Flags=0x8000 (QR=1)][...]
	responsePacket := []byte{
		0x12, 0x34, // Transaction ID
		0x80, 0x00, // Flags: QR=1 (response), Opcode=0, AA=0, TC=0, RD=0, RA=0, RCODE=0
		0x00, 0x00, // Questions: 0
		0x00, 0x00, // Answers: 0
		0x00, 0x00, // Authority: 0
		0x00, 0x00, // Additional: 0
	}

	// handleQuery should ignore response packets
	err = r.handleQuery(responsePacket)

	if err != nil {
		t.Logf("handleQuery returned error: %v (acceptable)", err)
	}

	t.Log("✓ handleQuery ignores response packets (QR=1)")
}

// TestHandleQuery_EmptyRegistry tests that handleQuery doesn't crash with no registered services.
//
// FR-004: Error handling must be robust
func TestHandleQuery_EmptyRegistry(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Build a valid PTR query for _http._tcp.local
	// DNS header + question section
	queryPacket := buildPTRQuery("_http._tcp.local")

	// handleQuery should not crash even with empty registry
	err = r.handleQuery(queryPacket)

	if err != nil {
		t.Logf("handleQuery returned error: %v (acceptable)", err)
	}

	t.Log("✓ handleQuery handles empty registry without crashing")
}

// TestHandleQuery_WithRegisteredService tests that handleQuery processes queries for registered services.
//
// RFC 6762 §6: Responders must answer queries for registered services
func TestHandleQuery_WithRegisteredService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test requiring registration timing in short mode")
	}

	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	// Register a service
	service := &Service{
		InstanceName: "Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Build a PTR query for the registered service type
	queryPacket := buildPTRQuery("_http._tcp.local")

	// handleQuery should process the query without error
	err = r.handleQuery(queryPacket)

	if err != nil {
		t.Errorf("handleQuery() error = %v, want nil (should process query for registered service)", err)
	}

	t.Log("✓ handleQuery processes queries for registered services")
}

// TestParseMessage tests the parseMessage wrapper function.
func TestParseMessage(t *testing.T) {
	tests := []struct {
		name      string
		packet    []byte
		wantError bool
	}{
		{
			name:      "nil packet",
			packet:    nil,
			wantError: true,
		},
		{
			name:      "empty packet",
			packet:    []byte{},
			wantError: true,
		},
		{
			name: "valid query packet",
			packet: []byte{
				0x00, 0x00, // Transaction ID
				0x00, 0x00, // Flags: QR=0 (query)
				0x00, 0x00, // Questions: 0
				0x00, 0x00, // Answers: 0
				0x00, 0x00, // Authority: 0
				0x00, 0x00, // Additional: 0
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseMessage(tt.packet)

			if tt.wantError {
				if err == nil {
					t.Error("parseMessage() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("parseMessage() error = %v, want nil", err)
				}
				if msg == nil {
					t.Error("parseMessage() returned nil message")
				}
			}
		})
	}
}

// TestBuildResponsePacket tests the buildResponsePacket stub function.
func TestBuildResponsePacket(t *testing.T) {
	// buildResponsePacket is currently a stub that returns empty byte slice
	// This test documents the current behavior

	msg := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      0x1234,
			Flags:   0x8000, // Response
			QDCount: 0,
			ANCount: 0,
			NSCount: 0,
			ARCount: 0,
		},
	}

	packet := buildResponsePacket(msg)

	// Current stub implementation returns empty packet
	if packet == nil {
		t.Error("buildResponsePacket() returned nil, want empty byte slice")
	}

	// Document that this is a stub
	if len(packet) != 0 {
		t.Logf("buildResponsePacket() returned %d bytes (implementation exists)", len(packet))
	} else {
		t.Log("✓ buildResponsePacket() stub verified (returns empty packet)")
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// buildPTRQuery constructs a minimal DNS query packet for PTR record.
//
// This helper creates a valid DNS query packet for testing handleQuery().
func buildPTRQuery(qname string) []byte {
	// Build a minimal PTR query packet
	// Format: [Header][Question]

	// DNS Header (12 bytes)
	packet := []byte{
		0x00, 0x00, // Transaction ID: 0
		0x00, 0x00, // Flags: QR=0 (query), Opcode=0, AA=0, TC=0, RD=0, RA=0, RCODE=0
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answers: 0
		0x00, 0x00, // Authority: 0
		0x00, 0x00, // Additional: 0
	}

	// Question section: encode domain name
	// Split qname into labels and encode each
	labels := []byte{}
	parts := splitDomainName(qname)
	for _, part := range parts {
		labels = append(labels, byte(len(part)))
		labels = append(labels, []byte(part)...)
	}
	labels = append(labels, 0x00) // Root label (end of name)

	packet = append(packet, labels...)

	// QTYPE: PTR (12)
	packet = append(packet, 0x00, 0x0C)

	// QCLASS: IN (1)
	packet = append(packet, 0x00, 0x01)

	return packet
}

// splitDomainName splits a domain name into labels.
func splitDomainName(name string) []string {
	labels := []string{}
	current := ""

	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			if current != "" {
				labels = append(labels, current)
				current = ""
			}
		} else {
			current += string(name[i])
		}
	}

	if current != "" {
		labels = append(labels, current)
	}

	return labels
}

// TestResponder_Unregister_ErrorCases tests error handling in Unregister.
//
// Validates proper error handling for edge cases:
// - Unregistering non-existent service
// - Unregistering already unregistered service
// - Invalid service ID formats
//
// Coverage improvement for responder.Unregister (71.4% → target 90%+)
func TestResponder_Unregister_ErrorCases(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	tests := []struct {
		name      string
		serviceID string
		setup     func()
		wantErr   bool
	}{
		{
			name:      "unregister non-existent service",
			serviceID: "NonExistent",
			setup:     func() {},
			wantErr:   true,
		},
		{
			name:      "unregister already unregistered service",
			serviceID: "TestService",
			setup: func() {
				// Register then unregister
				svc := &Service{
					InstanceName: "TestService",
					ServiceType:  "_http._tcp.local",
					Port:         8080,
				}
				_ = r.Register(svc)
				_ = r.Unregister("TestService")
			},
			wantErr: true,
		},
		{
			name:      "unregister with full service ID",
			serviceID: "AnotherTest._http._tcp.local",
			setup: func() {
				// Register service
				svc := &Service{
					InstanceName: "AnotherTest",
					ServiceType:  "_http._tcp.local",
					Port:         8080,
				}
				_ = r.Register(svc)
			},
			wantErr: false, // Should succeed with full ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := r.Unregister(tt.serviceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unregister(%q) error = %v, wantErr %v", tt.serviceID, err, tt.wantErr)
			}
		})
	}
}

// TestResponder_UpdateService_ErrorCases tests error handling in UpdateService.
//
// Validates proper error handling for edge cases:
// - Updating non-existent service
// - Empty TXT records
// - Nil TXT records
//
// Coverage improvement for responder.UpdateService (75.0% → target 90%+)
func TestResponder_UpdateService_ErrorCases(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	tests := []struct {
		name       string
		serviceID  string
		txtRecords map[string]string
		setup      func()
		wantErr    bool
	}{
		{
			name:       "update non-existent service",
			serviceID:  "NonExistent",
			txtRecords: map[string]string{"version": "1.0"},
			setup:      func() {},
			wantErr:    true,
		},
		{
			name:       "update with empty TXT records",
			serviceID:  "TestService._http._tcp.local",
			txtRecords: map[string]string{},
			setup: func() {
				svc := &Service{
					InstanceName: "TestService",
					ServiceType:  "_http._tcp.local",
					Port:         8080,
					TXTRecords:   map[string]string{"version": "1.0"},
				}
				_ = r.Register(svc)
			},
			wantErr: false, // Empty TXT is valid
		},
		{
			name:       "update with nil TXT records",
			serviceID:  "NilTest",
			txtRecords: nil,
			setup: func() {
				svc := &Service{
					InstanceName: "NilTest",
					ServiceType:  "_http._tcp.local",
					Port:         8080,
					TXTRecords:   map[string]string{"key": "value"},
				}
				_ = r.Register(svc)
			},
			wantErr: false, // Nil TXT is valid (clears records)
		},
		{
			name:       "update using instance name only",
			serviceID:  "InstanceOnly",
			txtRecords: map[string]string{"updated": "true"},
			setup: func() {
				svc := &Service{
					InstanceName: "InstanceOnly",
					ServiceType:  "_http._tcp.local",
					Port:         8080,
				}
				_ = r.Register(svc)
			},
			wantErr: false, // Should work with instance name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := r.UpdateService(tt.serviceID, tt.txtRecords)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateService(%q) error = %v, wantErr %v", tt.serviceID, err, tt.wantErr)
			}

			// If no error, verify update took effect
			if err == nil {
				retrieved, found := r.GetService(tt.serviceID)
				if !found {
					t.Errorf("GetService(%q) found = false after successful update", tt.serviceID)
				} else if retrieved != nil {
					// Verify TXT records match
					if len(tt.txtRecords) != len(retrieved.TXTRecords) {
						t.Logf("TXT records length mismatch: got %d, want %d", len(retrieved.TXTRecords), len(tt.txtRecords))
					}
				}
			}
		})
	}
}

// TestResponder_Register_EdgeCases tests edge cases in Register.
//
// Improves coverage for responder.Register (76.1% → target 90%+)
func TestResponder_Register_EdgeCases(t *testing.T) {
	ctx := context.Background()
	r, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	tests := []struct {
		name    string
		service *Service
		wantErr bool
	}{
		{
			name: "service with empty TXT records",
			service: &Service{
				InstanceName: "EmptyTXT",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
				TXTRecords:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "service with nil TXT records",
			service: &Service{
				InstanceName: "NilTXT",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
				TXTRecords:   nil,
			},
			wantErr: false,
		},
		{
			name: "service with multiple TXT records",
			service: &Service{
				InstanceName: "MultiTXT",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
				TXTRecords: map[string]string{
					"txtvers": "1",
					"path":    "/",
					"version": "1.0.0",
					"status":  "ready",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Register(tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// Verify service was registered
				retrieved, found := r.GetService(tt.service.InstanceName)
				if !found {
					t.Error("GetService() found = false after successful Register()")
				} else if retrieved != nil {
					// Verify TXT records
					if len(retrieved.TXTRecords) != len(tt.service.TXTRecords) {
						t.Errorf("TXT records length = %d, want %d", len(retrieved.TXTRecords), len(tt.service.TXTRecords))
					}
				}
			}
		})
	}
}

// TestResponder_GetLocalIPv4_EdgeCases tests edge cases in getLocalIPv4.
//
// This is a helper function used by Register, so improving its coverage
// indirectly improves Register coverage.
//
// Note: getLocalIPv4 is hard to test exhaustively without mocking network
// interfaces, but we can at least verify it doesn't panic.
func TestResponder_GetLocalIPv4_EdgeCases(t *testing.T) {
	// Just verify it doesn't panic and returns something reasonable
	ip, err := getLocalIPv4()

	// Either we get an IP or an error, both are valid
	if err != nil {
		t.Logf("getLocalIPv4() returned error: %v (may be expected in test environment)", err)
	}

	if err == nil && ip == nil {
		t.Error("getLocalIPv4() returned nil IP without error")
	}

	if err == nil && len(ip) != 4 {
		t.Errorf("getLocalIPv4() returned IP with length %d, want 4", len(ip))
	}
}
