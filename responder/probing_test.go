package responder

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Probing tests mirror BCT II.1-II.6 (RFC 6762 §8.1, §8.2, §9).

var otherHost = &net.UDPAddr{IP: net.IPv4(192, 168, 1, 77), Port: 5353}

// buildResponse builds a response packet holding one record.
func buildResponse(t *testing.T, name string, rrtype uint16, rdata []byte) []byte {
	t.Helper()
	pkt := make([]byte, 12)
	binary.BigEndian.PutUint16(pkt[2:4], 0x8400)
	binary.BigEndian.PutUint16(pkt[6:8], 1)
	enc, err := message.EncodeDomainName(name)
	if err != nil {
		t.Fatal(err)
	}
	pkt = append(pkt, enc...)
	pkt = binary.BigEndian.AppendUint16(pkt, rrtype)
	pkt = binary.BigEndian.AppendUint16(pkt, 0x8001)
	pkt = binary.BigEndian.AppendUint32(pkt, 120)
	pkt = binary.BigEndian.AppendUint16(pkt, uint16(len(rdata)))
	return append(pkt, rdata...)
}

// registerAsync runs Register in the background and calls inject once the
// first probe is on the wire.
func registerAsync(t *testing.T, r *Responder, mt *transport.MockTransport, svc *Service, inject func()) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- r.Register(svc) }()
	if inject != nil {
		deadline := time.Now().Add(2 * time.Second)
		for len(mt.SendCalls()) == 0 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		inject()
	}
	return done
}

func waitDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Register did not finish")
	}
}

// BCT II.1: probes ask for the host name and the instance (ANY, QU) and carry
// the proposed records in the Authority section.
func TestRegister_ProbeHasHostInstanceAndAuthority(t *testing.T) {
	r, mt := newConformanceResponder(t)
	waitDone(t, registerAsync(t, r, mt, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}, nil))

	probe, err := message.ParseMessage(mt.SendCalls()[0].Packet)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, q := range probe.Questions {
		names[q.QNAME] = q.QTYPE == 255 && q.QCLASS&0x8000 != 0
	}
	if !names["Alpha._http._tcp.local"] || !names["myhost.local"] {
		t.Errorf("probe questions %v, want ANY/QU for instance and host", names)
	}
	types := map[uint16]bool{}
	for _, a := range probe.Authorities {
		types[a.TYPE] = true
	}
	if !types[33] || !types[16] || !types[1] {
		t.Errorf("authority types %v, want SRV, TXT and A", types)
	}

	// The host is established: a second service probes only its instance.
	mt2 := mt
	before := len(mt2.SendCalls())
	waitDone(t, registerAsync(t, r, mt, &Service{InstanceName: "Bravo", ServiceType: "_http._tcp.local", Port: 81}, nil))
	probe2, _ := message.ParseMessage(mt.SendCalls()[before].Packet)
	if len(probe2.Questions) != 1 || probe2.Questions[0].QNAME != "Bravo._http._tcp.local" {
		t.Errorf("second probe questions %+v, want only the new instance", probe2.Questions)
	}
}

// BCT II.2: a response for the host name while probing makes the device pick
// a new host name.
func TestRegister_HostConflictRenamesHost(t *testing.T) {
	r, mt := newConformanceResponder(t)
	done := registerAsync(t, r, mt, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}, func() {
		_ = r.handleQuery(mt, buildResponse(t, "MYHOST.local", 1, []byte{192, 168, 1, 77}), otherHost, 0)
	})
	waitDone(t, done)
	if got := r.hostName(); got != "myhost-2.local" {
		t.Errorf("host name = %q, want myhost-2.local", got)
	}
	if _, ok := r.registry.Get("Alpha"); !ok {
		t.Error("service kept its name but is not registered")
	}
}

// RFC 6762 §8.1: a response for the instance name while probing renames the
// service.
func TestRegister_InstanceConflictRenamesService(t *testing.T) {
	r, mt := newConformanceResponder(t)
	svc := &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}
	done := registerAsync(t, r, mt, svc, func() {
		_ = r.handleQuery(mt, buildResponse(t, "Alpha._http._tcp.local", 16, []byte{3, 'x', '=', '1'}), otherHost, 0)
	})
	waitDone(t, done)
	if svc.InstanceName != "Alpha-2" || r.hostName() != "myhost.local" {
		t.Errorf("instance=%q host=%q, want Alpha-2 / myhost.local", svc.InstanceName, r.hostName())
	}
}

// BCT II.2/II.4, RFC 6762 §8.2: losing a simultaneous-probe tiebreak defers
// one second and keeps the name; winning it changes nothing.
func TestRegister_LostTiebreakDefersAndKeepsName(t *testing.T) {
	r, mt := newConformanceResponder(t)
	svc := &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}
	start := time.Now()
	done := registerAsync(t, r, mt, svc, func() {
		// Their probe: same instance, SRV with a larger port, so their data is
		// lexicographically later.
		q := buildTestQuery(t, 0, testQuestion{"Alpha._http._tcp.local", 255})
		q[9] = 1 // NSCOUNT = 1
		enc, _ := message.EncodeDomainName("Alpha._http._tcp.local")
		target, _ := message.EncodeName("other.local")
		rdata := append([]byte{0, 0, 0, 0, 0xff, 0xff}, target...)
		q = append(q, enc...)
		q = binary.BigEndian.AppendUint16(q, uint16(protocol.RecordTypeSRV))
		q = binary.BigEndian.AppendUint16(q, 1)
		q = binary.BigEndian.AppendUint32(q, 120)
		q = binary.BigEndian.AppendUint16(q, uint16(len(rdata)))
		q = append(q, rdata...)
		_ = r.handleQuery(mt, q, otherHost, 0)
	})
	waitDone(t, done)
	if svc.InstanceName != "Alpha" {
		t.Errorf("instance renamed to %q after a lost tiebreak, want Alpha kept", svc.InstanceName)
	}
	if d := time.Since(start); d < time.Second {
		t.Errorf("registered after %v, want >= 1 s deferral", d)
	}
	probes := 0
	for _, c := range mt.SendCalls() {
		if m, err := message.ParseMessage(c.Packet); err == nil && !m.Header.IsResponse() {
			probes++
		}
	}
	if probes < 4 {
		t.Errorf("%d probes sent, want a second probe round after the deferral (>= 4)", probes)
	}
}

// BCT II.6, RFC 6762 §9: a conflicting response after announcement sends the
// name back to probing (same name first).
func TestEstablishedConflict_Reprobes(t *testing.T) {
	r, mt := newConformanceResponder(t)
	waitDone(t, registerAsync(t, r, mt, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}, nil))
	before := len(mt.SendCalls())

	_ = r.handleQuery(mt, buildResponse(t, "Alpha._http._tcp.local", 16, []byte{3, 'x', '=', '1'}), otherHost, 0)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := r.registry.Get("Alpha"); ok && len(mt.SendCalls()) > before+3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	probe, _ := message.ParseMessage(mt.SendCalls()[before].Packet)
	if probe == nil || len(probe.Questions) == 0 || probe.Questions[0].QNAME != "Alpha._http._tcp.local" {
		t.Fatalf("no re-probe for Alpha after the conflict")
	}
	if _, ok := r.registry.Get("Alpha"); !ok {
		t.Error("Alpha not registered again after an uncontested re-probe")
	}

	// Our own record looped back is not a conflict.
	n := len(mt.SendCalls())
	ours := r.lastAnnouncedRecords
	for _, rr := range ours {
		if rr.Type == protocol.RecordTypeTXT {
			_ = r.handleQuery(mt, buildResponse(t, rr.Name, 16, rr.Data), otherHost, 0)
		}
	}
	time.Sleep(300 * time.Millisecond)
	if len(mt.SendCalls()) != n {
		t.Error("our own TXT record looped back started a re-probe")
	}
}

// RFC 6762 §8.1: 15 conflicts within 10 s → wait 5 s before each probe.
func TestProbeRateLimit(t *testing.T) {
	r, _ := newConformanceResponder(t)
	for i := 0; i < conflictBurst-1; i++ {
		r.noteConflict()
	}
	if r.overConflictLimit() {
		t.Fatal("rate limit active after 14 conflicts")
	}
	r.noteConflict()
	if !r.overConflictLimit() {
		t.Fatal("rate limit not active after 15 conflicts in 10 s")
	}
}

func TestRenameHost(t *testing.T) {
	r, _ := newConformanceResponder(t)
	if got := r.renameHost("myhost.local"); got != "myhost-2.local" || r.hostName() != got {
		t.Errorf("first rename = %q (current %q)", got, r.hostName())
	}
	if got := r.renameHost("myhost-2.local"); got != "myhost-3.local" {
		t.Errorf("second rename = %q", got)
	}
}

// BCT II.5 (SRV probing): BCT's simultaneous probe carries only an SRV
// record. Our probe proposes SRV and TXT. The TXT record depends on the SRV
// (as in Apple's mDNSResponder), so the tiebreak is decided by the SRV: our
// port 8080 beats their 8079, and we must not defer.
func TestTiebreak_TXTDependsOnSRV_WinOnSRV(t *testing.T) {
	r, mt := newConformanceResponder(t)
	svc := &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 8080}
	done := registerAsync(t, r, mt, svc, func() {
		q := buildTestQuery(t, 0, testQuestion{"Alpha._http._tcp.local", 255})
		q[9] = 1 // NSCOUNT = 1
		enc, _ := message.EncodeDomainName("Alpha._http._tcp.local")
		target, _ := message.EncodeName("myhost.local")
		rdata := append([]byte{0, 0, 0, 0, 0x1f, 0x8f}, target...) // port 8079
		q = append(q, enc...)
		q = binary.BigEndian.AppendUint16(q, uint16(protocol.RecordTypeSRV))
		q = binary.BigEndian.AppendUint16(q, 1)
		q = binary.BigEndian.AppendUint32(q, 120)
		q = binary.BigEndian.AppendUint16(q, uint16(len(rdata)))
		q = append(q, rdata...)
		_ = r.handleQuery(mt, q, otherHost, 0)
	})
	waitDone(t, done)
	probes := 0
	for _, c := range mt.SendCalls() {
		if m, err := message.ParseMessage(c.Packet); err == nil && !m.Header.IsResponse() {
			probes++
		}
	}
	if probes != 3 || svc.InstanceName != "Alpha" {
		t.Errorf("probes=%d name=%q, want 3 probes and Alpha kept (we win on SRV)", probes, svc.InstanceName)
	}
}
