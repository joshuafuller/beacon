package state

import (
	"context"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
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

// TestCompareBytesLexicographically tests lexicographic byte comparison.
//
// RFC 6762 §8.2.1: Tie-breaking uses lexicographic comparison of probe records.
// This is critical for simultaneous probe conflict resolution.
//
// Coverage improvement: compareBytesLexicographically (77.8% → target 100%)
func TestCompareBytesLexicographically(t *testing.T) {
	tests := []struct {
		name    string
		a       []byte
		b       []byte
		wantWin bool // true if a > b (we win)
	}{
		{
			name:    "a wins - first byte greater",
			a:       []byte{200, 1, 1, 1},
			b:       []byte{100, 1, 1, 1},
			wantWin: true,
		},
		{
			name:    "b wins - first byte greater",
			a:       []byte{100, 1, 1, 1},
			b:       []byte{200, 1, 1, 1},
			wantWin: false,
		},
		{
			name:    "a wins - second byte decides",
			a:       []byte{169, 254, 200, 50},
			b:       []byte{169, 254, 99, 200},
			wantWin: true, // 200 > 99 in third byte
		},
		{
			name:    "b wins - third byte decides",
			a:       []byte{169, 254, 1, 255},
			b:       []byte{169, 254, 2, 1},
			wantWin: false, // 1 < 2 in third byte
		},
		{
			name:    "a wins - longer when prefixes match",
			a:       []byte{192, 168, 1, 100, 5},
			b:       []byte{192, 168, 1, 100},
			wantWin: true, // len(a)=5 > len(b)=4
		},
		{
			name:    "b wins - longer when prefixes match",
			a:       []byte{10, 0, 0},
			b:       []byte{10, 0, 0, 1},
			wantWin: false, // len(a)=3 < len(b)=4
		},
		{
			name:    "identical - no winner",
			a:       []byte{192, 168, 1, 1},
			b:       []byte{192, 168, 1, 1},
			wantWin: false, // equal length, equal bytes
		},
		{
			name:    "empty slices - no winner",
			a:       []byte{},
			b:       []byte{},
			wantWin: false, // both empty
		},
		{
			name:    "a has data, b empty - a wins",
			a:       []byte{1},
			b:       []byte{},
			wantWin: true, // len(a)=1 > len(b)=0
		},
		{
			name:    "a empty, b has data - b wins",
			a:       []byte{},
			b:       []byte{1},
			wantWin: false, // len(a)=0 < len(b)=1
		},
		{
			name:    "single byte - a wins",
			a:       []byte{255},
			b:       []byte{0},
			wantWin: true, // 255 > 0
		},
		{
			name:    "single byte - b wins",
			a:       []byte{0},
			b:       []byte{255},
			wantWin: false, // 0 < 255
		},
		{
			name:    "max values - a wins by length",
			a:       []byte{255, 255, 255, 255, 1},
			b:       []byte{255, 255, 255, 255},
			wantWin: true, // all bytes match, a is longer
		},
		{
			name:    "ip addresses - a wins",
			a:       []byte{192, 168, 2, 1},  // 192.168.2.1
			b:       []byte{192, 168, 1, 255}, // 192.168.1.255
			wantWin: true,                     // 2 > 1 in third octet
		},
		{
			name:    "last byte decides",
			a:       []byte{1, 2, 3, 5},
			b:       []byte{1, 2, 3, 4},
			wantWin: true, // 5 > 4 in last byte
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareBytesLexicographically(tt.a, tt.b)
			if got != tt.wantWin {
				t.Errorf("compareBytesLexicographically(%v, %v) = %v, want %v",
					tt.a, tt.b, got, tt.wantWin)
			}
		})
	}
}
