package responder

import (
	"testing"
)

// TestConflictDetector_DetectConflict_RED tests conflict detection.
//
// TDD Phase: RED - Will fail until we implement ConflictDetector
//
// RFC 6762 §8.1: Conflicting responses during probing
// T019: Implement ConflictDetector with lexicographic tie-breaking
func TestConflictDetector_DetectConflict(t *testing.T) {
	detector := NewConflictDetector()

	// Service we're trying to register
	ourService := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Conflicting service (same name)
	conflictingService := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         9090, // Different port, but same name = conflict
	}

	hasConflict := detector.DetectConflict(ourService, conflictingService)
	if !hasConflict {
		t.Error("DetectConflict() = false, want true (same InstanceName)")
	}
}

// TestConflictDetector_NoConflict_RED tests non-conflicting services.
//
// TDD Phase: RED
//
// T019: Different instance names should not conflict
func TestConflictDetector_NoConflict(t *testing.T) {
	detector := NewConflictDetector()

	ourService := &Service{
		InstanceName: "My Printer",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	otherService := &Service{
		InstanceName: "Your Printer", // Different name
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	hasConflict := detector.DetectConflict(ourService, otherService)
	if hasConflict {
		t.Error("DetectConflict() = true, want false (different InstanceName)")
	}
}

// TestConflictDetector_CompareProbes_RED tests simultaneous probe tie-breaking.
//
// TDD Phase: RED
//
// RFC 6762 §8.2.1: Lexicographic comparison of probe records
// R002 Decision: Use bytes.Compare for lexicographic tie-breaking
// T019: Implement tie-breaking logic
func TestConflictDetector_CompareProbes(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name      string
		ourData   []byte
		theirData []byte
		wantWin   bool // true if we win, false if they win
	}{
		{
			name:      "we win - lexicographically later",
			ourData:   []byte{169, 254, 200, 50}, // 169.254.200.50
			theirData: []byte{169, 254, 99, 200}, // 169.254.99.200
			wantWin:   true,                      // 200 > 99 in third byte
		},
		{
			name:      "they win - lexicographically earlier",
			ourData:   []byte{169, 254, 99, 200},
			theirData: []byte{169, 254, 200, 50},
			wantWin:   false, // 99 < 200 in third byte
		},
		{
			name:      "identical data - no conflict",
			ourData:   []byte{192, 168, 1, 100},
			theirData: []byte{192, 168, 1, 100},
			wantWin:   false, // Identical = no conflict, both can proceed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWin := detector.CompareProbes(tt.ourData, tt.theirData)
			if gotWin != tt.wantWin {
				t.Errorf("CompareProbes() = %v, want %v", gotWin, tt.wantWin)
			}
		})
	}
}

// TestConflictDetector_Rename_RED tests automatic renaming on conflict.
//
// TDD Phase: RED
//
// RFC 6762 §8.2: Rename service on conflict (MyApp → MyApp (2))
// T019: Implement renaming with suffix increment
func TestConflictDetector_Rename(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name         string
		instanceName string
		wantRenamed  string
	}{
		{
			name:         "first rename - add (2)",
			instanceName: "My Printer",
			wantRenamed:  "My Printer (2)",
		},
		{
			name:         "second rename - increment to (3)",
			instanceName: "My Printer (2)",
			wantRenamed:  "My Printer (3)",
		},
		{
			name:         "tenth rename",
			instanceName: "My Printer (9)",
			wantRenamed:  "My Printer (10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRenamed := detector.Rename(tt.instanceName)
			if gotRenamed != tt.wantRenamed {
				t.Errorf("Rename(%q) = %q, want %q", tt.instanceName, gotRenamed, tt.wantRenamed)
			}
		})
	}
}

// TestConflictDetector_MultipleRecords_RED tests tie-breaking with multiple records.
//
// TDD Phase: RED
//
// RFC 6762 §8.2.1: Compare multiple records pairwise
// T019: Handle service with multiple records (SRV + TXT)
func TestConflictDetector_MultipleRecords(t *testing.T) {
	detector := NewConflictDetector()

	// Our service records (SRV + TXT)
	ourRecords := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x1f, 0x90}, // SRV: priority=0, weight=0, port=8080
		{0x00},                               // TXT: empty
	}

	// Their service records (SRV + TXT)
	theirRecords := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x1f, 0x91}, // SRV: priority=0, weight=0, port=8081
		{0x00},                               // TXT: empty
	}

	// Compare: our port 8080 (0x1f90) < their port 8081 (0x1f91), they win
	gotWin := detector.CompareMultipleRecords(ourRecords, theirRecords)
	if gotWin {
		t.Error("CompareMultipleRecords() = true, want false (they have higher port)")
	}
}

// TestConflictDetector_CompareMultipleRecords_EdgeCases tests edge cases in multi-record comparison.
//
// Covers additional scenarios beyond the basic test to reach higher coverage.
func TestConflictDetector_CompareMultipleRecords_EdgeCases(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name         string
		ourRecords   [][]byte
		theirRecords [][]byte
		wantWin      bool
		description  string
	}{
		{
			name:         "we have more records - we win",
			ourRecords:   [][]byte{{0x01}, {0x02}, {0x03}},
			theirRecords: [][]byte{{0x01}, {0x02}},
			wantWin:      true,
			description:  "RFC 6762 §8.2.1: longer list wins when all compared records match",
		},
		{
			name:         "they have more records - they win",
			ourRecords:   [][]byte{{0x01}, {0x02}},
			theirRecords: [][]byte{{0x01}, {0x02}, {0x03}},
			wantWin:      false,
			description:  "RFC 6762 §8.2.1: longer list wins when all compared records match",
		},
		{
			name:         "identical records - they win (equal length)",
			ourRecords:   [][]byte{{0x01}, {0x02}},
			theirRecords: [][]byte{{0x01}, {0x02}},
			wantWin:      false,
			description:  "Equal length and identical records - no winner",
		},
		{
			name:         "we win on first record",
			ourRecords:   [][]byte{{0x02}, {0x01}},
			theirRecords: [][]byte{{0x01}, {0x02}},
			wantWin:      true,
			description:  "First record comparison determines winner",
		},
		{
			name:         "they win on second record",
			ourRecords:   [][]byte{{0x01}, {0x01}},
			theirRecords: [][]byte{{0x01}, {0x02}},
			wantWin:      false,
			description:  "First records match, second record comparison determines winner",
		},
		{
			name:         "empty lists - no winner",
			ourRecords:   [][]byte{},
			theirRecords: [][]byte{},
			wantWin:      false,
			description:  "Empty lists are considered equal",
		},
		{
			name:         "we have records, they have empty - we win",
			ourRecords:   [][]byte{{0x01}},
			theirRecords: [][]byte{},
			wantWin:      true,
			description:  "Non-empty list beats empty list",
		},
		{
			name:         "they have records, we have empty - they win",
			ourRecords:   [][]byte{},
			theirRecords: [][]byte{{0x01}},
			wantWin:      false,
			description:  "Empty list loses to non-empty list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWin := detector.CompareMultipleRecords(tt.ourRecords, tt.theirRecords)
			if gotWin != tt.wantWin {
				t.Errorf("CompareMultipleRecords() = %v, want %v (%s)",
					gotWin, tt.wantWin, tt.description)
			}
		})
	}
}
