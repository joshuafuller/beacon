# BEACON Architecture Patterns

## Key Architectural Patterns

### 1. Transport Interface Abstraction (ADR-001)

**Why**: Decouples querier/responder from network implementation, enables testing

```go
type Transport interface {
    Send(ctx context.Context, packet []byte, dest net.Addr) error
    Receive(ctx context.Context) ([]byte, net.Addr, int, error)
    Close() error
}
```

**Implementations**:
- `UDPv4Transport` - Production IPv4 multicast (224.0.0.251:5353)
- `MockTransport` - Test double for unit testing
- `UDPv6Transport` - Stub for M2 future work

**Usage**:
- Querier uses transport for send/receive
- Responder uses transport for multicast
- Tests inject MockTransport

### 2. Buffer Pooling (ADR-002)

**Why**: Eliminates 900KB/sec allocations (9KB per receive call)

**Result**: 99% allocation reduction (9000 B/op → 48 B/op)

```go
// UDPv4Transport.Receive() uses buffer pool
bufPtr := GetBuffer()      // Get 9KB buffer from pool
defer PutBuffer(bufPtr)    // Return to pool
```

**Pattern**: `sync.Pool` for frequently allocated buffers

### 3. State Machine Pattern (Responder)

**Why**: RFC 6762 §8 requires complex probing/announcing state transitions

**States**:
- `StateIdle` - Initial state
- `StateProbing` - Sending 3 probes, detecting conflicts
- `StateAnnouncing` - Sending announcements
- `StateRegistered` - Service active, responding to queries

**Transitions**:
```
Idle → Probing (Register called)
Probing → Announcing (3 probes sent, no conflicts)
Probing → Idle (conflict detected, rename required)
Announcing → Registered (announcements sent)
Registered → Idle (Unregister called)
```

### 4. Registry Pattern (Service Management)

**Why**: Thread-safe service storage with concurrent access

```go
type Registry struct {
    mu       sync.RWMutex
    services map[string]*Service  // keyed by instance name
}
```

**Operations**:
- `Register(service)` - Write lock, add service
- `Get(name)` - Read lock, retrieve service
- `Remove(name)` - Write lock, delete service
- `List()` - Read lock, list all services

### 5. Builder Pattern (Message Construction)

**Why**: Simplifies complex DNS message creation

```go
// Query building
msg := BuildQuery(name, recordType)

// Response building
response := BuildResponse(answers, additionals)
```

**Features**:
- Automatic header flags (QR, AA, RD)
- DNS name encoding (RFC 1035 §3.1)
- Cache-flush bit handling (RFC 6762 §10.2)

### 6. Conflict Detection (RFC 6762 §8.2)

**Why**: Detect and resolve service name conflicts during probing

**Lexicographic Comparison**:
- Compare records byte-by-byte
- Lower lexicographic value wins
- Ties go to simultaneous probe defender

```go
detector := NewConflictDetector()
conflict := detector.DetectConflict(ourRecords, theirRecords)
if conflict {
    newName := service.Rename()  // Appends "-2", "-3", etc.
}
```

### 7. Known-Answer Suppression (RFC 6762 §7.1)

**Why**: Reduce network traffic by suppressing redundant answers

**Pattern**:
- Check incoming query for known-answer records
- Compare with our answers (name, type, rdata, TTL >50%)
- Suppress matching records from response

```go
builder.ApplyKnownAnswerSuppression(query, response)
// Removes redundant records from response
```

### 8. Rate Limiting (RFC 6762 §6.2)

**Why**: Prevent response flooding

**Implementation**:
- Per-interface rate limiting
- Token bucket algorithm
- 100 packets/second per interface limit

```go
limiter := NewRateLimiter(100 /* per second */)
if limiter.Allow(interfaceIndex) {
    // Send response
}
```

### 9. Interface-Specific IP Resolution (007)

**Why**: Multi-interface hosts must advertise correct IP per interface (RFC 6762 §15)

**Pattern**:
```go
// Extract interface index from IP_PKTINFO control message
interfaceIndex := extractInterfaceIndex(controlMessage)

// Resolve interface-specific IP
ip := getIPv4ForInterface(interfaceIndex)

// Build A record with correct IP
aRecord := buildARecord(hostname, ip)
```

**Key Files**:
- `internal/transport/udp.go` - Control message extraction
- `responder/responder.go` - IP resolver

### 10. Functional Options Pattern

**Why**: Flexible, backward-compatible configuration

```go
// Querier options
q := querier.New(
    querier.WithTimeout(5*time.Second),
    querier.WithInterfaces(ifaces),
)

// Responder options
r := responder.New(
    responder.WithHostname("myhost.local"),
    responder.WithTransport(transport),
)
```

### 11. Context-Aware Operations (F-9)

**Why**: Proper cancellation and timeouts

**Pattern**: All blocking operations accept `context.Context`

```go
func (q *Querier) Query(ctx context.Context, name string) ([]ResourceRecord, error) {
    // Respect context deadline
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case result := <-results:
        return result, nil
    }
}
```

### 12. Typed Errors (F-3)

**Why**: Structured error handling with wrapping

**Types**:
- `NetworkError` - Network operation failures
- `ValidationError` - Input validation failures
- `WireFormatError` - DNS message parsing failures

```go
// Create typed error
return &NetworkError{
    Op:      "send",
    Details: "multicast failed",
    Err:     underlyingErr,
}

// Check error type
var netErr *NetworkError
if errors.As(err, &netErr) {
    // Handle network error
}
```

## Common Anti-Patterns to Avoid

### ❌ Don't Swallow Errors
```go
// BAD
func foo() error {
    err := bar()
    return nil  // Error lost!
}

// GOOD
func foo() error {
    return bar()  // Propagate error
}
```

### ❌ Don't Violate Layer Boundaries
```go
// BAD - querier importing internal/network
import "github.com/joshuafuller/beacon/internal/network"

// GOOD - querier using transport abstraction
import "github.com/joshuafuller/beacon/internal/transport"
```

### ❌ Don't Ignore Context
```go
// BAD - ignoring context cancellation
for range time.Tick(interval) {
    doWork()
}

// GOOD - respecting context
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(interval):
        doWork()
    }
}
```

### ❌ Don't Leak Goroutines
```go
// BAD - goroutine never stops
go func() {
    for {
        doWork()
    }
}()

// GOOD - goroutine respects done channel
done := make(chan struct{})
go func() {
    for {
        select {
        case <-done:
            return
        default:
            doWork()
        }
    }
}()
```

## Performance Considerations

### Hot Paths (Profile Before Optimizing)
- `handleQuery()` - Responder query handling
- `collectResponses()` - Querier response collection  
- `ParseMessage()` - DNS message parsing
- `BuildResponse()` - DNS message building

### Optimization Techniques Used
1. **Buffer pooling** - Reuse 9KB buffers (99% allocation reduction)
2. **Pre-allocation** - `make([]byte, 0, capacity)`
3. **Avoid string concatenation** - Use `strings.Builder`
4. **Minimize allocations** - Reuse slices/maps in loops

### Benchmarking
```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Profile CPU
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```
