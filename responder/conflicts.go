package responder

import (
	"bytes"
	"context"
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
func (r *Responder) waitProbeRateLimit(ctx context.Context) error {
	if !r.overConflictLimit() {
		return nil
	}
	return r.sleep(ctx, conflictCooldown)
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

// sleep waits d, returning early with an error when ctx ends or the
// responder closes.
func (r *Responder) sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
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

// reprobeJob is one running RFC 6762 §9 re-probe. Unregister and Close
// cancel it and wait for done (#54).
type reprobeJob struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// startReprobe registers a re-probe job for key and runs fn in a tracked
// goroutine with a context that Close and Unregister can cancel. It does
// nothing if a re-probe for key is already running or the responder closed.
func (r *Responder) startReprobe(key string, fn func(ctx context.Context)) bool {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	if r.runCtx().Err() != nil {
		return false
	}
	if r.reprobing == nil {
		r.reprobing = make(map[string]*reprobeJob)
	}
	if _, running := r.reprobing[key]; running {
		return false
	}
	ctx, cancel := context.WithCancel(r.runCtx())
	job := &reprobeJob{cancel: cancel, done: make(chan struct{})}
	r.reprobing[key] = job
	r.reprobeWg.Add(1)
	go func() {
		defer r.reprobeWg.Done()
		defer close(job.done)
		defer r.endReprobe(key, job)
		defer cancel()
		fn(ctx)
	}()
	return true
}

func (r *Responder) endReprobe(key string, job *reprobeJob) {
	r.hostMu.Lock()
	defer r.hostMu.Unlock()
	if r.reprobing[key] == job {
		delete(r.reprobing, key)
	}
}

// cancelReprobe stops the re-probe for key, if any, and waits for it. It
// reports whether one was running.
func (r *Responder) cancelReprobe(key string) bool {
	r.hostMu.Lock()
	job, ok := r.reprobing[key]
	r.hostMu.Unlock()
	if !ok {
		return false
	}
	job.cancel()
	<-job.done
	return true
}

// cancelReprobes stops every re-probe and waits for them (Close).
func (r *Responder) cancelReprobes() {
	r.hostMu.Lock()
	jobs := make([]*reprobeJob, 0, len(r.reprobing))
	for _, job := range r.reprobing {
		jobs = append(jobs, job)
	}
	r.hostMu.Unlock()
	for _, job := range jobs {
		job.cancel()
	}
	r.reprobeWg.Wait()
}

// reprobeService takes a conflicted instance back to probing (RFC 6762 §9):
// it leaves the registry (its records are no longer verified unique) and is
// registered again, same name first; Register renames it only if the
// conflict persists (BCT II.6 warns when a device renames without
// re-probing).
func (r *Responder) reprobeService(instance string) {
	svc, ok := r.registry.Get(instance)
	if !ok {
		return
	}
	r.startReprobe("svc:"+instance, func(ctx context.Context) {
		_ = r.registry.Remove(instance)
		_ = r.register(ctx, fromInternalService(svc))
	})
}

// reprobeHost takes only the host name back to probing (RFC 6762 §9).
// Services stay registered and answered (#55). If the host must be renamed,
// every service is announced again so its SRV record names the new host.
func (r *Responder) reprobeHost() {
	r.startReprobe("host", func(ctx context.Context) {
		r.setHostEstablished(false)
		renamed, err := r.probeHostName(ctx)
		if err != nil || !renamed {
			return
		}
		r.announceServices(ctx)
	})
}

// probeHostName probes the current host name until it is unique, renaming
// it on conflicts (with the same rate limit and caps as Register). It
// reports whether the name changed.
func (r *Responder) probeHostName(ctx context.Context) (bool, error) {
	renamed := false
	for attempts := 0; attempts < maxNetworkRenameAttempts; attempts++ {
		if err := r.waitProbeRateLimit(ctx); err != nil {
			return renamed, err
		}
		host := r.hostName()
		hostRecords := hostAddressRecords(host)
		session := r.probes.start(state.ProbeSet{Names: []string{host}, Authority: hostRecords})
		finalState, result, err := r.runProbeAndAnnounce(ctx, hostRecords, host, session)
		r.probes.stop(session)
		if err != nil {
			return renamed, err
		}
		if finalState != state.StateConflictDetected {
			r.setHostEstablished(true)
			return renamed, nil
		}
		r.noteConflict()
		if result.Deferred {
			if err := r.sleep(ctx, time.Second); err != nil {
				return renamed, err
			}
			continue
		}
		r.renameHost(host)
		renamed = true
	}
	return renamed, fmt.Errorf("max rename attempts (%d) exceeded for host name", maxNetworkRenameAttempts)
}

// hostAddressRecords returns the unique address records for host.
func hostAddressRecords(host string) []*message.ResourceRecord {
	ipv4, err := getLocalIPv4()
	if err != nil {
		return nil
	}
	info := buildServiceInfo("host", "_host._tcp.local", host, 0, ipv4, nil)
	var out []*message.ResourceRecord
	for _, rr := range records.BuildRecordSet(info) {
		if rr.Type == protocol.RecordTypeA || rr.Type == protocol.RecordTypeAAAA {
			out = append(out, rr)
		}
	}
	return out
}

// announceServices announces every registered service again (RFC 6762
// §8.3), e.g. after the host was renamed and SRV targets changed.
func (r *Responder) announceServices(ctx context.Context) {
	ipv4, err := getLocalIPv4()
	if err != nil {
		return
	}
	for _, instance := range r.registry.List() {
		svc, ok := r.registry.Get(instance)
		if !ok {
			continue
		}
		info := buildServiceInfo(svc.InstanceName, svc.ServiceType, r.hostName(), svc.Port, ipv4, svc.TXT)
		if err := r.announce(ctx, records.BuildRecordSet(info)); err != nil {
			return
		}
	}
}

// announce sends RFC 6762 §8.3 announcements for recordSet without probing
// (the names are already verified unique).
func (r *Responder) announce(ctx context.Context, recordSet []*message.ResourceRecord) error {
	tr, err := r.registrationTransport(recordSet)
	if err != nil {
		return err
	}
	announcer := state.NewAnnouncer()
	announcer.SetTransport(tr)
	announcer.SetRecords(recordSet)
	return announcer.Announce(ctx, "", nil)
}
