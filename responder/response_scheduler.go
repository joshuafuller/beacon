package responder

import (
	"math/rand/v2"
	"net"
	"sync"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Response delays from RFC 6762 §6 and §7.2.
const (
	sharedDelayMin    = 20 * time.Millisecond  // shared answers: 20-120 ms
	sharedDelayMax    = 120 * time.Millisecond //
	truncatedDelayMin = 400 * time.Millisecond // TC set: 400-500 ms
	truncatedDelayMax = 500 * time.Millisecond //
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
	timer          *time.Timer
}

// responseScheduler delays multicast responses that RFC 6762 says to delay,
// merges answers to back-to-back queries from the same querier into one
// packet (§6), and applies known answers that arrive in follow-up packets
// (§7.2) before it sends.
type responseScheduler struct {
	mu      sync.Mutex
	pending map[schedulerKey]*pendingResponse
	closed  bool
}

// schedulerKey identifies one querier on one transport. Known-answer lists
// are per querier (RFC 6762 §7.2: "subsequent packets from the querying host").
type schedulerKey struct {
	tr   transport.Transport // transports are pointer types, so comparable
	host string
}

func pendingKey(tr transport.Transport, src net.Addr) schedulerKey {
	host := ""
	if udpAddr, ok := src.(*net.UDPAddr); ok {
		host = udpAddr.IP.String()
	} else if src != nil {
		host = src.String()
	}
	return schedulerKey{tr: tr, host: host}
}

// schedule adds matched questions and known answers to the pending response
// for (tr, src). It creates the pending response with a delay of delay, or
// pushes an existing one out to at least delay from now when extend is set.
// It reports false when nothing could be scheduled (scheduler closed).
func (s *responseScheduler) schedule(r *Responder, tr transport.Transport, ipv6Network bool, src net.Addr, interfaceIndex int,
	matched []matchedQuestion, knownAnswers []message.Answer, delay time.Duration, extend bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	if s.pending == nil {
		s.pending = make(map[schedulerKey]*pendingResponse)
	}

	key := pendingKey(tr, src)
	p, ok := s.pending[key]
	if !ok {
		p = &pendingResponse{tr: tr, ipv6Network: ipv6Network, src: src, interfaceIndex: interfaceIndex}
		s.pending[key] = p
		p.timer = time.AfterFunc(delay, func() { s.fire(r, key, p) })
	} else if extend {
		p.timer.Reset(delay)
	}
	p.matched = append(p.matched, matched...)
	p.knownAnswers = append(p.knownAnswers, knownAnswers...)
	return true
}

// addToPending merges matched questions into an already-pending response for
// (tr, src) without changing its delay. It reports whether one was pending.
func (s *responseScheduler) addToPending(tr transport.Transport, src net.Addr, matched []matchedQuestion, knownAnswers []message.Answer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pending[pendingKey(tr, src)]
	if !ok {
		return false
	}
	p.matched = append(p.matched, matched...)
	p.knownAnswers = append(p.knownAnswers, knownAnswers...)
	return true
}

// addKnownAnswers attaches a follow-up known-answer packet to the pending
// response for (tr, src). It reports whether such a response was pending.
func (s *responseScheduler) addKnownAnswers(tr transport.Transport, src net.Addr, knownAnswers []message.Answer, truncated bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pending[pendingKey(tr, src)]
	if !ok {
		return false
	}
	p.knownAnswers = append(p.knownAnswers, knownAnswers...)
	if truncated {
		// RFC 6762 §7.2: keep a 400-500 ms pause after the last TC packet.
		p.timer.Reset(randomDelay(truncatedDelayMin, truncatedDelayMax))
	}
	return true
}

// fire sends the pending response once its delay has passed.
func (s *responseScheduler) fire(r *Responder, key schedulerKey, p *pendingResponse) {
	if !s.take(key, p) {
		return
	}
	query := &message.DNSMessage{Answers: p.knownAnswers}
	r.respondToQuery(p.tr, p.ipv6Network, query, p.matched, p.src, p.interfaceIndex)
}

// take removes p from the pending set and reports whether it was still due
// (not closed, not already taken).
func (s *responseScheduler) take(key schedulerKey, p *pendingResponse) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.pending[key] != p {
		return false
	}
	delete(s.pending, key)
	return true
}

// close stops every pending response; nothing is sent after close returns.
func (s *responseScheduler) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for key, p := range s.pending {
		p.timer.Stop()
		delete(s.pending, key)
	}
}
