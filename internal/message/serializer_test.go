package message

import (
	"testing"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestSerializeMessage_NilMessage tests that nil input returns an error.
func TestSerializeMessage_NilMessage(t *testing.T) {
	_, err := SerializeMessage(nil)
	if err == nil {
		t.Fatal("SerializeMessage(nil) error = nil, want error")
	}
}

// TestSerializeMessage_EmptyMessage tests serialization of a header-only message.
func TestSerializeMessage_EmptyMessage(t *testing.T) {
	msg := &DNSMessage{
		Header: DNSHeader{
			ID:    0,
			Flags: protocol.FlagQR | protocol.FlagAA,
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v, want nil", err)
	}

	if len(data) != 12 {
		t.Fatalf("len(data) = %d, want 12 (header only)", len(data))
	}

	// Round-trip: parse the serialized data
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ID != msg.Header.ID {
		t.Errorf("ID = %d, want %d", parsed.Header.ID, msg.Header.ID)
	}
	if parsed.Header.Flags != msg.Header.Flags {
		t.Errorf("Flags = 0x%04x, want 0x%04x", parsed.Header.Flags, msg.Header.Flags)
	}
	if parsed.Header.QDCount != 0 {
		t.Errorf("QDCount = %d, want 0", parsed.Header.QDCount)
	}
	if parsed.Header.ANCount != 0 {
		t.Errorf("ANCount = %d, want 0", parsed.Header.ANCount)
	}
}

// TestSerializeMessage_RoundTrip_SinglePTR tests serialize -> parse round-trip with a PTR record.
func TestSerializeMessage_RoundTrip_SinglePTR(t *testing.T) {
	// Build a PTR RDATA: encoded name for "myprinter._http._tcp.local"
	ptrTarget, err := EncodeServiceInstanceName("myprinter", "_http._tcp.local")
	if err != nil {
		t.Fatalf("EncodeServiceInstanceName() error = %v", err)
	}

	msg := &DNSMessage{
		Header: DNSHeader{
			ID:    0,
			Flags: protocol.FlagQR | protocol.FlagAA,
		},
		Answers: []Answer{
			{
				NAME:     "_http._tcp.local",
				TYPE:     uint16(protocol.RecordTypePTR),
				CLASS:    uint16(protocol.ClassIN),
				TTL:      120,
				RDLENGTH: uint16(len(ptrTarget)),
				RDATA:    ptrTarget,
			},
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v", err)
	}

	// Round-trip: parse
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ANCount != 1 {
		t.Fatalf("ANCount = %d, want 1", parsed.Header.ANCount)
	}
	if !parsed.Header.IsResponse() {
		t.Error("IsResponse() = false, want true")
	}

	ans := parsed.Answers[0]
	if ans.NAME != "_http._tcp.local" {
		t.Errorf("NAME = %q, want %q", ans.NAME, "_http._tcp.local")
	}
	if ans.TYPE != uint16(protocol.RecordTypePTR) {
		t.Errorf("TYPE = %d, want %d (PTR)", ans.TYPE, protocol.RecordTypePTR)
	}
	if ans.TTL != 120 {
		t.Errorf("TTL = %d, want 120", ans.TTL)
	}
}

// TestSerializeMessage_RoundTrip_MultiSection tests a response with answers and additionals.
func TestSerializeMessage_RoundTrip_MultiSection(t *testing.T) {
	// PTR answer
	ptrTarget, err := EncodeServiceInstanceName("myprinter", "_http._tcp.local")
	if err != nil {
		t.Fatalf("EncodeServiceInstanceName() error = %v", err)
	}

	// A record additional
	aData := []byte{192, 168, 1, 100}

	msg := &DNSMessage{
		Header: DNSHeader{
			ID:    0,
			Flags: protocol.FlagQR | protocol.FlagAA,
		},
		Answers: []Answer{
			{
				NAME:     "_http._tcp.local",
				TYPE:     uint16(protocol.RecordTypePTR),
				CLASS:    uint16(protocol.ClassIN),
				TTL:      120,
				RDLENGTH: uint16(len(ptrTarget)),
				RDATA:    ptrTarget,
			},
		},
		Additionals: []Answer{
			{
				NAME:     "myhost.local",
				TYPE:     uint16(protocol.RecordTypeA),
				CLASS:    uint16(protocol.ClassIN) | 0x8000, // cache-flush bit set
				TTL:      4500,
				RDLENGTH: 4,
				RDATA:    aData,
			},
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v", err)
	}

	// Round-trip
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ANCount != 1 {
		t.Errorf("ANCount = %d, want 1", parsed.Header.ANCount)
	}
	if parsed.Header.ARCount != 1 {
		t.Errorf("ARCount = %d, want 1", parsed.Header.ARCount)
	}

	// Validate additional (A record)
	additional := parsed.Additionals[0]
	if additional.NAME != "myhost.local" {
		t.Errorf("Additional NAME = %q, want %q", additional.NAME, "myhost.local")
	}
	if additional.TYPE != uint16(protocol.RecordTypeA) {
		t.Errorf("Additional TYPE = %d, want %d (A)", additional.TYPE, protocol.RecordTypeA)
	}
	// Cache-flush bit should be preserved
	if additional.CLASS&0x8000 == 0 {
		t.Error("Additional CLASS cache-flush bit not preserved")
	}
	if additional.TTL != 4500 {
		t.Errorf("Additional TTL = %d, want 4500", additional.TTL)
	}
	if len(additional.RDATA) != 4 {
		t.Errorf("Additional RDATA len = %d, want 4", len(additional.RDATA))
	}
}

// TestSerializeMessage_ServiceNameWithSpaces tests names with spaces per RFC 6763 §4.3.
func TestSerializeMessage_ServiceNameWithSpaces(t *testing.T) {
	// Build PTR RDATA with a service instance name containing spaces
	ptrTarget, err := EncodeServiceInstanceName("My Web Server", "_http._tcp.local")
	if err != nil {
		t.Fatalf("EncodeServiceInstanceName() error = %v", err)
	}

	msg := &DNSMessage{
		Header: DNSHeader{
			Flags: protocol.FlagQR | protocol.FlagAA,
		},
		Answers: []Answer{
			{
				NAME:     "_http._tcp.local",
				TYPE:     uint16(protocol.RecordTypePTR),
				CLASS:    uint16(protocol.ClassIN),
				TTL:      120,
				RDLENGTH: uint16(len(ptrTarget)),
				RDATA:    ptrTarget,
			},
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v", err)
	}

	// Should be parseable
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ANCount != 1 {
		t.Fatalf("ANCount = %d, want 1", parsed.Header.ANCount)
	}
}

// TestSerializeMessage_CacheFlushBitPreservation tests the cache-flush bit round-trip.
func TestSerializeMessage_CacheFlushBitPreservation(t *testing.T) {
	aData := []byte{10, 0, 0, 1}

	msg := &DNSMessage{
		Header: DNSHeader{
			Flags: protocol.FlagQR | protocol.FlagAA,
		},
		Answers: []Answer{
			{
				NAME:     "myhost.local",
				TYPE:     uint16(protocol.RecordTypeA),
				CLASS:    uint16(protocol.ClassIN) | 0x8000, // cache-flush
				TTL:      4500,
				RDLENGTH: 4,
				RDATA:    aData,
			},
			{
				NAME:     "_http._tcp.local",
				TYPE:     uint16(protocol.RecordTypePTR),
				CLASS:    uint16(protocol.ClassIN), // no cache-flush (shared record)
				TTL:      120,
				RDLENGTH: 0,
				RDATA:    []byte{0}, // root
			},
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v", err)
	}

	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ANCount != 2 {
		t.Fatalf("ANCount = %d, want 2", parsed.Header.ANCount)
	}

	// A record should have cache-flush bit set
	if parsed.Answers[0].CLASS&0x8000 == 0 {
		t.Error("A record cache-flush bit not preserved (should be set)")
	}

	// PTR record should NOT have cache-flush bit
	if parsed.Answers[1].CLASS&0x8000 != 0 {
		t.Error("PTR record cache-flush bit incorrectly set (should be clear)")
	}
}

// TestSerializeMessage_WithQuestions tests serialization of a message with questions.
func TestSerializeMessage_WithQuestions(t *testing.T) {
	msg := &DNSMessage{
		Header: DNSHeader{
			ID:    0x1234,
			Flags: 0, // query
		},
		Questions: []Question{
			{
				QNAME:  "_http._tcp.local",
				QTYPE:  uint16(protocol.RecordTypePTR),
				QCLASS: uint16(protocol.ClassIN),
			},
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage() error = %v", err)
	}

	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if parsed.Header.ID != 0x1234 {
		t.Errorf("ID = 0x%04x, want 0x1234", parsed.Header.ID)
	}
	if parsed.Header.QDCount != 1 {
		t.Errorf("QDCount = %d, want 1", parsed.Header.QDCount)
	}
	if parsed.Questions[0].QNAME != "_http._tcp.local" {
		t.Errorf("QNAME = %q, want %q", parsed.Questions[0].QNAME, "_http._tcp.local")
	}
	if parsed.Questions[0].QTYPE != uint16(protocol.RecordTypePTR) {
		t.Errorf("QTYPE = %d, want %d", parsed.Questions[0].QTYPE, protocol.RecordTypePTR)
	}
}

// TestAnswerToResourceRecord tests the Answer-to-ResourceRecord conversion.
func TestAnswerToResourceRecord(t *testing.T) {
	tests := []struct {
		name           string
		answer         Answer
		wantType       protocol.RecordType
		wantClass      protocol.DNSClass
		wantCacheFlush bool
	}{
		{
			name: "A record with cache-flush",
			answer: Answer{
				NAME:  "myhost.local",
				TYPE:  1,
				CLASS: 0x8001, // IN + cache-flush
				TTL:   4500,
				RDATA: []byte{10, 0, 0, 1},
			},
			wantType:       protocol.RecordTypeA,
			wantClass:      protocol.ClassIN,
			wantCacheFlush: true,
		},
		{
			name: "PTR record without cache-flush",
			answer: Answer{
				NAME:  "_http._tcp.local",
				TYPE:  12,
				CLASS: 0x0001, // IN, no cache-flush
				TTL:   120,
				RDATA: []byte{0},
			},
			wantType:       protocol.RecordTypePTR,
			wantClass:      protocol.ClassIN,
			wantCacheFlush: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := answerToResourceRecord(&tt.answer)
			if rr.Type != tt.wantType {
				t.Errorf("Type = %d, want %d", rr.Type, tt.wantType)
			}
			if rr.Class != tt.wantClass {
				t.Errorf("Class = %d, want %d", rr.Class, tt.wantClass)
			}
			if rr.CacheFlush != tt.wantCacheFlush {
				t.Errorf("CacheFlush = %v, want %v", rr.CacheFlush, tt.wantCacheFlush)
			}
			if rr.TTL != tt.answer.TTL {
				t.Errorf("TTL = %d, want %d", rr.TTL, tt.answer.TTL)
			}
			if rr.Name != tt.answer.NAME {
				t.Errorf("Name = %q, want %q", rr.Name, tt.answer.NAME)
			}
		})
	}
}
