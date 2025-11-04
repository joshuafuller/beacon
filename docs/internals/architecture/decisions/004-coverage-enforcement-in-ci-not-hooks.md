# ADR-004: Coverage Enforcement in CI, Not Pre-commit Hooks

**Status**: Accepted
**Date**: 2025-11-03
**Deciders**: Project maintainers
**Related**: Constitution Principle III (TDD), REQ-F8-2 (≥80% coverage)

---

## Context

Beacon maintains ≥80% test coverage through disciplined TDD. Question arose: Should we enforce coverage as a pre-commit hook to prevent regression?

## Decision

**Enforce coverage in CI (`make ci-fast`, `make ci-full`), NOT in pre-commit hooks.**

Pre-commit hook checks:
- ✅ Code formatting (`gofmt`)
- ✅ Static analysis (`go vet`, `staticcheck`)
- ✅ Constitution enforcement (`semgrep`)
- ❌ Test coverage (moved to CI)

## Rationale

### Why Coverage Belongs in CI

1. **Measures aggregate health, not commit safety**
   - Coverage answers: "Is the codebase healthy?"
   - Pre-commit should answer: "Is this specific commit safe?"

2. **Coverage is a lagging indicator**
   - Reflects outcome of TDD practice
   - Not a deterministic check like formatting

3. **CI provides proper context**
   - Can review trends over time
   - Can make context-aware decisions (allow drops during refactoring)
   - Provides visibility without blocking workflow

### Why NOT in Pre-commit Hook

1. **Breaks TDD RED phase**
   ```bash
   # RED: Add failing test
   vim querier/feature_test.go
   git commit -m "RED: Add test for feature"
   # ❌ BLOCKED: Coverage dropped (new untested code paths)
   # Developer forced to skip hook or commit test+impl together
   ```

2. **Slow (3-5s vs. 1-2s)**
   - Pre-commit should be <2s for good UX
   - Coverage requires full test suite run
   - Creates friction in development flow

3. **Non-deterministic**
   - Same commit = different coverage depending on global state
   - Adding well-tested feature can drop global % (denominator grows)
   - Pre-commit checks should be deterministic

4. **Easy to game**
   ```go
   // 100% coverage, 0% value
   func TestFoo(t *testing.T) { Foo() }
   ```

5. **Wrong optimization target**
   - Optimizes for "never drop coverage" not "write good tests"
   - Encourages gaming metrics instead of testing behavior
   - TDD culture > coverage gate

## Implementation

### What We Built

1. **CI Enforcement** (already existed)
   - `make ci-fast`: Fails if coverage <80%
   - `make ci-full`: Comprehensive validation
   - Blocks merge, not commit

2. **Developer Visibility** (new)
   - `make test-coverage-report`: Pretty report by package
   - `./scripts/coverage-trend.sh`: Track trends over time
   - HTML reports via `make test-coverage`

3. **Pre-commit Hook** (unchanged)
   - Fast (~1-2s)
   - Deterministic (format, vet, semgrep)
   - Catches real bugs (resource leaks, races)

### Enforcement Points

| Check Point | Coverage Gate? | Speed | Blocks |
|-------------|---------------|-------|--------|
| Pre-commit | ❌ No | ~1-2s | Commit (on format/vet/semgrep) |
| `make ci-fast` | ✅ Yes (≥80%) | ~45s | Push |
| `make ci-full` | ✅ Yes (≥80%) | ~3m | Merge |

## Consequences

### Positive

- ✅ TDD workflow unblocked (RED phase works)
- ✅ Pre-commit stays fast (<2s)
- ✅ Coverage still enforced (before merge)
- ✅ Better visibility via reporting tools
- ✅ Context-aware enforcement (CI can show trends)

### Negative

- ⚠️ Developer can push commits with low coverage (caught in CI)
- ⚠️ Requires CI infrastructure (but we already have Makefile targets)

### Mitigations

- Make coverage highly visible (`make test-coverage-report`)
- Provide trend tracking (`./scripts/coverage-trend.sh`)
- Document philosophy in CLAUDE.md
- Trust TDD culture + code review

## Alternatives Considered

### 1. Pre-commit Coverage Gate
**Rejected**: Breaks TDD, slow, non-deterministic (see rationale above)

### 2. Coverage Diff Tool (only check new code)
**Not implemented**: Complex to build correctly, adds tooling dependency, still has timing issues
**Future**: Could revisit if coverage regression becomes a problem

### 3. Per-package Minimums
**Not implemented**: Could enforce 90%+ for security-critical packages
**Future**: Could add to `make ci-full` if needed

### 4. No Enforcement at All
**Rejected**: Coverage is valuable, just enforce in right place (CI)

## References

- CLAUDE.md (Coverage Philosophy section)
- Constitution Principle III (TDD)
- F-8 Testing Strategy (REQ-F8-2: ≥80% coverage)
- Pre-commit hook: `.githooks/pre-commit`
- CI targets: `Makefile` (ci-fast, ci-full)

## Notes

This decision aligns with Constitution principle that coverage is an **outcome** of good TDD practice, not a **gate** to enforce mechanically.

The goal is **sustainable high coverage through culture**, not **mandatory coverage through tooling**.
