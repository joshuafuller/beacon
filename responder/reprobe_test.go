package responder

import (
	"bytes"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
)

// Tests for #53, #54 and #55 (security review of the conflict handling).

func registerOrFail(t *testing.T, r *Responder, svc *Service) {
	t.Helper()
	if err := r.Register(svc); err != nil {
		t.Fatalf("Register(%q): %v", svc.InstanceName, err)
	}
}

func eventually(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}

// #53: a lost tiebreak counts toward the 15-in-10s probe rate limit.
func TestRegister_LostTiebreakCountsTowardRateLimit(t *testing.T) {
	r, mt := newConformanceResponder(t)
	svc := &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}
	done := registerAsync(t, r, mt, svc, func() {
		q := buildTestQuery(t, 0, testQuestion{"Alpha._http._tcp.local", 255})
		q[9] = 1
		enc, _ := message.EncodeDomainName("Alpha._http._tcp.local")
		target, _ := message.EncodeName("other.local")
		rdata := append([]byte{0, 0, 0, 0, 0xff, 0xff}, target...)
		q = append(q, enc...)
		q = append(q, 0, 33, 0, 1, 0, 0, 0, 120, 0, byte(len(rdata)))
		q = append(q, rdata...)
		_ = r.handleQuery(mt, q, otherHost, 0)
	})
	waitDone(t, done)
	r.hostMu.Lock()
	n := len(r.conflictTimes)
	r.hostMu.Unlock()
	if n != 1 {
		t.Errorf("conflicts recorded = %d after one lost tiebreak, want 1", n)
	}
}

// #54: Unregister during a re-probe cancels it; the service stays gone.
func TestReprobe_UnregisterCancels(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})

	_ = r.handleQuery(mt, buildResponse(t, "Alpha._http._tcp.local", 16, []byte{3, 'x', '=', '1'}), otherHost, 0)
	if err := r.Unregister("Alpha"); err != nil {
		t.Fatalf("Unregister during re-probe: %v", err)
	}
	time.Sleep(3 * time.Second)
	if _, ok := r.registry.Get("Alpha"); ok {
		t.Error("Alpha registered again after Unregister cancelled its re-probe")
	}
}

// #54: Close during a re-probe returns, and nothing is sent after it.
func TestReprobe_CloseCancels(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})

	_ = r.handleQuery(mt, buildResponse(t, "Alpha._http._tcp.local", 16, []byte{3, 'x', '=', '1'}), otherHost, 0)
	time.Sleep(50 * time.Millisecond)
	closed := make(chan struct{})
	go func() { _ = r.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("Close did not return during a re-probe")
	}
	n := len(mt.SendCalls())
	time.Sleep(2 * time.Second)
	if m := len(mt.SendCalls()); m != n {
		t.Errorf("%d packets sent after Close returned", m-n)
	}
}

// #55: a host conflict probes only the host name; services stay registered,
// including one with its own host name.
func TestReprobeHost_KeepsServices(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})
	registerOrFail(t, r, &Service{InstanceName: "Custom", ServiceType: "_http._tcp.local", Port: 81, Hostname: "custom.local"})
	before := len(mt.SendCalls())

	_ = r.handleQuery(mt, buildResponse(t, "myhost.local", 1, []byte{192, 168, 1, 77}), otherHost, 0)
	if !eventually(3*time.Second, func() bool { return len(mt.SendCalls()) > before && r.isHostEstablished() }) {
		t.Fatal("host name was not re-probed")
	}
	for _, name := range []string{"Alpha", "Custom"} {
		if _, ok := r.registry.Get(name); !ok {
			t.Errorf("%s left the registry during the host re-probe", name)
		}
	}
	probe, _ := message.ParseMessage(mt.SendCalls()[before].Packet)
	if len(probe.Questions) != 1 || probe.Questions[0].QNAME != "myhost.local" {
		t.Errorf("re-probe questions %+v, want only myhost.local", probe.Questions)
	}
}

// #55: when the re-probe renames the host, services are announced again
// with the new SRV target.
func TestReprobeHost_RenameUpdatesSRV(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})
	before := len(mt.SendCalls())

	conflict := buildResponse(t, "myhost.local", 1, []byte{192, 168, 1, 77})
	_ = r.handleQuery(mt, conflict, otherHost, 0)
	// Defend the name again while it is being re-probed: the host must rename.
	eventually(2*time.Second, func() bool { return len(mt.SendCalls()) > before })
	_ = r.handleQuery(mt, conflict, otherHost, 0)

	newTarget, _ := message.EncodeName("myhost-2.local")
	if !eventually(6*time.Second, func() bool {
		for _, c := range mt.SendCalls()[before:] {
			m, err := message.ParseMessage(c.Packet)
			if err != nil || !m.Header.IsResponse() {
				continue
			}
			for _, a := range m.Answers {
				if a.TYPE == 33 && bytes.HasSuffix(a.RDATA, newTarget) {
					return true
				}
			}
		}
		return false
	}) {
		t.Fatalf("no SRV announcement with target myhost-2.local (host now %q)", r.hostName())
	}
}

// Review of #54: Unregister with the full service ID also cancels a re-probe.
func TestReprobe_UnregisterByFullIDCancels(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})

	_ = r.handleQuery(mt, buildResponse(t, "Alpha._http._tcp.local", 16, []byte{3, 'x', '=', '1'}), otherHost, 0)
	eventually(time.Second, func() bool { _, ok := r.reprobeJobFor("svc:Alpha"); return ok })
	if err := r.Unregister("Alpha._http._tcp.local"); err != nil {
		t.Fatalf("Unregister(full ID) during re-probe: %v", err)
	}
	time.Sleep(3 * time.Second)
	if _, ok := r.registry.Get("Alpha"); ok {
		t.Error("Alpha registered again after Unregister(full ID)")
	}
}

// Review of #54: no re-probe can start while the service is being
// unregistered.
func TestUnregister_BlocksNewReprobe(t *testing.T) {
	r, _ := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})

	r.beginUnregister("Alpha")
	r.reprobeService("Alpha")
	if _, running := r.reprobeJobFor("svc:Alpha"); running {
		t.Error("a re-probe started while Alpha was being unregistered")
	}
	r.endUnregister("Alpha")
}
