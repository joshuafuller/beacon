# Beacon Development Guidelines

**Last Updated**: 2026-01-06 (Production-Ready Audit Complete - All Milestones Delivered)

This file provides context for Claude AI when working on the Beacon project.

---

## Claude Code Quick-Start (read first)

- **Know the layers**: Public APIs live in `querier/` & `responder/`; concrete mDNS logic stays under `internal/` to honor F-2 boundaries.
- **Stay milestone-aware**: We are in **M2 responder polish** (122/129 tasks done). Only touch M1 artifacts for regressions or docs.
- **Go-to specs**: Start with `specs/006-mdns-responder/{spec,plan,tasks}.md` plus `.specify/specs/F-2-architecture-layers.md` for import rules.
- **Primary commands**: `make test`, `make test-race`, `make semgrep-check`, and `go test ./... -run <Name>` for focused suites.
- **Token budget tip**: Quote just the sections you need (e.g., RFC 6762 §8) rather than pasting entire documents.
- **Summaries before detail**: Provide short status recaps before pasting large diffs or logs to keep startup context tight.
- **Citations & reports**: Always cite spec sections, ADRs, or RFC clauses when justifying protocol or architectural choices.
- **State machine focus**: Open tasks concentrate on probing/announcing polish—check `internal/state/*` tests first.
- **Hook awareness**: Pre-commit runs Semgrep; if reproducing failures, mention the exact rule ID.
- **Exit checklist**: Before finishing, ensure tests + `make semgrep-check` are green and documentation references are updated.

---

## Project Overview

**Beacon** is a lightweight, high-performance mDNS (Multicast DNS) library for Go, implementing RFC 6762 for service discovery on local networks.

**Current Status**: ✅ **PRODUCTION READY** - All Milestones Complete
- M1: mDNS Querier ✅ 100% Complete
- M1.1: Architectural Hardening ✅ 100% Complete
- M2: mDNS Responder ✅ 100% Complete (129/129 tasks)
- 007: Interface-Specific Addressing ✅ 100% Complete (116/116 tasks)
- Production Audit ✅ Complete (TODO/FIXME audit, documentation verified)

---

## Technology Stack

### Language & Runtime
- **Go 1.21+** (required)
- Core packages: `net`, `context`, `time`, `sync`, `encoding/binary`
- External (std-adjacent) deps: `golang.org/x/net` (multicast helpers), `golang.org/x/sys` (socket options)

### Architecture
- **Clean Architecture** with strict layer boundaries (F-2)
- **TDD Methodology** (RED → GREEN → REFACTOR)
- **Context-aware** operations throughout (F-9)

---

## Project Structure

```
beacon/
├── querier/                    # Public API - mDNS query interface
│   ├── querier.go             # Main Querier type
│   ├── querier_test.go        # Unit & integration tests
│   └── types.go               # Public types (ResourceRecord, etc.)
│
├── responder/                  # Public API - mDNS responder interface (006-mdns-responder)
│   ├── responder.go           # Main Responder type
│   ├── service.go             # Service definition and validation
│   ├── options.go             # Configuration options
│   └── conflict_detector.go   # RFC 6762 §8.2 tie-breaking logic
│
├── internal/                   # Internal implementation (not importable)
│   ├── transport/             # Network transport abstraction (M1-Refactoring)
│   │   ├── transport.go       # Transport interface
│   │   ├── udp.go             # UDPv4Transport (IPv4 multicast)
│   │   ├── buffer_pool.go     # sync.Pool for 9KB buffers
│   │   └── mock_transport.go  # Test double
│   │
│   ├── message/               # DNS message parsing/building
│   │   ├── builder.go         # Query/response message construction
│   │   ├── parser.go          # Response parsing
│   │   ├── name.go            # DNS name encoding (RFC 6763 §4.3)
│   │   └── message.go         # DNS message types
│   │
│   ├── protocol/              # mDNS protocol constants
│   │   └── constants.go       # Port 5353, multicast addresses, record types
│   │
│   ├── errors/                # Custom error types
│   │   └── errors.go          # NetworkError, ValidationError, etc.
│   │
│   ├── responder/             # Responder implementation (006-mdns-responder)
│   │   ├── registry.go        # Thread-safe service registry
│   │   ├── response_builder.go # RFC 6762 §6 response construction
│   │   └── known_answer.go    # RFC 6762 §7.1 suppression logic
│   │
│   ├── state/                 # State machine for probing/announcing (006-mdns-responder)
│   │   ├── machine.go         # State machine orchestration
│   │   ├── prober.go          # RFC 6762 §8.1 probing
│   │   └── announcer.go       # RFC 6762 §8.3 announcing
│   │
│   ├── records/               # DNS record construction (006-mdns-responder)
│   │   ├── record_set.go      # PTR, SRV, TXT, A record builders
│   │   └── ttl.go             # RFC 6762 §10 TTL values
│   │
│   ├── security/              # Input validation and rate limiting (006-mdns-responder)
│   │   ├── validation.go      # RFC-compliant input validation
│   │   └── rate_limiter.go    # RFC 6762 §6.2 per-interface rate limiting
│   │
│   └── network/               # Socket operations (legacy, being phased out)
│       └── socket.go          # Legacy socket code
│
├── tests/                      # Test suites
│   ├── contract/              # Contract tests (RFC compliance)
│   ├── integration/           # Real network tests
│   └── fuzz/                  # Fuzz testing
│
├── .specify/                   # ⭐ Spec Kit Framework Configuration
│   ├── specs/                 # Foundation specifications (F-1 through F-11)
│   │   ├── F-2-architecture-layers.md     # Layer boundaries
│   │   ├── F-3-error-handling.md          # Error propagation
│   │   └── F-9-transport-layer-config.md  # Socket configuration
│   ├── memory/                # Constitutional memory
│   │   └── constitution.md    # Project principles
│   └── templates/             # Spec/plan/task templates
│
├── specs/                      # ⭐ Feature Specifications (Milestone work)
│   ├── 001-spec-kit-migration/
│   ├── 002-mdns-querier/      # M1 implementation spec
│   ├── 003-m1-refactoring/    # M1-Refactoring spec (97 tasks complete)
│   │   ├── spec.md            # Requirements and success criteria
│   │   ├── plan.md            # Implementation strategy
│   │   └── tasks.md           # Executable tasks with checkpoints
│   └── 006-mdns-responder/    # M2 responder implementation (129 tasks, 122 complete)
│       ├── spec.md            # Responder requirements and RFC compliance
│       ├── plan.md            # State machine and record construction strategy
│       ├── tasks.md           # Executable tasks (94.6% complete)
│       ├── SECURITY_AUDIT.md  # Security validation (Grade: STRONG)
│       ├── CODE_REVIEW.md     # Code quality review (Grade: A)
│       └── PERFORMANCE_ANALYSIS.md # Performance profiling (Grade: A+)
│
├── RFC%20Docs/                   # ⭐ Protocol Specifications (SOURCE OF TRUTH)
│   ├── rfc6762.txt            # mDNS specification
│   └── rfc1035.txt            # DNS message format
│
├── docs/decisions/             # Architecture Decision Records (ADRs)
│   ├── 001-transport-interface-abstraction.md
│   ├── 002-buffer-pooling-pattern.md
│   └── 003-integration-test-timing-tolerance.md
│
└── archive/                    # Historical artifacts
    └── m1-refactoring/        # M1-Refactoring metrics and reports
        ├── README.md          # Archive documentation
        ├── reports/           # Completion reports
        └── metrics/           # Test/benchmark data

```

---

## Key Architectural Patterns

### 1. Transport Interface Abstraction
**Why**: Decouples querier from network implementation, enables IPv6 (M2), testability
**ADR**: docs/decisions/001-transport-interface-abstraction.md

```go
type Transport interface {
    Send(ctx context.Context, packet []byte, dest net.Addr) error
    Receive(ctx context.Context) ([]byte, net.Addr, int, error) // 4th return: interfaceIndex (007)
    Close() error
}
```

**Implementations**:
- `UDPv4Transport` - Production IPv4 multicast (224.0.0.251:5353)
- `MockTransport` - Test double for unit testing

---

### 2. Buffer Pooling
**Why**: Eliminates 900KB/sec allocations (9KB per receive call)
**ADR**: docs/decisions/002-buffer-pooling-pattern.md
**Result**: 99% allocation reduction (9000 B/op → 48 B/op)

```go
// UDPv4Transport.Receive() uses buffer pool
bufPtr := GetBuffer()
defer PutBuffer(bufPtr)  // Returns buffer to pool
```

---

### 3. Layer Boundaries (F-2)
**Rule**: Strict import restrictions

```
querier → transport → network
       ↘ protocol
       ↘ message
       ↘ errors
```

**Validation**: `grep -rn "internal/network" querier/` must return 0 matches

---

### 4. Error Propagation (FR-004)
**Rule**: Never swallow errors
**Pattern**: All errors wrapped in typed errors (`NetworkError`, `ValidationError`, `WireFormatError`)

```go
// ✅ CORRECT
func (t *UDPv4Transport) Close() error {
    return t.conn.Close()  // Propagate error
}

// ❌ WRONG
func (t *UDPv4Transport) Close() error {
    t.conn.Close()
    return nil  // Error swallowed!
}
```

---

## Common Commands

### Testing
```bash
# Run all tests
make test                    # Basic test run
make test-race              # With race detector
make test-coverage          # With coverage (≥80% gate)

# Detailed coverage report
make test-coverage-report   # Pretty formatted by package

# Track coverage over time
./scripts/coverage-trend.sh              # Record current
./scripts/coverage-trend.sh --show       # Show history
./scripts/coverage-trend.sh --graph      # Show trend graph

# Specific test suites
go test ./querier -run TestQuery              # Single test
go test ./tests/integration -v                # Integration tests
go test -bench=. -benchmem ./...              # Benchmarks
go test -fuzz=FuzzParseMessage -fuzztime=10s ./tests/fuzz  # Fuzz tests

# CI pipelines
make ci-fast                # Fast CI (unit + race + coverage)
make ci-full                # Full CI (all tests)
```

**Coverage checklist** (details: `scripts/coverage-trend.sh --help`, docs/COMPLIANCE_DASHBOARD.md):
- Target ≥80% per constitution; aim for ≥85% when touching hot paths.
- Run `make test-coverage-report` for package deltas before large refactors.
- Log snapshots via `./scripts/coverage-trend.sh` when coverage meaningfully shifts.
- Investigate dips in `internal/responder`, `internal/transport`, and `querier`—open focus areas.

### Building
```bash
# Build library (check for errors)
go build ./...

# Vet code
go vet ./...

# Format code
gofmt -w .

# Run all quality checks (via Makefile)
make test
```

### Semgrep (Static Analysis)

**Semgrep checklist** (full rule docs: `SEMGREP_RULES_SUMMARY.md`, quick ref: `.semgrep-tests/README.md`):
- Hooks: `.githooks/pre-commit` runs Semgrep automatically—rerun via `make semgrep-check` to reproduce CI.
- Prioritize ERROR rules first (timer/ticker, mutex, panic-on-input). Mention rule IDs when reporting failures.
- Use `make semgrep` for exploratory scans; switch to `make semgrep-check` before final commits or PRs.
- Suppress only with inline justification (`nosemgrep: <rule>` plus reason). Keep overrides rare and reviewed.
- Track rule updates in `SEMGREP_FINDINGS.md` before/after large changes.

---

## Code Style

### Go Standards
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` (no exceptions)
- Use `go vet` (must pass with zero warnings)

### Testing Standards
- **TDD**: Write tests FIRST (RED phase)
- **Coverage**: Maintain ≥85% (current: 84.8%)
- **Stability**: Zero flaky tests (achieved in M1-Refactoring)
- **Naming**: Test functions describe behavior (`TestQuery_Timeout_ReturnsEmptyResponse`)

### Documentation Standards
- **Godoc**: All exported types/functions must have documentation
- **ADRs**: Major architectural decisions documented in `docs/decisions/`
- **Comments**: Explain WHY, not WHAT
- **RFC References**: Link to RFC sections when implementing protocol behavior

---

## Recent Changes
- 008-documentation-production-polish: Added Go 1.21+ (matches Beacon minimum version), Markdown (GitHub Flavored), Hugo 0.120+ (static site)

### 007-interface-specific-addressing (Core Complete - 2025-11-06)
**Branch**: 007-interface-specific-addressing
**Status**: ✅ Core Implementation Complete (73/116 tasks, 62.9%)
**Summary**: RFC 6762 §15 compliance - Interface-specific IP addressing for multi-interface hosts

**Problem Solved**:
Multi-interface hosts (WiFi + Ethernet, multi-NIC servers) were advertising the same IP address on all interfaces, violating RFC 6762 §15 and causing connectivity failures.

**Solution**:
- Extract interface index from IP_PKTINFO (Linux) / IP_RECVIF (macOS/BSD) control messages
- Use interface-specific IP resolver `getIPv4ForInterface(interfaceIndex)` in response building

**Achievements**:

**Key Changes**:

**Testing**:

**Documentation**:

**Remaining**: Optional documentation (T085-T089), manual testing (T090-T099)

---

### 006-mdns-responder (94.6% Complete - 2025-11-04)
**Branch**: 006-mdns-responder
**Status**: 🚧 94.6% Complete (122/129 tasks), Final documentation polish in progress
**Summary**: Full mDNS responder implementation with RFC 6762/6763 compliance

**Achievements**:

**New Public APIs**:

**Key Files Created**:

**Documentation**:

**Remaining**: T123-T126 (documentation updates), T116-T117 (deferred - require macOS/Avahi)

---

### M1.1 Architectural Hardening (Complete - 2025-11-02)
**Branch**: 004-m1-1-architectural-hardening
**Status**: ✅ Complete (ready to merge)
**Summary**: Production-ready socket configuration, security, and interface management

**Achievements**:

**New Public APIs**:

**Key Files Changed**:

**Documentation**:

---

### M1-Refactoring (Complete - 2025-11-01)
**Branch**: 003-m1-refactoring
**Status**: ✅ Merged to main
**Summary**: Comprehensive architectural refactoring

**Achievements**:

**Key Files Changed**:

**Documentation**:

---

### M1 Implementation (Complete - 2025-10-XX)
**Branch**: 002-mdns-querier
**Summary**: Initial mDNS query implementation

**Features**:

---

## Next Steps (M1.2 / M2)

### Planned Features (M1.2 - Service Discovery)
1. **Service Browsing**
   - Browse for services (`_services._dns-sd._udp.local`)
   - Service instance enumeration
   - Service type discovery

2. **Enhanced Resource Record Parsing**
   - Better handling of compressed names
   - Support for additional record types (AAAA for IPv6)

### Planned Features (M2 - IPv6 & Advanced)
1. **IPv6 Support**
   - Dual-stack operation (IPv4 + IPv6)
   - IPv6 multicast (FF02::FB)
   - Per-interface transport binding

2. **Full Source Filtering**
   - Per-interface SourceFilter integration
   - Same-subnet validation (requires M2 per-interface transports)

3. **Observability**
   - Structured logging (F-6 Logging & Observability spec)
   - Metrics and telemetry

**Spec**: See `specs/` directory for feature specifications

---

## Specification-Driven Development (Spec Kit Framework)

**This is a Spec Kit project** - all development follows the [Spec Kit](https://github.com/github/spec-kit) framework's specification-driven methodology.

### Core Philosophy

Development follows a strict **Spec → Plan → Tasks → TDD → Validate** cycle:

1. **Specification** - Define WHAT and WHY
   - Requirements (functional/non-functional)
   - Success criteria
   - RFC compliance constraints

2. **Planning** - Define HOW
   - Architecture decisions (documented as ADRs)
   - Implementation strategy
   - Risk analysis

3. **Tasks** - Define executable steps
   - Granular, testable tasks (T001, T002, ...)
   - TDD cycles (RED → GREEN → REFACTOR)
   - Checkpoint validation

4. **TDD Execution** - STRICT test-first implementation
   - Write tests BEFORE code
   - Validate against F-Specs
   - Check RFC compliance

5. **Validation** - Prove completion
   - All tasks complete
   - All success criteria met
   - Completion report generated

### Key Directories

#### `.specify/` - Framework Configuration
- **`.specify/specs/`** - Foundation specifications
  - **BEACON_FOUNDATIONS.md** - ⭐ DNS/mDNS/DNS-SD conceptual foundation (919 lines)
    - DNS fundamentals, mDNS essentials, DNS-SD concepts
    - Terminology glossary, reference tables
    - **Read this first** to understand the domain
  - **F-2**: Layer Boundaries - Defines clean architecture constraints
  - **F-3**: Error Handling - Error propagation patterns (FR-004)
  - **F-9**: Transport Layer Configuration - Socket options, Avahi/Bonjour coexistence
  - **F-10**: Network Interface Management - Interface selection, VPN exclusion
  - **F-11**: Security Architecture - Rate limiting, source IP filtering
- **`.specify/memory/constitution.md`** - Constitutional principles (see below)
- **`.specify/templates/`** - Templates for specs, plans, tasks

#### `specs/` - Feature Specifications
- **`specs/[milestone]/spec.md`** - Feature specification
- **`specs/[milestone]/plan.md`** - Implementation plan
- **`specs/[milestone]/tasks.md`** - Executable tasks
- Example: `specs/003-m1-refactoring/` (97 tasks, all complete)

#### `RFC%20Docs/` - Protocol Specifications (CRITICAL)
- **RFC 6762**: mDNS specification - **ALL behavior must comply**
- **RFC 1035**: DNS message format
- RFCs are the source of truth for protocol behavior
- When in doubt, reference the RFC

#### `docs/decisions/` - Architecture Decision Records
- **ADR-001**: Transport Interface Abstraction
- **ADR-002**: Buffer Pooling Pattern
- **ADR-003**: Integration Test Timing Tolerance
- ADRs document WHY we made key architectural choices

## Common Task Recipes

| Scenario | Specs & References | Tests to Run First | Implementation Notes |
| --- | --- | --- | --- |
| Polish responder probing/announcing | `specs/006-mdns-responder/{plan,tasks}.md` (T118–T126), RFC 6762 §8, `.specify/specs/F-2-architecture-layers.md` | `go test ./internal/state -run Prober` then `make test` | Touch `internal/state/{prober,announcer}.go`; keep timers cancellable and cite RFC clauses when adjusting timings. |
| Add responder validation | `specs/006-mdns-responder/spec.md#validation`, `internal/security/validation.go` | `go test ./internal/security -run Validate` | Update validation errors only via `internal/errors`; document new constraints in `responder/service.go` Godoc. |
| Extend public responder API | `responder/responder.go`, `specs/006-mdns-responder/tasks.md` (API section), ADR-001 | `go test ./responder -run TestResponder` | Ensure new options thread through `internal/responder/registry.go`; add examples to `examples/` if behavior changes. |
| Investigate coverage regression | docs/COMPLIANCE_DASHBOARD.md, `scripts/coverage-trend.sh` | `make test-coverage-report` | Isolate packages <80% and add tests before refactors; log results via coverage script for history. |
| Semgrep failure triage | `SEMGREP_RULES_SUMMARY.md`, `.semgrep-tests/README.md` | `make semgrep-check` | Reproduce failing rule, adjust code or rule test; include rule ID + reasoning in PR description if suppression necessary. |

### Spec Kit Slash Commands

**IMPORTANT**: Use these commands when working on features:

- **`/speckit.specify`** - Create/update feature specification
  - Use when starting a new feature or milestone
  - Generates structured spec.md with requirements

- **`/speckit.plan`** - Generate implementation plan
  - Use after specification is complete
  - Creates plan.md with architecture decisions

- **`/speckit.tasks`** - Generate executable tasks
  - Use after plan is complete
  - Creates tasks.md with granular, testable tasks

- **`/speckit.implement`** - Execute implementation
  - Use to execute tasks in TDD cycles
  - Updates tasks.md as work progresses

- **`/speckit.analyze`** - Validate spec consistency
  - Use before starting implementation
  - Checks for conflicts between spec/plan/tasks

### Development Workflow Example (M1-Refactoring)

```bash
# 1. Create specification
/speckit.specify
# → Generated specs/003-m1-refactoring/spec.md

# 2. Plan implementation
/speckit.plan
# → Generated specs/003-m1-refactoring/plan.md with ADRs

# 3. Generate tasks
/speckit.tasks
# → Generated specs/003-m1-refactoring/tasks.md (97 tasks)

# 4. Execute tasks (TDD)
# T011-T018: RED - Write transport interface tests FIRST
# T019-T037: GREEN - Implement UDPv4Transport to make tests pass
# T038-T043: REFACTOR - Clean up, validate, checkpoint

# 5. Validate completion
# → COMPLETION_VALIDATION.md (9 criteria)
# → REFACTORING_COMPLETE.md (full report)
# → All tasks marked [x], all criteria met
```

---

## Constitutional Principles

From `.specify/memory/constitution.md`:

1. **Protocol Compliance First** - RFC 6762 (mDNS) compliance is non-negotiable
2. **Minimal External Dependencies** - Prefer standard library; currently only `golang.org/x/net` + `golang.org/x/sys`
3. **Context-Aware Operations** - All blocking operations accept `context.Context`
4. **Clean Architecture** - Strict layer boundaries (F-2)
5. **Test-Driven Development** - Tests written first (TDD)
6. **Performance Matters** - Optimize hot paths (buffer pooling example)

---

## Performance Characteristics

### Current Metrics (Post M1-Refactoring)
- **Query Latency**: 163 ns/op (9% improvement)
- **Allocations**: 48 B/op in receive path (99% reduction)
- **Concurrent Queries**: 100+ supported (NFR-002)
- **Test Coverage**: 84.8%
- **Flaky Tests**: 0

### NFRs (Non-Functional Requirements)
- **NFR-001**: Query processing overhead <100ms ✅
- **NFR-002**: Support ≥100 concurrent queries ✅
- **NFR-003**: ≥80% allocation reduction (achieved 99%) ✅

---

## Troubleshooting

### Common Issues

**Test Failures**:
```bash
# Clear test cache
go clean -testcache

# Run with verbose output
go test ./... -v
```

**Integration Test Flakiness**:
- Integration tests use real mDNS traffic
- Timing-sensitive (100ms tolerance added in M1-Refactoring)
- See ADR-003 for timing tolerance rationale

**Layer Boundary Violations**:
```bash
# Check for violations
grep -rn "internal/network" querier/

# Should return: no matches
```

---

## Git Workflow

### Branch Naming
- Feature: `XXX-feature-name` (e.g., `003-m1-refactoring`)
- Where XXX matches spec directory number

### Commit Messages
- Use conventional commits style
- Reference spec/task numbers
- Include WHY, not just WHAT

### Testing Before Commit
```bash
go test ./... -race
go vet ./...
gofmt -l . | grep . && echo "Files need formatting"
```

---

## Resources

### RFCs
- [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) - Multicast DNS (primary)
- [RFC 1035](https://www.rfc-editor.org/rfc/rfc1035.html) - DNS specification (wire format)

### Internal Docs
- `docs/decisions/` - Architecture Decision Records
- `specs/` - Feature specifications
- `archive/m1-refactoring/reports/` - Historical completion reports

### External
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)

---

<!-- MANUAL ADDITIONS START -->
- **Beacon-specific prompting**:
  - When citing requirements, reference spec IDs (e.g., `specs/006-mdns-responder/tasks.md#T121`) or ADR numbers alongside file paths.
  - Prefer layered summaries: 1) headline decision, 2) key evidence (tests/specs), 3) optional deep dive to conserve tokens.
  - Collapse large RFC excerpts into section pointers (e.g., "RFC 6762 §6.7, paragraphs 1-2") unless verbatim wording is essential.
- **Token stewardship**:
  - Keep conversation deltas ≤400 tokens where possible; link to prior context instead of repeating it.
  - Annotate any long log or diff with a one-line takeaway before pasting the snippet.
- **Failure reporting**:
  - For CI or Semgrep failures, provide command, exit status, and top 3 findings with rule IDs—offer remediation options if known.
  - If a required dependency is missing locally, suggest the Makefile/script that installs or mocks it.
- **PR etiquette**:
  - Summaries must map work to open specs/tasks and mention any Semgrep suppressions explicitly.
  - Highlight user-visible behavior changes and note testing gaps if something could not be run.
<!-- MANUAL ADDITIONS END -->

---

**Generated**: 2025-11-01 (living document)
**Project**: Beacon mDNS Library
**Status**: M2 Responder Implementation 94.6% complete (final polish + docs)
- When removing a test, justify the decision and note compensating coverage if needed.

## Active Technologies
- Go 1.21+ + Standard library + `golang.org/x/sys` (platform-specific socket options from M1.1), `golang.org/x/net` (multicast group management from M1.1, interface index extraction via `ipv4.PacketConn` from 007-interface-specific-addressing)
- In-memory (registered services, resource record sets with TTLs) (006-mdns-responder)
- Go 1.21+ (matches Beacon minimum version), Markdown (GitHub Flavored), Hugo 0.120+ (static site) (008-documentation-production-polish)
- Static files (markdown, examples, site content) in Git repository, no database required (008-documentation-production-polish)
