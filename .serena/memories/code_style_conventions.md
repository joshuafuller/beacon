# BEACON Code Style & Conventions

## Go Standards
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (no exceptions)
- Use `go vet` (must pass with zero warnings)
- Go 1.21+ idioms and patterns

## Naming Conventions

### Packages
- Short, lowercase, single word: `querier`, `responder`, `transport`
- No underscores or mixed caps

### Types
- PascalCase for exported types: `Querier`, `Responder`, `Transport`
- camelCase for unexported types: `registry`, `stateMachine`

### Functions/Methods
- PascalCase for exported: `New()`, `Query()`, `Register()`
- camelCase for unexported: `buildResponse()`, `parseMessage()`

### Variables
- camelCase: `serviceName`, `recordType`, `interfaceIndex`
- Acronyms in caps when starting a name: `TTL`, `IP`, `DNS`
- Acronyms lowercase when mid-name: `recordTTL`, `localIP`

## Documentation

### Godoc Comments
- All exported types/functions MUST have Godoc comments
- Start with the name of the item: `// Querier performs mDNS queries...`
- Explain WHY, not WHAT (code shows what)
- Reference RFCs when relevant: `// Per RFC 6762 §8.1, probing...`

### Code Comments
- Explain WHY, not WHAT
- Reference RFC sections for protocol behavior: `// RFC 6762 §7.1 - Known-Answer Suppression`
- Mark TODOs with task ID when applicable: `// TODO T032: Add logging when F-6 ready`

## Error Handling (F-3 Spec)

### Never Swallow Errors
```go
// ✅ CORRECT
func (t *UDPv4Transport) Close() error {
    return t.conn.Close()  // Propagate error
}

// ❌ WRONG
func (t *UDPv4Transport) Close() error {
    t.conn.Close()
    return nil  // Error swallowed!
}
```

### Typed Errors
- Use custom error types: `NetworkError`, `ValidationError`, `WireFormatError`
- Wrap errors with context: `fmt.Errorf("failed to parse message: %w", err)`

## Testing Standards

### Test-Driven Development (TDD)
1. **RED**: Write test first (it should fail)
2. **GREEN**: Write minimal code to make test pass
3. **REFACTOR**: Clean up code while keeping tests green

### Test Naming
- Format: `Test<TypeName>_<Method>_<Scenario>`
- Examples:
  - `TestQuerier_Query_Timeout_ReturnsEmptyResponse`
  - `TestResponder_Register_ConflictDetection`
  - `TestParseName_RFC1035_Compression`

### Test Structure
```go
func TestFoo_Bar(t *testing.T) {
    // Arrange - set up test data
    input := "test"
    
    // Act - execute the code under test
    result := Foo(input)
    
    // Assert - verify results
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Test Coverage
- Maintain ≥80% coverage (constitution requirement)
- Aim for ≥85% when touching hot paths
- 100% coverage not required (diminishing returns)
- Acceptable 0% coverage:
  - Examples (`examples/`)
  - Test hooks (internal test utilities)
  - IPv6 stubs (future work)

## Layer Boundaries (F-2 Spec)

### Import Rules
```
querier → transport → protocol, message, errors
       ↘ protocol
       ↘ message
       ↘ errors
       
responder → transport → protocol, message, errors
         ↘ state
         ↘ records
         ↘ security
```

### Violations
- ❌ NEVER import `internal/network` from public packages
- ❌ NEVER import `querier` or `responder` from `internal/`
- ✅ Validate with: `grep -rn "internal/network" querier/` (should return 0 matches)

## RFC Compliance

### Always Cite RFC Sections
```go
// Per RFC 6762 §8.1, we probe 3 times before announcing
const ProbeCount = 3

// RFC 6762 §7.1 - Known-Answer Suppression
func (rb *ResponseBuilder) ApplyKnownAnswerSuppression() { ... }
```

### RFC Sources of Truth
- RFC 6762: mDNS specification (primary)
- RFC 6763: DNS-SD specification
- RFC 1035: DNS message format
- Located in `RFC%20Docs/` directory

## Concurrency Patterns

### Context-Aware Operations
- All blocking operations MUST accept `context.Context`
- Check context cancellation in loops
- Propagate context through call chains

### Mutex Usage
- Embed `sync.RWMutex` for reader/writer locks
- Use defer for unlocks: `defer mu.Unlock()`
- Keep critical sections small

### Goroutine Management
- Always have a way to stop goroutines (context, done channel)
- Document goroutine lifecycle in Godoc
- Test for goroutine leaks in tests

## Performance Patterns

### Buffer Pooling
- Use `sync.Pool` for frequently allocated buffers
- Example: 9KB buffers in UDPv4Transport
- 99% allocation reduction achieved

### Avoid Allocations in Hot Paths
- Reuse slices/maps when possible
- Pre-allocate with capacity: `make([]byte, 0, 9000)`
- Profile with: `go test -bench=. -benchmem`

## Architecture Decision Records (ADRs)

When making significant architectural decisions:
1. Document in `docs/decisions/`
2. Format: `NNN-decision-title.md`
3. Include: Context, Decision, Consequences
4. Examples:
   - `001-transport-interface-abstraction.md`
   - `002-buffer-pooling-pattern.md`
