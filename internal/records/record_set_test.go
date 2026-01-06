package records

import (
	"testing"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestBuildTXTRecord_EmptyMandatory_RED tests RFC 6763 §6 mandatory TXT record.
//
// TDD Phase: RED - These tests will FAIL until we implement buildTXTRecord()
//
// RFC 6763 §6: "If a DNS-SD service has no TXT records, it MUST include a
// single TXT record consisting of a single zero byte (0x00)."
//
// FR-031: System MUST create mandatory TXT record with 0x00 byte if empty
// T027: Write TXT record tests
func TestBuildTXTRecord_Empty(t *testing.T) {
	txtRecords := map[string]string{} // Empty TXT records

	data := buildTXTRecord(txtRecords)

	// Empty TXT MUST be encoded as single 0x00 byte per RFC 6763 §6
	if len(data) != 1 || data[0] != 0x00 {
		t.Errorf("buildTXTRecord(empty) = %v, want [0x00]", data)
	}
}

// TestBuildTXTRecord_SingleKey_RED tests encoding a single key-value pair.
//
// TDD Phase: RED
//
// RFC 6763 §6.4: TXT record format
//   - Length byte + "key=value" string
//   - Example: "version=1.0" → [0x0b, 'v','e','r','s','i','o','n','=','1','.','0']
//
// T027: Test single key-value encoding
func TestBuildTXTRecord_SingleKey(t *testing.T) {
	txtRecords := map[string]string{
		"version": "1.0",
	}

	data := buildTXTRecord(txtRecords)

	// "version=1.0" = 11 bytes
	// Expected: [0x0b, 'v','e','r','s','i','o','n','=','1','.','0']
	if len(data) == 0 {
		t.Error("buildTXTRecord(single key) returned empty data, want encoded key-value")
	}

	// First byte should be length (11 = 0x0b)
	if data[0] != 0x0b {
		t.Errorf("buildTXTRecord(single key) length byte = 0x%02x, want 0x0b", data[0])
	}

	// Verify key-value string is present
	keyValue := "version=1.0"
	if len(data) < len(keyValue)+1 {
		t.Errorf("buildTXTRecord(single key) data too short: %d bytes, want at least %d", len(data), len(keyValue)+1)
	}
}

// TestBuildTXTRecord_MultipleKeys_RED tests encoding multiple key-value pairs.
//
// TDD Phase: RED
//
// RFC 6763 §6.4: Multiple key-value pairs are concatenated
//   - Each pair has its own length byte
//   - Example: "version=1.0" + "path=/api"
//
// T027: Test multiple key-value encoding
func TestBuildTXTRecord_MultipleKeys(t *testing.T) {
	txtRecords := map[string]string{
		"version": "1.0",
		"path":    "/api",
	}

	data := buildTXTRecord(txtRecords)

	// Should have at least 2 entries (version=1.0 and path=/api)
	// Each entry: length byte + data
	if len(data) < 20 { // Rough estimate: 11 (version) + 9 (path) bytes
		t.Errorf("buildTXTRecord(multiple keys) data too short: %d bytes", len(data))
	}

	// Verify we have multiple length-prefixed strings
	// (Detailed parsing will be done in GREEN phase)
	if data[0] == 0x00 {
		t.Error("buildTXTRecord(multiple keys) starts with 0x00, want length-prefixed strings")
	}
}

// TestBuildRecordSet_RED tests building complete record set for a service.
//
// TDD Phase: RED
//
// RFC 6763 §6: A registered service includes:
//   - PTR record: _service._proto.local → instance._service._proto.local
//   - SRV record: instance._service._proto.local → hostname:port
//   - TXT record: instance._service._proto.local → key-value pairs
//   - A record: hostname.local → IPv4 address
//
// FR-032: System MUST build complete record set (PTR, SRV, TXT, A)
// T027: Write record set tests
func TestBuildRecordSet_AllRecordTypes(t *testing.T) {
	service := ServiceInfo{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Hostname:     "myhost.local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
		TXTRecords:   map[string]string{"version": "1.0"},
	}

	recordSet := BuildRecordSet(&service)

	// Verify record set contains all 4 record types
	foundTypes := make(map[protocol.RecordType]bool)
	for _, record := range recordSet {
		foundTypes[record.Type] = true
	}

	wantTypes := []protocol.RecordType{
		protocol.RecordTypePTR,
		protocol.RecordTypeSRV,
		protocol.RecordTypeTXT,
		protocol.RecordTypeA,
	}

	for _, wantType := range wantTypes {
		if !foundTypes[wantType] {
			t.Errorf("BuildRecordSet() missing record type %v", wantType)
		}
	}

	// Should have exactly 4 records
	if len(recordSet) != 4 {
		t.Errorf("BuildRecordSet() returned %d records, want 4 (PTR, SRV, TXT, A)", len(recordSet))
	}
}

// TestBuildRecordSet_PTRRecord_RED tests PTR record construction.
//
// TDD Phase: RED
//
// RFC 6763 §6: PTR record format
//   - Name: _service._proto.local (e.g., "_http._tcp.local")
//   - RDATA: instance._service._proto.local (e.g., "My Printer._http._tcp.local")
//   - TTL: 120 seconds (service TTL per RFC 6762 §10)
//
// T027: Test PTR record
func TestBuildRecordSet_PTRRecord(t *testing.T) {
	service := ServiceInfo{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Hostname:     "myhost.local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
	}

	recordSet := BuildRecordSet(&service)

	// Find PTR record
	var ptrRecord *ResourceRecord
	for _, record := range recordSet {
		if record.Type == protocol.RecordTypePTR {
			ptrRecord = record
			break
		}
	}

	if ptrRecord == nil {
		t.Fatal("BuildRecordSet() did not include PTR record")
	}

	// Verify PTR record fields
	wantName := "_http._tcp.local"
	if ptrRecord.Name != wantName {
		t.Errorf("PTR record Name = %q, want %q", ptrRecord.Name, wantName)
	}

	// RFC 6762 §10: PTR records for DNS-SD services use 120 seconds
	// Service discovery records change more frequently than hostname records
	wantTTL := uint32(120)
	if ptrRecord.TTL != wantTTL {
		t.Errorf("PTR record TTL = %d, want %d (RFC 6762 §10: 120s for service records)", ptrRecord.TTL, wantTTL)
	}
}

// TestBuildRecordSet_SRVRecord_RED tests SRV record construction.
//
// TDD Phase: RED
//
// RFC 6763 §6: SRV record format
//   - Name: instance._service._proto.local
//   - RDATA: priority (0), weight (0), port, hostname
//   - TTL: 120 seconds
//   - Cache-flush: true (unique record per RFC 6762 §10.2)
//
// T027: Test SRV record
func TestBuildRecordSet_SRVRecord(t *testing.T) {
	service := ServiceInfo{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Hostname:     "myhost.local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
	}

	recordSet := BuildRecordSet(&service)

	// Find SRV record
	var srvRecord *ResourceRecord
	for _, record := range recordSet {
		if record.Type == protocol.RecordTypeSRV {
			srvRecord = record
			break
		}
	}

	if srvRecord == nil {
		t.Fatal("BuildRecordSet() did not include SRV record")
	}

	// Verify SRV record fields
	wantName := "My Printer._http._tcp.local"
	if srvRecord.Name != wantName {
		t.Errorf("SRV record Name = %q, want %q", srvRecord.Name, wantName)
	}

	wantTTL := uint32(120)
	if srvRecord.TTL != wantTTL {
		t.Errorf("SRV record TTL = %d, want %d", srvRecord.TTL, wantTTL)
	}

	// SRV is a unique record, should have cache-flush bit
	if !srvRecord.CacheFlush {
		t.Error("SRV record CacheFlush = false, want true (unique record)")
	}
}

// TestBuildRecordSet_ARecord_RED tests A record construction.
//
// TDD Phase: RED
//
// RFC 6762 §6: A record format
//   - Name: hostname.local
//   - RDATA: IPv4 address (4 bytes)
//   - TTL: 4500 seconds (hostname TTL per RFC 6762 §10)
//   - Cache-flush: true (unique record)
//
// T027: Test A record
func TestBuildRecordSet_ARecord(t *testing.T) {
	service := ServiceInfo{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Hostname:     "myhost.local",
		Port:         8080,
		IPv4Address:  []byte{192, 168, 1, 100},
	}

	recordSet := BuildRecordSet(&service)

	// Find A record
	var aRecord *ResourceRecord
	for _, record := range recordSet {
		if record.Type == protocol.RecordTypeA {
			aRecord = record
			break
		}
	}

	if aRecord == nil {
		t.Fatal("BuildRecordSet() did not include A record")
	}

	// Verify A record fields
	wantName := "myhost.local"
	if aRecord.Name != wantName {
		t.Errorf("A record Name = %q, want %q", aRecord.Name, wantName)
	}

	// RFC 6762 §10: A records use 4500 seconds (75 minutes)
	// Hostname records change less frequently than service discovery records
	wantTTL := uint32(4500)
	if aRecord.TTL != wantTTL {
		t.Errorf("A record TTL = %d, want %d (RFC 6762 §10: 4500s for hostname records)", aRecord.TTL, wantTTL)
	}

	// A is a unique record, should have cache-flush bit
	if !aRecord.CacheFlush {
		t.Error("A record CacheFlush = false, want true (unique record)")
	}

	// Verify IPv4 address data
	if len(aRecord.Data) != 4 {
		t.Errorf("A record Data length = %d, want 4 bytes", len(aRecord.Data))
	}
}

// TestResourceRecord_CanMulticast tests per-record multicast rate limiting.
//
// RFC 6762 §6.2: "A Multicast DNS responder MUST NOT multicast a given resource record
// on a given interface until at least one second has elapsed since the last time that
// resource record was multicast on that particular interface."
//
// Rate limiting is PER RECORD, PER INTERFACE to prevent network flooding.
//
// TDD Phase: RED (test written first)
//
// T069 [P] [US3]: Unit test per-record multicast rate limiting (1 second minimum)
func TestResourceRecord_CanMulticast(t *testing.T) {
	// Create a resource record
	rr := &ResourceRecord{
		Name:  "myservice._http._tcp.local",
		Type:  protocol.RecordTypePTR,
		Class: protocol.ClassIN,
		TTL:   4500,
		Data:  []byte{0x08, 'M', 'y', 'P', 'r', 'i', 'n', 't', 'e', 'r'},
	}

	// Interface ID (e.g., "eth0")
	interfaceID := "eth0"

	// Create record set tracker
	rs := NewRecordSet()

	// First multicast - should be allowed
	canMulticast := rs.CanMulticast(rr, interfaceID)
	if !canMulticast {
		t.Error("CanMulticast() = false for first multicast, want true")
	}

	// Record the multicast
	rs.RecordMulticast(rr, interfaceID)

	// Immediate retry - should be denied (< 1 second)
	canMulticast = rs.CanMulticast(rr, interfaceID)
	if canMulticast {
		t.Error("CanMulticast() = true immediately after multicast, want false (RFC 6762 §6.2: 1 second minimum)")
	}
}

// TestResourceRecord_CanMulticast_PerInterface tests rate limiting is per-interface.
//
// RFC 6762 §6.2: Rate limiting is "on a given interface" - different interfaces have
// independent rate limits.
//
// TDD Phase: RED
//
// T069 [P] [US3]: Verify rate limiting is per-interface
func TestResourceRecord_CanMulticast_PerInterface(t *testing.T) {
	rr := &ResourceRecord{
		Name:  "myservice._http._tcp.local",
		Type:  protocol.RecordTypePTR,
		Class: protocol.ClassIN,
		TTL:   4500,
		Data:  []byte{0x08, 'M', 'y', 'P', 'r', 'i', 'n', 't', 'e', 'r'},
	}

	rs := NewRecordSet()

	// Multicast on eth0
	rs.RecordMulticast(rr, "eth0")

	// Immediate multicast on eth0 - denied
	if rs.CanMulticast(rr, "eth0") {
		t.Error("CanMulticast(eth0) = true immediately after multicast on eth0, want false")
	}

	// Immediate multicast on wlan0 - allowed (different interface)
	if !rs.CanMulticast(rr, "wlan0") {
		t.Error("CanMulticast(wlan0) = false, want true (different interface from eth0)")
	}

	// Multicast on wlan0
	rs.RecordMulticast(rr, "wlan0")

	// Now wlan0 is also rate-limited
	if rs.CanMulticast(rr, "wlan0") {
		t.Error("CanMulticast(wlan0) = true immediately after multicast on wlan0, want false")
	}
}

// TestResourceRecord_CanMulticast_PerRecord tests rate limiting is per-record.
//
// RFC 6762 §6.2: Rate limiting is for "a given resource record" - different records
// have independent rate limits even on same interface.
//
// TDD Phase: RED
//
// T069 [P] [US3]: Verify rate limiting is per-record
func TestResourceRecord_CanMulticast_PerRecord(t *testing.T) {
	rr1 := &ResourceRecord{
		Name:  "service1._http._tcp.local",
		Type:  protocol.RecordTypePTR,
		Class: protocol.ClassIN,
		TTL:   4500,
		Data:  []byte{0x08, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '1'},
	}

	rr2 := &ResourceRecord{
		Name:  "service2._http._tcp.local",
		Type:  protocol.RecordTypePTR,
		Class: protocol.ClassIN,
		TTL:   4500,
		Data:  []byte{0x08, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '2'},
	}

	rs := NewRecordSet()

	// Multicast rr1 on eth0
	rs.RecordMulticast(rr1, "eth0")

	// Immediate multicast of rr1 - denied
	if rs.CanMulticast(rr1, "eth0") {
		t.Error("CanMulticast(rr1, eth0) = true immediately after multicast, want false")
	}

	// Immediate multicast of rr2 - allowed (different record)
	if !rs.CanMulticast(rr2, "eth0") {
		t.Error("CanMulticast(rr2, eth0) = false, want true (different record from rr1)")
	}
}

// TestResourceRecord_CanMulticast_ProbeDefense tests probe defense rate limit exception.
//
// RFC 6762 §6.2: "The one exception is that a Multicast DNS responder MUST respond
// quickly (at most 250 ms after detecting the conflict) when answering probe queries
// for the purpose of defending its name."
//
// Probe defense allows 250ms minimum instead of 1 second.
//
// TDD Phase: RED
//
// T070 [P] [US3]: Unit test probe defense rate limit exception
func TestResourceRecord_CanMulticast_ProbeDefense(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing test in short mode")
	}

	rr := &ResourceRecord{
		Name:  "myservice._http._tcp.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100},
	}

	rs := NewRecordSet()

	// Multicast rr on eth0
	rs.RecordMulticast(rr, "eth0")

	// Immediate probe defense - denied (< 250ms)
	canMulticast := rs.CanMulticastProbeDefense(rr, "eth0")
	if canMulticast {
		t.Error("CanMulticastProbeDefense() = true immediately, want false (< 250ms)")
	}

	// Regular multicast also denied (< 1 second)
	canMulticastRegular := rs.CanMulticast(rr, "eth0")
	if canMulticastRegular {
		t.Error("CanMulticast() = true immediately, want false (1 second minimum for regular responses)")
	}
}

// =============================================================================
// buildARecord() Tests - Coverage Increase from 66.7% to 90%+
// =============================================================================

// TestBuildARecord_ValidIPv4 tests A record construction with valid IPv4 address.
//
// Coverage: buildARecord line 221-234 (happy path)
// RFC 6762 §10: Hostname records use 4500 seconds (75 minutes) TTL
func TestBuildARecord_ValidIPv4(t *testing.T) {
	service := &ServiceInfo{
		Hostname:    "myhost.local",
		IPv4Address: []byte{192, 168, 1, 100},
	}

	record := buildARecord(service)

	// Verify record structure
	if record.Name != "myhost.local" {
		t.Errorf("buildARecord() Name = %q, want %q", record.Name, "myhost.local")
	}

	if record.Type != protocol.RecordTypeA {
		t.Errorf("buildARecord() Type = %v, want %v", record.Type, protocol.RecordTypeA)
	}

	if record.Class != protocol.ClassIN {
		t.Errorf("buildARecord() Class = %v, want %v", record.Class, protocol.ClassIN)
	}

	// RFC 6762 §10: A records use 4500 seconds (75 minutes)
	wantTTL := uint32(4500)
	if record.TTL != wantTTL {
		t.Errorf("buildARecord() TTL = %d, want %d (RFC 6762 §10)", record.TTL, wantTTL)
	}

	// Verify IPv4 address data
	if len(record.Data) != 4 {
		t.Errorf("buildARecord() Data length = %d, want 4", len(record.Data))
	}

	wantIP := []byte{192, 168, 1, 100}
	for i := 0; i < 4; i++ {
		if record.Data[i] != wantIP[i] {
			t.Errorf("buildARecord() Data[%d] = %d, want %d", i, record.Data[i], wantIP[i])
		}
	}

	// A is a unique record - should have cache-flush bit set
	if !record.CacheFlush {
		t.Error("buildARecord() CacheFlush = false, want true (A is unique record)")
	}

	t.Log("✓ buildARecord() with valid IPv4 creates correct A record")
}

// TestBuildARecord_InvalidIPv4_TooShort tests invalid IPv4 (< 4 bytes).
//
// Coverage: buildARecord line 221-225 (error path - invalid IPv4)
// When IPv4 address is invalid, function should use placeholder 0.0.0.0
func TestBuildARecord_InvalidIPv4_TooShort(t *testing.T) {
	service := &ServiceInfo{
		Hostname:    "myhost.local",
		IPv4Address: []byte{192, 168}, // Only 2 bytes - invalid!
	}

	record := buildARecord(service)

	// Should create placeholder 0.0.0.0
	wantIP := []byte{0, 0, 0, 0}
	if len(record.Data) != 4 {
		t.Errorf("buildARecord(invalid) Data length = %d, want 4", len(record.Data))
	}

	for i := 0; i < 4; i++ {
		if record.Data[i] != wantIP[i] {
			t.Errorf("buildARecord(invalid) Data[%d] = %d, want %d (placeholder 0.0.0.0)", i, record.Data[i], wantIP[i])
		}
	}

	t.Log("✓ buildARecord() with invalid IPv4 creates placeholder 0.0.0.0")
}

// TestBuildARecord_InvalidIPv4_TooLong tests invalid IPv4 (> 4 bytes).
//
// Coverage: buildARecord line 221-225 (error path - invalid IPv4)
func TestBuildARecord_InvalidIPv4_TooLong(t *testing.T) {
	service := &ServiceInfo{
		Hostname:    "myhost.local",
		IPv4Address: []byte{192, 168, 1, 100, 1}, // 5 bytes - invalid!
	}

	record := buildARecord(service)

	// Should create placeholder 0.0.0.0
	wantIP := []byte{0, 0, 0, 0}
	if len(record.Data) != 4 {
		t.Errorf("buildARecord(invalid) Data length = %d, want 4", len(record.Data))
	}

	for i := 0; i < 4; i++ {
		if record.Data[i] != wantIP[i] {
			t.Errorf("buildARecord(invalid) Data[%d] = %d, want %d (placeholder 0.0.0.0)", i, record.Data[i], wantIP[i])
		}
	}

	t.Log("✓ buildARecord() with too-long IPv4 creates placeholder 0.0.0.0")
}

// TestBuildARecord_EmptyIPv4 tests empty IPv4 address.
//
// Coverage: buildARecord line 221-225 (error path - empty IPv4)
func TestBuildARecord_EmptyIPv4(t *testing.T) {
	service := &ServiceInfo{
		Hostname:    "myhost.local",
		IPv4Address: []byte{}, // Empty - invalid!
	}

	record := buildARecord(service)

	// Should create placeholder 0.0.0.0
	wantIP := []byte{0, 0, 0, 0}
	if len(record.Data) != 4 {
		t.Errorf("buildARecord(empty) Data length = %d, want 4", len(record.Data))
	}

	for i := 0; i < 4; i++ {
		if record.Data[i] != wantIP[i] {
			t.Errorf("buildARecord(empty) Data[%d] = %d, want %d (placeholder 0.0.0.0)", i, record.Data[i], wantIP[i])
		}
	}

	t.Log("✓ buildARecord() with empty IPv4 creates placeholder 0.0.0.0")
}
