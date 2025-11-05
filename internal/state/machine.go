// Package state implements the mDNS service registration state machine per RFC 6762 §8.
//
// WHY THIS PACKAGE EXISTS:
// RFC 6762 §8 mandates a specific sequence before a service can be considered "established":
// 1. Probing phase (~750ms): Send 3 probe queries to detect naming conflicts
// 2. Announcing phase (~1s): Send 2 unsolicited announcements to inform the network
// 3. Established: Service is now discoverable by other mDNS clients
//
// This package enforces RFC-compliant timing, handles conflicts, and coordinates state transitions.
//
// DESIGN RATIONALE:
// - Goroutine-per-service: Each service registration runs independently in its own goroutine (R001)
// - Context-aware: All operations respect context cancellation per F-9
// - Testable: State machine is decoupled from transport for unit testing
//
// RFC COMPLIANCE:
// - RFC 6762 §8.1: Probing (3 probes, 250ms apart)
// - RFC 6762 §8.2: Conflict detection via simultaneous probe tie-breaking
// - RFC 6762 §8.3: Announcing (2 announcements, 1s apart)
// - RFC 6762 §9: Conflict resolution via service renaming
//
// PRIMARY TECHNICAL AUTHORITY: RFC 6762 §§8-9 (Probing, Announcing, Conflict Resolution)
package state

import (
	"context"
	"sync"
)

// Machine coordinates the service registration state machine per RFC 6762 §8.
//
// WHY: RFC 6762 §8 requires a multi-phase registration process to prevent naming conflicts
// and ensure all network participants are aware of the new service.
//
// OPERATION:
// The machine orchestrates three phases:
//
//  1. Probing Phase (RFC 6762 §8.1):
//     - Duration: ~750ms (3 probes × 250ms intervals)
//     - Purpose: Detect if another device is already using this name
//     - Action: Send probe queries (type ANY) and listen for responses
//     - Outcome: Either no conflict (proceed to announcing) or conflict detected (stop)
//
//  2. Announcing Phase (RFC 6762 §8.3):
//     - Duration: ~1s (2 announcements × 1s intervals)
//     - Purpose: Inform network that we're claiming this name
//     - Action: Send unsolicited multicast responses with all records (PTR, SRV, TXT, A)
//     - Outcome: Service is now established and discoverable
//
//  3. Established State:
//     - Service is fully registered and responding to queries
//     - Responder handles incoming queries via query handler goroutine
//
// CONFLICT HANDLING:
// If a conflict is detected during probing (RFC 6762 §8.2):
//   - Machine transitions to ConflictDetected state
//   - Caller (Responder.Register) renames service per RFC 6762 §9 (append "-2", "-3", etc.)
//   - Probing restarts with new name (max 10 rename attempts per FR-032)
//
// State flow:
//
//	Initial → Probing → Announcing → Established
//	Probing → ConflictDetected (if conflict)
//
// THREAD SAFETY:
// - Machine uses sync.RWMutex to protect state reads/writes
// - State transitions notify test hooks WITHOUT holding lock to prevent deadlocks
//
// State flow:
//
//	Initial → Probing → Announcing → Established
//	Probing → ConflictDetected (if conflict)
//
// R001 Decision: Goroutine-per-service architecture
// T037: Implement Machine struct
type Machine struct {
	prober         *Prober
	announcer      *Announcer
	mu             sync.RWMutex
	onStateChange  func(State)
	currentState   State
	injectConflict bool
}

// NewMachine creates a new state machine.
//
// T037: Initialize Machine
func NewMachine() *Machine {
	return &Machine{
		currentState: StateInitial,
		prober:       NewProber(),
		announcer:    NewAnnouncer(),
	}
}

// Run executes the state machine for a service.
//
// RFC 6762 §8: State machine execution
//  1. Probing: Send 3 probes, 250ms apart (~750ms)
//  2. Announcing: Send 2 announcements, 1s apart (~1s)
//  3. Established: Service fully registered
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceName: Full service name
//
// Returns:
//   - error: Context error if canceled, nil on success
//
// R001: Each service runs in its own goroutine
// T038: Implement Machine.run() with context cancellation
func (sm *Machine) Run(ctx context.Context, serviceName string) error {
	// Transition to Probing
	sm.setState(StateProbing)

	// Phase 1: Probing (~750ms)
	result := sm.prober.Probe(ctx, serviceName)
	if result.Error != nil {
		return result.Error
	}

	if result.Conflict || sm.injectConflict {
		// Conflict detected - stop here
		// Caller (Responder) will handle rename/retry
		sm.setState(StateConflictDetected)
		return nil
	}

	// Transition to Announcing
	sm.setState(StateAnnouncing)

	// Phase 2: Announcing (~1s)
	// Note: Records are built by Responder.Register() via BuildRecordSet()
	// Machine doesn't need records directly - Announcer uses test hooks for unit testing
	// Actual transport integration with records happens in US3 (Response to Queries)
	records := []byte{} // Placeholder for announcer interface
	err := sm.announcer.Announce(ctx, serviceName, records)
	if err != nil {
		return err
	}

	// Transition to Established
	sm.setState(StateEstablished)

	return nil
}

// GetState returns the current state.
//
// T038: State querying
func (sm *Machine) GetState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// setState transitions to a new state.
//
// T038: State transitions with callbacks
func (sm *Machine) setState(newState State) {
	// Manual unlock required: Must release lock before calling user callback to avoid deadlocks.
	// Callback may access state machine, so holding lock would cause deadlock.
	sm.mu.Lock() // nosemgrep: beacon-mutex-defer-unlock
	sm.currentState = newState
	sm.mu.Unlock()

	// Notify test hook (called WITHOUT lock to prevent deadlocks)
	if sm.onStateChange != nil {
		sm.onStateChange(newState)
	}
}

// SetInjectConflict is a test hook to inject conflict during probing.
//
// T062: Test hook for max rename attempts testing
func (sm *Machine) SetInjectConflict(inject bool) {
	sm.injectConflict = inject
}

// GetProber returns the internal Prober for integration with Responder.
//
// US2 GREEN: Allow Responder to access Prober for message capture
func (sm *Machine) GetProber() *Prober {
	return sm.prober
}

// GetAnnouncer returns the internal Announcer for integration with Responder.
//
// US2 GREEN: Allow Responder to access Announcer for message capture
func (sm *Machine) GetAnnouncer() *Announcer {
	return sm.announcer
}
