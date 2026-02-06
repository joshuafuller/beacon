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
