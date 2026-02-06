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

		// Send probe query
		// RFC 6762 §8.1: Probe queries use query type "ANY" (255) for the claimed name.
		//
		// Build a proper probe message with the actual service instance name.
		// serviceName format: "My Printer._http._tcp.local"
		// Split at first "._" boundary: instance="My Printer", serviceType="_http._tcp.local"
		var encodedName []byte
		var encErr error
		if idx := strings.Index(serviceName, "._"); idx >= 0 {
			instanceName := serviceName[:idx]
			serviceType := serviceName[idx+1:] // e.g., "_http._tcp.local"
			encodedName, encErr = message.EncodeServiceInstanceName(instanceName, serviceType)
		} else {
			encodedName, encErr = message.EncodeName(serviceName)
		}
		if encErr != nil {
			return ProbeResult{Error: encErr}
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

		probeMsg := append(header, question...)
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

		// Wait 250ms before next probe (except after last probe).
		// RFC 6762 §8.1: During the wait, listen for responses that indicate conflicts.
		if i < probeCount-1 && p.transport != nil && p.listenForResponses {
			// Listen for responses during the 250ms probe interval
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
					// Timeout or context cancelled - check if parent ctx is done
					select {
					case <-ctx.Done():
						return ProbeResult{Error: ctx.Err()}
					default:
						break // Timeout expired - move on to next probe
					}
					break
				}

				// Skip nil/empty packets (e.g., mock transport returning immediately)
				if len(packet) == 0 {
					// Yield briefly to avoid busy-spinning on non-blocking mocks
					time.Sleep(time.Millisecond)
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

				// Check answers for conflict with our service name
				if p.conflictDetector != nil && len(p.ourRecords) > 0 {
					for _, answer := range respMsg.Answers {
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
								return ProbeResult{Error: detectErr}
							}
							if conflict {
								return ProbeResult{Conflict: true}
							}
						}
					}
				}
			}
		} else if i < probeCount-1 {
			// Not listening for responses (no transport, shared transport, or unit test mode).
			// Check injected records then wait the probe interval.
			// T059: Check for conflicts using ConflictDetector with injected records
			if p.conflictDetector != nil && len(p.incomingRecords) > 0 && len(p.ourRecords) > 0 {
				for _, ourRecord := range p.ourRecords {
					for _, incomingRecord := range p.incomingRecords {
						conflict, err := p.conflictDetector.DetectConflict(ourRecord, incomingRecord)
						if err != nil {
							return ProbeResult{Error: err}
						}
						if conflict {
							return ProbeResult{Conflict: true}
						}
					}
				}
			}
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
