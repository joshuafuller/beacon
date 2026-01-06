# BEACON Production Polish - COMPREHENSIVE TASK LIST

**Status**: 4/4 RFC features implemented, NOW focus on test coverage + quality

Track progress with [x] for completed items.

---

## ✅ P0: CRITICAL RFC FEATURES (COMPLETE)

### T1: Goodbye Packets (RFC 6762 §9.4) ✅ COMPLETE
- [x] Write test: TestUnregister_SendsGoodbyePackets
- [x] Implement BuildGoodbyeRecords() in internal/records/record_set.go
- [x] Update Unregister() in responder/responder.go:266
- [x] Verify implementation

### T2: Source Address Validation (RFC 6762 §6.4) ⚠️ PARTIAL
- [ ] Write test: TestHandleQuery_RejectsWrongSubnet (SKIPPED - TODO)
- [x] Implement validateSourceAddress() function
- [x] Add subnet validation in handleQuery()
- [x] Basic implementation exists

### T3: TC Bit Truncation (RFC 6762 §6.5) ✅ COMPLETE
- [x] Write test: TestBuildResponse_SetsTCBitWhenTruncated
- [x] Set TC bit when response >9KB in BuildResponse()
- [x] Verify implementation

### T4: QU Bit + Unicast Responses (RFC 6762 §5.4) ⚠️ PARTIAL
- [ ] Write test: TestHandleQuery_QUBitUnicastResponse (SKIPPED)
- [x] Parse QU bit from question (0x8000 bitmask)
- [x] Implement unicast vs multicast logic
- [x] Basic implementation exists

---

## 🚨 P1: TEST COVERAGE TO 80%+ (CURRENT: 66.2%)

**Target**: 80%+ overall coverage (+13.8% needed)

### handleQuery Coverage (PRIORITY 1)
Current: Unknown, Target: 80%+

- [ ] Test PTR query handling (_services._dns-sd._udp.local)
- [ ] Test service type PTR query (_http._tcp.local)
- [ ] Test service instance queries (SRV, TXT, A)
- [ ] Test query parsing edge cases (malformed, truncated)
- [ ] Test interface index extraction (007 integration)
- [ ] Test error paths (parse failures, unknown types)
- [ ] Test known-answer suppression logic
- [ ] Test response building with multiple records
- [ ] COMPLETE TestHandleQuery_RejectsWrongSubnet (currently SKIPPED)

### collectResponses Coverage (PRIORITY 2)
Current: 47%, Target: 80%+

- [ ] Test timeout behavior (context cancellation)
- [ ] Test duplicate response handling
- [ ] Test response collection with multiple answers
- [ ] Test edge cases (empty responses, malformed packets)
- [ ] Test context cancellation mid-collection
- [ ] Test buffer pool interaction

### buildARecord Coverage (PRIORITY 3)
Current: Unknown, Target: 90%+

- [ ] Test A record construction with valid IPv4
- [ ] Test error path (invalid IP format)
- [ ] Test with interface-specific IPs (007 integration)

### Public API Coverage (PRIORITY 4)
Current: 0%, Target: 100%

Create `querier/options_test.go`:
- [ ] Test WithInterfaces() - custom interface selection
- [ ] Test WithInterfaceFilter() - interface filtering function
- [ ] Test WithRateLimit() - rate limiting configuration
- [ ] Test WithRateLimitThreshold() - threshold configuration
- [ ] Test WithRateLimitCooldown() - cooldown configuration

Update `responder/responder_test.go`:
- [ ] Test WithTransport() - custom transport injection
- [ ] Test option composition (multiple options together)

### Expected Coverage Impact
- handleQuery: +4.5% (32% → 80%)
- collectResponses: +2.8% (47% → 80%)
- Public APIs: +1.9% (0% → 100%)
- Other improvements: +4.6%
- **Total**: 66.2% → 80%+

---

## 🔍 P2: SEMGREP STATIC ANALYSIS (2-4 hours)

### Installation
- [ ] Install: `python3 -m pip install semgrep --user`
- [ ] Verify: `semgrep --version`
- [ ] Add to PATH: `export PATH=$PATH:~/.local/bin`

### Run Analysis
- [ ] Run: `make semgrep-check > /tmp/semgrep_findings.txt`
- [ ] Review: `cat /tmp/semgrep_findings.txt`

### Fix ERROR-Level Findings (CRITICAL)
Rules from SEMGREP_RULES_SUMMARY.md:
- [ ] Fix timer/ticker leaks (unclosed timers/tickers)
- [ ] Fix mutex issues (deadlocks, missing unlocks)
- [ ] Fix panic on user input (unbounded input, unvalidated conversions)

### Fix WARNING-Level Findings (Target: ≤5)
- [ ] Review each WARNING
- [ ] Fix or suppress with justification: `// nosemgrep: rule-id - reason`

### Verify Clean
- [ ] Final run: `make semgrep-check` shows 0 ERROR, ≤5 WARNING
- [ ] Commit hook passes: `.githooks/pre-commit` runs successfully

---

## 📚 P3: DOCUMENTATION (1-2 days)

### T123: Update Responder Examples in README
File: `README.md`
- [ ] Add responder usage example (Register, Unregister, UpdateService)
- [ ] Add code snippet showing basic responder setup
- [ ] Add conflict resolution example
- [ ] Link to detailed examples in `examples/` directory

### T124: Responder Usage Patterns Doc
Create: `docs/RESPONDER_USAGE.md`
- [ ] Document common patterns:
  - [ ] Basic service registration
  - [ ] Multi-service responder
  - [ ] Service update and unregistration
  - [ ] Conflict detection and resolution
  - [ ] Custom transport usage
- [ ] Include code examples for each pattern
- [ ] Document best practices

### T125: Conflict Resolution Patterns
Create: `docs/CONFLICT_RESOLUTION.md`
- [ ] Explain RFC 6762 §8.2 tie-breaking
- [ ] Document `ConflictDetector` usage
- [ ] Show examples:
  - [ ] Detecting conflicts during probing
  - [ ] Lexicographic comparison
  - [ ] Automatic renaming (with -2, -3 suffix)
- [ ] Include troubleshooting tips

### T126: Troubleshooting Guide
Create: `docs/TROUBLESHOOTING.md`
- [ ] Common issues:
  - [ ] "No responses received" - firewall, multicast issues
  - [ ] "Service name conflict" - conflict resolution
  - [ ] "Tests timing out" - integration test timing tolerance
  - [ ] "Build errors" - Go version, dependency issues
- [ ] Debug techniques:
  - [ ] Enabling verbose logging
  - [ ] Using Wireshark to inspect mDNS traffic
  - [ ] Checking network interfaces and multicast groups
- [ ] Platform-specific notes (Linux, macOS, Windows)

### Examples Documentation Enhancement
Files: `examples/*/README.md`
- [ ] Add step-by-step setup instructions
- [ ] Include expected output for each example
- [ ] Document requirements (Go version, network setup)
- [ ] Add troubleshooting section per example

---

## 🔧 P4: BUG FIXES & POLISH (1-2 days)

### Known Issues from Test Output
- [ ] Fix test that's causing overall suite failure (identify which one)
- [ ] Complete TestHandleQuery_RejectsWrongSubnet implementation
- [ ] Complete TestHandleQuery_QUBitUnicastResponse implementation

### TODO/FIXME Cleanup
From previous audit, address remaining items:
- [ ] Review all TODO comments in codebase
- [ ] Either implement or document deferral reason
- [ ] Remove obsolete TODOs

### Code Quality
- [ ] Run `go vet ./...` - must pass with 0 warnings
- [ ] Run `gofmt -l .` - must return empty (all files formatted)
- [ ] Check for unused imports/variables

---

## ✅ SUCCESS CRITERIA (ALL MUST BE TRUE)

### Phase 1: RFC Features ✅
- [x] Goodbye packets implemented
- [x] Source validation implemented (basic)
- [x] TC bit implemented
- [x] QU bit implemented (basic)
- [x] All P0 tests pass

### Phase 2: Coverage (REQUIRED)
- [ ] Coverage ≥80% verified (`make test-coverage-report | grep total`)
- [ ] All packages ≥75% coverage
- [ ] All tests pass (`make test && make test-race`)

### Phase 3: Quality (REQUIRED)
- [ ] Semgrep clean (0 ERROR, ≤5 WARNING)
- [ ] Build passes (`go build ./...`)
- [ ] No test failures

### Phase 4: Documentation (REQUIRED)
- [ ] RESPONDER_USAGE.md exists
- [ ] CONFLICT_RESOLUTION.md exists
- [ ] TROUBLESHOOTING.md exists
- [ ] README updated with responder examples

### Phase 5: Polish (REQUIRED)
- [ ] All TODOs addressed or documented
- [ ] `go vet` passes
- [ ] `gofmt` clean
- [ ] No skipped tests (except platform-specific)

---

## 📊 Progress Tracking

**Check after each task**:
```bash
# Check coverage
export PATH=$PATH:$HOME/go_installation/go/bin
make test-coverage-report

# Check Semgrep
make semgrep-check | head -20

# Check documentation
ls -la docs/*.md

# Check for TODOs
grep -rn "TODO\|FIXME" --include="*.go" | wc -l
```

---

## 🚨 DO NOT EXIT UNTIL

1. ✅ All P0 RFC features implemented (DONE)
2. ⏳ Coverage ≥80.0% (`make test-coverage-report` shows ≥80%)
3. ⏳ Semgrep clean (0 ERROR, ≤5 WARNING)
4. ⏳ All tests pass (`make test && make test-race`)
5. ⏳ All 3 documentation files exist
6. ⏳ No skipped tests without justification

**Current Status**: P0 complete, P1-P4 in progress

---

## Notes
- TDD methodology: Write test FIRST (RED), then implement (GREEN), then refactor
- Ralph implemented 4 RFC features but skipped some tests (marked TODO)
- Need to identify which test is failing the overall suite
- Test coverage is the next priority

**Last Updated**: 2026-01-06 after Ralph iteration 1
