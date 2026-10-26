package message

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// A 12-byte header that claims 65535 records per section must not make the
// parser allocate memory for records that cannot fit in the packet (#48).
func TestParseMessage_OversizedHeaderCounts_DoNotAllocate(t *testing.T) {
	pkt := []byte{0, 0, 0x84, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	if _, err := ParseMessage(pkt); err == nil {
		t.Fatal("ParseMessage accepted counts that cannot fit in 12 bytes")
	}
	res := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = ParseMessage(pkt)
		}
	})
	if got := res.AllocedBytesPerOp(); got >= 4096 {
		t.Errorf("ParseMessage allocated %d bytes for a 12-byte packet, want < 4096", got)
	}
}

// Valid messages still parse when every count is exact.
func TestParseMessage_ExactCounts_StillParse(t *testing.T) {
	pkt := make([]byte, 12)
	binary.BigEndian.PutUint16(pkt[4:6], 1)
	pkt = append(pkt, 0, 0, 1, 0, 1) // root name, type A, class IN
	if _, err := ParseMessage(pkt); err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
}

// RFC 6763 §4.3: an instance label with a dot survives parse + re-encode (#50).
func TestInstanceLabelWithDot_RoundTrips(t *testing.T) {
	wire, err := EncodeServiceInstanceName("Printer v1.2", "_http._tcp.local")
	if err != nil {
		t.Fatal(err)
	}
	name, _, err := ParseName(wire, 0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := EncodeDomainName(name)
	if err != nil {
		t.Fatalf("EncodeDomainName(%q): %v", name, err)
	}
	if !bytes.Equal(again, wire) {
		t.Errorf("re-encoded %q = %v, want %v", name, again, wire)
	}
}

// EncodeDomainName keeps host names strict and accepts instance labels.
func TestEncodeDomainName(t *testing.T) {
	if _, err := EncodeDomainName("My Printer._http._tcp.local"); err != nil {
		t.Errorf("instance name rejected: %v", err)
	}
	if _, err := EncodeDomainName("bad name.local"); err == nil {
		t.Error("host name with a space accepted")
	}
}
