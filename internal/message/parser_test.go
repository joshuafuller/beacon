package message

import (
	goerrors "errors"
	"net"
	"strings"
	"testing"

	"github.com/joshuafuller/beacon/internal/errors"
)

const testLocalName = "test.local"

// TestParseMessage_RFC1035_ValidResponse validates that ParseMessage correctly
// parses a complete DNS response message per RFC 1035 §4.1 (FR-009).
//
// RFC 1035 §4.1: "All communications inside of the domain protocol are carried
// in a single format called a message."
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseMessage_RFC1035_ValidResponse(t *testing.T) {
	// Build a simple response: "test.local" A record with IP 192.168.1.100
	// Header (12 bytes) + Question (17 bytes) + Answer (29 bytes)
	msg := make([]byte, 0)

	// Header: ID=0x1234, Flags=0x8000 (QR=1, response), QDCOUNT=1, ANCOUNT=1
	header := []byte{
		0x12, 0x34, // ID
		0x80, 0x00, // Flags: QR=1, OPCODE=0, AA=0, TC=0, RD=0, RA=0, Z=0, RCODE=0
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x01, // ANCOUNT = 1
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
	}
	msg = append(msg, header...)

	// Question: "test.local" A IN
	question := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,       // Name terminator
		0x00, 0x01, // QTYPE = A (1)
		0x00, 0x01, // QCLASS = IN (1)
	}
	msg = append(msg, question...)

	// Answer: "test.local" A IN 120 192.168.1.100
	answer := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,       // Name terminator
		0x00, 0x01, // TYPE = A (1)
		0x00, 0x01, // CLASS = IN (1)
		0x00, 0x00, 0x00, 0x78, // TTL = 120 seconds
		0x00, 0x04, // RDLENGTH = 4
		192, 168, 1, 100, // RDATA = IP address
	}
	msg = append(msg, answer...)

	// Parse the message
	parsed, err := ParseMessage(msg)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	// Verify header
	if parsed.Header.ID != 0x1234 {
		t.Errorf("Header.ID = 0x%04X, want 0x1234", parsed.Header.ID)
	}

	if !parsed.Header.IsResponse() {
		t.Error("Header.IsResponse() = false, want true per RFC 1035 §4.1.1")
	}

	if parsed.Header.QDCount != 1 {
		t.Errorf("Header.QDCount = %d, want 1", parsed.Header.QDCount)
	}

	if parsed.Header.ANCount != 1 {
		t.Errorf("Header.ANCount = %d, want 1", parsed.Header.ANCount)
	}

	// Verify question
	if len(parsed.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want 1", len(parsed.Questions))
	}

	if parsed.Questions[0].QNAME != testLocalName {
		t.Errorf("Questions[0].QNAME = %q, want %q", parsed.Questions[0].QNAME, testLocalName)
	}

	if parsed.Questions[0].QTYPE != 1 {
		t.Errorf("Questions[0].QTYPE = %d, want 1 (A)", parsed.Questions[0].QTYPE)
	}

	// Verify answer
	if len(parsed.Answers) != 1 {
		t.Fatalf("len(Answers) = %d, want 1", len(parsed.Answers))
	}

	if parsed.Answers[0].NAME != testLocalName {
		t.Errorf("Answers[0].NAME = %q, want %q", parsed.Answers[0].NAME, testLocalName)
	}

	if parsed.Answers[0].TYPE != 1 {
		t.Errorf("Answers[0].TYPE = %d, want 1 (A)", parsed.Answers[0].TYPE)
	}

	if parsed.Answers[0].TTL != 120 {
		t.Errorf("Answers[0].TTL = %d, want 120", parsed.Answers[0].TTL)
	}

	if len(parsed.Answers[0].RDATA) != 4 {
		t.Fatalf("len(Answers[0].RDATA) = %d, want 4", len(parsed.Answers[0].RDATA))
	}

	expectedIP := []byte{192, 168, 1, 100}
	for i, want := range expectedIP {
		if parsed.Answers[0].RDATA[i] != want {
			t.Errorf("Answers[0].RDATA[%d] = %d, want %d", i, parsed.Answers[0].RDATA[i], want)
		}
	}
}

// TestParseHeader_RFC1035_Format validates that ParseHeader correctly extracts
// all header fields per RFC 1035 §4.1.1 (FR-009).
//
// RFC 1035 §4.1.1: The header section is always present and contains fields
// which specify which of the remaining sections are present.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseHeader_RFC1035_Format(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		want   DNSHeader
	}{
		{
			name: "query header per RFC 1035 §4.1.1",
			header: []byte{
				0x00, 0x00, // ID = 0
				0x00, 0x00, // Flags: QR=0 (query)
				0x00, 0x01, // QDCOUNT = 1
				0x00, 0x00, // ANCOUNT = 0
				0x00, 0x00, // NSCOUNT = 0
				0x00, 0x00, // ARCOUNT = 0
			},
			want: DNSHeader{
				ID:      0,
				Flags:   0x0000,
				QDCount: 1,
				ANCount: 0,
				NSCount: 0,
				ARCount: 0,
			},
		},
		{
			name: "response header per RFC 1035 §4.1.1",
			header: []byte{
				0x12, 0x34, // ID = 0x1234
				0x81, 0x80, // Flags: QR=1, RD=1, RA=1
				0x00, 0x01, // QDCOUNT = 1
				0x00, 0x02, // ANCOUNT = 2
				0x00, 0x00, // NSCOUNT = 0
				0x00, 0x01, // ARCOUNT = 1
			},
			want: DNSHeader{
				ID:      0x1234,
				Flags:   0x8180,
				QDCount: 1,
				ANCount: 2,
				NSCount: 0,
				ARCount: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHeader(tt.header)
			if err != nil {
				t.Fatalf("ParseHeader failed: %v", err)
			}

			if got.ID != tt.want.ID {
				t.Errorf("ID = 0x%04X, want 0x%04X", got.ID, tt.want.ID)
			}

			if got.Flags != tt.want.Flags {
				t.Errorf("Flags = 0x%04X, want 0x%04X", got.Flags, tt.want.Flags)
			}

			if got.QDCount != tt.want.QDCount {
				t.Errorf("QDCount = %d, want %d", got.QDCount, tt.want.QDCount)
			}

			if got.ANCount != tt.want.ANCount {
				t.Errorf("ANCount = %d, want %d", got.ANCount, tt.want.ANCount)
			}
		})
	}
}

// TestParseHeader_TruncatedMessage validates that ParseHeader returns
// WireFormatError for truncated headers per FR-011, FR-015.
//
// FR-011: System MUST validate response message format and discard malformed packets
// FR-015: System MUST return WireFormatError for malformed response packets
func TestParseHeader_TruncatedMessage(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		errMsg string
	}{
		{
			name:   "empty message",
			header: []byte{},
			errMsg: "message too short",
		},
		{
			name:   "partial header (11 bytes)",
			header: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			errMsg: "message too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseHeader(tt.header)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.errMsg)
				return
			}

			// Verify it's a WireFormatError per FR-015
			var wireErr *errors.WireFormatError
			if !goerrors.As(err, &wireErr) {
				t.Errorf("expected WireFormatError per FR-015, got %T", err)
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
			}
		})
	}
}

// TestParseQuestion_RFC1035_Format validates that ParseQuestion correctly
// parses question sections per RFC 1035 §4.1.2 (FR-009).
//
// RFC 1035 §4.1.2: The question section contains QNAME, QTYPE, and QCLASS.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseQuestion_RFC1035_Format(t *testing.T) {
	// Question: "test.local" A IN
	questionData := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,       // Name terminator
		0x00, 0x01, // QTYPE = A (1)
		0x00, 0x01, // QCLASS = IN (1)
	}

	question, newOffset, err := ParseQuestion(questionData, 0)
	if err != nil {
		t.Fatalf("ParseQuestion failed: %v", err)
	}

	if question.QNAME != testLocalName {
		t.Errorf("QNAME = %q, want %q per RFC 1035 §4.1.2", question.QNAME, testLocalName)
	}

	if question.QTYPE != 1 {
		t.Errorf("QTYPE = %d, want 1 (A) per RFC 1035 §4.1.2", question.QTYPE)
	}

	if question.QCLASS != 1 {
		t.Errorf("QCLASS = %d, want 1 (IN) per RFC 1035 §4.1.2", question.QCLASS)
	}

	expectedOffset := len(questionData)
	if newOffset != expectedOffset {
		t.Errorf("newOffset = %d, want %d", newOffset, expectedOffset)
	}
}

// TestParseAnswer_RFC1035_Format validates that ParseAnswer correctly
// parses answer sections per RFC 1035 §4.1.3 (FR-009).
//
// RFC 1035 §4.1.3: The answer section contains RRs that answer the question.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseAnswer_RFC1035_Format(t *testing.T) {
	// Answer: "test.local" A IN 120 192.168.1.100
	answerData := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,       // Name terminator
		0x00, 0x01, // TYPE = A (1)
		0x00, 0x01, // CLASS = IN (1)
		0x00, 0x00, 0x00, 0x78, // TTL = 120
		0x00, 0x04, // RDLENGTH = 4
		192, 168, 1, 100, // RDATA
	}

	answer, newOffset, err := ParseAnswer(answerData, 0)
	if err != nil {
		t.Fatalf("ParseAnswer failed: %v", err)
	}

	if answer.NAME != testLocalName {
		t.Errorf("NAME = %q, want %q per RFC 1035 §4.1.3", answer.NAME, testLocalName)
	}

	if answer.TYPE != 1 {
		t.Errorf("TYPE = %d, want 1 (A) per RFC 1035 §4.1.3", answer.TYPE)
	}

	if answer.CLASS != 1 {
		t.Errorf("CLASS = %d, want 1 (IN) per RFC 1035 §4.1.3", answer.CLASS)
	}

	if answer.TTL != 120 {
		t.Errorf("TTL = %d, want 120 per RFC 1035 §4.1.3", answer.TTL)
	}

	if answer.RDLENGTH != 4 {
		t.Errorf("RDLENGTH = %d, want 4 per RFC 1035 §4.1.3", answer.RDLENGTH)
	}

	expectedOffset := len(answerData)
	if newOffset != expectedOffset {
		t.Errorf("newOffset = %d, want %d", newOffset, expectedOffset)
	}
}

// TestParseRDATA_ARecord validates that ParseRDATA correctly parses A record
// RDATA (IPv4 address) per RFC 1035 §3.4.1 (FR-009).
//
// RFC 1035 §3.4.1: A RDATA format is a 32-bit Internet address.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseRDATA_ARecord(t *testing.T) {
	rdata := []byte{192, 168, 1, 100}

	result, err := ParseRDATA(1, rdata) // TYPE = A (1)
	if err != nil {
		t.Fatalf("ParseRDATA failed: %v", err)
	}

	ip, ok := result.(net.IP)
	if !ok {
		t.Fatalf("ParseRDATA returned %T, want net.IP per RFC 1035 §3.4.1", result)
	}

	expected := net.IPv4(192, 168, 1, 100)
	if !ip.Equal(expected) {
		t.Errorf("IP = %s, want %s per RFC 1035 §3.4.1", ip, expected)
	}
}

// TestParseRDATA_PTRRecord validates that ParseRDATA correctly parses PTR record
// RDATA (domain name) per RFC 1035 §3.3.12 (FR-009).
//
// RFC 1035 §3.3.12: PTR RDATA format is a domain name.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseRDATA_PTRRecord(t *testing.T) {
	// PTR RDATA: "myservice._http._tcp.local"
	rdata := []byte{
		0x09, 'm', 'y', 's', 'e', 'r', 'v', 'i', 'c', 'e',
		0x05, '_', 'h', 't', 't', 'p',
		0x04, '_', 't', 'c', 'p',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,
	}

	result, err := ParseRDATA(12, rdata) // TYPE = PTR (12)
	if err != nil {
		t.Fatalf("ParseRDATA failed: %v", err)
	}

	name, ok := result.(string)
	if !ok {
		t.Fatalf("ParseRDATA returned %T, want string per RFC 1035 §3.3.12", result)
	}

	expected := "myservice._http._tcp.local"
	if name != expected {
		t.Errorf("PTR name = %q, want %q per RFC 1035 §3.3.12", name, expected)
	}
}

// TestParseRDATA_SRVRecord validates that ParseRDATA correctly parses SRV record
// RDATA (priority, weight, port, target) per RFC 2782 (FR-009).
//
// RFC 2782: SRV RDATA format is priority, weight, port, and target.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseRDATA_SRVRecord(t *testing.T) {
	// SRV RDATA: priority=10, weight=20, port=8080, target="server.local"
	rdata := []byte{
		0x00, 0x0A, // Priority = 10
		0x00, 0x14, // Weight = 20
		0x1F, 0x90, // Port = 8080
		0x06, 's', 'e', 'r', 'v', 'e', 'r',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,
	}

	result, err := ParseRDATA(33, rdata) // TYPE = SRV (33)
	if err != nil {
		t.Fatalf("ParseRDATA failed: %v", err)
	}

	srv, ok := result.(SRVData)
	if !ok {
		t.Fatalf("ParseRDATA returned %T, want SRVData per RFC 2782", result)
	}

	if srv.Priority != 10 {
		t.Errorf("Priority = %d, want 10 per RFC 2782", srv.Priority)
	}

	if srv.Weight != 20 {
		t.Errorf("Weight = %d, want 20 per RFC 2782", srv.Weight)
	}

	if srv.Port != 8080 {
		t.Errorf("Port = %d, want 8080 per RFC 2782", srv.Port)
	}

	if srv.Target != "server.local" {
		t.Errorf("Target = %q, want %q per RFC 2782", srv.Target, "server.local")
	}
}

// TestParseRDATA_TXTRecord validates that ParseRDATA correctly parses TXT record
// RDATA (text strings) per RFC 1035 §3.3.14 (FR-009).
//
// RFC 1035 §3.3.14: TXT RDATA format is one or more character strings.
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
func TestParseRDATA_TXTRecord(t *testing.T) {
	// TXT RDATA: "version=1.0" "path=/api"
	rdata := []byte{
		0x0B, 'v', 'e', 'r', 's', 'i', 'o', 'n', '=', '1', '.', '0',
		0x09, 'p', 'a', 't', 'h', '=', '/', 'a', 'p', 'i',
	}

	result, err := ParseRDATA(16, rdata) // TYPE = TXT (16)
	if err != nil {
		t.Fatalf("ParseRDATA failed: %v", err)
	}

	txt, ok := result.([]string)
	if !ok {
		t.Fatalf("ParseRDATA returned %T, want []string per RFC 1035 §3.3.14", result)
	}

	if len(txt) != 2 {
		t.Fatalf("len(TXT) = %d, want 2 per RFC 1035 §3.3.14", len(txt))
	}

	if txt[0] != "version=1.0" {
		t.Errorf("TXT[0] = %q, want %q per RFC 1035 §3.3.14", txt[0], "version=1.0")
	}

	if txt[1] != "path=/api" {
		t.Errorf("TXT[1] = %q, want %q per RFC 1035 §3.3.14", txt[1], "path=/api")
	}
}

// TestParseMessage_MalformedPacket validates that ParseMessage returns
// WireFormatError for malformed packets per FR-011, FR-015.
//
// FR-011: System MUST validate response message format and discard malformed packets
// FR-015: System MUST return WireFormatError for malformed response packets
func TestParseMessage_MalformedPacket(t *testing.T) {
	tests := []struct {
		name   string
		msg    []byte
		errMsg string
	}{
		{
			name:   "truncated header",
			msg:    []byte{0x00, 0x00, 0x00, 0x00},
			errMsg: "message too short",
		},
		{
			name: "truncated question section",
			msg: []byte{
				0x00, 0x00, // ID
				0x00, 0x00, // Flags
				0x00, 0x01, // QDCOUNT = 1
				0x00, 0x00, // ANCOUNT
				0x00, 0x00, // NSCOUNT
				0x00, 0x00, // ARCOUNT
				// Missing question section
			},
			errMsg: "unexpected end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMessage(tt.msg)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.errMsg)
				return
			}

			// Verify it's a WireFormatError per FR-015
			var wireErr *errors.WireFormatError
			if !goerrors.As(err, &wireErr) {
				t.Errorf("expected WireFormatError per FR-015, got %T", err)
			}
		})
	}
}

// TestParseMessage_WithCompression validates that ParseMessage correctly handles
// DNS name compression in answers per RFC 1035 §4.1.4 (FR-012).
//
// RFC 1035 §4.1.4: Message compression allows domain names to be replaced by
// pointers to prior occurrences.
//
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
func TestParseMessage_WithCompression(t *testing.T) {
	// Build message with compression pointer in answer
	msg := make([]byte, 0)

	// Header
	header := []byte{
		0x00, 0x00, // ID
		0x80, 0x00, // Flags: QR=1
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x01, // ANCOUNT = 1
		0x00, 0x00, // NSCOUNT
		0x00, 0x00, // ARCOUNT
	}
	msg = append(msg, header...)

	// Question: "test.local" A IN (starts at offset 12)
	question := []byte{
		0x04, 't', 'e', 's', 't',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,
		0x00, 0x01, // QTYPE = A
		0x00, 0x01, // QCLASS = IN
	}
	msg = append(msg, question...)

	// Answer: Use compression pointer to question name (offset 12)
	answer := []byte{
		0xC0, 0x0C, // Compression pointer to offset 12 ("test.local")
		0x00, 0x01, // TYPE = A
		0x00, 0x01, // CLASS = IN
		0x00, 0x00, 0x00, 0x78, // TTL = 120
		0x00, 0x04, // RDLENGTH = 4
		192, 168, 1, 100, // RDATA
	}
	msg = append(msg, answer...)

	parsed, err := ParseMessage(msg)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	// Verify answer name was decompressed per FR-012
	if len(parsed.Answers) != 1 {
		t.Fatalf("len(Answers) = %d, want 1", len(parsed.Answers))
	}

	if parsed.Answers[0].NAME != testLocalName {
		t.Errorf("Answer NAME = %q, want %q (decompressed per RFC 1035 §4.1.4)", parsed.Answers[0].NAME, testLocalName)
	}
}

// TestParseRDATAInMessage_CompressedSRVTarget proves the compression fix
// (FR-012): an SRV target encoded with a DNS compression pointer (as Avahi and
// Bonjour do) resolves correctly when parsed against the full message, whereas
// the context-free ParseRDATA cannot follow the pointer out of the isolated
// RDATA slice.
func TestParseRDATAInMessage_CompressedSRVTarget(t *testing.T) {
	// Build a synthetic message:
	//   [0..11]  12-byte header (zeros)
	//   [12..18] label sequence "local" + root: 05 'l' 'o' 'c' 'a' 'l' 00
	//   [19..]   SRV RDATA: prio(0) weight(0) port(8080) "host" <ptr->12>
	msg := make([]byte, 12)                                // header
	msg = append(msg, 0x05, 'l', 'o', 'c', 'a', 'l', 0x00) // "local." at offset 12

	rdataStart := len(msg) // 19
	srv := []byte{
		0x00, 0x00, // priority
		0x00, 0x00, // weight
		0x1f, 0x90, // port 8080
		0x04, 'h', 'o', 's', 't', // "host"
		0xC0, 0x0C, // compression pointer -> offset 12 ("local")
	}
	msg = append(msg, srv...)
	rdlength := len(srv)

	// New path: resolves the compressed target against the full message.
	data, err := ParseRDATAInMessage(33, msg, rdataStart, rdlength)
	if err != nil {
		t.Fatalf("ParseRDATAInMessage(SRV) returned error: %v", err)
	}
	srvData, ok := data.(SRVData)
	if !ok {
		t.Fatalf("expected SRVData, got %T", data)
	}
	if srvData.Target != "host.local" {
		t.Errorf("Target = %q, want %q (compression pointer must resolve)", srvData.Target, "host.local")
	}
	if srvData.Port != 8080 {
		t.Errorf("Port = %d, want 8080", srvData.Port)
	}

	// Old path: parsing the isolated RDATA slice cannot follow the pointer
	// (it references message bytes outside the slice) — documents the bug the
	// new function fixes.
	if _, err := ParseRDATA(33, srv); err == nil {
		t.Errorf("ParseRDATA(SRV) on isolated compressed RDATA unexpectedly succeeded; expected failure (cannot resolve pointer)")
	}
}

// TestParseRDATAInMessage_CompressedPTRTarget proves the same for PTR targets.
func TestParseRDATAInMessage_CompressedPTRTarget(t *testing.T) {
	msg := make([]byte, 12)                                                    // header
	msg = append(msg, 0x05, '_', 'h', 't', 't', 'p', 0x04, '_', 't', 'c', 'p', // "_http._tcp"
		0x05, 'l', 'o', 'c', 'a', 'l', 0x00) // "...local." starting at offset 12

	rdataStart := len(msg)
	// PTR target "Inst" + pointer to "_http._tcp.local" at offset 12.
	ptr := []byte{0x04, 'I', 'n', 's', 't', 0xC0, 0x0C}
	msg = append(msg, ptr...)

	data, err := ParseRDATAInMessage(12, msg, rdataStart, len(ptr))
	if err != nil {
		t.Fatalf("ParseRDATAInMessage(PTR) returned error: %v", err)
	}
	name, ok := data.(string)
	if !ok {
		t.Fatalf("expected string, got %T", data)
	}
	if name != "Inst._http._tcp.local" {
		t.Errorf("PTR target = %q, want %q (compression pointer must resolve)", name, "Inst._http._tcp.local")
	}
}
