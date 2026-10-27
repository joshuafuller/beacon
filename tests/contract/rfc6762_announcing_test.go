package contract

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/responder"
)

// testRFC6762Timing verifies RFC 6762 timing compliance for probing or announcing intervals.
//
// Parameters:
//   - t: Test context
//   - trackTimes: Callback that starts responder and returns captured timestamps
//   - expectedInterval: Expected interval between messages (250ms for probing, 1s for announcing)
//   - tolerance: Timing tolerance (e.g., 50ms for probing, 200ms for announcing)
//   - rfcSection: RFC reference for error messages (e.g., "RFC 6762 §8.1")
func testRFC6762Timing(t *testing.T, trackTimes func(*responder.Responder) []time.Time,
	expectedInterval time.Duration, tolerance time.Duration, rfcSection string) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	ctx := context.Background()
	r, err := responder.New(ctx, responder.WithHostname(uniqueHostname()))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	times := trackTimes(r)
	if len(times) < 2 {
		t.Fatalf("Expected at least 2 timestamps, got %d", len(times))
	}

	// Verify intervals match expected timing
	minInterval := expectedInterval - tolerance
	maxInterval := expectedInterval + tolerance

	for i := 1; i < len(times); i++ {
		interval := times[i].Sub(times[i-1])
		if interval < minInterval || interval > maxInterval {
			t.Errorf("interval[%d] = %v, want ~%v (range: %v-%v) per %s",
				i, interval, expectedInterval, minInterval, maxInterval, rfcSection)
		}
	}
}

// TestRFC6762_Announcing_TwoAnnouncements_RED tests RFC 6762 §8.3 announcing compliance.
//
// TDD Phase: RED - These tests will FAIL until announcing is implemented
//
// RFC 6762 §8.3: Announcing
//   - "The Multicast DNS responder MUST send at least two unsolicited
//     responses, one second apart."
//
// T029: Contract test for RFC 6762 §8.3 compliance
func TestRFC6762_Announcing_TwoAnnouncements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RFC contract test in short mode")
	}

	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Track announcement count
	var announcementCount int
	r.OnAnnounce(func() {
		announcementCount++
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// RFC 6762 §8.3 requires at least 2 announcements
	if announcementCount < 2 {
		t.Errorf("announcementCount = %d, want ≥2 per RFC 6762 §8.3", announcementCount)
	}
}

// TestRFC6762_Announcing_1sInterval_RED tests announcement timing per RFC 6762 §8.3.
//
// TDD Phase: RED
//
// RFC 6762 §8.3: "one second apart"
// T029: Test announcement timing compliance
func TestRFC6762_Announcing_1sInterval(t *testing.T) {
	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Use common timing test helper with 1s interval and 200ms tolerance
	testRFC6762Timing(t,
		func(r *responder.Responder) []time.Time {
			var announceTimes []time.Time
			r.OnAnnounce(func() {
				announceTimes = append(announceTimes, time.Now())
			})

			err := r.Register(service)
			if err != nil {
				t.Fatalf("Register() error = %v, want nil", err)
			}

			return announceTimes
		},
		1*time.Second, 200*time.Millisecond, "RFC 6762 §8.3")
}

// TestRFC6762_Announcing_ResponseFormat_RED tests announcement message format per RFC 6762 §8.3.
//
// TDD Phase: RED
//
// RFC 6762 §8.3: Announcements are responses (QR=1, AA=1)
//   - QR bit = 1 (response)
//   - AA bit = 1 (authoritative)
//   - Answer section contains service records
//
// T029: Test announcement message format compliance
func TestRFC6762_Announcing_ResponseFormat(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture announcement message bytes
	var announceMessage []byte
	r.OnAnnounce(func() {
		announceMessage = r.GetLastAnnounceMessage()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if len(announceMessage) < 12 {
		t.Fatalf("announcement message too short: %d bytes, want ≥12 (DNS header)", len(announceMessage))
	}

	// Parse DNS header flags (bytes 2-3)
	flags := binary.BigEndian.Uint16(announceMessage[2:4])

	// QR bit (bit 15) MUST be 1 for responses per RFC 6762 §18.2
	qr := (flags >> 15) & 0x01
	if qr != 1 {
		t.Errorf("announcement QR bit = %d, want 1 (response per RFC 6762 §18.2)", qr)
	}

	// AA bit (bit 10) MUST be 1 for authoritative answers per RFC 6762 §18.4
	aa := (flags >> 10) & 0x01
	if aa != 1 {
		t.Errorf("announcement AA bit = %d, want 1 (authoritative per RFC 6762 §18.4)", aa)
	}

	// OPCODE (bits 11-14) MUST be 0 per RFC 6762 §18.3
	opcode := (flags >> 11) & 0x0F
	if opcode != 0 {
		t.Errorf("announcement OPCODE = %d, want 0 (standard query per RFC 6762 §18.3)", opcode)
	}

	// RCODE (bits 0-3) MUST be 0 per RFC 6762 §18.11
	rcode := flags & 0x0F
	if rcode != 0 {
		t.Errorf("announcement RCODE = %d, want 0 (no error per RFC 6762 §18.11)", rcode)
	}
}

// TestRFC6762_Announcing_AnswerSection_RED tests announcement answer section per RFC 6762 §8.3.
//
// TDD Phase: RED
//
// RFC 6762 §8.3: Announcements include answer records
// RFC 6763 §6: Service announcements include PTR, SRV, TXT, A records
//
// T029: Test announcement includes all required records
func TestRFC6762_Announcing_AnswerSection(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture announcement message bytes
	var announceMessage []byte
	r.OnAnnounce(func() {
		announceMessage = r.GetLastAnnounceMessage()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if len(announceMessage) < 12 {
		t.Fatalf("announcement message too short: %d bytes", len(announceMessage))
	}

	// Parse ANCOUNT (bytes 6-7) - number of answer records
	ancount := binary.BigEndian.Uint16(announceMessage[6:8])

	// Should have at least 4 answers: PTR, SRV, TXT, A per RFC 6763 §6
	if ancount < 4 {
		t.Errorf("announcement ANCOUNT = %d, want ≥4 (PTR, SRV, TXT, A per RFC 6763 §6)", ancount)
	}

	// Message should be longer than just header if it has answers
	if len(announceMessage) <= 12 {
		t.Errorf("announcement size = %d bytes, want >12 (should include answer section)", len(announceMessage))
	}
}

// TestRFC6762_Announcing_CacheFlushBit_RED tests cache-flush bit per RFC 6762 §10.2.
//
// TDD Phase: RED
//
// RFC 6762 §10.2: "The cache-flush bit SHOULD be set on unique resource records"
//   - SRV, TXT, A records are unique → cache-flush = 1
//   - PTR records are shared → cache-flush = 0
//
// T029: Test cache-flush bit compliance
func TestRFC6762_Announcing_CacheFlushBit(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture record set sent in announcement
	var recordSet []*responder.ResourceRecord
	r.OnAnnounce(func() {
		recordSet = r.GetLastAnnouncedRecords()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Check cache-flush bit for each record type
	for _, record := range recordSet {
		switch record.Type {
		case protocol.RecordTypePTR:
			// PTR is shared, cache-flush SHOULD be 0
			if record.CacheFlush {
				t.Errorf("PTR record has cache-flush=true, want false (shared record per RFC 6762 §10.2)")
			}
		case protocol.RecordTypeSRV, protocol.RecordTypeTXT, protocol.RecordTypeA:
			// SRV, TXT, A are unique, cache-flush SHOULD be 1
			if !record.CacheFlush {
				t.Errorf("%v record has cache-flush=false, want true (unique record per RFC 6762 §10.2)", record.Type)
			}
		}
	}
}

// TestRFC6762_Announcing_TTL_RED tests TTL values per RFC 6762 §10.
//
// TDD Phase: RED
//
// RFC 6762 §10: TTL values
//   - Service records (PTR, SRV, TXT): 120 seconds
//   - Hostname records (A, AAAA): 4500 seconds (75 minutes)
//
// T029: Test TTL compliance
func TestRFC6762_Announcing_TTL(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture record set sent in announcement
	var recordSet []*responder.ResourceRecord
	r.OnAnnounce(func() {
		recordSet = r.GetLastAnnouncedRecords()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Check TTL for each record type
	for _, record := range recordSet {
		switch record.Type {
		case protocol.RecordTypePTR, protocol.RecordTypeSRV, protocol.RecordTypeTXT:
			// Service records: 120 seconds per RFC 6762 §10
			wantTTL := uint32(120)
			if record.TTL != wantTTL {
				t.Errorf("%v record TTL = %d, want %d (service TTL per RFC 6762 §10)", record.Type, record.TTL, wantTTL)
			}
		case protocol.RecordTypeA:
			// Hostname records: 4500 seconds per RFC 6762 §10
			wantTTL := uint32(4500)
			if record.TTL != wantTTL {
				t.Errorf("%v record TTL = %d, want %d (hostname TTL per RFC 6762 §10)", record.Type, record.TTL, wantTTL)
			}
		}
	}
}

// TestRFC6762_Announcing_MulticastAddress_RED tests multicast destination per RFC 6762 §5.
//
// TDD Phase: RED
//
// RFC 6762 §5: Multicast DNS Messages
//   - IPv4: 224.0.0.251, port 5353
//   - IPv6: FF02::FB, port 5353
//
// T029: Test announcements sent to correct multicast address
func TestRFC6762_Announcing_MulticastAddress(t *testing.T) {
	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() { _ = r.Close() }()

	service := &responder.Service{
		InstanceName: "RFC Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Capture destination address
	var destAddr string
	r.OnAnnounce(func() {
		destAddr = r.GetLastAnnounceDest()
	})

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// RFC 6762 §5: IPv4 multicast address
	wantAddr := "224.0.0.251:5353"
	if destAddr != wantAddr {
		t.Errorf("announcement dest = %q, want %q (RFC 6762 §5)", destAddr, wantAddr)
	}
}
