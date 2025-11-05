# User Story 2 Pre-Implementation Review

**Feature**: M2 mDNS Responder - Name Conflict Resolution (RFC 6762 §8.2)
**Branch**: 006-mdns-responder
**Phase**: Phase 4 (Tasks T048-T063)
**Status**: Ready for Implementation
**Date**: 2025-11-03

---

## Purpose

This document provides a **focused review** of Constitutional principles and F-Spec requirements relevant to **User Story 2: Name Conflict Resolution** before beginning implementation. This ensures proper policy compliance from the start.

---

## Constitutional Requirements

From `.specify/memory/constitution.md`:

### Principle I: RFC 6762 Compliance (PRIMARY)

**Requirement**: All mDNS behavior MUST comply with RFC 6762

**User Story 2 Specific**:
- **RFC 6762 §8.2**: Simultaneous Probe Tie-Breaking
- **RFC 6762 §9**: Conflict Resolution After Registration
- **RFC 6762 §8.1**: Probing (foundation from US1)

**Critical Rules**:
1. **Lexicographic comparison** for tie-breaking (case-insensitive, byte-wise)
2. **Defer if we lose** (suppress our probe, wait 1 second, pick new name)
3. **Continue if we win** (other host must defer)
4. **Rename on conflict** (append "-2", "-3", etc.)

**Validation**: All conflict detection logic must reference specific RFC sections

---

### Principle III: Test-Driven Development (TDD)

**Requirement**: ≥80% test coverage, RED → GREEN → REFACTOR

**User Story 2 Specific**:
- **T048-T053**: RED phase (write 6 failing tests FIRST)
- **T054-T060**: GREEN phase (implement to make tests pass)
- **T061-T063**: REFACTOR phase (clean up, validate)

**Critical Rules**:
1. Write tests BEFORE implementation
2. Verify tests FAIL initially (RED)
3. Implement minimal code to pass (GREEN)
4. Refactor with tests as safety net

**Validation**: Run `go test ./responder -v` after each task, expect specific pass/fail states

---

### Principle IV: Specification-Driven Development

**Requirement**: Spec → Plan → Tasks → Validate

**User Story 2 Specific**:
- **Spec**: `specs/006-mdns-responder/spec.md` §US-002
- **Plan**: `specs/006-mdns-responder/plan.md` §5.2.2
- **Tasks**: `specs/006-mdns-responder/tasks.md` T048-T063

**Critical Rules**:
1. Reference spec requirements in code comments
2. Update tasks.md as work progresses
3. Create checkpoints (CP-US2-1, CP-US2-2)
4. Generate validation report when complete

---

### Principle VIII: Phased, Iterative Delivery

**Requirement**: Complete phases sequentially

**User Story 2 Context**:
- **Phase 3 (US1)**: ✅ COMPLETE (Service registration working)
- **Phase 4 (US2)**: ⏳ NEXT (Name conflict resolution)

**Critical Rules**:
1. Do NOT start T048 until this review is complete
2. Complete US2 before starting US3 (Query Responding)
3. Validate Phase 4 complete before Phase 5

---

## F-Spec Requirements

### F-2: Package Structure and Layer Boundaries

**Location**: `.specify/specs/F-2-architecture-layers.md`

**User Story 2 Specific**:
- **ConflictDetector**: Lives in `responder/` package (NOT `internal/`)
- **Layer**: Domain logic (interfaces with Prober, StateMachine)
- **Dependencies**: Can import `internal/message` for DNS comparison, `internal/protocol` for constants

**Critical Rules**:
1. No circular dependencies (`responder/` ↔ `internal/`)
2. Public API surfaces in `responder/` package
3. Internal details in `internal/conflict/` if needed

**Validation**: Run `go list -f '{{.Imports}}' ./responder` - must not import `responder/`

---

### F-3: Error Handling and Propagation

**Location**: `.specify/specs/F-3-error-handling.md`

**User Story 2 Specific**:
- **ConflictDetector.DetectConflict()**: Returns `(bool, error)`
- **Errors**: `ConflictError` (wraps details about conflicting record)
- **Propagation**: Errors MUST be returned to caller (Prober), never swallowed

**Critical Rules**:
1. Use `fmt.Errorf("context: %w", err)` for wrapping
2. Define custom `ConflictError` type with fields: `OurRecord`, `TheirRecord`, `ComparisonResult`
3. Never use `panic()` in conflict detection (network input)

**Example**:
```go
type ConflictError struct {
    OurRecord      message.ResourceRecord
    TheirRecord    message.ResourceRecord
    WeDefer        bool  // true if we lost tie-break
}

func (e *ConflictError) Error() string {
    return fmt.Sprintf("name conflict: %s (we defer: %v)", e.OurRecord.Name, e.WeDefer)
}
```

---

### F-4: Concurrency and Context Management

**Location**: `.specify/specs/F-4-concurrency-context.md`

**User Story 2 Specific**:
- **ConflictDetector**: May be called concurrently by multiple Prober instances
- **Thread Safety**: Must be safe to call `DetectConflict()` from multiple goroutines
- **Context**: Accept `context.Context` if conflict detection becomes async in future

**Critical Rules**:
1. Use `sync.Mutex` if ConflictDetector has mutable state
2. Stateless functions preferred (pure comparison logic)
3. Document concurrency safety in godoc

**Example**:
```go
// DetectConflict checks if incomingRecord conflicts with ourRecord.
// Safe for concurrent use by multiple Prober instances.
func (cd *ConflictDetector) DetectConflict(ourRecord, incomingRecord message.ResourceRecord) (bool, error) {
    // Pure function - no shared mutable state
}
```

---

### F-8: Testing Strategy

**Location**: `.specify/specs/F-8-testing-strategy.md`

**User Story 2 Specific**:
- **Unit Tests**: `responder/conflict_detector_test.go` (6 tests minimum)
- **Integration Tests**: `responder/prober_test.go` (conflict scenarios)
- **Coverage Target**: ≥85% for new `ConflictDetector` code

**Test Cases (from tasks.md)**:
1. **T048**: No conflict (different names)
2. **T049**: Conflict detected (same name, different data)
3. **T050**: Tie-break (same name, same data) - we win
4. **T051**: Tie-break - we lose
5. **T052**: Lexicographic comparison edge cases
6. **T053**: Error handling (malformed records)

**Critical Rules**:
1. Use table-driven tests for tie-break scenarios
2. Test with RFC 6762 Appendix D examples
3. Run with `-race` flag (detect data races)

---

## RFC 6762 Requirements for US2

### §8.2: Simultaneous Probe Tie-Breaking

**Quote**:
> "When a host is probing for a set of records with the same name, if it receives a query for that name containing the same record types, the two hosts are in conflict. The host with the lexicographically later record wins."

**Implementation Requirements**:
1. **Compare record data** (RDATA field in DNS message)
2. **Lexicographic comparison** (byte-wise, case-insensitive for names)
3. **Winner continues probing**, loser defers
4. **Defer means**: Stop probing, wait 1 second, pick new name (append "-2")

**Example** (from RFC):
```
Our record:   myhost.local A 192.168.1.100
Their record: myhost.local A 192.168.1.50

Comparison: 192.168.1.100 > 192.168.1.50 lexicographically
Result: WE WIN (continue probing)
```

---

### §9: Conflict Resolution After Registration

**Quote**:
> "If a host receives a record that conflicts with one of its records, it MUST immediately rename the conflicting record by appending '-2' to the name."

**Implementation Requirements**:
1. **Detect conflicts** during probing AND after registration
2. **Rename immediately** (no delay)
3. **Increment suffix** ("-2", "-3", "-4", ...)
4. **Re-probe** with new name

**US2 Scope**: Probing conflicts only (post-registration conflicts in US4)

---

## What US1 Built (Foundation for US2)

From Phase 3 completion:

### 1. Service Struct (`responder/service.go`)

```go
type Service struct {
    InstanceName string
    ServiceType  string
    Domain       string
    Port         uint16
    TXTRecords   map[string]string
    state        ServiceState
    stateMu      sync.RWMutex
}
```

**US2 Uses**: `Service.InstanceName` is the name we're checking for conflicts

---

### 2. Registry (`responder/registry.go`)

```go
type Registry struct {
    services map[string]*Service
    mu       sync.RWMutex
}

func (r *Registry) Register(svc *Service) error
func (r *Registry) Unregister(instanceName string) error
```

**US2 Uses**: Registry tracks all services we're defending

---

### 3. StateMachine (`responder/state_machine.go`)

```go
type StateMachine struct {
    currentState ServiceState
    mu           sync.Mutex
}

func (sm *StateMachine) TransitionTo(newState ServiceState) error
```

**Current States**:
- `StateProbing`: Initial state, checking for conflicts
- `StateAnnouncing`: No conflicts, announcing presence
- `StateRegistered`: Fully registered, responding to queries

**US2 Adds**: Transition from `StateProbing` → `StateConflict` → `StateProbing` (with renamed service)

---

### 4. Prober (`responder/prober.go`)

```go
type Prober struct {
    transport Transport
}

func (p *Prober) Probe(ctx context.Context, svc *Service) error {
    // Send 3 probes, wait 250ms between each
    // US2 ADDS: Check responses for conflicts
}
```

**US2 Integration**: Prober calls `ConflictDetector.DetectConflict()` when receiving probe responses

---

## What US2 Will Add

### 1. ConflictDetector (`responder/conflict_detector.go`)

**Purpose**: Detect name conflicts and perform tie-breaking per RFC 6762 §8.2

**Interface**:
```go
type ConflictDetector struct {
    // Stateless - no fields needed
}

// DetectConflict checks if incomingRecord conflicts with ourRecord.
// Returns:
//   (true, nil)  - Conflict detected, we defer
//   (false, nil) - No conflict OR we win tie-break
//   (_, error)   - Error parsing/comparing records
func (cd *ConflictDetector) DetectConflict(
    ourRecord message.ResourceRecord,
    incomingRecord message.ResourceRecord,
) (bool, error)
```

**Key Methods**:
- `DetectConflict()`: Main entry point
- `lexicographicCompare()`: Byte-wise comparison (internal helper)
- `normalizeForComparison()`: Case-insensitive name handling (internal helper)

---

### 2. Prober Integration

**Changes to `responder/prober.go`**:
```go
func (p *Prober) Probe(ctx context.Context, svc *Service) error {
    detector := &ConflictDetector{}

    // Send 3 probes
    for i := 0; i < 3; i++ {
        p.sendProbe(svc)

        // US2 NEW: Listen for responses
        responses := p.receiveResponses(250 * time.Millisecond)

        // US2 NEW: Check each response for conflicts
        for _, response := range responses {
            conflict, err := detector.DetectConflict(ourRecord, response.Record)
            if err != nil {
                return fmt.Errorf("conflict detection failed: %w", err)
            }
            if conflict {
                return &ConflictError{...}  // Propagate to caller
            }
        }
    }

    return nil  // No conflicts detected
}
```

---

### 3. StateMachine Conflict Transition

**Changes to `responder/state_machine.go`**:
```go
const (
    StateProbing    ServiceState = "probing"
    StateConflict   ServiceState = "conflict"     // US2 NEW
    StateAnnouncing ServiceState = "announcing"
    StateRegistered ServiceState = "registered"
)

func (sm *StateMachine) HandleConflict() error {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    if sm.currentState != StateProbing {
        return fmt.Errorf("cannot handle conflict in state %s", sm.currentState)
    }

    sm.currentState = StateConflict
    return nil
}
```

---

### 4. Service Renaming Logic

**Changes to `responder/service.go`**:
```go
// Rename appends a numeric suffix to the instance name to resolve conflicts.
// Examples: "My Service" → "My Service-2" → "My Service-3"
func (s *Service) Rename() {
    s.stateMu.Lock()
    defer s.stateMu.Unlock()

    // Parse current suffix (if any)
    if strings.Contains(s.InstanceName, "-") {
        // Extract and increment: "My Service-2" → "My Service-3"
    } else {
        // First conflict: "My Service" → "My Service-2"
        s.InstanceName = s.InstanceName + "-2"
    }
}
```

---

## Pre-Implementation Checklist

Before starting Task T048, verify:

- [ ] **Constitution reviewed**: Principles I, III, IV, VIII understood
- [ ] **F-Specs reviewed**: F-2, F-3, F-4, F-8 requirements clear
- [ ] **RFC 6762 §8.2 read**: Simultaneous probe tie-breaking understood
- [ ] **US1 foundation understood**: Service, Registry, Prober, StateMachine
- [ ] **US2 scope clear**: Probing conflicts only (not post-registration)
- [ ] **TDD workflow ready**: Write tests first, verify RED phase
- [ ] **Test files created**: `responder/conflict_detector_test.go` exists (empty)
- [ ] **Branch verified**: On `006-mdns-responder`, up-to-date with main

**Command to verify**:
```bash
# Check branch
git rev-parse --abbrev-ref HEAD  # Should show: 006-mdns-responder

# Verify US1 complete (Phase 3)
grep "Phase 3" specs/006-mdns-responder/tasks.md  # Should show all [x]

# Create test file
touch responder/conflict_detector_test.go

# Ready to start T048
```

---

## Task Execution Order (T048-T063)

### RED Phase (Write Tests First)
- **T048**: Test no conflict (different names) - EXPECT FAIL
- **T049**: Test conflict detected (same name, different data) - EXPECT FAIL
- **T050**: Test tie-break (we win) - EXPECT FAIL
- **T051**: Test tie-break (we lose) - EXPECT FAIL
- **T052**: Test lexicographic edge cases - EXPECT FAIL
- **T053**: Test error handling - EXPECT FAIL

**Validation**: `go test ./responder -v` - 6 tests FAIL (no ConflictDetector implemented yet)

---

### GREEN Phase (Implement to Pass Tests)
- **T054**: Implement ConflictDetector struct (empty)
- **T055**: Implement DetectConflict() - basic name comparison
- **T056**: Implement lexicographic comparison helper
- **T057**: Implement tie-breaking logic (RFC 6762 §8.2)
- **T058**: Add error handling for malformed records
- **T059**: Integrate with Prober
- **T060**: Add StateMachine conflict transition

**Validation**: `go test ./responder -v` - ALL tests PASS

---

### REFACTOR Phase (Clean Up)
- **T061**: Add godoc comments, RFC references
- **T062**: Run benchmarks, optimize hot paths
- **T063**: Final validation (coverage ≥85%, go vet, semgrep)

**Validation**: `make ci-fast` - ALL checks PASS

---

## Success Criteria (from spec.md)

User Story 2 is COMPLETE when:

1. ✅ ConflictDetector implements RFC 6762 §8.2 tie-breaking
2. ✅ Prober detects conflicts during probing (3 probes)
3. ✅ StateMachine transitions to StateConflict on conflict
4. ✅ Service renames with numeric suffix ("-2", "-3", ...)
5. ✅ All 6 unit tests pass (table-driven)
6. ✅ Integration test: Two services, same name, one renames
7. ✅ Coverage ≥85% for conflict_detector.go
8. ✅ Go vet, semgrep pass (zero findings)

**Validation Command**:
```bash
make ci-fast  # Must pass before marking US2 complete
```

---

## References

### Constitutional
- `.specify/memory/constitution.md` - Principles I, III, IV, VIII

### F-Specs
- `.specify/specs/F-2-architecture-layers.md` - Package structure
- `.specify/specs/F-3-error-handling.md` - Error propagation
- `.specify/specs/F-4-concurrency-context.md` - Thread safety
- `.specify/specs/F-8-testing-strategy.md` - TDD requirements

### RFC
- `RFC%20Docs/RFC-6762-Multicast-DNS.txt` - §8.2 (Simultaneous Probe Tie-Breaking), §9 (Conflict Resolution)

### Project Docs
- `specs/006-mdns-responder/spec.md` - US2 requirements
- `specs/006-mdns-responder/plan.md` - Implementation strategy
- `specs/006-mdns-responder/tasks.md` - Tasks T048-T063

---

## Notes

**This review is a GATE** - do not proceed to T048 until:
1. All checklist items verified
2. RFC 6762 §8.2 understood
3. TDD workflow clear (RED → GREEN → REFACTOR)

**Estimated Time**: 4-6 hours (based on US1 velocity)

**Risk**: Lexicographic comparison is subtle - use RFC examples for validation

**Next Step**: Task T048 (write first failing test for DetectConflict)

---

**Review Complete**: 2025-11-03
**Reviewer**: Pre-implementation validation
**Status**: ✅ READY TO PROCEED with Phase 4: User Story 2
