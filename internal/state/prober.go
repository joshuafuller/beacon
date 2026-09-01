package state

import (
	"context"
	"encoding/binary"
	"net"
	"strings"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
)

// ProbeResult represents the result of probing.
type ProbeResult struct {
	Conflict bool  // true if naming conflict detected
	Error    error // error if probing failed
}

// Prober performs probing per RFC 6762 §8.1.
//
// RFC 6762 §8.1: "Before claiming a unique record, a host MUST send at least
// two probe queries, 250 milliseconds apart."
//
// Beacon implementation: Send exactly 3 probes for robust conflict detection.
//
// T039: Implement Prober
// T059: Integrate ConflictDetector with Prober (GREEN phase)
type Prober struct {
	// Transport for sending probe packets on the wire
	transport transport.Transport

	// listenForResponses enables the prober to call transport.Receive() during
	// probe intervals. When false (default), the prober only sends probes and
	// relies on an external receive loop (e.g., Responder's query handler) to
	// feed conflict information. Enable this when the prober has a dedicated
	// transport that is not shared with other goroutines.
	listenForResponses bool

	// Test hooks for injection
	onSendQuery             func()
	injectConflictAfter     int
	injectSimultaneousProbe bool
	ourProbeData            []byte
	theirProbeData          []byte

	// T059: ConflictDetector integration
	ourRecords       []message.ResourceRecord  // Our records being probed
	incomingRecords  []message.ResourceRecord  // Incoming probe responses (test hook)
	conflictDetector ConflictDetectorInterface // For detecting conflicts

	// US2 GREEN: Message capture for contract test validation
	lastProbeMessage []byte // Last sent probe message (wire format)
}

// ConflictDetectorInterface defines the interface for conflict detection.
// This allows us to use the ConflictDetector from responder package.
//
// T059: Interface for ConflictDetector integration
type ConflictDetectorInterface interface {
	DetectConflict(ourRecord, incomingRecord message.ResourceRecord) (bool, error)
}

// NewProber creates a new prober.
func NewProber() *Prober {
	return &Prober{}
}

// Probe sends probe queries to detect naming conflicts.
//
// RFC 6762 §8.1: Probing process
//   - Send 3 probe queries
//   - 250ms intervals between probes
//   - Total duration: ~750ms
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceName: Full service name (e.g., "My Printer._http._tcp.local")
//
// Returns:
//   - ProbeResult: Result with Conflict flag and any error
//
// T039: Implement probing with 3 queries × 250ms intervals
func (p *Prober) Probe(ctx context.Context, serviceName string) ProbeResult {
	const probeCount = 3

	for i := 0; i < probeCount; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ProbeResult{Error: ctx.Err()}
		default:
		}

		if err := p.sendProbe(ctx, serviceName); err != nil {
			return ProbeResult{Error: err}
		}

		if result, done := p.checkLegacyConflictInjection(i); done {
			return result
		}

		// Wait 250ms before next probe (except after last probe).
		// RFC 6762 §8.1: During the wait, listen for responses that indicate conflicts.
		if i < probeCount-1 {
			if p.transport != nil && p.listenForResponses {
				if result, done := p.listenForConflictDuringInterval(ctx); done {
					return result
				}
			} else {
				if result, done := p.waitProbeIntervalCheckingInjectedConflict(ctx); done {
					return result
				}
			}
		}
	}

	// No conflict detected
	return ProbeResult{Conflict: false}
}

// buildProbeMessage encodes serviceName and assembles the wire-format probe
// query: a DNS header (QDCOUNT=1) plus a question section for QTYPE=ANY (255),
// QCLASS=IN, per RFC 6762 §8.1.
//
// serviceName format: "My Printer._http._tcp.local" — split at the first
// "._" boundary into instance="My Printer", serviceType="_http._tcp.local"
// so the instance label is encoded per RFC 6763 §4.3 (allows spaces/UTF-8).
func buildProbeMessage(serviceName string) ([]byte, error) {
	var encodedName []byte
	var err error
	if idx := strings.Index(serviceName, "._"); idx >= 0 {
		instanceName := serviceName[:idx]
		serviceType := serviceName[idx+1:] // e.g., "_http._tcp.local"
		encodedName, err = message.EncodeServiceInstanceName(instanceName, serviceType)
	} else {
		encodedName, err = message.EncodeName(serviceName)
	}
	if err != nil {
		return nil, err
	}

	// Build DNS header (12 bytes) + question section (encodedName + 4 bytes QTYPE+QCLASS)
	// Header: QR=0, OPCODE=0, QDCOUNT=1, all else zero
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[4:6], 1) // QDCOUNT = 1

	// Question section: QNAME + QTYPE(ANY=255) + QCLASS(IN=1)
	question := make([]byte, len(encodedName)+4)
	copy(question, encodedName)
	binary.BigEndian.PutUint16(question[len(encodedName):], uint16(protocol.RecordTypeANY))
	binary.BigEndian.PutUint16(question[len(encodedName)+2:], uint16(protocol.ClassIN))

	return append(header, question...), nil
}

// sendProbe builds the probe message for serviceName, records it for
// contract-test capture, fires the onSendQuery test hook, and sends it to the
// mDNS multicast group per RFC 6762 §8.1.
func (p *Prober) sendProbe(ctx context.Context, serviceName string) error {
	probeMsg, err := buildProbeMessage(serviceName)
	if err != nil {
		return err
	}
	p.lastProbeMessage = probeMsg

	// Notify test hooks
	if p.onSendQuery != nil {
		p.onSendQuery()
	}

	// Send probe via transport (RFC 6762 §8.1: probes sent to mDNS multicast group)
	if p.transport != nil {
		dest := &net.UDPAddr{
			IP:   net.ParseIP(protocol.MulticastAddrIPv4),
			Port: protocol.Port,
		}
		if err := p.transport.Send(ctx, probeMsg, dest); err != nil {
			return err
		}
	}

	return nil
}

// checkLegacyConflictInjection evaluates the legacy injectConflictAfter and
// injectSimultaneousProbe test hooks for probe iteration probeIndex.
// done=true means Probe must return result immediately.
func (p *Prober) checkLegacyConflictInjection(probeIndex int) (result ProbeResult, done bool) {
	// Check for injected conflict (test hook - legacy)
	if p.injectConflictAfter > 0 && probeIndex >= p.injectConflictAfter {
		return ProbeResult{Conflict: true}, true
	}

	// Check for simultaneous probe (test hook for tie-breaking - legacy)
	if p.injectSimultaneousProbe {
		// Simulate lexicographic comparison
		// In production, this would use ConflictDetector.CompareProbes()
		weWin := compareBytesLexicographically(p.ourProbeData, p.theirProbeData)
		if !weWin {
			// We lose tie-break
			return ProbeResult{Conflict: true}, true
		}
		// We win tie-break, continue probing
	}

	return ProbeResult{}, false
}

// checkAnswersForConflict checks each answer against p.ourRecords via
// p.conflictDetector, per RFC 6762 §8.2 tie-breaking.
func (p *Prober) checkAnswersForConflict(answers []message.Answer) (conflict bool, err error) {
	if p.conflictDetector == nil || len(p.ourRecords) == 0 {
		return false, nil
	}

	for _, answer := range answers {
		// Convert Answer to ResourceRecord for conflict detection
		incoming := message.ResourceRecord{
			Name:  answer.NAME,
			Type:  protocol.RecordType(answer.TYPE),
			Class: protocol.DNSClass(answer.CLASS & 0x7FFF), // strip cache-flush bit
			TTL:   answer.TTL,
			Data:  answer.RDATA,
		}
		for _, ourRecord := range p.ourRecords {
			conflict, detectErr := p.conflictDetector.DetectConflict(ourRecord, incoming)
			if detectErr != nil {
				return false, detectErr
			}
			if conflict {
				return true, nil
			}
		}
	}

	return false, nil
}

// listenForConflictDuringInterval implements the RFC 6762 §8.1 250ms
// listen-for-response window between probes: it receives packets until the
// deadline, parses each response, and checks its answers for a conflict via
// checkAnswersForConflict. done=true means Probe must return result
// immediately.
func (p *Prober) listenForConflictDuringInterval(ctx context.Context) (result ProbeResult, done bool) {
	deadline := time.Now().Add(protocol.ProbeInterval)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		receiveCtx, cancelReceive := context.WithTimeout(ctx, remaining)
		packet, _, _, recvErr := p.transport.Receive(receiveCtx)
		cancelReceive()
		if recvErr != nil {
			// Timeout or context canceled - check if parent ctx is done
			select {
			case <-ctx.Done():
				return ProbeResult{Error: ctx.Err()}, true
			default:
				break // Timeout expired - move on to next probe
			}
			break
		}

		// Skip nil/empty packets (e.g., mock transport returning immediately)
		if len(packet) == 0 {
			// Yield briefly to avoid busy-spinning on non-blocking mocks
			timer := time.NewTimer(time.Millisecond)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return ProbeResult{Error: ctx.Err()}, true
			case <-timer.C:
			}
			continue
		}

		// Parse response and check for conflicts
		respMsg, parseErr := message.ParseMessage(packet)
		if parseErr != nil {
			continue // Malformed packet - ignore
		}

		// Only process responses (QR=1)
		if !respMsg.Header.IsResponse() {
			continue
		}

		conflict, err := p.checkAnswersForConflict(respMsg.Answers)
		if err != nil {
			return ProbeResult{Error: err}, true
		}
		if conflict {
			return ProbeResult{Conflict: true}, true
		}
	}

	return ProbeResult{}, false
}

// checkInjectedRecordsForConflict checks the test-injected p.incomingRecords
// against p.ourRecords via p.conflictDetector (legacy non-listening path).
//
// T059: Check for conflicts using ConflictDetector with injected records
func (p *Prober) checkInjectedRecordsForConflict() (conflict bool, err error) {
	if p.conflictDetector == nil || len(p.incomingRecords) == 0 || len(p.ourRecords) == 0 {
		return false, nil
	}

	for _, ourRecord := range p.ourRecords {
		for _, incomingRecord := range p.incomingRecords {
			conflict, detectErr := p.conflictDetector.DetectConflict(ourRecord, incomingRecord)
			if detectErr != nil {
				return false, detectErr
			}
			if conflict {
				return true, nil
			}
		}
	}

	return false, nil
}

// waitProbeIntervalCheckingInjectedConflict is the legacy non-listening path
// (no transport, a shared transport, or unit-test mode): it checks injected
// records for a conflict via checkInjectedRecordsForConflict, then blocks for
// the RFC 6762 §8.1 250ms probe interval (or until ctx is done). done=true
// means Probe must return result immediately.
func (p *Prober) waitProbeIntervalCheckingInjectedConflict(ctx context.Context) (result ProbeResult, done bool) {
	if conflict, err := p.checkInjectedRecordsForConflict(); err != nil {
		return ProbeResult{Error: err}, true
	} else if conflict {
		return ProbeResult{Conflict: true}, true
	}

	timer := time.NewTimer(protocol.ProbeInterval)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ProbeResult{Error: ctx.Err()}, true
	case <-timer.C:
		// Continue to next probe
	}

	return ProbeResult{}, false
}

// compareBytesLexicographically compares two byte slices lexicographically.
// Returns true if a > b (we win), false otherwise.
func compareBytesLexicographically(a, b []byte) bool {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] > b[i] {
			return true // We win
		} else if a[i] < b[i] {
			return false // They win
		}
	}

	// If all bytes match up to minLen, longer slice wins
	return len(a) > len(b)
}

// SetOurRecords sets the records we're probing for (test hook).
//
// T059: Test hook for ConflictDetector integration testing
func (p *Prober) SetOurRecords(records []message.ResourceRecord) {
	p.ourRecords = records
}

// InjectIncomingResponse injects incoming probe responses for testing.
//
// T059: Test hook for ConflictDetector integration testing
func (p *Prober) InjectIncomingResponse(records []message.ResourceRecord) {
	p.incomingRecords = records
}

// SetConflictDetector sets the conflict detector to use.
//
// T059: Allow injection of ConflictDetector for testing
func (p *Prober) SetConflictDetector(detector ConflictDetectorInterface) {
	p.conflictDetector = detector
}

// GetLastProbeMessage returns the last sent probe message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (p *Prober) GetLastProbeMessage() []byte {
	return p.lastProbeMessage
}

// SetLastProbeMessage sets the last probe message (for testing/transport integration).
//
// US2 GREEN: Allow transport layer to record sent messages
func (p *Prober) SetLastProbeMessage(msg []byte) {
	p.lastProbeMessage = msg
}

// SetTransport sets the transport used to send probe packets on the wire.
//
// RFC 6762 §8.1: Probes are sent to the mDNS multicast group 224.0.0.251:5353.
func (p *Prober) SetTransport(t transport.Transport) {
	p.transport = t
}

// EnableListenForResponses enables the prober to actively listen for responses
// by calling transport.Receive() during the 250ms probe intervals.
//
// This should only be enabled when the prober has a dedicated transport that is
// not shared with other goroutines (e.g., in standalone mode or testing).
// When a shared transport is in use (e.g., Responder's query handler), leave
// this disabled and rely on the external receive loop to feed conflict info.
//
// RFC 6762 §8.1: Listening for responses during probing is required for
// production conflict detection.
func (p *Prober) EnableListenForResponses() {
	p.listenForResponses = true
}

// SetOnSendQuery sets the callback to be called when a probe query is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (p *Prober) SetOnSendQuery(callback func()) {
	p.onSendQuery = callback
}
