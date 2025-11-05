# F-7: Resource Management

**Spec ID**: F-7
**Type**: Architecture
**Status**: Draft
**Dependencies**: F-2 (Package Structure), F-4 (Concurrency Model)
**References**: BEACON_FOUNDATIONS v1.1
**Governance**: Governed by [Beacon Constitution v1.0.0](../memory/constitution.md)
**RFC Validation**: Completed 2025-11-01 (No RFC-specific resource management requirements; implementation follows Go best practices)

---

## Overview

This specification defines Beacon's resource management strategy, including goroutine lifecycle, network connection management, memory allocation, resource limits, and graceful shutdown. Proper resource management prevents leaks, ensures clean shutdown, and maintains predictable performance.

**Constitutional Alignment**: This specification implements Principle VII (Excellence) by ensuring resource efficiency, preventing leaks, and maintaining predictable performance. While RFC 6762 and RFC 6763 do not mandate specific resource management approaches, proper resource handling is essential for enterprise-grade implementations.

---

## Requirements

### REQ-F7-1: No Resource Leaks
Components MUST NOT leak resources (goroutines, connections, memory).

**Rationale**: Leaks cause degradation over time, eventual failures.

### REQ-F7-2: Graceful Shutdown
Components MUST support graceful shutdown via `Stop()` or context cancellation.

**Rationale**: Clean shutdown prevents data loss, incomplete operations.

### REQ-F7-3: Cleanup on Error
Resources MUST be cleaned up even when errors occur.

**Rationale**: Error paths are common; leaking on errors is unacceptable.

### REQ-F7-4: Resource Limits
Components SHOULD enforce configurable resource limits.

**Rationale**: Prevents unbounded resource consumption, DoS.

### REQ-F7-5: Defer for Cleanup
Use `defer` for resource cleanup whenever possible.

**Rationale**: Ensures cleanup even on panic or error.

---

## Goroutine Management

### Pattern: Tracked Goroutines

Every goroutine must be tracked and cleanly shut down:

```go
type Component struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func (c *Component) Start(ctx context.Context) error {
    c.ctx, c.cancel = context.WithCancel(ctx)

    c.wg.Add(1)
    go c.worker()

    return nil
}

func (c *Component) worker() {
    defer c.wg.Done()
    // Worker logic
    for {
        select {
        case <-c.ctx.Done():
            return
        default:
            // do work
        }
    }
}

func (c *Component) Stop() error {
    c.cancel()   // Signal shutdown
    c.wg.Wait()  // Wait for completion
    return nil
}
```

**Properties**:
- `wg.Add(1)` before `go`
- `defer wg.Done()` first line in goroutine
- `ctx.Done()` checked in loop
- `Stop()` cancels context and waits

### Pattern: Request-Scoped Goroutines

For short-lived goroutines:

```go
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    resultChan := make(chan []Record, 1)
    errChan := make(chan error, 1)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                errChan <- fmt.Errorf("panic: %v", r)
            }
        }()

        records, err := q.performQuery(ctx, name)
        if err != nil {
            errChan <- err
            return
        }
        resultChan <- records
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case err := <-errChan:
        return nil, err
    case records := <-resultChan:
        return records, nil
    }
}
```

**Properties**:
- Goroutine completes when function returns
- Panic recovery prevents goroutine crash
- Context cancellation stops work
- Buffered channels prevent goroutine leak if cancelled

### Goroutine Leak Detection

Use testing to detect leaks:

```go
func TestNoGoroutineLeak(t *testing.T) {
    before := runtime.NumGoroutine()

    q, _ := New()
    q.Start(context.Background())
    time.Sleep(100 * time.Millisecond)
    q.Stop()

    time.Sleep(100 * time.Millisecond) // Allow cleanup

    after := runtime.NumGoroutine()
    if after > before {
        t.Errorf("goroutine leak: before=%d after=%d", before, after)
    }
}
```

---

## Network Connection Management

### Socket Lifecycle

```go
type Transport struct {
    conn4 *net.UDPConn  // IPv4 multicast socket
    conn6 *net.UDPConn  // IPv6 multicast socket
    mu    sync.Mutex
}

func (t *Transport) Open() error {
    t.mu.Lock()
    defer t.mu.Unlock()

    var err error

    // IPv4 socket
    t.conn4, err = openMulticastSocket("udp4", DefaultMulticastIPv4, DefaultPort)
    if err != nil {
        return fmt.Errorf("opening IPv4 socket: %w", err)
    }

    // IPv6 socket (cleanup conn4 on error)
    t.conn6, err = openMulticastSocket("udp6", DefaultMulticastIPv6, DefaultPort)
    if err != nil {
        t.conn4.Close() // Cleanup on error!
        t.conn4 = nil
        return fmt.Errorf("opening IPv6 socket: %w", err)
    }

    return nil
}

func (t *Transport) Close() error {
    t.mu.Lock()
    defer t.mu.Unlock()

    var errs []error

    if t.conn4 != nil {
        if err := t.conn4.Close(); err != nil {
            errs = append(errs, err)
        }
        t.conn4 = nil
    }

    if t.conn6 != nil {
        if err := t.conn6.Close(); err != nil {
            errs = append(errs, err)
        }
        t.conn6 = nil
    }

    if len(errs) > 0 {
        return fmt.Errorf("close errors: %v", errs)
    }
    return nil
}
```

**Properties**:
- Cleanup on partial failure
- Set to nil after close (idempotent)
- Mutex for thread-safety
- Return aggregated errors

### File Descriptor Limits

Monitor file descriptor usage:

```go
func (t *Transport) checkFDLimit() error {
    var rLimit syscall.Rlimit
    if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
        return err
    }

    if rLimit.Cur < 1024 {
        return fmt.Errorf("file descriptor limit too low: %d (recommend ≥1024)", rLimit.Cur)
    }

    return nil
}
```

### Socket Options

Set appropriate socket options:

```go
func openMulticastSocket(network, mcastAddr string, port int) (*net.UDPConn, error) {
    addr, err := net.ResolveUDPAddr(network, fmt.Sprintf("%s:%d", mcastAddr, port))
    if err != nil {
        return nil, err
    }

    conn, err := net.ListenUDP(network, addr)
    if err != nil {
        return nil, err
    }

    // Set socket options
    if err := conn.SetReadBuffer(DefaultReceiveBufferSize); err != nil {
        conn.Close()
        return nil, fmt.Errorf("setting read buffer: %w", err)
    }

    if err := conn.SetWriteBuffer(DefaultSendBufferSize); err != nil {
        conn.Close()
        return nil, fmt.Errorf("setting write buffer: %w", err)
    }

    return conn, nil
}
```

---

## Memory Management

### Buffer Pooling

Use `sync.Pool` for frequent allocations:

```go
var messageBufferPool = sync.Pool{
    New: func() interface{} {
        buf := make([]byte, DefaultMaxMessageSize)
        return &buf
    },
}

func (t *Transport) Receive() ([]byte, net.Addr, error) {
    bufPtr := messageBufferPool.Get().(*[]byte)
    buf := *bufPtr
    defer messageBufferPool.Put(bufPtr)

    n, addr, err := t.conn.ReadFrom(buf)
    if err != nil {
        return nil, nil, err
    }

    // Copy data (don't return pooled buffer!)
    data := make([]byte, n)
    copy(data, buf[:n])

    return data, addr, nil
}
```

**Properties**:
- Pool allocates on demand
- `Get()` retrieves buffer
- `defer Put()` returns buffer
- Copy data before returning (pool buffer reused)

### Cache Size Limits

Enforce cache limits:

```go
type Cache struct {
    mu       sync.RWMutex
    records  map[string]*Record
    maxSize  int
    eviction EvictionPolicy
}

func (c *Cache) Add(r *Record) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Check size limit
    if c.maxSize > 0 && len(c.records) >= c.maxSize {
        if err := c.evict(); err != nil {
            return fmt.Errorf("cache full: %w", err)
        }
    }

    c.records[r.Name] = r
    return nil
}

func (c *Cache) evict() error {
    // Evict based on policy (LRU, oldest, etc.)
    switch c.eviction {
    case EvictionLRU:
        return c.evictLRU()
    case EvictionOldest:
        return c.evictOldest()
    default:
        return errors.New("no eviction policy")
    }
}
```

### Memory Profiling

Support memory profiling:

```go
import _ "net/http/pprof"

// Enable pprof for debugging
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then: go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## Resource Limits

### Configuration

```go
type Config struct {
    MaxConcurrentQueries int           // 0 = unlimited
    MaxCacheSize         int           // 0 = unlimited
    MaxMessageSize       int           // bytes
    ReceiveBufferSize    int           // bytes
    SendBufferSize       int           // bytes
    MaxResponseDelay     time.Duration
}
```

### Query Concurrency Limit

```go
type Querier struct {
    sem chan struct{} // Semaphore
}

func NewQuerier(maxConcurrent int) *Querier {
    var sem chan struct{}
    if maxConcurrent > 0 {
        sem = make(chan struct{}, maxConcurrent)
    }
    return &Querier{sem: sem}
}

func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    if q.sem != nil {
        select {
        case q.sem <- struct{}{}:
            defer func() { <-q.sem }()
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }

    return q.performQuery(ctx, name)
}
```

### Rate Limiting

Enforce RFC rate limits:

```go
type RateLimiter struct {
    mu       sync.Mutex
    lastSend map[string]time.Time // name+type → last send time
    minInterval time.Duration     // RFC 6762: 1 second
}

func (r *RateLimiter) Allow(name, qtype string) bool {
    r.mu.Lock()
    defer r.mu.Unlock()

    key := name + ":" + qtype
    if last, ok := r.lastSend[key]; ok {
        if time.Since(last) < r.minInterval {
            return false // Too soon
        }
    }

    r.lastSend[key] = time.Now()
    return true
}
```

---

## Graceful Shutdown

### Shutdown Phases

1. **Stop accepting new work**
2. **Cancel ongoing work**
3. **Wait for completion**
4. **Close resources**

```go
type Server struct {
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    transport *Transport
    accepting atomic.Bool
}

func (s *Server) Start(ctx context.Context) error {
    s.ctx, s.cancel = context.WithCancel(ctx)
    s.accepting.Store(true)

    if err := s.transport.Open(); err != nil {
        return err
    }

    s.wg.Add(1)
    go s.receiveLoop()

    return nil
}

func (s *Server) receiveLoop() {
    defer s.wg.Done()

    for s.accepting.Load() {
        select {
        case <-s.ctx.Done():
            return
        default:
            packet, addr, err := s.transport.Receive()
            if err != nil {
                continue
            }
            s.handlePacket(packet, addr)
        }
    }
}

func (s *Server) Stop() error {
    // Phase 1: Stop accepting
    s.accepting.Store(false)

    // Phase 2: Cancel ongoing work
    s.cancel()

    // Phase 3: Wait for completion
    s.wg.Wait()

    // Phase 4: Close resources
    return s.transport.Close()
}
```

### Shutdown Timeout

Enforce shutdown timeout:

```go
func (s *Server) StopWithTimeout(timeout time.Duration) error {
    done := make(chan error, 1)

    go func() {
        done <- s.Stop()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timeout after %v", timeout)
    }
}
```

---

## Cleanup Patterns

### Pattern 1: Defer for Single Resource

```go
func process() error {
    file, err := os.Open("data.txt")
    if err != nil {
        return err
    }
    defer file.Close()

    // Use file
    return nil
}
```

### Pattern 2: Defer with Error Capture

```go
func process() (err error) {
    file, err := os.Open("data.txt")
    if err != nil {
        return err
    }
    defer func() {
        if cerr := file.Close(); cerr != nil && err == nil {
            err = cerr // Capture close error if no other error
        }
    }()

    return processFile(file)
}
```

### Pattern 3: Cleanup on Error

```go
func setupResources() (*Resources, error) {
    r := &Resources{}

    var err error
    r.conn1, err = openConn1()
    if err != nil {
        return nil, err
    }
    defer func() {
        if err != nil {
            r.conn1.Close() // Cleanup on error
        }
    }()

    r.conn2, err = openConn2()
    if err != nil {
        return nil, err
    }
    defer func() {
        if err != nil {
            r.conn2.Close() // Cleanup on error
        }
    }()

    return r, nil
}
```

### Pattern 4: Cleanup List

For multiple resources:

```go
type Cleanup struct {
    funcs []func() error
}

func (c *Cleanup) Add(f func() error) {
    c.funcs = append(c.funcs, f)
}

func (c *Cleanup) Run() error {
    var errs []error
    for i := len(c.funcs) - 1; i >= 0; i-- {
        if err := c.funcs[i](); err != nil {
            errs = append(errs, err)
        }
    }
    if len(errs) > 0 {
        return fmt.Errorf("cleanup errors: %v", errs)
    }
    return nil
}

// Usage
func setup() (*Resources, error) {
    var cleanup Cleanup
    defer func() {
        if err != nil {
            cleanup.Run()
        }
    }()

    conn1, err := openConn1()
    if err != nil {
        return nil, err
    }
    cleanup.Add(conn1.Close)

    conn2, err := openConn2()
    if err != nil {
        return nil, err
    }
    cleanup.Add(conn2.Close)

    return &Resources{conn1, conn2}, nil
}
```

---

## Finalizers (Use Sparingly)

Finalizers run when GC collects object, but are NOT reliable for cleanup.

**When to use**: Never for critical cleanup.
**When acceptable**: Warning users of unclosed resources.

```go
type Querier struct {
    closed atomic.Bool
}

func New() *Querier {
    q := &Querier{}
    runtime.SetFinalizer(q, func(q *Querier) {
        if !q.closed.Load() {
            log.Println("WARNING: Querier was not closed")
        }
    })
    return q
}

func (q *Querier) Close() error {
    q.closed.Store(true)
    runtime.SetFinalizer(q, nil)
    // actual cleanup
    return nil
}
```

**Never rely on finalizers for cleanup**. Always provide explicit `Close()` or `Stop()`.

---

## Testing Resource Management

### Test Leak Detection

```go
func TestNoLeaks(t *testing.T) {
    before := runtime.NumGoroutine()

    q, _ := New()
    q.Start(context.Background())
    time.Sleep(100 * time.Millisecond)
    q.Stop()

    // Force GC
    runtime.GC()
    time.Sleep(100 * time.Millisecond)

    after := runtime.NumGoroutine()
    if after > before+1 { // +1 for test goroutine tolerance
        t.Errorf("goroutine leak: before=%d after=%d", before, after)
    }
}
```

### Test Graceful Shutdown

```go
func TestGracefulShutdown(t *testing.T) {
    s := NewServer()
    s.Start(context.Background())

    time.Sleep(100 * time.Millisecond)

    // Shutdown should complete quickly
    done := make(chan error, 1)
    go func() {
        done <- s.Stop()
    }()

    select {
    case err := <-done:
        if err != nil {
            t.Errorf("Stop() error: %v", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("Stop() timeout")
    }
}
```

---

## Open Questions

**Q1**: Should we use `io.Closer` interface for all closeable resources?
- **Pro**: Standard interface
- **Con**: Some resources have `Stop()` not `Close()`
- **Decision**: Use `io.Closer` where appropriate, `Stop()` for active components

**Q2**: Automatic resource cleanup via context?
- **Example**: Resources cleaned up when context cancelled
- **Pro**: Convenient
- **Con**: Implicit, harder to reason about
- **Decision**: Explicit `Close()`/`Stop()` required, context for signaling only

**Q3**: Resource pooling for connections?
- **Decision**: Not initially, single connection per component

---

## Success Criteria

- [ ] No goroutine leaks
- [ ] No socket leaks
- [ ] No memory leaks
- [ ] Graceful shutdown works
- [ ] Resource limits enforced
- [ ] Tests verify leak-free operation
- [ ] `defer` used for cleanup

---

## Constitutional Compliance

This specification aligns with the [Beacon Constitution v1.0.0](../memory/constitution.md):

### Principle I: RFC Compliant
**Status**: ✅ **COMPLIANT**

- ✅ **No RFC resource management mandates**: RFC 6762 and RFC 6763 do not specify resource management implementation details
- ✅ **Supports RFC compliance**: Resource management patterns enable reliable mDNS/DNS-SD protocol implementation without interfering with RFC requirements
- ✅ **Follows Go best practices**: Implementation follows industry standards for enterprise-grade resource management

**Evidence**: REQ-F7-1 (No Resource Leaks), REQ-F7-2 (Graceful Shutdown) ensure protocol operations remain reliable and predictable, supporting RFC compliance without mandating specific resource approaches.

### Principle II: Spec-Driven Development
**Status**: ✅ **COMPLIANT**

- ✅ **Architecture specification exists**: This specification defines resource management patterns before implementation
- ✅ **Patterns govern implementation**: All resource management code will reference these patterns
- ✅ **Clear boundaries**: Resource management patterns integrate with package structure (F-2)

**Evidence**: This F-7 specification exists and defines patterns before M1 implementation begins.

### Principle III: Test-Driven Development
**Status**: ✅ **COMPLIANT**

- ✅ **Testable cleanup patterns**: REQ-F7-1 (No Resource Leaks) is verifiable via goroutine leak detection tests
- ✅ **Explicit test patterns**: Specification includes dedicated testing section (lines 690-741)
- ✅ **Leak detection tests**: Tests verify goroutine cleanup, connection cleanup, and memory management
- ✅ **Graceful shutdown tests**: REQ-F7-2 testable via shutdown verification

**Evidence**: Section "Testing Resource Management" (lines 690-741) defines specific test patterns for leak detection, shutdown verification, and resource cleanup validation.

### Principle IV: Phased Approach
**Status**: ✅ **COMPLIANT**

- ✅ **Incremental implementation**: Resource management patterns will be implemented across milestones:
  - M1: Basic goroutine tracking and cleanup (REQ-F7-1, REQ-F7-2)
  - M2: Network connection management (REQ-F7-3)
  - M3: Memory pooling and cache limits (REQ-F7-4)
  - M4+: Resource limits and rate limiting (REQ-F7-5)

### Principle V: Open Source
**Status**: ✅ **COMPLIANT**

- ✅ **Public specification**: This specification is publicly available in repository
- ✅ **MIT License**: All implementations will be publicly available under MIT license
- ✅ **Transparent patterns**: Resource management patterns are documented and accessible to all contributors

### Principle VI: Maintained
**Status**: ✅ **COMPLIANT**

- ✅ **Long-term stability**: Resource management is critical for preventing degradation over time
- ✅ **Version controlled**: Specification uses semantic versioning (current: 1.1)
- ✅ **Prevents leaks**: REQ-F7-1 ensures no resource leaks that could accumulate over time
- ✅ **Reliable operation**: Patterns establish predictable performance for maintained systems

**Evidence**: REQ-F7-1 through REQ-F7-5 prevent common issues that cause long-term degradation (leaks, unclean shutdown, resource exhaustion).

### Principle VII: Excellence
**Status**: ✅ **COMPLIANT**

- ✅ **Industry best practices**: Follows Go community standards for resource management
- ✅ **No leaks**: REQ-F7-1 enforces leak-free operation
- ✅ **Graceful shutdown**: REQ-F7-2 ensures clean component lifecycle
- ✅ **Predictable performance**: REQ-F7-4 (Resource Limits) and REQ-F7-5 (Defer for Cleanup) maintain consistent performance
- ✅ **Enterprise-grade**: Patterns support long-running, production services

**Evidence**:
- REQ-F7-1: No Resource Leaks (excellence in reliability)
- REQ-F7-2: Graceful Shutdown (excellence in lifecycle management)
- REQ-F7-3: Cleanup on Error (excellence in error handling)
- REQ-F7-4: Resource Limits (excellence in stability)
- REQ-F7-5: Defer for Cleanup (excellence in code quality)

**Overall Assessment**: This specification fully aligns with all constitutional principles and establishes resource management patterns essential for enterprise-grade, RFC-compliant mDNS/DNS-SD implementation.

---

## References

### Technical Sources of Truth (RFCs)

**Note**: RFC 6762 and RFC 6763 are the **PRIMARY TECHNICAL AUTHORITY** for Beacon. However, these RFCs do not mandate specific resource management implementation details. This specification follows Go best practices for resource management to support reliable mDNS/DNS-SD protocol implementation.

**Critical**: Per Constitution Principle I, RFC requirements override all other concerns. While RFCs don't specify resource management approaches, proper resource handling is essential for implementing RFC-compliant protocol operations reliably.

- [RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - No specific resource management requirements; implementation must support RFC-mandated protocol operations
- [RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt) - No specific resource management requirements; resource patterns enable reliable DNS-SD operation

### Project Governance

- [Beacon Constitution v1.0.0](../memory/constitution.md) - Principle VII (Excellence) requires resource efficiency, no leaks, graceful shutdown, and predictable performance

### Foundational Knowledge

- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md) - Architecture overview and terminology

### Architecture Specifications

- [F-2: Package Structure](./F-2-package-structure.md) - Component organization and lifecycle boundaries
- [F-4: Concurrency Model](./F-4-concurrency-model.md) - Goroutine lifecycle patterns and context propagation

### Go Best Practices

- [Effective Go: Defer](https://go.dev/doc/effective_go#defer) - Cleanup pattern guidance
- [Go Blog: Concurrency is not parallelism](https://go.dev/blog/waza-talk) - Goroutine management principles
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof) - Memory and goroutine leak detection
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency) - Channel and goroutine patterns
- [Go Wiki: Common Mistakes - Goroutine Leaks](https://go.dev/wiki/CommonMistakes) - Leak prevention guidance
- [`sync.Pool` documentation](https://pkg.go.dev/sync#Pool) - Memory pooling for performance
- [`runtime.SetFinalizer` documentation](https://pkg.go.dev/runtime#SetFinalizer) - Use sparingly, cleanup patterns preferred

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2025-11-01 | System | Updated to Constitution v1.0.0, BEACON_FOUNDATIONS v1.1; added RFC validation status and constitutional alignment section |
| 1.0 | 2025-10-31 | System | Initial draft |
