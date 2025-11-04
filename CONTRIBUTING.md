# Contributing to Beacon

Thank you for your interest in contributing to Beacon! We welcome contributions from the community and are grateful for your support.

---

## Table of Contents

- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Code Contribution Workflow](#code-contribution-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Community Guidelines](#community-guidelines)

---

## How Can I Contribute?

There are many ways to contribute to Beacon:

### 1. Report Bugs

**Before submitting a bug report**:
- Check [existing issues](https://github.com/joshuafuller/beacon/issues) to avoid duplicates
- Collect debug information (logs, packet captures, version info)
- Create a minimal reproduction case

**Submit a bug report**:
- Use the [Bug Report template](https://github.com/joshuafuller/beacon/issues/new?template=bug_report.md)
- Include Go version, OS, Beacon version
- Provide steps to reproduce
- Attach relevant logs or packet captures

### 2. Suggest Features

**Before suggesting a feature**:
- Check the [roadmap](ROADMAP.md) to see if it's already planned
- Review [RFC 6762](RFC%20Docs/rfc6762.txt) to ensure it's RFC-compliant
- Search [discussions](https://github.com/joshuafuller/beacon/discussions) for similar ideas

**Submit a feature request**:
- Use the [Feature Request template](https://github.com/joshuafuller/beacon/issues/new?template=feature_request.md)
- Explain the use case and motivation
- Provide examples of how it would work
- Reference relevant RFC sections

### 3. Improve Documentation

**Documentation contributions are highly valued**:
- Fix typos or unclear explanations
- Add examples for common use cases
- Improve API documentation
- Write guides for specific platforms or scenarios

**How to contribute docs**:
- Documentation lives in `docs/` and is written in Markdown
- Follow the [Documentation Style Guide](#documentation-style-guide)
- Submit a PR with your changes (see [workflow below](#code-contribution-workflow))

### 4. Write Code

**Types of code contributions**:
- Bug fixes
- Performance improvements
- New features (must have approved spec first)
- Test coverage improvements
- Platform-specific improvements (macOS, Windows)

**Before writing code**:
- Read this entire guide
- Set up your development environment
- Discuss large changes in an issue first

---

## Development Setup

### Prerequisites

- **Go 1.21 or later**
- **Git**
- **make** (optional but recommended)
- **Linux, macOS, or Windows** (Linux is best-tested)

### Clone the Repository

```bash
git clone https://github.com/joshuafuller/beacon.git
cd beacon
```

### Install Development Tools

```bash
# Install semgrep (static analysis)
pip install semgrep

# Verify installation
semgrep --version
go version
```

### Run Tests

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run with coverage
make test-coverage

# Verify everything works
make verify  # Runs format, vet, semgrep, tests
```

**Expected output**: All tests should pass.

### Project Structure

```
beacon/
├── querier/           # Public API - service discovery
├── responder/         # Public API - service announcement
├── internal/          # Internal implementation (not importable)
│   ├── transport/     # Network abstraction
│   ├── message/       # DNS encoding/decoding
│   ├── protocol/      # RFC constants
│   └── ...
├── tests/             # Test suites
│   ├── contract/      # RFC compliance tests
│   ├── integration/   # Real network tests
│   └── fuzz/          # Fuzz tests
├── docs/              # Documentation
│   ├── guides/        # User guides
│   ├── api/           # API reference
│   ├── development/   # Developer docs
│   └── internals/     # Technical docs
└── examples/          # Code examples
```

**See**: [Architecture Overview](docs/guides/architecture.md) for detailed design

---

## Code Contribution Workflow

### 1. Fork and Branch

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR-USERNAME/beacon.git
cd beacon

# Add upstream remote
git remote add upstream https://github.com/joshuafuller/beacon.git

# Create a feature branch
git checkout -b feature/my-improvement
```

**Branch naming**:
- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `perf/description` - Performance improvements

### 2. Make Changes

**Follow TDD (Test-Driven Development)**:

```bash
# 1. RED - Write failing test
# Edit: querier/querier_test.go
func TestNewFeature(t *testing.T) {
    // Test for new functionality (will fail)
}

# Verify test fails
go test ./querier

# 2. GREEN - Implement minimum code to pass
# Edit: querier/querier.go
// Implementation

# Verify test passes
go test ./querier

# 3. REFACTOR - Clean up code
# Improve code quality, add comments, etc.

# Verify tests still pass
go test ./querier
```

### 3. Verify Quality

```bash
# Run all quality checks
make verify

# This runs:
# - gofmt (code formatting)
# - go vet (static analysis)
# - semgrep (security and RFC compliance)
# - tests with race detector
# - coverage check (≥80%)
```

**All checks must pass before submitting PR.**

### 4. Commit Changes

**Write good commit messages**:

```bash
# Good commit message format:
git commit -m "feat: add IPv6 multicast support

- Implement FF02::FB multicast group
- Add IPv6Transport implementation
- Update tests for dual-stack

Fixes #123"
```

**Commit message format**:
```
<type>: <short description>

<detailed description>

<footer>
```

**Types**:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `test` - Adding tests
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `chore` - Maintenance tasks

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/my-improvement

# Create PR on GitHub
# Use the pull request template
```

**PR checklist**:
- [ ] Tests added or updated
- [ ] All tests pass (`make test-race`)
- [ ] Code coverage ≥80% (`make test-coverage`)
- [ ] `make semgrep-check` passes
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (for user-facing changes)
- [ ] PR description explains the change

### 6. Code Review

**What to expect**:
- Maintainers will review your PR
- Feedback may be provided (changes requested)
- CI must pass (automated tests)
- At least one maintainer approval required

**Be patient**: Reviews may take 1-7 days depending on size and complexity.

---

## Coding Standards

### Go Style Guide

**Follow**:
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Beacon [Constitution](.specify/memory/constitution.md) principles

**Key rules**:

#### 1. Formatting

```bash
# MUST pass before commit
gofmt -w .
```

**No exceptions** - all code must be gofmt'd.

#### 2. Naming Conventions

```go
// ✅ Good names
type QueryType uint16
func (q *Querier) Query(ctx context.Context, name string, qtype QueryType) ([]ResourceRecord, error)

// ❌ Bad names
type QT uint16
func (q *Querier) q(c context.Context, n string, t QT) ([]RR, error)
```

**Guidelines**:
- Exported names start with capital letter
- Use camelCase (not snake_case)
- Acronyms are all caps (HTML, URL, ID)
- Short names for short scopes (`i` for loop index, `ctx` for context)

#### 3. Error Handling

```go
// ✅ Good - always check errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ Bad - error swallowed
result, _ := doSomething()
```

**Never ignore errors** without explicit comment explaining why.

#### 4. Context Usage

```go
// ✅ Good - context first parameter
func (q *Querier) Query(ctx context.Context, name string, qtype QueryType) ([]ResourceRecord, error)

// ❌ Bad - context not first
func (q *Querier) Query(name string, qtype QueryType, ctx context.Context) ([]ResourceRecord, error)
```

**All blocking operations must accept context.**

#### 5. Resource Cleanup

```go
// ✅ Good - defer immediately after acquisition
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()

// ❌ Bad - no defer
f, err := os.Open("file.txt")
// ... use f ...
f.Close()  // May be skipped if error occurs
```

**Always use defer for cleanup.**

### RFC Compliance

**All protocol code must comply with RFC 6762/6763:**

```go
// ✅ Good - reference RFC section
// Implements RFC 6762 §8.1 - Probing
func (s *StateMachine) Probe() error {
    // ...
}

// ❌ Bad - no RFC reference
func (s *StateMachine) Probe() error {
    // ...
}
```

**See**: [RFC Compliance Matrix](docs/internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md)

---

## Testing Requirements

### Test Coverage

**Minimum coverage**: 80% (enforced by CI)

**Check coverage**:
```bash
make test-coverage-report
```

**Current coverage**: 81.3%

### Types of Tests

#### 1. Unit Tests

**Location**: `*_test.go` files next to implementation

```go
// querier/querier_test.go
func TestQuery_ValidInput(t *testing.T) {
    q, err := querier.New()
    if err != nil {
        t.Fatalf("New() failed: %v", err)
    }
    defer q.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
    if err != nil {
        t.Errorf("Query() failed: %v", err)
    }
}
```

**Run**:
```bash
go test ./querier -v
```

#### 2. Integration Tests

**Location**: `tests/integration/`

**Requirements**: Real network, may require permissions

**Run**:
```bash
go test ./tests/integration -v
```

#### 3. Contract Tests (RFC Compliance)

**Location**: `tests/contract/`

**Purpose**: Verify RFC 6762/6763 compliance

**Run**:
```bash
go test ./tests/contract -v
```

#### 4. Fuzz Tests

**Location**: `tests/fuzz/`

**Purpose**: Find crashes from malformed input

**Run**:
```bash
go test -fuzz=FuzzParseMessage -fuzztime=30s ./tests/fuzz
```

### Testing Best Practices

1. **Test behavior, not implementation**
2. **Use table-driven tests for similar cases**
3. **Always test error cases**
4. **Use `t.Parallel()` for parallel tests**
5. **Clean up resources (use `defer`)**
6. **Use descriptive test names** (`TestQuery_Timeout_ReturnsError`)

**See**: [Testing Guide](docs/development/testing.md)

---

## Documentation

### When to Update Documentation

**Always update docs when**:
- Adding a public API
- Changing behavior
- Adding features
- Fixing user-facing bugs

### Documentation Style Guide

#### 1. API Documentation (godoc)

```go
// ✅ Good godoc comment
// Query sends an mDNS query for the specified name and type, returning all
// matching resource records received within the context timeout.
//
// The name should be a fully-qualified domain name (e.g., "_http._tcp.local").
// Query types include PTR, SRV, TXT, and A (see QueryType constants).
//
// Query is thread-safe and may be called concurrently from multiple goroutines.
//
// Example:
//
//	results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
//	if err != nil {
//	    return err
//	}
func (q *Querier) Query(ctx context.Context, name string, qtype QueryType) ([]ResourceRecord, error)
```

**Guidelines**:
- First sentence is a summary (appears in godoc index)
- Use complete sentences
- Include examples for non-trivial APIs
- Document thread-safety
- Document context behavior

#### 2. User Guides

**Location**: `docs/guides/`

**Format**: Markdown

**Structure**:
- Clear title and audience
- Table of contents (if long)
- Code examples with output
- Troubleshooting section
- Links to related docs

**See**: [Getting Started Guide](docs/guides/getting-started.md) for example

#### 3. README Updates

**When to update**: Adding major features or changing installation

**Keep**:
- Quick start examples up-to-date
- Feature list current
- Badges accurate

---

## Community Guidelines

### Code of Conduct

We follow the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

**Summary**: Be respectful, inclusive, and professional.

**Report violations**: joshuafuller@gmail.com

### Communication Channels

- **GitHub Issues** - Bug reports, feature requests, tasks
- **GitHub Discussions** - Questions, ideas, general discussion
- **Pull Requests** - Code review, implementation discussion

**Response times**:
- Issues: 1-3 days for initial response
- PRs: 1-7 days for review (depending on size)
- Discussions: Best effort

### Recognition

**Contributors are recognized**:
- In release notes (CHANGELOG.md)
- In commit history
- In project README (for significant contributions)

---

## Development Process

### Spec-Driven Development

**Beacon follows the [Spec Kit](https://github.com/github/spec-kit) framework**:

1. **Specification** - Define WHAT and WHY (in `specs/`)
2. **Planning** - Define HOW (architecture, strategy)
3. **Tasks** - Break into granular, testable tasks
4. **TDD Implementation** - RED → GREEN → REFACTOR
5. **Validation** - Verify success criteria met

**For major features**: Spec must be approved before implementation begins.

**See**: [CLAUDE.md](CLAUDE.md) for detailed development guidelines

### Release Process

**Versioning**: [Semantic Versioning 2.0](https://semver.org/)

**Release cycle**:
- **Patch** (v0.1.x): Bug fixes, every 1-2 weeks
- **Minor** (v0.x.0): New features, every 1-3 months
- **Major** (v1.0.0): Breaking changes, as needed

**See**: [Release Process](docs/development/release-process.md)

---

## Questions?

**Need help?**
- [GitHub Discussions](https://github.com/joshuafuller/beacon/discussions) - Ask questions
- [Troubleshooting Guide](docs/guides/troubleshooting.md) - Common issues
- [Development Guide](docs/development/README.md) - Detailed dev docs

**Before asking**:
- Search existing issues and discussions
- Read relevant documentation
- Try to create a minimal reproduction

---

## Thank You!

Your contributions make Beacon better for everyone. We appreciate your time and effort! 🎉

---

**Last updated**: 2025-11-04
