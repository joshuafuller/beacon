package responder

import (
	"math/rand/v2"
	"net"
	"sync"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/responder"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Response delays from RFC 6762 §6 and §7.2.
const (
	sharedDelayMin    = 20 * time.Millisecond  // shared answers: 20-120 ms
	sharedDelayMax    = 120 * time.Millisecond //
	truncatedDelayMin = 400 * time.Millisecond // TC set: 400-500 ms
	truncatedDelayMax = 500 * time.Millisecond //
)

// Limits so that a querier (or a spoofed source) cannot hold or grow delayed
// responses without bound.
const (
	// maxPendingDelay caps how long TC follow-ups can push a response out.
	// RFC 6762 §7.2 lets a responder extend the delay "as necessary"; one
	// second covers any real multipacket known-answer list.
	maxPendingDelay = time.Second
	// maxPendingResponses caps concurrent delayed responses (one per querier).
	maxPendingResponses = 256
	// maxPendingQuestions / maxPendingKnownAnswers cap one pending response.
	maxPendingQuestions    = 64
	maxPendingKnownAnswers = 512
)

// randomDelay returns a uniformly random duration in [lo, hi]. The jitter
// only spreads responses from many hosts in time; it is not a security value.
func randomDelay(lo, hi time.Duration) time.Duration {
	return lo + rand.N(hi-lo+1) // #nosec G404 -- timing jitter, not security
}

// pendingResponse collects the questions and known answers that one querier
// sent while our multicast response to it is delayed.
type pendingResponse struct {
	tr             transport.Transport
	ipv6Network    bool
	src            net.Addr
	interfaceIndex int
	matched        []matchedQuestion
	knownAnswers   []message.Answer
	truncated      bool      // started by a TC query: follow-up known answers allowed
	created        time.Time // first packet; the response is never held past created+maxPendingDelay
	timer          *time.Timer
}

// responseScheduler delays multicast responses that RFC 6762 says to delay,
// merges answers to back-to-back queries from the same querier into one
// packet (§6), and applies known answers that arrive in follow-up packets
// (§7.2) before it sends.
type responseScheduler struct {
	mu       sync.Mutex
	pending  map[schedulerKey]*pendingResponse
	closed   bool
	inFlight sync.WaitGroup // fire calls between take and send; close waits for them
}

// schedulerKey identifies one querier on one interface of one transport.
// Known-answer lists are per querier (RFC 6762 §7.2: "subsequent packets from
// the querying host"); the interface and IPv6 zone keep link-local sources on
// different links apart.
type schedulerKey struct {
	tr             transport.Transport // transports are pointer types, so comparable
	host           string              // source IP including IPv6 zone
	interfaceIndex int
}

func pendingKey(tr transport.Transport, src net.Addr, interfaceIndex int) schedulerKey {
	host := ""
	if udpAddr, ok := src.(*net.UDPAddr); ok {
		host = udpAddr.IP.String()
		if udpAddr.Zone != "" {
			host += "%" + udpAddr.Zone
		}
	} else if src != nil {
		host = src.String()
	}
	return schedulerKey{tr: tr, host: host, interfaceIndex: interfaceIndex}
}

// schedule adds matched questions and known answers to the pending response
// for this querier, creating it with the given delay. truncated marks a TC
// query: it allows follow-up known answers and pushes the response out to at
// least delay from now (capped at maxPendingDelay after the first packet).
// It reports false when the caller must answer at once instead (scheduler
// closed or full).
func (s *responseScheduler) schedule(r *Responder, tr transport.Transport, ipv6Network bool, src net.Addr, interfaceIndex int,
	matched []matchedQuestion, knownAnswers []message.Answer, delay time.Duration, truncated bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	if s.pending == nil {
		s.pending = make(map[schedulerKey]*pendingResponse)
	}

	key := pendingKey(tr, src, interfaceIndex)
	p, ok := s.pending[key]
	if !ok {
		if len(s.pending) >= maxPendingResponses {
			return false
		}
		p = &pendingResponse{tr: tr, ipv6Network: ipv6Network, src: src, interfaceIndex: interfaceIndex, created: time.Now()}
		s.pending[key] = p
		p.timer = time.AfterFunc(delay, func() { s.fire(r, key, p) })
	} else if truncated {
		p.timer.Reset(p.capDelay(delay))
	}
	p.truncated = p.truncated || truncated
	p.add(matched, knownAnswers)
	return true
}

// capDelay limits delay so the response fires no later than
// created+maxPendingDelay.
func (p *pendingResponse) capDelay(delay time.Duration) time.Duration {
	if left := time.Until(p.created.Add(maxPendingDelay)); delay > left {
		return max(left, 0)
	}
	return delay
}

// add appends questions and known answers within the per-response caps.
func (p *pendingResponse) add(matched []matchedQuestion, knownAnswers []message.Answer) {
	if room := maxPendingQuestions - len(p.matched); room > 0 {
		p.matched = append(p.matched, matched[:min(room, len(matched))]...)
	}
	if room := maxPendingKnownAnswers - len(p.knownAnswers); room > 0 {
		p.knownAnswers = append(p.knownAnswers, knownAnswers[:min(room, len(knownAnswers))]...)
	}
}

// addToPending merges matched questions into an already-pending response for
// this querier without changing its delay. It reports whether one was pending.
func (s *responseScheduler) addToPending(tr transport.Transport, src net.Addr, interfaceIndex int, matched []matchedQuestion, knownAnswers []message.Answer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pending[pendingKey(tr, src, interfaceIndex)]
	if !ok {
		return false
	}
	p.add(matched, knownAnswers)
	return true
}

// addKnownAnswers attaches a follow-up known-answer packet (no questions) to
// the pending response for this querier. RFC 6762 §7.2: such packets follow a
// query with TC set, so they are ignored unless the pending response was
// started by one. It reports whether they were attached.
func (s *responseScheduler) addKnownAnswers(tr transport.Transport, src net.Addr, interfaceIndex int, knownAnswers []message.Answer, truncated bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pending[pendingKey(tr, src, interfaceIndex)]
	if !ok || !p.truncated {
		return false
	}
	p.add(nil, knownAnswers)
	if truncated {
		// RFC 6762 §7.2: keep a 400-500 ms pause after the last TC packet,
		// within maxPendingDelay of the first packet.
		p.timer.Reset(p.capDelay(randomDelay(truncatedDelayMin, truncatedDelayMax)))
	}
	return true
}

// fire sends the pending response once its delay has passed. Services that
// were unregistered during the delay are dropped, so a goodbye is never
// followed by a stale answer.
func (s *responseScheduler) fire(r *Responder, key schedulerKey, p *pendingResponse) {
	if !s.take(key, p) {
		return
	}
	defer s.inFlight.Done()

	matched := r.refreshMatched(p.matched)
	if len(matched) == 0 {
		return
	}
	query := &message.DNSMessage{Answers: p.knownAnswers}
	r.respondToQuery(p.tr, p.ipv6Network, query, matched, p.src, p.interfaceIndex)
}

// take removes p from the pending set and reports whether it was still due
// (not closed, not already taken). On true the caller must call inFlight.Done.
func (s *responseScheduler) take(key schedulerKey, p *pendingResponse) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.pending[key] != p {
		return false
	}
	delete(s.pending, key)
	s.inFlight.Add(1)
	return true
}

// close stops every pending response and waits for any that are being sent;
// nothing is sent after close returns.
func (s *responseScheduler) close() {
	s.stopAll()
	s.inFlight.Wait()
}

func (s *responseScheduler) stopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for key, p := range s.pending {
		p.timer.Stop()
		delete(s.pending, key)
	}
}

// refreshMatched re-reads every matched service from the registry and drops
// services that are no longer registered.
func (r *Responder) refreshMatched(matched []matchedQuestion) []matchedQuestion {
	out := make([]matchedQuestion, 0, len(matched))
	for _, m := range matched {
		var services []*responder.Service
		for _, svc := range m.services {
			if current, ok := r.registry.Get(svc.InstanceName); ok {
				services = append(services, current)
			}
		}
		if len(services) > 0 {
			out = append(out, matchedQuestion{question: m.question, services: services})
		}
	}
	return out
}
