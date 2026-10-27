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
		if hasUsableLink(linkSnapshot()) {
			r.onLinkChange()
		}
	})
}

// hasUsableLink reports whether a snapshot has an interface that is up and
// has an IPv4 address. A link that is up without one (a link-local address
// still being claimed, RFC 3927) cannot send mDNS yet.
func hasUsableLink(snapshot string) bool {
	for _, part := range strings.Split(snapshot, ";") {
		fields := strings.Fields(part)
		if len(fields) >= 3 && fields[1] == "up" {
			return true
		}
	}
	return false
}

// onLinkChange probes and announces again after a link change (RFC 6762
// §8): the host name first (renamed only if now taken), then every service
// under its current name. It runs as one tracked re-probe job, so Close
// cancels it. A newer change cancels a running job and starts again; the
// services that job had not finished are carried over, never dropped.
func (r *Responder) onLinkChange() {
	r.cancelReprobe("link")
	r.startReprobe("link", func(ctx context.Context) {
		services := append(r.takeLinkPending(), r.Services()...)
		r.setHostEstablished(false)
		if _, err := r.probeHostName(ctx); err != nil {
			r.addLinkPending(services)
			return
		}
		for i, svc := range services {
			if ctx.Err() != nil {
				r.addLinkPending(services[i:])
				return
			}
			_ = r.registry.Remove(svc.InstanceName)
			if err := r.register(ctx, svc); err != nil && ctx.Err() != nil {
				r.addLinkPending(services[i:])
				return
			}
		}
	})
}

// addLinkPending keeps services a cancelled link job had not re-registered.
func (r *Responder) addLinkPending(services []*Service) {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	r.linkPending = append(r.linkPending, services...)
}

// takeLinkPending returns and clears the carried-over services, dropping
// any that are registered again meanwhile.
func (r *Responder) takeLinkPending() []*Service {
	pending := r.swapLinkPending()
	out := pending[:0:0]
	for _, svc := range pending {
		if _, registered := r.registry.Get(svc.InstanceName); !registered {
			out = append(out, svc)
		}
	}
	return out
}

// dropLinkPending removes instance (or "<instance>.<type>") from the
// carried-over list; Unregister calls it so the service stays gone.
func (r *Responder) dropLinkPending(serviceID string) {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	kept := r.linkPending[:0]
	for _, svc := range r.linkPending {
		if !namesEqual(svc.InstanceName, serviceID) && !namesEqual(svc.InstanceName+"."+svc.ServiceType, serviceID) {
			kept = append(kept, svc)
		}
	}
	r.linkPending = kept
}

// swapLinkPending returns the carried-over list and clears it.
func (r *Responder) swapLinkPending() []*Service {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	pending := r.linkPending
	r.linkPending = nil
	return pending
}
