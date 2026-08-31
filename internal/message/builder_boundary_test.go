package message

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestBuildResponse_AnswerCountBoundary is a regression test for the ANCOUNT
// uint16 boundary (RFC 6762 §18): 65535 is the largest valid answer count,
// 65536 must be rejected rather than silently accepted.
//
// Before this validation existed, buildResponseHeader capped answerCount to
// 65535 when writing the ANCOUNT field, but BuildResponse still serialized
// every answer into the body. That produced a wire-format message whose
// header claimed 65535 answers while the body actually contained 65536,
// corrupting the message for any parser reading strictly off ANCOUNT.
func TestBuildResponse_AnswerCountBoundary(t *testing.T) {
	makeAnswers := func(n int) []*ResourceRecord {
		answers := make([]*ResourceRecord, n)
		for i := range answers {
			answers[i] = &ResourceRecord{
				Name:  "a.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{127, 0, 0, 1},
			}
		}
		return answers
	}

	t.Run("65535 answers is accepted (uint16 max)", func(t *testing.T) {
		response, err := BuildResponse(makeAnswers(65535))
		if err != nil {
			t.Fatalf("BuildResponse() error = %v, want nil at uint16 max (65535)", err)
		}

		ancount := binary.BigEndian.Uint16(response[6:8])
		if ancount != 65535 {
			t.Errorf("ANCOUNT = %d, want 65535", ancount)
		}
	})

	t.Run("65536 answers is rejected, not silently capped", func(t *testing.T) {
		_, err := BuildResponse(makeAnswers(65536))
		if err == nil {
			t.Fatalf("BuildResponse() error = nil, want ValidationError for 65536 answers " +
				"(RFC 6762 §18: ANCOUNT is a wire-format uint16 field)")
		}
	})
}

// TestSerializeResourceRecord_RDATALengthBoundary is a regression test for the
// RDLENGTH uint16 boundary (RFC 1035 §3.2.1): 65535 bytes of RDATA is the
// largest valid length, 65536 bytes must be rejected rather than silently
// accepted.
//
// Before this validation existed, the RDLENGTH field was capped to 65535 when
// rdataLen exceeded it, but the full (uncapped) rr.Data was still appended as
// RDATA. That produced a record whose RDLENGTH understated the actual RDATA
// size, corrupting the wire format for any parser reading strictly off
// RDLENGTH to find the start of the next record.
func TestSerializeResourceRecord_RDATALengthBoundary(t *testing.T) {
	makeRecord := func(dataLen int) *ResourceRecord {
		return &ResourceRecord{
			Name:  "a.local",
			Type:  protocol.RecordTypeTXT,
			Class: protocol.ClassIN,
			TTL:   120,
			Data:  bytes.Repeat([]byte{0x00}, dataLen),
		}
	}

	encodedName, err := EncodeName("a.local")
	if err != nil {
		t.Fatalf("EncodeName() error = %v, want nil", err)
	}
	rdlengthOffset := len(encodedName) + 8 // NAME + TYPE(2) + CLASS(2) + TTL(4)

	t.Run("65535-byte RDATA is accepted (uint16 max)", func(t *testing.T) {
		record, err := SerializeResourceRecord(makeRecord(65535))
		if err != nil {
			t.Fatalf("SerializeResourceRecord() error = %v, want nil at uint16 max (65535)", err)
		}

		wantLen := len(encodedName) + 10 + 65535
		if len(record) != wantLen {
			t.Fatalf("record length = %d, want %d", len(record), wantLen)
		}

		rdlength := binary.BigEndian.Uint16(record[rdlengthOffset : rdlengthOffset+2])
		if rdlength != 65535 {
			t.Errorf("RDLENGTH = %d, want 65535", rdlength)
		}
	})

	t.Run("65536-byte RDATA is rejected, not silently truncated", func(t *testing.T) {
		_, err := SerializeResourceRecord(makeRecord(65536))
		if err == nil {
			t.Fatalf("SerializeResourceRecord() error = nil, want ValidationError for 65536-byte RDATA " +
				"(RFC 1035 §3.2.1: RDLENGTH is a wire-format uint16 field)")
		}
	})
}
