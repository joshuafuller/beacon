# Changelog

All notable changes to the Beacon mDNS library will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### M1-Refactoring (2025-11-01) - Internal Architecture Overhaul

Complete architectural refactoring to prepare for M1.1 and M2 milestones. **Zero breaking changes to public API**.

#### Added
- **Transport Interface Abstraction** ([ADR-001](docs/decisions/001-transport-interface-abstraction.md))
  - New `internal/transport` package with `Transport` interface
  - `UDPv4Transport`: Production IPv4 UDP multicast implementation
  - `MockTransport`: Test double for unit testing without network sockets
  - Context propagation support (F-9 REQ-F9-7) for cancellation and deadlines
  - Enables future IPv6 support (M2) via `UDPv6Transport` implementation

- **Buffer Pooling Optimization** ([ADR-002](docs/decisions/002-buffer-pooling-pattern.md))
  - `sync.Pool`-based buffer reuse for receive operations
  - Reduces allocations by **99%** (9000 B/op → 48 B/op)
  - Eliminates 900 KB/sec GC pressure at 100 queries/second
  - Thread-safe, automatic scaling with load

#### Changed
- **Querier Package Refactoring**
  - `querier.Querier` now uses `transport.Transport` interface instead of `net.PacketConn`
  - Removed direct dependency on `internal/network` (fixes layer boundary violation)
  - Maintains 100% API compatibility (no breaking changes)
  - Improved testability (can inject `MockTransport` for deterministic tests)

#### Fixed
- **P0: Layer Boundary Violation** ([Issue #P0-001](specs/003-m1-refactoring/tasks.md))
  - `querier` package no longer imports `internal/network`
  - Proper layer separation: `querier` → `transport` → `network`
  - Complies with F-2 architecture specification

- **P0: Error Swallowing in Close()** (FR-004 Compliance)
  - `UDPv4Transport.Close()` now propagates socket close errors
  - `Querier.Close()` propagates transport close errors
  - Errors wrapped in `errors.NetworkError` for context
  - End-to-end validation via T063-T064 tests

#### Performance
- **Zero Abstraction Overhead**
  - Query latency: **179 ns/op → 163 ns/op** (9% faster!)
  - Transport interface adds no measurable overhead
  - Better cache locality from refactored code

- **Buffer Pool Impact**
  - Receive path allocations: **9000 B/op → 48 B/op** (99% reduction)
  - Exceeds FR-003 target of ≥80% allocation reduction
  - GC pressure reduced from 900 KB/sec to near-zero

#### Testing
- **Test Coverage**: 84.8% (maintained from 83.9% baseline)
- **New Tests Added**:
  - 13 unit tests for `UDPv4Transport` (T011-T016, T044-T048, T063)
  - 3 contract tests for `Transport` interface (T017-T019)
  - 2 integration tests for error propagation (T027-T028, T064)
- **Test Quality**:
  - All tests pass with race detector (`-race`)
  - Fuzz testing: 1000 executions, zero panics
  - **Fixed 3 flaky tests** (now 100% stable):
    - `TestUDPv4Transport_ReceiveReturnsBufferToPool`: Resilient to real mDNS traffic
    - `TestUDPv4Transport_Receive_PropagatesContextDeadline`: Resilient to real mDNS traffic
    - `TestQuery_RealNetwork_Timeout`: Added 100ms jitter tolerance (see [ADR-003](docs/decisions/003-integration-test-timing-tolerance.md))
- **Test Suite Stability**: 9/9 packages PASS (was 8/9 with 1 flaky test)

#### Technical Details
- **TDD Methodology**: STRICT RED → GREEN → REFACTOR
  - All 97 tasks completed following test-first approach
  - Tests written before implementation (validates no false positives)
- **Context Propagation**:
  - `Send()` and `Receive()` accept `context.Context`
  - Context deadlines propagate to socket `SetReadDeadline()`
  - Supports cancellation via `ctx.Done()` channel
- **M1.1 Alignment**:
  - F-9 REQ-F9-7: Context propagation implemented ✅
  - F-9 REQ-F9-1: `UDPv4Transport` extensible via `ListenConfig` ✅
  - F-9 REQ-F9-2: Platform-specific socket options supported ✅

#### Documentation
- Created [ADR-001: Transport Interface Abstraction](docs/decisions/001-transport-interface-abstraction.md)
- Created [ADR-002: Buffer Pooling Pattern](docs/decisions/002-buffer-pooling-pattern.md)
- Summarised final metrics in [PLAN_COMPLETE.md](specs/003-m1-refactoring/PLAN_COMPLETE.md)
- Updated godoc for all new packages

#### Migration Guide
**No migration needed** - Public API is unchanged. Internal refactoring only.

If you were using internal packages (not recommended):
- Replace `internal/network.CreateSocket()` with `transport.NewUDPv4Transport()`
- Replace direct `net.PacketConn` usage with `transport.Transport` interface
- See [ADR-001](docs/decisions/001-transport-interface-abstraction.md) for details

#### Future Impact
This refactoring enables:
- **M1.1**: Context-aware service discovery, improved error handling
- **M2**: IPv6 dual-stack support via `UDPv6Transport`
- **M3+**: Alternative transports (TCP, QUIC) by implementing `Transport` interface

#### References
- **Milestone Plan**: [specs/003-m1-refactoring/plan.md](specs/003-m1-refactoring/plan.md)
- **Tasks**: [specs/003-m1-refactoring/tasks.md](specs/003-m1-refactoring/tasks.md) (97/97 complete)
- **Completion Report**: [PLAN_COMPLETE.md](specs/003-m1-refactoring/PLAN_COMPLETE.md)
- **Benchmarks**: [baseline_metrics.md](specs/003-m1-refactoring/baseline_metrics.md)

---

## [0.1.0] - 2025-10-XX - M1 Milestone (Initial Release)

### Added
- **Core mDNS Query Functionality**
  - IPv4 UDP multicast support (224.0.0.251:5353)
  - Record type support: A, PTR, SRV, TXT
  - RFC 6762 compliant query/response handling
  - Context-aware timeout support

- **Querier Package** (Public API)
  - `querier.New()`: Create new querier instance
  - `querier.Query()`: Execute mDNS query with timeout
  - `querier.Close()`: Release resources
  - Support for 100+ concurrent queries (NFR-002)

- **Internal Packages**
  - `internal/protocol`: mDNS protocol constants (port, multicast address)
  - `internal/message`: DNS message parsing and building (RFC 1035, RFC 6762)
  - `internal/network`: UDP socket creation and I/O
  - `internal/errors`: Custom error types (NetworkError, ValidationError, WireFormatError)

- **Testing**
  - Contract tests for RFC compliance
  - Integration tests with real network
  - Fuzz testing for message parser
  - Benchmark suite for performance validation

### Performance
- Query processing overhead: <100ms (NFR-001)
- Supports 100 concurrent queries without resource leaks (NFR-002)
- Test coverage: 83.9%

---

## Legend

### Change Categories
- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Vulnerability fixes
- **Performance**: Performance improvements

### Priority Markers
- **P0**: Critical (blocks milestone)
- **P1**: High (should fix)
- **P2**: Medium (nice to have)

---

[Unreleased]: https://github.com/joshuafuller/beacon/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/joshuafuller/beacon/releases/tag/v0.1.0
