package responder

// Option is a functional option for configuring a Responder.
//
// RFC 6762 §5: mDNS Message Format and Configuration
//
// Options follow the functional options pattern (per F-5: API Usability) to provide
// flexible, backwards-compatible configuration for the responder. This allows adding
// new configuration parameters without breaking existing code.
//
// All options are applied during New() initialization before starting the background
// query handler goroutine.
//
// Functional Requirements:
//   - FR-202: Flexible responder configuration
//   - T044: Functional options pattern
//
// Example:
//
//	resp, err := responder.New(ctx,
//	    responder.WithHostname("mydevice.local"),
//	)
type Option func(*Responder) error

// WithHostname sets a custom hostname for the responder's A/AAAA records.
//
// RFC 6762 §6.1: Resource Records and A/AAAA Records
//
// The hostname is used to construct A/AAAA records that map the service hostname
// to IP addresses. If not provided via this option, the system hostname (from
// os.Hostname()) is used automatically with ".local" appended.
//
// The hostname should end with ".local" for mDNS compliance. If the provided hostname
// doesn't end with ".local", it should be adjusted by the caller.
//
// Parameters:
//   - hostname: Custom hostname for A/AAAA records (e.g., "mydevice.local")
//
// Returns:
//   - Option: Configuration function that sets the hostname
//
// Functional Requirements:
//   - FR-202: Configurable hostname for responder
//   - T044: WithHostname option implementation
//
// Example:
//
//	// Use custom hostname for all services
//	resp, err := responder.New(ctx, responder.WithHostname("server.local"))
//	if err != nil {
//	    return err
//	}
//
//	// Service will be advertised with A record: server.local → 192.168.1.100
//	service := &responder.Service{
//	    InstanceName: "My Service",
//	    ServiceType:  "_http._tcp.local",
//	    Port:         8080,
//	}
//	resp.Register(service)
func WithHostname(hostname string) Option {
	return func(r *Responder) error {
		r.hostname = hostname
		return nil
	}
}
