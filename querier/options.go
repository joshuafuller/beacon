package querier

import (
	"net"
	"time"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Option is a functional option for configuring a Querier.
//
// Options follow the functional options pattern per F-5 (API Usability).
// This allows flexible, backwards-compatible configuration.
//
// Example:
//
//	q, err := querier.New(
//	    querier.WithTimeout(2 * time.Second),
//	)
type Option func(*Querier) error

// WithTimeout sets the default timeout for mDNS queries.
//
// The timeout determines how long to wait for responses before returning.
// Per FR-006, timeout is configurable with a reasonable default.
//
// Default: 1 second (per SC-002: discover 95% of devices within 1 second)
//
// Example:
//
//	q, err := querier.New(querier.WithTimeout(2 * time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(q *Querier) error {
		q.defaultTimeout = timeout
		return nil
	}
}

// WithInterfaces configures the Querier to use only the specified interfaces.
// This overrides the default interface selection logic (which excludes VPN/Docker).
//
// Use this when you need explicit control over which interfaces send mDNS queries.
//
// FR-011: System MUST provide WithInterfaces([]net.Interface) functional option
//
// Example:
//
//	ifaces, _ := net.Interfaces()
//	eth0 := ifaces[0]
//	q, _ := querier.New(querier.WithInterfaces([]net.Interface{eth0}))
//
// If the provided list is empty, the default interface selection logic is used
// (same as not calling WithInterfaces at all).
//
// FR-011 Compliance: Empty list is accepted and treated as "use all interfaces"
func WithInterfaces(ifaces []net.Interface) Option {
	return func(q *Querier) error {
		// FR-011: Accept empty list - use default interface selection
		// This allows callers to conditionally pass interfaces without special casing
		if len(ifaces) == 0 {
			// Don't set explicitInterfaces - use default selection
			return nil
		}

		// Set explicit interface list (overrides filter)
		q.explicitInterfaces = ifaces
		return nil
	}
}

// WithInterfaceFilter configures the Querier with a custom interface selection filter.
// The filter function is called for each available interface; return true to include.
//
// This option is ignored if WithInterfaces() is also specified (explicit list takes priority).
//
// FR-012: System MUST provide WithInterfaceFilter(func(net.Interface) bool) functional option
//
// Example (allow only Ethernet interfaces):
//
//	q, _ := querier.New(querier.WithInterfaceFilter(func(iface net.Interface) bool {
//	    return strings.HasPrefix(iface.Name, "eth")
//	}))
//
// Default behavior (if neither WithInterfaces nor WithInterfaceFilter is specified):
//   - Excludes VPN (utun*, tun*, ppp*, wg*, tailscale*, wireguard*)
//   - Excludes Docker (docker0, veth*, br-*)
//   - Excludes loopback, down, and non-multicast interfaces
//
// FR-012 Compliance: Nil filter is accepted and treated as "accept all interfaces"
func WithInterfaceFilter(filter func(net.Interface) bool) Option {
	return func(q *Querier) error {
		// FR-012: Accept nil filter - use default interface selection
		// This allows callers to conditionally set filters without special casing
		if filter == nil {
			// Don't set custom filter - use default selection
			return nil
		}

		// Set custom filter (will be ignored if explicit list is set)
		q.interfaceFilter = filter
		return nil
	}
}

// WithRateLimit enables or disables rate limiting.
// Rate limiting protects against multicast storms by tracking per-source-IP query rates.
//
// FR-033: System MUST provide WithRateLimit(bool) option to enable/disable rate limiting
//
// Default: Enabled (true)
//
// When enabled, sources exceeding the threshold (default: 100 qps) are rate-limited
// for a cooldown period (default: 60 seconds).
//
// Example (disable rate limiting for testing):
//
//	q, _ := querier.New(querier.WithRateLimit(false))
func WithRateLimit(enabled bool) Option {
	return func(q *Querier) error {
		q.rateLimitEnabled = enabled
		return nil
	}
}

// WithRateLimitThreshold sets the query rate threshold (queries per second per source IP).
// Sources exceeding this threshold are rate-limited for the cooldown period.
//
// FR-027: Rate limiter MUST have configurable threshold (default: 100 queries/second)
//
// Default: 100 queries/second (balances storm protection with legitimate high-volume use)
//
// This option is ignored if rate limiting is disabled via WithRateLimit(false).
//
// Example (stricter threshold for untrusted networks):
//
//	q, _ := querier.New(querier.WithRateLimitThreshold(50))
func WithRateLimitThreshold(threshold int) Option {
	return func(q *Querier) error {
		if threshold <= 0 {
			return &errors.ValidationError{
				Field:   "rateLimitThreshold",
				Value:   threshold,
				Message: "threshold must be greater than 0",
			}
		}

		q.rateLimitThreshold = threshold
		return nil
	}
}

// WithRateLimitCooldown sets the duration to drop packets from a source after
// it exceeds the rate limit threshold.
//
// FR-028: Rate limiter MUST have configurable cooldown duration (default: 60 seconds)
//
// Default: 60 seconds (long enough to detect persistent storms, short enough to recover)
//
// This option is ignored if rate limiting is disabled via WithRateLimit(false).
//
// Example (shorter cooldown for transient storms):
//
//	q, _ := querier.New(querier.WithRateLimitCooldown(30 * time.Second))
func WithRateLimitCooldown(cooldown time.Duration) Option {
	return func(q *Querier) error {
		if cooldown <= 0 {
			return &errors.ValidationError{
				Field:   "rateLimitCooldown",
				Value:   cooldown,
				Message: "cooldown must be greater than 0",
			}
		}

		q.rateLimitCooldown = cooldown
		return nil
	}
}

// WithIPv6 enables dual-stack mDNS operation per RFC 6762 §20.
//
// When enabled, the Querier creates a second UDP transport bound to the IPv6
// mDNS multicast group (FF02::FB:5353) and sends queries on both IPv4 and
// IPv6 networks. AAAA records are accepted and surfaced in responses.
//
// If the host has no IPv6-capable multicast interface, New() gracefully falls
// back to IPv4-only operation rather than returning an error.
//
// Example:
//
//	q, err := querier.New(querier.WithIPv6())
func WithIPv6() Option {
	return func(q *Querier) error {
		q.ipv6Enabled = true
		return nil
	}
}

// WithTransport overrides the default IPv4 transport (normally a real UDP
// multicast socket bound to 224.0.0.251:5353). New() skips creating that
// socket when this option is supplied.
//
// This exists for test isolation (T100): injecting transport.NewMockTransport()
// lets tests drive Querier's aggregation/dedup/DiscoverServices logic via
// fixture packets without a real socket racing against live mDNS traffic on
// the host's network.
//
// Example:
//
//	mock := transport.NewMockTransport()
//	q, _ := New(WithTransport(mock))
//
// See: specs/004-m1-1-architectural-hardening/tasks.md Phase 8, T100
func WithTransport(t transport.Transport) Option {
	return func(q *Querier) error {
		q.transport = t
		return nil
	}
}

// WithIPv6Transport overrides the default IPv6 transport and implies WithIPv6
// (dual-stack is meaningless without it). New() skips creating the real
// UDPv6Transport when this option is supplied.
//
// Companion to WithTransport (T100) for tests that need to exercise dual-stack
// Query() behavior (e.g. verifying both transports are sent to) without
// opening real IPv4 and IPv6 multicast sockets.
//
// Example:
//
//	q, _ := New(WithTransport(mockV4), WithIPv6Transport(mockV6))
func WithIPv6Transport(t transport.Transport) Option {
	return func(q *Querier) error {
		q.transport6 = t
		q.ipv6Enabled = true
		return nil
	}
}
