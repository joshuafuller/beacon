package responder

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

// RFC 6762 §8: "Whenever a Multicast DNS responder starts up, wakes up from
// sleep, receives an indication of a network interface 'Link Change' event,
// or has any other reason to believe that its network connectivity may have
// changed in some relevant way, it MUST perform the two startup steps below:
// Probing and Announcing." BCT II.17 (hot-plugging) checks it.

// linkPollInterval is how often the usable interfaces are compared.
const linkPollInterval = time.Second

// linkSnapshot describes the interfaces mDNS can use: up, running (carrier),
// multicast-capable, not loopback, with their IPv4 addresses. Two equal
// snapshots mean "no relevant change". A variable so tests can drive it.
var linkSnapshot = func() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "error"
	}
	var parts []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}
		state := "down"
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagRunning != 0 {
			state = "up"
		}
		var addrs []string
		if list, err := iface.Addrs(); err == nil {
			for _, a := range list {
				if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					addrs = append(addrs, ipnet.IP.String())
				}
			}
		}
		sort.Strings(addrs)
		parts = append(parts, fmt.Sprintf("%s %s %s", iface.Name, state, strings.Join(addrs, ",")))
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// watchLinks calls onChange each time linkSnapshot differs from the previous
// one, until ctx ends.
func watchLinks(ctx context.Context, every time.Duration, onChange func()) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	prev := linkSnapshot()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if cur := linkSnapshot(); cur != prev {
				prev = cur
				onChange()
			}
		}
	}
}

// runLinkWatcher runs watchLinks until Close; the caller did reprobeWg.Add.
func (r *Responder) runLinkWatcher() {
	defer r.reprobeWg.Done()
	watchLinks(r.lifeCtx, linkPollInterval, func() {
		// Probing while no link is up would only lose packets; the change
		// that brings a link back up triggers the re-probe.
		if strings.Contains(linkSnapshot(), " up ") {
			r.onLinkChange()
		}
	})
}

// onLinkChange probes and announces again after a link change (RFC 6762
// §8): the host name first (renamed only if now taken), then every service
// under its current name. It runs as one tracked re-probe job, so Close
// cancels it; a second change while it runs is covered by the running job.
func (r *Responder) onLinkChange() {
	r.startReprobe("link", func(ctx context.Context) {
		r.setHostEstablished(false)
		if _, err := r.probeHostName(ctx); err != nil {
			return
		}
		for _, svc := range r.Services() {
			if ctx.Err() != nil {
				return
			}
			_ = r.registry.Remove(svc.InstanceName)
			_ = r.register(ctx, svc)
		}
	})
}
