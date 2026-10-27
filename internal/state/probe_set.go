package state

import (
	"context"
	"encoding/binary"
	"math/rand/v2"
	"net"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// ProbeSet describes one RFC 6762 §8.1 probe: the names asked for (one
// question each, type ANY, QU bit set) and the records this host proposes to
// use, sent in the Authority section for simultaneous-probe tiebreaking
// (§8.2).
type ProbeSet struct {
	Names     []string
	Authority []*message.ResourceRecord
}

// ProbeEvent is something the responder saw on the network while probing.
type ProbeEvent struct {
	// Name is the probed name the event is about.
	Name string
	// LostTiebreak is set when another host probes for the same name with
	// lexicographically later records (§8.2): defer one second and probe
	// again with the same name. Otherwise the event is a conflicting answer
	// (§8.1): choose a new name.
	LostTiebreak bool
}

// SetProbeSet makes the prober send multi-question probes with an Authority
// section and listen for events instead of reading the transport itself.
func (p *Prober) SetProbeSet(ps ProbeSet, events <-chan ProbeEvent) {
	p.probeSet = &ps
	p.events = events
}

// LastResult returns the result of the most recent Probe call.
func (p *Prober) LastResult() ProbeResult {
	return p.lastResult
}

// probeWithSet runs RFC 6762 §8.1 probing for p.probeSet: a random 0-250 ms
// wait, three probes 250 ms apart, then 250 ms more. Any event that arrives
// after the first probe ends probing; events from before it are stale and
// ignored (§8.1).
func (p *Prober) probeWithSet(ctx context.Context) ProbeResult {
	probeMsg, err := buildProbeSetMessage(*p.probeSet)
	if err != nil {
		return ProbeResult{Error: err}
	}

	initial := rand.N(protocol.ProbeInterval + 1) // #nosec G404 -- timing jitter, not security
	if err := sleepCtx(ctx, initial); err != nil {
		return ProbeResult{Error: err}
	}
	p.drainEvents()

	const probeCount = 3
	for i := 0; i < probeCount; i++ {
		p.lastProbeMessage = probeMsg
		if p.onSendQuery != nil {
			p.onSendQuery()
		}
		if p.transport != nil {
			dest := &net.UDPAddr{IP: net.ParseIP(protocol.MulticastAddrIPv4), Port: protocol.Port}
			if err := p.transport.Send(ctx, probeMsg, dest); err != nil {
				return ProbeResult{Error: err}
			}
		}
		if result, done := p.waitForEvent(ctx, protocol.ProbeInterval); done {
			return result
		}
	}
	return ProbeResult{}
}

// waitForEvent waits d for a probe event.
func (p *Prober) waitForEvent(ctx context.Context, d time.Duration) (ProbeResult, bool) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ProbeResult{Error: ctx.Err()}, true
	case ev := <-p.events:
		if ev.LostTiebreak {
			return ProbeResult{Conflict: true, ConflictName: ev.Name, Deferred: true}, true
		}
		return ProbeResult{Conflict: true, ConflictName: ev.Name}, true
	case <-timer.C:
		return ProbeResult{}, false
	}
}

// drainEvents drops events received before the first probe (stale, §8.1).
func (p *Prober) drainEvents() {
	for {
		select {
		case <-p.events:
		default:
			return
		}
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// buildProbeSetMessage builds the probe query: one QU question of type ANY per
// name, and the proposed records in the Authority section without the
// cache-flush bit (RFC 6762 §8.1, §8.2; BCT fails probes whose proposed
// answers set it).
func buildProbeSetMessage(ps ProbeSet) ([]byte, error) {
	pkt := make([]byte, 12)
	// #nosec G115 -- a probe holds a handful of names and records.
	binary.BigEndian.PutUint16(pkt[4:6], uint16(len(ps.Names))) //nolint:gosec
	// #nosec G115 -- as above.
	binary.BigEndian.PutUint16(pkt[8:10], uint16(len(ps.Authority))) //nolint:gosec
	for _, name := range ps.Names {
		enc, err := message.EncodeDomainName(name)
		if err != nil {
			return nil, err
		}
		pkt = append(pkt, enc...)
		pkt = binary.BigEndian.AppendUint16(pkt, uint16(protocol.RecordTypeANY))
		pkt = binary.BigEndian.AppendUint16(pkt, uint16(protocol.ClassIN)|0x8000) // QU
	}
	for _, rr := range ps.Authority {
		proposed := *rr
		proposed.CacheFlush = false
		enc, err := message.SerializeResourceRecord(&proposed)
		if err != nil {
			return nil, err
		}
		pkt = append(pkt, enc...)
	}
	return pkt, nil
}
