package responder

import (
	"testing"
)

// TestService_Validate_RED tests service validation per RFC 6763 §4.
//
// TDD Phase: RED - These tests will FAIL until we implement Service.Validate()
//
// RFC 6763 §4: Service Instance Names
//   - ServiceType format: "_service._proto.domain" (e.g., "_http._tcp.local")
//   - InstanceName: 1-63 octets (RFC 1035 label length)
//   - Port: 1-65535 (uint16 range)
//   - TXT records: total size ≤1300 bytes (RFC 6763 §6.2)
//
// FR-026: System MUST validate service registration parameters
// T021: Write service validation tests
func TestService_Validate_ServiceType(t *testing.T) {
	tests := []struct {
		name        string
		serviceType string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid service type - _http._tcp.local",
			serviceType: "_http._tcp.local",
			wantErr:     false,
		},
		{
			name:        "valid service type - _printer._tcp.local",
			serviceType: "_printer._tcp.local",
			wantErr:     false,
		},
		{
			name:        "invalid - missing leading underscore",
			serviceType: "http._tcp.local",
			wantErr:     true,
			errContains: "invalid service type format",
		},
		{
			name:        "invalid - missing protocol underscore",
			serviceType: "_http.tcp.local",
			wantErr:     true,
			errContains: "invalid service type format",
		},
		{
			name:        "invalid - empty string",
			serviceType: "",
			wantErr:     true,
			errContains: "service type cannot be empty",
		},
		{
			name:        "invalid - invalid protocol (must be _tcp or _udp)",
			serviceType: "_http._sctp.local",
			wantErr:     true,
			errContains: "invalid service type format",
		},
		{
			name:        "invalid - missing domain",
			serviceType: "_http._tcp",
			wantErr:     true,
			errContains: "invalid service type format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: "Test Service",
				ServiceType:  tt.serviceType,
				Port:         8080,
			}

			err := service.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestService_Validate_InstanceName tests instance name validation per RFC 1035 §2.3.4.
//
// TDD Phase: RED
//
// RFC 1035 §2.3.4: Labels are 1-63 octets
// RFC 6763 §4: Instance names should be human-readable
//
// T021: InstanceName length validation
func TestService_Validate_InstanceName(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid - single word",
			instanceName: "MyPrinter",
			wantErr:      false,
		},
		{
			name:         "valid - with spaces",
			instanceName: "My Awesome Printer",
			wantErr:      false,
		},
		{
			name:         "valid - 63 characters (max)",
			instanceName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 63 chars
			wantErr:      false,
		},
		{
			name:         "invalid - empty string",
			instanceName: "",
			wantErr:      true,
			errContains:  "instance name cannot be empty",
		},
		{
			name:         "invalid - 64 characters (exceeds max)",
			instanceName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 64 chars
			wantErr:      true,
			errContains:  "instance name exceeds 63 octets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: tt.instanceName,
				ServiceType:  "_http._tcp.local",
				Port:         8080,
			}

			err := service.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestService_Validate_Port tests port validation.
//
// TDD Phase: RED
//
// RFC 6763 §6: Port must be in valid range (1-65535)
//
// T021: Port range validation
func TestService_Validate_Port(t *testing.T) {
	tests := []struct {
		name        string
		port        uint16
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid - port 80",
			port:    80,
			wantErr: false,
		},
		{
			name:    "valid - port 8080",
			port:    8080,
			wantErr: false,
		},
		{
			name:    "valid - port 65535 (max)",
			port:    65535,
			wantErr: false,
		},
		{
			name:        "invalid - port 0",
			port:        0,
			wantErr:     true,
			errContains: "port must be in range 1-65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: "Test Service",
				ServiceType:  "_http._tcp.local",
				Port:         tt.port,
			}

			err := service.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestService_Validate_TXTRecords tests TXT record size validation per RFC 6763 §6.2.
//
// TDD Phase: RED
//
// RFC 6763 §6.2: Total TXT record size SHOULD NOT exceed 1300 bytes
//
// T021: TXT record size validation
func TestService_Validate_TXTRecords(t *testing.T) {
	tests := []struct {
		name        string
		txtRecords  map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid - small TXT records",
			txtRecords: map[string]string{
				"version": "1.0",
				"path":    "/api",
			},
			wantErr: false,
		},
		{
			name:       "valid - empty TXT (will add 0x00 byte per RFC 6763 §6)",
			txtRecords: map[string]string{},
			wantErr:    false,
		},
		{
			name: "invalid - TXT records exceed 1300 bytes",
			txtRecords: map[string]string{
				"large": string(make([]byte, 1400)), // 1400 bytes > 1300 limit
			},
			wantErr:     true,
			errContains: "TXT records exceed 1300 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: "Test Service",
				ServiceType:  "_http._tcp.local",
				Port:         8080,
				TXTRecords:   tt.txtRecords,
			}

			err := service.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestService_Rename tests the Rename() method per RFC 6762 §9 conflict resolution.
//
// TDD Phase: RED - These tests will FAIL until we implement Service.Rename()
//
// RFC 6762 §9: "If a host receives a response containing a record that conflicts
// with one of its unique records, the host MUST immediately rename the record by
// appending a numeric suffix (starting with '-2') to the instance name."
//
// Examples:
//   - "My Service" → "My Service-2" (first conflict)
//   - "My Service-2" → "My Service-3" (second conflict)
//   - "My Service-10" → "My Service-11" (double-digit)
//
// FR-030: System MUST rename service on conflict (US2)
// T061: Write Service.Rename() tests (RED phase)
func TestService_Rename(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		wantAfter    string
		description  string
	}{
		{
			name:         "first_rename_appends_-2",
			instanceName: "My Service",
			wantAfter:    "My Service-2",
			description:  "RFC 6762 §9: first conflict appends -2",
		},
		{
			name:         "second_rename_increments_to_-3",
			instanceName: "My Service-2",
			wantAfter:    "My Service-3",
			description:  "subsequent conflicts increment suffix",
		},
		{
			name:         "third_rename_increments_to_-4",
			instanceName: "My Service-3",
			wantAfter:    "My Service-4",
			description:  "continue incrementing",
		},
		{
			name:         "tenth_rename_increments_to_-11",
			instanceName: "My Service-10",
			wantAfter:    "My Service-11",
			description:  "double-digit suffixes work",
		},
		{
			name:         "preserves_spaces",
			instanceName: "My Awesome Printer",
			wantAfter:    "My Awesome Printer-2",
			description:  "preserves spaces in name",
		},
		{
			name:         "single_word_name",
			instanceName: "Printer",
			wantAfter:    "Printer-2",
			description:  "works with single word names",
		},
		{
			name:         "name_with_hyphen_not_suffix",
			instanceName: "My-Service",
			wantAfter:    "My-Service-2",
			description:  "hyphen in name (not suffix) doesn't interfere",
		},
		{
			name:         "name_ending_with_number_not_suffix",
			instanceName: "Service2",
			wantAfter:    "Service2-2",
			description:  "number at end (not suffix format) appends -2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: tt.instanceName,
				ServiceType:  "_http._tcp.local",
				Port:         8080,
			}

			service.Rename()

			if service.InstanceName != tt.wantAfter {
				t.Errorf("Rename() InstanceName = %q, want %q (%s)",
					service.InstanceName, tt.wantAfter, tt.description)
			}
		})
	}
}

// TestService_Rename_MaxLength tests that renamed instance names stay within RFC 1035 §2.3.4 limits.
//
// TDD Phase: RED
//
// RFC 1035 §2.3.4: Labels are 1-63 octets
// RFC 6762 §9: Renaming must respect label length limits
//
// T061: Rename() respects 63-octet limit
func TestService_Rename_MaxLength(t *testing.T) {
	tests := []struct {
		name         string
		instanceName string
		wantLen      int
		description  string
	}{
		{
			name:         "short_name_no_truncation",
			instanceName: "MyService",
			wantLen:      12, // "MyService-2" = 11 chars
			description:  "short names don't need truncation",
		},
		{
			name: "long_name_truncation",
			// 60 characters - renaming to -2 would make 63 (max), -10 would make 64 (over limit)
			instanceName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 60 chars
			wantLen:      63,
			description:  "long names truncated to fit suffix within 63 octets",
		},
		{
			name: "max_length_name",
			// 63 characters - must truncate before adding suffix
			instanceName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 63 chars
			wantLen:      63,
			description:  "max length names truncated before suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				InstanceName: tt.instanceName,
				ServiceType:  "_http._tcp.local",
				Port:         8080,
			}

			service.Rename()

			if len(service.InstanceName) > 63 {
				t.Errorf("Rename() InstanceName length = %d, want ≤63 (RFC 1035 §2.3.4 violation)",
					len(service.InstanceName))
			}

			// For this RED phase, we just check the length constraint
			// GREEN phase will implement proper truncation logic
		})
	}
}

// Helper function for substring checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
