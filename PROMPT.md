Focus on P1 tasks from @fix_plan.md - increase test coverage to 80%+.

## Current Status
- RFC Features (P0): ✅ COMPLETE (4/4 implemented)
- Test Coverage: 66.2% → Need 80%+ (+13.8%)
- @fix_plan.md has full task breakdown

## Priority Tasks (Do in Order)

### 1. handleQuery Coverage Tests
Add tests to `responder/handlequery_test.go`:
- Test PTR query handling (_services._dns-sd._udp.local)
- Test service type PTR query (_http._tcp.local)
- Test service instance queries (SRV, TXT, A)
- Test query parsing edge cases (malformed, truncated)
- Test interface index extraction
- Test error paths (parse failures, unknown types)
- Test known-answer suppression logic
- Complete TestHandleQuery_RejectsWrongSubnet (currently SKIPPED)

### 2. collectResponses Coverage Tests
Add tests to `querier/collectresponses_test.go`:
- Test timeout behavior (context cancellation)
- Test duplicate response handling
- Test response collection with multiple answers
- Test edge cases (empty responses, malformed packets)
- Test context cancellation mid-collection
- Test buffer pool interaction

### 3. Public API Coverage Tests
Create `querier/options_test.go` (NEW FILE):
- Test WithInterfaces() - custom interface selection
- Test WithInterfaceFilter() - interface filtering function
- Test WithRateLimit() - rate limiting configuration
- Test WithRateLimitThreshold() - threshold configuration
- Test WithRateLimitCooldown() - cooldown configuration

## After Each Test File
1. Run tests: `go test ./[package] -v`
2. Check coverage: `make test-coverage-report`
3. Mark tasks complete in @fix_plan.md (change [ ] to [x])
4. Continue to next file

## Verification
```bash
export PATH=$PATH:$HOME/go_installation/go/bin
make test-coverage-report | grep "total.*statements"
# Target: XX.X% ≥ 80.0%
```

Write the test files directly. Do not ask for permission.
