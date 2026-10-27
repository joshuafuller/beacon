package state

import (
	"context"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
)

func testProbeSet() ProbeSet {
	srv := &message.ResourceRecord{Name: "Alpha._http._tcp.local", Type: protocol.RecordTypeSRV, Class: protocol.ClassIN, TTL: 120,
		Data: []byte{0, 0, 0, 0, 0x1f, 0x90, 6, 'm', 'y', 'h', 'o', 's', 't', 5, 'l', 'o', 'c', 'a', 'l', 0}, CacheFlush: true}
	a := &message.ResourceRecord{Name: "myhost.local", Type: protocol.RecordTypeA, Class: protocol.ClassIN, TTL: 120,
		Data: []byte{169, 254, 20, 20}, CacheFlush: true}
	return ProbeSet{Names: []string{"Alpha._http._tcp.local", "myhost.local"}, Authority: []*message.ResourceRecord{srv, a}}
}

// RFC 6762 §8.1/§8.2: QU questions of type ANY, proposed records in the
// Authority section, cache-flush bit clear (BCT II.1).
func TestBuildProbeSetMessage_Format(t *testing.T) {
	pkt, err := buildProbeSetMessage(testProbeSet())
	if err != nil {
		t.Fatal(err)
	}
	msg, err := message.ParseMessage(pkt)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Header.Flags != 0 || len(msg.Questions) != 2 || len(msg.Authorities) != 2 {
		t.Fatalf("flags=%#x qd=%d ns=%d, want 0/2/2", msg.Header.Flags, len(msg.Questions), len(msg.Authorities))
	}
	for _, q := range msg.Questions {
		if q.QTYPE != uint16(protocol.RecordTypeANY) || q.QCLASS != uint16(protocol.ClassIN)|0x8000 {
			t.Errorf("question %s type=%d class=%#x, want ANY, IN|QU", q.QNAME, q.QTYPE, q.QCLASS)
		}
	}
	for _, rr := range msg.Authorities {
		if rr.CLASS&0x8000 != 0 {
			t.Errorf("proposed %s type %d has cache-flush set", rr.NAME, rr.TYPE)
		}
	}
}

func runProbeWithEvent(t *testing.T, ev *ProbeEvent, after time.Duration) (ProbeResult, int) {
	t.Helper()
	mt := transport.NewMockTransport()
	events := make(chan ProbeEvent, 4)
	p := NewProber()
	p.SetTransport(mt)
	p.SetProbeSet(testProbeSet(), events)
	if ev != nil {
		go func() {
			deadline := time.Now().Add(2 * time.Second)
			for len(mt.SendCalls()) == 0 && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			time.Sleep(after)
			events <- *ev
		}()
	}
	res := p.Probe(context.Background(), "Alpha._http._tcp.local")
	return res, len(mt.SendCalls())
}

func TestProbeSet_NoConflict_ThreeProbes(t *testing.T) {
	start := time.Now()
	res, sent := runProbeWithEvent(t, nil, 0)
	if res.Conflict || res.Error != nil || sent != 3 {
		t.Fatalf("result=%+v sent=%d, want clean 3 probes", res, sent)
	}
	if d := time.Since(start); d < 750*time.Millisecond || d > 1100*time.Millisecond {
		t.Errorf("probing took %v, want 750-1000 ms (0-250 ms wait + 3x250 ms)", d)
	}
}

func TestProbeSet_ConflictAfterFirstProbe(t *testing.T) {
	res, sent := runProbeWithEvent(t, &ProbeEvent{Name: "myhost.local"}, 50*time.Millisecond)
	if !res.Conflict || res.Deferred || res.ConflictName != "myhost.local" || sent != 1 {
		t.Fatalf("result=%+v sent=%d, want conflict on myhost.local after 1 probe", res, sent)
	}
}

func TestProbeSet_LostTiebreak_Deferred(t *testing.T) {
	res, _ := runProbeWithEvent(t, &ProbeEvent{Name: "Alpha._http._tcp.local", LostTiebreak: true}, 50*time.Millisecond)
	if !res.Conflict || !res.Deferred {
		t.Fatalf("result=%+v, want deferred (lost tiebreak)", res)
	}
}

// Events from before the first probe are stale and ignored (RFC 6762 §8.1).
func TestProbeSet_StaleEventIgnored(t *testing.T) {
	events := make(chan ProbeEvent, 4)
	events <- ProbeEvent{Name: "myhost.local"}
	p := NewProber()
	p.SetTransport(transport.NewMockTransport())
	p.SetProbeSet(testProbeSet(), events)
	if res := p.Probe(context.Background(), "x"); res.Conflict {
		t.Fatalf("stale pre-probe event caused %+v", res)
	}
}
