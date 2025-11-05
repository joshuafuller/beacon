package records

import (
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestTTL_GetRemainingTTL_RED tests remaining TTL calculation.
//
// TDD Phase: RED - These tests will FAIL until we implement TTL management
//
// RFC 6762 §10: TTL values decrease over time
// T017: Implement TTL calculation (GetRemainingTTL, IsExpired)
func TestTTL_GetRemainingTTL(t *testing.T) {
	tests := []struct {
		name       string
		ttl        uint32
		elapsed    time.Duration
		wantRemain uint32
	}{
		{
			name:       "fresh record - no time elapsed",
			ttl:        protocol.TTLHostname, // 4500 seconds
			elapsed:    0,
			wantRemain: 4500,
		},
		{
			name:       "half TTL elapsed",
			ttl:        protocol.TTLService, // 120 seconds
			elapsed:    60 * time.Second,
			wantRemain: 60,
		},
		{
			name:       "almost expired",
			ttl:        protocol.TTLService, // 120 seconds
			elapsed:    119 * time.Second,
			wantRemain: 1,
		},
		{
			name:       "fully elapsed returns 0",
			ttl:        protocol.TTLService, // 120 seconds
			elapsed:    120 * time.Second,
			wantRemain: 0,
		},
		{
			name:       "over-elapsed returns 0",
			ttl:        protocol.TTLService, // 120 seconds
			elapsed:    200 * time.Second,
			wantRemain: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a record with a specific TTL
			record := &RecordTTL{
				TTL:       tt.ttl,
				CreatedAt: time.Now().Add(-tt.elapsed), // Simulate elapsed time
			}

			gotRemain := record.GetRemainingTTL()
			if gotRemain != tt.wantRemain {
				t.Errorf("GetRemainingTTL() = %d, want %d (ttl=%d, elapsed=%v)",
					gotRemain, tt.wantRemain, tt.ttl, tt.elapsed)
			}
		})
	}
}

// TestTTL_IsExpired_RED tests expiration checking.
//
// TDD Phase: RED
//
// RFC 6762 §10: Records expire when TTL reaches zero
// T017: Implement IsExpired() method
func TestTTL_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		ttl         uint32
		elapsed     time.Duration
		wantExpired bool
	}{
		{
			name:        "fresh record not expired",
			ttl:         protocol.TTLService,
			elapsed:     0,
			wantExpired: false,
		},
		{
			name:        "half TTL not expired",
			ttl:         protocol.TTLService, // 120 seconds
			elapsed:     60 * time.Second,
			wantExpired: false,
		},
		{
			name:        "one second before expiry not expired",
			ttl:         protocol.TTLService, // 120 seconds
			elapsed:     119 * time.Second,
			wantExpired: false,
		},
		{
			name:        "exactly at TTL is expired",
			ttl:         protocol.TTLService, // 120 seconds
			elapsed:     120 * time.Second,
			wantExpired: true,
		},
		{
			name:        "past TTL is expired",
			ttl:         protocol.TTLService, // 120 seconds
			elapsed:     200 * time.Second,
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &RecordTTL{
				TTL:       tt.ttl,
				CreatedAt: time.Now().Add(-tt.elapsed),
			}

			gotExpired := record.IsExpired()
			if gotExpired != tt.wantExpired {
				t.Errorf("IsExpired() = %v, want %v (ttl=%d, elapsed=%v)",
					gotExpired, tt.wantExpired, tt.ttl, tt.elapsed)
			}
		})
	}
}

// TestTTL_ServiceVsHostname_RED tests different TTL values per record type.
//
// TDD Phase: RED
//
// RFC 6762 §10:
//   - Service records (SRV, TXT): 120 seconds
//   - Hostname records (A, AAAA): 4500 seconds (75 minutes)
//
// T015: ResourceRecordSet must use correct TTL per record type
func TestTTL_ServiceVsHostname(t *testing.T) {
	tests := []struct {
		name       string
		recordType protocol.RecordType
		wantTTL    uint32
	}{
		{
			name:       "SRV record uses TTLService (120s) per RFC 6762 §10",
			recordType: protocol.RecordTypeSRV,
			wantTTL:    protocol.TTLService,
		},
		{
			name:       "TXT record uses TTLService (120s) per RFC 6762 §10",
			recordType: protocol.RecordTypeTXT,
			wantTTL:    protocol.TTLService,
		},
		{
			name:       "A record uses TTLHostname (4500s) per RFC 6762 §10",
			recordType: protocol.RecordTypeA,
			wantTTL:    protocol.TTLHostname,
		},
		{
			name:       "PTR record uses TTLService (120s) per RFC 6762 §10",
			recordType: protocol.RecordTypePTR,
			wantTTL:    protocol.TTLService,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a record with the appropriate TTL
			record := &RecordTTL{
				RecordType: tt.recordType,
				TTL:        GetTTLForRecordType(tt.recordType),
				CreatedAt:  time.Now(),
			}

			if record.TTL != tt.wantTTL {
				t.Errorf("TTL = %d, want %d for record type %s",
					record.TTL, tt.wantTTL, tt.recordType)
			}
		})
	}
}

// TestTTL_CreatedAtTimestamp_RED tests that records store creation time.
//
// TDD Phase: RED
//
// R004 Decision: Store creation timestamp for TTL calculation
// T017: RecordTTL must track CreatedAt timestamp
func TestTTL_CreatedAtTimestamp(t *testing.T) {
	before := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp precision

	record := NewRecordTTL(protocol.RecordTypeA, protocol.TTLHostname)

	time.Sleep(10 * time.Millisecond)
	after := time.Now()

	if record.CreatedAt.Before(before) {
		t.Errorf("CreatedAt %v is before record creation %v", record.CreatedAt, before)
	}

	if record.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v is after record creation %v", record.CreatedAt, after)
	}
}

// Note: RecordTTL, GetRemainingTTL, IsExpired, GetTTLForRecordType, and NewRecordTTL
// are now implemented in ttl.go (T017 GREEN phase)

// TestGetTTLForRecordType tests RFC 6762 §10 TTL values for all record types.
//
// RFC 6762 §10 specifies different TTLs for different record types:
// - A/AAAA records (hostnames): 4500 seconds (75 minutes)
// - Service records (PTR, SRV, TXT): 120 seconds (2 minutes)
//
// Coverage improvement: GetTTLForRecordType (75% → 100%)
func TestGetTTLForRecordType(t *testing.T) {
	tests := []struct {
		name       string
		recordType protocol.RecordType
		wantTTL    uint32
		rfcNote    string
	}{
		{
			name:       "A record uses TTLHostname (4500s)",
			recordType: protocol.RecordTypeA,
			wantTTL:    protocol.TTLHostname,
			rfcNote:    "RFC 6762 §10: hostname records use 4500s",
		},
		{
			name:       "PTR record uses TTLService (120s)",
			recordType: protocol.RecordTypePTR,
			wantTTL:    protocol.TTLService,
			rfcNote:    "RFC 6762 §10: service discovery records use 120s",
		},
		{
			name:       "SRV record uses TTLService (120s)",
			recordType: protocol.RecordTypeSRV,
			wantTTL:    protocol.TTLService,
			rfcNote:    "RFC 6762 §10: service discovery records use 120s",
		},
		{
			name:       "TXT record uses TTLService (120s)",
			recordType: protocol.RecordTypeTXT,
			wantTTL:    protocol.TTLService,
			rfcNote:    "RFC 6762 §10: service discovery records use 120s",
		},
		{
			name:       "AAAA record (unknown type) defaults to TTLService",
			recordType: protocol.RecordType(28), // AAAA = 28 (not yet defined in protocol)
			wantTTL:    protocol.TTLService,
			rfcNote:    "Default case: unknown types use TTLService",
		},
		{
			name:       "NS record (unknown type) defaults to TTLService",
			recordType: protocol.RecordType(2), // NS = 2 (not defined in protocol)
			wantTTL:    protocol.TTLService,
			rfcNote:    "Default case: unknown types use TTLService",
		},
		{
			name:       "CNAME record (unknown type) defaults to TTLService",
			recordType: protocol.RecordType(5), // CNAME = 5 (not defined in protocol)
			wantTTL:    protocol.TTLService,
			rfcNote:    "Default case: unknown types use TTLService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTTLForRecordType(tt.recordType)

			if got != tt.wantTTL {
				t.Errorf("GetTTLForRecordType(%v) = %d, want %d (%s)",
					tt.recordType, got, tt.wantTTL, tt.rfcNote)
			}
		})
	}
}

// TestGetTTLForRecordType_Values validates the actual TTL constant values.
//
// Ensures protocol constants match RFC 6762 §10 requirements.
//
// Coverage improvement: Validates TTL constant correctness
func TestGetTTLForRecordType_Values(t *testing.T) {
	// RFC 6762 §10: Hostname records use 4500 seconds (75 minutes)
	if protocol.TTLHostname != 4500 {
		t.Errorf("protocol.TTLHostname = %d, want 4500 (RFC 6762 §10: 75 minutes)",
			protocol.TTLHostname)
	}

	// RFC 6762 §10: Service discovery records use 120 seconds (2 minutes)
	if protocol.TTLService != 120 {
		t.Errorf("protocol.TTLService = %d, want 120 (RFC 6762 §10: 2 minutes)",
			protocol.TTLService)
	}
}
