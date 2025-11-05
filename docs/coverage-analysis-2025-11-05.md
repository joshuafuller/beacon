# Code Coverage Analysis Report
**Date**: 2025-11-05
**Branch**: claude/investigate-code-coverage-011CUpRzFA4E8EJSWZqrTena
**Overall Coverage**: 68.5%
**Target**: ≥80% (Constitution), ≥85% (Ideal)
**Status**: ❌ Below Target (-11.5% from minimum, -16.5% from ideal)

---

## Executive Summary

Code coverage has dropped significantly below our targets. However, the situation is nuanced:

**Key Findings:**
1. ✅ **Contract tests ARE working** (36/36 passing) - Protocol compliance validated
2. ❌ **Public API unit tests are sparse** - Critical gap for documentation & maintainability
3. ⚠️ **Query handling untested** - Major functionality gap (integration tests skipped)
4. ℹ️ **Many "0%" lines are test infrastructure** - Actually exercised by contract tests

**Bottom Line:** We need targeted improvements in PUBLIC API testing and query handling, but our protocol compliance is solid.

---

## Coverage by Package

| Package | Coverage | Status | Priority |
|---------|----------|--------|----------|
| `examples/discover` | 0.0% | ⚪ N/A | Low (example code) |
| `internal/network` | 42.0% | ❌ FAILING | Low (legacy, being phased out) |
| `querier` | 54.2% | ❌ Below | **HIGH** (public API) |
| `responder` | 66.8% | ❌ Below | **HIGH** (public API) |
| `internal/transport` | 71.1% | ⚠️ Below | Medium (hot path) |
| `internal/state` | 75.4% | ⚠️ Below | Medium (state machine) |
| `internal/responder` | 77.8% | ⚠️ Below | Medium |
| `internal/message` | 82.0% | ✅ Near | Low |
| `internal/records` | 87.3% | ✅ Good | Low |
| `internal/security` | 92.1% | ✅ Good | Low |
| `internal/errors` | 93.3% | ✅ Good | Low |
| `internal/protocol` | 98.0% | ✅ Excellent | Low |
| **Contract tests** | 36/36 ✅ | All passing | N/A |
| **Fuzz tests** | 109,471 exec | 0 crashes | N/A |

---

## Critical Gaps (HIGH VALUE - Must Fix)

### 1. Querier Public Options (querier/options.go)

**Impact**: Public API, user-facing configuration
**Risk**: Users may hit bugs in uncovered validation logic
**Coverage**: 5 functions at 0%, 1 at 100%

#### Untested Functions:
```go
WithInterfaces()           0% - Interface list validation (empty list check)
WithInterfaceFilter()      0% - Filter function validation (nil check)
WithRateLimit()            0% - Rate limiting toggle
WithRateLimitThreshold()   0% - Threshold validation (must be > 0)
WithRateLimitCooldown()    0% - Cooldown validation (must be > 0)
```

**Why This Matters:**
- These functions have validation logic that could fail
- Users rely on clear error messages when misconfigured
- Options pattern is part of our public API contract
- Zero tests = zero documentation of expected behavior

**Recommendation:** Add unit tests for each option covering:
- Happy path (valid input)
- Error cases (nil, zero, negative, empty)
- Error message clarity

**Estimated Effort:** 1-2 hours (straightforward unit tests)

---

### 2. Querier Record Helpers (querier/records.go)

**Impact**: Public API, type safety
**Risk**: Users may receive incorrect data or panics
**Coverage**: 33-66% (only happy paths tested)

#### Poorly Covered Functions:
```go
RecordType.String()        0% - String representation
AsA()                     33% - IPv4 address extraction
AsPTR()                   67% - PTR target extraction
AsSRV()                   33% - SRV data extraction
AsTXT()                   33% - TXT strings extraction
```

**Why This Matters:**
- Type assertion failures not tested (Data field is interface{})
- Wrong-type handling not verified (returns nil/empty)
- Users depend on these for safe type conversions
- Missing tests for malformed data scenarios

**Existing Test Gap:**
```go
// TestResourceRecordAccessors only tests:
ptrRecord.AsA() // PTR record -> AsA() should return nil ✓
ptrRecord.AsPTR() // PTR record -> AsPTR() should return string ✓

// NOT tested:
// - A record with non-IP Data (type assertion failure)
// - SRV record with wrong data type
// - TXT record with non-[]string Data
// - RecordType.String() for all types
```

**Recommendation:** Expand TestResourceRecordAccessors to cover:
- All 4 record types × 4 accessor methods (16 combinations)
- Invalid Data types (type assertion edge cases)
- RecordType.String() for A, PTR, SRV, TXT

**Estimated Effort:** 2-3 hours

---

### 3. Responder Query Handling (responder/responder.go)

**Impact**: CRITICAL - Core responder functionality
**Risk**: Query processing broken, no tests to catch regressions
**Coverage**: 0% for handleQuery() and helpers

#### Untested Functions:
```go
handleQuery()              0% - Main query processing loop
parseMessage()             0% - DNS message parsing
buildResponsePacket()      0% - Response construction
```

**Why This Matters:**
- `handleQuery()` is the MAIN code path for responder query processing
- RFC 6762 §6 compliance depends on this working correctly
- Integration tests are **SKIPPED** with comment: "Query handling not yet implemented"
- Contract tests don't exercise query/response cycle

**Evidence from Integration Test:**
```go
// tests/integration/query_response_test.go:55
t.Skip("Query handling not yet implemented (US3 in progress)")
```

**Current State:**
- Code EXISTS (handleQuery implemented at responder.go:576)
- Integration tests EXIST but are SKIPPED
- No unit tests for handleQuery logic

**Recommendation:**
1. **Immediate**: Unskip integration tests (tests/integration/query_response_test.go)
2. **Short-term**: Add unit tests for handleQuery using MockTransport
3. **Verify**: RFC 6762 §6 query response behavior

**Estimated Effort:** 4-6 hours (includes debugging why tests were skipped)

---

## Low-Priority Gaps (Test Infrastructure - Skip or Defer)

### 4. Responder Test Hooks

**Coverage**: Multiple functions at 0%
```go
OnProbe()                  0%
OnAnnounce()               0%
GetLastProbeMessage()      0%
GetLastAnnounceMessage()   0%
GetLastAnnouncedRecords()  0%
GetLastAnnounceDest()      0%
InjectSimultaneousProbe()  0%
```

**Why This is OK:**
- These are TEST SUPPORT functions, not production code
- Used by contract tests in `tests/contract/`
- Contract tests are 36/36 passing
- Coverage tool shows 0% because contract tests are in different package

**Evidence:**
```bash
$ grep "OnProbe\|GetLastProbeMessage" tests/contract/*.go
rfc6762_probing_test.go:  resp.OnProbe(func() { probeCount++ })
rfc6762_announcing_test.go:  msg := resp.GetLastProbeMessage()
```

**Recommendation:** No action needed. Consider adding package-level comment explaining these are contract test hooks.

---

### 5. Platform Socket Configuration (internal/transport/socket_*.go)

**Coverage**: platformControl() and PlatformControl() at 0%

**Why This is Lower Priority:**
- Used only by legacy `internal/network` package (being phased out)
- Socket options ARE being set (tests would fail on port 5353 otherwise)
- Platform-specific code is hard to test without real sockets
- M1.1 Architectural Hardening validated this works on Linux/macOS/Windows

**Recommendation:** Defer until `internal/network` is fully removed. If needed, add integration test that verifies SO_REUSEPORT is set (check via `/proc/<pid>/fdinfo` on Linux).

---

### 6. Internal State Machine Helpers (internal/state/)

**Coverage**: Multiple getters/setters at 0%
```go
GetProber()                0%
GetAnnouncer()             0%
SetLastProbeMessage()      0%
SetOnSendQuery()           0%
SetRecords()               0%
```

**Why This is OK:**
- Similar to responder test hooks - used by contract tests
- Core state machine logic (Probe, Announce) is well tested (75-94%)
- Only accessor methods are uncovered

**Recommendation:** No action unless state machine bugs found.

---

## Root Cause Analysis

### Why Did Coverage Drop?

1. **M2 Responder Implementation (006-mdns-responder)**
   - Added ~3,000 LOC of new code
   - Focus was on RFC compliance (contract tests) over unit test coverage
   - Integration tests written but SKIPPED (waiting for full implementation?)
   - Test hooks added but not unit-tested (used in contract tests instead)

2. **Public API Testing Deprioritized**
   - Options pattern functions added without corresponding tests
   - Only 1 option tested (WithTimeout) out of 6
   - Record accessor tests are minimal (happy path only)

3. **Test Architecture Choices**
   - Heavy reliance on contract tests (good for RFC compliance)
   - Light on unit tests (bad for API documentation & regression prevention)
   - Integration tests skipped (unclear why - US3 may be incomplete?)

---

## Recommendations (Prioritized)

### Tier 1: Critical (Do This Week)

1. **Unskip and Fix Query Response Integration Tests** (4-6 hours)
   - File: `tests/integration/query_response_test.go`
   - Remove `t.Skip()` calls
   - Fix any failures
   - Verify RFC 6762 §6 compliance
   - **Impact**: +5-10% coverage, validates critical functionality

2. **Add Querier Options Tests** (1-2 hours)
   - File: `querier/querier_test.go`
   - Add TestWithInterfaces, TestWithInterfaceFilter, TestWithRateLimit*, etc.
   - Cover validation error paths
   - **Impact**: +3-5% coverage, documents public API

### Tier 2: Important (Do This Month)

3. **Expand Record Accessor Tests** (2-3 hours)
   - File: `querier/querier_test.go`
   - Expand TestResourceRecordAccessors
   - Test all type combinations and error cases
   - **Impact**: +2-3% coverage, prevents type assertion panics

4. **Add Responder Query Handling Unit Tests** (3-4 hours)
   - File: `responder/responder_test.go`
   - Mock transport to inject test queries
   - Verify handleQuery() logic
   - Test parseMessage() and buildResponsePacket()
   - **Impact**: +4-6% coverage, prevents regressions

### Tier 3: Nice to Have (Future)

5. **Document Test Infrastructure** (1 hour)
   - Add package comments explaining test hooks
   - Update CLAUDE.md with coverage interpretation guide
   - **Impact**: No coverage change, improves maintainability

6. **Platform Socket Tests** (Defer)
   - Only if planning to expand socket options
   - **Impact**: +1-2% coverage

---

## Success Criteria

After implementing Tier 1 + Tier 2 recommendations:

| Metric | Current | Target | Expected After Fixes |
|--------|---------|--------|----------------------|
| Overall Coverage | 68.5% | ≥85% | ~80-85% |
| querier/ | 54.2% | ≥85% | ~75-85% |
| responder/ | 66.8% | ≥85% | ~75-80% |
| Integration Tests | 2 skipped | 0 skipped | 0 skipped |

---

## Testing Philosophy Improvements

### Current Strengths
✅ Excellent RFC compliance testing (contract tests)
✅ Strong fuzz testing (109k executions, 0 crashes)
✅ Good security/validation coverage (92%)

### Current Weaknesses
❌ Sparse public API unit tests
❌ Integration tests not maintained (skipped)
❌ Options pattern under-tested

### Recommended Changes

1. **Adopt "API-First Testing"**
   - Every public function should have unit tests
   - Focus on user-facing behavior, not internals
   - Test validation and error messages explicitly

2. **Keep Integration Tests Green**
   - Never commit skipped integration tests
   - If test needs implementation, mark as `t.Skip()` with issue link
   - Review skipped tests weekly

3. **Separate Test Infrastructure from Production Code**
   - Move test hooks to `testing.go` files
   - Add build tags: `//go:build testing`
   - Don't count test code in coverage

---

## Files to Modify

**High Priority:**
- `querier/querier_test.go` - Add option and record accessor tests
- `tests/integration/query_response_test.go` - Unskip and fix
- `responder/responder_test.go` - Add handleQuery unit tests

**Documentation:**
- `CLAUDE.md` - Update coverage section with interpretation guide
- `docs/TESTING.md` - Document testing philosophy (create if needed)

---

## Notes

- **Contract Tests**: All 36 contract tests passing demonstrates RFC compliance
- **Fuzz Tests**: 109,471 executions with 0 crashes shows robustness
- **False Negatives**: Many "0%" lines are test hooks exercised in different packages
- **Real Issue**: Public API under-tested, query handling untested

**Key Insight:** Coverage number is low, but the quality of TESTED code is high. We just need to expand PUBLIC API testing and unskip integration tests.

---

**Report Generated**: 2025-11-05
**Author**: Coverage Analysis Investigation
**Next Review**: After Tier 1 fixes implemented
