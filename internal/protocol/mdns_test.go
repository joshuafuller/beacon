package protocol

import (
	"testing"
)

// TestPort validates that the mDNS port constant is 5353 per RFC 6762 §5 (FR-004).
//
// RFC 6762 §5 states: "Multicast DNS uses UDP port 5353."
//
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
func TestPort(t *testing.T) {
	want := 5353
	if Port != want {
		t.Errorf("Port = %d, want %d per RFC 6762 §5", Port, want)
	}
}

// TestMulticastAddrIPv4 validates that the mDNS IPv4 multicast address is
// 224.0.0.251 per RFC 6762 §5 (FR-004).
//
// RFC 6762 §5 states: "The IPv4 link-local multicast address is 224.0.0.251."
//
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
func TestMulticastAddrIPv4(t *testing.T) {
	// Test validates constant matches RFC value, hardcoded string is intentional
	want := "224.0.0.251"          // nosemgrep: beacon-hardcoded-multicast-address
	if MulticastAddrIPv4 != want { // nosemgrep: beacon-hardcoded-multicast-address
		t.Errorf("MulticastAddrIPv4 = %s, want %s per RFC 6762 §5", MulticastAddrIPv4, want) // nosemgrep: beacon-hardcoded-multicast-address
	}
}

// TestMulticastGroupIPv4 validates that MulticastGroupIPv4() returns the correct
// UDP address for mDNS multicast per RFC 6762 §5 (FR-004).
//
// RFC 6762 §5: "224.0.0.251:5353"
//
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries
func TestMulticastGroupIPv4(t *testing.T) {
	addr := MulticastGroupIPv4()

	// Test validates constants match RFC values, hardcoded strings are intentional
	wantIP := "224.0.0.251" // nosemgrep: beacon-hardcoded-multicast-address
	wantPort := 5353

	if addr.IP.String() != wantIP { // nosemgrep: beacon-hardcoded-multicast-address
		t.Errorf("MulticastGroupIPv4().IP = %s, want %s per RFC 6762 §5", addr.IP, wantIP) // nosemgrep: beacon-hardcoded-multicast-address
	}

	if addr.Port != wantPort {
		t.Errorf("MulticastGroupIPv4().Port = %d, want %d per RFC 6762 §5", addr.Port, wantPort)
	}

	// Verify it's a valid multicast address
	if !addr.IP.IsMulticast() {
		t.Errorf("MulticastGroupIPv4().IP is not a multicast address")
	}
}

// TestRecordType_String validates that RecordType.String() returns correct
// human-readable names per RFC 1035 (FR-002).
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
func TestRecordType_String(t *testing.T) {
	tests := []struct {
		name       string
		recordType RecordType
		want       string
	}{
		{
			name:       "A record",
			recordType: RecordTypeA,
			want:       "A",
		},
		{
			name:       "PTR record",
			recordType: RecordTypePTR,
			want:       "PTR",
		},
		{
			name:       "TXT record",
			recordType: RecordTypeTXT,
			want:       "TXT",
		},
		{
			name:       "SRV record",
			recordType: RecordTypeSRV,
			want:       "SRV",
		},
		{
			name:       "Unknown record type",
			recordType: RecordType(999),
			want:       "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.recordType.String()
			if got != tt.want {
				t.Errorf("RecordType(%d).String() = %s, want %s", tt.recordType, got, tt.want)
			}
		})
	}
}

// TestRecordType_IsSupported validates that RecordType.IsSupported() returns
// true for M1-supported types (A, PTR, SRV, TXT) per FR-002.
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
// FR-014: System MUST return ValidationError for invalid query names or unsupported record types
func TestRecordType_IsSupported(t *testing.T) {
	tests := []struct {
		name       string
		recordType RecordType
		want       bool
	}{
		{
			name:       "A record supported per FR-002",
			recordType: RecordTypeA,
			want:       true,
		},
		{
			name:       "PTR record supported per FR-002",
			recordType: RecordTypePTR,
			want:       true,
		},
		{
			name:       "TXT record supported per FR-002",
			recordType: RecordTypeTXT,
			want:       true,
		},
		{
			name:       "SRV record supported per FR-002",
			recordType: RecordTypeSRV,
			want:       true,
		},
		{
			name:       "AAAA record not supported in M1",
			recordType: RecordType(28), // AAAA (IPv6)
			want:       false,
		},
		{
			name:       "MX record not supported in M1",
			recordType: RecordType(15), // MX (mail exchange)
			want:       false,
		},
		{
			name:       "Unknown record type not supported",
			recordType: RecordType(999),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.recordType.IsSupported()
			if got != tt.want {
				t.Errorf("RecordType(%d).IsSupported() = %v, want %v", tt.recordType, got, tt.want)
			}
		})
	}
}

// TestRecordType_Values validates that record type constants have the correct
// numeric values per RFC 1035 §3.2.2 and RFC 2782.
//
// RFC 1035 §3.2.2 defines: A=1, PTR=12, TXT=16
// RFC 2782 defines: SRV=33
//
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
func TestRecordType_Values(t *testing.T) {
	tests := []struct {
		name       string
		recordType RecordType
		wantValue  uint16
	}{
		{
			name:       "A record value per RFC 1035 §3.2.2",
			recordType: RecordTypeA,
			wantValue:  1,
		},
		{
			name:       "PTR record value per RFC 1035 §3.2.2",
			recordType: RecordTypePTR,
			wantValue:  12,
		},
		{
			name:       "TXT record value per RFC 1035 §3.2.2",
			recordType: RecordTypeTXT,
			wantValue:  16,
		},
		{
			name:       "SRV record value per RFC 2782",
			recordType: RecordTypeSRV,
			wantValue:  33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uint16(tt.recordType)
			if got != tt.wantValue {
				t.Errorf("RecordType constant = %d, want %d", got, tt.wantValue)
			}
		})
	}
}

// TestClassIN validates that ClassIN has the correct value (1) per RFC 1035 §3.2.4.
//
// RFC 1035 §3.2.4: "IN = 1 the Internet"
func TestClassIN(t *testing.T) {
	want := uint16(1)
	got := uint16(ClassIN)
	if got != want {
		t.Errorf("ClassIN = %d, want %d per RFC 1035 §3.2.4", got, want)
	}
}

// TestDNSHeaderFlags validates that DNS header flag constants have the correct
// bit values per RFC 1035 §4.1.1 and RFC 6762 §18.
//
// RFC 1035 §4.1.1 defines header format with bit positions.
// RFC 6762 §18 defines mDNS-specific header field requirements.
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18
func TestDNSHeaderFlags(t *testing.T) {
	tests := []struct {
		name      string
		flag      uint16
		wantValue uint16
		rfcRef    string
	}{
		{
			name:      "QR bit (bit 15) per RFC 1035 §4.1.1",
			flag:      FlagQR,
			wantValue: 0x8000,
			rfcRef:    "RFC 1035 §4.1.1, RFC 6762 §18.2",
		},
		{
			name:      "AA bit (bit 10) per RFC 1035 §4.1.1",
			flag:      FlagAA,
			wantValue: 0x0400,
			rfcRef:    "RFC 1035 §4.1.1, RFC 6762 §18.4",
		},
		{
			name:      "TC bit (bit 9) per RFC 1035 §4.1.1",
			flag:      FlagTC,
			wantValue: 0x0200,
			rfcRef:    "RFC 1035 §4.1.1, RFC 6762 §18.5",
		},
		{
			name:      "RD bit (bit 8) per RFC 1035 §4.1.1",
			flag:      FlagRD,
			wantValue: 0x0100,
			rfcRef:    "RFC 1035 §4.1.1, RFC 6762 §18.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.flag != tt.wantValue {
				t.Errorf("Flag = 0x%04X, want 0x%04X per %s", tt.flag, tt.wantValue, tt.rfcRef)
			}
		})
	}
}

// TestOpcodeQuery validates that OpcodeQuery is zero per RFC 6762 §18.3.
//
// RFC 6762 §18.3 states: "In both multicast query and multicast response messages,
// the OPCODE MUST be zero on transmission."
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18 (OPCODE=0 per §18.3)
func TestOpcodeQuery(t *testing.T) {
	want := uint16(0)
	if OpcodeQuery != want {
		t.Errorf("OpcodeQuery = %d, want %d per RFC 6762 §18.3", OpcodeQuery, want)
	}
}

// TestRCodeNoError validates that RCodeNoError is zero per RFC 6762 §18.11.
//
// RFC 6762 §18.11 states: "Multicast DNS messages received with non-zero
// Response Codes MUST be silently ignored."
//
// FR-022: System MUST ignore responses with RCODE != 0 per RFC 6762 §18.11
func TestRCodeNoError(t *testing.T) {
	want := uint16(0)
	if RCodeNoError != want {
		t.Errorf("RCodeNoError = %d, want %d per RFC 6762 §18.11", RCodeNoError, want)
	}
}

// TestDNSNameConstraints validates DNS name constraint constants per RFC 1035 §3.1.
//
// RFC 1035 §3.1 defines domain name format with label and name length limits.
//
// FR-003: System MUST validate queried names follow DNS naming rules
func TestDNSNameConstraints(t *testing.T) {
	tests := []struct {
		name      string
		constant  int
		wantValue int
		rfcRef    string
	}{
		{
			name:      "MaxLabelLength per RFC 1035 §3.1",
			constant:  MaxLabelLength,
			wantValue: 63,
			rfcRef:    "RFC 1035 §3.1 (labels ≤63 bytes)",
		},
		{
			name:      "MaxNameLength per RFC 1035 §3.1",
			constant:  MaxNameLength,
			wantValue: 255,
			rfcRef:    "RFC 1035 §3.1 (total name ≤255 bytes)",
		},
		{
			name:      "MaxCompressionPointers per RFC 1035 §4.1.4",
			constant:  MaxCompressionPointers,
			wantValue: 256,
			rfcRef:    "RFC 1035 §4.1.4 (loop detection)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.wantValue {
				t.Errorf("%s = %d, want %d per %s", tt.name, tt.constant, tt.wantValue, tt.rfcRef)
			}
		})
	}
}

// TestCompressionMask validates the compression pointer mask (0xC0) per RFC 1035 §4.1.4.
//
// RFC 1035 §4.1.4 states: "The pointer takes the form of a two octet sequence where
// the first two bits are ones."
//
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4 (message compression)
func TestCompressionMask(t *testing.T) {
	want := byte(0xC0) // Binary: 11000000 (high 2 bits = 11)
	if CompressionMask != want {
		t.Errorf("CompressionMask = 0x%02X, want 0x%02X per RFC 1035 §4.1.4", CompressionMask, want)
	}
}

// TestMulticastGroupIPv4_IsLinkLocal validates that the multicast address
// is link-local per RFC 6762 §5.
//
// RFC 6762 §5 specifies that mDNS uses "link-local multicast".
func TestMulticastGroupIPv4_IsLinkLocal(t *testing.T) {
	addr := MulticastGroupIPv4()

	// 224.0.0.251 is in the 224.0.0.0/24 link-local multicast range
	ip := addr.IP.To4()
	if ip == nil {
		t.Fatal("MulticastGroupIPv4() returned non-IPv4 address")
	}

	// Link-local multicast range: 224.0.0.0 - 224.0.0.255
	if ip[0] != 224 || ip[1] != 0 || ip[2] != 0 {
		t.Errorf("MulticastGroupIPv4() IP %s is not in link-local range 224.0.0.0/24 per RFC 6762 §5", ip)
	}
}

// TestMulticastGroupIPv4_NotNil validates that MulticastGroupIPv4() never returns nil.
func TestMulticastGroupIPv4_NotNil(t *testing.T) {
	addr := MulticastGroupIPv4()
	if addr == nil {
		t.Fatal("MulticastGroupIPv4() returned nil")
	}
	if addr.IP == nil {
		t.Fatal("MulticastGroupIPv4().IP is nil")
	}
}
