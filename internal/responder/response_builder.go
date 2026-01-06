// Package responder implements mDNS response building per RFC 6762 §6.
package responder

import (
	"fmt"
	"strings"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/records"
)

// ResponseBuilder constructs mDNS responses per RFC 6762 §6.
//
// RFC 6762 §6: "When a Multicast DNS responder constructs and sends a Multicast DNS
// response message, the Resource Record Sections of that message must contain only
// records for which that responder is explicitly authoritative."
//
// Response structure per RFC 6762 §6:
//   - Answer section: Records that directly answer the query
//   - Additional section: Records that reduce round-trips (SRV, TXT, A for PTR queries)
//
// R005 Decision: Greedy packing with priority ordering (answer > additional),
// graceful truncation at 9000 bytes per RFC 6762 §17.
//
// T075: Implement ResponseBuilder struct
type ResponseBuilder struct {
	maxPacketSize int // RFC 6762 §17: 9000 bytes maximum
}

// ServiceWithIP extends Service with IP address for testing.
//
// T075: Service type for ResponseBuilder tests
type ServiceWithIP struct {
	InstanceName string
	ServiceType  string
	Domain       string
	Port         int
	IPv4Address  []byte
	TXTRecords   map[string]string
	Hostname     string
}

// NewResponseBuilder creates a new ResponseBuilder with RFC 6762 defaults.
//
// RFC 6762 §17: Maximum packet size is 9000 bytes.
//
// T075: ResponseBuilder constructor
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		maxPacketSize: 9000, // RFC 6762 §17
	}
}

// BuildResponse constructs an mDNS response for a query per RFC 6762 §6.
//
// RFC 6762 §6: For a PTR query, the response MUST contain:
//   - Answer section: PTR record pointing to service instance
//   - Additional section: SRV, TXT, A records (reduces round-trips)
//
// R005 Decision: Greedy packing - add all answer records (critical), then
// add additional records until 9000-byte limit reached.
//
// Returns:
//   - *message.DNSMessage: Response message
//   - error: If response construction fails
//
// T076: Implement BuildResponse()
func (rb *ResponseBuilder) BuildResponse(service *ServiceWithIP, query *message.DNSMessage) (*message.DNSMessage, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	// Build response header per RFC 6762 §6
	// Flags: QR=1 (response), OPCODE=0, AA=1 (authoritative), TC=0, RD=0, RA=0, Z=0, RCODE=0
	// Bit 15 (QR=1): 0x8000
	// Bit 10 (AA=1): 0x0400
	// Total: 0x8400
	flags := uint16(0x8400) // QR=1, AA=1

	response := &message.DNSMessage{
		Header: message.DNSHeader{
			ID:      query.Header.ID, // Match query ID
			Flags:   flags,           // Response with authoritative answer
			QDCount: 0,               // No questions in response per RFC 6762 §6
			ANCount: 0,               // Will be set after building
			NSCount: 0,               // No authority records
			ARCount: 0,               // Will be set after building
		},
		Questions:   []message.Question{}, // RFC 6762 §6: No questions in response
		Answers:     []message.Answer{},   // Will populate based on query
		Authorities: []message.Answer{},   // Empty for mDNS
		Additionals: []message.Answer{},   // Will populate with SRV, TXT, A
	}

	// Convert Service to records.ServiceInfo for record building
	serviceInfo := &records.ServiceInfo{
		InstanceName: service.InstanceName,
		ServiceType:  service.ServiceType,
		Hostname:     rb.getHostname(service),
		Port:         service.Port,
		IPv4Address:  service.IPv4Address,
		TXTRecords:   service.TXTRecords,
	}

	// Build all records for this service
	allRecords := records.BuildRecordSet(serviceInfo)

	// T095: Convert query known-answers (Answer section) to ResourceRecords for suppression
	knownAnswers := make([]*message.ResourceRecord, 0, len(query.Answers))
	for _, answer := range query.Answers {
		// Convert message.Answer to message.ResourceRecord for ApplyKnownAnswerSuppression
		knownAnswers = append(knownAnswers, &message.ResourceRecord{
			Name:       answer.NAME,
			Type:       protocol.RecordType(answer.TYPE),
			Class:      protocol.DNSClass(answer.CLASS),
			TTL:        answer.TTL,
			Data:       answer.RDATA,
			CacheFlush: (answer.CLASS & 0x8000) != 0,
		})
	}

	// For PTR query, answer is PTR record, additional is SRV/TXT/A
	// For now, assume first question is PTR query (will enhance later)
	if len(query.Questions) > 0 {
		question := query.Questions[0]

		// Check if this is a PTR query
		if question.QTYPE == uint16(protocol.RecordTypePTR) {
			// Add PTR record to answer section (with known-answer suppression)
			for _, rr := range allRecords {
				if rr.Type == protocol.RecordTypePTR {
					// T095: Apply known-answer suppression per RFC 6762 §7.1
					if rb.ApplyKnownAnswerSuppression(rr, knownAnswers) {
						response.Answers = append(response.Answers, rb.recordToAnswer(rr))
					}
					// T096: TODO - log suppressed record
					break
				}
			}

			// Add SRV, TXT, A to additional section (with known-answer suppression)
			for _, rr := range allRecords {
				if rr.Type == protocol.RecordTypeSRV || rr.Type == protocol.RecordTypeTXT || rr.Type == protocol.RecordTypeA {
					// T095: Apply known-answer suppression per RFC 6762 §7.1
					if rb.ApplyKnownAnswerSuppression(rr, knownAnswers) {
						response.Additionals = append(response.Additionals, rb.recordToAnswer(rr))
					}
					// T096: TODO - log suppressed record
				}
			}
		}
	}

	// Update counts
	response.Header.ANCount = uint16(len(response.Answers))
	response.Header.ARCount = uint16(len(response.Additionals))

	// Check packet size limit (RFC 6762 §17: 9000 bytes)
	estimatedSize := rb.EstimatePacketSize(response)
	if estimatedSize > rb.maxPacketSize {
		// R005: Gracefully truncate additional records
		originalAdditionalCount := len(response.Additionals)
		response.Additionals = rb.truncateAdditionals(response, estimatedSize)
		response.Header.ARCount = uint16(len(response.Additionals))

		// RFC 6762 §6.5: Set TC bit when truncated
		// Bit 9 (TC=1): 0x0200
		if len(response.Additionals) < originalAdditionalCount {
			response.Header.Flags |= 0x0200 // Set TC bit
		}
	}

	return response, nil
}

// EstimatePacketSize estimates the wire format size of a DNS message.
//
// RFC 6762 §17: Maximum packet size is 9000 bytes.
//
// Estimation formula (R005 decision):
//   - Header: 12 bytes
//   - Each record: ~60 bytes average (name + type + class + TTL + rdlength + rdata)
//
// T077: Implement EstimatePacketSize()
func (rb *ResponseBuilder) EstimatePacketSize(msg *message.DNSMessage) int {
	// Header is always 12 bytes
	size := 12

	// Estimate answer records
	for _, answer := range msg.Answers {
		size += rb.estimateRecordSize(&answer)
	}

	// Estimate additional records
	for _, additional := range msg.Additionals {
		size += rb.estimateRecordSize(&additional)
	}

	return size
}

// estimateRecordSize estimates the size of a single resource record.
//
// R005 decision: Conservative estimate
//   - Name: ~50 bytes (with compression)
//   - Type: 2 bytes
//   - Class: 2 bytes
//   - TTL: 4 bytes
//   - RDLength: 2 bytes
//   - RData: actual data length
//
// T077: Helper for packet size estimation
func (rb *ResponseBuilder) estimateRecordSize(answer *message.Answer) int {
	// Name (compressed): ~50 bytes average
	// Type (2) + Class (2) + TTL (4) + RDLength (2) = 10 bytes
	// RDATA: len(answer.RDATA)
	return 50 + 10 + len(answer.RDATA)
}

// truncateAdditionals removes additional records until packet size is acceptable.
//
// R005 Decision: Graceful truncation - keep answer section intact (critical),
// remove additional records (nice-to-have) until under 9000 bytes.
//
// T077: Implement truncation
func (rb *ResponseBuilder) truncateAdditionals(msg *message.DNSMessage, currentSize int) []message.Answer {
	// Remove additional records one by one until under limit
	additionals := make([]message.Answer, 0, len(msg.Additionals))
	size := currentSize

	for _, additional := range msg.Additionals {
		recordSize := rb.estimateRecordSize(&additional)
		if size-recordSize >= rb.maxPacketSize {
			// Skip this record
			size -= recordSize
			continue
		}
		additionals = append(additionals, additional)
	}

	return additionals
}

// recordToAnswer converts a ResourceRecord to an Answer.
//
// T076: Helper for response building
func (rb *ResponseBuilder) recordToAnswer(rr *message.ResourceRecord) message.Answer {
	return message.Answer{
		NAME:     rr.Name,
		TYPE:     uint16(rr.Type),
		CLASS:    uint16(rr.Class),
		TTL:      rr.TTL,
		RDLENGTH: uint16(len(rr.Data)),
		RDATA:    rr.Data,
	}
}

// getHostname returns the hostname for the service.
//
// If service.Hostname is set, use it. Otherwise, construct from instance name.
//
// T076: Helper
func (rb *ResponseBuilder) getHostname(service *ServiceWithIP) string {
	if service.Hostname != "" {
		return service.Hostname
	}
	// Default: instancename.local
	return service.InstanceName + ".local"
}

// ApplyKnownAnswerSuppression determines if a record should be included in the response
// based on known-answer suppression per RFC 6762 §7.1.
//
// RFC 6762 §7.1: "A Multicast DNS responder MUST NOT answer a Multicast DNS query if
// the answer it would give is already included in the Answer Section with an RR TTL
// at least half the correct value."
//
// Parameters:
//   - ourRecord: The record we would send in the response
//   - knownAnswers: Records from the query's Answer Section (known-answer list)
//
// Returns:
//   - true: Include the record in response (no suppression)
//   - false: Suppress the record (already in known-answer list with TTL ≥50%)
//
// T092: Implement known-answer suppression logic
func (rb *ResponseBuilder) ApplyKnownAnswerSuppression(ourRecord *message.ResourceRecord, knownAnswers []*message.ResourceRecord) bool {
	// No known-answers → no suppression
	if len(knownAnswers) == 0 {
		return true // Include in response
	}

	// Check if ourRecord matches any known-answer
	for _, knownAnswer := range knownAnswers {
		// RFC 6762 §7.1: Records must match on Name, Type, Class, and RDATA
		if !recordsMatch(ourRecord, knownAnswer) {
			continue // Not a match, check next known-answer
		}

		// Records match - check TTL threshold
		// RFC 6762 §7.1: Suppress if known-answer TTL ≥ 50% of true TTL
		ttlThreshold := ourRecord.TTL / 2 // 50% of true TTL

		if knownAnswer.TTL >= ttlThreshold {
			// Known-answer TTL ≥50% → suppress (querier's cache is fresh enough)
			return false // Do NOT include in response
		}

		// Known-answer TTL <50% → respond to refresh before expiration
		return true // Include in response
	}

	// No matching known-answer found → include in response
	return true
}

// recordsMatch checks if two resource records match per RFC 6762 §7.1 criteria.
//
// RFC 6762 §7.1: Records match if Name, Type, Class, and RDATA are identical.
//
// Parameters:
//   - a, b: Resource records to compare
//
// Returns:
//   - true: Records match (same Name, Type, Class, RDATA)
//   - false: Records differ
//
// T092: Helper for known-answer matching
func recordsMatch(a, b *message.ResourceRecord) bool {
	// Name comparison (case-insensitive per RFC 1035 §2.3.3)
	// DNS names are case-insensitive: "example.local" == "EXAMPLE.LOCAL"
	if !strings.EqualFold(a.Name, b.Name) {
		return false
	}

	// Type must match
	if a.Type != b.Type {
		return false
	}

	// Class must match (ignore cache-flush bit for comparison)
	// RFC 6762 §10.2: Cache-flush bit is NOT part of record identity
	classA := a.Class & 0x7FFF // Mask out cache-flush bit
	classB := b.Class & 0x7FFF
	if classA != classB {
		return false
	}

	// RDATA must match byte-for-byte
	if len(a.Data) != len(b.Data) {
		return false
	}
	for i := range a.Data {
		if a.Data[i] != b.Data[i] {
			return false
		}
	}

	// All criteria match
	return true
}
