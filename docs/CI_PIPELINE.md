# CI Pipeline Documentation

**Last Updated**: 2025-11-05
**Status**: Complete CI Overhaul

## Overview

The Beacon CI pipeline is designed to provide fast feedback while ensuring comprehensive validation before merging code. The pipeline runs on GitHub Actions and consists of multiple jobs that run in parallel for optimal speed.

## Pipeline Jobs

### 1. Fast CI (`ci-fast`)
**Purpose**: Quick feedback loop for every commit
**Runs on**: All pushes and pull requests
**Timeout**: 15 minutes

**What it does**:
- ✅ Code formatting check (`gofmt`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint` v2.5.0)
- ✅ Security checks (`semgrep` - optional)
- ✅ Race condition detection (`go test -race`)
- ✅ Code coverage (minimum 80% required)
- ✅ Coverage upload to Codecov

**Key Features**:
- Uses Go module caching for faster builds
- Installs `golangci-lint` v2.5.0 with proper v2 configuration
- Temporarily allows lint warnings during technical debt cleanup
- Semgrep is optional and won't fail the build if not installed

### 2. Full CI (`ci-full`)
**Purpose**: Comprehensive validation for PRs and main branch
**Runs on**: Pull requests and pushes to main
**Timeout**: 20 minutes

**What it does**:
- ✅ All checks from Fast CI
- ✅ Enhanced static analysis (`staticcheck`)
- ✅ RFC compliance tests (contract tests)
- ✅ Integration tests (real network tests)
- ✅ Fuzz tests (30 seconds)
- ✅ Benchmark tests (5 iterations)

**Key Features**:
- Comprehensive test suite covering all test types
- Validates RFC 6762/6763 compliance
- Performance regression detection via benchmarks
- Integration tests with real mDNS traffic

### 3. Build Matrix (`build-matrix`)
**Purpose**: Cross-platform and multi-version validation
**Runs on**: All pushes and pull requests
**Timeout**: 10 minutes per combination

**Test Matrix**:
- **Operating Systems**: Ubuntu, macOS, Windows
- **Go Versions**: 1.21, 1.22, 1.23
- **Total**: 9 combinations (3 OS × 3 versions)

**What it does**:
- ✅ Build verification (`go build ./...`)
- ✅ Unit tests in short mode (`go test -short`)
- ✅ Race detection (Linux/macOS only, as Windows race detector is slower)

**Key Features**:
- `fail-fast: false` - All combinations run even if one fails
- Platform-specific race detection (skipped on Windows for speed)
- Validates compatibility across supported Go versions

### 4. Security Scan (`security`)
**Purpose**: Security vulnerability detection
**Runs on**: All pushes and pull requests
**Timeout**: 10 minutes

**What it does**:
- ✅ Security scanning with `gosec`
- ✅ SARIF report generation
- ✅ Upload to GitHub Security tab

**Key Features**:
- Integrates with GitHub Security features
- Non-blocking (continues on error)
- Results visible in Pull Request security tab

### 5. Go Report Card (`goreportcard`)
**Purpose**: Update external Go Report Card score
**Runs on**: Only on pushes to main branch
**Timeout**: 5 minutes

**What it does**:
- Triggers Go Report Card refresh after main branch updates

## Key Improvements

### Configuration Migration
1. **golangci-lint v2 Migration**
   - Migrated `.golangci.yml` from v1 to v2 format
   - Added `version: 2` declaration
   - Updated linter configuration structure
   - Fixed output format specification

2. **Version Pinning**
   - Go: 1.21 (minimum supported version)
   - golangci-lint: v2.5.0
   - staticcheck: latest

### Performance Optimizations
1. **Caching Strategy**
   - Go module caching via `actions/setup-go@v5`
   - Additional build cache for faster subsequent runs
   - Cache key based on `go.sum` hash

2. **Concurrency Control**
   - Cancel in-progress runs when new commits pushed
   - Parallel job execution where possible
   - Optimized timeout values per job

3. **Fast Feedback**
   - Separate fast and full CI jobs
   - Short test mode for build matrix
   - Early failure detection

### Reliability Improvements
1. **Error Handling**
   - Graceful handling of optional tools (semgrep)
   - Continue-on-error for non-critical checks
   - Clear error messages with actionable feedback

2. **Timeout Protection**
   - All jobs have explicit timeouts
   - Prevents hung builds from blocking the queue

3. **Artifact Management**
   - Coverage reports uploaded for 30 days
   - Test results preserved for 7 days
   - SARIF reports for security analysis

## Configuration Files

### `.golangci.yml`
- **Version**: v2
- **Default Linters**: standard (includes errcheck, govet, ineffassign, staticcheck, unused)
- **Additional Linters**: misspell, bodyclose, gosec, revive, gocyclo, dupl, goconst, prealloc
- **Settings**:
  - Complexity limit: 15
  - Duplication threshold: 100 tokens
  - Test file exclusions for appropriate linters

### `.github/workflows/ci.yml`
- **Concurrency**: Cancels outdated runs automatically
- **Environment Variables**: Centralized version configuration
- **Jobs**: 5 parallel jobs for comprehensive coverage

## Running Locally

### Fast CI (recommended for development)
```bash
make ci-fast
```

This runs the same checks as the Fast CI job:
- Format check
- Vet
- Lint (warnings allowed)
- Semgrep (if installed)
- Race detector
- Coverage check (≥80%)

### Full CI (comprehensive validation)
```bash
make ci-full
```

This runs all tests including:
- All Fast CI checks
- Staticcheck
- Contract tests (RFC compliance)
- Integration tests
- Fuzz tests (30 seconds)
- Benchmarks

### Individual Checks
```bash
# Just formatting check
make fmt-check

# Just linting
golangci-lint run --config .golangci.yml ./...

# Just tests with race detector
make test-race

# Just coverage
make test-coverage
```

## Troubleshooting

### golangci-lint v2 Configuration Error
**Symptom**: `unsupported version of the configuration`

**Solution**: The configuration must have `version: 2` at the top. The new format uses:
- `linters.default: standard` instead of enable-all
- `linters.settings` instead of `linters-settings`
- `output.formats` as a map instead of array

### Coverage Below 80%
**Symptom**: CI fails with coverage below minimum

**Solution**:
1. Run `make test-coverage-report` to see detailed coverage
2. Add tests for uncovered code paths
3. Use `./scripts/coverage-trend.sh` to track progress

### Race Condition Detected
**Symptom**: Tests fail with race detector

**Solution**:
1. Run `make test-race` locally to reproduce
2. Fix the race condition using proper synchronization
3. Reference F-8 Testing Strategy and REQ-F8-5

### Semgrep Not Installed
**Symptom**: Warning about semgrep not available

**Solution**: Install semgrep (optional):
```bash
pip install semgrep
```

Or skip locally - CI will handle it gracefully.

### Build Matrix Failures
**Symptom**: Specific OS or Go version fails

**Solution**:
1. Check if it's a platform-specific issue
2. Use build tags if needed for platform differences
3. Ensure compatibility with Go 1.21+ (minimum version)

## Best Practices

1. **Run Fast CI locally before pushing**
   - Catches most issues before CI runs
   - Saves CI minutes and time

2. **Watch for lint warnings**
   - Currently set to not fail CI during technical debt cleanup
   - Will be enforced strictly once cleanup complete

3. **Keep coverage above 80%**
   - Aim for 85%+ on hot paths
   - Constitution requires ≥80%

4. **Test on your platform**
   - Use `go test -short` for quick validation
   - Full tests for comprehensive validation

5. **Review security findings**
   - Check GitHub Security tab after PR
   - Address high-severity findings promptly

## Future Improvements

1. **Dependency Review**
   - Add automatic dependency vulnerability scanning
   - Integrate with GitHub Dependabot

2. **Performance Tracking**
   - Track benchmark results over time
   - Alert on performance regressions

3. **Test Flakiness Detection**
   - Automatic detection of flaky tests
   - Historical test result analysis

4. **Coverage Trending**
   - Automatic coverage trend graphs
   - Alert on coverage decreases

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint v2 Migration Guide](https://golangci-lint.run/docs/product/migration-guide)
- Beacon Constitution (`.specify/memory/constitution.md`)
- F-8 Testing Strategy (`.specify/specs/F-8-testing-strategy.md`)
- Makefile (`/Makefile`)

## Support

For CI issues:
1. Check this documentation first
2. Review the Makefile for local commands
3. Check GitHub Actions logs for detailed error messages
4. Reference the Constitution and F-Specs for requirements
