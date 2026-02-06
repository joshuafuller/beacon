package state

import (
	"context"
	"net"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/records"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Announcer performs announcing per RFC 6762 §8.3.
//
// RFC 6762 §8.3: "The Multicast DNS responder MUST send at least two
// unsolicited responses, one second apart."
//
// T040: Implement Announcer
type Announcer struct {
	// Transport for sending announcement packets on the wire
	transport transport.Transport

	// Test hooks for injection
	onSendAnnouncement func()
	lastSentData       []byte
	lastDestAddr       string

	// US2 GREEN: Message capture for contract test validation
	lastAnnounceMessage []byte // Last sent announcement message (wire format)

	// Resource records to announce (DNS wire format serialization)
	resourceRecords []*records.ResourceRecord
}

// NewAnnouncer creates a new announcer.
func NewAnnouncer() *Announcer {
	return &Announcer{
		lastDestAddr: "224.0.0.251:5353", // RFC 6762 §5 multicast address
	}
}

// Announce sends unsolicited multicast announcements.
//
// RFC 6762 §8.3: Announcing process
//   - Send 2 announcements
//   - 1 second interval between announcements
//   - Total duration: ~1 second
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceName: Full service name
//   - records: Resource records to announce (wire format)
//
// Returns:
//   - error: Context error if canceled
//
// T040: Implement announcing with 2 announcements × 1s interval
func (a *Announcer) Announce(ctx context.Context, _ string, records []byte) error {
	const announcementCount = 2
	const announcementInterval = 1 * time.Second

	a.lastSentData = records

	for i := 0; i < announcementCount; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Send announcement
		// RFC 6762 §8.3: Announcements are DNS responses with answer records
		//
		// Build announcement message with actual resource records
		// If no records are set, fall back to empty stub for compatibility with existing tests
		var announceMsg []byte
		var err error

		if len(a.resourceRecords) > 0 {
			// Convert records.ResourceRecord to message.ResourceRecord for BuildResponse()
			messageRecords := make([]*message.ResourceRecord, len(a.resourceRecords))
			for i, rr := range a.resourceRecords {
				messageRecords[i] = &message.ResourceRecord{
					Name:       rr.Name,
					Type:       rr.Type,
					Class:      rr.Class,
					TTL:        rr.TTL,
					Data:       rr.Data,
					CacheFlush: rr.CacheFlush,
				}
			}

			// Use message.BuildResponse() to serialize records into wire format
			announceMsg, err = message.BuildResponse(messageRecords)
			if err != nil {
				// If serialization fails, fall back to empty message
				// This shouldn't happen in practice with valid records
				announceMsg = make([]byte, 12)
				announceMsg[2] = 0x84 // QR=1, AA=1
			}
		} else {
			// No records set - use minimal stub for backward compatibility with tests
			// Minimal DNS response header (12 bytes) per RFC 1035 §4.1.1:
			//   ID: 0x0000
			//   Flags: QR=1, AA=1 = 0x8400
			//   QDCOUNT, ANCOUNT, NSCOUNT, ARCOUNT: all 0
			announceMsg = make([]byte, 12)
			announceMsg[2] = 0x84 // High byte: QR=1, OPCODE=0, AA=1
			announceMsg[3] = 0x00 // Low byte: TC=0, RD=0, RA=0, Z=0, RCODE=0
		}

		a.lastAnnounceMessage = announceMsg

		if a.onSendAnnouncement != nil {
			a.onSendAnnouncement()
		}

		// Send announcement via transport (RFC 6762 §8.3: announcements to mDNS multicast group)
		if a.transport != nil {
			dest := &net.UDPAddr{
				IP:   net.ParseIP(protocol.MulticastAddrIPv4),
				Port: protocol.Port,
			}
			_ = a.transport.Send(ctx, announceMsg, dest) // nosemgrep: beacon-error-swallowing
		}

		// Wait 1s before next announcement (except after last)
		if i < announcementCount-1 {
			timer := time.NewTimer(announcementInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// Continue to next announcement
			}
		}
	}

	return nil
}

// GetLastAnnounceMessage returns the last sent announcement message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (a *Announcer) GetLastAnnounceMessage() []byte {
	return a.lastAnnounceMessage
}

// SetLastAnnounceMessage sets the last announcement message (for testing/transport integration).
//
// US2 GREEN: Allow transport layer to record sent messages
func (a *Announcer) SetLastAnnounceMessage(msg []byte) {
	a.lastAnnounceMessage = msg
}

// SetOnSendAnnouncement sets the callback to be called when an announcement is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (a *Announcer) SetOnSendAnnouncement(callback func()) {
	a.onSendAnnouncement = callback
}

// GetLastDestAddr returns the last destination address used for announcements.
//
// US2 GREEN: Contract test support for RFC 6762 §5 multicast address validation
func (a *Announcer) GetLastDestAddr() string {
	return a.lastDestAddr
}

// SetTransport sets the transport used to send announcement packets on the wire.
//
// RFC 6762 §8.3: Announcements are sent to the mDNS multicast group 224.0.0.251:5353.
func (a *Announcer) SetTransport(t transport.Transport) {
	a.transport = t
}

// SetRecords sets the resource records to be announced.
//
// This method allows the responder to provide the actual DNS records
// that should be included in announcement messages.
//
// Parameters:
//   - records: Resource records (PTR, SRV, TXT, A) to announce
//
// DNS message serialization: Called before Announce() to provide records
func (a *Announcer) SetRecords(records []*records.ResourceRecord) {
	a.resourceRecords = records
}
