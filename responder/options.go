package responder

import (
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
