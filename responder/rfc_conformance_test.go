package responder

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	iresp "github.com/joshuafuller/beacon/internal/responder"
	"github.com/joshuafuller/beacon/internal/security"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Wire-level RFC 6762 conformance tests for query handling (issues #40, #41,
// #44, #45, #46). Each test sends real query bytes into handleQuery and checks
// the bytes the responder puts on the wire.

func newConformanceResponder(t *testing.T, instances ...string) (*Responder, *transport.MockTransport) {
	t.Helper()
	mt := transport.NewMockTransport()
	r := &Responder{
		ctx:              context.Background(),
		transport:        mt,
		registry:         iresp.NewRegistry(),
		hostname:         "myhost.local",
		responseBuilder:  iresp.NewResponseBuilder(),
		rateLimiter:      security.NewRateLimiter(100, 60*time.Second, 10000),
		queryHandlerDone: make(chan struct{}),
	}
	for _, name := range instances {
		svc := &Service{InstanceName: name, ServiceType: "_http._tcp.local", Port: 80}
		if err := r.RegisterServiceWithoutProbing(svc); err != nil {
			t.Fatalf("register %q: %v", name, err)
		}
	}
	return r, mt
}

type testQuestion struct {
	name  string
	qtype uint16
}

// buildTestQuery builds a query packet with the given ID and questions.
func buildTestQuery(t *testing.T, id uint16, questions ...testQuestion) []byte {
	t.Helper()
	pkt := make([]byte, 12)
	binary.BigEndian.PutUint16(pkt[0:2], id)
	binary.BigEndian.PutUint16(pkt[4:6], uint16(len(questions)))
	for _, q := range questions {
		enc, err := message.EncodeDomainName(q.name)
		if err != nil {
			t.Fatalf("encode %q: %v", q.name, err)
		}
		pkt = append(pkt, enc...)
		pkt = binary.BigEndian.AppendUint16(pkt, q.qtype)
		pkt = binary.BigEndian.AppendUint16(pkt, 1)
	}
	return pkt
}

var mdnsSource = &net.UDPAddr{IP: net.IPv4(192, 168, 1, 50), Port: 5353}

func sentMessages(t *testing.T, mt *transport.MockTransport) ([]*message.DNSMessage, []transport.SendCall) {
	t.Helper()
	calls := mt.SendCalls()
	msgs := make([]*message.DNSMessage, 0, len(calls))
	for _, c := range calls {
		m, err := message.ParseMessage(c.Packet)
		if err != nil {
			t.Fatalf("responder sent unparseable packet: %v", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, calls
}

func answersOfType(msgs []*message.DNSMessage, rrtype uint16) []message.Answer {
	var out []message.Answer
	for _, m := range msgs {
		for _, a := range m.Answers {
			if a.TYPE == rrtype {
				out = append(out, a)
			}
		}
	}
	return out
}

// RFC 6763 §4.1: a browse lists every instance of the service type (#40).
func TestHandleQuery_PTRBrowse_ReturnsEveryInstance(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha", "Bravo")
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
	flushResponses(r)

	msgs, _ := sentMessages(t, mt)
	if len(msgs) != 1 {
		t.Fatalf("sent %d packets, want 1", len(msgs))
	}
	got := map[string]bool{}
	for _, a := range answersOfType(msgs, 12) {
		target, err := message.ParseRDATA(12, a.RDATA)
		if err != nil {
			t.Fatalf("PTR rdata: %v", err)
		}
		got[target.(string)] = true
	}
	for _, want := range []string{"Alpha._http._tcp.local", "Bravo._http._tcp.local"} {
		if !got[want] {
			t.Errorf("PTR answers %v missing %q", got, want)
		}
	}
}

// RFC 6762 §6: each question in a multi-question query is answered (#41).
func TestHandleQuery_SecondQuestionAnswered(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"other.local", 1}, testQuestion{"_http._tcp.local", 12})
	_ = r.handleQuery(mt, q, mdnsSource, 0)
	flushResponses(r)

	msgs, _ := sentMessages(t, mt)
	if len(answersOfType(msgs, 12)) != 1 {
		t.Errorf("PTR answers = %d, want 1 for the second question", len(answersOfType(msgs, 12)))
	}
}

// RFC 6762 §6: do not put empty responses on the wire (#41).
func TestHandleQuery_NoAnswers_SendsNothing(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"other.local", 1}), mdnsSource, 0)
	flushResponses(r)
	if n := len(mt.SendCalls()); n != 0 {
		t.Errorf("sent %d packets for a query we cannot answer, want 0", n)
	}
}

// RFC 6762 §18.1: multicast responses MUST have ID zero (#44).
func TestHandleQuery_MulticastResponse_HasZeroID(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	_ = r.handleQuery(mt, buildTestQuery(t, 0x1234, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
	flushResponses(r)

	_, calls := sentMessages(t, mt)
	if len(calls) != 1 {
		t.Fatalf("sent %d packets, want 1", len(calls))
	}
	if calls[0].Dest != nil {
		t.Fatalf("dest = %v, want multicast (nil)", calls[0].Dest)
	}
	if id := binary.BigEndian.Uint16(calls[0].Packet[0:2]); id != 0 {
		t.Errorf("multicast response ID = 0x%04x, want 0 (RFC 6762 §18.1)", id)
	}
}

// RFC 1035 §2.3.3 / RFC 6762 §16: names match without regard to ASCII case (#45).
func TestHandleQuery_NamesMatchCaseInsensitively(t *testing.T) {
	tests := []testQuestion{
		{"_HTTP._TCP.LOCAL", 12},
		{"ALPHA._HTTP._TCP.LOCAL", 33},
		{"ALPHA._HTTP._TCP.LOCAL", 16},
		{"MYHOST.LOCAL", 1},
	}
	for _, tq := range tests {
		r, mt := newConformanceResponder(t, "Alpha")
		_ = r.handleQuery(mt, buildTestQuery(t, 0, tq), mdnsSource, 0)
		flushResponses(r)
		msgs, _ := sentMessages(t, mt)
		if len(answersOfType(msgs, tq.qtype)) == 0 {
			t.Errorf("query %s type %d: no answer", tq.name, tq.qtype)
		}
	}
}

// RFC 6762 §6.7: a query from a port other than 5353 gets a conventional
// unicast DNS response to that address and port (#46).
func TestHandleQuery_LegacyUnicast(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	legacy := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 50), Port: 40000}
	_ = r.handleQuery(mt, buildTestQuery(t, 0x77, testQuestion{"_http._tcp.local", 12}), legacy, 0)

	msgs, calls := sentMessages(t, mt)
	if len(calls) != 1 {
		t.Fatalf("sent %d packets, want 1", len(calls))
	}
	if calls[0].Dest == nil || calls[0].Dest.String() != legacy.String() {
		t.Fatalf("dest = %v, want %v", calls[0].Dest, legacy)
	}
	m := msgs[0]
	if m.Header.ID != 0x77 {
		t.Errorf("ID = 0x%04x, want 0x0077", m.Header.ID)
	}
	if len(m.Questions) != 1 || m.Questions[0].QNAME != "_http._tcp.local" || m.Questions[0].QTYPE != 12 {
		t.Errorf("questions = %+v, want the original PTR question", m.Questions)
	}
	all := append(append([]message.Answer{}, m.Answers...), m.Additionals...)
	if len(m.Answers) == 0 {
		t.Fatal("legacy response has no answers")
	}
	for _, a := range all {
		if a.CLASS&0x8000 != 0 {
			t.Errorf("%s type %d has cache-flush bit set", a.NAME, a.TYPE)
		}
		if a.TTL > 10 {
			t.Errorf("%s type %d TTL = %d, want <= 10", a.NAME, a.TYPE, a.TTL)
		}
	}
}

// A legacy SRV query for an instance name with a space must echo the question
// intact (the question encoder must accept RFC 6763 §4.3 instance labels).
func TestHandleQuery_LegacyUnicast_InstanceWithSpace(t *testing.T) {
	r, mt := newConformanceResponder(t, "My Printer")
	legacy := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 50), Port: 40000}
	_ = r.handleQuery(mt, buildTestQuery(t, 9, testQuestion{"My Printer._http._tcp.local", 33}), legacy, 0)

	msgs, _ := sentMessages(t, mt)
	if len(msgs) != 1 || len(msgs[0].Answers) != 1 || len(msgs[0].Questions) != 1 {
		t.Fatalf("got %d packets; want 1 packet with 1 question and 1 SRV answer", len(msgs))
	}
	if q := msgs[0].Questions[0].QNAME; q != "My Printer._http._tcp.local" {
		t.Errorf("echoed question = %q", q)
	}
}

// RFC 6762 §7.1: when the querier already knows every answer, send nothing (#41).
func TestHandleQuery_AllKnownAnswers_SendsNothing(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	target, _ := message.EncodeDomainName("Alpha._http._tcp.local")
	owner, _ := message.EncodeName("_http._tcp.local")
	q[7] = 1 // ANCOUNT = 1: the known answer
	q = append(q, owner...)
	q = binary.BigEndian.AppendUint16(q, 12)
	q = binary.BigEndian.AppendUint16(q, 1)
	q = binary.BigEndian.AppendUint32(q, 120)
	q = binary.BigEndian.AppendUint16(q, uint16(len(target)))
	q = append(q, target...)

	_ = r.handleQuery(mt, q, mdnsSource, 0)
	flushResponses(r)
	if n := len(mt.SendCalls()); n != 0 {
		t.Errorf("sent %d packets although the only answer is known, want 0", n)
	}
}
