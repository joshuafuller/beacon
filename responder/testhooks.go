package responder

import "fmt"

// This file contains test-only hooks on the Responder.
//
// They exist solely to enable black-box contract tests (in package
// tests/contract and responder's own *_test.go) to observe and inject protocol
// behavior — captured probe/announce messages, probe/announce callbacks, fast
// registration without the ~1.75s probing delay, and conflict injection — without
// reaching into internal packages or weakening the F-2 layer boundary.
//
// They are NOT part of the responder's runtime behavior. If the contract tests
// are ever moved white-box (into package responder with an export_test.go), this
// surface can be removed from the public API entirely. See B2 in the refactor
// punch-list for that deliberate, deferred decision.

// OnProbe sets a callback to be called when a probe is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (r *Responder) OnProbe(callback func()) {
	// Store callback for future machines
	r.onProbeCallback = callback

	// Also apply to current machine if it exists
	if r.lastMachine != nil {
		prober := r.lastMachine.GetProber()
		if prober != nil {
			prober.SetOnSendQuery(callback)
		}
	}
}

// OnAnnounce sets a callback to be called when an announcement is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (r *Responder) OnAnnounce(callback func()) {
	// Store callback for future machines
	r.onAnnounceCallback = callback

	// Also apply to current machine if it exists
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			announcer.SetOnSendAnnouncement(callback)
		}
	}
}

// GetLastProbeMessage returns the last sent probe message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (r *Responder) GetLastProbeMessage() []byte {
	if r.lastMachine != nil {
		prober := r.lastMachine.GetProber()
		if prober != nil {
			return prober.GetLastProbeMessage()
		}
	}
	return nil
}

// GetLastAnnounceMessage returns the last sent announcement message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (r *Responder) GetLastAnnounceMessage() []byte {
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			return announcer.GetLastAnnounceMessage()
		}
	}
	return nil
}

// GetLastAnnouncedRecords returns the last announced record set.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 and RFC 6763 §6 validation
func (r *Responder) GetLastAnnouncedRecords() []*ResourceRecord {
	return r.lastAnnouncedRecords
}

// GetLastAnnounceDest returns the last announcement destination address.
//
// US2 GREEN: Contract test support for RFC 6762 §5 multicast address validation
func (r *Responder) GetLastAnnounceDest() string {
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			return announcer.GetLastDestAddr()
		}
	}
	return ""
}

// RegisterServiceWithoutProbing is a test hook that registers a service directly
// in the registry without performing probing or announcing.
//
// This is intended ONLY for contract tests that need to test query handling
// without the ~1.75s overhead of the full Register() flow.
//
// Parameters:
//   - service: The service to register
//
// Returns:
//   - error: validation or registry error
func (r *Responder) RegisterServiceWithoutProbing(service *Service) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}
	if err := service.Validate(); err != nil {
		return err
	}
	return r.registry.Register(toInternalService(service))
}

// InjectConflictDuringProbing is a test hook to inject conflicts during probing.
//
// When enabled, the state machine will always report StateConflictDetected,
// forcing the rename loop to trigger.
//
// T062: Test hook for max rename attempts testing
func (r *Responder) InjectConflictDuringProbing(inject bool) {
	r.injectConflict = inject
}

// InjectSimultaneousProbe is a test hook for injecting simultaneous probe scenarios.
//
// This method is currently a stub placeholder for future simultaneous probe testing
// per RFC 6762 §8.2 tiebreaking. It will be implemented when detailed conflict
// resolution testing is added.
//
// Parameters:
//   - First parameter: Our probe packet (currently unused)
//   - Second parameter: Incoming probe packet (currently unused)
//
// T062: Test hook infrastructure for conflict scenarios
func (r *Responder) InjectSimultaneousProbe([]byte, []byte) {}
