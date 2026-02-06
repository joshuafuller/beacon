package responder

import (
	"testing"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// TestConflictDetector_NoConflict verifies that DetectConflict returns false
// when the incoming record has a different name than our record.
//
// RFC 6762 §8.2: "A host is probing for a set of records with the same name"
// - Different names = no conflict
//
// Task: T048 (RED phase - expect FAIL until ConflictDetector implemented)
func TestConflictDetector_NoConflict_DifferentNames(t *testing.T) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // 192.168.1.100
	}

	incomingRecord := message.ResourceRecord{
		Name:  "otherservice.local", // Different name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // 192.168.1.50
	}

	conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
	if err != nil {
		t.Fatalf("DetectConflict() unexpected error: %v", err)
	}

	if conflict {
		t.Errorf("DetectConflict() = true, want false (different names should not conflict)")
	}
}

// TestConflictDetector_ConflictDetected verifies that DetectConflict returns true
// when the incoming record has the same name but different data (we lose tie-break).
//
// RFC 6762 §8.2: "The host with the lexicographically later record wins"
// - Same name, different data = conflict
// - Lexicographic comparison determines winner
//
// Task: T049 (RED phase - expect FAIL)
func TestConflictDetector_ConflictDetected_SameNameDifferentData(t *testing.T) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // 192.168.1.50 (lexicographically earlier)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local", // Same name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // 192.168.1.100 (lexicographically later - they win)
	}

	conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
	if err != nil {
		t.Fatalf("DetectConflict() unexpected error: %v", err)
	}

	if !conflict {
		t.Errorf("DetectConflict() = false, want true (same name + we lose tie-break = conflict)")
	}
}

// TestConflictDetector_TieBreak_WeWin verifies that DetectConflict returns false
// when we win the lexicographic tie-break (our data > their data).
//
// RFC 6762 §8.2: "The host with the lexicographically later record wins"
// - Our data > their data = we win = NO conflict (they defer)
//
// Task: T050 (RED phase - expect FAIL)
func TestConflictDetector_TieBreak_WeWin(t *testing.T) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // 192.168.1.100 (lexicographically later - we win)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local", // Same name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // 192.168.1.50 (lexicographically earlier)
	}

	conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
	if err != nil {
		t.Fatalf("DetectConflict() unexpected error: %v", err)
	}

	if conflict {
		t.Errorf("DetectConflict() = true, want false (we win tie-break, they defer)")
	}
}

// TestConflictDetector_TieBreak_WeLose verifies that DetectConflict returns true
// when we lose the lexicographic tie-break (our data < their data).
//
// RFC 6762 §8.2: "The host with the lexicographically later record wins"
// - Our data < their data = we lose = CONFLICT (we defer)
//
// Task: T051 (RED phase - expect FAIL)
func TestConflictDetector_TieBreak_WeLose(t *testing.T) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // 192.168.1.50 (lexicographically earlier - we lose)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local", // Same name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // 192.168.1.100 (lexicographically later - they win)
	}

	conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
	if err != nil {
		t.Fatalf("DetectConflict() unexpected error: %v", err)
	}

	if !conflict {
		t.Errorf("DetectConflict() = false, want true (we lose tie-break, we must defer)")
	}
}

// TestConflictDetector_LexicographicComparison tests edge cases in lexicographic
// comparison of DNS record data per RFC 6762 §8.2.
//
// RFC 6762 §8.2 comparison algorithm:
// 1. Compare class (excluding cache-flush bit) - numerically greater wins
// 2. Compare type - numerically greater wins
// 3. Compare rdata bytewise (UNSIGNED 0-255) - greater byte wins
// 4. If one runs out of data, longer wins
// 5. If identical, no conflict (fault tolerance)
//
// Edge cases:
// - Identical data (exact tie = no conflict per RFC §8.2)
// - Single byte difference
// - Different lengths
// - Zero bytes in data
// - UNSIGNED byte interpretation (200 > 99, not negative)
//
// Task: T052 (RED phase - expect FAIL)
func TestConflictDetector_LexicographicComparison_EdgeCases(t *testing.T) {
	detector := &ConflictDetector{}

	tests := []struct {
		name           string
		ourData        []byte
		theirData      []byte
		expectConflict bool
		description    string
	}{
		{
			name:           "identical_data",
			ourData:        []byte{192, 168, 1, 100},
			theirData:      []byte{192, 168, 1, 100},
			expectConflict: false, // RFC 6762 §8.2: exact tie = no conflict (fault tolerance)
			description:    "identical data is no conflict (fault tolerance)",
		},
		{
			name:           "single_byte_difference_we_win",
			ourData:        []byte{192, 168, 1, 101},
			theirData:      []byte{192, 168, 1, 100},
			expectConflict: false, // We win (101 > 100)
			description:    "single byte difference, we win",
		},
		{
			name:           "single_byte_difference_we_lose",
			ourData:        []byte{192, 168, 1, 99},
			theirData:      []byte{192, 168, 1, 100},
			expectConflict: true, // We lose (99 < 100)
			description:    "single byte difference, we lose",
		},
		{
			name:           "different_lengths_we_win",
			ourData:        []byte{192, 168, 1, 100, 1}, // Longer
			theirData:      []byte{192, 168, 1, 100},
			expectConflict: false, // We win (longer is lexicographically later)
			description:    "longer data wins (lexicographically later)",
		},
		{
			name:           "different_lengths_we_lose",
			ourData:        []byte{192, 168, 1, 100},
			theirData:      []byte{192, 168, 1, 100, 1}, // Longer
			expectConflict: true,                        // We lose (shorter is lexicographically earlier)
			description:    "shorter data loses (lexicographically earlier)",
		},
		{
			name:           "zero_bytes_we_win",
			ourData:        []byte{0, 0, 0, 1},
			theirData:      []byte{0, 0, 0, 0},
			expectConflict: false, // We win (1 > 0 in last byte)
			description:    "zero bytes handled correctly",
		},
		{
			name:           "rfc_example_169_254_200_50_wins",
			ourData:        []byte{169, 254, 99, 200}, // 169.254.99.200
			theirData:      []byte{169, 254, 200, 50}, // 169.254.200.50
			expectConflict: true,                      // We lose (99 < 200 in 3rd byte) - RFC 6762 §8.2 example
			description:    "RFC 6762 §8.2 example: 169.254.200.50 wins over 169.254.99.200",
		},
		{
			name:           "unsigned_byte_interpretation",
			ourData:        []byte{169, 254, 99, 200}, // 200 is UNSIGNED (not -56)
			theirData:      []byte{169, 254, 99, 100}, // 100 < 200
			expectConflict: false,                     // We win (200 > 100 as UNSIGNED)
			description:    "byte 200 interpreted as UNSIGNED 200, not signed -56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ourRecord := message.ResourceRecord{
				Name:  "myservice.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  tt.ourData,
			}

			incomingRecord := message.ResourceRecord{
				Name:  "myservice.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  tt.theirData,
			}

			conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
			if err != nil {
				t.Fatalf("DetectConflict() unexpected error: %v", err)
			}

			if conflict != tt.expectConflict {
				t.Errorf("DetectConflict() = %v, want %v (%s)", conflict, tt.expectConflict, tt.description)
			}
		})
	}
}

// TestConflictDetector_ClassAndTypeComparison tests that class and type are
// compared before RDATA per RFC 6762 §8.2.
//
// RFC 6762 §8.2: "The determination of 'lexicographically later' is performed
// by first comparing the record class (excluding the cache-flush bit), then
// the record type, then raw comparison of the binary content of the rdata"
//
// Task: T052 extension (RED phase - expect FAIL)
func TestConflictDetector_ClassAndTypeComparison(t *testing.T) {
	detector := &ConflictDetector{}

	tests := []struct {
		name           string
		ourType        protocol.RecordType
		ourClass       protocol.DNSClass
		theirType      protocol.RecordType
		theirClass     protocol.DNSClass
		expectConflict bool
		description    string
	}{
		{
			name:           "same_class_same_type_compare_rdata",
			ourType:        protocol.RecordTypeA,
			ourClass:       protocol.ClassIN,
			theirType:      protocol.RecordTypeA,
			theirClass:     protocol.ClassIN,
			expectConflict: true, // Falls through to RDATA comparison (we lose in test data)
			description:    "same class+type: compare RDATA",
		},
		{
			name:           "same_class_different_type_we_win",
			ourType:        protocol.RecordTypeSRV, // 33 > 16
			ourClass:       protocol.ClassIN,
			theirType:      protocol.RecordTypeTXT, // 16
			theirClass:     protocol.ClassIN,
			expectConflict: false, // We win (33 > 16), regardless of RDATA
			description:    "same class, our type higher: we win",
		},
		{
			name:           "same_class_different_type_we_lose",
			ourType:        protocol.RecordTypeTXT, // 16 < 33
			ourClass:       protocol.ClassIN,
			theirType:      protocol.RecordTypeSRV, // 33
			theirClass:     protocol.ClassIN,
			expectConflict: true, // We lose (16 < 33), regardless of RDATA
			description:    "same class, our type lower: we lose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ourRecord := message.ResourceRecord{
				Name:  "myservice.local",
				Type:  tt.ourType,
				Class: tt.ourClass,
				TTL:   120,
				Data:  []byte{192, 168, 1, 50}, // We lose RDATA comparison (50 < 100)
			}

			incomingRecord := message.ResourceRecord{
				Name:  "myservice.local",
				Type:  tt.theirType,
				Class: tt.theirClass,
				TTL:   120,
				Data:  []byte{192, 168, 1, 100}, // They win RDATA comparison (100 > 50)
			}

			conflict, err := detector.DetectConflict(ourRecord, incomingRecord)
			if err != nil {
				t.Fatalf("DetectConflict() unexpected error: %v", err)
			}

			if conflict != tt.expectConflict {
				t.Errorf("DetectConflict() = %v, want %v (%s)", conflict, tt.expectConflict, tt.description)
			}
		})
	}
}

// TestConflictDetector_ErrorHandling verifies that DetectConflict returns
// appropriate errors for malformed or invalid records.
//
// Error cases:
// - Empty name
// - Nil data
// - Invalid record type
//
// Task: T053 (RED phase - expect FAIL)
func TestConflictDetector_ErrorHandling(t *testing.T) {
	detector := &ConflictDetector{}

	tests := []struct {
		name           string
		ourRecord      message.ResourceRecord
		incomingRecord message.ResourceRecord
		expectError    bool
		description    string
	}{
		{
			name: "empty_name_our_record",
			ourRecord: message.ResourceRecord{
				Name:  "", // Empty name
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 100},
			},
			incomingRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 50},
			},
			expectError: true,
			description: "empty name in our record should return error",
		},
		{
			name: "empty_name_incoming_record",
			ourRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 100},
			},
			incomingRecord: message.ResourceRecord{
				Name:  "", // Empty name
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 50},
			},
			expectError: true,
			description: "empty name in incoming record should return error",
		},
		{
			name: "nil_data_our_record",
			ourRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  nil, // Nil data
			},
			incomingRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 50},
			},
			expectError: true,
			description: "nil data in our record should return error",
		},
		{
			name: "nil_data_incoming_record",
			ourRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  []byte{192, 168, 1, 100},
			},
			incomingRecord: message.ResourceRecord{
				Name:  "valid.local",
				Type:  protocol.RecordTypeA,
				Class: protocol.ClassIN,
				TTL:   120,
				Data:  nil, // Nil data
			},
			expectError: true,
			description: "nil data in incoming record should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := detector.DetectConflict(tt.ourRecord, tt.incomingRecord)

			if tt.expectError && err == nil {
				t.Errorf("DetectConflict() expected error but got nil (%s)", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("DetectConflict() unexpected error: %v (%s)", err, tt.description)
			}
		})
	}
}

// BenchmarkConflictDetector_DetectConflict benchmarks the full conflict detection path.
//
// T063: Benchmark ConflictDetector performance
func BenchmarkConflictDetector_DetectConflict(b *testing.B) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100},
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // We win
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectConflict(ourRecord, incomingRecord)
	}
}

// BenchmarkConflictDetector_LexicographicCompare_ClassDiffers benchmarks class comparison (fastest path).
//
// T063: Benchmark lexicographic comparison - class differs
func BenchmarkConflictDetector_LexicographicCompare_ClassDiffers(b *testing.B) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN, // 0x0001
		TTL:   120,
		Data:  []byte{192, 168, 1, 100},
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: 0x0002, // Different class (rare in practice)
		TTL:   120,
		Data:  []byte{192, 168, 1, 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectConflict(ourRecord, incomingRecord)
	}
}

// BenchmarkConflictDetector_LexicographicCompare_TypeDiffers benchmarks type comparison (second path).
//
// T063: Benchmark lexicographic comparison - type differs
func BenchmarkConflictDetector_LexicographicCompare_TypeDiffers(b *testing.B) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA, // 1
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100},
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeSRV, // 33 (different type)
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{0x00, 0x00, 0x00, 0x00, 0x1f, 0x90}, // Priority=0, Weight=0, Port=8080
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectConflict(ourRecord, incomingRecord)
	}
}

// BenchmarkConflictDetector_LexicographicCompare_RDATACompare benchmarks full RDATA comparison (slowest path).
//
// T063: Benchmark lexicographic comparison - RDATA comparison
func BenchmarkConflictDetector_LexicographicCompare_RDATACompare(b *testing.B) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // Same class, same type, RDATA differs
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // Full RDATA comparison needed
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectConflict(ourRecord, incomingRecord)
	}
}

// BenchmarkConflictDetector_LexicographicCompare_RFCExample benchmarks the RFC 6762 §8.2 example case.
//
// T063: Benchmark RFC 6762 §8.2 example (169.254.200.50 vs 169.254.99.200)
func BenchmarkConflictDetector_LexicographicCompare_RFCExample(b *testing.B) {
	detector := &ConflictDetector{}

	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{169, 254, 99, 200}, // 169.254.99.200 (we lose)
	}

	incomingRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{169, 254, 200, 50}, // 169.254.200.50 (they win)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectConflict(ourRecord, incomingRecord)
	}
}

// ==============================================================================
// Godoc Example (T043) - Documentation & Production Polish
// ==============================================================================

// ExampleConflictDetector demonstrates RFC 6762 §8.2 tie-breaking logic.
//
// This example shows how the ConflictDetector determines whether an incoming
// mDNS probe conflicts with our own probe by performing lexicographic comparison
// of resource records.
//
// RFC 6762 §8.2: "When a host that is probing for a record sees another host
// issue a query for the same record... The two records are compared and the
// lexicographically later data wins."
func ExampleConflictDetector() {
	detector := &ConflictDetector{}

	// Our record - probing for myservice.local with IP 192.168.1.50
	ourRecord := message.ResourceRecord{
		Name:  "myservice.local",
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 50}, // 192.168.1.50
	}

	// Scenario 1: No conflict - different name
	differentName := message.ResourceRecord{
		Name:  "otherservice.local", // Different name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100},
	}

	conflict, _ := detector.DetectConflict(ourRecord, differentName)
	// conflict = false (different names don't conflict)
	_ = conflict

	// Scenario 2: Conflict detected - same name, we lose tie-break
	welose := message.ResourceRecord{
		Name:  "myservice.local", // Same name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 100}, // 192.168.1.100 > 192.168.1.50 (they win)
	}

	conflict, _ = detector.DetectConflict(ourRecord, welose)
	// conflict = true (we must defer and retry with new name)
	_ = conflict

	// Scenario 3: No conflict - same name, we win tie-break
	weWin := message.ResourceRecord{
		Name:  "myservice.local", // Same name
		Type:  protocol.RecordTypeA,
		Class: protocol.ClassIN,
		TTL:   120,
		Data:  []byte{192, 168, 1, 25}, // 192.168.1.25 < 192.168.1.50 (we win)
	}

	conflict, _ = detector.DetectConflict(ourRecord, weWin)
	// conflict = false (they defer to us)
	_ = conflict
}
