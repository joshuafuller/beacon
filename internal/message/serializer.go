// Package message implements DNS message serialization per RFC 1035.
package message

import (
	"encoding/binary"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// SerializeMessage converts a DNSMessage struct to wire-format bytes per RFC 1035 §4.1.
//
// The function serializes:
//   - Header (12 bytes): ID, Flags, QDCount, ANCount, NSCount, ARCount
//   - Question section: QNAME + QTYPE + QCLASS for each question
//   - Answer section: resource records
//   - Authority section: resource records
//   - Additional section: resource records
//
// Section counts in the header are set to match the actual serialized records.
//
// Parameters:
//   - msg: The DNS message to serialize
//
// Returns:
//   - []byte: Wire-format DNS message
//   - error: if serialization fails (nil resource record, invalid name, etc.)
func SerializeMessage(msg *DNSMessage) ([]byte, error) {
	if msg == nil {
		return nil, &errors.ValidationError{
			Field:   "DNSMessage",
			Value:   nil,
			Message: "cannot serialize nil message",
		}
	}

	// Pre-allocate buffer with reasonable initial capacity
	buf := make([]byte, 0, 512)

	// Reserve 12 bytes for header (will be filled after we know section counts)
	buf = append(buf, make([]byte, 12)...)

	// Serialize questions
	for _, q := range msg.Questions {
		qBytes, err := serializeQuestion(&q)
		if err != nil {
			return nil, err
		}
		buf = append(buf, qBytes...)
	}

	// Serialize answers
	for _, a := range msg.Answers {
		rr := answerToResourceRecord(&a)
		rrBytes, err := SerializeResourceRecord(rr)
		if err != nil {
			return nil, err
		}
		buf = append(buf, rrBytes...)
	}

	// Serialize authorities
	for _, a := range msg.Authorities {
		rr := answerToResourceRecord(&a)
		rrBytes, err := SerializeResourceRecord(rr)
		if err != nil {
			return nil, err
		}
		buf = append(buf, rrBytes...)
	}

	// Serialize additionals
	for _, a := range msg.Additionals {
		rr := answerToResourceRecord(&a)
		rrBytes, err := SerializeResourceRecord(rr)
		if err != nil {
			return nil, err
		}
		buf = append(buf, rrBytes...)
	}

	// Fill in header with actual counts
	binary.BigEndian.PutUint16(buf[0:2], msg.Header.ID)
	binary.BigEndian.PutUint16(buf[2:4], msg.Header.Flags)
	binary.BigEndian.PutUint16(buf[4:6], uint16(len(msg.Questions)))
	binary.BigEndian.PutUint16(buf[6:8], uint16(len(msg.Answers)))
	binary.BigEndian.PutUint16(buf[8:10], uint16(len(msg.Authorities)))
	binary.BigEndian.PutUint16(buf[10:12], uint16(len(msg.Additionals)))

	return buf, nil
}

// serializeQuestion serializes a DNS question to wire format per RFC 1035 §4.1.2.
func serializeQuestion(q *Question) ([]byte, error) {
	encodedName, err := EncodeName(q.QNAME)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 0, len(encodedName)+4)
	buf = append(buf, encodedName...)

	// QTYPE (2 bytes)
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, q.QTYPE)
	buf = append(buf, typeBytes...)

	// QCLASS (2 bytes)
	classBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classBytes, q.QCLASS)
	buf = append(buf, classBytes...)

	return buf, nil
}

// answerToResourceRecord converts an Answer (parsed format) to a ResourceRecord (builder format).
//
// The conversion maps:
//   - Answer.NAME   -> ResourceRecord.Name
//   - Answer.TYPE   -> ResourceRecord.Type
//   - Answer.CLASS  -> ResourceRecord.Class (cache-flush bit stripped, set as CacheFlush bool)
//   - Answer.TTL    -> ResourceRecord.TTL
//   - Answer.RDATA  -> ResourceRecord.Data
func answerToResourceRecord(a *Answer) *ResourceRecord {
	// RFC 6762 §10.2: Cache-flush bit is bit 15 of CLASS
	cacheFlush := (a.CLASS & 0x8000) != 0
	class := a.CLASS & 0x7FFF // Strip cache-flush bit

	return &ResourceRecord{
		Name:       a.NAME,
		Type:       protocol.RecordType(a.TYPE),
		Class:      protocol.DNSClass(class),
		TTL:        a.TTL,
		Data:       a.RDATA,
		CacheFlush: cacheFlush,
	}
}
