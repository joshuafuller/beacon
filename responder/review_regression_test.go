package responder

// Regression tests from the adversarial review of PR #51. Each failed on 1502dcd.

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/security"
)

var legacySrc = &net.UDPAddr{IP: net.IPv4(192, 168, 1, 50), Port: 40000}

// P1: RFC 6762 §6.7 for the service enumeration meta-query.
func TestReview_LegacyServiceEnumeration(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	_ = r.handleQuery(mt, buildTestQuery(t, 0x77, testQuestion{"_services._dns-sd._udp.local", 12}), legacySrc, 0)
	calls := mt.SendCalls()
	if len(calls) != 1 {
		t.Fatalf("sent %d", len(calls))
	}
	m, _ := message.ParseMessage(calls[0].Packet)
	t.Logf("dest=%v id=0x%04x qd=%d", calls[0].Dest, m.Header.ID, m.Header.QDCount)
	if calls[0].Dest == nil || m.Header.ID != 0x77 || m.Header.QDCount != 1 {
		t.Errorf("legacy enum query not answered per §6.7: dest=%v id=0x%04x qd=%d", calls[0].Dest, m.Header.ID, m.Header.QDCount)
	}
}

// P2: a browse with many instances is split into MTU-sized multicast packets
// with TC clear (RFC 6762 §17, §18.5), and every instance is answered.
func TestReview_MergedBrowseSizeAndTC(t *testing.T) {
	for _, n := range []int{1, 10, 40, 200} {
		var names []string
		for i := 0; i < n; i++ {
			names = append(names, fmt.Sprintf("Instance %03d with a fairly long name xxxxxxxxxx", i))
		}
		r, mt := newConformanceResponder(t, names...)
		_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
		flushResponses(r)
		calls := mt.SendCalls()
		ptrs := 0
		for _, c := range calls {
			flags := binary.BigEndian.Uint16(c.Packet[2:4])
			if c.Dest == nil && flags&0x0200 != 0 {
				t.Errorf("n=%d: multicast response has TC=1 (RFC 6762 §18.5 MUST be zero)", n)
			}
			if len(c.Packet) > 1440 {
				t.Errorf("n=%d: packet payload %d bytes > 1440 (MTU, RFC 6762 §17)", n, len(c.Packet))
			}
			m, err := message.ParseMessage(c.Packet)
			if err != nil {
				t.Fatal(err)
			}
			ptrs += int(m.Header.ANCount)
		}
		if ptrs != n {
			t.Errorf("n=%d: %d PTR answers in %d packets, want %d", n, ptrs, len(calls), n)
		}
	}
}

func appendKnownPTR(q []byte, ownerOff int, rdata []byte) []byte {
	binary.BigEndian.PutUint16(q[6:8], binary.BigEndian.Uint16(q[6:8])+1)
	q = append(q, 0xC0|byte(ownerOff>>8), byte(ownerOff)) // owner = pointer to question name
	q = binary.BigEndian.AppendUint16(q, 12)
	q = binary.BigEndian.AppendUint16(q, 1)
	q = binary.BigEndian.AppendUint32(q, 4500)
	q = binary.BigEndian.AppendUint16(q, uint16(len(rdata)))
	return append(q, rdata...)
}

// P3: known-answer PTR for Alpha suppresses the PTR, but Alpha's SRV/TXT still go out.
func TestReview_KnownAnswerLeavesSuppressedInstanceAdditionals(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha", "Bravo")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	target, _ := message.EncodeDomainName("Alpha._http._tcp.local")
	q = appendKnownPTR(q, 12, target)
	_ = r.handleQuery(mt, q, mdnsSource, 0)
	flushResponses(r)
	msgs, _ := sentMessages(t, mt)
	for _, m := range msgs {
		for _, a := range m.Answers {
			t.Logf("AN %s type %d", a.NAME, a.TYPE)
		}
		for _, a := range m.Additionals {
			t.Logf("AR %s type %d", a.NAME, a.TYPE)
			if strings.HasPrefix(a.NAME, "Alpha.") {
				t.Errorf("additional %s type %d sent for instance whose PTR answer was suppressed", a.NAME, a.TYPE)
			}
		}
	}
}

// P4: known answer with a compressed PTR target (what Bonjour/Avahi send) does not suppress.
func TestReview_KnownAnswerCompressedTargetNotMatched(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	rdata := append([]byte{5, 'A', 'l', 'p', 'h', 'a'}, 0xC0, 12) // Alpha + pointer to _http._tcp.local
	q = appendKnownPTR(q, 12, rdata)
	_ = r.handleQuery(mt, q, mdnsSource, 0)
	flushResponses(r)
	if n := len(mt.SendCalls()); n != 0 {
		t.Errorf("sent %d packets although the PTR is a known answer (compressed RDATA not compared as a name)", n)
	}
}

// P5: legacy query with one question we cannot re-encode: a bare 12-byte header goes out.
func TestReview_LegacyEchoUnencodableQuestion(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	pkt := make([]byte, 12)
	binary.BigEndian.PutUint16(pkt[0:2], 0x99)
	binary.BigEndian.PutUint16(pkt[4:6], 2)
	pkt = append(pkt, 8)
	pkt = append(pkt, "bad name"...)
	pkt = append(pkt, 5)
	pkt = append(pkt, "local"...)
	pkt = append(pkt, 0, 0, 1, 0, 1)
	enc, _ := message.EncodeName("_http._tcp.local")
	pkt = append(pkt, enc...)
	pkt = append(pkt, 0, 12, 0, 1)
	_ = r.handleQuery(mt, pkt, legacySrc, 0)
	for _, c := range mt.SendCalls() {
		m, _ := message.ParseMessage(c.Packet)
		t.Logf("dest=%v len=%d id=0x%04x an=%d qd=%d", c.Dest, len(c.Packet), m.Header.ID, m.Header.ANCount, m.Header.QDCount)
		if m.Header.ANCount == 0 {
			t.Errorf("legacy response is an empty header (ID 0x%04x, ANCOUNT 0)", m.Header.ID)
		}
	}
}

// P6: PTR + SRV questions in one query: SRV is in Answer and again in Additional.
func TestReview_MergeDuplicatesAnswerInAdditional(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}, testQuestion{"Alpha._http._tcp.local", 33}), mdnsSource, 0)
	flushResponses(r)
	msgs, _ := sentMessages(t, mt)
	inAN := false
	for _, a := range msgs[0].Answers {
		if a.TYPE == 33 {
			inAN = true
		}
	}
	for _, a := range msgs[0].Additionals {
		if a.TYPE == 33 && inAN {
			t.Errorf("SRV %s is in both Answer and Additional sections", a.NAME)
		}
	}
}

// Instances whose name starts with "_" or contains "._" (RFC 6763 §4.3
// allows any UTF-8) must still be answered.
func TestReview_UnderscoreInstance(t *testing.T) {
	for _, inst := range []string{"_My Printer", "Printer._v2"} {
		r, mt := newConformanceResponder(t, inst)
		_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
		flushResponses(r)
		msgs, _ := sentMessages(t, mt)
		if len(msgs) != 1 || msgs[0].Header.ANCount == 0 {
			t.Fatalf("%q: no PTR answer", inst)
		}
		target, err := message.ParseRDATA(12, msgs[0].Answers[0].RDATA)
		if err != nil {
			t.Fatal(err)
		}
		wire, _ := message.EncodeServiceInstanceName(inst, "_http._tcp.local")
		if !bytesEqual(msgs[0].Answers[0].RDATA, wire) {
			t.Errorf("%q: PTR target %q encoded with wrong label split", inst, target)
		}
	}
}

func bytesEqual(a, b []byte) bool { return string(a) == string(b) }

func TestReview_BadHostFlag(t *testing.T) {
	r, mt := newConformanceResponder(t, "Beacon BCT")
	if err := WithHostname("bad name.local")(r); err != nil {
		t.Logf("WithHostname rejected: %v", err)
		return
	}
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
	flushResponses(r)
	for _, c := range mt.SendCalls() {
		m, _ := message.ParseMessage(c.Packet)
		t.Logf("len=%d an=%d", len(c.Packet), m.Header.ANCount)
		if m.Header.ANCount == 0 {
			t.Error("-host 'bad name' accepted; responder sends empty 12-byte headers")
		}
	}
}

// A packet repeating the enumeration question gets one reply, and the
// per-source rate limit applies to enumeration replies too.
func TestReview_EnumerationAnsweredOncePerPacketAndRateLimited(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	qs := make([]testQuestion, 5)
	for i := range qs {
		qs[i] = testQuestion{"_services._dns-sd._udp.local", 12}
	}
	_ = r.handleQuery(mt, buildTestQuery(t, 1, qs...), legacySrc, 0)
	if n := len(mt.SendCalls()); n != 1 {
		t.Errorf("5 enumeration questions produced %d replies, want 1", n)
	}

	r2, mt2 := newConformanceResponder(t, "Alpha")
	r2.rateLimiter = security.NewRateLimiter(1, time.Minute, 10)
	for i := 0; i < 5; i++ {
		_ = r2.handleQuery(mt2, buildTestQuery(t, 1, testQuestion{"_services._dns-sd._udp.local", 12}), legacySrc, 0)
	}
	if n := len(mt2.SendCalls()); n > 1 {
		t.Errorf("rate limit 1/s allowed %d enumeration replies, want <= 1", n)
	}
}
