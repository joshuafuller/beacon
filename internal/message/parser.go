// Package message implements DNS message parsing per RFC 1035.
package message

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/joshuafuller/beacon/internal/errors"
)

// SRVData represents SRV record data per RFC 2782.
//
// SRV records provide the location of services (hostname and port).
type SRVData struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
}

// ParseMessage parses a complete DNS message from wire format per RFC 1035 §4.1.
//
// The message consists of:
//   - Header: 12 bytes (always present)
//   - Question section: Variable length (QDCOUNT entries)
//   - Answer section: Variable length (ANCOUNT entries)
//   - Authority section: Variable length (NSCOUNT entries, M1 ignores)
//   - Additional section: Variable length (ARCOUNT entries, M1 ignores)
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-011: System MUST validate response message format and discard malformed packets
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
//
// Parameters:
//   - msg: The complete DNS message buffer
//
// Returns:
//   - message: The parsed DNS message structure
//   - error: WireFormatError if the message is malformed
func ParseMessage(msg []byte) (*DNSMessage, error) {
	// Parse header
	header, err := ParseHeader(msg)
	if err != nil {
		return nil, err
	}

	offset := 12 // Header is always 12 bytes

	// Parse question section
	questions := make([]Question, header.QDCount)
	for i := uint16(0); i < header.QDCount; i++ {
		question, newOffset, err := ParseQuestion(msg, offset)
		if err != nil {
			return nil, err
		}
		questions[i] = question
		offset = newOffset
	}

	// Parse answer section
	answers := make([]Answer, header.ANCount)
	for i := uint16(0); i < header.ANCount; i++ {
		answer, newOffset, err := ParseAnswer(msg, offset)
		if err != nil {
			return nil, err
		}
		answers[i] = answer
		offset = newOffset
	}

	// Parse authority section (M1: ignored per FR-010, but we parse for completeness)
	authorities := make([]Answer, header.NSCount)
	for i := uint16(0); i < header.NSCount; i++ {
		authority, newOffset, err := ParseAnswer(msg, offset)
		if err != nil {
			return nil, err
		}
		authorities[i] = authority
		offset = newOffset
	}

	// Parse additional section (M1: ignored per FR-010, but we parse for completeness)
	additionals := make([]Answer, header.ARCount)
	for i := uint16(0); i < header.ARCount; i++ {
		additional, newOffset, err := ParseAnswer(msg, offset)
		if err != nil {
			return nil, err
		}
		additionals[i] = additional
		offset = newOffset
	}

	return &DNSMessage{
		Header:      header,
		Questions:   questions,
		Answers:     answers,
		Authorities: authorities,
		Additionals: additionals,
	}, nil
}

// ParseHeader parses the DNS message header per RFC 1035 §4.1.1.
//
// Header format (12 bytes):
//   - ID (2 bytes): Transaction ID
//   - Flags (2 bytes): QR, OPCODE, AA, TC, RD, RA, Z, RCODE
//   - QDCOUNT (2 bytes): Number of questions
//   - ANCOUNT (2 bytes): Number of answers
//   - NSCOUNT (2 bytes): Number of authority records
//   - ARCOUNT (2 bytes): Number of additional records
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-011: System MUST validate response message format and discard malformed packets
//
// Parameters:
//   - msg: The complete DNS message buffer (must be at least 12 bytes)
//
// Returns:
//   - header: The parsed DNS header
//   - error: WireFormatError if the header is malformed
func ParseHeader(msg []byte) (DNSHeader, error) {
	if len(msg) < 12 {
		return DNSHeader{}, &errors.WireFormatError{
			Operation: "parse header",
			Offset:    0,
			Message:   fmt.Sprintf("message too short for header: %d bytes, expected at least 12", len(msg)),
		}
	}

	header := DNSHeader{
		ID:      binary.BigEndian.Uint16(msg[0:2]),
		Flags:   binary.BigEndian.Uint16(msg[2:4]),
		QDCount: binary.BigEndian.Uint16(msg[4:6]),
		ANCount: binary.BigEndian.Uint16(msg[6:8]),
		NSCount: binary.BigEndian.Uint16(msg[8:10]),
		ARCount: binary.BigEndian.Uint16(msg[10:12]),
	}

	return header, nil
}

// ParseQuestion parses a DNS question section entry per RFC 1035 §4.1.2.
//
// Question format:
//   - QNAME (variable): Domain name (label-encoded, can be compressed)
//   - QTYPE (2 bytes): Query type
//   - QCLASS (2 bytes): Query class
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
//
// Parameters:
//   - msg: The complete DNS message buffer
//   - offset: The starting offset of this question entry
//
// Returns:
//   - question: The parsed question
//   - newOffset: The offset immediately after this question entry
//   - error: WireFormatError if the question is malformed
func ParseQuestion(msg []byte, offset int) (Question, int, error) {
	// Parse QNAME
	qname, newOffset, err := ParseName(msg, offset)
	if err != nil {
		return Question{}, offset, err
	}

	// Check bounds for QTYPE and QCLASS (4 bytes)
	if newOffset+4 > len(msg) {
		return Question{}, offset, &errors.WireFormatError{
			Operation: "parse question",
			Offset:    newOffset,
			Message:   "truncated question: not enough bytes for QTYPE and QCLASS",
		}
	}

	// Parse QTYPE
	qtype := binary.BigEndian.Uint16(msg[newOffset : newOffset+2])

	// Parse QCLASS
	qclass := binary.BigEndian.Uint16(msg[newOffset+2 : newOffset+4])

	question := Question{
		QNAME:  qname,
		QTYPE:  qtype,
		QCLASS: qclass,
	}

	return question, newOffset + 4, nil
}

// ParseAnswer parses a DNS answer/authority/additional section entry per RFC 1035 §4.1.3.
//
// Answer format:
//   - NAME (variable): Domain name (label-encoded, can be compressed)
//   - TYPE (2 bytes): Record type
//   - CLASS (2 bytes): Record class
//   - TTL (4 bytes): Time-to-live
//   - RDLENGTH (2 bytes): Resource data length
//   - RDATA (variable): Resource data (RDLENGTH bytes)
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
//
// Parameters:
//   - msg: The complete DNS message buffer
//   - offset: The starting offset of this answer entry
//
// Returns:
//   - answer: The parsed answer
//   - newOffset: The offset immediately after this answer entry
//   - error: WireFormatError if the answer is malformed
func ParseAnswer(msg []byte, offset int) (Answer, int, error) {
	// Parse NAME
	name, newOffset, err := ParseName(msg, offset)
	if err != nil {
		return Answer{}, offset, err
	}

	// Check bounds for TYPE, CLASS, TTL, RDLENGTH (10 bytes)
	if newOffset+10 > len(msg) {
		return Answer{}, offset, &errors.WireFormatError{
			Operation: "parse answer",
			Offset:    newOffset,
			Message:   "truncated answer: not enough bytes for fixed fields",
		}
	}

	// Parse TYPE
	rtype := binary.BigEndian.Uint16(msg[newOffset : newOffset+2])

	// Parse CLASS
	class := binary.BigEndian.Uint16(msg[newOffset+2 : newOffset+4])

	// Parse TTL
	ttl := binary.BigEndian.Uint32(msg[newOffset+4 : newOffset+8])

	// Parse RDLENGTH
	rdlength := binary.BigEndian.Uint16(msg[newOffset+8 : newOffset+10])

	newOffset += 10

	// Check bounds for RDATA
	if newOffset+int(rdlength) > len(msg) {
		return Answer{}, offset, &errors.WireFormatError{
			Operation: "parse answer",
			Offset:    newOffset,
			Message:   fmt.Sprintf("truncated RDATA: expected %d bytes, only %d available", rdlength, len(msg)-newOffset),
		}
	}

	// Extract RDATA
	rdata := make([]byte, rdlength)
	copy(rdata, msg[newOffset:newOffset+int(rdlength)])

	answer := Answer{
		NAME:        name,
		TYPE:        rtype,
		CLASS:       class,
		TTL:         ttl,
		RDLENGTH:    rdlength,
		RDATA:       rdata,
		RDATAOffset: newOffset, // RDATA begins here within msg (after the 10 fixed bytes)
	}

	return answer, newOffset + int(rdlength), nil
}

// ParseRDATA parses type-specific RDATA into Go types per RFC 1035.
//
// Supported types (per FR-002):
//   - A (1): IPv4 address → net.IP
//   - PTR (12): Domain name → string
//   - TXT (16): Text strings → []string
//   - SRV (33): Service location → SRVData
//
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-012: System MUST decompress DNS names in RDATA (PTR, SRV target)
//
// Parameters:
//   - recordType: The DNS record type (A, PTR, SRV, TXT)
//   - rdata: The raw RDATA bytes
//
// Returns:
//   - parsed: Type-specific parsed data (net.IP, string, []string, or SRVData)
//   - error: WireFormatError if RDATA is malformed
func ParseRDATA(recordType uint16, rdata []byte) (interface{}, error) {
	// No surrounding-message context: treat rdata as a self-contained buffer.
	// DNS name-compression pointers inside PTR/SRV targets reference earlier
	// bytes of the full message and therefore CANNOT be resolved here — callers
	// that have the full message should use ParseRDATAInMessage so that
	// responders which compress (e.g. Avahi/Bonjour) interoperate (FR-012).
	return ParseRDATAInMessage(recordType, rdata, 0, len(rdata))
}

// ParseRDATAInMessage parses the type-specific RDATA located at rdataStart (of
// length rdlength) within the full message msg. Unlike ParseRDATA, it resolves
// DNS name-compression pointers in PTR and SRV targets against msg, which is
// required to interoperate with responders that compress those names (FR-012).
//
// For A and TXT records (no embedded names) it is equivalent to ParseRDATA.
//
// Parameters:
//   - recordType: The DNS record type (A, PTR, SRV, TXT)
//   - msg: The complete DNS message (for compression-pointer resolution)
//   - rdataStart: Byte offset of RDATA within msg
//   - rdlength: Length of RDATA in bytes
func ParseRDATAInMessage(recordType uint16, msg []byte, rdataStart, rdlength int) (interface{}, error) {
	if rdataStart < 0 || rdlength < 0 || rdataStart+rdlength > len(msg) {
		return nil, &errors.WireFormatError{
			Operation: "parse RDATA",
			Offset:    rdataStart,
			Message:   fmt.Sprintf("RDATA out of bounds: start=%d len=%d msg=%d", rdataStart, rdlength, len(msg)),
		}
	}
	rdata := msg[rdataStart : rdataStart+rdlength]

	switch recordType {
	case 1: // A record: IPv4 address (4 bytes)
		if len(rdata) != 4 {
			return nil, &errors.WireFormatError{
				Operation: "parse A record",
				Offset:    rdataStart,
				Message:   fmt.Sprintf("invalid A record length: %d bytes, expected 4", len(rdata)),
			}
		}
		return net.IPv4(rdata[0], rdata[1], rdata[2], rdata[3]), nil

	case 12: // PTR record: Domain name (may be compressed against msg)
		name, _, err := ParseName(msg, rdataStart)
		if err != nil {
			return nil, err
		}
		return name, nil

	case 16: // TXT record: Text strings
		var strings []string
		offset := 0
		for offset < len(rdata) {
			// Each string is length-prefixed
			if offset >= len(rdata) {
				break
			}
			length := int(rdata[offset])
			offset++

			if offset+length > len(rdata) {
				return nil, &errors.WireFormatError{
					Operation: "parse TXT record",
					Offset:    offset,
					Message:   fmt.Sprintf("truncated TXT string: expected %d bytes, only %d available", length, len(rdata)-offset),
				}
			}

			str := string(rdata[offset : offset+length])
			strings = append(strings, str)
			offset += length
		}
		return strings, nil

	case 33: // SRV record: Priority, Weight, Port, Target (target may be compressed)
		if len(rdata) < 6 {
			return nil, &errors.WireFormatError{
				Operation: "parse SRV record",
				Offset:    rdataStart,
				Message:   fmt.Sprintf("truncated SRV record: %d bytes, expected at least 6", len(rdata)),
			}
		}

		priority := binary.BigEndian.Uint16(rdata[0:2])
		weight := binary.BigEndian.Uint16(rdata[2:4])
		port := binary.BigEndian.Uint16(rdata[4:6])

		// Target is a domain name starting 6 bytes into RDATA; resolve against
		// the full message so compression pointers work.
		target, _, err := ParseName(msg, rdataStart+6)
		if err != nil {
			return nil, err
		}

		return SRVData{
			Priority: priority,
			Weight:   weight,
			Port:     port,
			Target:   target,
		}, nil

	default:
		return nil, &errors.WireFormatError{
			Operation: "parse RDATA",
			Offset:    rdataStart,
			Message:   fmt.Sprintf("unsupported record type: %d", recordType),
		}
	}
}
