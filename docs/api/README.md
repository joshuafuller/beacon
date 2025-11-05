# Beacon API Reference

Complete API documentation for the Beacon mDNS library.

---

## Quick Navigation

### Public APIs

- **[Querier API](querier.md)** - Service discovery
- **[Responder API](responder.md)** - Service announcement
- **[Common Types](types.md)** - Shared types and constants

### Packages

| Package | Purpose | Stability |
|---------|---------|-----------|
| `github.com/joshuafuller/beacon/querier` | Service discovery (queries) | ✅ Stable |
| `github.com/joshuafuller/beacon/responder` | Service announcement (responses) | ✅ Stable |

### Internal Packages

Internal packages (`internal/*`) are **not part of the public API** and may change without notice. Do not import them directly.

---

## API Stability Guarantee

Beacon follows [Semantic Versioning 2.0.0](https://semver.org/):

- **Major version (v1.x.x)**: Breaking changes to public API
- **Minor version (v0.x.0)**: New features, backward compatible
- **Patch version (v0.0.x)**: Bug fixes, backward compatible

**Current version**: v0.1.0 (pre-1.0, API may change)

**Stability commitment**:
- ✅ Public APIs (`querier`, `responder`) follow semantic versioning
- ✅ Deprecations announced one minor version before removal
- ❌ Internal APIs (`internal/*`) have no stability guarantee

---

## Common Patterns

### Creating and Closing Resources

**All Beacon resources must be closed** to avoid socket leaks:

```go
// Querier
q, err := querier.New()
if err != nil {
    return err
}
defer q.Close()  // Always close

// Responder
r, err := responder.New(ctx)
if err != nil {
    return err
}
defer r.Close()  // Always close
```

---

### Context Usage

**All blocking operations accept `context.Context`**:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
results, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    results, _ := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
}()

// Cancel the query
cancel()
```

---

### Error Handling

**Always check errors**:

```go
q, err := querier.New()
if err != nil {
    // Handle creation errors
    // Common: permission denied, address in use
    return fmt.Errorf("create querier: %w", err)
}
defer q.Close()

results, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
if err != nil {
    // Handle query errors
    // Common: context timeout, network errors
    return fmt.Errorf("query failed: %w", err)
}
```

**Error types**:
- `NetworkError` - Network I/O failures
- `ValidationError` - Invalid input (service name, port, etc.)
- `WireFormatError` - Malformed DNS packets
- `context.DeadlineExceeded` - Operation timed out
- `context.Canceled` - Operation was canceled

---

## API Overview

### Querier API

**Purpose**: Discover services on the local network

**Basic usage**:
```go
import "github.com/joshuafuller/beacon/querier"

q, _ := querier.New()
defer q.Close()

results, _ := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
```

**See**: [Querier API Documentation](querier.md)

---

### Responder API

**Purpose**: Announce services on the local network

**Basic usage**:
```go
import "github.com/joshuafuller/beacon/responder"

r, _ := responder.New(ctx)
defer r.Close()

svc := &responder.Service{
    Instance: "My Service",
    Service:  "_http._tcp",
    Domain:   "local",
    Port:     8080,
}

r.Register(ctx, svc)
```

**See**: [Responder API Documentation](responder.md)

---

## Type Reference

### QueryType

DNS query types for service discovery:

```go
const (
    RecordTypePTR querier.RecordType = 12  // Browse for service instances
    RecordTypeSRV querier.RecordType = 33  // Get service location (host:port)
    RecordTypeTXT querier.RecordType = 16  // Get service metadata
    RecordTypeA   querier.RecordType = 1   // Resolve hostname to IPv4
)
```

**See**: [Types Documentation](types.md)

---

### ResourceRecord

DNS resource record (response from query):

```go
type ResourceRecord struct {
    Name  string        // Record name (e.g., "My Server._http._tcp.local")
    Type  QueryType     // Record type (PTR, SRV, TXT, A)
    Class uint16        // DNS class (always 1 for IN)
    TTL   uint32        // Time-to-live in seconds
    Data  []byte        // Record data (format depends on Type)
}
```

**See**: [Types Documentation](types.md)

---

### Service

Service definition for announcement:

```go
type Service struct {
    Instance string     // Service instance name (e.g., "My Web Server")
    Service  string     // Service type (e.g., "_http._tcp")
    Domain   string     // Domain (always "local" for mDNS)
    Port     uint16     // Service port number (1-65535)
    TXT      []string   // TXT record strings (metadata)

    // Optional (auto-detected if not provided)
    HostName string     // Hostname (default: os.Hostname())
    IPs      []net.IP   // IP addresses (default: auto-detected)
}
```

**See**: [Responder API Documentation](responder.md)

---

## Examples

### Discover HTTP Services

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/joshuafuller/beacon/querier"
)

func main() {
    q, _ := querier.New()
    defer q.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    results, _ := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)

    for _, rr := range results {
        fmt.Printf("Found: %s\n", rr.Name)
    }
}
```

---

### Announce a Service

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "github.com/joshuafuller/beacon/responder"
)

func main() {
    r, _ := responder.New(context.Background())
    defer r.Close()

    svc := &responder.Service{
        Instance: "My API",
        Service:  "_http._tcp",
        Domain:   "local",
        Port:     8080,
        TXT:      []string{"version=1.0", "path=/api/v1"},
    }

    r.Register(context.Background(), svc)

    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
}
```

---

### Multiple Services

```go
r, _ := responder.New(context.Background())
defer r.Close()

// Register HTTP service
httpSvc := &responder.Service{
    Instance: "My Web Server",
    Service:  "_http._tcp",
    Port:     8080,
}
r.Register(ctx, httpSvc)

// Register SSH service
sshSvc := &responder.Service{
    Instance: "My SSH Server",
    Service:  "_ssh._tcp",
    Port:     22,
}
r.Register(ctx, sshSvc)

// Both services are now announced
```

---

## Thread Safety

### Querier

**Thread-safe**: Multiple goroutines can call `Query()` concurrently.

```go
q, _ := querier.New()
defer q.Close()

// Safe: concurrent queries
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
    }()
}
wg.Wait()
```

---

### Responder

**Thread-safe**: All methods can be called concurrently.

```go
r, _ := responder.New(ctx)
defer r.Close()

// Safe: concurrent registration
r.Register(ctx, service1)  // Goroutine 1
r.Register(ctx, service2)  // Goroutine 2
```

---

## Performance Considerations

### Buffer Pooling

Beacon uses `sync.Pool` internally for receive buffers. **No user action required** - this is automatic.

**Result**: 99% allocation reduction (9000 B/op → 48 B/op)

---

### Query Batching

**Not recommended**: Sending many queries in quick succession

```go
// ❌ Avoid: hammers the network
for i := 0; i < 100; i++ {
    q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
}
```

**Better**: Space out queries

```go
// ✅ Better: rate-limited queries
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for range ticker.C {
    q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
}
```

---

### Service Registration

**Efficient**: Register multiple services per responder

```go
// ✅ Good: one responder, multiple services
r, _ := responder.New(ctx)
r.Register(ctx, service1)
r.Register(ctx, service2)
r.Register(ctx, service3)
```

**Inefficient**: Multiple responders for multiple services

```go
// ❌ Avoid: creates unnecessary sockets
r1, _ := responder.New(ctx)
r1.Register(ctx, service1)

r2, _ := responder.New(ctx)  // Extra socket!
r2.Register(ctx, service2)
```

---

## Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for version history and breaking changes.

---

## Further Reading

- [Getting Started Guide](../guides/getting-started.md) - Tutorials and examples
- [Architecture Overview](../guides/architecture.md) - How Beacon works
- [Troubleshooting Guide](../guides/troubleshooting.md) - Common issues

---

**Questions?** [Open a discussion](https://github.com/joshuafuller/beacon/discussions) or see [Contributing Guide](../../CONTRIBUTING.md).
