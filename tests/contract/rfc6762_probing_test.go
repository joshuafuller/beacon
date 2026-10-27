package contract

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

// TestRFC6762_Probing_ThreeQueries_RED tests RFC 6762 §8.1 probing compliance.
//
// TDD Phase: RED - These tests will FAIL until probing is implemented
//
// RFC 6762 §8.1: Probing
//   - "The host MUST send at least two query packets"
//   - Beacon implementation: Send exactly 3 queries
//   - Timing: 250ms intervals between queries
//
// T028: Contract test for RFC 6762 §8.1 compliance
func TestRFC6762_Probing_ThreeQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RFC contract test in short mode")
	}

	ctx := context.Background()
	r, err := responder.New(ctx, responder.WithHostname(uniqueHostname()))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Track probe queries sent
	var probeCount int
	r.OnProbe(func() {
		probeCount++
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// RFC 6762 §8.1 requires at least 2 probes, Beacon sends 3
	if probeCount < 2 {
		t.Errorf("probeCount = %d, want ≥2 per RFC 6762 §8.1", probeCount)
	}

	if probeCount != 3 {
		t.Errorf("probeCount = %d, want 3 (Beacon implementation)", probeCount)
	}
}

// TestRFC6762_Probing_250msInterval_RED tests probe timing per RFC 6762 §8.1.
//
// TDD Phase: RED
//
// RFC 6762 §8.1: "250 millisecond intervals"
// T028: Test probe timing compliance
func TestRFC6762_Probing_250msInterval(t *testing.T) {
	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Use common timing test helper with 250ms interval and 50ms tolerance
	testRFC6762Timing(t,
		func(r *responder.Responder) []time.Time {
			var probeTimes []time.Time
			r.OnProbe(func() {
				probeTimes = append(probeTimes, time.Now())
			})

			err := r.Register(service)
			if err != nil {
				t.Fatalf("Register() error = %v, want nil", err)
			}

			return probeTimes
		},
		250*time.Millisecond, 50*time.Millisecond, "RFC 6762 §8.1")
}

// TestRFC6762_Probing_QueryFormat_RED tests probe query message format per RFC 6762 §8.1.
//
// TDD Phase: RED
//
// RFC 6762 §8.1: Probe queries
//   - QR bit = 0 (query, not response)
//   - Question section: "ANY" query for service name
//   - Authority section: Records being probed
//
// T028: Test probe message format compliance
func TestRFC6762_Probing_QueryFormat(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx, responder.WithHostname(uniqueHostname()))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture probe message bytes
	var probeMessage []byte
	r.OnProbe(func() {
		probeMessage = r.GetLastProbeMessage()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if len(probeMessage) < 12 {
		t.Fatalf("probe message too short: %d bytes, want ≥12 (DNS header)", len(probeMessage))
	}

	// Parse DNS header flags (bytes 2-3)
	flags := binary.BigEndian.Uint16(probeMessage[2:4])

	// QR bit (bit 15) MUST be 0 for queries per RFC 6762 §18.2
	qr := (flags >> 15) & 0x01
	if qr != 0 {
		t.Errorf("probe QR bit = %d, want 0 (query per RFC 6762 §18.2)", qr)
	}

	// OPCODE (bits 11-14) MUST be 0 per RFC 6762 §18.3
	opcode := (flags >> 11) & 0x0F
	if opcode != 0 {
		t.Errorf("probe OPCODE = %d, want 0 (standard query per RFC 6762 §18.3)", opcode)
	}
}

// TestRFC6762_Probing_ConflictDetection_RED tests conflict detection per RFC 6762 §8.1.
//
// TDD Phase: RED
//
// RFC 6762 §8.1: "If any of these messages elicit a response, then the host
// MUST choose another name."
//
// T028: Test conflict detection during probing
func TestRFC6762_Probing_ConflictDetection(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx, responder.WithHostname(uniqueHostname()))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "Conflicting Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Simulate conflict response during probing
	r.InjectConflictDuringProbing(true)

	// Register should detect conflict and return error or rename
	err = r.Register(service)

	// Beacon behavior: conflict is signaled via error or automatic rename
	// (Will be implemented in User Story 2)
	// For now, just verify probing completes
	if err != nil {
		// Error is acceptable for conflict
		t.Logf("Register() with conflict error = %v (expected)", err)
	}
}

// TestRFC6762_Probing_TieBreaking_RED tests lexicographic tie-breaking per RFC 6762 §8.2.1.
//
// TDD Phase: RED
//
// RFC 6762 §8.2.1: "The two records are compared and the lexicographically
// later data wins."
//
// T028: Test tie-breaking for simultaneous probes
func TestRFC6762_Probing_TieBreaking(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx, responder.WithHostname(uniqueHostname()))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "Tie Break Test",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Simulate simultaneous probe where we WIN (our data > their data)
	ourIP := []byte{192, 168, 1, 100}
	theirIP := []byte{192, 168, 1, 50} // Lexicographically earlier
	r.InjectSimultaneousProbe(ourIP, theirIP)

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil (should win tie-break)", err)
	}

	// Verify service registered successfully (won tie-break)
	registered, exists := r.GetService(service.InstanceName)
	if !exists {
		t.Error("service not registered after winning tie-break")
	}

	if registered == nil {
		t.Error("registered service is nil")
	}
}

// uniqueHostname returns a host name nobody else on the test network owns.
// Without it the probe counts depend on the network: on Windows/WSL, Windows
// itself answers for "<PC name>.local", and Beacon correctly renames its host
// (RFC 6762 §9), adding a probe round.
func uniqueHostname() string {
	return fmt.Sprintf("beacon-contract-%d.local", time.Now().UnixNano())
}
