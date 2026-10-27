package transport

import (
	"context"
	"testing"
)

// RFC 6762 §11: "All Multicast DNS responses (including responses sent via
// unicast) SHOULD be sent with IP TTL set to 255." BCT III.3 warns otherwise.
func TestUDPv4Transport_IPTTLIs255(t *testing.T) {
	tr, err := NewUDPv4Transport()
	if err != nil {
		t.Skipf("no IPv4 multicast socket: %v", err)
	}
	defer func() { _ = tr.Close() }()

	if ttl, err := tr.ipv4Conn.MulticastTTL(); err != nil || ttl != 255 {
		t.Errorf("multicast TTL = %d (err %v), want 255", ttl, err)
	}
	if ttl, err := tr.ipv4Conn.TTL(); err != nil || ttl != 255 {
		t.Errorf("unicast TTL = %d (err %v), want 255", ttl, err)
	}
}

func TestUDPv6Transport_HopLimitIs255(t *testing.T) {
	tr, err := NewUDPv6Transport()
	if err != nil {
		t.Skipf("no IPv6 multicast socket: %v", err)
	}
	defer func() { _ = tr.Close() }()

	if hl, err := tr.ipv6Conn.MulticastHopLimit(); err != nil || hl != 255 {
		t.Errorf("multicast hop limit = %d (err %v), want 255", hl, err)
	}
	if hl, err := tr.ipv6Conn.HopLimit(); err != nil || hl != 255 {
		t.Errorf("unicast hop limit = %d (err %v), want 255", hl, err)
	}
}

// The responder passes dest == nil for "multicast to the mDNS group". A nil
// dest must reach 224.0.0.251:5353, not fail in WriteTo (found by Apple BCT:
// no multicast reply to any query was ever sent).
func TestUDPv4Transport_SendNilDestMulticasts(t *testing.T) {
	tr, err := NewUDPv4Transport()
	if err != nil {
		t.Skipf("no IPv4 multicast socket: %v", err)
	}
	defer func() { _ = tr.Close() }()
	pkt := []byte{0, 0, 0x84, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if err := tr.Send(context.Background(), pkt, nil); err != nil {
		t.Fatalf("Send(nil dest) = %v, want multicast send", err)
	}
}
