// Package protocol defines mDNS protocol constants and validation logic
// per RFC 6762 (Multicast DNS).
//
// This package implements the protocol requirements from spec.md including:
//   - mDNS port and multicast address (FR-004)
//   - DNS record types (FR-002)
//   - RFC 6762 header field validation (FR-020, FR-021, FR-022)
//
// PRIMARY TECHNICAL AUTHORITY: RFC 6762 (Multicast DNS)
package protocol

import (
	"net"
	"time"
)

// mDNS Protocol Constants per RFC 6762
const (
	// Port is the mDNS port number (5353) per RFC 6762 §5.
	//
	// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
	Port = 5353

	// MulticastAddrIPv4 is the mDNS IPv4 multicast address (224.0.0.251) per RFC 6762 §5.
	//
	// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
	MulticastAddrIPv4 = "224.0.0.251"

	// MulticastAddrIPv6 is the deployed link-local mDNS IPv6 multicast address per RFC 6762 §22.
	//
	// RFC 6762 §22: FF02::FB is the deployed link-local IPv6 multicast group.
	MulticastAddrIPv6 = "FF02::FB"
)

// MulticastGroupIPv4 returns the mDNS IPv4 multicast group address.
//
// This is a convenience function for creating net.UDPAddr for mDNS multicast.
//
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
func MulticastGroupIPv4() *net.UDPAddr {
	return &net.UDPAddr{
		// This IS the protocol package that defines MulticastAddrIPv4 constant
		IP:   net.ParseIP(MulticastAddrIPv4), // nosemgrep: beacon-hardcoded-multicast-address
		Port: Port,
	}
}

// MulticastGroupIPv6 returns the mDNS IPv6 multicast group address.
//
// RFC 6762 §§20, 22: Dual-stack mDNS uses FF02::FB:5353 for its IPv6 zone.
func MulticastGroupIPv6() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.ParseIP(MulticastAddrIPv6), // nosemgrep: beacon-hardcoded-multicast-address
		Port: Port,
	}
}

// RecordType represents a DNS record type per RFC 1035 §3.2.2.
//
// M1 supports A, PTR, SRV, and TXT record types.
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
type RecordType uint16

// Supported DNS record types for M1 per RFC 1035 and RFC 2782 (SRV).
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
const (
	// RecordTypeA represents an A (IPv4 address) record per RFC 1035 §3.4.1.
	//
	// Type value: 1
	RecordTypeA RecordType = 1

	// RecordTypePTR represents a PTR (pointer/domain name) record per RFC 1035 §3.3.12.
	//
	// Used for service instance enumeration in DNS-SD.
	// Type value: 12
	RecordTypePTR RecordType = 12

	// RecordTypeTXT represents a TXT (text strings) record per RFC 1035 §3.3.14.
	//
	// Used for service metadata in DNS-SD.
	// Type value: 16
	RecordTypeTXT RecordType = 16

	// RecordTypeSRV represents an SRV (service location) record per RFC 2782.
	//
	// Used for service host/port information in DNS-SD.
	// Type value: 33
	RecordTypeSRV RecordType = 33

	// RecordTypeAAAA represents an AAAA (IPv6 address) record per RFC 3596.
	//
	// Type value: 28
	RecordTypeAAAA RecordType = 28

	// RecordTypeANY represents a query for all record types per RFC 1035 §3.2.3.
	//
	// RFC 6762 §8.1: "All probe queries SHOULD be done using... query type 'ANY' (255)"
	// Used for probing to detect conflicts for all record types.
	// Type value: 255
	RecordTypeANY RecordType = 255
)

// String returns the human-readable name for a RecordType.
func (rt RecordType) String() string {
	switch rt {
	case RecordTypeA:
		return "A"
	case RecordTypePTR:
		return "PTR"
	case RecordTypeTXT:
		return "TXT"
	case RecordTypeSRV:
		return "SRV"
	case RecordTypeAAAA:
		return "AAAA"
	case RecordTypeANY:
		return "ANY"
	default:
		return "UNKNOWN"
	}
}

// IsSupported returns true if the RecordType is supported.
//
// FR-002: System MUST support querying for A, AAAA, PTR, SRV, and TXT record types
// FR-014: System MUST return ValidationError for invalid query names or unsupported record types
// RFC 6762 §8.1: ANY type (255) is required for probing
func (rt RecordType) IsSupported() bool {
	switch rt {
	case RecordTypeA, RecordTypeAAAA, RecordTypePTR, RecordTypeTXT, RecordTypeSRV, RecordTypeANY:
		return true
	default:
		return false
	}
}

// DNSClass represents a DNS class per RFC 1035 §3.2.4.
//
// M1 uses the IN (Internet) class for all queries.
type DNSClass uint16

const (
	// ClassIN is the Internet (IN) class per RFC 1035 §3.2.4.
	//
	// Class value: 1
	ClassIN DNSClass = 1
)

// DNS Header Flags per RFC 1035 §4.1.1 and RFC 6762 §18
const (
	// FlagQR is the Query/Response bit (bit 15).
	//
	// RFC 6762 §18.2: In query messages the QR bit MUST be zero.
	// RFC 6762 §18.2: In response messages the QR bit MUST be one.
	//
	// FR-020: System MUST set DNS header fields per RFC 6762 §18 (QR=0 per §18.2)
	// FR-021: System MUST validate received responses have QR=1 per RFC 6762 §18.2
	FlagQR uint16 = 1 << 15 // 0x8000

	// FlagAA is the Authoritative Answer bit (bit 10).
	//
	// RFC 6762 §18.4: In query messages, the Authoritative Answer (AA) bit MUST be zero on transmission.
	//
	// FR-020: System MUST set DNS header fields per RFC 6762 §18 (AA=0 per §18.4)
	FlagAA uint16 = 1 << 10 // 0x0400

	// FlagTC is the Truncated bit (bit 9).
	//
	// RFC 6762 §18.5: In query messages, if the TC bit is set, it indicates that additional
	// Known-Answer records may be following shortly.
	//
	// M1 does not implement Known-Answer suppression, so TC=0.
	//
	// FR-020: System MUST set DNS header fields per RFC 6762 §18 (TC=0 per §18.5)
	FlagTC uint16 = 1 << 9 // 0x0200

	// FlagRD is the Recursion Desired bit (bit 8).
	//
	// RFC 6762 §18.6: In query messages, the Recursion Desired (RD) bit SHOULD be zero.
	//
	// M1 enforces RD=0 as MUST for simplicity.
	//
	// FR-020: System MUST set DNS header fields per RFC 6762 §18 (RD=0 per §18.6)
	FlagRD uint16 = 1 << 8 // 0x0100
)

// OPCODE values per RFC 1035 §4.1.1
const (
	// OpcodeQuery is the standard query OPCODE (0).
	//
	// RFC 6762 §18.3: In both multicast query and multicast response messages,
	// the OPCODE MUST be zero on transmission.
	//
	// FR-020: System MUST set DNS header fields per RFC 6762 §18 (OPCODE=0 per §18.3)
	OpcodeQuery uint16 = 0
)

// RCODE values per RFC 1035 §4.1.1
const (
	// RCodeNoError is the no error RCODE (0).
	//
	// RFC 6762 §18.11: Multicast DNS messages received with non-zero
	// Response Codes MUST be silently ignored.
	//
	// FR-022: System MUST ignore responses with RCODE != 0 per RFC 6762 §18.11
	RCodeNoError uint16 = 0
)

// DNS Name Constraints per RFC 1035 §3.1
const (
	// MaxLabelLength is the maximum length of a DNS label (63 bytes) per RFC 1035 §3.1.
	//
	// FR-003: System MUST validate queried names follow DNS naming rules (labels ≤63 bytes)
	MaxLabelLength = 63

	// MaxNameLength is the maximum length of a DNS name (255 bytes) per RFC 1035 §3.1.
	//
	// FR-003: System MUST validate queried names follow DNS naming rules (total name ≤255 bytes)
	MaxNameLength = 255

	// MaxCompressionPointers is the maximum number of compression pointer jumps allowed
	// when decompressing DNS names per RFC 1035 §4.1.4.
	//
	// This prevents infinite loops in malformed packets with circular compression pointers.
	//
	// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4 (message compression)
	MaxCompressionPointers = 256
)

// Compression pointer mask per RFC 1035 §4.1.4
const (
	// CompressionMask identifies a compression pointer (high 2 bits = 11).
	//
	// RFC 1035 §4.1.4: Message compression uses a pointer where the first two bits
	// are ones (0xC0), and the remaining 14 bits specify an offset.
	//
	// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4 (message compression)
	CompressionMask byte = 0xC0
)

// TTL values per RFC 6762 §10
const (
	// TTLService is the recommended TTL for service records (SRV, TXT) - 120 seconds per RFC 6762 §10.
	//
	// RFC 6762 §10: "The recommended TTL value for Multicast DNS resource records
	// with a host name as the resource record's name (e.g., A, AAAA, HINFO, etc.) or
	// contained within the resource record's rdata (e.g., SRV, reverse mapping PTR
	// record, etc.) is 120 seconds."
	//
	// FR-019: System MUST use RFC 6762 §10 TTL values (120s service records, 4500s hostname records)
	TTLService = 120

	// TTLHostname is the recommended TTL for hostname records (A, AAAA) - 4500 seconds (75 minutes) per RFC 6762 §10.
	//
	// RFC 6762 §10: "The recommended TTL value for other Multicast DNS resource records is 75 minutes (4500 seconds)."
	//
	// FR-019: System MUST use RFC 6762 §10 TTL values (120s service records, 4500s hostname records)
	TTLHostname = 4500
)

// Timing constants per RFC 6762 §8
const (
	// ProbeInterval is the interval between probe packets - 250 milliseconds per RFC 6762 §8.1.
	//
	// RFC 6762 §8.1: "When ready to send its Multicast DNS probe packet(s) the host should
	// first verify that the hardware address is ready by sending a standard ARP Request for
	// the desired IP address and then wait 250 milliseconds."
	//
	// F-4 REQ-F4-6: mDNS timing operations MUST use RFC-mandated delays from protocol package
	// Constitution Principle I: RFC MUST requirements cannot be configurable
	//
	// This IS the protocol package defining the constant - nosemgrep comment prevents
	// false positive from beacon-rfc-timing-local-const rule
	ProbeInterval = 250 * time.Millisecond // nosemgrep: beacon-rfc-timing-local-const
)
