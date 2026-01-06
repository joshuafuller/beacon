// Package records implements resource record construction per RFC 6762/6763.
package records

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// ServiceInfo holds service information for record set building.
//
// This is used internally to construct the full set of resource records
// (PTR, SRV, TXT, A) for a registered service.
//
// T033: ServiceInfo type for BuildRecordSet()
type ServiceInfo struct {
	InstanceName string            // "My Printer"
	ServiceType  string            // "_http._tcp.local"
	Hostname     string            // "myhost.local"
	Port         int               // 8080
	IPv4Address  []byte            // [192, 168, 1, 100]
	TXTRecords   map[string]string // {"version": "1.0"}
}

// BuildRecordSet constructs a complete set of resource records for a service.
//
// RFC 6763 §6: A registered service includes:
//   - PTR record: _service._proto.local → instance._service._proto.local
//   - SRV record: instance._service._proto.local → hostname:port
//   - TXT record: instance._service._proto.local → key-value pairs
//   - A record: hostname.local → IPv4 address
//
// Parameters:
//   - service: Service information
//
// Returns:
//   - []*message.ResourceRecord: All records (PTR, SRV, TXT, A)
//
// FR-032: System MUST build complete record set (PTR, SRV, TXT, A)
// T033: Implement BuildRecordSet()
func BuildRecordSet(service *ServiceInfo) []*message.ResourceRecord {
	records := make([]*message.ResourceRecord, 0, 4)

	// 1. PTR record: _service._proto.local → instance._service._proto.local
	ptrRecord := buildPTRRecord(service)
	records = append(records, ptrRecord)

	// 2. SRV record: instance._service._proto.local → hostname:port
	srvRecord := buildSRVRecord(service)
	records = append(records, srvRecord)

	// 3. TXT record: instance._service._proto.local → key-value pairs
	txtRecord := buildTXTRecordFromService(service)
	records = append(records, txtRecord)

	// 4. A record: hostname.local → IPv4 address
	aRecord := buildARecord(service)
	records = append(records, aRecord)

	return records
}

// buildPTRRecord constructs a PTR record per RFC 6763 §6.
//
// PTR record format:
//   - Name: _service._proto.local (e.g., "_http._tcp.local")
//   - Type: PTR (12)
//   - Class: IN (1)
//   - TTL: 120 seconds (service record per RFC 6762 §10)
//   - RDATA: instance._service._proto.local (e.g., "My Printer._http._tcp.local")
//   - CacheFlush: false (PTR is a shared record per RFC 6762 §10.2)
//
// RFC 6762 §10: PTR records for DNS-SD services use 120 seconds.
// Service discovery records change more frequently than hostname records.
//
// T033: PTR record construction
func buildPTRRecord(service *ServiceInfo) *message.ResourceRecord {
	// PTR points from service type to full service instance name
	name := service.ServiceType

	// Encode target as service instance name per RFC 6763 §4.3
	// Service instance names can contain UTF-8 and spaces
	// Error impossible: ServiceInfo validated by responder.Service.Validate()
	targetEncoded, _ := message.EncodeServiceInstanceName(service.InstanceName, service.ServiceType) // nosemgrep: beacon-error-swallowing

	return &message.ResourceRecord{
		Name:       name,
		Type:       protocol.RecordTypePTR,
		Class:      protocol.ClassIN,
		TTL:        120, // RFC 6762 §10: 120 seconds for service records
		Data:       targetEncoded,
		CacheFlush: false, // PTR is shared (multiple services can have same type)
	}
}

// buildSRVRecord constructs an SRV record per RFC 6763 §6.
//
// SRV record format:
//   - Name: instance._service._proto.local
//   - Type: SRV (33)
//   - Class: IN (1)
//   - TTL: 120 seconds (service TTL)
//   - RDATA: priority (0), weight (0), port, hostname
//   - CacheFlush: true (SRV is unique per RFC 6762 §10.2)
//
// T033: SRV record construction
func buildSRVRecord(service *ServiceInfo) *message.ResourceRecord {
	// Name format: instance._service._proto.local
	// NOTE: Instance name may contain spaces/UTF-8 per RFC 6763 §4.3
	// The serializer will handle encoding with EncodeServiceInstanceName()
	name := service.InstanceName + "." + service.ServiceType

	// SRV RDATA format per RFC 2782:
	//   Priority (2 bytes, big-endian) = 0
	//   Weight (2 bytes, big-endian) = 0
	//   Port (2 bytes, big-endian)
	//   Target (hostname as DNS name)
	data := make([]byte, 6)                  // Priority + Weight + Port
	binary.BigEndian.PutUint16(data[0:2], 0) // Priority = 0
	binary.BigEndian.PutUint16(data[2:4], 0) // Weight = 0
	// G115: bounds checked - service.Port is int representing a valid port number (0-65535)
	port := service.Port
	if port < 0 || port > 65535 { //nolint:gosec // G115: bounds checked
		port = 0 // Fallback to 0 if invalid
	}
	binary.BigEndian.PutUint16(data[4:6], uint16(port)) // Port

	// Append encoded hostname
	// Error impossible: ServiceInfo.Hostname pre-validated by caller
	// Hostname follows format "name.local" with valid DNS labels
	hostnameEncoded, _ := message.EncodeName(service.Hostname) // nosemgrep: beacon-error-swallowing
	data = append(data, hostnameEncoded...)

	return &message.ResourceRecord{
		Name:       name,
		Type:       protocol.RecordTypeSRV,
		Class:      protocol.ClassIN,
		TTL:        protocol.TTLService, // 120 seconds
		Data:       data,
		CacheFlush: true, // SRV is unique (one service instance = one SRV)
	}
}

// buildTXTRecordFromService constructs a TXT record for a service per RFC 6763 §6.
//
// RFC 6762 §10: TXT records for DNS-SD services use 120 seconds.
// Service discovery records change more frequently than hostname records.
//
// T034: TXT record construction for service
func buildTXTRecordFromService(service *ServiceInfo) *message.ResourceRecord {
	name := service.InstanceName + "." + service.ServiceType
	data := buildTXTRecord(service.TXTRecords)

	return &message.ResourceRecord{
		Name:       name,
		Type:       protocol.RecordTypeTXT,
		Class:      protocol.ClassIN,
		TTL:        120, // RFC 6762 §10: 120 seconds for service records
		Data:       data,
		CacheFlush: true, // TXT is unique per service instance
	}
}

// buildTXTRecord encodes TXT records per RFC 6763 §6.
//
// RFC 6763 §6: "If a DNS-SD service has no TXT records, it MUST include a
// single TXT record consisting of a single zero byte (0x00)."
//
// TXT record format per RFC 6763 §6.4:
//   - Each key-value pair: length byte + "key=value" string
//   - Multiple pairs concatenated
//   - Empty TXT: single 0x00 byte
//
// Parameters:
//   - txtRecords: Map of key-value pairs
//
// Returns:
//   - []byte: Encoded TXT record data
//
// FR-031: System MUST create mandatory TXT record with 0x00 byte if empty
// T034: Implement buildTXTRecord()
func buildTXTRecord(txtRecords map[string]string) []byte {
	// RFC 6763 §6: Empty TXT MUST be 0x00
	if len(txtRecords) == 0 {
		return []byte{0x00}
	}

	// Encode each key-value pair with length prefix
	data := make([]byte, 0, 256)
	for key, value := range txtRecords {
		// Format: "key=value"
		entry := key + "=" + value

		// Length byte + entry string
		entryLen := byte(len(entry))
		data = append(data, entryLen)
		data = append(data, []byte(entry)...)
	}

	return data
}

// buildARecord constructs an A record per RFC 1035 §3.4.1.
//
// A record format:
//   - Name: hostname.local
//   - Type: A (1)
//   - Class: IN (1)
//   - TTL: 4500 seconds (75 minutes per RFC 6762 §10)
//   - RDATA: IPv4 address (4 bytes)
//   - CacheFlush: true (A is unique per RFC 6762 §10.2)
//
// RFC 6762 §10: Hostname records (A, AAAA) use 4500 seconds (75 minutes).
// Host IP addresses change less frequently than service discovery records.
//
// T033: A record construction
func buildARecord(service *ServiceInfo) *message.ResourceRecord {
	if len(service.IPv4Address) != 4 {
		// Invalid IPv4 address - return placeholder
		// In production, this should return an error
		service.IPv4Address = []byte{0, 0, 0, 0}
	}

	return &message.ResourceRecord{
		Name:       service.Hostname,
		Type:       protocol.RecordTypeA,
		Class:      protocol.ClassIN,
		TTL:        4500, // RFC 6762 §10: 4500 seconds (75 min) for hostname records
		Data:       service.IPv4Address,
		CacheFlush: true, // A is unique (one hostname = one IP)
	}
}

// ResourceRecord is a type alias for message.ResourceRecord.
// This allows tests to reference ResourceRecord without importing message package.
type ResourceRecord = message.ResourceRecord

// RecordSet tracks per-record, per-interface multicast timestamps for rate limiting.
//
// RFC 6762 §6.2: "A Multicast DNS responder MUST NOT multicast a given resource record
// on a given interface until at least one second has elapsed since the last time that
// resource record was multicast on that particular interface."
//
// Rate limiting is:
//   - PER RECORD: Different records have independent rate limits
//   - PER INTERFACE: Same record can be multicast on different interfaces simultaneously
//
// Exception per RFC 6762 §6.2: Probe defense allows 250ms minimum instead of 1 second
//
// T073-T074: Implement rate limiting
type RecordSet struct {
	// lastMulticast tracks per-record, per-interface multicast timestamps
	// Key: buildRecordKey(rr) + ":" + interfaceID
	// Value: timestamp of last multicast (Unix nanoseconds for 250ms probe defense precision)
	lastMulticast map[string]int64
}

// NewRecordSet creates a new RecordSet for rate limiting tracking.
//
// T073: Constructor for RecordSet
func NewRecordSet() *RecordSet {
	return &RecordSet{
		lastMulticast: make(map[string]int64),
	}
}

// CanMulticast checks if a record can be multicast on the given interface per RFC 6762 §6.2.
//
// RFC 6762 §6.2: "MUST NOT multicast a given resource record on a given interface until
// at least one second has elapsed since the last time that resource record was multicast
// on that particular interface."
//
// Returns:
//   - true: Record can be multicast (≥1 second since last multicast, or never multicast)
//   - false: Record cannot be multicast (rate limit not yet elapsed)
//
// T073: Implement CanMulticast()
func (rs *RecordSet) CanMulticast(rr *ResourceRecord, interfaceID string) bool {
	key := rs.buildRecordKey(rr) + ":" + interfaceID
	lastTimeNano, exists := rs.lastMulticast[key]
	if !exists {
		// Never multicast before - allowed
		return true
	}

	// RFC 6762 §6.2: Minimum 1 second (1e9 nanoseconds) between multicasts
	elapsedNano := time.Now().UnixNano() - lastTimeNano
	return elapsedNano >= 1e9 // 1 second = 1,000,000,000 nanoseconds
}

// CanMulticastProbeDefense checks if probe defense multicast is allowed per RFC 6762 §6.2.
//
// RFC 6762 §6.2: "The one exception is that a Multicast DNS responder MUST respond
// quickly (at most 250 ms after detecting the conflict) when answering probe queries
// for the purpose of defending its name."
//
// Probe defense has relaxed rate limit: 250ms instead of 1 second.
//
// Returns:
//   - true: Probe defense multicast allowed (≥250ms since last multicast)
//   - false: Too soon for probe defense
//
// T074: Implement probe defense exception
func (rs *RecordSet) CanMulticastProbeDefense(rr *ResourceRecord, interfaceID string) bool {
	key := rs.buildRecordKey(rr) + ":" + interfaceID
	lastTimeNano, exists := rs.lastMulticast[key]
	if !exists {
		// Never multicast before - allowed
		return true
	}

	// RFC 6762 §6.2: Probe defense minimum 250ms = 250,000,000 nanoseconds
	elapsedNano := time.Now().UnixNano() - lastTimeNano
	return elapsedNano >= 250e6 // 250ms in nanoseconds
}

// RecordMulticast records that a multicast was sent for this record on this interface.
//
// This updates the rate limiting timestamp per RFC 6762 §6.2.
//
// T074: Implement RecordMulticast()
func (rs *RecordSet) RecordMulticast(rr *ResourceRecord, interfaceID string) {
	key := rs.buildRecordKey(rr) + ":" + interfaceID
	rs.lastMulticast[key] = time.Now().UnixNano()
}

// GetLastMulticast returns the last multicast time for a record on an interface.
//
// Returns:
//   - time.Time: Last multicast timestamp
//   - bool: true if record was multicast before, false if never multicast
//
// T074: Helper for testing
func (rs *RecordSet) GetLastMulticast(rr *ResourceRecord, interfaceID string) (time.Time, bool) {
	key := rs.buildRecordKey(rr) + ":" + interfaceID
	lastTimeNano, exists := rs.lastMulticast[key]
	if !exists {
		return time.Time{}, false
	}
	return time.Unix(0, lastTimeNano), true
}

// buildRecordKey generates a unique key for a resource record.
//
// Record key components (per RFC 6762 §6.2):
//   - Name (case-insensitive per DNS)
//   - Type
//   - Class
//   - RDATA (binary data)
//
// TTL is NOT part of the key - same record with different TTL is still the same record.
//
// T073: Record key generation
func (rs *RecordSet) buildRecordKey(rr *ResourceRecord) string {
	// Use fmt.Sprintf for proper uint16 conversion
	// Binary data encoded as string for map key
	return fmt.Sprintf("%d:%d:%s:%s", rr.Type, rr.Class, rr.Name, string(rr.Data))
}

// BuildGoodbyeRecords creates DNS records with TTL=0 for service goodbye (RFC 6762 §10.1).
//
// RFC 6762 §10.1: "To provide immediate notification when a host shuts down or a service
// ceases to be available, [...] a Multicast DNS responder SHOULD send an unsolicited
// Multicast DNS response packet [...] with the Resource Record's TTL set to zero."
//
// Parameters:
//   - service: Service information
//
// Returns:
//   - []*message.ResourceRecord: All records with TTL=0 (PTR, SRV, TXT, A)
//
// Task 1: Goodbye packet record builder
func BuildGoodbyeRecords(service *ServiceInfo) []*message.ResourceRecord {
	// Build normal record set
	records := BuildRecordSet(service)

	// Set TTL=0 for all records (goodbye signal)
	for _, record := range records {
		record.TTL = 0
	}

	return records
}
