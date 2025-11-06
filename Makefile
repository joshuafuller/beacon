# Beacon Project Makefile
# Updated: 2025-11-03
# Governance: Aligned with Beacon Constitution v1.0.0

.PHONY: help test test-race test-coverage test-coverage-report test-integration test-contract test-fuzz test-benchmark lint fmt fmt-check vet vet-staticcheck build clean verify ci-fast ci-full all

# Default target
.DEFAULT_GOAL := help

# Variables
GO := go
GOLANGCI_LINT := golangci-lint
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
MIN_COVERAGE := 80

## help: Display this help message
help:
	@echo "Beacon Project - Make Targets"
	@echo ""
	@echo "Testing (Constitution Principle III: TDD, F-8 Testing Strategy):"
	@echo "  make test                  - Run all unit tests (fast, excludes fuzz)"
	@echo "  make test-race             - Run tests with race detector (fast, excludes fuzz)"
	@echo "  make test-coverage         - Run tests with coverage report (≥80% required, fast)"
	@echo "  make test-coverage-report  - Detailed coverage report by package (pretty output)"
	@echo "  make test-integration      - Run integration tests"
	@echo "  make test-contract         - Run API contract tests (RFC compliance, REQ-F8-6)"
	@echo "  make test-fuzz             - Run fuzz tests (SLOW: 10,000 iterations, use test-fuzz-ci for CI)"
	@echo "  make test-fuzz-ci          - Run fuzz tests for CI (30 seconds, F-8 recommendation)"
	@echo "  make test-benchmark        - Run benchmark tests (F-8 Testing Strategy)"
	@echo ""
	@echo "Code Quality (Constitution Principle VII: Excellence):"
	@echo "  make lint              - Run golangci-lint"
	@echo "  make semgrep           - Run Semgrep (informational only)"
	@echo "  make semgrep-check     - Run Semgrep and fail on findings (for CI)"
	@echo "  make fmt               - Format code with gofmt"
	@echo "  make fmt-check         - Check if code is formatted (no changes)"
	@echo "  make vet               - Run go vet"
	@echo "  make vet-staticcheck   - Run go vet + staticcheck"
	@echo ""
	@echo "Build:"
	@echo "  make build             - Build all packages"
	@echo "  make clean             - Remove build artifacts"
	@echo ""
	@echo "CI/CD Pipelines:"
	@echo "  make verify            - Quick validation checks (fmt-check, vet, lint, test)"
	@echo "  make ci-fast           - Fast CI feedback (unit tests + race + coverage, no integration/fuzz)"
	@echo "  make ci-full           - Full CI validation (all tests including integration, fuzz, benchmarks)"
	@echo ""
	@echo "Composite:"
	@echo "  make all               - Run full validation pipeline (fmt, vet, lint, test-race, test-coverage, test-contract)"
	@echo ""
	@echo "Coverage Tracking:"
	@echo "  ./scripts/coverage-trend.sh         - Record current coverage"
	@echo "  ./scripts/coverage-trend.sh --show  - Show coverage history"
	@echo "  ./scripts/coverage-trend.sh --graph - Show trend graph"

## test: Run all unit tests (excludes fuzz tests for speed)
test:
	@echo "Running unit tests (excludes fuzz tests)..."
	$(GO) test -v $$(go list ./... | grep -v /tests/fuzz)

## test-race: Run tests with race detector (FR-019, REQ-F8-5: MUST pass with zero race conditions)
test-race:
	@echo "Running tests with race detector (FR-019, REQ-F8-5, excludes fuzz tests)..."
	$(GO) test -race -v $$(go list ./... | grep -v /tests/fuzz)

## test-coverage: Run tests with coverage report (SC-010, REQ-F8-2: ≥80% required)
test-coverage:
	@echo "Running tests with coverage (REQ-F8-2: ≥80% required, excludes fuzz tests)..."
	$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic $$(go list ./... | grep -v /tests/fuzz)
	@echo ""
	@echo "Coverage Report:"
	$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML report: $(COVERAGE_HTML)"
	@echo ""
	@echo "Checking minimum coverage (≥$(MIN_COVERAGE)%)..."
	@COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < $(MIN_COVERAGE)" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below minimum $(MIN_COVERAGE)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets minimum $(MIN_COVERAGE)%"; \
	fi

## test-coverage-report: Generate detailed coverage report by package
test-coverage-report:
	@echo "Generating detailed coverage report by package (excludes fuzz tests)..."
	@$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic $$(go list ./... | grep -v /tests/fuzz) || true
	@if [ ! -f $(COVERAGE_FILE) ]; then \
		echo "❌ Failed to generate coverage report (tests may have failed)"; \
		exit 1; \
	fi
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║               BEACON TEST COVERAGE REPORT                          ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep -v "total:" | awk '{printf "%-60s %6s\n", $$1, $$3}'
	@echo ""
	@echo "────────────────────────────────────────────────────────────────────"
	@COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	printf "%-60s %6.1f%%\n" "TOTAL COVERAGE" $$COVERAGE; \
	echo "────────────────────────────────────────────────────────────────────"; \
	echo ""; \
	if [ $$(echo "$$COVERAGE < $(MIN_COVERAGE)" | bc -l) -eq 1 ]; then \
		printf "Status: ❌ Below target ($(MIN_COVERAGE)%% required)\n"; \
		exit 1; \
	elif [ $$(echo "$$COVERAGE < 85" | bc -l) -eq 1 ]; then \
		printf "Status: ⚠️  Meets minimum, aim for 85%%+\n"; \
	else \
		printf "Status: ✅ Excellent coverage!\n"; \
	fi
	@echo ""

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v ./tests/integration/...

## test-contract: Run API contract tests (REQ-F8-6: RFC Compliance Testing)
test-contract:
	@echo "Running API contract tests (RFC compliance, REQ-F8-6)..."
	@echo "  Validating RFC 6762 compliance (mDNS protocol requirements)"
	$(GO) test -v ./tests/contract/...

## test-fuzz: Run fuzz tests (NFR-003: 10,000 iterations)
test-fuzz:
	@echo "Running fuzz tests (NFR-003: 10,000 iterations)..."
	$(GO) test -fuzz=FuzzMessageParser -fuzztime=10000x ./tests/fuzz/...

## test-fuzz-ci: Run fuzz tests for CI (F-8 recommendation: 30 seconds)
test-fuzz-ci:
	@echo "Running fuzz tests for CI (30 seconds, F-8 Testing Strategy)..."
	$(GO) test -fuzz=FuzzMessageParser -fuzztime=30s ./tests/fuzz/...

## test-benchmark: Run benchmark tests (F-8 Testing Strategy)
test-benchmark:
	@echo "Running benchmark tests (F-8 Testing Strategy)..."
	$(GO) test -bench=. -benchmem -count=5 ./...

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run --config .golangci.yml ./...

## lint-warn: Run golangci-lint but don't fail on warnings (for CI during technical debt cleanup)
lint-warn:
	@echo "Running golangci-lint (warning mode - won't fail CI)..."
	$(GOLANGCI_LINT) run --config .golangci.yml ./... || true
	@echo "⚠️  Lint warnings found (tracked in GitHub Issues)"

## semgrep: Run Semgrep security and quality checks (informational only)
semgrep:
	@echo "Running Semgrep security and quality checks (informational)..."
	@if ! command -v semgrep > /dev/null; then \
		echo "⚠️  semgrep not installed. Install with: pip install semgrep"; \
		exit 1; \
	fi
	@semgrep --config=.semgrep.yml . || true
	@echo ""
	@echo "ℹ️  Semgrep findings are informational. See SEMGREP_RULES_SUMMARY.md for details."

## semgrep-check: Run Semgrep and fail on findings (for CI/CD)
semgrep-check:
	@echo "Running Semgrep security and quality checks (strict mode)..."
	@if ! command -v semgrep > /dev/null; then \
		echo "⚠️  semgrep not installed. Install with: pip install semgrep"; \
		echo "   Run: pip install semgrep"; \
		exit 1; \
	fi
	@echo "   Enforcing Constitution principles and F-Spec requirements..."
	@semgrep --config=.semgrep.yml --error . --exclude .semgrep-tests
	@echo "✅ Semgrep: No issues found"

## fmt: Format code with gofmt
fmt:
	@echo "Formatting code..."
	@gofmt -l -w .
	@echo "Code formatted."

## fmt-check: Check if code is formatted (for CI validation)
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted."; \
	fi

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## vet-staticcheck: Run go vet + staticcheck (F-8 Testing Strategy recommendation)
vet-staticcheck: vet
	@echo "Running staticcheck..."
	@if ! command -v staticcheck > /dev/null; then \
		echo "⚠️  staticcheck not installed. Installing..."; \
		$(GO) install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@staticcheck ./...

## build: Build all packages
build:
	@echo "Building all packages..."
	$(GO) build ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f *.test
	rm -f *.prof
	$(GO) clean -cache -testcache
	@echo "Clean complete."

## verify: Quick validation checks (fast validation without formatting)
## Checks formatting, runs vet/lint/semgrep, runs unit tests
verify: fmt-check vet lint semgrep-check test
	@echo ""
	@echo "✅ Quick validation checks passed!"
	@echo "   - Code properly formatted (gofmt check)"
	@echo "   - No vet issues (go vet)"
	@echo "   - No lint issues (golangci-lint)"
	@echo "   - No Semgrep issues (Constitution/F-Spec enforcement)"
	@echo "   - Unit tests passing"

## ci-fast: Fast CI feedback loop (for quick validation on every commit)
## Excludes slow tests (integration, fuzz, benchmarks) for fast feedback
## NOTE: Using lint-warn temporarily during technical debt cleanup (see GitHub Issues)
ci-fast: fmt-check vet lint-warn semgrep-check test-race test-coverage
	@echo ""
	@echo "✅ Fast CI checks passed!"
	@echo "   - Code properly formatted (gofmt check)"
	@echo "   - No vet issues (go vet)"
	@echo "   - Lint warnings checked (golangci-lint - warnings only)"
	@echo "   - No Semgrep issues (Constitution/F-Spec enforcement)"
	@echo "   - Zero race conditions (go test -race, REQ-F8-5)"
	@echo "   - Coverage ≥$(MIN_COVERAGE)% (go test -cover, REQ-F8-2)"

## ci-full: Full CI validation (comprehensive checks for PRs/releases)
## Includes all tests: unit, integration, contract, fuzz, benchmarks
## NOTE: Using lint-warn temporarily during technical debt cleanup (see GitHub Issues)
ci-full: fmt-check vet-staticcheck lint-warn semgrep-check test-race test-coverage test-contract test-integration test-fuzz-ci test-benchmark
	@echo ""
	@echo "✅ Full CI validation passed!"
	@echo "   - Code properly formatted (gofmt check)"
	@echo "   - No vet/staticcheck issues (go vet + staticcheck)"
	@echo "   - Lint warnings checked (golangci-lint - warnings only)"
	@echo "   - No Semgrep issues (Constitution/F-Spec enforcement)"
	@echo "   - Zero race conditions (go test -race, REQ-F8-5)"
	@echo "   - Coverage ≥$(MIN_COVERAGE)% (go test -cover, REQ-F8-2)"
	@echo "   - RFC compliance validated (test-contract, REQ-F8-6)"
	@echo "   - Integration tests passed"
	@echo "   - Fuzz tests passed (30 seconds)"
	@echo "   - Benchmarks completed"

## all: Run full validation pipeline (F-8 Testing Strategy compliance)
## Includes: fmt → vet → lint → semgrep-check → test-race → test-coverage → test-contract
all: fmt vet lint semgrep-check test-race test-coverage test-contract
	@echo ""
	@echo "✅ All validation checks passed!"
	@echo "   - Code formatted (gofmt)"
	@echo "   - No vet issues (go vet)"
	@echo "   - No lint issues (golangci-lint)"
	@echo "   - No Semgrep issues (Constitution/F-Spec enforcement)"
	@echo "   - Zero race conditions (go test -race, REQ-F8-5)"
	@echo "   - Coverage ≥$(MIN_COVERAGE)% (go test -cover, REQ-F8-2)"
	@echo "   - RFC compliance validated (test-contract, REQ-F8-6)"
