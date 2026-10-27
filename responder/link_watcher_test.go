package responder

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
)

// RFC 6762 §8 / BCT II.17 (hot-plugging): after a link change the responder
// probes and announces again, host name and services, keeping their names.
func TestLinkChange_ReprobesHostAndServices(t *testing.T) {
	r, mt := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})
	before := len(mt.SendCalls())

	r.onLinkChange()

	probed := map[string]bool{}
	ok := eventually(8*time.Second, func() bool {
		for _, c := range mt.SendCalls()[before:] {
			m, err := message.ParseMessage(c.Packet)
			if err != nil || m.Header.IsResponse() {
				continue
			}
			for _, q := range m.Questions {
				probed[q.QNAME] = true
			}
		}
		_, registered := r.registry.Get("Alpha")
		return probed["myhost.local"] && probed["Alpha._http._tcp.local"] && registered && r.isHostEstablished()
	})
	if !ok {
		t.Fatalf("after a link change: probed=%v, want host and Alpha probed and Alpha registered again", probed)
	}
}

// The watcher calls onLinkChange when the usable interfaces change, and not
// when they stay the same.
func TestLinkWatcher_DetectsChange(t *testing.T) {
	var mu sync.Mutex
	state := "eth0 up 169.254.20.20"
	orig := linkSnapshot
	t.Cleanup(func() { linkSnapshot = orig })
	linkSnapshot = func() string {
		mu.Lock()
		defer mu.Unlock()
		return state
	}

	changes := make(chan struct{}, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchLinks(ctx, 10*time.Millisecond, func() { changes <- struct{}{} })

	select {
	case <-changes:
		t.Fatal("change reported while the links stayed the same")
	case <-time.After(100 * time.Millisecond):
	}

	mu.Lock()
	state = "eth0 down"
	mu.Unlock()
	select {
	case <-changes:
	case <-time.After(time.Second):
		t.Fatal("link down not reported")
	}
}

// A link that is up but has no IPv4 address yet (link-local still being
// claimed) is not usable; probing starts once the address exists.
func TestHasUsableLink(t *testing.T) {
	for snap, want := range map[string]bool{
		"eth0 down ":                      false,
		"eth0 up ":                        false,
		"eth0 up 169.254.20.20":           true,
		"eth0 down ;wlan0 up 192.168.1.4": true,
	} {
		if got := hasUsableLink(snap); got != want {
			t.Errorf("hasUsableLink(%q) = %v, want %v", snap, got, want)
		}
	}
}

// A newer link change restarts a running link re-probe, so the final state
// of the link is what gets probed.
func TestLinkChange_RestartsRunningJob(t *testing.T) {
	r, _ := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})
	r.onLinkChange()
	first, ok := r.reprobeJobFor("link")
	if !ok {
		t.Fatal("no link job after the first change")
	}
	r.onLinkChange()
	second, ok := r.reprobeJobFor("link")
	if !ok || second == first {
		t.Fatal("second change did not restart the link job")
	}
	select {
	case <-first.done:
	default:
		t.Error("first link job still running after the restart")
	}
}

// Restarting the link job must not lose services the cancelled job had
// taken out of the registry to re-probe.
func TestLinkChange_RestartKeepsServices(t *testing.T) {
	r, _ := newConformanceResponder(t)
	registerOrFail(t, r, &Service{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80})
	registerOrFail(t, r, &Service{InstanceName: "Bravo", ServiceType: "_http._tcp.local", Port: 81})

	r.onLinkChange()
	// Restart while the first job is re-registering Alpha (it has removed
	// Alpha from the registry): host probe + announce take ~2 s.
	if !eventually(5*time.Second, func() bool { _, ok := r.registry.Get("Alpha"); return !ok }) {
		t.Fatal("first job never took Alpha out for re-probing")
	}
	r.onLinkChange()

	if !eventually(15*time.Second, func() bool {
		_, a := r.registry.Get("Alpha")
		_, b := r.registry.Get("Bravo")
		_, running := r.reprobeJobFor("link")
		return a && b && !running
	}) {
		t.Fatalf("services after restarted link job: %+v", r.Services())
	}
}

// Unregister also removes a service waiting to be re-registered by a link
// job, so a later link change cannot bring it back (#58 class of bug).
func TestUnregister_DropsLinkPending(t *testing.T) {
	r, _ := newConformanceResponder(t)
	r.addLinkPending([]*Service{{InstanceName: "Alpha", ServiceType: "_http._tcp.local", Port: 80}})
	_ = r.Unregister("Alpha")
	if left := r.takeLinkPending(); len(left) != 0 {
		t.Errorf("pending after Unregister = %+v, want none", left)
	}
}
