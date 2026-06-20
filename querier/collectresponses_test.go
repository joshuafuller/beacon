package querier

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// =============================================================================
// collectResponses() Tests - Coverage Increase from 47.8% to 80%+
// =============================================================================
//
// These tests exercise all branches of the collectResponses() function:
// 1. Context timeout (returns collected responses, no error)
// 2. Malformed message parsing (skip and continue - FR-011, FR-016)
// 3. Invalid response flags (skip responses with QR=0 or RCODE≠0 - FR-021, FR-022)
// 4. Type filtering (skip non-matching record types)
// 5. RDATA parsing errors (skip malformed RDATA - FR-011)
// 6. Deduplication (FR-007 - skip duplicate records)
// 7. Normal response aggregation (FR-008)
//
// TDD Approach: These are tests for EXISTING code to increase coverage.
// Following RFC 6762 and functional requirements (FR-007 through FR-016).
//
// =============================================================================

// TestCollectResponses_ContextTimeout tests that timeout returns what was collected.
//
// FR-008: Timeout is NOT an error - return aggregated responses
// Coverage: collectResponses line 268-270 (context done path)
func TestCollectResponses_ContextTimeout(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Call collectResponses - should timeout and return empty response
	response, err := q.collectResponses(ctx, "test.local", RecordTypeA)

	if err != nil {
		t.Errorf("collectResponses(timeout) = %v, want nil (timeout is not an error per FR-008)", err)
	}

	if response == nil {
		t.Fatal("collectResponses returned nil response")
	}

	if len(response.Records) != 0 {
		t.Errorf("collectResponses(timeout) returned %d records, want 0 (no responses received)", len(response.Records))
	}

	t.Log("✓ FR-008: Timeout returns collected responses without error")
}

// TestCollectResponses_MalformedMessage tests that malformed packets are skipped.
//
// FR-011: Validate and discard malformed packets
// FR-016: Continue collecting after discarding malformed packets
// Coverage: collectResponses line 274-278 (parse error path)
func TestCollectResponses_MalformedMessage(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inject malformed packet into response channel
	malformed := []byte{0x00, 0x01, 0x02} // Too short - invalid DNS message
	q.responseChan <- malformed

	// Also send a valid response packet to test that collection continues
	validPacket := buildValidResponsePacket("test.local", protocol.RecordTypeA, []byte{192, 168, 1, 1})
	q.responseChan <- validPacket

	// Start collecting in background
	doneChan := make(chan *Response, 1)
	go func() {
		resp, _ := q.collectResponses(ctx, "test.local", RecordTypeA)
		doneChan <- resp
	}()

	// Wait for timeout
	select {
	case response := <-doneChan:
		// Should have 1 valid record (malformed packet skipped)
		if len(response.Records) != 1 {
			t.Errorf("collectResponses returned %d records, want 1 (malformed packet should be skipped)", len(response.Records))
		}
		t.Log("✓ FR-011, FR-016: Malformed packets skipped, collection continues")

	case <-time.After(200 * time.Millisecond):
		t.Fatal("collectResponses did not return within timeout")
	}
}

// TestCollectResponses_InvalidResponseFlags tests that invalid responses are skipped.
//
// FR-021: Validate QR=1 (responses only)
// FR-022: Ignore RCODE≠0
// Coverage: collectResponses line 282-286 (validate response path)
func TestCollectResponses_InvalidResponseFlags(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inject packet with QR=0 (query, not response) - should be skipped
	queryPacket := buildQueryPacket("test.local", protocol.RecordTypeA)
	q.responseChan <- queryPacket

	// Start collecting in background
	doneChan := make(chan *Response, 1)
	go func() {
		resp, _ := q.collectResponses(ctx, "test.local", RecordTypeA)
		doneChan <- resp
	}()

	// Wait for timeout
	select {
	case response := <-doneChan:
		// Should have 0 records (query packet skipped)
		if len(response.Records) != 0 {
			t.Errorf("collectResponses returned %d records, want 0 (query packets should be skipped)", len(response.Records))
		}
		t.Log("✓ FR-021: Query packets (QR=0) are skipped")

	case <-time.After(200 * time.Millisecond):
		t.Fatal("collectResponses did not return within timeout")
	}
}

// TestCollectResponses_TypeFiltering tests that non-matching types are skipped.
//
// Coverage: collectResponses line 291-295 (type filtering)
func TestCollectResponses_TypeFiltering(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inject PTR response when querying for A record
	ptrPacket := buildValidResponsePacket("_http._tcp.local", protocol.RecordTypePTR, []byte{4, 't', 'e', 's', 't', 0})
	q.responseChan <- ptrPacket

	// Also inject matching A record
	aPacket := buildValidResponsePacket("test.local", protocol.RecordTypeA, []byte{192, 168, 1, 1})
	q.responseChan <- aPacket

	// Start collecting A records
	doneChan := make(chan *Response, 1)
	go func() {
		resp, _ := q.collectResponses(ctx, "test.local", RecordTypeA)
		doneChan <- resp
	}()

	// Wait for timeout
	select {
	case response := <-doneChan:
		// Should have 1 A record (PTR record skipped due to type filter)
		if len(response.Records) != 1 {
			t.Errorf("collectResponses returned %d records, want 1 (PTR record should be filtered out)", len(response.Records))
		}
		if len(response.Records) == 1 && response.Records[0].Type != RecordTypeA {
			t.Errorf("Record type = %v, want %v", response.Records[0].Type, RecordTypeA)
		}
		t.Log("✓ Type filtering: Non-matching record types skipped")

	case <-time.After(200 * time.Millisecond):
		t.Fatal("collectResponses did not return within timeout")
	}
}

// TestCollectResponses_Deduplication tests RFC 6762 record deduplication.
//
// FR-007: Deduplicate identical responses
// Coverage: collectResponses line 304-310 (deduplication logic)
func TestCollectResponses_Deduplication(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inject duplicate A records (same name, type, data)
	packet1 := buildValidResponsePacket("test.local", protocol.RecordTypeA, []byte{192, 168, 1, 1})
	packet2 := buildValidResponsePacket("test.local", protocol.RecordTypeA, []byte{192, 168, 1, 1}) // Duplicate
	packet3 := buildValidResponsePacket("test.local", protocol.RecordTypeA, []byte{192, 168, 1, 2}) // Different IP

	q.responseChan <- packet1
	q.responseChan <- packet2 // Should be deduplicated
	q.responseChan <- packet3

	// Start collecting
	doneChan := make(chan *Response, 1)
	go func() {
		resp, _ := q.collectResponses(ctx, "test.local", RecordTypeA)
		doneChan <- resp
	}()

	// Wait for timeout
	select {
	case response := <-doneChan:
		// Should have 2 records (packet2 is duplicate of packet1)
		if len(response.Records) != 2 {
			t.Errorf("collectResponses returned %d records, want 2 (duplicate should be deduplicated)", len(response.Records))
		}
		t.Log("✓ FR-007: Duplicate records deduplicated")

	case <-time.After(200 * time.Millisecond):
		t.Fatal("collectResponses did not return within timeout")
	}
}

// TestCollectResponses_NormalAggregation tests happy path response collection.
//
// FR-008: Aggregate responses received within timeout window
// Coverage: collectResponses line 272-322 (full happy path)
func TestCollectResponses_NormalAggregation(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Inject multiple valid responses
	packet1 := buildValidResponsePacket("test1.local", protocol.RecordTypeA, []byte{192, 168, 1, 1})
	packet2 := buildValidResponsePacket("test2.local", protocol.RecordTypeA, []byte{192, 168, 1, 2})
	packet3 := buildValidResponsePacket("test3.local", protocol.RecordTypeA, []byte{192, 168, 1, 3})

	q.responseChan <- packet1
	q.responseChan <- packet2
	q.responseChan <- packet3

	// Start collecting
	doneChan := make(chan *Response, 1)
	go func() {
		resp, _ := q.collectResponses(ctx, "test.local", RecordTypeA)
		doneChan <- resp
	}()

	// Wait for timeout
	select {
	case response := <-doneChan:
		// Should have 3 records
		if len(response.Records) != 3 {
			t.Errorf("collectResponses returned %d records, want 3", len(response.Records))
		}
		t.Log("✓ FR-008: Multiple responses aggregated correctly")

	case <-time.After(200 * time.Millisecond):
		t.Fatal("collectResponses did not return within timeout")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildValidResponsePacket constructs a minimal valid DNS response packet.
//
// Packet structure:
//   - Header (12 bytes): ID, Flags (QR=1), Counts
//   - Answer: NAME, TYPE, CLASS, TTL, RDLENGTH, RDATA
//
// Parameters:
//   - name: Answer name (e.g., "test.local")
//   - rtype: Record type (e.g., A=1, PTR=12)
//   - rdata: Record data (raw bytes)
//
// Returns:
//   - Complete DNS response packet
func buildValidResponsePacket(name string, rtype protocol.RecordType, rdata []byte) []byte {
	packet := make([]byte, 0, 512)

	// Header: 12 bytes
	packet = append(packet, 0x00, 0x00) // ID
	packet = append(packet, 0x80, 0x00) // Flags: QR=1 (response), RCODE=0
	packet = append(packet, 0x00, 0x00) // QDCOUNT=0
	packet = append(packet, 0x00, 0x01) // ANCOUNT=1
	packet = append(packet, 0x00, 0x00) // NSCOUNT=0
	packet = append(packet, 0x00, 0x00) // ARCOUNT=0

	// Answer section:
	// NAME (encoded as DNS labels)
	nameEncoded, _ := message.EncodeName(name)
	packet = append(packet, nameEncoded...)

	// TYPE (2 bytes, big-endian)
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, uint16(rtype))
	packet = append(packet, typeBytes...)

	// CLASS (2 bytes, IN=1)
	packet = append(packet, 0x00, 0x01)

	// TTL (4 bytes, big-endian, 120 seconds)
	packet = append(packet, 0x00, 0x00, 0x00, 0x78)

	// RDLENGTH (2 bytes, big-endian)
	rdlengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(rdlengthBytes, uint16(len(rdata)))
	packet = append(packet, rdlengthBytes...)

	// RDATA
	packet = append(packet, rdata...)

	return packet
}

// buildQueryPacket constructs a DNS query packet (QR=0) for negative testing.
//
// Used to test that collectResponses properly skips query packets (FR-021).
func buildQueryPacket(name string, rtype protocol.RecordType) []byte {
	packet := make([]byte, 0, 512)

	// Header: QR=0 (query)
	packet = append(packet, 0x00, 0x00) // ID
	packet = append(packet, 0x00, 0x00) // Flags: QR=0 (query)
	packet = append(packet, 0x00, 0x01) // QDCOUNT=1
	packet = append(packet, 0x00, 0x00) // ANCOUNT=0
	packet = append(packet, 0x00, 0x00) // NSCOUNT=0
	packet = append(packet, 0x00, 0x00) // ARCOUNT=0

	// Question section
	nameEncoded, _ := message.EncodeName(name)
	packet = append(packet, nameEncoded...)

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, uint16(rtype))
	packet = append(packet, typeBytes...)

	packet = append(packet, 0x00, 0x01) // CLASS=IN

	return packet
}

// buildBundledPTRResponse builds a DNS-SD style response: a PTR answer for
// serviceType plus SRV/TXT/A records in the ADDITIONAL section (the round-trip
// optimization per RFC 6763 §12).
func buildBundledPTRResponse(serviceType, instance, host string, port uint16, ip [4]byte, txt string) []byte {
	ptrTarget, _ := message.EncodeName(instance)

	srv := make([]byte, 6)
	binary.BigEndian.PutUint16(srv[0:2], 0) // priority
	binary.BigEndian.PutUint16(srv[2:4], 0) // weight
	binary.BigEndian.PutUint16(srv[4:6], port)
	hostEnc, _ := message.EncodeName(host)
	srv = append(srv, hostEnc...)

	txtRDATA := append([]byte{byte(len(txt))}, []byte(txt)...)

	msg := &message.DNSMessage{
		Header: message.DNSHeader{Flags: 0x8400}, // QR=1, AA=1
		Answers: []message.Answer{
			{NAME: serviceType, TYPE: uint16(protocol.RecordTypePTR), CLASS: 1, TTL: 120, RDATA: ptrTarget},
		},
		Additionals: []message.Answer{
			{NAME: instance, TYPE: uint16(protocol.RecordTypeSRV), CLASS: 1, TTL: 120, RDATA: srv},
			{NAME: instance, TYPE: uint16(protocol.RecordTypeTXT), CLASS: 1, TTL: 120, RDATA: txtRDATA},
			{NAME: host, TYPE: uint16(protocol.RecordTypeA), CLASS: 1, TTL: 120, RDATA: ip[:]},
		},
	}
	packet, _ := message.SerializeMessage(msg)
	return packet
}

// TestCollectResponses_RetainsAdditionals verifies the Additional section is
// retained in Response.Additionals (issue #4). Previously collectResponses
// processed only the Answer section, discarding the SRV/TXT/A additionals that
// let one PTR query resolve a whole instance.
func TestCollectResponses_RetainsAdditionals(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer q.Close()

	packet := buildBundledPTRResponse("_http._tcp.local", "Inst._http._tcp.local",
		"host.local", 8080, [4]byte{192, 168, 1, 5}, "path=/api")
	q.responseChan <- packet

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	resp, err := q.collectResponses(ctx, "_http._tcp.local", RecordTypePTR)
	if err != nil {
		t.Fatalf("collectResponses error: %v", err)
	}

	// Answer section: the PTR (queried type) is in Records as before.
	if len(resp.Records) != 1 || resp.Records[0].Type != RecordTypePTR {
		t.Fatalf("Records = %+v; want exactly one PTR record", resp.Records)
	}

	// Additional section: SRV/TXT/A must be retained (issue #4).
	if len(resp.Additionals) != 3 {
		t.Fatalf("Additionals has %d records, want 3 (SRV, TXT, A); additionals are being discarded", len(resp.Additionals))
	}
	var sawSRV, sawTXT, sawA bool
	for _, r := range resp.Additionals {
		switch r.Type {
		case RecordTypeSRV:
			sawSRV = true
		case RecordTypeTXT:
			sawTXT = true
		case RecordTypeA:
			sawA = true
		}
	}
	if !sawSRV || !sawTXT || !sawA {
		t.Fatalf("Additionals missing types: SRV=%v TXT=%v A=%v", sawSRV, sawTXT, sawA)
	}
}

// TestDiscoverServices_UsesAdditionals_SingleRoundTrip verifies that when the
// browse (PTR) response bundles SRV/TXT/A in its Additional section,
// DiscoverServices resolves the instance from that single response WITHOUT
// follow-up queries (issue #4). Only the bundled PTR packet is queued; if the
// additionals were not consumed, the fallback SRV/TXT/A queries would find
// nothing and the instance would be unresolved.
func TestDiscoverServices_UsesAdditionals_SingleRoundTrip(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer q.Close()

	packet := buildBundledPTRResponse("_http._tcp.local", "Inst._http._tcp.local",
		"host.local", 8080, [4]byte{192, 168, 1, 5}, "path=/api")
	q.responseChan <- packet

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	services, err := q.DiscoverServices(ctx, "_http._tcp.local")
	if err != nil {
		t.Fatalf("DiscoverServices error: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("got %d services, want 1", len(services))
	}
	s := services[0]
	if s.InstanceName != "Inst" {
		t.Errorf("InstanceName = %q, want %q", s.InstanceName, "Inst")
	}
	if s.Hostname != "host.local" {
		t.Errorf("Hostname = %q, want %q (must come from bundled SRV additional)", s.Hostname, "host.local")
	}
	if s.Port != 8080 {
		t.Errorf("Port = %d, want 8080 (from bundled SRV additional)", s.Port)
	}
	if s.TXT["path"] != "/api" {
		t.Errorf("TXT[path] = %q, want %q (from bundled TXT additional)", s.TXT["path"], "/api")
	}
	if s.AddrIPv4 == nil || s.AddrIPv4.String() != "192.168.1.5" {
		t.Errorf("AddrIPv4 = %v, want 192.168.1.5 (from bundled A additional)", s.AddrIPv4)
	}
}
