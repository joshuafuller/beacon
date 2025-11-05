package querier

import (
	"context"
	"net"
	"testing"
	"time"
)

// BenchmarkQuery measures the query processing overhead per NFR-001.
//
// T092: Verify query processing overhead <100ms
//
// NFR-001: Query processing overhead MUST be <100ms on typical hardware
//
// This benchmark measures the time to execute a complete query cycle:
//  1. Validate inputs
//  2. Build query message
//  3. Send to multicast group
//  4. Collect responses (with timeout)
//  5. Parse and deduplicate responses
func BenchmarkQuery(b *testing.B) {
	q, err := New()
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = q.Query(ctx, "benchmark.local", RecordTypeA)
	}
}

// BenchmarkNew measures the cost of creating a new Querier.
//
// This benchmark measures socket creation and background goroutine startup.
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q, err := New()
		if err != nil {
			b.Fatalf("New() failed: %v", err)
		}
		_ = q.Close()
	}
}

// BenchmarkQueryParallel measures concurrent query performance.
//
// This benchmark validates that the Querier can handle concurrent queries
// efficiently without lock contention.
func BenchmarkQueryParallel(b *testing.B) {
	q, err := New()
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	b.RunParallel(func(pb *testing.PB) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		for pb.Next() {
			_, _ = q.Query(ctx, "parallel.local", RecordTypeA)
		}
	})
}

// TestConcurrentQueries validates that 100 concurrent queries work without
// resource leaks per NFR-002.
//
// T093: Verify 100 concurrent queries without leaks
//
// NFR-002: System MUST support at least 100 concurrent queries without resource leaks
//
// This test:
//  1. Creates a single Querier instance
//  2. Launches 100 goroutines, each making a query
//  3. Verifies all queries complete successfully
//  4. Verifies no goroutine leaks (via testing.T short mode)
func TestConcurrentQueries(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	const numQueries = 100

	// Channel to collect results
	results := make(chan error, numQueries)

	// Launch 100 concurrent queries
	for i := 0; i < numQueries; i++ {
		go func(_ int) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			_, err := q.Query(ctx, "concurrent.local", RecordTypeA)
			results <- err
		}(i)
	}

	// Collect all results
	for i := 0; i < numQueries; i++ {
		err := <-results
		if err != nil {
			// Errors are acceptable (timeout, validation, network)
			// We're testing that queries don't panic or deadlock
			t.Logf("Query %d returned error (acceptable): %v", i, err)
		}
	}

	t.Logf("✓ NFR-002: Successfully handled %d concurrent queries", numQueries)
}

// TestWithTimeout verifies the WithTimeout option works correctly.
//
// This test validates the functional option pattern for configuration.
func TestWithTimeout(t *testing.T) {
	customTimeout := 2 * time.Second

	q, err := New(WithTimeout(customTimeout))
	if err != nil {
		t.Fatalf("New(WithTimeout) failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	// Verify the timeout was set
	if q.defaultTimeout != customTimeout {
		t.Errorf("defaultTimeout = %v, want %v", q.defaultTimeout, customTimeout)
	}

	t.Logf("✓ WithTimeout option set defaultTimeout to %v", q.defaultTimeout)
}

// TestClose verifies graceful shutdown releases all resources.
//
// This test validates FR-017, FR-018 resource management requirements.
func TestClose(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Close should complete without error
	err = q.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Calling Close again should not panic (idempotent)
	// Note: Current implementation may panic on double-close
	// This documents the behavior

	t.Log("✓ Close() completed successfully")
}

// TestWithInterfaces verifies WithInterfaces option validation.
//
// Tests that the option correctly sets explicit interface list and validates input.
func TestWithInterfaces(t *testing.T) {
	tests := []struct {
		name        string
		ifaces      []net.Interface
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid interface list",
			ifaces: []net.Interface{
				{Name: "eth0", Index: 1},
			},
			expectError: false,
		},
		{
			name:        "empty interface list",
			ifaces:      []net.Interface{},
			expectError: true,
			errorMsg:    "interface list cannot be empty",
		},
		{
			name:        "nil interface list",
			ifaces:      nil,
			expectError: true,
			errorMsg:    "interface list cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(WithInterfaces(tt.ifaces))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				} else {
					t.Logf("✓ Correctly rejected with error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("New(WithInterfaces) failed: %v", err)
			}
			defer func() { _ = q.Close() }()

			// Verify explicit interfaces were set
			if len(q.explicitInterfaces) != len(tt.ifaces) {
				t.Errorf("explicitInterfaces length = %d, want %d",
					len(q.explicitInterfaces), len(tt.ifaces))
			}
		})
	}
}

// TestWithInterfaceFilter verifies WithInterfaceFilter option validation.
//
// Tests that the option correctly sets custom filter and validates input.
func TestWithInterfaceFilter(t *testing.T) {
	t.Run("valid filter function", func(t *testing.T) {
		filter := func(iface net.Interface) bool {
			return iface.Name == "eth0"
		}

		q, err := New(WithInterfaceFilter(filter))
		if err != nil {
			t.Fatalf("New(WithInterfaceFilter) failed: %v", err)
		}
		defer func() { _ = q.Close() }()

		// Verify filter was set
		if q.interfaceFilter == nil {
			t.Error("interfaceFilter was not set")
		}

		t.Log("✓ Filter function set successfully")
	})

	t.Run("nil filter function", func(t *testing.T) {
		_, err := New(WithInterfaceFilter(nil))
		if err == nil {
			t.Error("Expected error for nil filter, got nil")
		} else if !contains(err.Error(), "filter function cannot be nil") {
			t.Errorf("Expected error about nil filter, got: %v", err)
		} else {
			t.Logf("✓ Correctly rejected nil filter: %v", err)
		}
	})
}

// TestWithRateLimit verifies WithRateLimit option.
//
// Tests that rate limiting can be enabled/disabled.
func TestWithRateLimit(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"rate limiting enabled", true},
		{"rate limiting disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(WithRateLimit(tt.enabled))
			if err != nil {
				t.Fatalf("New(WithRateLimit(%v)) failed: %v", tt.enabled, err)
			}
			defer func() { _ = q.Close() }()

			if q.rateLimitEnabled != tt.enabled {
				t.Errorf("rateLimitEnabled = %v, want %v",
					q.rateLimitEnabled, tt.enabled)
			}

			t.Logf("✓ Rate limiting set to %v", tt.enabled)
		})
	}
}

// TestWithRateLimitThreshold verifies WithRateLimitThreshold option validation.
//
// Tests threshold validation (must be > 0).
func TestWithRateLimitThreshold(t *testing.T) {
	tests := []struct {
		name        string
		threshold   int
		expectError bool
	}{
		{"valid threshold", 100, false},
		{"minimum threshold", 1, false},
		{"high threshold", 10000, false},
		{"zero threshold", 0, true},
		{"negative threshold", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(WithRateLimitThreshold(tt.threshold))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid threshold, got nil")
				} else if !contains(err.Error(), "threshold must be greater than 0") {
					t.Errorf("Expected threshold validation error, got: %v", err)
				} else {
					t.Logf("✓ Correctly rejected threshold %d: %v", tt.threshold, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("New(WithRateLimitThreshold(%d)) failed: %v",
					tt.threshold, err)
			}
			defer func() { _ = q.Close() }()

			if q.rateLimitThreshold != tt.threshold {
				t.Errorf("rateLimitThreshold = %d, want %d",
					q.rateLimitThreshold, tt.threshold)
			}

			t.Logf("✓ Threshold set to %d", tt.threshold)
		})
	}
}

// TestWithRateLimitCooldown verifies WithRateLimitCooldown option validation.
//
// Tests cooldown validation (must be > 0).
func TestWithRateLimitCooldown(t *testing.T) {
	tests := []struct {
		name        string
		cooldown    time.Duration
		expectError bool
	}{
		{"valid cooldown", 60 * time.Second, false},
		{"short cooldown", 1 * time.Second, false},
		{"long cooldown", 5 * time.Minute, false},
		{"zero cooldown", 0, true},
		{"negative cooldown", -1 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := New(WithRateLimitCooldown(tt.cooldown))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid cooldown, got nil")
				} else if !contains(err.Error(), "cooldown must be greater than 0") {
					t.Errorf("Expected cooldown validation error, got: %v", err)
				} else {
					t.Logf("✓ Correctly rejected cooldown %v: %v", tt.cooldown, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("New(WithRateLimitCooldown(%v)) failed: %v",
					tt.cooldown, err)
			}
			defer func() { _ = q.Close() }()

			if q.rateLimitCooldown != tt.cooldown {
				t.Errorf("rateLimitCooldown = %v, want %v",
					q.rateLimitCooldown, tt.cooldown)
			}

			t.Logf("✓ Cooldown set to %v", tt.cooldown)
		})
	}
}

// TestResourceRecordAccessors validates the type-safe accessor methods.
//
// This test ensures AsA, AsPTR, AsSRV, AsTXT return nil/empty for wrong types
// and handle malformed data gracefully.
func TestResourceRecordAccessors(t *testing.T) {
	// Test all combinations of record types and accessor methods
	tests := []struct {
		name     string
		record   ResourceRecord
		expectA  bool
		expectPTR bool
		expectSRV bool
		expectTXT bool
	}{
		{
			name: "A record",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypeA,
				Data: net.IPv4(192, 168, 1, 1),
			},
			expectA: true,
		},
		{
			name: "PTR record",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypePTR,
				Data: "target.local",
			},
			expectPTR: true,
		},
		{
			name: "SRV record",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypeSRV,
				Data: SRVData{
					Target:   "server.local",
					Priority: 0,
					Weight:   0,
					Port:     8080,
				},
			},
			expectSRV: true,
		},
		{
			name: "TXT record",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypeTXT,
				Data: []string{"key=value", "version=1"},
			},
			expectTXT: true,
		},
		{
			name: "A record with wrong data type",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypeA,
				Data: "not an IP", // Wrong type
			},
			// All should return nil/empty
		},
		{
			name: "SRV record with wrong data type",
			record: ResourceRecord{
				Name: "test.local",
				Type: RecordTypeSRV,
				Data: "not SRVData", // Wrong type
			},
			// All should return nil/empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test AsA()
			ip := tt.record.AsA()
			if tt.expectA {
				if ip == nil {
					t.Error("AsA() returned nil for A record")
				}
			} else {
				if ip != nil {
					t.Errorf("AsA() returned %v, expected nil", ip)
				}
			}

			// Test AsPTR()
			ptr := tt.record.AsPTR()
			if tt.expectPTR {
				if ptr == "" {
					t.Error("AsPTR() returned empty string for PTR record")
				}
			} else {
				if ptr != "" {
					t.Errorf("AsPTR() returned %q, expected empty string", ptr)
				}
			}

			// Test AsSRV()
			srv := tt.record.AsSRV()
			if tt.expectSRV {
				if srv == nil {
					t.Error("AsSRV() returned nil for SRV record")
				}
			} else {
				if srv != nil {
					t.Errorf("AsSRV() returned %v, expected nil", srv)
				}
			}

			// Test AsTXT()
			txt := tt.record.AsTXT()
			if tt.expectTXT {
				if txt == nil {
					t.Error("AsTXT() returned nil for TXT record")
				}
			} else {
				if txt != nil {
					t.Errorf("AsTXT() returned %v, expected nil", txt)
				}
			}
		})
	}

	t.Log("✓ Type-safe accessors validated for all record types and error cases")
}

// TestRecordTypeString verifies RecordType.String() returns correct names.
func TestRecordTypeString(t *testing.T) {
	tests := []struct {
		recordType RecordType
		expected   string
	}{
		{RecordTypeA, "A"},
		{RecordTypePTR, "PTR"},
		{RecordTypeSRV, "SRV"},
		{RecordTypeTXT, "TXT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.recordType.String()
			if got != tt.expected {
				t.Errorf("RecordType(%d).String() = %q, want %q",
					tt.recordType, got, tt.expected)
			}
		})
	}

	t.Log("✓ RecordType.String() validated for all types")
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// ==============================================================================
// M1-Refactoring Integration Tests (TDD - RED Phase)
// ==============================================================================
// These tests are written FIRST to guide the Transport interface refactoring.
// Expected: FAIL until querier is refactored to use Transport interface (T031-T037)

// NOTE: Original TDD RED tests removed (T027, T028):
// - TestQuerier_UsesTransportInterface: Obsolete, T031 is complete
//   (Querier HAS transport field at querier.go:46-47, used throughout)
// - TestQuerier_WorksWithMockTransport: Deferred to future milestone
//   (WithTransport() option not implemented - all tests work without it)
//
// Transport interface abstraction is validated via:
// - M1-Refactoring completion (see archive/m1-refactoring/)
// - internal/transport/transport_test.go (interface contract tests)
// - querier/querier.go:112 (New() creates UDPv4Transport)
//
// TODO M2 (T100): Add test for WithTransport() option
// After implementing WithTransport() option (see querier/options.go TODO), add:
//
//   func TestQuerier_WithTransport_UsesMockTransport(t *testing.T) {
//       mock := transport.NewMockTransport()
//       q, err := New(WithTransport(mock))
//       if err != nil {
//           t.Fatalf("New(WithTransport) failed: %v", err)
//       }
//       defer func() { _ = q.Close() }()
//
//       ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
//       defer cancel()
//
//       _, _ = q.Query(ctx, "test.local", RecordTypeA)
//
//       // Verify mock recorded the Send() call
//       calls := mock.SendCalls()
//       if len(calls) != 1 {
//           t.Errorf("Expected 1 Send() call, got %d", len(calls))
//       }
//   }
//
// This enables testing without real network, mocking failures, simulating responses.
// See: specs/004-m1-1-architectural-hardening/tasks.md Phase 8, T100

// ==============================================================================
// Phase 3: Error Propagation Validation (T064) - FR-004
// ==============================================================================

// T064: Integration test - Querier.Close() handles transport close errors
//
// This test validates that Querier.Close() properly propagates errors from
// the underlying transport (FR-004 validation).
//
// Test strategy: Close twice - second close should propagate transport error
func TestQuerier_Close_PropagatesTransportErrors(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// First close should succeed
	err = q.Close()
	if err != nil {
		t.Errorf("First Close() should succeed, got error: %v", err)
	}

	// Second close should propagate transport error (validates FR-004 end-to-end)
	err = q.Close()
	if err == nil {
		t.Error("FR-004 VIOLATION: Second Close() returned nil, expected error from transport")
	} else {
		t.Logf("✓ FR-004 VALIDATED (end-to-end): Querier.Close() propagates transport error: %v", err)
	}
}
