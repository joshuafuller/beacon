# Beacon Documentation Standards

**Version**: 1.0
**Last Updated**: 2025-11-05
**Exemplar**: `internal/state/machine.go`

This document defines documentation standards for the Beacon mDNS library to ensure RFC traceability, maintainability, and auditability.

---

## Table of Contents

1. [Philosophy](#philosophy)
2. [Package-Level Documentation](#package-level-documentation)
3. [Type Documentation](#type-documentation)
4. [Function Documentation](#function-documentation)
5. [Constant Documentation](#constant-documentation)
6. [Inline Comments](#inline-comments)
7. [Traceability](#traceability)
8. [Examples](#examples)
9. [Review Checklist](#review-checklist)

---

## Philosophy

### Core Principles

1. **RFC Traceability**: Every protocol decision must link to specific RFC section
2. **Explain WHY, Not WHAT**: Code shows what; comments explain why
3. **Architectural Rationale**: Document decisions with references to ADRs
4. **Task Linkage**: Connect code to implementation tasks for auditability
5. **Godoc Compliance**: All documentation must render correctly in godoc

### Audience

- **External Users**: Package-level docs, public APIs (via godoc)
- **Internal Maintainers**: Implementation rationale, RFC compliance
- **Auditors**: Traceability to specs, ADRs, and RFCs

---

## Package-Level Documentation

Every package must have comprehensive package-level documentation explaining its purpose and scope.

### Template

```go
// Package <name> <one-line purpose>.
//
// ## WHY THIS PACKAGE EXISTS
//
// <2-3 sentences explaining the problem this package solves and its role
// in the overall architecture. Reference relevant user stories or requirements.>
//
// ## PRIMARY TECHNICAL AUTHORITY
//
// - RFC 6762 §X: <Section title and brief description>
// - RFC 6763 §Y: <Section title and brief description>
// - ADR-NNN: <Decision title>
//
// ## DESIGN RATIONALE
//
// <Explain key architectural decisions, tradeoffs, and why this approach
// was chosen. Reference ADRs for major decisions.>
//
// ## RFC COMPLIANCE
//
// This package implements the following RFC requirements:
//
// - RFC 6762 §X.Y: <Requirement description>
// - RFC 6763 §Z: <Requirement description>
//
// ## KEY CONCEPTS
//
// <Define domain-specific terminology and concepts used in this package.
// This helps developers understand the domain model.>
//
// ## EXAMPLE USAGE
//
// <Provide a simple, compilable example showing typical usage.>
package name
```

### Example

```go
// Package responder implements mDNS responder functionality for service registration
// and query response per RFC 6762.
//
// ## WHY THIS PACKAGE EXISTS
//
// Applications need to advertise network services (HTTP servers, printers, etc.)
// so they can be discovered by other devices on the local network without
// centralized DNS servers. This package provides the RFC 6762 compliant responder
// implementation for Beacon's mDNS library.
//
// ## PRIMARY TECHNICAL AUTHORITY
//
// - RFC 6762 §5-§14: Multicast DNS protocol specification
// - RFC 6763 §4-§7: DNS-SD service types and naming
// - ADR-005: State machine architecture for probing/announcing
//
// ## DESIGN RATIONALE
//
// The responder uses a goroutine-per-service architecture (ADR-005) to isolate
// state machines and simplify concurrency. Services progress through states
// (Initial → Probing → Announcing → Established) independently, allowing
// concurrent registration without complex locking.
//
// ## RFC COMPLIANCE
//
// This package implements the following RFC requirements:
//
// - RFC 6762 §8.1: Probing for name conflicts (3 probes, 250ms apart)
// - RFC 6762 §8.2: Simultaneous probe tiebreaking (lexicographic comparison)
// - RFC 6762 §8.3: Announcing registered services (2 announcements, 1s apart)
// - RFC 6762 §10: Resource record TTL values (75 minutes default, 10s goodbye)
//
// ## KEY CONCEPTS
//
// - Service: An instance of a network service with name, type, port, and metadata
// - Probing: Conflict detection phase before announcing (RFC 6762 §8.1)
// - Announcing: Broadcasting service availability after successful probing
// - Goodbye: TTL=0 packet sent during graceful shutdown
//
// ## EXAMPLE USAGE
//
//	resp, err := responder.New(responder.WithInterface("en0"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer resp.Close()
//
//	service := responder.Service{
//	    InstanceName: "MyApp",
//	    ServiceType:  "_http._tcp.local.",
//	    Port:         8080,
//	    TXT:          []string{"version=1.0", "path=/"},
//	}
//
//	if err := resp.Register(context.Background(), service); err != nil {
//	    log.Fatal(err)
//	}
package responder
```

---

## Type Documentation

Every exported type must have documentation explaining its purpose, RFC relationship, and usage.

### Template

```go
// <TypeName> <one-line purpose>.
//
// RFC 6762 §X: <Which RFC section this type implements or relates to>
//
// <2-3 sentences explaining what this type represents in the protocol,
// its role in the system, and any important constraints or invariants.>
//
// Wire format (for protocol types):
//
//	0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                   FIELD1                      |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// Functional Requirements:
//   - FR-XXX: <Requirement this type helps satisfy>
//   - US-Y: <User story this supports>
//
// Example:
//
//	<Short, compilable example showing typical usage>
type TypeName struct {
    // Field1 <purpose and RFC reference if applicable>
    Field1 string

    // Field2 <purpose, constraints, and rationale>
    Field2 int
}
```

### Example

```go
// Machine coordinates the service registration state machine per RFC 6762 §8.
//
// RFC 6762 §8: Probing and Announcing on Startup
//
// Each service progresses through states independently to detect name conflicts
// (probing) and broadcast availability (announcing). The machine orchestrates
// transitions and provides cancellation support via context.
//
// State flow:
//
//	Initial → Probing → Announcing → Established
//	Probing → ConflictDetected (if name conflict detected)
//
// Functional Requirements:
//   - FR-201: Service registration with conflict detection
//   - FR-202: Graceful cancellation support
//   - US-1: Service registration
//   - US-2: Name conflict resolution
//
// Design Decision:
//   - R001: Goroutine-per-service architecture for isolation and simplicity
//   - ADR-005: State machine pattern for clear lifecycle management
//
// Example:
//
//	machine := state.NewMachine()
//	err := machine.Run(ctx, "MyApp._http._tcp.local.")
//	if err != nil {
//	    return err
//	}
//	if machine.GetState() == state.StateConflictDetected {
//	    // Handle rename
//	}
type Machine struct {
    // prober handles the RFC 6762 §8.1 probing phase (3 probes, 250ms apart)
    prober *Prober

    // announcer handles the RFC 6762 §8.3 announcing phase (2 announcements, 1s apart)
    announcer *Announcer

    // mu protects currentState from concurrent access
    mu sync.RWMutex

    // onStateChange is called after state transitions (test hook)
    onStateChange func(State)

    // currentState tracks the current position in the state machine
    currentState State

    // injectConflict is a test hook to simulate conflict detection
    injectConflict bool
}
```

---

## Function Documentation

Every exported function must document its purpose, RFC compliance, algorithm, parameters, and return values.

### Template

```go
// <FunctionName> <one-line purpose>.
//
// RFC 6762 §X.Y: <Which RFC requirement this function implements>
//
// <2-4 sentences explaining what this function does, why it exists,
// and any important algorithm details or RFC compliance notes.>
//
// Algorithm (for complex functions):
//  1. <Step 1 with RFC reference if applicable>
//  2. <Step 2>
//  3. <Step 3>
//
// RFC Quote (if critical for understanding):
//   "Direct quote from RFC explaining the requirement"
//   (RFC 6762 §X.Y, paragraph N)
//
// Parameters:
//   - param1: <Description, constraints, and valid values>
//   - param2: <Description>
//
// Returns:
//   - result: <Description of return value>
//   - error: <Error conditions and types returned>
//
// Functional Requirements:
//   - FR-XXX: <Requirement satisfied>
//   - Task T-YYY: <Implementation task>
//
// Example:
//
//	<Short, compilable usage example if the function is complex>
func FunctionName(param1 string, param2 int) (result, error) {
    // Implementation
}
```

### Example

```go
// Run executes the state machine for a service registration.
//
// RFC 6762 §8: Probing and Announcing on Startup
//
// This function orchestrates the two-phase registration process: probing to
// detect name conflicts, then announcing to advertise the service. The entire
// process takes approximately 1.75 seconds (~750ms probing + ~1s announcing).
//
// Algorithm:
//  1. Transition to Probing state
//  2. Send 3 probe queries spaced 250ms apart (RFC 6762 §8.1)
//  3. If conflict detected, return with StateConflictDetected
//  4. Transition to Announcing state
//  5. Send 2 unsolicited announcements spaced 1s apart (RFC 6762 §8.3)
//  6. Transition to Established state
//
// RFC Quote:
//   "When ready to announce its presence, the host first probes to see if
//   anyone else is using the name. Only if it successfully completes probing
//   for the desired name without conflict is the host able to claim and use
//   that name."
//   (RFC 6762 §8, paragraph 1)
//
// Parameters:
//   - ctx: Context for cancellation. Cancelling stops probing/announcing immediately.
//   - serviceName: Fully qualified service name (e.g., "MyApp._http._tcp.local.")
//
// Returns:
//   - error: Context error if canceled, nil on successful registration or conflict
//
// Functional Requirements:
//   - FR-201: Service registration state machine
//   - FR-202: Context-aware cancellation
//   - Task T-038: Implement Machine.Run() with context cancellation
//   - Decision R001: Goroutine-per-service architecture
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := machine.Run(ctx, "MyApp._http._tcp.local.")
//	if err != nil {
//	    return fmt.Errorf("registration failed: %w", err)
//	}
func (sm *Machine) Run(ctx context.Context, serviceName string) error {
    // Implementation
}
```

---

## Constant Documentation

Every constant must explain its value, RFC source, and rationale.

### Template

```go
const (
    // <ConstantName> <one-line explanation>.
    //
    // RFC 6762 §X.Y: <RFC requirement>
    //
    // <2-3 sentences explaining WHY this specific value, including any
    // tradeoffs, alternative values considered, or protocol constraints.>
    //
    // Example: <Usage context or example if helpful>
    ConstantName = value
)
```

### Example

```go
const (
    // ProbeCount is the number of probe queries sent during the probing phase.
    //
    // RFC 6762 §8.1: Probing
    //
    // The RFC specifies exactly 3 probes to balance conflict detection reliability
    // with registration speed. Fewer probes increase collision risk; more probes
    // delay service availability unnecessarily. The 3-probe sequence provides
    // ~750ms of conflict detection (3 × 250ms) before announcing.
    //
    // RFC Quote:
    //   "A host probes to see if a resource record set with a given name, rrtype,
    //   and rrclass is already in use, by sending query messages asking for that
    //   set, and seeing if any other host responds."
    //   (RFC 6762 §8.1, paragraph 1)
    ProbeCount = 3

    // ProbeInterval is the time between successive probe queries.
    //
    // RFC 6762 §8.1: Probing
    //
    // The RFC mandates 250ms spacing between probes. This interval balances
    // timely registration with giving other devices adequate time to respond
    // to probe queries. Shorter intervals risk missing responses due to network
    // latency; longer intervals delay service availability.
    ProbeInterval = 250 * time.Millisecond

    // DefaultTTL is the default TTL for service records.
    //
    // RFC 6762 §10: Resource Record TTL Values and Cache Coherency
    //
    // The RFC recommends 75 minutes (4500 seconds) as a default TTL, providing
    // a balance between cache efficiency and timely updates. Services expected
    // to change frequently should use shorter TTLs; static services can use longer.
    //
    // RFC Quote:
    //   "It is recommended that Multicast DNS resource records for a service with
    //   a continuously changing list of available services use a TTL of 75 minutes."
    //   (RFC 6762 §10, paragraph 2)
    DefaultTTL = 75 * time.Minute
)
```

---

## Inline Comments

Inline comments explain non-obvious decisions, RFC requirements, and "why" rationale.

### Guidelines

1. **Explain WHY, Not WHAT**: The code itself shows what; comments explain why
2. **RFC References**: Link protocol behavior to RFC sections
3. **Rationale**: Explain tradeoffs, alternatives considered, gotchas
4. **Semgrep Suppressions**: Always justify with reason

### Template

```go
// <Explain the rationale, RFC requirement, or gotcha being addressed>
// RFC 6762 §X.Y: <Relevant RFC text if applicable>
//
// Rationale: <Why this approach vs alternatives>
code
```

### Examples

```go
// Manual unlock required: Must release lock before calling user callback to avoid deadlocks.
// Callback may access state machine, so holding lock would cause deadlock.
sm.mu.Lock() // nosemgrep: beacon-mutex-defer-unlock
sm.currentState = newState
sm.mu.Unlock()

// Notify test hook (called WITHOUT lock to prevent deadlocks)
if sm.onStateChange != nil {
    sm.onStateChange(newState)
}
```

```go
// RFC 6762 §8.2: Simultaneous Probe Tiebreaking
// When multiple hosts probe for the same name simultaneously, the host with
// the lexicographically later record data wins. This ensures deterministic
// conflict resolution without requiring coordination.
if bytes.Compare(ourData, theirData) > 0 {
    // We win the tiebreak - continue probing
    return true
}
// We lose - must rename and reprobe
return false
```

```go
// Known-answer suppression per RFC 6762 §7.1
// If the querier already knows about this record (included in known-answer section),
// we omit it from the response to reduce network traffic.
for _, knownAnswer := range query.Answers {
    if recordMatches(record, knownAnswer) {
        continue // Skip this record
    }
}
```

---

## Traceability

All code must be traceable to requirements, specifications, and RFCs.

### Traceability Elements

1. **RFC Sections**: Link to specific RFC section (e.g., "RFC 6762 §8.1")
2. **Functional Requirements**: Reference FR-XXX IDs from specs
3. **User Stories**: Reference US-X from feature specs
4. **Tasks**: Reference T-XXX from tasks.md
5. **ADRs**: Reference ADR-XXX for architectural decisions
6. **Decisions**: Reference R-XXX from implementation plans

### Example

```go
// ConflictDetector implements RFC 6762 §8.2 simultaneous probe tiebreaking.
//
// Functional Requirements:
//   - FR-202: Conflict detection and resolution
//   - US-2: Name conflict resolution
//
// Implementation:
//   - Task T-045: Implement lexicographic comparison logic
//   - Decision R003: Use bytes.Compare for deterministic ordering
//   - ADR-006: Conflict detection strategy
type ConflictDetector struct {
    // ...
}
```

---

## Examples

Provide examples for complex APIs, packages, and algorithms.

### Guidelines

1. **Compilable**: Examples must compile (or use `// Example:` format)
2. **Realistic**: Show real-world usage, not trivial cases
3. **Complete**: Include error handling and resource cleanup
4. **Focused**: Demonstrate one concept at a time

### Package Example

```go
// ## EXAMPLE USAGE
//
// Register a single service:
//
//	resp, err := responder.New(responder.WithInterface("en0"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer resp.Close()
//
//	service := responder.Service{
//	    InstanceName: "MyApp",
//	    ServiceType:  "_http._tcp.local.",
//	    Port:         8080,
//	    TXT:          []string{"version=1.0"},
//	}
//
//	err = resp.Register(context.Background(), service)
//	if err != nil {
//	    log.Fatal(err)
//	}
```

### Function Example

```go
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := machine.Run(ctx, "MyApp._http._tcp.local.")
//	if err == context.DeadlineExceeded {
//	    return fmt.Errorf("registration timed out")
//	}
//	if machine.GetState() == StateConflictDetected {
//	    return fmt.Errorf("name conflict detected")
//	}
```

---

## Review Checklist

Use this checklist when reviewing documentation:

### Package-Level
- [ ] WHY THIS PACKAGE EXISTS section present and clear
- [ ] PRIMARY TECHNICAL AUTHORITY lists relevant RFCs/ADRs
- [ ] DESIGN RATIONALE explains architectural decisions
- [ ] RFC COMPLIANCE lists all implemented requirements
- [ ] KEY CONCEPTS defines domain terminology
- [ ] EXAMPLE USAGE provides compilable example

### Type Documentation
- [ ] Purpose clearly stated in first sentence
- [ ] RFC section(s) referenced
- [ ] Wire format documented (for protocol types)
- [ ] Functional requirements listed
- [ ] Usage example provided (for complex types)
- [ ] Field comments explain purpose and constraints

### Function Documentation
- [ ] Purpose clearly stated in first sentence
- [ ] RFC section(s) referenced
- [ ] Algorithm steps documented (for complex logic)
- [ ] All parameters documented with constraints
- [ ] All return values documented
- [ ] Error conditions documented
- [ ] Functional requirements/tasks listed
- [ ] Example provided (for complex functions)

### Constant Documentation
- [ ] Value explained with rationale
- [ ] RFC section quoted or referenced
- [ ] Tradeoffs mentioned (why this value vs alternatives)

### Inline Comments
- [ ] Non-obvious logic explained with "WHY"
- [ ] RFC requirements referenced where applicable
- [ ] Semgrep suppressions justified
- [ ] Gotchas and pitfalls documented

### Traceability
- [ ] RFC sections linked to protocol behavior
- [ ] Functional requirements (FR-XXX) referenced
- [ ] Tasks (T-XXX) referenced
- [ ] ADRs referenced for architectural decisions
- [ ] User stories (US-X) referenced where applicable

### Examples
- [ ] Examples compile or use proper format
- [ ] Examples show realistic usage
- [ ] Examples include error handling
- [ ] Examples focused on one concept

---

## Tools

### Godoc Generation

```bash
# Generate godoc locally
godoc -http=:6060

# View package documentation
open http://localhost:6060/pkg/github.com/joshuafuller/beacon/responder/
```

### Documentation Coverage

```bash
# Check for undocumented exports
go doc -all ./responder | grep -E "^(type|func|const|var)" | grep -v "//"
```

### RFC Reference Validation

```bash
# Find all RFC references in code
grep -r "RFC 6762" --include="*.go" .

# Validate RFC section references exist
./scripts/validate-rfc-refs.sh
```

---

## References

- **Exemplar**: `internal/state/machine.go`
- **RFC 6762**: Multicast DNS specification
- **RFC 6763**: DNS-Based Service Discovery
- **Effective Go**: https://go.dev/doc/effective_go
- **Go Doc Comments**: https://go.dev/doc/comment

---

**Document Status**: Living document - update as standards evolve
**Next Review**: After Phase 1 (Priority 1 packages) completion
