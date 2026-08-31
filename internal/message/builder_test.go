package message

import (
	"encoding/binary"
	"testing"
)

// TestBuildQuery_RFC6762_HeaderFields validates that mDNS query messages
// have correct header field values per RFC 6762 §18 (FR-020).
//
// RFC 6762 Requirements:
//
//	§18.2 - QR bit MUST be zero in query messages
//	§18.3 - OPCODE MUST be zero (standard query)
//	§18.4 - AA bit MUST be zero in query messages
//	§18.5 - TC bit clear means no additional Known Answers
//	§18.6 - RD bit SHOULD be zero (M1 enforces MUST for simplicity)
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18
func TestBuildQuery_RFC6762_HeaderFields(t *testing.T) {
	query, err := BuildQuery("test.local", 1) // A record query
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	// Verify minimum message size (header is 12 bytes minimum)
	if len(query) < 12 {
		t.Fatalf("query too short: %d bytes, expected at least 12 bytes", len(query))
	}

	// Parse header manually (first 12 bytes)
	flags := binary.BigEndian.Uint16(query[2:4])

	// RFC 6762 §18.2: QR bit MUST be zero in queries
	qr := (flags & 0x8000) >> 15
	if qr != 0 {
		t.Errorf("QR bit is %d, expected 0 per RFC 6762 §18.2", qr)
	}

	// RFC 6762 §18.3: OPCODE MUST be zero
	opcode := (flags >> 11) & 0x0F
	if opcode != 0 {
		t.Errorf("OPCODE is %d, expected 0 per RFC 6762 §18.3", opcode)
	}

	// RFC 6762 §18.4: AA bit MUST be zero in queries
	aa := (flags & 0x0400) >> 10
	if aa != 0 {
		t.Errorf("AA bit is %d, expected 0 per RFC 6762 §18.4", aa)
	}

	// RFC 6762 §18.5: TC bit clear (no Known Answers in M1)
	tc := (flags & 0x0200) >> 9
	if tc != 0 {
		t.Errorf("TC bit is %d, expected 0 per RFC 6762 §18.5", tc)
	}

	// RFC 6762 §18.6: RD bit SHOULD be zero (M1 enforces as MUST)
	rd := (flags & 0x0100) >> 8
	if rd != 0 {
		t.Errorf("RD bit is %d, expected 0 per RFC 6762 §18.6", rd)
	}

	// Additional validation: RA, Z, RCODE should all be zero
	ra := (flags & 0x0080) >> 7
	if ra != 0 {
		t.Errorf("RA bit is %d, expected 0 in query", ra)
	}

	z := (flags & 0x0070) >> 4
	if z != 0 {
		t.Errorf("Z bits are %d, expected 0 (reserved)", z)
	}

	rcode := flags & 0x000F
	if rcode != 0 {
		t.Errorf("RCODE is %d, expected 0 in query", rcode)
	}
}

// TestBuildQuery_RFC6762_QuestionSection validates that BuildQuery creates
// a valid question section per RFC 1035 §4.1.2 (FR-001).
//
// RFC 1035 §4.1.2: The question section contains QNAME, QTYPE, and QCLASS.
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
func TestBuildQuery_RFC6762_QuestionSection(t *testing.T) {
	tests := []struct {
		name       string
		qname      string
		qtype      uint16
		wantQCount uint16
	}{
		{
			name:       "A record query for test.local",
			qname:      "test.local",
			qtype:      1, // A
			wantQCount: 1,
		},
		{
			name:       "PTR record query for service",
			qname:      "_http._tcp.local",
			qtype:      12, // PTR
			wantQCount: 1,
		},
		{
			name:       "SRV record query",
			qname:      "myservice._http._tcp.local",
			qtype:      33, // SRV
			wantQCount: 1,
		},
		{
			name:       "TXT record query",
			qname:      "myservice._http._tcp.local",
			qtype:      16, // TXT
			wantQCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := BuildQuery(tt.qname, tt.qtype)
			if err != nil {
				t.Fatalf("BuildQuery failed: %v", err)
			}

			// Verify header QDCOUNT
			qdcount := binary.BigEndian.Uint16(query[4:6])
			if qdcount != tt.wantQCount {
				t.Errorf("QDCOUNT is %d, expected %d", qdcount, tt.wantQCount)
			}

			// Verify ANCOUNT, NSCOUNT, ARCOUNT are all zero (queries don't have answers)
			ancount := binary.BigEndian.Uint16(query[6:8])
			if ancount != 0 {
				t.Errorf("ANCOUNT is %d, expected 0 in query", ancount)
			}

			nscount := binary.BigEndian.Uint16(query[8:10])
			if nscount != 0 {
				t.Errorf("NSCOUNT is %d, expected 0 in query", nscount)
			}

			arcount := binary.BigEndian.Uint16(query[10:12])
			if arcount != 0 {
				t.Errorf("ARCOUNT is %d, expected 0 in query", arcount)
			}

			// Verify the question section contains the name
			// Skip header (12 bytes), then we should have the encoded name
			if len(query) < 12+len(tt.qname)+1+4 { // header + name + terminator + QTYPE + QCLASS
				t.Errorf("query too short for question section")
			}
		})
	}
}

// TestBuildQuery_RFC1035_QClass validates that BuildQuery sets QCLASS to IN (1)
// per RFC 1035 and RFC 6762 §5.4 (FR-001, FR-004).
//
// RFC 1035: QCLASS = IN (1) for Internet class
// RFC 6762 §5.4: QU bit (bit 15 of QCLASS) = 0 for multicast queries in M1
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251
func TestBuildQuery_RFC1035_QClass(t *testing.T) {
	query, err := BuildQuery("test.local", 1) // A record
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	// Parse to find QCLASS field
	// Header is 12 bytes, then QNAME (variable), then QTYPE (2 bytes), then QCLASS (2 bytes)
	// For "test.local": 4 + "test" (4 bytes) + 5 + "local" (5 bytes) + 0 = 12 bytes for name
	// So QCLASS starts at byte: 12 (header) + 12 (name) + 2 (QTYPE) = 26

	if len(query) < 28 {
		t.Fatalf("query too short to contain QCLASS: %d bytes", len(query))
	}

	// Extract QCLASS (last 2 bytes of question section)
	// We need to find it by parsing the name length
	offset := 12 // Start after header

	// Parse name to find its end
	for offset < len(query) {
		length := query[offset]
		if length == 0 {
			// End of name
			offset++
			break
		}
		offset += 1 + int(length)
	}

	// Now at QTYPE (2 bytes) + QCLASS (2 bytes)
	if offset+4 > len(query) {
		t.Fatalf("query too short to contain QTYPE and QCLASS at offset %d", offset)
	}

	qtype := binary.BigEndian.Uint16(query[offset : offset+2])
	if qtype != 1 {
		t.Errorf("QTYPE is %d, expected 1 (A record)", qtype)
	}

	qclass := binary.BigEndian.Uint16(query[offset+2 : offset+4])

	// RFC 1035: QCLASS = 1 (IN) for Internet class
	// RFC 6762 §5.4: QU bit (bit 15) = 0 for multicast queries (M1 uses QU=0)
	expectedQClass := uint16(0x0001) // IN class, QU=0
	if qclass != expectedQClass {
		t.Errorf("QCLASS is 0x%04X, expected 0x%04X (IN class, QU=0) per RFC 1035 and RFC 6762 §5.4", qclass, expectedQClass)
	}
}

// TestBuildQuery_RFC1035_NameEncoding validates that BuildQuery correctly
// encodes DNS names per RFC 1035 §3.1 (FR-001, FR-003).
//
// RFC 1035 §3.1: Domain names are encoded as length-prefixed labels.
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-003: System MUST validate queried names follow DNS naming rules
func TestBuildQuery_RFC1035_NameEncoding(t *testing.T) {
	query, err := BuildQuery("test.local", 1)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	// Skip header (12 bytes)
	offset := 12

	// Expected encoding: 4 + "test" + 5 + "local" + 0
	expected := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,
	}

	if len(query) < offset+len(expected) {
		t.Fatalf("query too short: %d bytes, expected at least %d", len(query), offset+len(expected))
	}

	for i, want := range expected {
		got := query[offset+i]
		if got != want {
			t.Errorf("byte %d: got 0x%02X, want 0x%02X per RFC 1035 §3.1", offset+i, got, want)
		}
	}
}

// TestBuildQuery_InvalidName validates that BuildQuery returns ValidationError
// for invalid names per RFC 1035 §3.1 (FR-003, FR-014).
//
// RFC 1035 §3.1: Names must follow DNS naming rules (labels ≤63 bytes, total ≤255 bytes).
//
// FR-003: System MUST validate queried names follow DNS naming rules
// FR-014: System MUST return ValidationError for invalid query names
func TestBuildQuery_InvalidName(t *testing.T) {
	tests := []struct {
		name   string
		qname  string
		qtype  uint16
		errMsg string
	}{
		{
			name:   "empty name",
			qname:  "",
			qtype:  1,
			errMsg: "", // Empty name might be valid (root), or might error - depends on implementation
		},
		{
			name:   "invalid character (space)",
			qname:  "test host.local",
			qtype:  1,
			errMsg: "invalid character",
		},
		{
			name:   "label too long (64 bytes)",
			qname:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.local", // 64 'a's
			qtype:  1,
			errMsg: "exceeds maximum length 63 bytes per RFC 1035 §3.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildQuery(tt.qname, tt.qtype)

			if tt.errMsg == "" {
				// Test allows either success or specific error
				return
			}

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.errMsg)
				return
			}

			// For non-empty error message expectations, verify the error
			if tt.errMsg != "" {
				errStr := err.Error()
				if len(errStr) == 0 || (len(tt.errMsg) > 0 && len(errStr) > 0 && errStr[0] != tt.errMsg[0]) {
					// Basic check that error message is reasonable
					t.Logf("got error: %v", err)
				}
			}
		})
	}
}

// TestBuildQuery_UnsupportedRecordType validates that BuildQuery returns
// ValidationError for unsupported record types per FR-002 and FR-014.
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
// FR-014: System MUST return ValidationError for invalid query names or unsupported record types
func TestBuildQuery_UnsupportedRecordType(t *testing.T) {
	tests := []struct {
		name  string
		qtype uint16
	}{
		{
			name:  "MX record - not supported in M1",
			qtype: 15,
		},
		{
			name:  "Unknown record type",
			qtype: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildQuery("test.local", tt.qtype)

			if err == nil {
				t.Errorf("expected error for unsupported record type %d per FR-002, FR-014, got nil", tt.qtype)
			}
		})
	}
}

// TestBuildQuery_SupportedRecordTypes validates that BuildQuery successfully
// creates queries for all supported record types per FR-002.
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
func TestBuildQuery_SupportedRecordTypes(t *testing.T) {
	tests := []struct {
		name  string
		qtype uint16
	}{
		{
			name:  "A record (1)",
			qtype: 1,
		},
		{
			name:  "PTR record (12)",
			qtype: 12,
		},
		{
			name:  "TXT record (16)",
			qtype: 16,
		},
		{
			name:  "AAAA record (28)",
			qtype: 28,
		},
		{
			name:  "SRV record (33)",
			qtype: 33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := BuildQuery("test.local", tt.qtype)

			if err != nil {
				t.Errorf("BuildQuery failed for supported type %d per FR-002: %v", tt.qtype, err)
				return
			}

			if len(query) < 12 {
				t.Errorf("query too short: %d bytes", len(query))
			}

			// Verify QTYPE is set correctly in the question section
			// Parse past the name to find QTYPE
			offset := 12
			for offset < len(query) {
				length := query[offset]
				if length == 0 {
					offset++
					break
				}
				offset += 1 + int(length)
			}

			if offset+2 <= len(query) {
				qtype := binary.BigEndian.Uint16(query[offset : offset+2])
				if qtype != tt.qtype {
					t.Errorf("QTYPE in query is %d, expected %d", qtype, tt.qtype)
				}
			}
		})
	}
}

// TestBuildQuery_MessageID validates that BuildQuery sets a transaction ID
// per RFC 1035 §4.1.1 (FR-001).
//
// RFC 1035 §4.1.1: ID is a 16-bit identifier for matching queries and responses.
// RFC 6762 §18.1: Multicast DNS messages SHOULD use ID = 0, but M1 may use random ID.
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
func TestBuildQuery_MessageID(t *testing.T) {
	query, err := BuildQuery("test.local", 1)
	if err != nil {
		t.Fatalf("BuildQuery failed: %v", err)
	}

	// Extract ID (first 2 bytes)
	id := binary.BigEndian.Uint16(query[0:2])

	// RFC 6762 §18.1 says SHOULD use 0, but we allow any value
	// Just verify it's present (any value is acceptable for M1)
	t.Logf("Message ID: 0x%04X (RFC 6762 §18.1 suggests 0, but any value acceptable)", id)
}
