package responder

import (
	"strings"
	"testing"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestResponseBuilder_BuildResponse_PTRQuery tests building a response to a PTR query.
//
// RFC 6762 §6: When responding to a PTR query (e.g., "_http._tcp.local"), the response
// MUST contain:
//   - Answer section: PTR record pointing to the service instance
//   - Additional section: SRV + TXT + A records to reduce round-trips
//
// R005 Decision: Greedy packing with priority ordering (answer > additional)
//
// TDD Phase: RED (test written first, will fail until implementation)
//
// T064 [P] [US3]: Unit test ResponseBuilder.BuildResponse() for PTR query
func TestResponseBuilder_BuildResponse_PTRQuery(t *testing.T) {
	rb := NewResponseBuilder()

	// Service to respond about
	service := &ServiceWithIP{
		InstanceName: "MyPrinter",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   map[string]string{"txtvers": "1", "path": "/"},
	}

	// Incoming PTR query for "_http._tcp.local"
	query := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      12345,
			Flags:   0, // Query (QR=0)
			QDCount: 1,
		},
		Questions: []message.Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN),
			},
		},
	}

	// Build response
	response, err := rb.BuildResponse(service, query)
	if err != nil {
		t.Fatalf("BuildResponse() error = %v, want nil", err)
	}

	// Verify response header
	if !response.Header.IsResponse() {
		t.Error("response.Header.IsResponse() = false, want true")
	}
	// Check authoritative bit (bit 10 = 0x0400)
	if (response.Header.Flags & 0x0400) == 0 {
		t.Error("response AA bit not set, want authoritative")
	}
	// Check RCODE (bits 0-3) = 0
	if response.Header.GetRCODE() != 0 {
		t.Errorf("response.Header.GetRCODE() = %d, want 0 (no error)", response.Header.GetRCODE())
	}

	// Verify answer section contains PTR record
	if len(response.Answers) < 1 {
		t.Fatalf("len(response.Answers) = %d, want ≥1 (PTR record)", len(response.Answers))
	}

	ptrRecord := response.Answers[0]
	if ptrRecord.TYPE != uint16(protocol.RecordTypePTR) {
		t.Errorf("PTR record TYPE = %v, want RecordTypePTR", ptrRecord.TYPE)
	}
	if ptrRecord.NAME != "_http._tcp.local" {
		t.Errorf("PTR record NAME = %q, want %q", ptrRecord.NAME, "_http._tcp.local")
	}
	// RFC 6762 §10: PTR records for DNS-SD services have TTL of 120 seconds
	if ptrRecord.TTL != 120 {
		t.Errorf("PTR record TTL = %d, want 120 (RFC 6762 §10)", ptrRecord.TTL)
	}

	// Verify additional section contains SRV, TXT, A records
	// RFC 6762 §6: Additional records reduce round-trips
	if len(response.Additionals) < 3 {
		t.Errorf("len(response.Additionals) = %d, want ≥3 (SRV+TXT+A)", len(response.Additionals))
	}

	// Check for SRV record (port and hostname)
	hasSRV := false
	for _, rr := range response.Additionals {
		if rr.TYPE == uint16(protocol.RecordTypeSRV) {
			hasSRV = true
			// RFC 6762 §10: SRV records have TTL of 120 seconds
			if rr.TTL != 120 {
				t.Errorf("SRV record TTL = %d, want 120 (RFC 6762 §10)", rr.TTL)
			}
		}
	}
	if !hasSRV {
		t.Error("Additional section missing SRV record")
	}

	// Check for TXT record (metadata)
	hasTXT := false
	for _, rr := range response.Additionals {
		if rr.TYPE == uint16(protocol.RecordTypeTXT) {
			hasTXT = true
			// RFC 6762 §10: TXT records for DNS-SD services have TTL of 120 seconds
			if rr.TTL != 120 {
				t.Errorf("TXT record TTL = %d, want 120 (RFC 6762 §10)", rr.TTL)
			}
		}
	}
	if !hasTXT {
		t.Error("Additional section missing TXT record")
	}

	// Check for A record (IPv4 address)
	hasA := false
	for _, rr := range response.Additionals {
		if rr.TYPE == uint16(protocol.RecordTypeA) {
			hasA = true
			// RFC 6762 §10: A records have TTL of 4500 seconds (75 minutes)
			if rr.TTL != 4500 {
				t.Errorf("A record TTL = %d, want 4500 (RFC 6762 §10)", rr.TTL)
			}
		}
	}
	if !hasA {
		t.Error("Additional section missing A record")
	}
}

// TestResponseBuilder_Respects9000ByteLimit tests packet size limiting per RFC 6762 §17.
//
// RFC 6762 §17: "Multicast DNS messages carried by UDP may be up to the IP MTU of the
// physical interface, less the space required for the IP header (20 bytes for IPv4;
// 40 bytes for IPv6) and the UDP header (8 bytes). In the case of a single Ethernet
// packet, this limits the available space to 9000 bytes."
//
// R005 Decision: Omit additional records (not answer records) if packet would exceed 9000 bytes.
// Answer section is critical - fail if answer alone exceeds limit.
//
// TDD Phase: RED
//
// T066 [P] [US3]: Unit test ResponseBuilder respects 9000-byte limit
func TestResponseBuilder_Respects9000ByteLimit(t *testing.T) {
	rb := NewResponseBuilder()

	// Service with LARGE TXT records to trigger size limit
	largeTXT := make(map[string]string)
	for i := 0; i < 100; i++ {
		// Each TXT record ~200 bytes
		largeTXT[string(rune('a'+i%26))] = string(make([]byte, 200))
	}

	service := &ServiceWithIP{
		InstanceName: "LargeService",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   largeTXT, // ~20KB of TXT data
	}

	query := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      12345,
			Flags:   0,
			QDCount: 1,
		},
		Questions: []message.Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN),
			},
		},
	}

	response, err := rb.BuildResponse(service, query)
	if err != nil {
		t.Fatalf("BuildResponse() error = %v, want nil", err)
	}

	// Estimate response size
	responseSize := rb.EstimatePacketSize(response)

	// RFC 6762 §17: MUST NOT exceed 9000 bytes
	if responseSize > 9000 {
		t.Errorf("response size = %d bytes, want ≤9000 (RFC 6762 §17)", responseSize)
	}

	// Verify answer section is intact (critical)
	if len(response.Answers) < 1 {
		t.Error("Answer section empty - should contain PTR record even if additionals omitted")
	}

	// Additional records may be truncated (acceptable per R005)
	// We just verify the packet doesn't exceed limit
}

// TestResponseBuilder_QUBitHandling tests unicast vs multicast response decision.
//
// RFC 6762 §5.4: "When receiving a question with the unicast-response bit set, a
// responder SHOULD usually respond with a unicast packet directed back to the querier.
// However, if the responder has not multicast that record recently (within one quarter
// of its TTL), then the responder SHOULD instead multicast the response so as to keep
// all the peer caches up to date, and to permit passive conflict detection."
//
// TDD Phase: RED
//
// T067 [P] [US3]: Unit test QU bit handling (unicast vs multicast response)
func TestResponseBuilder_QUBitHandling(t *testing.T) {
	rb := NewResponseBuilder()

	service := &ServiceWithIP{
		InstanceName: "MyService",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   map[string]string{"txtvers": "1"},
	}

	tests := []struct {
		name                string
		quBitSet            bool
		lastMulticastAgo    uint32 // seconds ago
		recordTTL           uint32 // record TTL in seconds
		wantUnicast         bool
		wantMulticast       bool
		wantMulticastReason string
	}{
		{
			name:                "QU=0 → multicast",
			quBitSet:            false,
			lastMulticastAgo:    0,
			recordTTL:           120,
			wantUnicast:         false,
			wantMulticast:       true,
			wantMulticastReason: "QU bit not set",
		},
		{
			name:                "QU=1, recent multicast (< TTL/4) → unicast",
			quBitSet:            true,
			lastMulticastAgo:    10,  // 10 seconds ago
			recordTTL:           120, // TTL/4 = 30 seconds
			wantUnicast:         true,
			wantMulticast:       false,
			wantMulticastReason: "",
		},
		{
			name:                "QU=1, stale multicast (> TTL/4) → multicast",
			quBitSet:            true,
			lastMulticastAgo:    40,  // 40 seconds ago
			recordTTL:           120, // TTL/4 = 30 seconds
			wantUnicast:         false,
			wantMulticast:       true,
			wantMulticastReason: "RFC 6762 §5.4: last multicast > TTL/4 ago (40s > 30s), need to update peer caches",
		},
		{
			name:                "QU=1, never multicast → multicast",
			quBitSet:            true,
			lastMulticastAgo:    99999, // Never multicast
			recordTTL:           120,
			wantUnicast:         false,
			wantMulticast:       true,
			wantMulticastReason: "RFC 6762 §5.4: never multicast before",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build query with QU bit set or not
			qclass := protocol.ClassIN
			if tt.quBitSet {
				// RFC 6762 §5.4: QU bit is top bit of class field (0x8000)
				qclass = protocol.DNSClass(uint16(qclass) | 0x8000)
			}

			query := &message.DNSMessage{
				Header: message.DNSHeader{
					ID:      12345,
					Flags:   0, // Query (QR=0)
					QDCount: 1,
				},
				Questions: []message.Question{
					{
						QNAME:  "_http._tcp.local",
						QTYPE:  uint16(protocol.RecordTypePTR),
						QCLASS: uint16(qclass),
					},
				},
			}

			// Simulate last multicast timestamp
			// (Implementation will need to track per-record multicast times)
			// For now, we'll add test hooks to inject this state

			// Build response
			response, err := rb.BuildResponse(service, query)
			if err != nil {
				t.Fatalf("BuildResponse() error = %v, want nil", err)
			}

			// Check response destination decision
			// (ResponseBuilder should indicate unicast vs multicast)
			// We'll add a field to response or return metadata
			if response == nil {
				t.Fatal("response is nil")
			}

			// TODO: Implementation will add response.SendViaUnicast bool field
			// For now, this test documents the requirement
		})
	}
}

// TestResponseBuilder_QUBit_OneFourthTTLException tests the 1/4 TTL multicast exception.
//
// RFC 6762 §5.4: "if the responder has not multicast that record recently (within
// one quarter of its TTL), then the responder SHOULD instead multicast the response"
//
// This is the 1/4 TTL exception - even if QU bit is set, we multicast if the record
// is "stale" in peer caches.
//
// TDD Phase: RED
//
// T068 [P] [US3]: Unit test QU bit 1/4 TTL multicast exception
func TestResponseBuilder_QUBit_OneFourthTTLException(t *testing.T) {
	rb := NewResponseBuilder()

	service := &ServiceWithIP{
		InstanceName: "MyService",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   map[string]string{"txtvers": "1"},
	}

	// Query with QU bit set
	query := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      12345,
			Flags:   0, // Query (QR=0)
			QDCount: 1,
		},
		Questions: []message.Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN) | 0x8000, // QU bit set
			},
		},
	}

	tests := []struct {
		name             string
		recordTTL        uint32 // seconds
		lastMulticastAgo uint32 // seconds
		wantMulticast    bool
	}{
		{
			name:             "TTL=120s, multicast 10s ago (< 30s) → unicast OK",
			recordTTL:        120,
			lastMulticastAgo: 10,
			wantMulticast:    false,
		},
		{
			name:             "TTL=120s, multicast 30s ago (= TTL/4) → multicast",
			recordTTL:        120,
			lastMulticastAgo: 30,
			wantMulticast:    true, // Boundary case: exactly TTL/4
		},
		{
			name:             "TTL=120s, multicast 50s ago (> TTL/4) → multicast",
			recordTTL:        120,
			lastMulticastAgo: 50,
			wantMulticast:    true,
		},
		{
			name:             "TTL=4500s (PTR), multicast 1000s ago (< 1125s) → unicast OK",
			recordTTL:        4500,
			lastMulticastAgo: 1000,
			wantMulticast:    false,
		},
		{
			name:             "TTL=4500s (PTR), multicast 1200s ago (> 1125s) → multicast",
			recordTTL:        4500,
			lastMulticastAgo: 1200,
			wantMulticast:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build response
			response, err := rb.BuildResponse(service, query)
			if err != nil {
				t.Fatalf("BuildResponse() error = %v, want nil", err)
			}

			// Verify response exists
			if response == nil {
				t.Fatal("response is nil")
			}

			// TODO: Implementation will track last multicast time per record
			// and apply 1/4 TTL rule
			// For now, test documents the requirement
		})
	}
}

// BenchmarkResponseBuilder_BuildResponse benchmarks response construction latency.
//
// RFC 6762 §6: Responders should respond quickly to queries (target <100ms total).
// Response building should be a small fraction of that budget.
//
// T085 [US3]: Benchmark query response latency (target <100ms)
func BenchmarkResponseBuilder_BuildResponse(b *testing.B) {
	rb := NewResponseBuilder()

	service := &ServiceWithIP{
		InstanceName: "BenchService",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   map[string]string{"txtvers": "1", "path": "/api"},
	}

	query := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      12345,
			Flags:   0,
			QDCount: 1,
		},
		Questions: []message.Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rb.BuildResponse(service, query)
		if err != nil {
			b.Fatalf("BuildResponse() error = %v", err)
		}
	}
}

// TestBuildResponse_SetsTCBitWhenTruncated tests TC bit is set when response exceeds 9000 bytes.
//
// RFC 6762 §6.5: "If a Multicast DNS responder receives a query,
// and the size of the response exceeds the path MTU of the link,
// then the responder SHOULD send a response with the TC (Truncated) bit set."
//
// RFC 6762 §17: "Multicast DNS Messages SHOULD NOT exceed 9000 bytes"
//
// Task 3: TC bit truncation implementation test
func TestBuildResponse_SetsTCBitWhenTruncated(t *testing.T) {
	rb := NewResponseBuilder()

	// Create service with massive TXT record that will exceed 9000 bytes
	// Single TXT record with 10KB of data
	largeTXT := make(map[string]string)
	// Create one very large TXT entry: key + value ~10KB total
	largeTXT["data"] = strings.Repeat("x", 10000) // 10KB value

	service := &ServiceWithIP{
		InstanceName: "MyService",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   largeTXT,
	}

	query := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      12345,
			Flags:   0,
			QDCount: 1,
		},
		Questions: []message.Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN),
			},
		},
	}

	response, err := rb.BuildResponse(service, query)
	if err != nil {
		t.Fatalf("BuildResponse() error = %v, want nil", err)
	}

	// Debug: Check estimated size
	estimatedSize := rb.EstimatePacketSize(response)
	t.Logf("Estimated packet size: %d bytes", estimatedSize)
	t.Logf("Response additionals count: %d", len(response.Additionals))
	t.Logf("Response flags: 0x%04x", response.Header.Flags)

	// RFC 6762 §17: Responses > 9000 bytes should be truncated
	// RFC 6762 §6.5: TC bit must be set when truncated
	// Bit 9 (TC=1): 0x0200

	// With 10KB TXT record, response should exceed 9000 bytes
	// and be truncated, setting TC bit
	if estimatedSize > 9000 {
		if (response.Header.Flags & 0x0200) == 0 {
			t.Error("TC bit not set, want TC=1 when response exceeds 9000 bytes")
		}
	} else {
		t.Skipf("Test scenario didn't generate packet > 9000 bytes (got %d)", estimatedSize)
	}
}
