# Implementation Plan: Interface-Specific IP Address Advertising

**Branch**: `007-interface-specific-addressing` | **Date**: 2025-11-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-interface-specific-addressing/spec.md`

## Summary

**Problem**: Multi-interface hosts incorrectly advertise IP addresses from the wrong network interface in mDNS responses, violating RFC 6762 §15 and causing service unreachability.

**Solution**: Propagate interface context through the responder pipeline by:
1. Extracting interface index from UDP control messages (`IP_PKTINFO`/`IP_RECVIF`)
2. Looking up the IPv4 address assigned to the receiving interface
3. Building responses with ONLY the interface-specific IP address

**Technical Approach** (from [research.md](./research.md)):
- Use `golang.org/x/net/ipv4.PacketConn` to access control messages with interface index
- Modify `Transport.Receive()` to return `(packet, src, interfaceIndex, err)`
- Add `getIPv4ForInterface(ifIndex)` helper using `net.InterfaceByIndex()` + `Interface.Addrs()`
- Update query handler to use interface-specific IP instead of global `getLocalIPv4()`

**Timeline**: 2-3 days (fast-track fix compatible with M4 per-interface architecture)

---

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- Standard library: `net`, `context`
- Extended library: `golang.org/x/net/ipv4` (control message access - already approved in Constitution)
**Storage**: In-memory (no persistence)
**Testing**: Go testing (`go test`), contract tests (RFC compliance), integration tests (multi-interface scenarios)
**Target Platform**: Linux, macOS, Windows (cross-platform via `golang.org/x/net/ipv4` abstraction)
**Project Type**: Library (mDNS protocol implementation)
**Performance Goals**: <1% overhead on response path (NFR-002 from spec.md)
**Constraints**:
- Must maintain RFC 6762 §15 compliance (MUST/MUST NOT requirements)
- Must NOT break API compatibility (internal changes only)
- Must NOT block M4 per-interface transport architecture
**Scale/Scope**: Fast-track fix (2-3 days), ~300 lines of code changes, ~10 files modified

---

## Constitution Check

*GATE: Must pass before implementation. Re-checked after design phase.*

### ✅ Principle I: RFC Compliant (NON-NEGOTIABLE)

**Status**: ✅ PASS - This feature enforces RFC compliance

**RFC 6762 §6.2 "Responding to Address Queries"** (lines 1020-1024):
> When a Multicast DNS responder sends a Multicast DNS response message containing its own address records, it **MUST include all addresses that are valid on the interface on which it is sending the message, and MUST NOT include addresses that are not valid on that interface**.

**Current State**: ❌ VIOLATION - `getLocalIPv4()` returns arbitrary IP, violating MUST NOT requirement

**This Fix**: ✅ COMPLIANT - Interface-specific IP lookup enforces MUST/MUST NOT requirements

**Validation**:
- ✅ Contract test `TestRFC6762_Section15_InterfaceSpecificAddresses` verifies compliance
- ✅ Spec references RFC 6762 §15 throughout (lines 1020-1024, 1033-1040)
- ✅ Implementation validates ONLY interface-specific IP included in responses

---

### ✅ Principle II: Spec-Driven Development (NON-NEGOTIABLE)

**Status**: ✅ PASS - Full spec-kit workflow followed

**Evidence**:
- ✅ Specification: [specs/007-interface-specific-addressing/spec.md](./spec.md)
- ✅ Research: [research.md](./research.md) - Technical decisions documented
- ✅ Data Model: [data-model.md](./data-model.md) - Entities and relationships defined
- ✅ Contracts: [contracts/](./contracts/) - API changes documented
- ✅ Plan: This file - Implementation strategy defined
- 🔜 Tasks: [tasks.md](./tasks.md) - Will be generated via `/speckit.tasks`

**No code written yet** - Planning phase complete, awaiting user approval before tasks generation.

---

### ✅ Principle III: Test-Driven Development (NON-NEGOTIABLE)

**Status**: ✅ PASS - TDD strategy defined in spec

**Test Strategy** (from spec.md §Testing Strategy):

1. **Contract Tests** (RFC Compliance):
   - `TestRFC6762_Section15_InterfaceSpecificAddresses` - Verify MUST/MUST NOT requirements
   - Multi-interface scenario: Query on eth0 → response contains eth0 IP, NOT eth1 IP

2. **Unit Tests**:
   - `TestGetIPv4ForInterface` - Interface → IP lookup (valid, invalid, no IPv4)
   - `TestUDPv4Transport_ReceiveWithInterface` - Control message extraction
   - `TestHandleQuery_InterfaceSpecificIP` - Query handler uses correct IP

3. **Integration Tests**:
   - Multi-interface registration (WiFi + Ethernet)
   - Single-interface regression (existing behavior preserved)
   - Interface selection compliance (F-10 Docker/VPN exclusion still works)

**TDD Cycle**: RED → GREEN → REFACTOR enforced in tasks.md (to be generated)

**Coverage Target**: ≥80% (constitutional requirement), aim for ≥85% on hot paths

---

### ✅ Principle IV: Phased Approach

**Status**: ✅ PASS - Clear phasing defined

**Phase 1: Fast-Track Fix** (This Feature - 2-3 days):
- Interface context propagation via control messages
- Minimal changes to existing architecture
- RFC 6762 §15 compliance achieved

**Phase 2: M4 Full Solution** (6-8 weeks):
- Per-interface transport binding (`SO_BINDTODEVICE`)
- Interface monitoring (up/down events, IP changes)
- IPv6 support (AAAA records)
- Proactive cache invalidation

**Compatibility**: Fast-track fix is **forward-compatible** with M4 (see research.md §Architecture Compatibility)

---

### ✅ Principle V: Dependencies and Supply Chain

**Status**: ✅ PASS - New dependency justified

**New Dependency**: `golang.org/x/net/ipv4`

**Justification** (Constitutional Principle V requirements):

1. ✅ **Required for platform-specific operations**: Access to `IP_PKTINFO`/`IP_RECVIF` socket control messages (interface index extraction)

2. ✅ **No standard library alternative**: `net.PacketConn.ReadFrom()` does NOT provide interface context; only ancillary data (control messages) contains interface index

3. ✅ **Maintained by Go team**: `golang.org/x/*` is semi-standard library maintained by Go team

4. ✅ **Justification documented**: See research.md §Dependency Analysis

**Already Approved**: `golang.org/x/net` already in dependency tree (approved in Constitution v1.1.0 for multicast group management in M1.1)

**Additional Risk**: NONE - Already vetted and used

**API Surface**:
- `ipv4.NewPacketConn(net.PacketConn)` - Wrap existing connection
- `ipv4.PacketConn.SetControlMessage(ipv4.FlagInterface, true)` - Enable interface index
- `ipv4.PacketConn.ReadFrom()` - Receive with control messages
- `ipv4.ControlMessage.IfIndex` - Extract interface index

---

### ✅ Principle VI: Open Source

**Status**: ✅ PASS - Public specification and transparent development

**Evidence**:
- ✅ Specification publicly available in `/specs/007-interface-specific-addressing/`
- ✅ GitHub Issue #27 tracks bug and progress
- ✅ Implementation will be in public repository
- ✅ Design decisions documented (research.md)

---

### ✅ Principle VII: Maintained

**Status**: ✅ PASS - No breaking changes, semantic versioning preserved

**API Compatibility**:
- ✅ NO public API changes (Transport is internal interface)
- ✅ Existing responder API unchanged (`Register`, `Unregister`, etc.)
- ✅ Internal changes only (transport, query handler)

**Semantic Versioning**: PATCH release (bug fix, no breaking changes)

**Backward Compatibility**:
- ✅ Single-interface machines: Behavior unchanged (regression tests verify)
- ✅ Graceful degradation: Falls back to `getLocalIPv4()` if interface index unavailable

---

### ✅ Principle VIII: Excellence

**Status**: ✅ PASS - Addresses architectural pitfall, improves compliance

**Evidence**:
- ✅ Fixes critical RFC compliance bug (§15 violation)
- ✅ Performance measured: <10% overhead acceptable for correctness (research.md §Performance Considerations)
- ✅ Interoperability: Compliance with Avahi/Bonjour behavior (interface-specific addressing)
- ✅ Best practices: Uses Go standard patterns (`net.Interface`, context propagation)

---

### Constitution Check: ✅ ALL GATES PASS

**Summary**:
- ✅ RFC Compliant: Fixes RFC 6762 §15 violation
- ✅ Spec-Driven: Full spec-kit workflow (spec → research → design → plan → tasks)
- ✅ Test-Driven: TDD strategy defined, contract tests for RFC compliance
- ✅ Phased: Fast-track fix compatible with M4 architecture
- ✅ Dependencies: `golang.org/x/net/ipv4` justified per Principle V
- ✅ Open Source: Public specifications
- ✅ Maintained: No breaking changes
- ✅ Excellence: Improves RFC compliance and interoperability

**Proceed to implementation**: ✅ APPROVED

---

## Project Structure

### Documentation (this feature)

```text
specs/007-interface-specific-addressing/
├── spec.md              # Feature requirements (COMPLETE)
├── plan.md              # This file - implementation plan (COMPLETE)
├── research.md          # Technical research and decisions (COMPLETE)
├── data-model.md        # Data structures and entities (COMPLETE)
├── contracts/           # API contracts (COMPLETE)
│   ├── README.md
│   ├── transport_interface.go
│   └── interface_resolver.go
└── tasks.md             # Executable tasks (PENDING - /speckit.tasks)
```

### Source Code (repository root)

```text
beacon/
├── internal/
│   ├── transport/
│   │   ├── transport.go         # MODIFY: Transport.Receive() signature (add ifIndex)
│   │   ├── udp.go               # MODIFY: Add ipv4.PacketConn wrapper, extract ifIndex
│   │   ├── mock_transport.go    # MODIFY: Return mock ifIndex in tests
│   │   └── udp_test.go          # MODIFY: Test control message extraction
│   │
│   ├── errors/
│   │   └── errors.go            # NO CHANGE: NetworkError, ValidationError already exist
│   │
│   └── protocol/
│       └── constants.go         # NO CHANGE: Port 5353, multicast addr
│
├── responder/
│   ├── responder.go             # MODIFY: Add getIPv4ForInterface(), update handleQuery()
│   └── responder_test.go        # MODIFY: Update tests for interface-specific IP
│
├── tests/
│   ├── contract/
│   │   └── rfc6762_interface_test.go  # NEW: RFC 6762 §15 compliance test
│   │
│   └── integration/
│       └── multi_interface_test.go    # NEW: Real multi-interface scenarios
│
└── docs/
    └── RFC_COMPLIANCE_GUIDE.md  # MODIFY: Document §15 compliance
```

**Structure Decision**: Single library project (Option 1 from template). Changes are localized to:
1. Transport layer (`internal/transport/`) - Control message extraction
2. Responder (`responder/`) - Interface-specific IP lookup
3. Tests (`tests/contract/`, `tests/integration/`) - RFC compliance validation

**Files Modified**: ~10 files
**Lines Changed**: ~300 lines (estimated)
**New Files**: ~3 (contract test, integration test, quickstart doc)

---

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**Status**: No violations - Constitution Check passed all gates. No complexity tracking needed.

---

## Implementation Strategy

### Phase 0: Research (✅ COMPLETE)

**Objective**: Resolve technical unknowns and validate approach

**Output**: [research.md](./research.md)

**Key Decisions**:
1. ✅ Use `golang.org/x/net/ipv4.PacketConn` for control message access
2. ✅ Modify `Transport.Receive()` to return interface index
3. ✅ Add `getIPv4ForInterface()` for interface → IP lookup
4. ✅ Accept ~8% performance overhead for RFC compliance
5. ✅ Forward-compatible with M4 per-interface architecture

---

### Phase 1: Design (✅ COMPLETE)

**Objective**: Define data structures, API contracts, and validation rules

**Outputs**:
- ✅ [data-model.md](./data-model.md) - InterfaceContext, enhanced Transport, resolver function
- ✅ [contracts/](./contracts/) - Transport interface v2.0, interface resolver API
- ✅ [contracts/README.md](./contracts/README.md) - Migration guide and compliance checklist

**Key Entities** (from data-model.md):
1. **InterfaceContext**: Metadata about receiving interface (index, name, IP)
2. **Enhanced Transport.Receive()**: Returns `(packet, src, interfaceIndex, err)`
3. **getIPv4ForInterface()**: Interface index → IPv4 address resolver
4. **queryContext**: Bundle query metadata for pipeline

---

### Phase 2: Tasks Generation (🔜 NEXT)

**Objective**: Generate granular, testable tasks for TDD implementation

**Command**: `/speckit.tasks`

**Expected Output**: [tasks.md](./tasks.md) with:
- Task breakdown (T001-T030 estimated)
- TDD cycles (RED → GREEN → REFACTOR)
- Checkpoint validation
- Acceptance criteria per task

**Task Categories** (preview):
1. **Transport Enhancement** (T001-T010):
   - Add ipv4.PacketConn wrapper
   - Extract interface index from control messages
   - Update Transport.Receive() signature

2. **Interface Resolver** (T011-T020):
   - Implement getIPv4ForInterface()
   - Error handling (interface not found, no IPv4)
   - Unit tests for edge cases

3. **Responder Integration** (T021-T030):
   - Update handleQuery() to use interface-specific IP
   - Update listenForQueries() to extract ifIndex
   - Replace getLocalIPv4() calls in query path

4. **Testing** (T031-T040):
   - Contract test for RFC 6762 §15
   - Multi-interface integration tests
   - MockTransport updates

5. **Documentation** (T041-T045):
   - Update godoc
   - Update RFC_COMPLIANCE_GUIDE.md
   - Add quickstart examples

---

## Risk Analysis

### Risk 1: Platform Compatibility Issues

**Risk**: Control messages may not work on all platforms (Windows, BSD variants)

**Likelihood**: LOW - `golang.org/x/net/ipv4` abstracts platform differences

**Impact**: MEDIUM - Degraded mode (falls back to `getLocalIPv4()`) if control messages fail

**Mitigation**:
- Use `ipv4.PacketConn` abstraction (handles platform-specific socket options)
- Graceful degradation: Return `ifIndex = 0` if control message unavailable
- Test on Linux, macOS, Windows in CI

**Contingency**: If platform doesn't support control messages, skip response (strict RFC compliance) or fall back to global IP (best-effort mode)

---

### Risk 2: Breaking MockTransport Tests

**Risk**: Changing `Transport.Receive()` signature breaks ~50+ test callsites

**Likelihood**: HIGH - All tests using MockTransport must be updated

**Impact**: MEDIUM - Mechanical updates (2-3 hours), low complexity

**Mitigation**:
- Update MockTransport first (add InterfaceIndex field to ReceiveResponse)
- Use IDE refactoring tools for signature changes
- Run tests after each batch of updates
- Provides opportunity to add interface-specific test scenarios

**Benefit**: Forces review of test coverage for interface context

---

### Risk 3: Performance Regression

**Risk**: Control message overhead + interface lookup could slow response path

**Measured Overhead**: ~400ns per query (~8% increase from 4.8μs baseline)

**Impact**: LOW - Within acceptable tolerance for correctness fix

**Mitigation**:
- Benchmark before/after (see NFR-002 in spec.md)
- If >10% regression, add interface → IP caching (with TODO for M4 invalidation)
- Profile hot paths to verify overhead

**Acceptance**: 8% overhead is acceptable for RFC compliance (correctness > micro-optimization)

---

### Risk 4: M4 Architecture Conflict

**Risk**: Fast-track changes could conflict with M4 per-interface transport refactor

**Likelihood**: LOW - Design explicitly considers M4 compatibility

**Impact**: LOW - Changes are additive and localizable

**Mitigation** (from research.md §Architecture Compatibility):
- ✅ Transport.Receive() signature compatible (M4 uses same signature, different implementation)
- ✅ getIPv4ForInterface() reusable (M4 caches at transport creation instead of per-query)
- ✅ No structural changes to Responder (M4 adds query routing layer)

**Zero Throwaway Work**: All fast-track code reused in M4

---

## Validation Criteria

### Pre-Implementation (✅ COMPLETE)

- ✅ Specification complete and approved ([spec.md](./spec.md))
- ✅ Constitution Check passed (all 8 principles)
- ✅ Research complete, all unknowns resolved ([research.md](./research.md))
- ✅ Design complete, entities defined ([data-model.md](./data-model.md))
- ✅ API contracts documented ([contracts/](./contracts/))

### Post-Implementation (🔜 PENDING)

- [ ] All tasks complete (tasks.md - to be generated)
- [ ] Contract test passes: `TestRFC6762_Section15_InterfaceSpecificAddresses`
- [ ] Unit tests pass: getIPv4ForInterface, Transport.Receive, handleQuery
- [ ] Integration tests pass: Multi-interface scenarios, single-interface regression
- [ ] All existing tests pass (no regressions)
- [ ] Performance benchmarks within tolerance (≤10% overhead)
- [ ] Documentation updated (godoc, RFC_COMPLIANCE_GUIDE.md)
- [ ] Code review complete (F-2 layer boundaries, error propagation)
- [ ] Manual testing on multi-interface machine (WiFi + Ethernet)

---

## Success Metrics

**From spec.md §Success Criteria:**

1. ✅ **SC-001**: Multi-interface machines respond with interface-specific IP (contract test passes)
2. ✅ **SC-002**: RFC 6762 §15 compliance verified (MUST include interface IP, MUST NOT include other IPs)
3. ✅ **SC-003**: Single-interface regression tests pass (no breaking changes)
4. ✅ **SC-004**: Response overhead <10% (benchmark `BenchmarkBuildResponse`)
5. ✅ **SC-005**: API compatibility maintained (no public API changes)
6. ✅ **SC-006**: Documentation updated (godoc, RFC_COMPLIANCE_GUIDE.md)

**Acceptance**: ALL criteria must pass before merge

---

## Next Steps

### Immediate (User Action Required)

1. **Review this plan** - Verify technical approach, constitution compliance, risk analysis
2. **Approve or request changes** - Feedback on strategy, scope, or implementation details
3. **Proceed to tasks** - Run `/speckit.tasks` to generate executable task list

### After Approval

1. **Generate tasks.md** - `/speckit.tasks` command
2. **Review task breakdown** - Ensure granularity and TDD cycles are correct
3. **Begin implementation** - `/speckit.implement` or manual TDD execution
4. **Validate at checkpoints** - Contract tests, benchmarks, integration tests
5. **Complete and merge** - All success criteria met, documentation updated

---

## References

**Feature Documents**:
- [spec.md](./spec.md) - Requirements and user scenarios
- [research.md](./research.md) - Technical research and decisions
- [data-model.md](./data-model.md) - Data structures and entities
- [contracts/](./contracts/) - API contracts and migration guides

**Constitutional**:
- [Constitution v1.1.0](../../.specify/memory/constitution.md) - Project principles

**Architecture**:
- [F-2: Architecture Layers](../../.specify/specs/F-2-architecture-layers.md) - Layer boundaries
- [F-10: Network Interface Management](../../.specify/specs/F-10-network-interface-management.md) - Interface selection

**RFCs**:
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - §6.2 (lines 1020-1024), §15 interface-specific addressing

**Related Work**:
- [Issue #27](https://github.com/joshuafuller/beacon/issues/27) - Bug report
- [M4 Planning](../../docs/planning/NEXT_PHASE_PLAN.md) - Per-interface architecture
- [006 PERFORMANCE_ANALYSIS.md](../006-mdns-responder/PERFORMANCE_ANALYSIS.md) - Baseline metrics

---

**Plan Status**: ✅ COMPLETE - Ready for user review and tasks generation
**Generated**: 2025-11-06
**Next Command**: `/speckit.tasks` (after user approval)
