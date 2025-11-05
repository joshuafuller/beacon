# F-3: Error Handling Strategy

**Spec ID**: F-3
**Type**: Architecture
**Status**: Validated (2025-11-01)
**Dependencies**: F-2 (Package Structure)
**References**:
- BEACON_FOUNDATIONS v1.1
- RFC 6762 (Multicast DNS)
- RFC 6763 (DNS-SD)
- Beacon Constitution v1.0.0

**Revision Notes**:
- **2025-11-01**: RFC validation completed against RFC 6762 §18 (Security Considerations), RFC 6763 §6 (TXT record format), and §7 (service names)
- Added mDNS-specific error types (TruncationError, WireFormatError)
- Added DNS-SD validation error examples (service type, TXT records)
- Enhanced error scenarios matrix for RFC compliance
- Updated governance references to Constitution v1.0.0

---

## Overview

This specification defines Beacon's error handling strategy, including error types, wrapping conventions, sentinel errors, and user-facing error messages. Effective error handling enables:
- **Debuggability**: Sufficient context to diagnose issues
- **Recoverability**: Errors classified for appropriate handling
- **Usability**: Clear error messages for application developers

**Constitutional Alignment**: This specification supports Principle I (RFC Compliant) by defining error types that enforce RFC requirements, Principle III (TDD) by making errors testable and predictable, and Principle VII (Excellence) by providing clear, actionable error messages for best-in-class user experience.

**RFC Compliance**: All error types and handling patterns are designed to enforce RFC 6762 and RFC 6763 requirements, particularly:
- RFC 6762 §18 (Security) - WireFormatError for malformed packets
- RFC 6762 §7.2 (Truncation) - TruncationError for TC bit handling
- RFC 6763 §6 (TXT records) - Validation errors for TXT format
- RFC 6763 §7 (Service names) - Validation errors for service type format

---

## Requirements

### REQ-F3-1: Explicit Error Returns
Functions that can fail MUST return `error` as the last return value.

**Rationale**: Go idiom, explicit error handling.

### REQ-F3-2: Error Wrapping
Errors SHOULD be wrapped with context using `fmt.Errorf("context: %w", err)`.

**Rationale**: Preserves error chain for diagnosis while adding context.

### REQ-F3-3: Error Type Safety
Error types MUST be comparable using `errors.Is()` and `errors.As()`.

**Rationale**: Go 1.13+ error handling best practices.

### REQ-F3-4: Non-Nil Error Checking
Callers MUST check errors before using other return values.

**Pattern**:
```go
result, err := Operation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
// use result
```

### REQ-F3-5: User-Friendly Messages
Public API errors MUST have clear, actionable messages for application developers.

**Bad**: `"error: 0x4a"`
**Good**: `"failed to join multicast group 224.0.0.251: permission denied (requires CAP_NET_RAW or root)"`

---

## Error Categories

Beacon errors fall into eight categories:

### 1. Protocol Errors
Violations of DNS/mDNS/DNS-SD protocol.

**Examples**:
- Malformed DNS message
- Invalid label encoding
- RFC compliance violation
- Unsupported record type

**Type**: `*ProtocolError`

**Handling**: Usually not recoverable, log and discard packet.

### 2. Network Errors
Network I/O failures.

**Examples**:
- Socket creation failure
- Permission denied (port 5353, multicast)
- Network unreachable
- Timeout
- Interface down

**Type**: `*NetworkError` (or wrapped `net.Error`)

**Handling**: May be recoverable (retry), may be permanent (permission), may be transient (interface down).

### 3. Validation Errors
Invalid input from user/application.

**Examples**:
- Invalid domain name format
- Invalid service type (RFC 6763 §7)
- Invalid TXT record format (RFC 6763 §6)
- Empty required field
- TTL out of range
- Service name exceeds 15 characters

**Type**: `*ValidationError`

**Handling**: Not recoverable, user must fix input.

### 4. Conflict Errors
mDNS name conflicts during probing or operation.

**Examples**:
- Name already in use
- Probing detected conflict (RFC 6762 §8.1)
- Too many conflicts (rate limited)

**Type**: `*ConflictError`

**Handling**: Recoverable via rename or user intervention.

### 5. Resource Errors
System resource exhaustion.

**Examples**:
- Out of memory
- Too many open sockets
- Message exceeds maximum size
- Cache full

**Type**: `*ResourceError`

**Handling**: May be transient (retry after cleanup) or permanent (message too large).

### 6. Truncation Errors
mDNS message truncation handling (RFC 6762 §7.2).

**Examples**:
- Message truncated (TC bit set)
- Known-Answer list truncated
- Additional records truncated

**Type**: `*TruncationError`

**Handling**: Query again without known-answer suppression, or handle partial response.

### 7. Wire Format Errors
Malformed packets that may indicate security issues (RFC 6762 §18).

**Examples**:
- Invalid name compression pointer
- Label exceeds 63 bytes
- Message exceeds packet bounds
- Invalid UTF-8 in labels

**Type**: `*WireFormatError`

**Handling**: Not recoverable, discard packet, may indicate attack.

### 8. Context Errors
Context cancellation or timeout.

**Examples**:
- `context.Canceled`
- `context.DeadlineExceeded`

**Type**: Standard `context` package errors

**Handling**: Operation cancelled by user or timeout, not an error condition per se.

---

## Error Types

### Base Error Interface

All Beacon errors implement standard `error` interface:
```go
type error interface {
    Error() string
}
```

### Protocol Error

```go
// ProtocolError represents a DNS/mDNS/DNS-SD protocol violation.
type ProtocolError struct {
    Op      string // Operation that failed (e.g., "parse message")
    Field   string // Field that caused error (e.g., "header.flags")
    Message string // Human-readable description
    Err     error  // Underlying error (if any)
}

func (e *ProtocolError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %s: %v", e.Op, e.Field, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s: %s", e.Op, e.Field, e.Message)
}

func (e *ProtocolError) Unwrap() error {
    return e.Err
}
```

**Example**:
```go
return &ProtocolError{
    Op:      "parse header",
    Field:   "qdcount",
    Message: "question count exceeds packet size",
}
// Error: "parse header: qdcount: question count exceeds packet size"
```

### Network Error

```go
// NetworkError represents a network I/O failure.
type NetworkError struct {
    Op      string // Operation (e.g., "send query", "join multicast group")
    Addr    string // Address involved (if applicable)
    Message string // Human-readable description
    Err     error  // Underlying error (often net.Error)
}

func (e *NetworkError) Error() string {
    if e.Addr != "" {
        return fmt.Sprintf("%s [%s]: %s", e.Op, e.Addr, e.Message)
    }
    return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *NetworkError) Unwrap() error {
    return e.Err
}

func (e *NetworkError) Timeout() bool {
    if ne, ok := e.Err.(net.Error); ok {
        return ne.Timeout()
    }
    return false
}

func (e *NetworkError) Temporary() bool {
    if ne, ok := e.Err.(net.Error); ok {
        return ne.Temporary()
    }
    return false
}
```

**Example**:
```go
return &NetworkError{
    Op:      "join multicast group",
    Addr:    "224.0.0.251",
    Message: "permission denied (requires CAP_NET_RAW or root)",
    Err:     syscall.EPERM,
}
// Error: "join multicast group [224.0.0.251]: permission denied (requires CAP_NET_RAW or root)"
```

### Validation Error

```go
// ValidationError represents invalid user input.
type ValidationError struct {
    Field   string // Field that failed validation
    Value   string // Invalid value (if safe to include)
    Message string // What's wrong
}

func (e *ValidationError) Error() string {
    if e.Value != "" {
        return fmt.Sprintf("invalid %s: %s: %s", e.Field, e.Value, e.Message)
    }
    return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}
```

**Example**:
```go
return &ValidationError{
    Field:   "service type",
    Value:   "_invalid service_._tcp",
    Message: "service name contains spaces (must be letters, digits, hyphens only)",
}
// Error: "invalid service type: _invalid service_._tcp: service name contains spaces"
```

### Conflict Error

```go
// ConflictError represents an mDNS name conflict.
type ConflictError struct {
    Name      string   // Name that conflicts
    RecordType string  // Record type (A, AAAA, SRV, etc.)
    Message   string   // Description
    Conflicts int      // Number of conflicts detected (for rate limiting)
}

func (e *ConflictError) Error() string {
    if e.Conflicts > 1 {
        return fmt.Sprintf("name conflict: %s (%s): %s (%d conflicts)",
            e.Name, e.RecordType, e.Message, e.Conflicts)
    }
    return fmt.Sprintf("name conflict: %s (%s): %s", e.Name, e.RecordType, e.Message)
}
```

**Example**:
```go
return &ConflictError{
    Name:       "myhost.local",
    RecordType: "A",
    Message:    "another host is using this name",
    Conflicts:  1,
}
// Error: "name conflict: myhost.local (A): another host is using this name"
```

### Resource Error

```go
// ResourceError represents system resource exhaustion.
type ResourceError struct {
    Resource string // Resource that's exhausted (e.g., "memory", "sockets")
    Op       string // Operation that failed
    Message  string // Description
    Err      error  // Underlying error (if any)
}

func (e *ResourceError) Error() string {
    return fmt.Sprintf("%s: %s: %s", e.Op, e.Resource, e.Message)
}

func (e *ResourceError) Unwrap() error {
    return e.Err
}
```

**Example**:
```go
return &ResourceError{
    Resource: "message size",
    Op:       "serialize message",
    Message:  "message exceeds 9000 byte limit for multicast",
}
// Error: "serialize message: message size: message exceeds 9000 byte limit"
```

### Truncation Error

```go
// TruncationError represents mDNS message truncation (RFC 6762 §7.2).
// When TC bit is set, responder indicates message was truncated.
type TruncationError struct {
    Name     string // Name being queried
    Truncated string // What was truncated (e.g., "known answers", "additional records")
    Message   string // Description
}

func (e *TruncationError) Error() string {
    return fmt.Sprintf("truncated response for %s: %s: %s", e.Name, e.Truncated, e.Message)
}
```

**Example**:
```go
return &TruncationError{
    Name:      "myhost.local",
    Truncated: "known answers",
    Message:   "TC bit set, query again without known-answer suppression",
}
// Error: "truncated response for myhost.local: known answers: TC bit set, query again without known-answer suppression"
```

**RFC 6762 §7.2**: When TC bit is set, querier SHOULD reissue query without known-answer section.

### Wire Format Error

```go
// WireFormatError represents a malformed DNS packet (RFC 6762 §18).
// May indicate security issue (buffer overflow attempt, malicious packet).
type WireFormatError struct {
    Op      string // Operation (e.g., "parse message", "decompress name")
    Field   string // Field that's malformed (e.g., "name pointer", "label")
    Offset  int    // Byte offset in packet where error occurred
    Message string // Description
    Err     error  // Underlying error (if any)
}

func (e *WireFormatError) Error() string {
    if e.Offset >= 0 {
        return fmt.Sprintf("%s: %s at offset %d: %s", e.Op, e.Field, e.Offset, e.Message)
    }
    return fmt.Sprintf("%s: %s: %s", e.Op, e.Field, e.Message)
}

func (e *WireFormatError) Unwrap() error {
    return e.Err
}
```

**Example**:
```go
return &WireFormatError{
    Op:      "decompress name",
    Field:   "name pointer",
    Offset:  42,
    Message: "pointer points beyond packet boundary (possible attack)",
}
// Error: "decompress name: name pointer at offset 42: pointer points beyond packet boundary (possible attack)"
```

**RFC 6762 §18**: Implementers should guard against malformed packets, name compression loops, and other attacks.

---

## Sentinel Errors

Sentinel errors are predefined `error` values for common conditions.

### When to Use Sentinels
Use for:
- Well-known error conditions
- Errors that callers may want to handle specifically
- Errors without additional context needed

### Definition Pattern

```go
package beacon

import "errors"

var (
    // ErrClosed indicates operation on closed resource.
    ErrClosed = errors.New("beacon: resource closed")

    // ErrNotFound indicates requested record not found.
    ErrNotFound = errors.New("beacon: not found")

    // ErrTimeout indicates operation timed out.
    ErrTimeout = errors.New("beacon: timeout")

    // ErrInvalidName indicates invalid domain name.
    ErrInvalidName = errors.New("beacon: invalid domain name")

    // ErrConflict indicates name conflict detected.
    ErrConflict = errors.New("beacon: name conflict")

    // ErrTruncated indicates message was truncated (TC bit set).
    ErrTruncated = errors.New("beacon: message truncated")

    // ErrMalformed indicates malformed wire format packet.
    ErrMalformed = errors.New("beacon: malformed packet")

    // ErrInvalidServiceType indicates invalid DNS-SD service type.
    ErrInvalidServiceType = errors.New("beacon: invalid service type")

    // ErrInvalidTXTRecord indicates invalid TXT record format.
    ErrInvalidTXTRecord = errors.New("beacon: invalid TXT record")
)
```

### Usage

**Returning**:
```go
if cache.Get(name) == nil {
    return nil, ErrNotFound
}
```

**Checking**:
```go
record, err := cache.Get("myhost.local")
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

**Wrapping**:
```go
if err := probe(name); err != nil {
    if errors.Is(err, ErrConflict) {
        return fmt.Errorf("probing failed: %w", err)
    }
}
```

---

## Error Wrapping

### Wrapping Pattern

Add context while preserving error chain:

```go
func HighLevel() error {
    if err := MidLevel(); err != nil {
        return fmt.Errorf("high level operation: %w", err)
    }
    return nil
}

func MidLevel() error {
    if err := LowLevel(); err != nil {
        return fmt.Errorf("mid level: %w", err)
    }
    return nil
}

func LowLevel() error {
    return errors.New("low level failure")
}
```

**Result**: `"high level operation: mid level: low level failure"`

**Unwrapping**: `errors.Unwrap()` or `errors.Is()` can check original error.

### What Context to Add

Good context answers:
- **What** were you doing?
- **Where** in the process did it fail?
- **Which** resource/name/operation?

**Example**:
```go
if err := sendQuery(query, addr); err != nil {
    return fmt.Errorf("sending query for %s to %s: %w", query.Name, addr, err)
}
```

### When NOT to Wrap

Don't wrap if adding no new information:

**Bad**:
```go
if err != nil {
    return fmt.Errorf("error: %w", err) // No new info!
}
```

**Good**:
```go
if err != nil {
    return err // Just return as-is
}
```

---

## Error Handling Patterns

### Pattern 1: Check and Wrap

```go
func Process() error {
    data, err := Fetch()
    if err != nil {
        return fmt.Errorf("fetching data: %w", err)
    }

    if err := Validate(data); err != nil {
        return fmt.Errorf("validating data: %w", err)
    }

    return nil
}
```

### Pattern 2: Check Specific Error

```go
record, err := cache.Lookup(name)
if err != nil {
    if errors.Is(err, ErrNotFound) {
        // Handle not found specifically
        return defaultRecord, nil
    }
    return nil, fmt.Errorf("cache lookup: %w", err)
}
```

### Pattern 3: Type Assertion for Details

```go
if err := operation(); err != nil {
    var netErr *NetworkError
    if errors.As(err, &netErr) {
        if netErr.Timeout() {
            // Retry on timeout
            return retry(operation)
        }
    }
    return fmt.Errorf("operation failed: %w", err)
}
```

### Pattern 4: Cleanup on Error

```go
func Start() error {
    conn, err := openConnection()
    if err != nil {
        return fmt.Errorf("opening connection: %w", err)
    }
    defer func() {
        if err != nil {
            conn.Close() // Cleanup if subsequent operations fail
        }
    }()

    if err = configure(conn); err != nil {
        return fmt.Errorf("configuring connection: %w", err)
    }

    err = nil // Success, don't close
    return nil
}
```

### Pattern 5: Context Cancellation

```go
func Query(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // Return context.Canceled or context.DeadlineExceeded
    case result := <-resultChan:
        return processResult(result)
    }
}
```

---

## Logging vs Returning

### RULE-1: Return errors to caller
Functions SHOULD return errors, not just log them.

**Rationale**: Caller decides how to handle (retry, ignore, propagate, log).

### RULE-2: Log at boundaries
Log errors at system boundaries (public API entry points, goroutines).

**Example**:
```go
// Public API - log here
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    records, err := q.query(ctx, name)
    if err != nil {
        log.Printf("query failed: name=%s err=%v", name, err)
        return nil, err
    }
    return records, nil
}

// Internal function - just return error
func (q *Querier) query(ctx context.Context, name string) ([]Record, error) {
    // Don't log here, let caller handle
    return nil, fmt.Errorf("not implemented")
}
```

### RULE-3: Don't log and return
Avoid logging AND returning the same error (causes duplicate logs).

**Anti-pattern**:
```go
func process() error {
    if err := step1(); err != nil {
        log.Printf("step1 failed: %v", err) // DON'T
        return err                           // Now caller might log too
    }
}
```

**Better**:
```go
func process() error {
    if err := step1(); err != nil {
        return fmt.Errorf("step1: %w", err) // Just return with context
    }
}
```

---

## Testing Error Handling

### Test Error Construction

```go
func TestProtocolError(t *testing.T) {
    err := &ProtocolError{
        Op:      "parse",
        Field:   "header",
        Message: "invalid flags",
    }

    want := "parse: header: invalid flags"
    if got := err.Error(); got != want {
        t.Errorf("Error() = %q, want %q", got, want)
    }
}
```

### Test Error Wrapping

```go
func TestErrorWrapping(t *testing.T) {
    base := errors.New("base error")
    wrapped := fmt.Errorf("context: %w", base)

    if !errors.Is(wrapped, base) {
        t.Error("wrapped error should match base")
    }
}
```

### Test Error Handling

```go
func TestHandlesNotFound(t *testing.T) {
    cache := NewCache()
    _, err := cache.Lookup("nonexistent")

    if !errors.Is(err, ErrNotFound) {
        t.Errorf("expected ErrNotFound, got %v", err)
    }
}
```

---

## Documentation

### Error Documentation in GoDoc

Document what errors a function can return:

```go
// Query sends an mDNS query and returns matching records.
//
// Returns:
//   - ValidationError if name is invalid
//   - NetworkError if send fails
//   - context.Canceled if ctx is cancelled
//   - context.DeadlineExceeded if ctx times out
func Query(ctx context.Context, name string) ([]Record, error) {
    // ...
}
```

### Error Messages

Error messages SHOULD:
- Be lowercase (unless proper noun)
- Not end with punctuation
- Be specific and actionable
- Include relevant values (names, addresses)

**Good**:
```
"failed to join multicast group 224.0.0.251: permission denied"
"invalid service name '_my service': contains spaces"
"name conflict for myhost.local: already in use"
```

**Bad**:
```
"Error!" // vague
"An error occurred while processing your request" // generic
"0x4a" // not human-readable
```

---

## RFC Error Scenarios Matrix

This matrix maps RFC requirements to appropriate error types for implementation guidance.

### mDNS Protocol Errors (RFC 6762)

| RFC Section | Scenario | Error Type | Sentinel Error | Handling |
|-------------|----------|------------|----------------|----------|
| §6 | Response delay violation | ProtocolError | - | Log and continue |
| §7.2 | Message truncated (TC bit) | TruncationError | ErrTruncated | Requery without known-answers |
| §8.1 | Probe conflict detected | ConflictError | ErrConflict | Rename or abort |
| §8.2 | Simultaneous probe tiebreaking | ConflictError | ErrConflict | Compare records lexicographically |
| §8.3 | Announce rate limiting | ResourceError | - | Back off, retry later |
| §17 | Message exceeds 9000 bytes | ResourceError | - | Split message or reject |
| §18 | Name compression loop | WireFormatError | ErrMalformed | Discard packet |
| §18 | Label exceeds 63 bytes | WireFormatError | ErrMalformed | Discard packet |
| §18 | Invalid UTF-8 in label | WireFormatError | ErrMalformed | Discard packet |
| §18 | Pointer beyond packet | WireFormatError | ErrMalformed | Discard packet (attack?) |

### DNS-SD Protocol Errors (RFC 6763)

| RFC Section | Scenario | Error Type | Sentinel Error | Handling |
|-------------|----------|------------|----------------|----------|
| §6 | TXT record exceeds 1300 bytes | ValidationError | ErrInvalidTXTRecord | Reject or warn |
| §6 | TXT pair exceeds 255 bytes | ValidationError | ErrInvalidTXTRecord | Reject |
| §6 | Malformed key=value pair | ValidationError | ErrInvalidTXTRecord | Reject |
| §7 | Service name exceeds 15 chars | ValidationError | ErrInvalidServiceType | Reject |
| §7 | Invalid service type format | ValidationError | ErrInvalidServiceType | Reject |
| §7 | Missing underscore prefix | ValidationError | ErrInvalidServiceType | Reject |
| §7 | Invalid protocol (_tcp/_udp) | ValidationError | ErrInvalidServiceType | Reject |

### Network Errors

| Scenario | Error Type | Sentinel Error | Handling |
|----------|------------|----------------|----------|
| Port 5353 bind fails (permission) | NetworkError | - | Report clear error with sudo hint |
| Multicast join fails | NetworkError | - | Retry or skip interface |
| Interface down | NetworkError | - | Skip interface or wait |
| Send timeout | NetworkError | ErrTimeout | Retry or report |
| Receive buffer full | NetworkError | - | Increase buffer or drop |

### Usage Examples

**Truncation handling**:
```go
resp, err := sendQuery(query)
if err != nil {
    var truncErr *TruncationError
    if errors.As(err, &truncErr) {
        // RFC 6762 §7.2: Requery without known-answer suppression
        return sendQuery(queryWithoutKnownAnswers(query))
    }
    return nil, err
}
```

**Wire format security**:
```go
msg, err := parseMessage(packet)
if err != nil {
    var wireErr *WireFormatError
    if errors.As(err, &wireErr) {
        // RFC 6762 §18: Possible attack, log and discard
        log.Warn("malformed packet", "offset", wireErr.Offset, "err", err)
        return nil // Discard packet
    }
    return nil, err
}
```

**DNS-SD validation**:
```go
func validateServiceType(serviceType string) error {
    if !strings.HasPrefix(serviceType, "_") {
        return &ValidationError{
            Field:   "service type",
            Value:   serviceType,
            Message: "must start with underscore (RFC 6763 §7)",
        }
    }

    parts := strings.Split(serviceType, ".")
    if len(parts) != 2 {
        return &ValidationError{
            Field:   "service type",
            Value:   serviceType,
            Message: "must be '_<servicename>._tcp' or '_<servicename>._udp' (RFC 6763 §7)",
        }
    }

    serviceName := strings.TrimPrefix(parts[0], "_")
    if len(serviceName) > 15 {
        return &ValidationError{
            Field:   "service name",
            Value:   serviceName,
            Message: fmt.Sprintf("exceeds 15 characters (%d) (RFC 6763 §7)", len(serviceName)),
        }
    }

    return nil
}
```

**Probe conflict handling**:
```go
func probe(name string) error {
    if err := sendProbes(name); err != nil {
        var conflictErr *ConflictError
        if errors.As(err, &conflictErr) {
            // RFC 6762 §8.1: Conflict detected, must rename
            if conflictErr.Conflicts > 10 {
                return fmt.Errorf("too many conflicts, giving up: %w", err)
            }
            return probe(renameName(name)) // Try again with new name
        }
        return err
    }
    return nil
}
```

---

## Open Questions

**Q1**: Should we use structured errors with fields, or wrapped strings?
- **Current**: Structured (ProtocolError, NetworkError types)
- **Alternative**: Just wrap with fmt.Errorf
- **Decision**: Structured for categorization, wrapped for context

**Q2**: Error codes (numeric)?
- **Pro**: Machine-readable
- **Con**: Non-idiomatic in Go
- **Decision**: No error codes, use errors.Is() and types

**Q3**: Stack traces?
- **Pro**: Helps debugging
- **Con**: Not standard in Go, adds dependency
- **Decision**: No stack traces initially, rely on error wrapping for context

---

## Success Criteria

- [x] Error types defined for all categories
- [x] Sentinel errors documented
- [x] Wrapping conventions established
- [x] Error messages clear and actionable
- [x] Tests verify error handling
- [x] Documentation includes error returns
- [x] RFC validation completed (2025-11-01)
- [x] Constitutional alignment verified

---

## Constitution Check

**Principle I - RFC Compliant**: ✅ PASS
- Error types enforce RFC 6762 §18 (WireFormatError for security)
- TruncationError implements RFC 6762 §7.2 requirements
- ValidationError enforces RFC 6763 §6 (TXT) and §7 (service names)
- All protocol violations result in appropriate error types
- RFC references documented in error scenarios matrix

**Principle II - Spec-Driven Development**: ✅ PASS
- This specification defines error handling before implementation
- Error types, patterns, and scenarios fully documented
- Provides clear guidance for developers

**Principle III - Test-Driven Development**: ✅ PASS
- Error types are testable (constructable, comparable)
- Testing patterns documented
- Examples show how to test error handling
- Sentinel errors support errors.Is() for testing

**Principle VII - Excellence**: ✅ PASS
- Clear, actionable error messages defined
- User-friendly guidance (e.g., "requires CAP_NET_RAW or root")
- Best practices from Go ecosystem incorporated
- Error documentation standards established

**Overall Assessment**: This specification aligns with all relevant constitutional principles and enforces RFC compliance through structured error types.

---

## References

- [Beacon Constitution v1.0.0](../memory/constitution.md)
- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md)
- [RFC 6762 - Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)
- [RFC 6763 - DNS-SD](../../RFC%20Docs/RFC-6763-DNS-SD.txt)
- Go Blog: [Error Handling](https://go.dev/blog/error-handling-and-go)
- Go Blog: [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- Effective Go: [Errors](https://go.dev/doc/effective_go#errors)

---

## Version History

| Version | Date | Changes | Validated Against |
|---------|------|---------|-------------------|
| 1.2 | 2025-11-01 | Validated against Constitution v1.0.0 and BEACON_FOUNDATIONS v1.1; RFC validation completed | RFC 6762, RFC 6763, Constitution v1.0.0 |
| 1.1 | 2025-11-01 | Added mDNS-specific error types, DNS-SD validation examples, RFC scenarios matrix | - |
| 1.0 | 2025-10-31 | Initial specification | - |
