// Package message implements DNS message construction per RFC 6762.
package message

// nosemgrep: beacon-external-dependencies
import (
	"crypto/rand" // Standard library, required for secure DNS query ID generation per gosec G404
	"encoding/binary"
	"math/big"
	"strings"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// BuildQuery constructs an mDNS query message per RFC 6762 §18.
//
// The query message consists of:
//   - Header: 12 bytes with flags set per RFC 6762 §18
//   - Question section: QNAME (variable), QTYPE (2 bytes), QCLASS (2 bytes)
//
// RFC 6762 §18 Query Requirements:
//
//	§18.2: QR bit MUST be zero (query)
//	§18.3: OPCODE MUST be zero (standard query)
//	§18.4: AA bit MUST be zero
//	§18.5: TC bit clear (no Known Answers in M1)
//	§18.6: RD bit SHOULD be zero (M1 enforces MUST)
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
// FR-003: System MUST validate queried names follow DNS naming rules
// FR-020: System MUST set DNS header fields per RFC 6762 §18
//
// Parameters:
//   - name: The DNS name to query (e.g., "printer.local")
//   - recordType: The DNS record type (A=1, PTR=12, TXT=16, SRV=33)
//
// Returns:
//   - query: The wire format DNS query message
//   - error: ValidationError if name or recordType is invalid
func BuildQuery(name string, recordType uint16) ([]byte, error) {
	// Validate record type per FR-002
	if !protocol.RecordType(recordType).IsSupported() {
		return nil, &errors.ValidationError{
			Field:   "recordType",
			Value:   recordType,
			Message: "unsupported record type (M1 supports A, PTR, SRV, TXT)",
		}
	}

	// Encode name per RFC 1035 §3.1 (this also validates per FR-003)
	encodedName, err := EncodeName(name)
	if err != nil {
		return nil, err // EncodeName already returns ValidationError
	}

	// Build DNS header per RFC 6762 §18
	header := buildQueryHeader()

	// Build question section per RFC 1035 §4.1.2
	question := buildQuestionSection(encodedName, recordType)

	// Combine header + question
	query := append(header, question...)

	return query, nil
}

// buildQueryHeader constructs a DNS header for an mDNS query per RFC 6762 §18.
//
// Header format (12 bytes):
//   - ID (2 bytes): Transaction ID
//   - Flags (2 bytes): QR, OPCODE, AA, TC, RD, RA, Z, RCODE
//   - QDCOUNT (2 bytes): Number of questions (always 1 for M1)
//   - ANCOUNT (2 bytes): Number of answers (always 0 for queries)
//   - NSCOUNT (2 bytes): Number of authority records (always 0 for queries)
//   - ARCOUNT (2 bytes): Number of additional records (always 0 for queries)
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18
func buildQueryHeader() []byte {
	header := make([]byte, 12)

	// ID: RFC 6762 §18.1 suggests 0, but M1 uses random ID for future compatibility
	// Use crypto/rand for cryptographically secure random number generation (G404)
	idBig, err := rand.Int(rand.Reader, big.NewInt(65536))
	if err != nil {
		// Fallback to 0 if crypto/rand fails (should not happen in practice)
		idBig = big.NewInt(0)
	}
	// G115: rand.Int is called with upper bound 65536, so result is in range [0, 65535]
	// Safe conversion to uint16 using modulo to ensure no overflow
	id := uint16(idBig.Uint64() % 65536) //nolint:gosec // G115: rand.Int bounds upper limit to 65536
	binary.BigEndian.PutUint16(header[0:2], id)

	// Flags: Set per RFC 6762 §18
	// QR=0 (§18.2), OPCODE=0 (§18.3), AA=0 (§18.4), TC=0 (§18.5),
	// RD=0 (§18.6), RA=0, Z=0, RCODE=0
	flags := uint16(0x0000)
	binary.BigEndian.PutUint16(header[2:4], flags)

	// QDCOUNT: 1 question
	binary.BigEndian.PutUint16(header[4:6], 1)

	// ANCOUNT: 0 answers (queries don't have answers)
	binary.BigEndian.PutUint16(header[6:8], 0)

	// NSCOUNT: 0 authority records
	binary.BigEndian.PutUint16(header[8:10], 0)

	// ARCOUNT: 0 additional records (M1 doesn't send additional records)
	binary.BigEndian.PutUint16(header[10:12], 0)

	return header
}

// buildQuestionSection constructs a DNS question section per RFC 1035 §4.1.2.
//
// Question format:
//   - QNAME (variable): Encoded domain name (length-prefixed labels)
//   - QTYPE (2 bytes): Query type (A, PTR, SRV, TXT)
//   - QCLASS (2 bytes): Query class (IN=1, QU bit=0 for multicast)
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
func buildQuestionSection(encodedName []byte, recordType uint16) []byte {
	// Question section size: name + QTYPE (2) + QCLASS (2)
	question := make([]byte, 0, len(encodedName)+4)

	// QNAME: Already encoded by EncodeName
	question = append(question, encodedName...)

	// QTYPE: Record type (2 bytes, big-endian)
	qtype := make([]byte, 2)
	binary.BigEndian.PutUint16(qtype, recordType)
	question = append(question, qtype...)

	// QCLASS: IN (1) with QU bit=0 per RFC 6762 §5.4
	// M1 uses standard multicast queries (QU=0)
	qclass := make([]byte, 2)
	binary.BigEndian.PutUint16(qclass, uint16(protocol.ClassIN)) // 0x0001
	question = append(question, qclass...)

	return question
}

// BuildResponse constructs an mDNS response message per RFC 6762 §18.
//
// The response message consists of:
//   - Header: 12 bytes with flags set per RFC 6762 §18 (QR=1, AA=1)
//   - Answer section: Variable, contains the resource records
//
// RFC 6762 §18 Response Requirements:
//
//	§18.2: QR bit MUST be one (response)
//	§18.3: OPCODE MUST be zero (standard query)
//	§18.4: AA bit MUST be one (authoritative answer)
//	§18.11: RCODE MUST be zero
//
// FR-023: System MUST construct valid mDNS response messages per RFC 6762
// T012: Implement BuildResponse() to make T011 tests pass
//
// Parameters:
//   - answers: The resource records to include in the answer section
//
// Returns:
//   - []byte: The wire format DNS response message
//   - error: ValidationError if answers are invalid
func BuildResponse(answers []*ResourceRecord) ([]byte, error) {
	// Build response header
	header := buildResponseHeader(len(answers))

	// Start with header
	response := make([]byte, 0, 512) // Pre-allocate reasonable size
	response = append(response, header...)

	// Add answer records
	for _, answer := range answers {
		answerBytes, err := SerializeResourceRecord(answer)
		if err != nil {
			return nil, err
		}
		response = append(response, answerBytes...)
	}

	return response, nil
}

// buildResponseHeader constructs a DNS header for an mDNS response per RFC 6762 §18.
//
// Header format (12 bytes):
//   - ID (2 bytes): Transaction ID (0 for responses per RFC 6762 §18.1)
//   - Flags (2 bytes): QR=1, AA=1, OPCODE=0, RCODE=0
//   - QDCOUNT (2 bytes): Number of questions (0 for unsolicited responses)
//   - ANCOUNT (2 bytes): Number of answers
//   - NSCOUNT (2 bytes): Number of authority records (0)
//   - ARCOUNT (2 bytes): Number of additional records (0 for now)
//
// FR-023: System MUST set response header fields per RFC 6762 §18
// T012: Build response headers with QR=1, AA=1
func buildResponseHeader(answerCount int) []byte {
	header := make([]byte, 12)

	// ID: RFC 6762 §18.1 recommends 0 for responses
	binary.BigEndian.PutUint16(header[0:2], 0)

	// Flags: QR=1 (response), AA=1 (authoritative), OPCODE=0, RCODE=0
	// QR bit (bit 15): 1 = response
	// AA bit (bit 10): 1 = authoritative
	flags := protocol.FlagQR | protocol.FlagAA
	binary.BigEndian.PutUint16(header[2:4], flags)

	// QDCOUNT: 0 questions (unsolicited response)
	binary.BigEndian.PutUint16(header[4:6], 0)

	// ANCOUNT: Number of answer records
	// G115: RFC 6762 §4.3 specifies ANCOUNT as uint16, max 65535. DNS message size limit
	// (9000 bytes per RFC 6762) ensures answerCount never exceeds uint16.
	// Defensive bounds check for safety.
	if answerCount > 65535 { //nolint:gosec // G115: bounds checked, max message size 9000 bytes
		answerCount = 65535 // Cap at maximum uint16
	}
	binary.BigEndian.PutUint16(header[6:8], uint16(answerCount))

	// NSCOUNT: 0 authority records
	binary.BigEndian.PutUint16(header[8:10], 0)

	// ARCOUNT: 0 additional records (for now)
	binary.BigEndian.PutUint16(header[10:12], 0)

	return header
}

// serializeResourceRecord serializes a resource record to wire format.
//
// Resource record format per RFC 1035 §3.2.1:
//   - NAME (variable): Domain name
//   - TYPE (2 bytes): Record type (A, PTR, SRV, TXT)
//   - CLASS (2 bytes): Class (IN=1), with cache-flush bit if set
//   - TTL (4 bytes): Time to live in seconds
//   - RDLENGTH (2 bytes): Length of RDATA
//   - RDATA (variable): Record data
//
// RFC 6762 §10.2: Cache-flush bit (bit 15 of CLASS) for unique records
//
// T012: Serialize resource records with cache-flush support
func SerializeResourceRecord(rr *ResourceRecord) ([]byte, error) {
	if rr == nil {
		return nil, &errors.ValidationError{
			Field:   "ResourceRecord",
			Value:   nil,
			Message: "cannot serialize nil resource record",
		}
	}

	// Encode the domain name
	// Detect service instance names per RFC 6763 §4.3:
	// If the name contains a service type pattern (_service._proto.local),
	// split it and encode the instance portion separately to allow UTF-8/spaces.
	var encodedName []byte
	var err error

	// Check if this is a service instance name format: "instance._service._proto.local"
	// Pattern: contains "._" which indicates a service type
	if strings.Contains(rr.Name, "._") {
		// Split into instance name and service type
		parts := strings.SplitN(rr.Name, "._", 2)
		if len(parts) == 2 {
			// parts[0] = instance name (may contain spaces/UTF-8)
			// parts[1] = service type (e.g., "http._tcp.local")
			instanceName := parts[0]
			serviceType := "_" + parts[1] // Restore leading underscore

			// Use special encoding for service instance names
			encodedName, err = EncodeServiceInstanceName(instanceName, serviceType)
			if err != nil {
				return nil, err
			}
		} else {
			// Fallback to normal encoding
			encodedName, err = EncodeName(rr.Name)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Normal DNS name (not a service instance)
		encodedName, err = EncodeName(rr.Name)
		if err != nil {
			return nil, err
		}
	}

	// Calculate total size
	recordSize := len(encodedName) + 10 + len(rr.Data) // name + type(2) + class(2) + ttl(4) + rdlength(2) + rdata

	record := make([]byte, 0, recordSize)

	// NAME
	record = append(record, encodedName...)

	// TYPE (2 bytes)
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, uint16(rr.Type))
	record = append(record, typeBytes...)

	// CLASS (2 bytes) with cache-flush bit if requested
	class := uint16(rr.Class)
	if rr.CacheFlush {
		// Set cache-flush bit (bit 15) per RFC 6762 §10.2
		class |= 0x8000
	}
	classBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classBytes, class)
	record = append(record, classBytes...)

	// TTL (4 bytes)
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, rr.TTL)
	record = append(record, ttlBytes...)

	// RDLENGTH (2 bytes)
	// G115: RFC 1035 §3.2.1 specifies RDLENGTH as uint16, max 65535. DNS message size
	// limit (9000 bytes per RFC 6762) ensures rdata length never exceeds uint16.
	// Defensive bounds check for safety.
	rdataLen := len(rr.Data)
	if rdataLen > 65535 { //nolint:gosec // G115: bounds checked, max message size 9000 bytes
		rdataLen = 65535 // Cap at maximum uint16
	}
	rdlengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(rdlengthBytes, uint16(rdataLen))
	record = append(record, rdlengthBytes...)

	// RDATA
	record = append(record, rr.Data...)

	return record, nil
}

// ResourceRecord represents a DNS resource record for response building.
//
// This type is used by the response builder to serialize records into wire format.
// Full implementation will be in internal/records/record_set.go (T015).
//
// T012: Minimal ResourceRecord type for response building
// T015: Will be replaced by full ResourceRecordSet implementation
type ResourceRecord struct {
	Name       string              // Domain name (e.g., "printer.local")
	Type       protocol.RecordType // Record type (A, PTR, SRV, TXT)
	Class      protocol.DNSClass   // Class (usually IN=1)
	TTL        uint32              // Time to live in seconds
	Data       []byte              // Record data (wire format)
	CacheFlush bool                // RFC 6762 §10.2 cache-flush bit for unique records
}
