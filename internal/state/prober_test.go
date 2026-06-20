package state

import (
	"context"
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
)

const testServiceName = "My Printer._http._tcp.local"

// mockConflictDetector is a mock implementation of ConflictDetectorInterface for testing.
//
// T059: Mock ConflictDetector for integration testing
type mockConflictDetector struct {
	detectFunc func(ourRecord, incomingRecord message.ResourceRecord) (bool, error)
}

func (m *mockConflictDetector) DetectConflict(ourRecord, incomingRecord message.ResourceRecord) (bool, error) {
	if m.detectFunc != nil {
		return m.detectFunc(ourRecord, incomingRecord)
	}
	// Default: use simple lexicographic comparison
	if ourRecord.Name != incomingRecord.Name {
		return false, nil
	}
	// Compare data bytewise
	for i := 0; i < len(ourRecord.Data) && i < len(incomingRecord.Data); i++ {
		if ourRecord.Data[i] < incomingRecord.Data[i] {
			return true, nil // We lose
		} else if ourRecord.Data[i] > incomingRecord.Data[i] {
			return false, nil // We win
		}
	}
	// If equal or one is shorter, no conflict
	return false, nil
}

// TestProber_Probe_RED tests probing per RFC 6762 §8.1.
//
// TDD Phase: RED - These tests will FAIL until we implement Prober
//
// RFC 6762 §8.1: Probing
//   - Send 3 probe queries
//   - 250ms intervals between probes
//   - Total duration: ~750ms (0ms, 250ms, 500ms)
//
// FR-027: System MUST perform probing before announcing
// T024: Write prober tests
func TestProber_Probe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	ctx := context.Background()
	prober := NewProber()

	start := time.Now()
	result := prober.Probe(ctx, testServiceName)
	elapsed := time.Since(start)

	if result.Conflict {
		t.Error("Probe() detected conflict, want no conflict")
	}

	if result.Error != nil {
		t.Fatalf("Probe() error = %v, want nil", result.Error)
	}

	// Probing should take ~500ms (3 probes × 250ms intervals)
	// Allow ±200ms tolerance for test timing
	minDuration := 350 * time.Millisecond // 750ms - 200ms
	maxDuration := 650 * time.Millisecond // 750ms + 200ms

	if elapsed < minDuration || elapsed > maxDuration {
		t.Errorf("Probe() took %v, want ~500ms (range: %v-%v)", elapsed, minDuration, maxDuration)
	}
}

// TestProber_Probe_ThreeQueries_RED tests that prober sends exactly 3 queries.
//
// TDD Phase: RED
//
// RFC 6762 §8.1: "The host MUST send at least two query packets"
// Beacon implementation: Send exactly 3 queries for robust detection
//
// T024: Test prober sends 3 queries
func TestProber_Probe_ThreeQueries(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Track query count via mock transport
	queryCount := 0
	prober.onSendQuery = func() {
		queryCount++
	}

	result := prober.Probe(ctx, testServiceName)
	if result.Error != nil {
		t.Fatalf("Probe() error = %v, want nil", result.Error)
	}

	if queryCount != 3 {
		t.Errorf("Probe() sent %d queries, want 3", queryCount)
	}
}

// TestProber_Probe_ConflictDetection_RED tests conflict detection during probing.
//
// TDD Phase: RED
//
// RFC 6762 §8.1: If a conflicting response is received, probing fails
// T024: Test conflict detection during probing
func TestProber_Probe_ConflictDetection(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Simulate receiving a conflicting response after first probe
	prober.injectConflictAfter = 1 // After 1st probe

	result := prober.Probe(ctx, testServiceName)

	if !result.Conflict {
		t.Error("Probe() Conflict = false, want true (simulated conflict)")
	}

	if result.Error != nil {
		t.Errorf("Probe() error = %v, want nil (conflict is not an error)", result.Error)
	}
}

// TestProber_Probe_Cancellation_RED tests context cancellation during probing.
//
// TDD Phase: RED
//
// FR-009: All blocking operations MUST respect context cancellation
// T024: Test prober respects context cancellation
func TestProber_Probe_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	prober := NewProber()

	// Cancel context after 100ms (before probing completes)
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := prober.Probe(ctx, testServiceName)
	elapsed := time.Since(start)

	if result.Error == nil {
		t.Error("Probe() error = nil, want context.Canceled")
	}

	// Should abort quickly after cancellation (~100ms, not full 750ms)
	if elapsed > 300*time.Millisecond {
		t.Errorf("Probe() took %v after cancellation, want <300ms", elapsed)
	}
}

// TestProber_Probe_TieBreaking_RED tests lexicographic tie-breaking for simultaneous probes.
//
// TDD Phase: RED
//
// RFC 6762 §8.2.1: When simultaneous probes occur, use lexicographic comparison
// T024: Test tie-breaking with ConflictDetector integration
func TestProber_Probe_TieBreaking(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Simulate simultaneous probe where we WIN (our data > their data)
	prober.injectSimultaneousProbe = true
	prober.ourProbeData = []byte{192, 168, 1, 100}  // Our IP
	prober.theirProbeData = []byte{192, 168, 1, 50} // Their IP (lexicographically earlier)

	result := prober.Probe(ctx, testServiceName)

	// We should win the tie-break, so no conflict
	if result.Conflict {
		t.Error("Probe() Conflict = true, want false (we won tie-break)")
	}

	if result.Error != nil {
		t.Fatalf("Probe() error = %v, want nil", result.Error)
	}
}

// TestProber_Probe_TieBreaking_Lose_RED tests losing lexicographic tie-breaking.
//
// TDD Phase: RED
//
// RFC 6762 §8.2.1: If their data is lexicographically later, we lose
// T024: Test losing tie-break results in conflict
func TestProber_Probe_TieBreaking_Lose(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Simulate simultaneous probe where we LOSE (their data > our data)
	prober.injectSimultaneousProbe = true
	prober.ourProbeData = []byte{192, 168, 1, 50}    // Our IP (lexicographically earlier)
	prober.theirProbeData = []byte{192, 168, 1, 100} // Their IP

	result := prober.Probe(ctx, testServiceName)

	// We should lose the tie-break, so conflict detected
	if !result.Conflict {
		t.Error("Probe() Conflict = false, want true (we lost tie-break)")
	}
}

// TestProber_ConflictDetectorIntegration tests that Prober uses ConflictDetector
// to check incoming probe responses for conflicts.
//
// TDD Phase: RED - This test will FAIL until we integrate ConflictDetector with Prober
//
// RFC 6762 §8.1: "If any conflicting Multicast DNS response is received before
// all three of the probe queries have been sent, then the probing host knows
// that there is already a host on the network using that name."
//
// RFC 6762 §8.2: Use lexicographic comparison for simultaneous probes
//
// FR-029: System MUST detect naming conflicts via lexicographic comparison (US2)
// T059: Integrate ConflictDetector with Prober (RED phase)
func TestProber_ConflictDetectorIntegration(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Simulate receiving a probe response with resource records
	// The prober should use ConflictDetector to check if these conflict with our records
	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // Our IP (lexicographically earlier)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // Their IP (lexicographically later - they win)
	}

	// Set up prober with our records
	prober.SetOurRecords([]message.ResourceRecord{ourRecord})

	// Set up ConflictDetector (using mock)
	detector := &mockConflictDetector{}
	prober.SetConflictDetector(detector)

	// Inject incoming probe response
	prober.InjectIncomingResponse([]message.ResourceRecord{incomingRecord})

	// Probe should detect conflict via ConflictDetector
	result := prober.Probe(ctx, testServiceName)

	if !result.Conflict {
		t.Error("Probe() Conflict = false, want true (ConflictDetector should detect we lose)")
	}

	if result.Error != nil {
		t.Errorf("Probe() error = %v, want nil (conflict is not an error)", result.Error)
	}
}

// TestProber_ConflictDetectorIntegration_NoConflict tests that Prober correctly
// identifies when there is NO conflict (we win the tie-break).
//
// TDD Phase: RED
//
// RFC 6762 §8.2: If we win the lexicographic comparison, no conflict
//
// T059: Integrate ConflictDetector with Prober (RED phase)
func TestProber_ConflictDetectorIntegration_NoConflict(t *testing.T) {
	ctx := context.Background()
	prober := NewProber()

	// Our record wins the tie-break
	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // Our IP (lexicographically later - we win)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // Their IP (lexicographically earlier)
	}

	// Set up prober with our records
	prober.SetOurRecords([]message.ResourceRecord{ourRecord})

	// Set up ConflictDetector (using mock)
	detector := &mockConflictDetector{}
	prober.SetConflictDetector(detector)

	// Inject incoming probe response
	prober.InjectIncomingResponse([]message.ResourceRecord{incomingRecord})

	// Probe should NOT detect conflict (we win tie-break)
	result := prober.Probe(ctx, testServiceName)

	if result.Conflict {
		t.Error("Probe() Conflict = true, want false (we win tie-break, no conflict)")
	}

	if result.Error != nil {
		t.Errorf("Probe() error = %v, want nil", result.Error)
	}
}

// TestProber_MessageCapture verifies probe message capture for contract tests.
func TestProber_MessageCapture(t *testing.T) {
	p := NewProber()
	ctx := context.Background()

	result := p.Probe(ctx, "test.local")
	if result.Error != nil {
		t.Fatalf("Probe() error = %v", result.Error)
	}

	msg := p.GetLastProbeMessage()
	if len(msg) < 12 {
		t.Errorf("GetLastProbeMessage() = %d bytes, want ≥12 (DNS header)", len(msg))
		t.Logf("Message bytes: %v", msg)
	} else {
		t.Logf("✓ Captured probe message: %d bytes", len(msg))
	}
}

// TestProber_MessageCapture_WithSpaces tests message capture with spaces in service name.
func TestProber_MessageCapture_WithSpaces(t *testing.T) {
	p := NewProber()
	ctx := context.Background()

	// Use the same service name format as the contract test
	serviceName := "RFC Test Service._http._tcp.local"
	result := p.Probe(ctx, serviceName)
	if result.Error != nil {
		t.Fatalf("Probe() error = %v", result.Error)
	}

	msg := p.GetLastProbeMessage()
	t.Logf("Message length with spaces: %d bytes", len(msg))
	if len(msg) < 12 {
		t.Errorf("GetLastProbeMessage() = %d bytes, want ≥12 (DNS header)", len(msg))
		t.Logf("This confirms the hypothesis: BuildQuery fails with spaces in the name")
	} else {
		t.Logf("✓ Captured probe message: %d bytes", len(msg))
		t.Logf("Hypothesis rejected: BuildQuery works with spaces")
	}
}

// TestProber_BuildQuery_Error tests what error BuildQuery returns with spaces.
func TestProber_BuildQuery_Error(t *testing.T) {
	// Directly test BuildQuery with spaces
	serviceName := "RFC Test Service._http._tcp.local"
	msg, err := message.BuildQuery(serviceName, uint16(protocol.RecordTypeANY))

	if err != nil {
		t.Logf("BuildQuery error: %v", err)
		t.Logf("Error type: %T", err)
	} else {
		t.Logf("BuildQuery succeeded unexpectedly, message length: %d", len(msg))
	}
}

// TestProber_TransportSend verifies that Prober sends probe packets via transport.
//
// RFC 6762 §8.1: Probes MUST be sent to the mDNS multicast group 224.0.0.251:5353.
// This test verifies that Send() is called 3 times with the probe message bytes
// and the correct multicast destination.
func TestProber_TransportSend(t *testing.T) {
	mock := transport.NewMockTransport()
	prober := NewProber()
	prober.SetTransport(mock)

	ctx := context.Background()
	result := prober.Probe(ctx, testServiceName)
	if result.Error != nil {
		t.Fatalf("Probe() error = %v", result.Error)
	}

	calls := mock.SendCalls()
	if len(calls) != 3 {
		t.Fatalf("Send() called %d times, want 3", len(calls))
	}

	// Verify destination is mDNS multicast address
	for i, call := range calls {
		if call.Dest == nil {
			t.Errorf("Send() call %d: dest = nil, want 224.0.0.251:5353", i)
			continue
		}
		destStr := call.Dest.String()
		if destStr != "224.0.0.251:5353" {
			t.Errorf("Send() call %d: dest = %q, want %q", i, destStr, "224.0.0.251:5353")
		}
	}

	// Verify sent bytes match lastProbeMessage
	lastMsg := prober.GetLastProbeMessage()
	if lastMsg == nil {
		t.Fatal("GetLastProbeMessage() = nil")
	}
	for i, call := range calls {
		if len(call.Packet) != len(lastMsg) {
			t.Errorf("Send() call %d: packet len = %d, want %d", i, len(call.Packet), len(lastMsg))
		}
	}
}

// TestProber_ProbeMessageContainsServiceName verifies the probe message QNAME
// contains the actual service instance name, not the root domain.
//
// RFC 6762 §8.1: Probe queries MUST ask for the name being claimed.
func TestProber_ProbeMessageContainsServiceName(t *testing.T) {
	prober := NewProber()
	ctx := context.Background()

	result := prober.Probe(ctx, testServiceName)
	if result.Error != nil {
		t.Fatalf("Probe() error = %v", result.Error)
	}

	msg := prober.GetLastProbeMessage()
	if len(msg) < 12 {
		t.Fatalf("GetLastProbeMessage() too short: %d bytes", len(msg))
	}

	// Parse the message to extract the QNAME
	parsed, err := message.ParseMessage(msg)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if len(parsed.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(parsed.Questions))
	}

	q := parsed.Questions[0]

	// Verify QTYPE = ANY (255)
	if q.QTYPE != uint16(protocol.RecordTypeANY) {
		t.Errorf("QTYPE = %d, want %d (ANY)", q.QTYPE, protocol.RecordTypeANY)
	}

	// Verify QCLASS = IN (1)
	if q.QCLASS != uint16(protocol.ClassIN) {
		t.Errorf("QCLASS = %d, want %d (IN)", q.QCLASS, protocol.ClassIN)
	}

	// Verify QNAME contains the service name components, not root domain
	// testServiceName = "My Printer._http._tcp.local"
	// The parsed QNAME should contain the instance name and service type
	qname := q.QNAME
	if qname == "." || qname == "" {
		t.Errorf("QNAME = %q, want service instance name containing %q", qname, testServiceName)
	}

	// Check that the QNAME contains key parts of the service name
	if !strings.Contains(qname, "My Printer") {
		t.Errorf("QNAME %q does not contain instance name %q", qname, "My Printer")
	}
	if !strings.Contains(qname, "_http") {
		t.Errorf("QNAME %q does not contain service type %q", qname, "_http")
	}
	if !strings.Contains(qname, "_tcp") {
		t.Errorf("QNAME %q does not contain protocol %q", qname, "_tcp")
	}
}

// buildTestResponsePacket constructs a minimal DNS response packet with one A record answer.
// This is used by tests to simulate mDNS responses received during probing.
func buildTestResponsePacket(name string, ip net.IP) []byte {
	// Encode the name for the answer section
	var encodedName []byte
	if idx := strings.Index(name, "._"); idx >= 0 {
		instanceName := name[:idx]
		serviceType := name[idx+1:]
		var err error
		encodedName, err = message.EncodeServiceInstanceName(instanceName, serviceType)
		if err != nil {
			// Fallback for simple names
			encodedName, _ = message.EncodeName(name)
		}
	} else {
		encodedName, _ = message.EncodeName(name)
	}

	// DNS Header (12 bytes): QR=1 (response), ANCOUNT=1
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[2:4], 0x8400) // QR=1, AA=1
	binary.BigEndian.PutUint16(header[6:8], 1)      // ANCOUNT = 1

	// Answer section: NAME + TYPE(A=1) + CLASS(IN=1) + TTL + RDLENGTH + RDATA
	answer := make([]byte, 0, len(encodedName)+10+4)
	answer = append(answer, encodedName...)

	typeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBuf, uint16(protocol.RecordTypeA))
	answer = append(answer, typeBuf...)

	classBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(classBuf, uint16(protocol.ClassIN))
	answer = append(answer, classBuf...)

	ttlBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBuf, 120)
	answer = append(answer, ttlBuf...)

	rdlenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(rdlenBuf, 4) // A record = 4 bytes
	answer = append(answer, rdlenBuf...)

	answer = append(answer, ip.To4()...)

	return append(header, answer...)
}

// TestProber_ListensForConflictDuringProbing verifies that the prober listens for
// incoming mDNS responses during the 250ms probe interval and detects conflicts.
//
// RFC 6762 §8.1: "A host MUST listen for responses during probing"
func TestProber_ListensForConflictDuringProbing(t *testing.T) {
	mock := transport.NewMockTransport()
	prober := NewProber()
	prober.SetTransport(mock)
	prober.EnableListenForResponses()

	// Our record: A record for the service name with our IP
	ourRecord := message.ResourceRecord{
		Name:  "My Printer._http._tcp.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // Our IP (lexicographically earlier - we lose)
	}
	prober.SetOurRecords([]message.ResourceRecord{ourRecord})

	// Set up conflict detector that detects conflict when names match and their data wins
	detector := &mockConflictDetector{
		detectFunc: func(our, incoming message.ResourceRecord) (bool, error) {
			// Conflict if same name and incoming data is lexicographically greater
			if our.Name != incoming.Name {
				return false, nil
			}
			for i := 0; i < len(our.Data) && i < len(incoming.Data); i++ {
				if our.Data[i] < incoming.Data[i] {
					return true, nil // We lose - conflict
				} else if our.Data[i] > incoming.Data[i] {
					return false, nil // We win
				}
			}
			return false, nil
		},
	}
	prober.SetConflictDetector(detector)

	// Build a response packet claiming the same name with a different (winning) IP
	responsePacket := buildTestResponsePacket(
		"My Printer._http._tcp.local",
		net.IPv4(192, 168, 1, 100), // Their IP - lexicographically greater, they win
	)

	// Queue the response so the prober receives it during the first probe interval.
	// Use onSendQuery to queue the response after the first probe is sent.
	firstSend := true
	prober.SetOnSendQuery(func() {
		if firstSend {
			firstSend = false
			// Queue a conflicting response for the prober to receive during the wait
			mock.QueueReceive(responsePacket, &net.UDPAddr{
				IP:   net.IPv4(192, 168, 1, 100),
				Port: 5353,
			}, 0)
		}
	})

	ctx := context.Background()
	result := prober.Probe(ctx, testServiceName)

	if !result.Conflict {
		t.Error("Probe() Conflict = false, want true (conflict response received during probing)")
	}

	if result.Error != nil {
		t.Errorf("Probe() error = %v, want nil (conflict is not an error)", result.Error)
	}
}

// TestProber_NoConflictWhenResponseDoesntMatch verifies that a response for a
// different service name does not trigger a conflict.
func TestProber_NoConflictWhenResponseDoesntMatch(t *testing.T) {
	mock := transport.NewMockTransport()
	prober := NewProber()
	prober.SetTransport(mock)
	prober.EnableListenForResponses()

	// Our record
	ourRecord := message.ResourceRecord{
		Name:  "My Printer._http._tcp.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50},
	}
	prober.SetOurRecords([]message.ResourceRecord{ourRecord})

	// Conflict detector only flags conflict if names match
	detector := &mockConflictDetector{
		detectFunc: func(our, incoming message.ResourceRecord) (bool, error) {
			if our.Name != incoming.Name {
				return false, nil // Different names - no conflict
			}
			// Same name - check data
			for i := 0; i < len(our.Data) && i < len(incoming.Data); i++ {
				if our.Data[i] < incoming.Data[i] {
					return true, nil
				} else if our.Data[i] > incoming.Data[i] {
					return false, nil
				}
			}
			return false, nil
		},
	}
	prober.SetConflictDetector(detector)

	// Build a response for a DIFFERENT service name
	responsePacket := buildTestResponsePacket(
		"Other Service._http._tcp.local",
		net.IPv4(192, 168, 1, 100),
	)

	// Queue the non-matching response
	firstSend := true
	prober.SetOnSendQuery(func() {
		if firstSend {
			firstSend = false
			mock.QueueReceive(responsePacket, &net.UDPAddr{
				IP:   net.IPv4(192, 168, 1, 100),
				Port: 5353,
			}, 0)
		}
	})

	ctx := context.Background()
	result := prober.Probe(ctx, testServiceName)

	if result.Conflict {
		t.Error("Probe() Conflict = true, want false (response was for different service name)")
	}

	if result.Error != nil {
		t.Errorf("Probe() error = %v, want nil", result.Error)
	}
}
