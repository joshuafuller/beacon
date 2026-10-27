package responder

import (
	"bytes"
	"sort"
	"sync"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/state"
)

// probeSession is one in-progress probe (RFC 6762 §8.1): the names being
// probed, the records proposed for them, and the channel the prober reads.
type probeSession struct {
	set    state.ProbeSet
	events chan state.ProbeEvent
}

// probeTracker routes packets the query handler receives to the probes that
// are running, so probing does not need its own reader on the shared socket.
type probeTracker struct {
	mu       sync.Mutex
	sessions map[*probeSession]struct{}
}

func (t *probeTracker) start(set state.ProbeSet) *probeSession {
	s := &probeSession{set: set, events: make(chan state.ProbeEvent, 8)}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sessions == nil {
		t.sessions = make(map[*probeSession]struct{})
	}
	t.sessions[s] = struct{}{}
	return s
}

func (t *probeTracker) stop(s *probeSession) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, s)
}

func (t *probeTracker) active() []*probeSession {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]*probeSession, 0, len(t.sessions))
	for s := range t.sessions {
		out = append(out, s)
	}
	return out
}

func (s *probeSession) send(ev state.ProbeEvent) {
	select {
	case s.events <- ev:
	default: // the prober already has an event to act on
	}
}

// probedName returns the probed name equal to name, if any.
func (s *probeSession) probedName(name string) (string, bool) {
	for _, n := range s.set.Names {
		if namesEqual(n, name) {
			return n, true
		}
	}
	return "", false
}

// proposes reports whether rr is identical to a record this probe proposes.
// An identical record is our own (looped back) or a proxy for us, not a
// conflict (RFC 6762 §9: conflicts are records with different rdata).
func (s *probeSession) proposes(rr message.Answer) bool {
	for _, ours := range s.set.Authority {
		if namesEqual(ours.Name, rr.NAME) && uint16(ours.Type) == rr.TYPE && bytes.Equal(ours.Data, rr.RDATA) {
			return true
		}
	}
	return false
}

// onResponse handles a response packet: during probing, any record (TTL > 0)
// whose name is being probed and that is not ours is a conflict (RFC 6762
// §8.1: "any answer containing a record with that name, of any type, MUST be
// considered a conflicting response"). msg's RDATA must be decompressed.
func (t *probeTracker) onResponse(msg *message.DNSMessage) {
	for _, s := range t.active() {
		for _, section := range [][]message.Answer{msg.Answers, msg.Additionals} {
			for _, rr := range section {
				if rr.TTL == 0 || s.proposes(rr) {
					continue
				}
				if name, ok := s.probedName(rr.NAME); ok {
					s.send(state.ProbeEvent{Name: name})
				}
			}
		}
	}
}

// onProbe handles a query that carries proposed records in its Authority
// section (another host probing). For each name we are also probing, the
// two record sets are compared as RFC 6762 §8.2 describes; if theirs is
// lexicographically later we lose: defer one second and probe again.
func (t *probeTracker) onProbe(msg *message.DNSMessage) {
	for _, s := range t.active() {
		for _, q := range msg.Questions {
			name, ok := s.probedName(q.QNAME)
			if !ok {
				continue
			}
			var theirs, ours []tiebreakRecord
			for _, rr := range msg.Authorities {
				if namesEqual(rr.NAME, name) {
					theirs = append(theirs, tiebreakRecord{class: rr.CLASS &^ 0x8000, rrtype: rr.TYPE, rdata: rr.RDATA})
				}
			}
			for _, rr := range s.set.Authority {
				if namesEqual(rr.Name, name) {
					ours = append(ours, tiebreakRecord{class: uint16(rr.Class) &^ 0x8000, rrtype: uint16(rr.Type), rdata: rr.Data})
				}
			}
			ours, theirs = dropDependentTXT(ours, theirs)
			if len(theirs) > 0 && compareTiebreak(ours, theirs) < 0 {
				s.send(state.ProbeEvent{Name: name, LostTiebreak: true})
			}
		}
	}
}

// dropDependentTXT removes TXT records from both sets when we probe an SRV
// for the name. A DNS-SD service's TXT record depends on its SRV record for
// conflict detection (Apple mDNSResponder marks it DependentOn the SRV), so
// the SRV alone decides the tiebreak. Without this, a peer whose probe
// carries only the SRV (as Apple's BCT does) makes our TXT (type 16) sort
// before its SRV (type 33) and we wrongly "lose" (BCT II.5).
func dropDependentTXT(ours, theirs []tiebreakRecord) ([]tiebreakRecord, []tiebreakRecord) {
	hasSRV := false
	for _, r := range ours {
		if r.rrtype == uint16(protocol.RecordTypeSRV) {
			hasSRV = true
			break
		}
	}
	if !hasSRV {
		return ours, theirs
	}
	keep := func(rs []tiebreakRecord) []tiebreakRecord {
		out := rs[:0:0]
		for _, r := range rs {
			if r.rrtype != uint16(protocol.RecordTypeTXT) {
				out = append(out, r)
			}
		}
		return out
	}
	return keep(ours), keep(theirs)
}

type tiebreakRecord struct {
	class, rrtype uint16
	rdata         []byte
}

func (a tiebreakRecord) compare(b tiebreakRecord) int {
	switch {
	case a.class != b.class:
		return int(a.class) - int(b.class)
	case a.rrtype != b.rrtype:
		return int(a.rrtype) - int(b.rrtype)
	default:
		return bytes.Compare(a.rdata, b.rdata)
	}
}

// compareTiebreak compares two proposed record sets per RFC 6762 §8.2.1:
// each set is sorted, then records are compared pairwise (class, type,
// rdata); the first difference decides, and if one set runs out first the
// set with records left is later. It returns <0 when ours is earlier (we
// lose), 0 when identical, >0 when ours is later (we win).
func compareTiebreak(ours, theirs []tiebreakRecord) int {
	sortRecords := func(rs []tiebreakRecord) {
		sort.Slice(rs, func(i, j int) bool { return rs[i].compare(rs[j]) < 0 })
	}
	sortRecords(ours)
	sortRecords(theirs)
	for i := 0; i < len(ours) && i < len(theirs); i++ {
		if c := ours[i].compare(theirs[i]); c != 0 {
			return c
		}
	}
	return len(ours) - len(theirs)
}
