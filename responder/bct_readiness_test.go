package responder

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Tests for behaviour that Apple's Bonjour Conformance Test (BCT 1.5.4)
// checks. Each names the BCT subtest it mirrors. Helpers live in
// rfc_conformance_test.go.

// BCT II.7/II.8: queries with type ANY (255) must be answered with every
// record of the name (RFC 6762 §6, RFC 1035 §3.2.3).
func TestHandleQuery_ANY_AnswersAllRecordsOfName(t *testing.T) {
	tests := []struct {
		name      string
		wantTypes []uint16
	}{
		{"Alpha._http._tcp.local", []uint16{33, 16}}, // SRV + TXT
		{"myhost.local", []uint16{1}},                // A
		{"_http._tcp.local", []uint16{12}},           // PTR
	}
	for _, tc := range tests {
		r, mt := newConformanceResponder(t, "Alpha")
		_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{tc.name, 255}), mdnsSource, 0)
		flushResponses(r)
		msgs, _ := sentMessages(t, mt)
		for _, want := range tc.wantTypes {
			if len(answersOfType(msgs, want)) == 0 {
				t.Errorf("ANY %s: no type %d answer", tc.name, want)
			}
		}
		for _, m := range msgs {
			for _, a := range m.Answers {
				if !namesEqual(a.NAME, tc.name) {
					t.Errorf("ANY %s: answer for other name %s", tc.name, a.NAME)
				}
			}
		}
	}
}

// BCT II.9: known answers whose RDATA uses name compression still suppress
// our answer (RFC 6762 §7.1; RDATA is compared after decompression).
func TestHandleQuery_CompressedKnownAnswer_Suppresses(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	q[7] = 1 // ANCOUNT = 1
	// Owner: pointer to the question name at offset 12.
	q = append(q, 0xC0, 12)
	q = binary.BigEndian.AppendUint16(q, 12)
	q = binary.BigEndian.AppendUint16(q, 1)
	q = binary.BigEndian.AppendUint32(q, 120)
	// RDATA: "Alpha" label + pointer to "_http._tcp.local" at offset 12.
	rdata := append([]byte{5}, "Alpha"...)
	rdata = append(rdata, 0xC0, 12)
	q = binary.BigEndian.AppendUint16(q, uint16(len(rdata)))
	q = append(q, rdata...)

	_ = r.handleQuery(mt, q, mdnsSource, 0)
	flushResponses(r)
	if n := len(mt.SendCalls()); n != 0 {
		t.Errorf("sent %d packets although the compressed known answer covers it, want 0", n)
	}
}

// waitForSends polls until mt has at least n sends or timeout passes.
func waitForSends(mt *transport.MockTransport, n int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if got := len(mt.SendCalls()); got >= n {
			return got
		}
		time.Sleep(2 * time.Millisecond)
	}
	return len(mt.SendCalls())
}

// BCT II.8: shared (PTR) answers to multicast queries are delayed by a random
// 20-120 ms (RFC 6762 §6).
func TestHandleQuery_SharedAnswer_RandomDelay(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	start := time.Now()
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
	if n := len(mt.SendCalls()); n != 0 {
		t.Fatalf("PTR answer sent at once (%d packets), want 20-120 ms delay", n)
	}
	if waitForSends(mt, 1, 300*time.Millisecond) != 1 {
		t.Fatal("no delayed PTR answer within 300 ms")
	}
	if d := time.Since(start); d < 20*time.Millisecond || d > 140*time.Millisecond {
		t.Errorf("PTR answer after %v, want 20-120 ms", d)
	}
}

// BCT II.7: unique answers (SRV/TXT/A) are not delayed.
func TestHandleQuery_UniqueAnswer_Immediate(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"Alpha._http._tcp.local", 33}), mdnsSource, 0)
	if n := len(mt.SendCalls()); n != 1 {
		t.Errorf("SRV answer: %d packets sent at once, want 1", n)
	}
}

// BCT II.15: two back-to-back shared queries get one aggregated response.
func TestHandleQuery_BackToBackQueries_Aggregated(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	if err := r.RegisterServiceWithoutProbing(&Service{InstanceName: "Bravo", ServiceType: "_ipp._tcp.local", Port: 631}); err != nil {
		t.Fatal(err)
	}
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12}), mdnsSource, 0)
	_ = r.handleQuery(mt, buildTestQuery(t, 0, testQuestion{"_ipp._tcp.local", 12}), mdnsSource, 0)
	waitForSends(mt, 1, 300*time.Millisecond)
	time.Sleep(150 * time.Millisecond) // a second packet would have arrived by now

	msgs, _ := sentMessages(t, mt)
	if len(msgs) != 1 {
		t.Fatalf("sent %d packets, want 1 aggregated response", len(msgs))
	}
	if n := len(answersOfType(msgs, 12)); n != 2 {
		t.Errorf("aggregated response has %d PTR answers, want 2", n)
	}
}

// BCT II.10: a query with TC set waits 400-500 ms for more known answers
// (RFC 6762 §7.2); a later known-answer packet suppresses the answer.
func TestHandleQuery_TruncatedQuery_WaitsForKnownAnswers(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	q[2] |= 0x02 // TC bit
	_ = r.handleQuery(mt, q, mdnsSource, 0)

	// Follow-up packet: no questions, one known answer (our PTR, full TTL).
	ka := make([]byte, 12)
	ka[7] = 1
	owner, _ := message.EncodeName("_http._tcp.local")
	target, _ := message.EncodeDomainName("Alpha._http._tcp.local")
	ka = append(ka, owner...)
	ka = binary.BigEndian.AppendUint16(ka, 12)
	ka = binary.BigEndian.AppendUint16(ka, 1)
	ka = binary.BigEndian.AppendUint32(ka, 120)
	ka = binary.BigEndian.AppendUint16(ka, uint16(len(target)))
	ka = append(ka, target...)
	time.Sleep(50 * time.Millisecond)
	_ = r.handleQuery(mt, ka, mdnsSource, 0)

	if n := waitForSends(mt, 1, 700*time.Millisecond); n != 0 {
		t.Errorf("sent %d packets although a follow-up known answer covers it, want 0", n)
	}
}

func TestHandleQuery_TruncatedQuery_Delay400to500ms(t *testing.T) {
	r, mt := newConformanceResponder(t, "Alpha")
	q := buildTestQuery(t, 0, testQuestion{"_http._tcp.local", 12})
	q[2] |= 0x02 // TC bit
	start := time.Now()
	_ = r.handleQuery(mt, q, mdnsSource, 0)
	if waitForSends(mt, 1, 700*time.Millisecond) != 1 {
		t.Fatal("no answer within 700 ms")
	}
	if d := time.Since(start); d < 400*time.Millisecond || d > 520*time.Millisecond {
		t.Errorf("answer after %v, want 400-500 ms", d)
	}
}

// flushResponses sends every delayed response now (test helper).
func flushResponses(r *Responder) {
	r.scheduler.mu.Lock()
	pending := make(map[schedulerKey]*pendingResponse, len(r.scheduler.pending))
	for k, p := range r.scheduler.pending {
		pending[k] = p
	}
	r.scheduler.mu.Unlock()
	for k, p := range pending {
		p.timer.Stop()
		r.scheduler.fire(r, k, p)
	}
}
