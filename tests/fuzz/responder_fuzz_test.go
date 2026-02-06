// Package fuzz provides fuzz testing for the responder package.
//
// Fuzz testing validates that the responder handles invalid inputs without
// crashes or panics per NFR-003.
package fuzz

import (
	"context"
	"testing"

	"github.com/joshuafuller/beacon/responder"
)

// FuzzServiceRegistration tests service registration with random inputs.
//
// This fuzzer validates that the responder handles malformed service
// definitions without crashes or panics per NFR-003 (Safety & Robustness).
//
// The fuzzer tests:
//   - Valid service definitions (should register successfully)
//   - Empty/nil strings in InstanceName, ServiceType
//   - Out-of-range port numbers (negative, > 65535)
//   - Very long instance names (> 63 octets)
//   - Malformed service types (missing underscores, invalid protocol)
//   - Invalid TXT record sizes (> 1300 bytes)
//   - Special characters and UTF-8 in instance names
//   - Random byte sequences interpreted as strings
//
// Expected behavior:
//   - Valid inputs: Successful registration
//   - Invalid inputs: Return validation error (not panic)
//
// NFR-003: System MUST handle invalid input without crashes or panics
// T118: Fuzz test for service registration with invalid input
//
// Run with: go test -fuzz=FuzzServiceRegistration -fuzztime=10s ./tests/fuzz/
func FuzzServiceRegistration(f *testing.F) {
	// Seed corpus: Valid service
	f.Add("My Service", "_http._tcp.local", 8080, "key1", "value1")

	// Seed corpus: Edge cases
	f.Add("", "_http._tcp.local", 8080, "", "")                                                           // Empty instance name
	f.Add("Service", "", 8080, "", "")                                                                    // Empty service type
	f.Add("Service", "_http._tcp.local", 0, "", "")                                                       // Port = 0
	f.Add("Service", "_http._tcp.local", 65536, "", "")                                                   // Port > 65535
	f.Add("Service", "_http._tcp.local", -1, "", "")                                                      // Port < 0
	f.Add("VeryLongNameThatExceeds63OctetsLimitPerRFC1035Section2.3.4", "_http._tcp.local", 8080, "", "") // Long name
	f.Add("Service", "no-underscore.tcp.local", 8080, "", "")                                             // Missing underscore
	f.Add("Service", "_http._invalid.local", 8080, "", "")                                                // Invalid protocol
	f.Add("Service With Spaces", "_http._tcp.local", 8080, "", "")                                        // Spaces in name (valid per RFC 6763 §4.1)

	f.Fuzz(func(t *testing.T, instanceName, serviceType string, port int, txtKey, txtValue string) {
		// Create responder (each iteration gets its own responder for isolation)
		ctx := context.Background()
		r, err := responder.New(ctx)
		if err != nil {
			// Skip if responder creation fails (shouldn't happen in normal operation)
			t.Skip("Failed to create responder:", err)
		}
		defer func() {
			// Ensure responder is closed even if test panics
			_ = r.Close()
		}()

		// Build TXT records if both key and value are non-empty
		txtRecords := make(map[string]string)
		if txtKey != "" && txtValue != "" {
			txtRecords[txtKey] = txtValue
		}

		// Construct service from fuzz inputs
		// Clamp port to uint16 range (fuzz input is int)
		var portU16 uint16
		if port > 0 && port <= 65535 {
			portU16 = uint16(port)
		}
		svc := &responder.Service{
			InstanceName: instanceName,
			ServiceType:  serviceType,
			Port:         portU16,
			TXTRecords:   txtRecords,
		}

		// Attempt registration - should NEVER panic
		// Valid inputs should succeed, invalid inputs should return error
		err = r.Register(svc)

		// We don't assert on the error result - the goal is to ensure NO PANIC
		// Invalid inputs are expected to return errors, not crash
		_ = err

		// If registration succeeded, attempt to retrieve the service
		if err == nil {
			_, found := r.GetService(instanceName)
			// Service may or may not be found depending on internal state
			// but this should never panic
			_ = found
		}
	})
}

// FuzzServiceUpdate tests service TXT record updates with random inputs.
//
// This fuzzer validates that UpdateService handles malformed TXT records
// without crashes or panics per NFR-003.
//
// The fuzzer tests:
//   - Valid TXT record updates
//   - Very large TXT records (> 1300 bytes)
//   - Empty keys/values
//   - Special characters in keys/values
//   - Non-existent service IDs
//
// Expected behavior:
//   - Valid inputs: Successful update
//   - Invalid inputs: Return error (not panic)
//
// NFR-003: System MUST handle invalid input without crashes or panics
// T118: Fuzz test for service updates with invalid input
//
// Run with: go test -fuzz=FuzzServiceUpdate -fuzztime=10s ./tests/fuzz/
func FuzzServiceUpdate(f *testing.F) {
	// Seed corpus: Valid updates
	f.Add("My Service", "key1", "value1")
	f.Add("", "", "")                                       // Empty everything
	f.Add("Service", "longkey", string(make([]byte, 1400))) // TXT record > 1300 bytes
	f.Add("Service", "=", "=")                              // Special characters

	f.Fuzz(func(t *testing.T, serviceID, txtKey, txtValue string) {
		// Create responder and register a service
		ctx := context.Background()
		r, err := responder.New(ctx)
		if err != nil {
			t.Skip("Failed to create responder:", err)
		}
		defer func() { _ = r.Close() }()

		// Register a known service first
		svc := &responder.Service{
			InstanceName: "Test Service",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		}
		err = r.Register(svc)
		if err != nil {
			t.Skip("Failed to register test service:", err)
		}

		// Build TXT records from fuzz inputs
		txtRecords := make(map[string]string)
		if txtKey != "" {
			txtRecords[txtKey] = txtValue
		}

		// Attempt update with fuzzy serviceID and TXT records - should NEVER panic
		err = r.UpdateService(serviceID, txtRecords)

		// Don't assert on error - goal is to ensure NO PANIC
		// Invalid serviceID or TXT records should return errors, not crash
		_ = err
	})
}

// FuzzServiceUnregister tests service unregistration with random inputs.
//
// This fuzzer validates that Unregister handles malformed service IDs
// without crashes or panics per NFR-003.
//
// The fuzzer tests:
//   - Valid service IDs
//   - Non-existent service IDs
//   - Empty strings
//   - Very long strings
//   - Special characters
//
// Expected behavior:
//   - Valid service ID: Successful unregistration
//   - Invalid service ID: Return error (not panic)
//
// NFR-003: System MUST handle invalid input without crashes or panics
// T118: Fuzz test for unregistration with invalid input
//
// Run with: go test -fuzz=FuzzServiceUnregister -fuzztime=10s ./tests/fuzz/
func FuzzServiceUnregister(f *testing.F) {
	// Seed corpus
	f.Add("My Service")
	f.Add("")                         // Empty
	f.Add("NonExistentService")       // Not registered
	f.Add(string(make([]byte, 1000))) // Very long

	f.Fuzz(func(t *testing.T, serviceID string) {
		// Create responder and register a service
		ctx := context.Background()
		r, err := responder.New(ctx)
		if err != nil {
			t.Skip("Failed to create responder:", err)
		}
		defer func() { _ = r.Close() }()

		// Register a known service
		svc := &responder.Service{
			InstanceName: "Test Service",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		}
		err = r.Register(svc)
		if err != nil {
			t.Skip("Failed to register test service:", err)
		}

		// Attempt unregistration with fuzzy serviceID - should NEVER panic
		err = r.Unregister(serviceID)

		// Don't assert on error - goal is to ensure NO PANIC
		_ = err
	})
}
