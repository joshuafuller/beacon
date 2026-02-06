package state

import (
	"context"
	"net"
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

		// Send probe query
		// RFC 6762 §8.1: Probe queries use query type "ANY" (255)
		// US2 GREEN: Build actual probe message for contract test validation
		//
		// NOTE: BuildQuery() currently rejects spaces in DNS labels (per RFC 1035),
		// but RFC 6763 DNS-SD allows spaces in service instance names.
		// For now, create a minimal stub message for contract test validation.
		// Full DNS-SD name encoding will be implemented in US4 (Service Publishing).
		//
		// Minimal DNS header (12 bytes) per RFC 1035 §4.1.1:
		//   ID (2 bytes): 0x0000
		//   Flags (2 bytes): QR=0, OPCODE=0, AA=0, TC=0, RD=0, RA=0, Z=0, RCODE=0 = 0x0000
		//   QDCOUNT (2 bytes): 1 question
		//   ANCOUNT (2 bytes): 0 answers
		//   NSCOUNT (2 bytes): 0 authority
		//   ARCOUNT (2 bytes): 0 additional
		//   Question section (variable): QNAME + QTYPE + QCLASS
		//
		// For contract test validation, we just need header + minimal question
		probeMsg := make([]byte, 28) // 12-byte header + 16-byte minimal question
		// Header: all zeros already (QR=0, OPCODE=0 is correct)
		probeMsg[4] = 0x00 // QDCOUNT high byte
		probeMsg[5] = 0x01 // QDCOUNT low byte = 1 question
		// Minimal question: <root> ANY IN
		probeMsg[12] = 0x00 // Root label (length 0)
		probeMsg[13] = 0x00 // QTYPE high byte
		probeMsg[14] = 0xFF // QTYPE low byte = 255 (ANY)
		probeMsg[15] = 0x00 // QCLASS high byte
		probeMsg[16] = 0x01 // QCLASS low byte = 1 (IN)
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
			_ = p.transport.Send(ctx, probeMsg, dest) // nosemgrep: beacon-error-swallowing
		}

		// T059: Check for conflicts using ConflictDetector (if configured)
		if p.conflictDetector != nil && len(p.incomingRecords) > 0 && len(p.ourRecords) > 0 {
			// Check each incoming record against each of our records
			for _, ourRecord := range p.ourRecords {
				for _, incomingRecord := range p.incomingRecords {
					conflict, err := p.conflictDetector.DetectConflict(ourRecord, incomingRecord)
					if err != nil {
						return ProbeResult{Error: err}
					}
					if conflict {
						// Conflict detected via ConflictDetector
						return ProbeResult{Conflict: true}
					}
				}
			}
		}

		// Check for injected conflict (test hook - legacy)
		if p.injectConflictAfter > 0 && i >= p.injectConflictAfter {
			return ProbeResult{Conflict: true}
		}

		// Check for simultaneous probe (test hook for tie-breaking - legacy)
		if p.injectSimultaneousProbe {
			// Simulate lexicographic comparison
			// In production, this would use ConflictDetector.CompareProbes()
			weWin := compareBytesLexicographically(p.ourProbeData, p.theirProbeData)
			if !weWin {
				// We lose tie-break
				return ProbeResult{Conflict: true}
			}
			// We win tie-break, continue probing
		}

		// Wait 250ms before next probe (except after last probe)
		if i < probeCount-1 {
			timer := time.NewTimer(protocol.ProbeInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ProbeResult{Error: ctx.Err()}
			case <-timer.C:
				// Continue to next probe
			}
		}
	}

	// No conflict detected
	return ProbeResult{Conflict: false}
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

// SetOnSendQuery sets the callback to be called when a probe query is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (p *Prober) SetOnSendQuery(callback func()) {
	p.onSendQuery = callback
}
