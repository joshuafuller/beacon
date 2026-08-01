package message

import (
	"encoding/binary"
	"testing"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestSerializeMessage_SectionCountBoundary is a regression test for the DNS
// section-count uint16 boundaries (RFC 1035 §4.1.1): QDCOUNT, ANCOUNT,
// NSCOUNT, and ARCOUNT are each 16-bit fields, so 65535 entries in a section
// is the largest valid count and 65536 must be rejected.
//
// Before this validation existed, SerializeMessage wrote each count with a
// bare `uint16(len(...))` conversion and no bounds check at all. For a
// section with 65536 entries, that conversion silently wraps around
// (65536 mod 65536 = 0): the header would claim the section is empty while
// the body still contains all 65536 serialized records.
func TestSerializeMessage_SectionCountBoundary(t *testing.T) {
	makeAnswers := func(n int) []Answer {
		answers := make([]Answer, n)
		for i := range answers {
			answers[i] = Answer{
				NAME:  "a.local",
				TYPE:  uint16(protocol.RecordTypeA),
				CLASS: uint16(protocol.ClassIN),
				TTL:   120,
				RDATA: []byte{127, 0, 0, 1},
			}
		}
		return answers
	}

	t.Run("65535 answers in a message is accepted (uint16 max)", func(t *testing.T) {
		msg := &DNSMessage{
			Header:  DNSHeader{ID: 1},
			Answers: makeAnswers(65535),
		}

		wire, err := SerializeMessage(msg)
		if err != nil {
			t.Fatalf("SerializeMessage() error = %v, want nil at uint16 max (65535)", err)
		}

		ancount := binary.BigEndian.Uint16(wire[6:8])
		if ancount != 65535 {
			t.Errorf("ANCOUNT = %d, want 65535", ancount)
		}
	})

	t.Run("65536 answers in a message is rejected, not silently wrapped", func(t *testing.T) {
		msg := &DNSMessage{
			Header:  DNSHeader{ID: 1},
			Answers: makeAnswers(65536),
		}

		_, err := SerializeMessage(msg)
		if err == nil {
			t.Fatalf("SerializeMessage() error = nil, want error for 65536 answers " +
				"(RFC 1035 §4.1.1: ANCOUNT is a wire-format uint16 field)")
		}
	})
}
