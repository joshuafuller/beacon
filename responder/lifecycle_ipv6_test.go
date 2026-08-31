package responder

import (
	"context"
	"net"
	"testing"

	internalresponder "github.com/joshuafuller/beacon/internal/responder"
)

// RFC 6762 §20: A dual-stack host should register its names in both the IPv4
// and IPv6 .local zones. RFC 6762 §8.4 applies that rule to record updates.
func TestUpdateServiceAnnouncesOnBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	ipv4Transport := &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
		ipv4Sends++
		return nil
	}}
	ipv6Transport := &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
		ipv6Sends++
		return nil
	}}
	r := &Responder{
		ctx:        context.Background(),
		transport:  ipv4Transport,
		transport6: ipv6Transport,
		registry:   internalresponder.NewRegistry(),
		hostname:   "host.local",
	}
	if err := r.registry.Register(&internalresponder.Service{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}); err != nil {
		t.Fatalf("register fixture: %v", err)
	}

	if err := r.UpdateService("Example", map[string]string{"version": "2"}); err != nil {
		t.Fatalf("UpdateService() failed: %v", err)
	}
	if ipv4Sends != 1 || ipv6Sends != 1 {
		t.Fatalf("update sends: IPv4=%d IPv6=%d, want one announcement in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends)
	}
}

// RFC 6762 §§10.1, 20: A dual-stack responder sends goodbye records into
// both independent .local zones when a service is unregistered.
func TestUnregisterSendsGoodbyeOnBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	r := &Responder{
		ctx: context.Background(),
		transport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv4Sends++
			return nil
		}},
		transport6: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv6Sends++
			return nil
		}},
		registry: internalresponder.NewRegistry(),
		hostname: "host.local",
	}
	if err := r.registry.Register(&internalresponder.Service{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}); err != nil {
		t.Fatalf("register fixture: %v", err)
	}

	if err := r.Unregister("Example"); err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}
	if ipv4Sends != 1 || ipv6Sends != 1 {
		t.Fatalf("goodbye sends: IPv4=%d IPv6=%d, want one goodbye in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends)
	}
}

// RFC 6762 §§8.1, 8.3, 20: A dual-stack registration probes and announces
// in both independent .local zones before becoming established.
func TestProbeAndAnnounceUsesBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	r := &Responder{
		ctx: context.Background(),
		transport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv4Sends++
			return nil
		}},
		transport6: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv6Sends++
			return nil
		}},
	}

	if _, err := r.runProbeAndAnnounce(nil, "Example._http._tcp.local"); err != nil {
		t.Fatalf("runProbeAndAnnounce() failed: %v", err)
	}
	const lifecyclePackets = 5 // Three probes plus two announcements.
	if ipv4Sends != lifecyclePackets || ipv6Sends != lifecyclePackets {
		t.Fatalf("lifecycle sends: IPv4=%d IPv6=%d, want %d in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends, lifecyclePackets)
	}
}
