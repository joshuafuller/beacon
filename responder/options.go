package responder

import (
	"github.com/joshuafuller/beacon/internal/security"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Option is a functional option for configuring a Responder.
//
// This pattern allows flexible configuration without breaking API compatibility.
//
// T044: Implement functional options pattern
type Option func(*Responder) error

// WithTransport sets a custom transport for the responder (primarily for testing).
//
// If not provided, a production UDPv4Transport will be created.
//
// Parameters:
//   - t: Custom transport implementation (e.g., MockTransport for testing)
//
// Returns:
//   - Option: Configuration function
//
// Example:
//
//	mockTransport := transport.NewMockTransport()
//	r, err := New(ctx, WithTransport(mockTransport))
//
// 007-interface-specific-addressing: Added to support contract testing
func WithTransport(t transport.Transport) Option {
	return func(r *Responder) error {
		r.transport = t
		return nil
	}
}

// WithRateLimiter sets a custom rate limiter for the responder.
//
// If not provided, a default rate limiter is created with:
//   - 100 queries/second per source IP
//   - 60 second cooldown after threshold exceeded
//   - 10,000 max tracked source IPs
//
// Parameters:
//   - rl: Custom rate limiter configuration
//
// Returns:
//   - Option: Configuration function
//
// FR-026: Per-source-IP rate limiting
func WithRateLimiter(rl *security.RateLimiter) Option {
	return func(r *Responder) error {
		r.rateLimiter = rl
		return nil
	}
}

// WithIPv6 enables dual-stack operation by adding an IPv6 multicast transport.
//
// RFC 6762 §11: IPv6 mDNS uses the link-local multicast address FF02::FB on port 5353.
//
// If IPv6 is unavailable on this host, this option degrades gracefully and the
// responder continues in IPv4-only mode.
func WithIPv6() Option {
	return func(r *Responder) error {
		tr6, err := transport.NewUDPv6Transport()
		if err != nil {
			// Graceful degradation: IPv6 unavailable, continue IPv4-only.
			return nil
		}
		r.transport6 = tr6
		return nil
	}
}

// WithHostname sets a custom hostname for the responder.
//
// If not provided, the system hostname will be used.
//
// Parameters:
//   - hostname: Custom hostname (e.g., "myhost.local")
//
// Returns:
//   - Option: Configuration function
//
// Example:
//
//	r, err := New(ctx, WithHostname("mydevice.local"))
//
// T044: WithHostname option
func WithHostname(hostname string) Option {
	return func(r *Responder) error {
		r.hostname = hostname
		return nil
	}
}
