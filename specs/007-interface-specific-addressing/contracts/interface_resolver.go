// Package contracts defines the API contracts for interface-specific addressing.
package contracts

import "net"

// InterfaceResolver provides interface → IPv4 address lookup.
//
// This is a NEW internal API for RFC 6762 §15 compliance.
//
// Location: responder/responder.go (private helper function)
// Version: 1.0 (007-interface-specific-addressing)
type InterfaceResolver interface {
	// GetIPv4ForInterface returns the IPv4 address assigned to the specified interface.
	//
	// RFC 6762 §15: Responses MUST include addresses valid on the receiving interface.
	//
	// Parameters:
	//   - ifIndex: Network interface index (from Transport.Receive or ipv4.ControlMessage)
	//
	// Returns:
	//   - []byte: IPv4 address (4 bytes) in network byte order
	//   - error: NetworkError if interface not found, ValidationError if no IPv4
	//
	// Algorithm:
	//   1. net.InterfaceByIndex(ifIndex) → net.Interface
	//   2. iface.Addrs() → []net.Addr
	//   3. Filter for first IPv4 address (ipnet.IP.To4() != nil)
	//   4. Return IPv4 bytes or error
	//
	// Edge Cases:
	//   - Interface not found → NetworkError ("interface index N not found")
	//   - Interface has no IPv4 → ValidationError ("no IPv4 address found on interface")
	//   - Interface has multiple IPs → returns first IPv4 (consistent behavior)
	//
	// Performance:
	//   - ~250ns per call (InterfaceByIndex + Addrs lookup)
	//   - No caching in fast-track fix (deferred to M4)
	//
	// Example:
	//   ipv4, err := getIPv4ForInterface(2)  // Look up interface index 2
	//   if err != nil {
	//       return err  // Interface not found or no IPv4
	//   }
	//   // ipv4 == []byte{192, 168, 1, 100}
	GetIPv4ForInterface(ifIndex int) ([]byte, error)
}

// Concrete Implementation (not interface - just function)
//
// func getIPv4ForInterface(ifIndex int) ([]byte, error) {
//     iface, err := net.InterfaceByIndex(ifIndex)
//     if err != nil {
//         return nil, &errors.NetworkError{
//             Operation: "lookup interface",
//             Err:       err,
//             Details:   fmt.Sprintf("interface index %d not found", ifIndex),
//         }
//     }
//
//     addrs, err := iface.Addrs()
//     if err != nil {
//         return nil, &errors.NetworkError{
//             Operation: "get interface addresses",
//             Err:       err,
//             Details:   fmt.Sprintf("failed to get addresses for %s", iface.Name),
//         }
//     }
//
//     for _, addr := range addrs {
//         if ipnet, ok := addr.(*net.IPNet); ok {
//             if ipv4 := ipnet.IP.To4(); ipv4 != nil {
//                 return ipv4, nil
//             }
//         }
//     }
//
//     return nil, &errors.ValidationError{
//         Field:   "interface",
//         Value:   iface.Name,
//         Message: "no IPv4 address found on interface",
//     }
// }

// FallbackBehavior defines how the responder handles interface resolution failures.
type FallbackBehavior int

const (
	// SkipResponse - Do not send response if interface lookup fails (strict RFC compliance)
	// Recommended: This ensures we never violate RFC 6762 §15 MUST NOT requirement
	SkipResponse FallbackBehavior = iota

	// UseGlobalIP - Fall back to getLocalIPv4() if interface lookup fails (best-effort)
	// Use case: Platform doesn't support control messages, or interface removed
	// Tradeoff: May violate RFC 6762 §15 on multi-interface hosts (same as current bug)
	UseGlobalIP
)

// InterfaceResolutionPolicy defines how responder handles interface context.
//
// This is a configuration option for future extensibility (not implemented in fast-track).
type InterfaceResolutionPolicy struct {
	// FallbackOnError defines behavior when getIPv4ForInterface fails
	FallbackOnError FallbackBehavior

	// LogFailures enables warning logs for interface resolution errors
	LogFailures bool

	// CacheEnabled enables interface → IP caching (M4 feature)
	// Fast-track fix: Always false
	CacheEnabled bool

	// CacheTTL defines how long to cache interface IPs before refresh
	// Fast-track fix: Ignored (no caching)
	CacheTTL int // seconds
}

// DefaultPolicy returns the recommended interface resolution policy.
//
// Fast-track defaults:
//   - SkipResponse on error (strict RFC compliance)
//   - LogFailures enabled (debugging)
//   - No caching (deferred to M4)
func DefaultPolicy() InterfaceResolutionPolicy {
	return InterfaceResolutionPolicy{
		FallbackOnError: SkipResponse,
		LogFailures:     true,
		CacheEnabled:    false,
		CacheTTL:        0,
	}
}

// ValidationRules for IPv4 addresses in responses
//
// RFC 6762 §15 Requirements:
//   - MUST include addresses valid on receiving interface
//   - MUST NOT include addresses from other interfaces
//
// Additional Validation:
//   - IPv4 address must be 4 bytes
//   - IPv4 address must not be loopback (127.0.0.0/8)
//   - IPv4 address must not be unspecified (0.0.0.0)
func IsValidIPv4ForResponse(ip net.IP) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false // Not IPv4
	}

	if ipv4.IsLoopback() {
		return false // 127.0.0.0/8
	}

	if ipv4.IsUnspecified() {
		return false // 0.0.0.0
	}

	return true
}
