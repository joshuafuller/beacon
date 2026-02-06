package state

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/transport"
)

// TestAnnouncer_Announce_RED tests announcing per RFC 6762 §8.3.
//
// TDD Phase: RED - These tests will FAIL until we implement Announcer
//
// RFC 6762 §8.3: Announcing
//   - Send 2 unsolicited multicast announcements
//   - 1 second interval between announcements
//   - Total duration: ~1 second (0ms, 1000ms)
//
// FR-028: System MUST send announcements after successful probing
// T025: Write announcer tests
func TestAnnouncer_Announce(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	ctx := context.Background()
	announcer := NewAnnouncer()

	records := []byte{} // Empty for now, will be ResourceRecords in GREEN phase

	start := time.Now()
	err := announcer.Announce(ctx, testServiceName, records)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Announce() error = %v, want nil", err)
	}

	// Announcing should take ~1 second (2 announcements × 1s interval)
	// Allow ±200ms tolerance for test timing
	minDuration := 800 * time.Millisecond  // 1s - 200ms
	maxDuration := 1200 * time.Millisecond // 1s + 200ms

	if elapsed < minDuration || elapsed > maxDuration {
		t.Errorf("Announce() took %v, want ~1s (range: %v-%v)", elapsed, minDuration, maxDuration)
	}
}

// TestAnnouncer_Announce_TwoAnnouncements_RED tests that announcer sends exactly 2 announcements.
//
// TDD Phase: RED
//
// RFC 6762 §8.3: "The Multicast DNS responder MUST send at least two unsolicited
// responses, one second apart."
//
// T025: Test announcer sends 2 announcements
func TestAnnouncer_Announce_TwoAnnouncements(t *testing.T) {
	ctx := context.Background()
	announcer := NewAnnouncer()

	records := []byte{}

	// Track announcement count via mock transport
	announcementCount := 0
	announcer.onSendAnnouncement = func() {
		announcementCount++
	}

	err := announcer.Announce(ctx, testServiceName, records)
	if err != nil {
		t.Fatalf("Announce() error = %v, want nil", err)
	}

	if announcementCount != 2 {
		t.Errorf("Announce() sent %d announcements, want 2", announcementCount)
	}
}

// TestAnnouncer_Announce_Cancellation_RED tests context cancellation during announcing.
//
// TDD Phase: RED
//
// FR-009: All blocking operations MUST respect context cancellation
// T025: Test announcer respects context cancellation
func TestAnnouncer_Announce_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	announcer := NewAnnouncer()

	records := []byte{}

	// Cancel context after 500ms (before announcing completes at ~1s)
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := announcer.Announce(ctx, testServiceName, records)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Announce() error = nil, want context.Canceled")
	}

	// Should abort quickly after cancellation (~500ms, not full 1s)
	if elapsed > 700*time.Millisecond {
		t.Errorf("Announce() took %v after cancellation, want <700ms", elapsed)
	}
}

// TestAnnouncer_Announce_RecordsIncluded_RED tests that announcements include resource records.
//
// TDD Phase: RED
//
// RFC 6762 §8.3: Announcements MUST include PTR, SRV, TXT, A records
// T025: Test announcer includes all required records
func TestAnnouncer_Announce_RecordsIncluded(t *testing.T) {
	ctx := context.Background()
	announcer := NewAnnouncer()

	records := []byte{0x01, 0x02, 0x03} // Placeholder record data

	// Capture sent announcement data
	var sentData []byte
	announcer.onSendAnnouncement = func() {
		sentData = announcer.lastSentData
	}

	err := announcer.Announce(ctx, testServiceName, records)
	if err != nil {
		t.Fatalf("Announce() error = %v, want nil", err)
	}

	// Verify records were included in announcement
	if len(sentData) == 0 {
		t.Error("Announce() sent empty data, want records included")
	}
}

// TestAnnouncer_Announce_MulticastAddress_RED tests that announcements are sent to multicast address.
//
// TDD Phase: RED
//
// RFC 6762 §5: Announcements MUST be sent to 224.0.0.251:5353
// T025: Test announcer sends to correct multicast address
func TestAnnouncer_Announce_MulticastAddress(t *testing.T) {
	ctx := context.Background()
	announcer := NewAnnouncer()

	records := []byte{}

	// Capture destination address
	var destAddr string
	announcer.onSendAnnouncement = func() {
		destAddr = announcer.lastDestAddr
	}

	err := announcer.Announce(ctx, testServiceName, records)
	if err != nil {
		t.Fatalf("Announce() error = %v, want nil", err)
	}

	wantAddr := "224.0.0.251:5353"
	if destAddr != wantAddr {
		t.Errorf("Announce() sent to %q, want %q", destAddr, wantAddr)
	}
}

// TestAnnouncer_TransportSend verifies that Announcer sends packets via transport.
//
// RFC 6762 §8.3: Announcements are DNS responses (QR=1, AA=1) sent to
// 224.0.0.251:5353. This test verifies Send() is called 2 times with valid
// DNS response packets and the correct multicast destination.
func TestAnnouncer_TransportSend(t *testing.T) {
	mock := transport.NewMockTransport()
	announcer := NewAnnouncer()
	announcer.SetTransport(mock)

	ctx := context.Background()
	err := announcer.Announce(ctx, testServiceName, []byte{})
	if err != nil {
		t.Fatalf("Announce() error = %v", err)
	}

	calls := mock.SendCalls()
	if len(calls) != 2 {
		t.Fatalf("Send() called %d times, want 2", len(calls))
	}

	for i, call := range calls {
		// Verify destination
		if call.Dest == nil {
			t.Errorf("Send() call %d: dest = nil, want 224.0.0.251:5353", i)
			continue
		}
		if call.Dest.String() != "224.0.0.251:5353" {
			t.Errorf("Send() call %d: dest = %q, want %q", i, call.Dest.String(), "224.0.0.251:5353")
		}

		// Verify packet is a valid DNS response (QR=1, AA=1)
		if len(call.Packet) < 12 {
			t.Errorf("Send() call %d: packet too short (%d bytes), want >= 12", i, len(call.Packet))
			continue
		}
		flags := binary.BigEndian.Uint16(call.Packet[2:4])
		if flags&0x8000 == 0 {
			t.Errorf("Send() call %d: QR bit not set (flags=0x%04x)", i, flags)
		}
		if flags&0x0400 == 0 {
			t.Errorf("Send() call %d: AA bit not set (flags=0x%04x)", i, flags)
		}
	}
}
