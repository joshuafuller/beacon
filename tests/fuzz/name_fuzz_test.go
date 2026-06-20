// Package fuzz provides fuzz testing for DNS message handling.
//
// This file fuzzes DNS name encoding/parsing (RFC 1035 §3.1, RFC 6763 §4.1.2)
// for two properties:
//   - ParseName MUST NOT panic on arbitrary bytes (NFR-003).
//   - Encode→Parse→Encode is idempotent at the wire-byte level for any name
//     that EncodeName accepts (a name that encodes must parse back to a name
//     that re-encodes to the identical bytes).
package fuzz

import (
	"bytes"
	"testing"

	"github.com/joshuafuller/beacon/internal/message"
)

// FuzzNameRoundTrip verifies that names accepted by EncodeName survive an
// encode→parse→encode cycle without loss, and that ParseName never panics on
// the resulting bytes.
//
// Comparing the wire bytes (rather than the strings) avoids baking in any
// assumption about case-folding or trailing-dot normalization; it asserts the
// codec is self-consistent, which is the property that matters on the wire.
//
// Run with: go test -fuzz=FuzzNameRoundTrip -fuzztime=20s ./tests/fuzz/
func FuzzNameRoundTrip(f *testing.F) {
	// Seed corpus: representative valid mDNS / DNS-SD names.
	for _, s := range []string{
		"test.local",
		"_http._tcp.local",
		"My Printer._http._tcp.local",
		"a.b.c.d.local",
		"x.local",
		"_services._dns-sd._udp.local",
		"my-device.local",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, name string) {
		enc1, err := message.EncodeName(name)
		if err != nil {
			// Input is not a name EncodeName accepts; nothing to round-trip.
			return
		}

		// Property 1: ParseName must not panic and must succeed on output that
		// EncodeName produced.
		parsed, _, err := message.ParseName(enc1, 0)
		if err != nil {
			t.Fatalf("EncodeName accepted %q but ParseName rejected its output (% x): %v", name, enc1, err)
		}

		// Property 2: re-encoding the parsed name yields identical wire bytes.
		enc2, err := message.EncodeName(parsed)
		if err != nil {
			t.Fatalf("round-trip name %q (from %q) failed to re-encode: %v", parsed, name, err)
		}
		if !bytes.Equal(enc1, enc2) {
			t.Fatalf("round-trip mismatch:\n  input:  %q -> % x\n  parsed: %q -> % x", name, enc1, parsed, enc2)
		}
	})
}

// FuzzParseNameRaw verifies ParseName never panics on arbitrary byte buffers
// and arbitrary (possibly out-of-range) offsets per NFR-003.
//
// Run with: go test -fuzz=FuzzParseNameRaw -fuzztime=20s ./tests/fuzz/
func FuzzParseNameRaw(f *testing.F) {
	f.Add([]byte{0x04, 't', 'e', 's', 't', 0x00}, 0)
	f.Add([]byte{0xc0, 0x00}, 0)               // compression pointer to self
	f.Add([]byte{0x05, 'l', 'o', 'c', 'a'}, 0) // length exceeds remaining bytes

	f.Fuzz(func(t *testing.T, buf []byte, offset int) {
		// Must not panic regardless of buffer contents or offset; errors are fine.
		_, _, _ = message.ParseName(buf, offset)
	})
}
