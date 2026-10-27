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
