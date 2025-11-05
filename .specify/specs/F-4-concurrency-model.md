# F-4: Concurrency Model

**Spec ID**: F-4
**Version**: 1.1
**Type**: Architecture
**Status**: Active
**Last Updated**: 2025-11-01
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling)
**References**: BEACON_FOUNDATIONS v1.1
**Governance**: Beacon Constitution v1.0.0
**RFC Compliance**: Validated against RFC 6762 (mDNS) and RFC 6763 (DNS-SD) as of 2025-11-01

---

## Constitution Compliance

This specification adheres to the [Beacon Constitution v1.0.0](../memory/constitution.md):

- **RFC Compliant (Principle I)**: All mDNS timing requirements (probe intervals, response delays, rate limiting) strictly follow RFC 6762 §6, §8.1, and §8.3. Timer values are non-negotiable and defined by RFC mandates.
- **Spec-Driven (Principle II)**: This architecture specification governs concurrency patterns across all Beacon components before implementation.
- **Test-Driven (Principle III)**: All concurrency patterns include testing guidance with race detection (`go test -race`) as mandatory (REQ-F4-5).
- **Excellence (Principle VII)**: Follows Go best practices for concurrency, including goroutine lifecycle management, proper synchronization, and context propagation.

**RFC Validation**: This specification has been validated against RFC 6762 (Multicast DNS) and RFC 6763 (DNS-SD) to ensure timing requirements align with protocol mandates.

---

## Overview

This specification defines Beacon's concurrency model, including goroutine lifecycle, synchronization patterns, channel usage, and thread-safety guarantees. Go's concurrency primitives (goroutines, channels, mutexes) enable safe concurrent operation while maintaining clarity.

---

## Requirements

### REQ-F4-1: Goroutine Ownership
Every goroutine MUST have a clear owner responsible for its lifecycle.

**Rationale**: Prevents goroutine leaks, ensures clean shutdown.

### REQ-F4-2: Context Propagation
Long-running operations MUST accept `context.Context` for cancellation.

**Rationale**: Enables timeouts, cancellation, graceful shutdown.

### REQ-F4-3: Thread-Safe Public APIs
Public API methods MUST be safe for concurrent use unless documented otherwise.

**Rationale**: Users expect thread-safety from library APIs.

### REQ-F4-4: Synchronization Over Sharing
Prefer communication via channels over sharing memory with locks.

**Rationale**: "Don't communicate by sharing memory; share memory by communicating."

### REQ-F4-5: No Data Races
Code MUST pass `go test -race` without data race warnings.

**Rationale**: Data races cause undefined behavior.

### REQ-F4-6: RFC Timing Compliance
mDNS timing operations MUST use RFC-mandated delays (probe intervals, response delays, rate limiting).

**Rationale**: RFC 6762 mandates specific timing to ensure interoperability and avoid network flooding.

### REQ-F4-7: Timer Management
Timers MUST be properly stopped to avoid leaks. Use `defer timer.Stop()` or ensure cleanup on all paths.

**Rationale**: Leaked timers consume memory and goroutines.

### REQ-F4-8: Random Jitter
Response delays MUST use random jitter within RFC-specified ranges to avoid synchronization.

**Rationale**: RFC 6762 §6 requires randomized delays to prevent response storms.

---

## Goroutine Patterns

### Pattern 1: Request-Scoped Goroutines

For operations tied to a single request:

```go
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    resultChan := make(chan []Record, 1)
    errChan := make(chan error, 1)

    go func() {
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
- Goroutine completes when request completes
- Context cancellation stops work
- No explicit cleanup needed (goroutine exits)

### Pattern 2: Long-Running Background Goroutines

For ongoing background work:

```go
type Responder struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func (r *Responder) Start(ctx context.Context) error {
    r.ctx, r.cancel = context.WithCancel(ctx)

    r.wg.Add(1)
    go r.receiveLoop()

    r.wg.Add(1)
    go r.cacheCleanupLoop()

    return nil
}

func (r *Responder) receiveLoop() {
    defer r.wg.Done()
    for {
        select {
        case <-r.ctx.Done():
            return
        default:
            packet, err := r.transport.Receive()
            if err != nil {
                continue
            }
            r.handlePacket(packet)
        }
    }
}

func (r *Responder) Stop() error {
    r.cancel()   // Signal goroutines to stop
    r.wg.Wait()  // Wait for all goroutines to finish
    return nil
}
```

**Properties**:
- Goroutines run for lifetime of component
- Context signals shutdown
- WaitGroup ensures clean shutdown
- Stop() waits for completion

### Pattern 3: Worker Pool

For bounded concurrency:

```go
type QueryProcessor struct {
    workers   int
    queryChan chan Query
    wg        sync.WaitGroup
}

func (p *QueryProcessor) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
}

func (p *QueryProcessor) worker(ctx context.Context, id int) {
    defer p.wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case query := <-p.queryChan:
            p.processQuery(query)
        }
    }
}

func (p *QueryProcessor) Submit(query Query) {
    p.queryChan <- query
}

func (p *QueryProcessor) Stop() {
    close(p.queryChan)
    p.wg.Wait()
}
```

**Properties**:
- Fixed number of workers
- Work distributed via channel
- Bounded resource usage

---

## Synchronization Primitives

### Mutex for Exclusive Access

Use `sync.Mutex` when only one goroutine should access data:

```go
type Cache struct {
    mu      sync.Mutex
    records map[string]Record
}

func (c *Cache) Add(r Record) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.records[r.Name] = r
}

func (c *Cache) Get(name string) (Record, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    r, ok := c.records[name]
    return r, ok
}
```

**Guidelines**:
- Always `defer mu.Unlock()` immediately after `Lock()`
- Keep critical sections small
- Don't call external functions while holding lock (deadlock risk)

### RWMutex for Read-Heavy Workloads

Use `sync.RWMutex` when reads >> writes:

```go
type Cache struct {
    mu      sync.RWMutex
    records map[string]Record
}

func (c *Cache) Add(r Record) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.records[r.Name] = r
}

func (c *Cache) Get(name string) (Record, bool) {
    c.mu.RLock() // Read lock (multiple readers allowed)
    defer c.mu.RUnlock()
    r, ok := c.records[name]
    return r, ok
}
```

**When to use**:
- Reads are much more frequent than writes
- Read operations are not trivial (worth locking overhead)

### Atomic Operations

Use `sync/atomic` for simple counters/flags:

```go
type Stats struct {
    queryCount atomic.Int64
}

func (s *Stats) IncrementQueries() {
    s.queryCount.Add(1)
}

func (s *Stats) QueryCount() int64 {
    return s.queryCount.Load()
}
```

**When to use**:
- Simple counters
- Flags (0/1)
- No need for mutex overhead

---

## Channel Patterns

### Pattern 1: Result Channel (Buffered 1)

For single-value results:

```go
resultChan := make(chan Result, 1) // Buffered to avoid goroutine leak
go func() {
    result := compute()
    resultChan <- result // Non-blocking send (buffer size 1)
}()

select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultChan:
    return result
}
```

**Why buffered**: If context cancelled before receive, sender doesn't block forever.

### Pattern 2: Error Channel

For communicating errors:

```go
errChan := make(chan error, 1)
go func() {
    if err := operation(); err != nil {
        errChan <- err
    }
}()

select {
case err := <-errChan:
    return err
case <-ctx.Done():
    return ctx.Err()
}
```

### Pattern 3: Fan-Out (Multiple Receivers)

Distribute work to multiple workers:

```go
workChan := make(chan Work, 100)

for i := 0; i < numWorkers; i++ {
    go worker(workChan)
}

for _, work := range workItems {
    workChan <- work
}
close(workChan) // Signal no more work
```

### Pattern 4: Fan-In (Multiple Senders, One Receiver)

Combine results from multiple sources:

```go
func fanIn(channels ...<-chan Result) <-chan Result {
    out := make(chan Result)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan Result) {
            defer wg.Done()
            for result := range c {
                out <- result
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### Pattern 5: Done Channel (Broadcast Close)

Signal shutdown to multiple goroutines:

```go
type Server struct {
    done chan struct{}
}

func (s *Server) Start() {
    s.done = make(chan struct{})

    go s.worker1()
    go s.worker2()
}

func (s *Server) worker1() {
    for {
        select {
        case <-s.done:
            return
        default:
            // do work
        }
    }
}

func (s *Server) Stop() {
    close(s.done) // Broadcasts to all workers
}
```

**Property**: Closing channel broadcasts to all receivers.

---

## Context Usage

### Passing Context

All long-running operations MUST accept `context.Context`:

```go
// Good
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error)

// Bad - no context
func (q *Querier) Query(name string) ([]Record, error)
```

### Context Best Practices

**RULE-1**: Context is first parameter:
```go
func Operation(ctx context.Context, arg1, arg2 string) error
```

**RULE-2**: Don't store context in structs (except for lifecycle):
```go
// Bad
type Querier struct {
    ctx context.Context
}

// Good - pass per operation
func (q *Querier) Query(ctx context.Context, ...) error
```

**RULE-3**: Respect context cancellation:
```go
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultChan:
    return process(result)
}
```

**RULE-4**: Use context.WithTimeout for operations with deadlines:
```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
records, err := q.Query(ctx, name)
```

### Context Values (Use Sparingly)

Avoid using context for passing values:

```go
// Bad - passing data via context
ctx = context.WithValue(ctx, "timeout", 5*time.Second)

// Good - explicit parameters
func Query(ctx context.Context, name string, timeout time.Duration) error
```

**Exception**: Request-scoped values (trace IDs, request IDs) in server contexts.

---

## Thread-Safety Guarantees

### Public APIs

All public package APIs MUST be thread-safe:

```go
// Thread-safe: Multiple goroutines can call simultaneously
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error)
```

### Internal Packages

Internal packages MAY document thread-safety requirements:

```go
// NOT thread-safe: Caller must synchronize access
type MessageBuilder struct {
    // ...
}
```

### Documentation

Document thread-safety in GoDoc:

```go
// Querier sends mDNS queries.
// It is safe for concurrent use by multiple goroutines.
type Querier struct {
    // ...
}

// MessageBuilder constructs DNS messages.
// It is NOT safe for concurrent use. Each goroutine should have its own builder.
type MessageBuilder struct {
    // ...
}
```

---

## Common Pitfalls

### Pitfall 1: Goroutine Leaks

**Problem**: Goroutine blocks forever, never exits.

```go
// BAD - goroutine leak if context never cancelled
go func() {
    result := <-someChan // Blocks forever if channel never receives
    process(result)
}()
```

**Solution**: Use select with context:

```go
go func() {
    select {
    case <-ctx.Done():
        return
    case result := <-someChan:
        process(result)
    }
}()
```

### Pitfall 2: Closing a Channel Twice

**Problem**: `panic: close of closed channel`

```go
close(ch)
close(ch) // PANIC!
```

**Solution**: Use `sync.Once` or ensure single closer:

```go
var once sync.Once
once.Do(func() { close(ch) })
```

### Pitfall 3: Race on Shared Map

**Problem**: Concurrent map access without synchronization.

```go
// BAD - data race
var cache = make(map[string]Record)

func Add(r Record) {
    cache[r.Name] = r // RACE if multiple goroutines call Add
}
```

**Solution**: Use mutex or `sync.Map`:

```go
var (
    mu    sync.Mutex
    cache = make(map[string]Record)
)

func Add(r Record) {
    mu.Lock()
    defer mu.Unlock()
    cache[r.Name] = r
}
```

### Pitfall 4: Calling Method on Nil Receiver

**Problem**: Method called before initialization.

```go
var q *Querier
q.Query(ctx, name) // Panic if q is nil
```

**Solution**: Check for nil or ensure initialization:

```go
func (q *Querier) Query(ctx context.Context, name string) error {
    if q == nil {
        return errors.New("querier not initialized")
    }
    // ...
}
```

### Pitfall 5: Forgetting to Start Goroutine

**Problem**: Blocking operation not run in goroutine.

```go
// BAD - blocks forever
resultChan := make(chan Result)
resultChan <- compute() // Deadlock! No receiver
result := <-resultChan
```

**Solution**: Send in goroutine or use buffered channel:

```go
// Option 1: Goroutine
resultChan := make(chan Result)
go func() {
    resultChan <- compute()
}()
result := <-resultChan

// Option 2: Buffered
resultChan := make(chan Result, 1)
resultChan <- compute() // Non-blocking (buffer size 1)
result := <-resultChan
```

---

## Testing Concurrency

### Data Race Detection

Always run tests with race detector:

```bash
go test -race ./...
```

### Testing Concurrent Access

```go
func TestConcurrentAdd(t *testing.T) {
    cache := NewCache()
    var wg sync.WaitGroup

    // 100 goroutines adding concurrently
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cache.Add(Record{Name: fmt.Sprintf("record-%d", id)})
        }(i)
    }

    wg.Wait()

    // Verify all added
    if cache.Size() != 100 {
        t.Errorf("expected 100 records, got %d", cache.Size())
    }
}
```

### Testing Goroutine Lifecycle

```go
func TestGracefulShutdown(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())

    r := NewResponder()
    if err := r.Start(ctx); err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // Let it run briefly
    time.Sleep(100 * time.Millisecond)

    // Shutdown
    cancel()
    if err := r.Stop(); err != nil {
        t.Fatalf("Stop failed: %v", err)
    }

    // Verify goroutines exited (would fail with -race if leaked)
}
```

---

## Lifecycle Management

### Component Lifecycle

Components follow this pattern:

```go
type Component struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup

    mu    sync.Mutex
    state componentState
}

type componentState int

const (
    stateNew componentState = iota
    stateStarted
    stateStopped
)

func (c *Component) Start(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.state != stateNew {
        return errors.New("component already started")
    }

    c.ctx, c.cancel = context.WithCancel(ctx)
    c.state = stateStarted

    c.wg.Add(1)
    go c.run()

    return nil
}

func (c *Component) run() {
    defer c.wg.Done()
    // Component logic
}

func (c *Component) Stop() error {
    c.mu.Lock()
    if c.state != stateStarted {
        c.mu.Unlock()
        return errors.New("component not started")
    }
    c.state = stateStopped
    c.mu.Unlock()

    c.cancel()
    c.wg.Wait()
    return nil
}
```

---

## mDNS Timing Patterns

mDNS has specific timing requirements defined in RFC 6762. These patterns show how to implement them safely with Go's concurrency primitives.

### Probe Timing (RFC 6762 §8.1)

Probing MUST send 3 queries exactly 250ms apart:

```go
func (r *Responder) probe(ctx context.Context, name string) error {
    const (
        probeCount    = 3
        probeInterval = 250 * time.Millisecond
        initialDelay  = 0 // Can add 0-250ms random delay if desired
    )

    // Optional: Random delay 0-250ms before first probe
    if initialDelay > 0 {
        delay := time.Duration(rand.Int63n(int64(initialDelay)))
        timer := time.NewTimer(delay)
        defer timer.Stop()
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-timer.C:
        }
    }

    // Send 3 probes, 250ms apart
    for i := 0; i < probeCount; i++ {
        if err := r.sendProbe(ctx, name); err != nil {
            return fmt.Errorf("probe %d: %w", i+1, err)
        }

        // Wait 250ms before next probe (except after last probe)
        if i < probeCount-1 {
            timer := time.NewTimer(probeInterval)
            defer timer.Stop()

            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-timer.C:
                // Continue to next probe
            }
        }
    }

    return nil
}
```

**Key Points**:
- MUST use exactly 250ms intervals (not configurable)
- MUST send exactly 3 probes (not configurable)
- MAY add 0-250ms random delay before first probe
- Use `defer timer.Stop()` to avoid timer leaks
- Respect context cancellation

### Response Delays (RFC 6762 §6)

Responses for shared records MUST be delayed 20-120ms (random):

```go
func (r *Responder) sendResponse(ctx context.Context, response Message) error {
    delay := randomDelay(20*time.Millisecond, 120*time.Millisecond)

    timer := time.NewTimer(delay)
    defer timer.Stop()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-timer.C:
        return r.transport.Send(response)
    }
}

func randomDelay(min, max time.Duration) time.Duration {
    delta := max - min
    return min + time.Duration(rand.Int63n(int64(delta)))
}
```

**Exceptions**:
- **Unique records (sole owner)**: No delay (send immediately)
- **Probe responses (defending)**: Minimal delay (time-critical)
- **TC bit set**: 400-500ms delay (see below)

### TC Bit Delay (RFC 6762 §7.2)

When TC bit is set, delay response 400-500ms:

```go
func (r *Responder) sendTruncatedResponse(ctx context.Context, response Message) error {
    // RFC 6762 §7.2: 400-500ms delay when TC bit set
    delay := randomDelay(400*time.Millisecond, 500*time.Millisecond)

    timer := time.NewTimer(delay)
    defer timer.Stop()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-timer.C:
        response.Header.TC = true
        return r.transport.Send(response)
    }
}
```

### Announce Timing (RFC 6762 §8.3)

Announcements MUST be sent at least 1 second apart, at least 2 times:

```go
func (r *Responder) announce(ctx context.Context, records []Record) error {
    const (
        minAnnounceCount    = 2
        minAnnounceInterval = 1 * time.Second
    )

    announceCount := r.config.AnnounceCount // >= minAnnounceCount
    if announceCount < minAnnounceCount {
        announceCount = minAnnounceCount
    }

    for i := 0; i < announceCount; i++ {
        if err := r.sendAnnouncement(ctx, records); err != nil {
            return fmt.Errorf("announcement %d: %w", i+1, err)
        }

        // Wait at least 1 second before next announcement
        if i < announceCount-1 {
            timer := time.NewTimer(minAnnounceInterval)
            defer timer.Stop()

            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-timer.C:
                // Continue to next announcement
            }
        }
    }

    return nil
}
```

**Key Points**:
- MUST wait at least 1 second between announcements
- MUST send at least 2 announcements
- MAY send more than 2 (configurable, but minimum enforced)

### Rate Limiting (RFC 6762 §6)

Queries MUST NOT be sent more frequently than once per second (except probes):

```go
type RateLimiter struct {
    mu           sync.Mutex
    lastQueryTime map[string]time.Time
}

func (rl *RateLimiter) CanQuery(name string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    last, ok := rl.lastQueryTime[name]
    if !ok {
        rl.lastQueryTime[name] = time.Now()
        return true
    }

    // RFC 6762 §6: Minimum 1 second between queries for same name
    if time.Since(last) < 1*time.Second {
        return false
    }

    rl.lastQueryTime[name] = time.Now()
    return true
}

func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    if !q.rateLimiter.CanQuery(name) {
        return nil, &ProtocolError{
            Op:      "query",
            Field:   "rate limit",
            Message: "query rate limited (max 1/second per name)",
        }
    }

    return q.sendQuery(ctx, name)
}
```

**Exceptions**:
- Probe queries at 250ms intervals ARE allowed (RFC 6762 §8.1)
- This is per-name rate limiting, not global

### Timer Management Best Practices

**Pattern 1: Always Stop Timers**

```go
// Good - timer stopped in all paths
func operation(ctx context.Context) error {
    timer := time.NewTimer(5 * time.Second)
    defer timer.Stop() // Ensures cleanup

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-timer.C:
        return doWork()
    }
}
```

**Pattern 2: Ticker for Repeated Operations**

```go
func (r *Responder) cacheCleanupLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop() // Critical: stop ticker on exit

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            r.cleanupExpiredRecords()
        }
    }
}
```

**Pattern 3: Reset Timer for Retry Logic**

```go
func retryWithBackoff(ctx context.Context, operation func() error) error {
    backoff := 100 * time.Millisecond
    timer := time.NewTimer(backoff)
    defer timer.Stop()

    for {
        err := operation()
        if err == nil {
            return nil
        }

        // Exponential backoff
        backoff *= 2
        if backoff > 10*time.Second {
            backoff = 10 * time.Second
        }

        timer.Reset(backoff)

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-timer.C:
            // Retry
        }
    }
}
```

**Anti-pattern: Timer Leak**

```go
// BAD - timer leaked if context cancelled
func badOperation(ctx context.Context) error {
    timer := time.NewTimer(5 * time.Second)
    // Missing: defer timer.Stop()

    select {
    case <-ctx.Done():
        return ctx.Err() // Timer still running!
    case <-timer.C:
        return doWork()
    }
}
```

### mDNS-Specific Example: Probing with Conflict Detection

```go
func (r *Responder) probeWithConflictDetection(ctx context.Context, name string) error {
    const (
        probeCount    = 3
        probeInterval = 250 * time.Millisecond
    )

    // Channel to receive probe responses (potential conflicts)
    conflictChan := make(chan Message, 10)
    defer close(conflictChan)

    // Start listening for conflicts
    r.registerConflictListener(name, conflictChan)
    defer r.unregisterConflictListener(name)

    // Send probes
    for i := 0; i < probeCount; i++ {
        if err := r.sendProbe(ctx, name); err != nil {
            return fmt.Errorf("probe %d: %w", i+1, err)
        }

        // Wait 250ms or until conflict detected
        if i < probeCount-1 {
            timer := time.NewTimer(probeInterval)
            defer timer.Stop()

            select {
            case <-ctx.Done():
                return ctx.Err()
            case conflict := <-conflictChan:
                return &ConflictError{
                    Name:       name,
                    RecordType: "A",
                    Message:    fmt.Sprintf("conflict detected during probe %d", i+1),
                }
            case <-timer.C:
                // Continue to next probe
            }
        }
    }

    // Final check: wait brief period for late responses
    timer := time.NewTimer(250 * time.Millisecond)
    defer timer.Stop()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case conflict := <-conflictChan:
        return &ConflictError{
            Name:       name,
            RecordType: "A",
            Message:    "conflict detected after probing complete",
        }
    case <-timer.C:
        // No conflict, safe to use name
        return nil
    }
}
```

**Key Points**:
- Combines RFC timing (250ms probes) with conflict detection
- Uses select to handle multiple events (timer, conflict, cancellation)
- Properly stops all timers on all paths
- Returns appropriate error types

---

## Open Questions

**Q1**: Should we use `context.Context` in struct fields for long-running components?
- **Current**: Yes, for lifecycle management (Start/Stop pattern)
- **Concern**: Generally discouraged in Go
- **Decision**: Acceptable for component lifecycle, not for passing through operations

**Q2**: Buffered vs unbuffered channels?
- **Current**: Buffered (size 1) for result channels to avoid leaks
- **Guideline**: Unbuffered for synchronization, buffered (small) for decoupling

**Q3**: Should we provide context.Context in callbacks?
- **Example**: `OnServiceDiscovered(ctx context.Context, service Service)`
- **Decision**: Yes if callback may take time, no if trivial

---

## Success Criteria

- [ ] Goroutine lifecycle patterns defined
- [ ] Synchronization strategies documented
- [ ] Channel patterns cataloged
- [ ] Context usage guidelines established
- [ ] Thread-safety guarantees specified
- [ ] Common pitfalls identified
- [ ] All tests pass with `-race`

---

## References

### Governance
- [Beacon Constitution v1.0.0](../memory/constitution.md)
- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md)

### RFCs
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - Multicast DNS (February 2013)
  - §6: Response timing and rate limiting
  - §8.1: Probing requirements (250ms intervals, 3 probes)
  - §8.3: Announcing requirements (1 second intervals, minimum 2 announcements)
- [RFC 6763](../../RFC%20Docs/RFC-6763-DNS-SD.txt) - DNS-Based Service Discovery (February 2013)

### Go Resources
- Go Blog: [Concurrency is not parallelism](https://go.dev/blog/waza-talk)
- Go Blog: [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- Effective Go: [Concurrency](https://go.dev/doc/effective_go#concurrency)
- Go Wiki: [Common Mistakes](https://go.dev/wiki/CommonMistakes)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.1 | 2025-11-01 | Governance alignment: Updated references to Constitution v1.0.0 and FOUNDATIONS v1.1; added Constitution Compliance section; confirmed RFC validation against RFC 6762 and RFC 6763; enhanced references section with governance and RFC links |
| 1.0 | 2025-11-01 | Initial version with mDNS timing patterns (RFC 6762 §6, §8.1, §8.3), timer management, and RFC-critical requirements (REQ-F4-6, REQ-F4-7, REQ-F4-8) |
