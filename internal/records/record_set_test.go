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

// TestBuildARecord_EdgeCases tests buildARecord with various IPv4 address edge cases.
//
// buildARecord has special handling for invalid IPv4 addresses (not 4 bytes).
// Per the code comment: "Invalid IPv4 address - return placeholder"
//
// Coverage improvement: buildARecord (66.7% → 100%)
func TestBuildARecord_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		ipv4Address []byte
		wantIP      []byte
		description string
	}{
		{
			name:        "valid IPv4 address (4 bytes)",
			ipv4Address: []byte{192, 168, 1, 100},
			wantIP:      []byte{192, 168, 1, 100},
			description: "Normal case: valid 4-byte IPv4 address",
		},
		{
			name:        "empty IPv4 address",
			ipv4Address: []byte{},
			wantIP:      []byte{0, 0, 0, 0},
			description: "Edge case: empty slice gets placeholder 0.0.0.0",
		},
		{
			name:        "nil IPv4 address",
			ipv4Address: nil,
			wantIP:      []byte{0, 0, 0, 0},
			description: "Edge case: nil gets placeholder 0.0.0.0",
		},
		{
			name:        "too short IPv4 address (3 bytes)",
			ipv4Address: []byte{192, 168, 1},
			wantIP:      []byte{0, 0, 0, 0},
			description: "Edge case: < 4 bytes gets placeholder 0.0.0.0",
		},
		{
			name:        "too long IPv4 address (5 bytes)",
			ipv4Address: []byte{192, 168, 1, 100, 255},
			wantIP:      []byte{0, 0, 0, 0},
			description: "Edge case: > 4 bytes gets placeholder 0.0.0.0",
		},
		{
			name:        "loopback address",
			ipv4Address: []byte{127, 0, 0, 1},
			wantIP:      []byte{127, 0, 0, 1},
			description: "Valid case: loopback 127.0.0.1",
		},
		{
			name:        "broadcast address",
			ipv4Address: []byte{255, 255, 255, 255},
			wantIP:      []byte{255, 255, 255, 255},
			description: "Valid case: broadcast 255.255.255.255",
		},
		{
			name:        "zero address",
			ipv4Address: []byte{0, 0, 0, 0},
			wantIP:      []byte{0, 0, 0, 0},
			description: "Valid case: 0.0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &ServiceInfo{
				InstanceName: "Test Service",
				ServiceType:  "_http._tcp.local",
				Hostname:     "testhost.local",
				Port:         8080,
				IPv4Address:  tt.ipv4Address,
			}

			record := buildARecord(service)

			// Verify record was created
			if record == nil {
				t.Fatal("buildARecord() returned nil")
			}

			// Verify record type
			if record.Type != protocol.RecordTypeA {
				t.Errorf("Type = %v, want RecordTypeA", record.Type)
			}

			// Verify hostname
			if record.Name != "testhost.local" {
				t.Errorf("Name = %q, want \"testhost.local\"", record.Name)
			}

			// Verify TTL per RFC 6762 §10
			wantTTL := uint32(4500)
			if record.TTL != wantTTL {
				t.Errorf("TTL = %d, want %d (RFC 6762 §10: 4500s for hostname records)",
					record.TTL, wantTTL)
			}

			// Verify cache-flush bit (A is unique)
			if !record.CacheFlush {
				t.Error("CacheFlush = false, want true (A is unique record)")
			}

			// Verify IPv4 address
			gotIP := record.Data

			if len(gotIP) != 4 {
				t.Errorf("Data length = %d, want 4 bytes", len(gotIP))
			}

			for i := 0; i < 4 && i < len(gotIP); i++ {
				if gotIP[i] != tt.wantIP[i] {
					t.Errorf("Data[%d] = %d, want %d (%s)", i, gotIP[i], tt.wantIP[i], tt.description)
				}
			}

			// Verify service.IPv4Address was modified if it was invalid
			if len(tt.ipv4Address) != 4 {
				// Should have been set to placeholder
				if len(service.IPv4Address) != 4 {
					t.Errorf("service.IPv4Address length = %d, want 4 (should be fixed to placeholder)",
						len(service.IPv4Address))
				}
				for i := 0; i < 4 && i < len(service.IPv4Address); i++ {
					if service.IPv4Address[i] != 0 {
						t.Errorf("service.IPv4Address[%d] = %d, want 0 (placeholder)", i, service.IPv4Address[i])
					}
				}
			}

			t.Logf("✓ %s", tt.description)
		})
	}
}

// TestBuildARecord_RFC6762_Compliance tests RFC 6762 compliance of buildARecord.
//
// Validates that A records conform to RFC 6762 requirements.
//
// Coverage improvement: buildARecord RFC validation
func TestBuildARecord_RFC6762_Compliance(t *testing.T) {
	service := &ServiceInfo{
		InstanceName: "My Service",
		ServiceType:  "_http._tcp.local",
		Hostname:     "myhost.local",
		Port:         8080,
		IPv4Address:  []byte{10, 0, 0, 1},
	}

	record := buildARecord(service)

	// RFC 6762 §10: Hostname records use 4500 seconds (75 minutes)
	if record.TTL != 4500 {
		t.Errorf("TTL = %d, want 4500 (RFC 6762 §10: hostname records)", record.TTL)
	}

	// RFC 6762 §10.2: A records are unique (cache-flush bit set)
	if !record.CacheFlush {
		t.Error("CacheFlush = false, want true (RFC 6762 §10.2: unique records)")
	}

	// RFC 1035 §3.2.2: Class IN
	if record.Class != protocol.ClassIN {
		t.Errorf("Class = %v, want ClassIN (RFC 1035 §3.2.2)", record.Class)
	}

	// RFC 1035 §3.2.2: Type A
	if record.Type != protocol.RecordTypeA {
		t.Errorf("Type = %v, want RecordTypeA (RFC 1035 §3.2.2)", record.Type)
	}

	// RFC 1035 §3.4.1: A record RDATA is 4 octets
	data := record.Data
	if len(data) != 4 {
		t.Errorf("Data length = %d, want 4 (RFC 1035 §3.4.1: A record is 4 octets)", len(data))
	}
}
