# Tasks: Interface-Specific IP Address Advertising

**Input**: Design documents from `/specs/007-interface-specific-addressing/`
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/)

**Tests**: TDD approach - tests written FIRST (RED → GREEN → REFACTOR cycle)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Beacon uses single library project structure:
- Implementation: `internal/transport/`, `responder/`, `internal/`
- Tests: `tests/contract/`, `tests/integration/`, `*_test.go` files
- Documentation: `docs/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Foundation for interface-specific addressing implementation

- [ ] T001 Review current Transport interface implementation in internal/transport/transport.go
- [ ] T002 Review current UDPv4Transport implementation in internal/transport/udp.go
- [ ] T003 [P] Review current responder query handling in responder/responder.go (lines 605-652)
- [ ] T004 [P] Review MockTransport implementation in internal/transport/mock_transport.go

**Checkpoint**: Understanding of current architecture complete

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core transport layer changes that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Transport Interface Enhancement

- [ ] T005 Update Transport interface in internal/transport/transport.go to add interfaceIndex return value to Receive()
- [ ] T006 Add golang.org/x/net/ipv4 import to internal/transport/udp.go
- [ ] T007 Add ipv4Conn field to UDPv4Transport struct in internal/transport/udp.go
- [ ] T008 Update NewUDPv4Transport() in internal/transport/udp.go to wrap connection with ipv4.NewPacketConn()
- [ ] T009 Call SetControlMessage(ipv4.FlagInterface, true) in NewUDPv4Transport() to enable interface index extraction
- [ ] T010 Update UDPv4Transport.Receive() in internal/transport/udp.go to use ipv4Conn.ReadFrom() instead of conn.ReadFrom()
- [ ] T011 Extract cm.IfIndex from control message in UDPv4Transport.Receive() and return as interfaceIndex (0 if unavailable)

### MockTransport Update

- [ ] T012 Add InterfaceIndex field to ReceiveResponse struct in internal/transport/mock_transport.go
- [ ] T013 Update MockTransport.Receive() to return interfaceIndex from ReceiveResponse (0 as default)

### Interface Resolver Function

- [ ] T014 Add getIPv4ForInterface(ifIndex int) function to responder/responder.go (after getLocalIPv4)
- [ ] T015 Implement interface lookup using net.InterfaceByIndex(ifIndex) in getIPv4ForInterface()
- [ ] T016 Implement address lookup using iface.Addrs() in getIPv4ForInterface()
- [ ] T017 Filter for first IPv4 address in getIPv4ForInterface()
- [ ] T018 Add NetworkError return for interface not found in getIPv4ForInterface()
- [ ] T019 Add ValidationError return for no IPv4 address in getIPv4ForInterface()
- [ ] T020 Add godoc comment to getIPv4ForInterface() citing RFC 6762 §15

**Checkpoint**: Foundation ready - Transport returns interface index, resolver function ready

---

## Phase 3: User Story 1 - Laptop with WiFi and Ethernet (Priority: P1) 🎯 MVP

**Goal**: Multi-interface machines respond to queries with interface-specific IP addresses. Query on eth0 → eth0 IP, query on wlan0 → wlan0 IP.

**Independent Test**: Register service on machine with 2+ interfaces, send queries from each network, verify responses contain ONLY the interface-specific IP.

### Tests for User Story 1 (TDD - Write Tests FIRST)

> **RED Phase**: Write these tests, ensure they FAIL before implementation

- [ ] T021 [P] [US1] Create tests/contract/rfc6762_interface_test.go with package and imports
- [ ] T022 [US1] Write TestRFC6762_Section15_InterfaceSpecificAddresses test skeleton in tests/contract/rfc6762_interface_test.go
- [ ] T023 [US1] Add scenario 1 to contract test: Query on interface 1 → MUST include interface 1 IP, MUST NOT include interface 2 IP
- [ ] T024 [US1] Add scenario 2 to contract test: Query on interface 2 → MUST include interface 2 IP, MUST NOT include interface 1 IP
- [ ] T025 [US1] Add scenario 3 to contract test: Single-interface regression - verify behavior unchanged
- [ ] T026 [US1] Run contract test and verify it FAILS (RED phase complete)

### Implementation for User Story 1 (GREEN Phase)

- [ ] T027 [US1] Update Responder.listenForQueries() in responder/responder.go to extract interfaceIndex from transport.Receive()
- [ ] T028 [US1] Pass interfaceIndex to handleQuery() or make available via closure in responder/responder.go
- [ ] T029 [US1] Update handleQuery() in responder/responder.go to call getIPv4ForInterface(interfaceIndex) instead of getLocalIPv4()
- [ ] T030 [US1] Add graceful fallback: if interfaceIndex == 0, call getLocalIPv4() (degraded mode)
- [ ] T031 [US1] Add error handling: if getIPv4ForInterface() fails, skip response for that query
- [ ] T032 [US1] Add debug logging for interface-specific IP lookup (ifIndex, IP, errors)
- [ ] T033 [US1] Run contract test and verify it PASSES (GREEN phase complete)

### Refactoring for User Story 1 (REFACTOR Phase)

- [ ] T034 [US1] Review getIPv4ForInterface() error messages for clarity
- [ ] T035 [US1] Review handleQuery() logic for readability
- [ ] T036 [US1] Add inline comments citing RFC 6762 §15 where interface-specific IP is used
- [ ] T037 [US1] Update getLocalIPv4() godoc to mark as "DEPRECATED for response building" (still used for registration)

**Checkpoint**: User Story 1 complete - Multi-interface advertising works, contract test passes, single-interface regression verified

---

## Phase 4: User Story 2 - Server with Multiple NICs (Priority: P2)

**Goal**: Production servers with multiple VLANs advertise correct IP per VLAN. Query on VLAN1 → VLAN1 IP only, query on VLAN2 → VLAN2 IP only.

**Independent Test**: Register service on multi-NIC server (3+ interfaces), query from each VLAN, verify responses contain ONLY the VLAN-specific IP.

### Tests for User Story 2

> **RED Phase**: Write tests for multi-NIC server scenario

- [ ] T038 [P] [US2] Create tests/integration/multi_interface_test.go with package and imports
- [ ] T039 [US2] Write TestMultiNICServer_VLANIsolation test skeleton in tests/integration/multi_interface_test.go
- [ ] T040 [US2] Add scenario: 3-NIC server (10.0.1.10, 10.0.2.10, 10.0.3.10) - query on VLAN1 returns only VLAN1 IP
- [ ] T041 [US2] Add scenario: Same server - query on VLAN2 returns only VLAN2 IP
- [ ] T042 [US2] Add scenario: Verify connection failure when wrong IP advertised (validates fix)
- [ ] T043 [US2] Run integration test and verify it FAILS (RED phase complete)

### Implementation for User Story 2 (GREEN Phase)

- [ ] T044 [US2] Verify getIPv4ForInterface() handles multiple NIC scenarios correctly (code review)
- [ ] T045 [US2] Add unit test TestGetIPv4ForInterface_MultipleNICs in responder/responder_test.go
- [ ] T046 [US2] Test edge case: Interface index out of range → NetworkError
- [ ] T047 [US2] Test edge case: Interface exists but no IPv4 → ValidationError
- [ ] T048 [US2] Run all tests (contract + integration) and verify they PASS (GREEN phase complete)

### Refactoring for User Story 2

- [ ] T049 [US2] Review error handling paths for multi-NIC failures
- [ ] T050 [US2] Add performance measurement: Benchmark getIPv4ForInterface() lookup time

**Checkpoint**: User Story 2 complete - Multi-NIC servers work correctly, VLAN isolation verified

---

## Phase 5: User Story 3 - Docker and VPN Interfaces (Priority: P3)

**Goal**: Systems with Docker/VPN interfaces exclude virtual interfaces from mDNS advertising. Query on physical interface → physical IP only (not Docker/VPN IPs).

**Independent Test**: Register service on machine with physical + Docker + VPN interfaces, query from physical network, verify response contains ONLY physical interface IP.

### Tests for User Story 3

> **RED Phase**: Write tests for Docker/VPN interface exclusion

- [ ] T051 [P] [US3] Add TestDockerVPNExclusion test to tests/integration/multi_interface_test.go
- [ ] T052 [US3] Add scenario: Machine with physical (192.168.1.100), Docker (172.17.0.1), VPN (10.8.0.2) - query on physical returns physical IP only
- [ ] T053 [US3] Add scenario: Verify Docker/VPN interfaces not advertised (leverages F-10 interface selection)
- [ ] T054 [US3] Add scenario: Docker container querying host → response includes Docker bridge IP (special case)
- [ ] T055 [US3] Run integration test and verify it FAILS (RED phase complete)

### Implementation for User Story 3 (GREEN Phase)

- [ ] T056 [US3] Verify F-10 interface selection logic still works with new interface-specific addressing (code review)
- [ ] T057 [US3] Test interaction: DefaultInterfaces() filtering + getIPv4ForInterface() lookup
- [ ] T058 [US3] Verify Docker/VPN interfaces are excluded during transport initialization (F-10 behavior preserved)
- [ ] T059 [US3] Run all tests and verify they PASS (GREEN phase complete)

### Refactoring for User Story 3

- [ ] T060 [US3] Document interaction between F-10 interface selection and interface-specific IP lookup
- [ ] T061 [US3] Add example to godoc showing Docker/VPN exclusion behavior

**Checkpoint**: User Story 3 complete - Docker/VPN interfaces correctly excluded, F-10 compatibility verified

---

## Phase 6: Test Coverage & Validation

**Purpose**: Ensure comprehensive test coverage and validate all scenarios

### Unit Tests

- [ ] T062 [P] Write TestGetIPv4ForInterface_ValidInterface in responder/responder_test.go
- [ ] T063 [P] Write TestGetIPv4ForInterface_InterfaceNotFound in responder/responder_test.go
- [ ] T064 [P] Write TestGetIPv4ForInterface_NoIPv4Address in responder/responder_test.go
- [ ] T065 [P] Write TestGetIPv4ForInterface_MultipleIPv4Addresses (returns first) in responder/responder_test.go
- [ ] T066 [P] Write TestUDPv4Transport_ReceiveWithInterface in internal/transport/udp_test.go
- [ ] T067 [P] Write TestUDPv4Transport_ControlMessageUnavailable (graceful degradation) in internal/transport/udp_test.go
- [ ] T068 Write TestHandleQuery_InterfaceSpecificIP in responder/responder_test.go
- [ ] T069 Write TestHandleQuery_FallbackToGlobalIP (when ifIndex == 0) in responder/responder_test.go
- [ ] T070 Write TestHandleQuery_SkipResponseOnError (when interface lookup fails) in responder/responder_test.go

### MockTransport Test Updates

- [ ] T071 Update all existing tests using MockTransport to handle 4-value Receive() return
- [ ] T072 Add mock interface index values to test fixtures (e.g., 1 for eth0, 2 for wlan0)
- [ ] T073 Verify all responder tests pass with updated MockTransport

### Performance Benchmarks

- [ ] T074 [P] Create BenchmarkGetIPv4ForInterface in responder/responder_test.go
- [ ] T075 [P] Create BenchmarkHandleQuery_WithInterfaceContext in responder/responder_test.go
- [ ] T076 Run benchmarks and compare to baseline (from 006-mdns-responder/PERFORMANCE_ANALYSIS.md)
- [ ] T077 Verify performance overhead ≤10% (NFR-002 from spec.md)

### Regression Tests

- [ ] T078 Run all existing responder tests and verify no regressions
- [ ] T079 Run all existing transport tests and verify no regressions
- [ ] T080 Run full test suite with race detector: go test ./... -race
- [ ] T081 Verify test coverage ≥80% for modified files: go test -cover ./responder ./internal/transport

**Checkpoint**: All tests pass, coverage targets met, no performance regressions

---

## Phase 7: Documentation & Compliance

**Purpose**: Update documentation to reflect interface-specific addressing behavior

- [ ] T082 [P] Update responder/responder.go godoc for Responder to document interface-specific behavior
- [ ] T083 [P] Update internal/transport/transport.go godoc for Transport.Receive() to document interfaceIndex return
- [ ] T084 [P] Update internal/transport/udp.go godoc for UDPv4Transport to document control message usage
- [ ] T085 Update docs/RFC_COMPLIANCE_GUIDE.md to document RFC 6762 §15 compliance
- [ ] T086 Add section to RFC_COMPLIANCE_GUIDE.md showing example: Query on eth0 → eth0 IP in response
- [ ] T087 Add note about graceful degradation when control messages unavailable
- [ ] T088 Update CLAUDE.md §Recent Changes to document 007-interface-specific-addressing completion
- [ ] T089 Update CLAUDE.md §Active Technologies to note golang.org/x/net/ipv4 usage

**Checkpoint**: Documentation complete, RFC compliance documented

---

## Phase 8: Integration & Manual Validation

**Purpose**: Manual testing on real multi-interface hardware

### Manual Test Plan

- [ ] T090 Setup manual test environment: Laptop with WiFi + Ethernet on different networks
- [ ] T091 Register mDNS service on multi-interface laptop
- [ ] T092 Send query from WiFi network, capture response, verify WiFi IP advertised
- [ ] T093 Send query from Ethernet network, capture response, verify Ethernet IP advertised
- [ ] T094 Verify connection succeeds to correct IP from each network
- [ ] T095 Test single-interface machine (regression): Verify existing behavior preserved
- [ ] T096 Document manual test results in specs/007-interface-specific-addressing/MANUAL_TEST_REPORT.md

### Platform Testing

- [ ] T097 [P] Test on Linux (control messages via IP_PKTINFO)
- [ ] T098 [P] Test on macOS (control messages via IP_RECVIF)
- [ ] T099 [P] Test on Windows (control messages via LPFN_WSARECVMSG)
- [ ] T100 Document platform compatibility results

**Checkpoint**: Manual validation complete, multi-platform verified

---

## Phase 9: Polish & Final Review

**Purpose**: Final cleanup and preparation for merge

### Code Quality

- [ ] T101 Run gofmt on all modified files
- [ ] T102 Run go vet on all modified files
- [ ] T103 Run make semgrep-check and verify no new violations
- [ ] T104 Review all added comments for clarity and RFC citations
- [ ] T105 Review error messages for consistency and helpfulness
- [ ] T106 Check for TODOs and ensure all are tracked (none blocking)

### Final Validation

- [ ] T107 Run make test (full test suite)
- [ ] T108 Run make test-race (with race detector)
- [ ] T109 Run make test-coverage and verify ≥80% for modified packages
- [ ] T110 Verify all acceptance criteria from spec.md §Success Criteria met
- [ ] T111 Create completion report: specs/007-interface-specific-addressing/COMPLETION_REPORT.md

### Issue & PR Preparation

- [ ] T112 Update GitHub Issue #27 with implementation summary
- [ ] T113 Prepare PR description linking spec.md, plan.md, and completion report
- [ ] T114 Add before/after examples to PR description (wrong IP → correct IP)
- [ ] T115 Note RFC 6762 §15 compliance fix in PR
- [ ] T116 Ready for code review and merge

**Checkpoint**: Feature complete, ready for PR and merge

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - Core fix for multi-interface addressing
- **User Story 2 (Phase 4)**: Depends on US1 - Validates multi-NIC server scenario (same implementation, different test)
- **User Story 3 (Phase 5)**: Depends on US1 - Validates Docker/VPN exclusion (same implementation + F-10 compatibility)
- **Test Coverage (Phase 6)**: Depends on US1-US3 - Comprehensive coverage of all scenarios
- **Documentation (Phase 7)**: Can proceed in parallel with Phase 6
- **Integration (Phase 8)**: Depends on Phase 6 completion - Manual validation
- **Polish (Phase 9)**: Depends on all previous phases - Final cleanup

### User Story Dependencies

- **User Story 1 (P1)**: Core implementation - MUST complete first
  - Implements interface-specific IP lookup
  - Contract test for RFC 6762 §15 compliance
  - Single-interface regression verification

- **User Story 2 (P2)**: Extends US1 testing to multi-NIC servers
  - Same code as US1, different test scenario
  - Verifies VLAN isolation behavior
  - Can start after US1 implementation complete

- **User Story 3 (P3)**: Validates interaction with F-10 interface selection
  - Leverages US1 implementation
  - Tests Docker/VPN exclusion
  - Verifies F-10 compatibility preserved

### Within Each User Story (TDD Cycle)

1. **RED**: Write tests FIRST, verify they FAIL
2. **GREEN**: Implement minimum code to make tests PASS
3. **REFACTOR**: Clean up, improve clarity, add documentation

### Parallel Opportunities

**Phase 1 (Setup)**: T001-T004 can all run in parallel (read-only review tasks)

**Phase 2 (Foundational)**: Limited parallelism due to dependencies
- T005-T011: Sequential (Transport interface changes)
- T012-T013: Can run in parallel after T011 (MockTransport update)
- T014-T020: Sequential (Interface resolver implementation)

**Phase 3 (US1 Tests)**: T021-T026 contract test tasks can be parallelized

**Phase 6 (Test Coverage)**:
- T062-T067: All unit tests can run in parallel (different test functions)
- T074-T075: Benchmarks can run in parallel
- T078-T081: Regression/coverage checks sequential

**Phase 7 (Documentation)**: T082-T084 can run in parallel (different files)

**Phase 8 (Platform Testing)**: T097-T099 can run in parallel (different platforms)

---

## Parallel Example: User Story 1 (Phase 3)

```bash
# RED Phase - Write tests in parallel:
Task T021: "Create tests/contract/rfc6762_interface_test.go" [P]
Task T022: "Write TestRFC6762_Section15_InterfaceSpecificAddresses skeleton"
# Then T023-T026 complete test sequentially

# After GREEN Phase complete, REFACTOR tasks can run in parallel:
Task T034: "Review getIPv4ForInterface() error messages" [P]
Task T035: "Review handleQuery() logic" [P]
Task T036: "Add RFC citations" [P]
```

## Parallel Example: Phase 6 (Test Coverage)

```bash
# All unit tests can run in parallel:
Task T062: "TestGetIPv4ForInterface_ValidInterface" [P]
Task T063: "TestGetIPv4ForInterface_InterfaceNotFound" [P]
Task T064: "TestGetIPv4ForInterface_NoIPv4Address" [P]
Task T065: "TestGetIPv4ForInterface_MultipleIPv4Addresses" [P]
Task T066: "TestUDPv4Transport_ReceiveWithInterface" [P]
Task T067: "TestUDPv4Transport_ControlMessageUnavailable" [P]
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (understand current code)
2. Complete Phase 2: Foundational (transport + resolver - CRITICAL)
3. Complete Phase 3: User Story 1 (core fix + contract test)
4. **STOP and VALIDATE**: Run contract test, verify RFC 6762 §15 compliance
5. Ready for basic multi-interface fix (stops here if fast-track needed)

**Estimated Effort**: ~1-2 days for MVP (P1 only)

### Full Delivery (All User Stories)

1. Complete Phase 1-3 → MVP complete
2. Add Phase 4: User Story 2 → Multi-NIC server validation
3. Add Phase 5: User Story 3 → Docker/VPN compatibility
4. Complete Phase 6-9 → Full test coverage, docs, validation
5. Ready for production merge

**Estimated Effort**: ~2-3 days for full implementation

### Incremental Validation

- After Phase 2: Foundation ready - verify Transport.Receive() returns interface index
- After Phase 3: US1 complete - verify contract test passes
- After Phase 4: US2 complete - verify multi-NIC scenario works
- After Phase 5: US3 complete - verify F-10 compatibility
- After Phase 6: All tests pass - verify coverage ≥80%
- After Phase 8: Manual testing complete - verify real hardware behavior
- After Phase 9: Ready for merge

---

## Success Criteria (From spec.md)

- [ ] **SC-001**: Multi-interface machines respond with interface-specific IP (contract test passes)
- [ ] **SC-002**: RFC 6762 §15 compliance verified (MUST include interface IP, MUST NOT include other IPs)
- [ ] **SC-003**: Single-interface regression tests pass (no breaking changes)
- [ ] **SC-004**: Response overhead <10% (benchmark `BenchmarkHandleQuery_WithInterfaceContext`)
- [ ] **SC-005**: API compatibility maintained (no public API changes)
- [ ] **SC-006**: Documentation updated (godoc, RFC_COMPLIANCE_GUIDE.md)

---

## Notes

- **[P]** tasks = different files, no dependencies, can run in parallel
- **[Story]** label maps task to specific user story for traceability
- **TDD Cycle**: RED (failing test) → GREEN (minimal implementation) → REFACTOR (cleanup)
- Each user story is independently testable and delivers value
- US1 is MVP - can stop here for fast-track fix
- US2 and US3 validate edge cases and compatibility
- Commit after each logical task group
- Stop at any checkpoint to validate independently
- Estimated total effort: 2-3 days for full implementation

---

**Task Generation Complete**: 116 tasks organized across 9 phases
**MVP Scope**: Phase 1-3 (T001-T037) delivers core fix
**Full Scope**: All phases (T001-T116) delivers production-ready feature
**Next Step**: Begin with Phase 1 (Setup) review tasks
