package querier

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Wire-level RFC 6762 / RFC 6763 conformance tests for the querier
// (issues #42, #43, #47, #49).

// buildAResponse builds a one-answer A response for name.
func buildAResponse(t *testing.T, name string, ip [4]byte) []byte {
	t.Helper()
	pkt := make([]byte, 12)
	binary.BigEndian.PutUint16(pkt[2:4], 0x8400)
	binary.BigEndian.PutUint16(pkt[6:8], 1)
	enc, err := message.EncodeName(name)
	if err != nil {
		t.Fatalf("encode %q: %v", name, err)
	}
	pkt = append(pkt, enc...)
	pkt = binary.BigEndian.AppendUint16(pkt, 1)
	pkt = binary.BigEndian.AppendUint16(pkt, 0x8001)
	pkt = binary.BigEndian.AppendUint32(pkt, 120)
	pkt = binary.BigEndian.AppendUint16(pkt, 4)
	return append(pkt, ip[:]...)
}

func shortCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	t.Cleanup(cancel)
	return ctx
}

// Answers whose owner name differs from the question are not results (#43).
func TestCollectResponses_FiltersByName(t *testing.T) {
	q := &Querier{responseChan: make(chan []byte, 10)}
	q.responseChan <- buildAResponse(t, "someone-else.local", [4]byte{10, 0, 0, 9})
	q.responseChan <- buildAResponse(t, "PRINTER.local", [4]byte{10, 0, 0, 7})

	resp, err := q.collectResponses(shortCtx(t), "printer.local", RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Records) != 1 || resp.Records[0].Name != "PRINTER.local" {
		t.Errorf("records = %+v, want only the PRINTER.local answer", resp.Records)
	}
}

// Packets queued before Query sends are stale and are not results (#43).
func TestQuery_DiscardsPacketsQueuedBeforeSend(t *testing.T) {
	q := &Querier{
		transport:      transport.NewMockTransport(),
		responseChan:   make(chan []byte, 10),
		defaultTimeout: 50 * time.Millisecond,
	}
	q.responseChan <- buildAResponse(t, "printer.local", [4]byte{10, 0, 0, 1})

	resp, err := q.Query(context.Background(), "printer.local", RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Records) != 0 {
		t.Errorf("records = %+v, want none (packet predates the query)", resp.Records)
	}
}

// RFC 6763 §4.3: instance labels may contain spaces (#42).
func TestQuery_InstanceNameWithSpace(t *testing.T) {
	mt := transport.NewMockTransport()
	q := &Querier{transport: mt, responseChan: make(chan []byte, 1), defaultTimeout: 10 * time.Millisecond}

	if _, err := q.Query(context.Background(), "My Printer._http._tcp.local", RecordTypeSRV); err != nil {
		t.Fatalf("Query: %v", err)
	}
	calls := mt.SendCalls()
	if len(calls) != 1 {
		t.Fatalf("sent %d packets, want 1", len(calls))
	}
	pkt := calls[0].Packet
	if pkt[12] != 10 || string(pkt[13:23]) != "My Printer" {
		t.Errorf("first question label = %q (len %d), want 10-byte \"My Printer\"", pkt[13:13+int(pkt[12])], pkt[12])
	}

	// Host names stay strict.
	if _, err := q.Query(context.Background(), "bad name.local", RecordTypeA); err == nil {
		t.Error("A query for \"bad name.local\" accepted, want validation error")
	}
}

// RFC 6762 §6: responses from a source port other than 5353 MUST be ignored (#47).
func TestShouldDrop_SourcePort(t *testing.T) {
	q := &Querier{}
	pkt := buildAResponse(t, "printer.local", [4]byte{192, 168, 1, 9})
	if !q.shouldDrop(pkt, &net.UDPAddr{IP: net.IPv4(192, 168, 1, 9), Port: 40000}, false) {
		t.Error("response from port 40000 accepted, want dropped")
	}
	if q.shouldDrop(pkt, &net.UDPAddr{IP: net.IPv4(192, 168, 1, 9), Port: 5353}, false) {
		t.Error("response from port 5353 dropped, want accepted")
	}
}

// RFC 6762 §11: on-link sources are valid even when not RFC 1918 (#49).
func TestShouldDrop_OnLinkNonPrivateSource(t *testing.T) {
	orig := localInterfaceAddrs
	t.Cleanup(func() { localInterfaceAddrs = orig; resetOnLinkCache() })
	_, cgnat, _ := net.ParseCIDR("100.64.1.0/24")
	localInterfaceAddrs = func() ([]net.Addr, error) { return []net.Addr{cgnat}, nil }
	resetOnLinkCache()

	q := &Querier{}
	pkt := buildAResponse(t, "printer.local", [4]byte{100, 64, 1, 5})
	if q.shouldDrop(pkt, &net.UDPAddr{IP: net.IPv4(100, 64, 1, 5), Port: 5353}, false) {
		t.Error("on-link 100.64.1.5 dropped, want accepted")
	}
	if !q.shouldDrop(pkt, &net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 5353}, false) {
		t.Error("off-link 8.8.8.8 accepted, want dropped")
	}
}

// RFC 6763 §4.3: DiscoverServices resolves an instance whose name has a space
// through the SRV/TXT/A fallback queries (#42).
func TestDiscoverServices_InstanceWithSpace_Fallback(t *testing.T) {
	q := newQuerierWithBlockingMock(t)
	defer q.Close()

	const serviceType = "_beacon-space._tcp.local"
	const instance = "My Printer._beacon-space._tcp.local"
	const host = "host.local"
	const port = uint16(8080)

	ptr, _ := message.EncodeDomainName(instance)
	defer feedResponseUntilStopped(q, buildValidResponsePacket(serviceType, 12, ptr))()

	srv := make([]byte, 6)
	binary.BigEndian.PutUint16(srv[4:6], port)
	hostEnc, _ := message.EncodeName(host)
	defer feedResponseUntilStopped(q, buildValidResponsePacket(instance, 33, append(srv, hostEnc...)))()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	services, err := q.DiscoverServices(ctx, serviceType)
	if err != nil {
		t.Fatal(err)
	}
	s := findServiceByInstance(services, "My Printer")
	if s == nil {
		t.Fatalf("instance \"My Printer\" not found in %+v", services)
	}
	if s.Hostname != host || s.Port != port {
		t.Errorf("Hostname/Port = %q/%d, want %q/%d", s.Hostname, s.Port, host, port)
	}
}

// RFC 6762 §16: bundled additionals whose names differ only in case are used (#45).
func TestDiscoverServices_UpperCaseAdditionals(t *testing.T) {
	q := newQuerierWithBlockingMock(t)
	defer q.Close()

	packet := buildBundledPTRResponse("_HTTP._TCP.LOCAL", "INST._HTTP._TCP.LOCAL",
		"HOST.LOCAL", 8080, [4]byte{192, 168, 1, 5}, "path=/api")
	defer feedResponseUntilStopped(q, packet)()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	services, err := q.DiscoverServices(ctx, "_http._tcp.local")
	if err != nil {
		t.Fatal(err)
	}
	s := findServiceByInstance(services, "INST")
	if s == nil {
		t.Fatalf("instance \"INST\" not found in %+v", services)
	}
	if s.Port != 8080 || s.AddrIPv4 == nil || s.TXT["path"] != "/api" {
		t.Errorf("service not resolved from additionals: %+v", *s)
	}
}

// An empty or root name is never a valid SRV/TXT question.
func TestQuery_EmptyInstanceName_Rejected(t *testing.T) {
	q := &Querier{transport: transport.NewMockTransport(), responseChan: make(chan []byte, 1), defaultTimeout: 10 * time.Millisecond}
	for _, name := range []string{"", "."} {
		if _, err := q.Query(context.Background(), name, RecordTypeSRV); err == nil {
			t.Errorf("Query(%q, SRV) accepted, want validation error", name)
		}
	}
}
