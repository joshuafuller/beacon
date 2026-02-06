# BEACON Development Commands

## Environment Setup
```bash
# Go compiler (required for all commands)
export PATH=$PATH:$HOME/go_installation/go/bin
export GOPATH=$HOME/go

# Verify installation
go version  # Should show: go1.23.5 linux/amd64
```

## Build Commands
```bash
# Build all packages
go build ./...

# Build specific packages
go build ./querier
go build ./responder
go build ./internal/...

# Build examples
go build ./examples/discover
go build ./examples/interface-specific
```

## Test Commands
```bash
# Run all tests (MOST COMMON - use this after changes)
make test

# Run with race detector (use before committing)
make test-race

# Run with coverage
make test-coverage

# Detailed coverage report
make test-coverage-report

# Coverage for specific package
go test ./querier -cover -v
go test ./responder -cover -v

# Run specific test
go test ./responder -run TestResponder_Register -v

# Integration tests
go test ./tests/integration -v

# Contract tests (RFC compliance)
go test ./tests/contract -v

# Fuzz tests
go test -fuzz=FuzzParseMessage -fuzztime=10s ./tests/fuzz
```

## Code Quality
```bash
# Static analysis (Semgrep)
make semgrep-check

# Go vet
go vet ./...

# Format code
gofmt -w .

# Check formatting
gofmt -l .
```

## Coverage Analysis
```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out -o coverage.html
```

## Common Workflows

### After Making Changes
1. `gofmt -w .` - Format code
2. `go vet ./...` - Run static analysis
3. `make test` - Run all tests
4. `make test-coverage-report` - Check coverage

### Before Committing
1. `make test-race` - Run tests with race detector
2. `make semgrep-check` - Run Semgrep
3. `make test-coverage-report` - Verify coverage ≥80%
4. `go build ./...` - Verify build

### Debugging Tests
```bash
# Run with verbose output
go test ./path/to/package -v

# Run with race detector
go test ./path/to/package -race -v

# Run with timeout
go test ./path/to/package -timeout 30s -v

# Clear test cache
go clean -testcache
```

## System Utilities (Linux/WSL2)
- `ls` - /usr/bin/ls
- `grep` - /usr/bin/grep
- `find` - /usr/bin/find
- `sed` - /usr/bin/sed
- `awk` - /usr/bin/awk
- `tail` - /usr/bin/tail
- `head` - /usr/bin/head
- `git` - Available

## Ralph Commands (Autonomous Development)
```bash
# View Ralph status
cat status.json

# View Ralph logs
tail -f logs/ralph.log

# Check task progress
cat @fix_plan.md
```
