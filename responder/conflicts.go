package responder

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/records"
	"github.com/joshuafuller/beacon/internal/state"
)

// Probe rate limit, RFC 6762 §8.1: "If fifteen conflicts occur within any
// ten-second period, then the host MUST wait at least five seconds before
// each successive additional probe attempt."
const (
	conflictWindow   = 10 * time.Second
	conflictBurst    = 15
	conflictCooldown = 5 * time.Second
)

// probeSetFor builds the RFC 6762 §8.1 probe for a service: the instance
// name (and the host name when it is not yet established) asked as ANY, with
// the unique records proposed for those names in the Authority section.
func probeSetFor(serviceName, hostname string, probeHost bool, recordSet []*message.ResourceRecord) state.ProbeSet {
	ps := state.ProbeSet{Names: []string{serviceName}}
	if probeHost {
		ps.Names = append(ps.Names, hostname)
	}
	for _, rr := range recordSet {
		switch rr.Type {
		case protocol.RecordTypeSRV, protocol.RecordTypeTXT:
			if namesEqual(rr.Name, serviceName) {
				ps.Authority = append(ps.Authority, rr)
			}
		case protocol.RecordTypeA, protocol.RecordTypeAAAA:
			if probeHost && namesEqual(rr.Name, hostname) {
				ps.Authority = append(ps.Authority, rr)
			}
		}
	}
	return ps
}

func (r *Responder) isHostEstablished() bool {
	r.hostMu.RLock()
	defer r.hostMu.RUnlock()
	return r.hostEstablished
}

func (r *Responder) setHostEstablished(v bool) {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	r.hostEstablished = v
}

var renameSuffix = regexp.MustCompile(`^(.+)-(\d+)$`)

// renameHost picks the next host name after a conflict ("myhost.local" →
// "myhost-2.local" → "myhost-3.local", RFC 6762 §9) and makes it current
// when old is still the current name. It returns the new name.
func (r *Responder) renameHost(old string) string {
	label := strings.TrimSuffix(strings.TrimSuffix(old, "."), ".local")
	if m := renameSuffix.FindStringSubmatch(label); m != nil {
		n, _ := strconv.Atoi(m[2]) // nosemgrep: beacon-error-swallowing -- regex guarantees digits
		label = fmt.Sprintf("%s-%d", m[1], n+1)
	} else {
		label += "-2"
	}
	if len(label) > protocol.MaxLabelLength {
		label = label[len(label)-protocol.MaxLabelLength:]
	}
	next := label + ".local"

	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	if namesEqual(r.hostname, old) {
		r.hostname = next
		r.hostEstablished = false
	}
	return next
}

// noteConflict records a probe conflict for the rate limit.
func (r *Responder) noteConflict() {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	r.conflictTimes = append(r.conflictTimes, time.Now())
}

// waitProbeRateLimit waits conflictCooldown before a probe attempt when
// conflictBurst conflicts happened within conflictWindow (RFC 6762 §8.1).
func (r *Responder) waitProbeRateLimit() error {
	if !r.overConflictLimit() {
		return nil
	}
	return r.sleep(conflictCooldown)
}

func (r *Responder) overConflictLimit() bool {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	cutoff := time.Now().Add(-conflictWindow)
	kept := r.conflictTimes[:0]
	for _, t := range r.conflictTimes {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	r.conflictTimes = kept
	return len(kept) >= conflictBurst
}

// sleep waits d, returning early with an error when the responder closes.
func (r *Responder) sleep(d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-r.ctx.Done():
		return r.ctx.Err()
	case <-r.queryHandlerDone:
		return fmt.Errorf("responder closed")
	case <-timer.C:
		return nil
	}
}

// checkEstablishedConflicts handles RFC 6762 §9 for names already in use:
// a response with a unique record (SRV/TXT of a registered instance, A/AAAA
// of our host name) whose data differs from ours means another host claims
// the name. The name goes back to probing; Register then renames it if the
// conflict persists. msg's RDATA must be decompressed.
func (r *Responder) checkEstablishedConflicts(msg *message.DNSMessage) {
	host := r.hostName()
	for _, section := range [][]message.Answer{msg.Answers, msg.Additionals} {
		for _, rr := range section {
			if rr.TTL == 0 {
				continue // goodbye, not a claim
			}
			switch rr.TYPE {
			case uint16(protocol.RecordTypeA), uint16(protocol.RecordTypeAAAA):
				if r.isHostEstablished() && namesEqual(rr.NAME, host) && !isOwnAddress(rr.RDATA) {
					r.reprobeHost()
				}
			case uint16(protocol.RecordTypeSRV), uint16(protocol.RecordTypeTXT):
				if instance, ok := r.conflictingInstance(rr); ok {
					r.reprobeService(instance)
				}
			}
		}
	}
}

// conflictingInstance returns the registered instance whose SRV/TXT record
// has rr's name but different data.
func (r *Responder) conflictingInstance(rr message.Answer) (string, bool) {
	for _, instance := range r.registry.List() {
		svc, ok := r.registry.Get(instance)
		if !ok || !namesEqual(rr.NAME, svc.InstanceName+"."+svc.ServiceType) {
			continue
		}
		info := buildServiceInfo(svc.InstanceName, svc.ServiceType, r.hostName(), svc.Port, nil, svc.TXT)
		for _, ours := range records.BuildRecordSet(info) {
			if uint16(ours.Type) == rr.TYPE && namesEqual(ours.Name, rr.NAME) {
				return instance, !bytes.Equal(ours.Data, rr.RDATA)
			}
		}
	}
	return "", false
}

// isOwnAddress reports whether rdata (an A or AAAA record) is an address of
// this host.
func isOwnAddress(rdata []byte) bool {
	ip := net.IP(rdata)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return true // cannot tell: do not start a conflict on our own data
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.Equal(ip) {
			return true
		}
	}
	return false
}

// startReprobe marks key as being re-probed; false if it already is.
func (r *Responder) startReprobe(key string) bool {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	if r.reprobing == nil {
		r.reprobing = make(map[string]bool)
	}
	if r.reprobing[key] {
		return false
	}
	r.reprobing[key] = true
	return true
}

func (r *Responder) endReprobe(key string) {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	delete(r.reprobing, key)
}

// reprobeService takes a conflicted instance back to probing (RFC 6762 §9).
// Register probes the same name first and renames only if the conflict
// persists (BCT II.6 warns when a device renames without re-probing).
func (r *Responder) reprobeService(instance string) {
	if !r.startReprobe("svc:" + instance) {
		return
	}
	svc, ok := r.registry.Get(instance)
	if !ok {
		r.endReprobe("svc:" + instance)
		return
	}
	_ = r.registry.Remove(instance)
	go func() {
		defer r.endReprobe("svc:" + instance)
		_ = r.Register(fromInternalService(svc))
	}()
}

// reprobeHost takes the host name back to probing (RFC 6762 §9). Services
// are removed and registered again; the first registration probes the host
// name, renames it if the conflict persists, and the rest follow it.
func (r *Responder) reprobeHost() {
	if !r.startReprobe("host") {
		return
	}
	r.setHostEstablished(false)
	var services []*Service
	for _, instance := range r.registry.List() {
		if svc, ok := r.registry.Get(instance); ok {
			services = append(services, fromInternalService(svc))
			_ = r.registry.Remove(instance)
		}
	}
	go func() {
		defer r.endReprobe("host")
		for _, svc := range services {
			_ = r.Register(svc)
		}
	}()
}
