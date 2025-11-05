// Package responder implements mDNS service registration and response per RFC 6762.
package responder

import (
	"fmt"
	"regexp"
	"strconv"
)

// Service represents an mDNS service to be registered per RFC 6763.
//
// RFC 6763 §4: Service Instance Names
// RFC 6763 §6: Text Record Semantics
//
// Service encapsulates all information needed to advertise a network service via mDNS.
// The service name follows DNS-SD conventions with three components:
//   1. InstanceName: Human-readable label (e.g., "My Printer")
//   2. ServiceType: Protocol identifier (e.g., "_ipp._tcp")
//   3. Domain: Always ".local" for mDNS
//
// Full DNS name: "My Printer._ipp._tcp.local"
//
// Wire Format (DNS-SD PTR/SRV/TXT Records):
//
//	PTR Record:  _ipp._tcp.local. → My Printer._ipp._tcp.local.
//	SRV Record:  My Printer._ipp._tcp.local. → hostname.local:631
//	TXT Record:  My Printer._ipp._tcp.local. → ["version=1.0", "path=/"]
//	A Record:    hostname.local. → 192.168.1.100
//
// Functional Requirements:
//   - FR-026: System MUST validate service registration parameters
//   - US-1: Service registration with probing/announcing
//   - T031: Service struct design
//
// Validation Constraints (per RFC 6763 §4.3, RFC 1035 §2.3.4):
//   - InstanceName: 1-63 octets (UTF-8)
//   - ServiceType: Must match "_service._proto.local" pattern
//   - Port: 1-65535
//   - TXTRecords: Total size ≤1300 bytes (recommended ≤200 bytes)
//
// Example:
//
//	service := &responder.Service{
//	    InstanceName: "My Web Server",
//	    ServiceType:  "_http._tcp.local",
//	    Port:         8080,
//	    TXTRecords:   map[string]string{
//	        "version": "1.0",
//	        "path":    "/api",
//	    },
//	    Hostname:     "server.local",  // Optional - defaults to system hostname
//	}
//
//	if err := service.Validate(); err != nil {
//	    log.Fatalf("Invalid service: %v", err)
//	}
type Service struct {
	// InstanceName is the human-readable service instance name (e.g., "My Printer").
	// RFC 1035 §2.3.4: Labels are 1-63 octets.
	InstanceName string

	// ServiceType is the service type (e.g., "_http._tcp.local").
	// Format: "_service._proto.local" where proto is _tcp or _udp.
	ServiceType string

	// Port is the service port number (1-65535).
	Port int

	// TXTRecords contains optional service metadata as key-value pairs.
	// RFC 6763 §6.2: Total size SHOULD NOT exceed 1300 bytes.
	// RFC 6763 §6: If empty, a single TXT record with 0x00 byte MUST be created.
	TXTRecords map[string]string

	// Hostname is the hostname for the A/AAAA record (optional).
	// If not provided, system hostname will be used.
	Hostname string
}

// Validate validates the service fields per RFC 6762/6763 requirements.
//
// RFC 6763 §4: Service Instance Names
// RFC 1035 §2.3.4: Label length limits
// RFC 6763 §6.2: TXT record size limits
//
// Returns:
//   - error: ValidationError if any field is invalid, nil if valid
//
// FR-026: System MUST validate service registration parameters
// T032: Implement Service.Validate()
func (s *Service) Validate() error {
	// Validate InstanceName
	if s.InstanceName == "" {
		return fmt.Errorf("instance name cannot be empty")
	}

	// RFC 1035 §2.3.4: Labels are 1-63 octets
	if len(s.InstanceName) > 63 {
		return fmt.Errorf("instance name exceeds 63 octets (got %d)", len(s.InstanceName))
	}

	// Validate ServiceType format
	if err := validateServiceType(s.ServiceType); err != nil {
		return err
	}

	// Validate Port
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("port must be in range 1-65535 (got %d)", s.Port)
	}

	// Validate TXT records size
	if err := validateTXTRecordsSize(s.TXTRecords); err != nil {
		return err
	}

	return nil
}

// Rename renames the service by appending or incrementing a numeric suffix per RFC 6762 §9.
//
// RFC 6762 §9: "If a host receives a response containing a record that conflicts
// with one of its unique records, the host MUST immediately rename the record by
// appending a numeric suffix (starting with '-2') to the instance name."
//
// Renaming algorithm:
//  1. If name has no suffix (e.g., "My Service") → append "-2"
//  2. If name has suffix (e.g., "My Service-2") → increment to "-3"
//  3. Truncate if needed to stay within 63-octet limit (RFC 1035 §2.3.4)
//
// Examples:
//   - "My Service" → "My Service-2"
//   - "My Service-2" → "My Service-3"
//   - "My Service-10" → "My Service-11"
//
// FR-030: System MUST rename service on conflict (US2)
// T061: Implement Service.Rename() (GREEN phase)
func (s *Service) Rename() {
	// Pattern: matches "-N" suffix at end of string where N is a positive integer
	// E.g., "My Service-2", "Printer-10"
	suffixPattern := regexp.MustCompile(`^(.+)-(\d+)$`)

	if matches := suffixPattern.FindStringSubmatch(s.InstanceName); matches != nil {
		// Name already has a suffix - increment it
		baseName := matches[1]  // "My Service"
		suffixStr := matches[2] // "2"

		// Parse existing suffix (guaranteed to be valid digits by regex)
		// Error is impossible because regex ensures suffixStr contains only digits
		suffix, _ := strconv.Atoi(suffixStr) // nosemgrep: beacon-error-swallowing
		suffix++                             // Increment: 2 → 3

		// Reconstruct name with incremented suffix
		newName := fmt.Sprintf("%s-%d", baseName, suffix)

		// Truncate if needed to fit within 63-octet limit
		s.InstanceName = truncateToFit(newName, 63)
	} else {
		// Name has no suffix - append "-2"
		newName := s.InstanceName + "-2"

		// Truncate if needed to fit within 63-octet limit
		s.InstanceName = truncateToFit(newName, 63)
	}
}

// truncateToFit truncates a name to fit within maxLen octets while preserving suffix.
//
// RFC 1035 §2.3.4: Labels are 1-63 octets
// RFC 6762 §9: Renaming must respect label length limits
//
// Algorithm:
//  1. If name fits within maxLen, return as-is
//  2. If name is too long, truncate the base name (not the suffix)
//
// Examples:
//   - truncateToFit("Short-2", 63) → "Short-2" (no change)
//   - truncateToFit("VeryLongNameThatExceedsLimit...-2", 63) → "VeryLongNameThatExceedsLi-2" (truncated)
//
// T061: Truncation logic for 63-octet limit
func truncateToFit(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name // Fits within limit
	}

	// Name is too long - need to truncate
	// Strategy: Preserve the suffix (e.g., "-2"), truncate the base name

	// Find the suffix
	suffixPattern := regexp.MustCompile(`^(.+?)(-\d+)$`)
	if matches := suffixPattern.FindStringSubmatch(name); matches != nil {
		baseName := matches[1] // "VeryLongNameThatExceedsLimit..."
		suffix := matches[2]   // "-2"

		// Calculate how much space we have for the base name
		maxBaseLen := maxLen - len(suffix)

		if maxBaseLen < 1 {
			// Edge case: suffix itself is too long (shouldn't happen in practice)
			// Just truncate the whole thing
			return name[:maxLen]
		}

		// Truncate base name and append suffix
		return baseName[:maxBaseLen] + suffix
	}

	// No suffix found (shouldn't happen in Rename() flow, but handle it)
	return name[:maxLen]
}

// serviceTypeRegex matches valid service type patterns per RFC 6763 §4.
// Format: _service._proto.local where service is alphanumeric+hyphens, proto is _tcp or _udp
var serviceTypeRegex = regexp.MustCompile(`^_[a-z0-9-]+\._(tcp|udp)\.local$`)

// validateServiceType validates the service type format per RFC 6763 §4.
//
// Format: "_service._proto.local"
// Example: "_http._tcp.local"
//
// Requirements:
//   - Must start with underscore "_"
//   - Protocol must be _tcp or _udp
//   - Must end with ".local" for mDNS
//
// T032: ServiceType validation
func validateServiceType(serviceType string) error {
	if serviceType == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	// Use regex for robust validation
	if !serviceTypeRegex.MatchString(serviceType) {
		return fmt.Errorf("invalid service type format (must be _service._proto.local, e.g., \"_http._tcp.local\")")
	}

	return nil
}

// validateTXTRecordsSize validates that TXT records don't exceed RFC limits.
//
// RFC 6763 §6.2: "The total size of a typical DNS-SD TXT record is intended to be
// small -- 200 bytes or less. In cases where more data is justified, the maximum
// SHOULD NOT exceed 1300 bytes."
//
// T032: TXT record size validation
func validateTXTRecordsSize(txtRecords map[string]string) error {
	if len(txtRecords) == 0 {
		// Empty TXT is valid - will create mandatory 0x00 byte per RFC 6763 §6
		return nil
	}

	// Calculate total size: length byte + key=value for each pair
	totalSize := 0
	for key, value := range txtRecords {
		// Each entry: length byte + "key=value"
		entrySize := 1 + len(key) + 1 + len(value) // 1 for '=', 1 for length prefix
		totalSize += entrySize
	}

	// RFC 6763 §6.2: SHOULD NOT exceed 1300 bytes
	if totalSize > 1300 {
		return fmt.Errorf("TXT records exceed 1300 bytes (got %d)", totalSize)
	}

	return nil
}
