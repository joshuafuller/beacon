# Pull Request

## Description

**What does this PR do?**:


**Related issue**: (e.g., Fixes #123, Closes #456)

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Refactoring (no functional changes)
- [ ] Test coverage improvement

## Changes Made

**List the specific changes made**:

-
-
-

## Testing

**How has this been tested?**:

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Contract tests added/updated (if RFC-related)
- [ ] Fuzz tests added/updated (if parsing/validation-related)
- [ ] Manual testing performed

**Test coverage**:
- Coverage before: %
- Coverage after: %

**Test commands run**:
```bash
make test
make test-race
make test-coverage
```

## RFC Compliance (if applicable)

- [ ] This change implements RFC 6762 §X.Y
- [ ] This change implements RFC 6763 §X.Y
- [ ] RFC Compliance Matrix updated
- [ ] Contract tests validate RFC compliance

## Documentation

- [ ] README.md updated (if needed)
- [ ] CHANGELOG.md updated (for user-facing changes)
- [ ] API documentation (godoc) added/updated
- [ ] User guides updated (docs/guides/)
- [ ] Examples added/updated
- [ ] Architecture Decision Record (ADR) created (for significant architectural changes)

## Code Quality

- [ ] Code follows Go style guide (gofmt)
- [ ] Code passes `go vet`
- [ ] Code passes `make semgrep-check`
- [ ] All tests pass with race detector (`make test-race`)
- [ ] Code coverage ≥80% (`make test-coverage`)
- [ ] Errors are properly wrapped and typed
- [ ] Resources are cleaned up (defer used correctly)
- [ ] Context is used for cancellation/timeouts
- [ ] Code includes RFC section references (if protocol code)

## Breaking Changes

**If this is a breaking change, describe**:

- What breaks?
- Why is this necessary?
- Migration path for users?

## Performance Impact

**Does this change affect performance?**:

- [ ] No performance impact
- [ ] Performance improvement (include benchmarks)
- [ ] Potential performance regression (justified by X)

**Benchmark results** (if applicable):
```
Before:
BenchmarkX-8   1000000   1234 ns/op   567 B/op   8 allocs/op

After:
BenchmarkX-8   2000000    456 ns/op   123 B/op   2 allocs/op
```

## Additional Notes

**Anything else reviewers should know**:


## Checklist

- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Screenshots / Logs (if applicable)

**Before**:


**After**:


---

**For Maintainers**:

- [ ] Spec approved (for new features)
- [ ] ADR created (for architectural changes)
- [ ] Security implications reviewed
- [ ] Platform compatibility verified (Linux/macOS/Windows)
- [ ] Backward compatibility maintained
