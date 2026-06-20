// Package message defines DNS message wire format structures per RFC 1035.
//
// This package implements the DNS message parsing and encoding requirements
// from spec.md, following RFC 1035 wire format specifications.
//
// Architecture: Per F-2, this package is internal (not importable by external users).
// Public API entities are in the beacon/querier package.
//
// PRIMARY TECHNICAL AUTHORITY: RFC 1035 (DNS wire format), RFC 6762 (mDNS extensions)
package message

// DNSHeader represents the DNS message header per RFC 1035 §4.1.1.
//
// The header is always 12 bytes and contains metadata about the message.
//
// Wire format (big-endian):
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      ID                       |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    QDCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ANCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    NSCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ARCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18
// FR-021: System MUST validate received responses have QR=1 per RFC 6762 §18.2
type DNSHeader struct {
	// ID is the transaction ID (16 bits).
	//
	// RFC 6762 §18.1: Multicast DNS messages SHOULD use ID = 0 for one-shot queries.
	// M1 uses random ID for potential future compatibility.
	ID uint16

	// Flags contains bit-packed header flags (16 bits).
	//
	// Bit layout per RFC 1035 §4.1.1:
	//   QR (bit 15): 0=query, 1=response
	//   OPCODE (bits 11-14): 0=standard query
	//   AA (bit 10): Authoritative Answer
	//   TC (bit 9): Truncated
	//   RD (bit 8): Recursion Desired
	//   RA (bit 7): Recursion Available
	//   Z (bits 4-6): Reserved (must be zero)
	//   RCODE (bits 0-3): Response Code
	//
	// RFC 6762 §18 requirements for queries:
	//   QR=0, OPCODE=0, AA=0, TC=0, RD=0, Z=0, RCODE=0
	//
	// RFC 6762 §18 requirements for responses:
	//   QR=1 (FR-021), RCODE=0 (FR-022, ignore non-zero)
	Flags uint16

	// QDCount is the number of entries in the Question section (16 bits).
	QDCount uint16

	// ANCount is the number of entries in the Answer section (16 bits).
	ANCount uint16

	// NSCount is the number of entries in the Authority section (16 bits).
	//
	// M1: Authority section is ignored per FR-010.
	NSCount uint16

	// ARCount is the number of entries in the Additional section (16 bits).
	//
	// M1: Additional section is ignored per FR-010.
	ARCount uint16
}

// IsQuery returns true if this is a query message (QR bit = 0) per RFC 1035 §4.1.1.
func (h *DNSHeader) IsQuery() bool {
	// QR bit is bit 15 (0x8000)
	return (h.Flags & 0x8000) == 0
}

// IsResponse returns true if this is a response message (QR bit = 1) per RFC 1035 §4.1.1.
//
// FR-021: System MUST validate received responses have QR=1 per RFC 6762 §18.2
func (h *DNSHeader) IsResponse() bool {
	// QR bit is bit 15 (0x8000)
	return (h.Flags & 0x8000) != 0
}

// GetRCODE extracts the response code from the Flags field per RFC 1035 §4.1.1.
//
// RCODE is bits 0-3 of the Flags field.
//
// FR-022: System MUST ignore responses with RCODE != 0 per RFC 6762 §18.11
func (h *DNSHeader) GetRCODE() uint8 {
	// RCODE is bits 0-3 (mask 0x000F)
	// G115: bounds checked - bitwise AND with 0x000F always produces value 0-15, safe for uint8
	return uint8(h.Flags & 0x000F) //nolint:gosec // G115: bounds checked
}

// GetOPCODE extracts the operation code from the Flags field per RFC 1035 §4.1.1.
//
// OPCODE is bits 11-14 of the Flags field.
//
// RFC 6762 §18.3: OPCODE MUST be zero on transmission.
func (h *DNSHeader) GetOPCODE() uint8 {
	// OPCODE is bits 11-14 (shift right 11, mask 0x0F)
	// G115: bounds checked - bitwise AND with 0x0F always produces value 0-15, safe for uint8
	return uint8((h.Flags >> 11) & 0x0F) //nolint:gosec // G115: bounds checked
}

// Question represents a DNS question section entry per RFC 1035 §4.1.2.
//
// The question section contains the query being asked.
//
// Wire format:
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                                               |
//	/                     QNAME                     /
//	/                                               /
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     QTYPE                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     QCLASS                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
type Question struct {
	// QNAME is the domain name being queried (variable length, label-encoded).
	//
	// RFC 1035 §3.1: Domain names are sequences of labels, each prefixed by a length byte.
	// Example: "printer.local" → 7printer5local0
	//
	// FR-003: System MUST validate queried names follow DNS naming rules (labels ≤63 bytes, total name ≤255 bytes)
	QNAME string

	// QTYPE is the query type (16 bits).
	//
	// M1 supports: A (1), PTR (12), SRV (33), TXT (16) per FR-002.
	QTYPE uint16

	// QCLASS is the query class (16 bits).
	//
	// RFC 1035: IN = 1 (Internet class)
	// RFC 6762 §5.4: QU bit (bit 15) = 0 for multicast queries (M1 default)
	//
	// M1 uses QCLASS = 0x0001 (IN, no QU bit per FR-001)
	QCLASS uint16
}

// Answer represents a DNS answer/authority/additional section entry per RFC 1035 §4.1.3.
//
// The answer section contains resource records returned by the responder.
//
// Wire format:
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                                               |
//	/                                               /
//	/                      NAME                     /
//	|                                               |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      TYPE                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     CLASS                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      TTL                      |
//	|                                               |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                   RDLENGTH                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
//	/                     RDATA                     /
//	/                                               /
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-010: System MUST extract Answer, Authority, and Additional sections (M1: Answer only)
type Answer struct {
	// NAME is the domain name this record refers to (variable length, can be compressed).
	//
	// RFC 1035 §4.1.4: Names can use compression pointers (high 2 bits = 11).
	//
	// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
	NAME string

	// TYPE is the resource record type (16 bits).
	//
	// M1 supports: A (1), PTR (12), SRV (33), TXT (16) per FR-002.
	TYPE uint16

	// CLASS is the resource record class (16 bits).
	//
	// RFC 1035: IN = 1 (Internet class)
	// RFC 6762 §10.2: Cache-flush bit (bit 15) can be set in responses
	//
	// M1: CLASS = 0x0001 (IN) or 0x8001 (IN + cache-flush, M1 ignores cache-flush bit)
	CLASS uint16

	// TTL is the time-to-live in seconds (32 bits).
	//
	// RFC 1035: TTL specifies how long the record can be cached.
	// M1: TTL is parsed but not used (caching deferred to M2).
	TTL uint32

	// RDLENGTH is the length of RDATA in bytes (16 bits).
	//
	// FR-011: System MUST validate response message format (RDLENGTH must match actual RDATA length)
	RDLENGTH uint16

	// RDATA is the type-specific resource data (variable length, RDLENGTH bytes).
	//
	// Format depends on TYPE:
	//   A (1):   4 bytes (IPv4 address)
	//   PTR (12): Domain name (label-encoded, can be compressed)
	//   SRV (33): 2 bytes priority + 2 bytes weight + 2 bytes port + domain name
	//   TXT (16): Text strings (length-prefixed strings)
	//
	// FR-012: System MUST decompress DNS names in RDATA (PTR, SRV target)
	RDATA []byte

	// RDATAOffset is the byte offset of RDATA within the full message it was
	// parsed from. It lets ParseRDATAInMessage resolve DNS name-compression
	// pointers in PTR/SRV targets against the whole message (FR-012); it is 0
	// for Answers not produced by ParseMessage.
	RDATAOffset int
}

// DNSMessage represents a complete DNS message per RFC 1035 §4.1.
//
// The message consists of a header and up to four sections: Question, Answer,
// Authority, and Additional.
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
type DNSMessage struct {
	// Header is the DNS message header (12 bytes, always present).
	Header DNSHeader

	// Questions is the question section (variable length, QDCount entries).
	//
	// M1 queries have 1 question per query.
	Questions []Question

	// Answers is the answer section (variable length, ANCount entries).
	//
	// FR-010: System MUST extract Answer section from responses
	Answers []Answer

	// Authorities is the authority section (variable length, NSCount entries).
	//
	// FR-010: M1 ignores Authority section (out of scope)
	Authorities []Answer

	// Additionals is the additional section (variable length, ARCount entries).
	//
	// FR-010: M1 ignores Additional section (deferred to M2 for cache pre-population)
	Additionals []Answer
}
