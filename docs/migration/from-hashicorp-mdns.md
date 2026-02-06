# Migrating from hashicorp/mdns to Beacon

This guide helps users migrate from the [hashicorp/mdns](https://github.com/hashicorp/mdns) library to Beacon.

## Why Migrate to Beacon?

### Advantages of Beacon

- **Modern Go Design**: Built with Go 1.21+ features, clean architecture, and strict layer boundaries
- **Context Support**: All operations accept `context.Context` for proper cancellation and timeouts
- **Production-Ready**: Comprehensive test coverage (>85%), zero flaky tests, race detector clean
- **RFC Compliance**: Full RFC 6762/6763 compliance with documented validation
- **Performance**: Buffer pooling reduces allocations by 99% (9000 B/op → 48 B/op)
- **Better Error Handling**: Typed errors (`NetworkError`, `ValidationError`, `WireFormatError`)
- **Multi-Interface Support**: Proper interface-specific IP addressing (RFC 6762 §15)
- **Active Maintenance**: Regular updates, responsive issue tracking, clear roadmap

### When to Stay with hashicorp/mdns

- **Legacy Codebase**: If your project requires Go 1.16 or earlier
- **Minimal Dependencies**: hashicorp/mdns has fewer external dependencies
- **Battle-Tested**: hashicorp/mdns has been in production at HashiCorp for many years

---

## API Comparison Table

| **Feature** | **hashicorp/mdns** | **Beacon** | **Notes** |
|-------------|-------------------|-----------|-----------|
| **Querier Creation** | `mdns.NewClient()` | `querier.New(ctx)` | Beacon requires context |
| **Query Execution** | `client.Query(params)` | `q.Query(ctx, name, qtype)` | Beacon uses explicit parameters |
| **Service Registration** | `mdns.NewMDNSService()` | `responder.Service{}` | Beacon uses struct initialization |
| **Responder Creation** | `mdns.NewServer()` | `responder.New(ctx)` | Beacon requires context |
| **Service Registration** | `server.Register(service)` | `r.RegisterService(ctx, svc)` | Context-aware in Beacon |
| **Service Deregistration** | `server.Shutdown()` | `r.DeregisterService(ctx, name, svcType)` | Beacon supports granular deregistration |
| **Browsing** | `client.Query(&QueryParam{Service: "_http._tcp"})` | `q.Query(ctx, "_http._tcp.local", querier.TypePTR)` | Beacon uses explicit PTR queries |
| **TXT Records** | `Info []string` | `TXTRecords map[string]string` | Beacon uses key-value map |
| **Resource Records** | `entries <-chan *ServiceEntry` | `records []querier.ResourceRecord` | Beacon uses slices, not channels |
| **Context Support** | ❌ None | ✅ All APIs | Beacon built with context from ground up |
| **Error Types** | Generic errors | Typed errors | Beacon: `NetworkError`, `ValidationError`, etc. |
| **IPv6 Support** | Partial | Full (RFC 6762) | Beacon supports dual-stack |
| **Buffer Pooling** | ❌ None | ✅ Built-in | 99% allocation reduction |
| **Rate Limiting** | ❌ None | ✅ RFC 6762 §6.2 | Per-interface rate limiting |
| **Conflict Detection** | Basic | RFC 6762 §8.2 | Lexicographic tie-breaking |

---

## Migration Examples

### Example 1: Basic Service Query

#### hashicorp/mdns
```go
package main

import (
    "fmt"
    "time"

    "github.com/hashicorp/mdns"
)

func main() {
    // Create a channel for results
    entriesCh := make(chan *mdns.ServiceEntry, 10)

    // Query for HTTP services
    params := &mdns.QueryParam{
        Service: "_http._tcp",
        Timeout: 5 * time.Second,
        Entries: entriesCh,
    }

    // Execute query
    go mdns.Query(params)

    // Collect results
    timeout := time.After(5 * time.Second)
    for {
        select {
        case entry := <-entriesCh:
            fmt.Printf("Found service: %s at %s:%d\n",
                entry.Name, entry.AddrV4, entry.Port)
        case <-timeout:
            return
        }
    }
}
```

#### Beacon
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/joshuafuller/beacon/querier"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Create querier
    q, err := querier.New(ctx)
    if err != nil {
        panic(err)
    }
    defer q.Close()

    // Query for HTTP services (PTR query)
    records, err := q.Query(ctx, "_http._tcp.local", querier.TypePTR)
    if err != nil {
        panic(err)
    }

    // Process results
    for _, record := range records {
        fmt.Printf("Found service: %s\n", record.Name)
        // To get full details (SRV, TXT), query the PTR target
    }
}
```

**Key Differences**:
- **Context**: Beacon uses `context.Context` for cancellation
- **Results**: Beacon returns `[]ResourceRecord` instead of channel
- **Lifecycle**: Explicit `Close()` call required
- **Error Handling**: Beacon returns typed errors

---

### Example 2: Service Registration

#### hashicorp/mdns
```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "github.com/hashicorp/mdns"
)

func main() {
    // Create service
    service, _ := mdns.NewMDNSService(
        "My Web Server",
        "_http._tcp",
        "",
        "",
        8080,
        nil, // IPs (auto-detected)
        []string{"version=1.0", "status=ready"},
    )

    // Create responder
    server, _ := mdns.NewServer(&mdns.Config{Zone: service})
    defer server.Shutdown()

    // Wait for signal
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
    <-sig
}
```

#### Beacon
```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/joshuafuller/beacon/responder"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create responder
    resp, err := responder.New(ctx)
    if err != nil {
        panic(err)
    }
    defer resp.Close()

    // Define service
    svc := &responder.Service{
        InstanceName: "My Web Server",
        ServiceType:  "_http._tcp.local",
        Port:         8080,
        TXTRecords: map[string]string{
            "version": "1.0",
            "status":  "ready",
        },
    }

    // Register service
    if err := resp.RegisterService(ctx, svc); err != nil {
        panic(err)
    }

    // Wait for signal
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
    <-sig

    // Deregister service (optional, Close() handles this)
    resp.DeregisterService(ctx, svc.InstanceName, svc.ServiceType)
}
```

**Key Differences**:
- **Service Definition**: Beacon uses `responder.Service` struct
- **TXT Records**: Map instead of `[]string`
- **Registration**: Explicit `RegisterService()` call
- **Deregistration**: Granular deregistration support
- **Context**: All operations accept context

---

### Example 3: Dynamic Service Updates

#### hashicorp/mdns
```go
// hashicorp/mdns does not support dynamic updates
// You must shutdown and recreate the server

package main

import (
    "time"

    "github.com/hashicorp/mdns"
)

func main() {
    for i := 0; i < 10; i++ {
        // Create service with updated TXT records
        service, _ := mdns.NewMDNSService(
            "Counter Service",
            "_http._tcp",
            "",
            "",
            8080,
            nil,
            []string{fmt.Sprintf("count=%d", i)},
        )

        // Shutdown old server
        if server != nil {
            server.Shutdown()
        }

        // Create new server
        server, _ = mdns.NewServer(&mdns.Config{Zone: service})

        time.Sleep(5 * time.Second)
    }
}
```

#### Beacon
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/joshuafuller/beacon/responder"
)

func main() {
    ctx := context.Background()

    resp, err := responder.New(ctx)
    if err != nil {
        panic(err)
    }
    defer resp.Close()

    svc := &responder.Service{
        InstanceName: "Counter Service",
        ServiceType:  "_http._tcp.local",
        Port:         8080,
        TXTRecords:   map[string]string{"count": "0"},
    }

    resp.RegisterService(ctx, svc)

    // Update TXT records dynamically
    for i := 1; i < 10; i++ {
        time.Sleep(5 * time.Second)

        svc.TXTRecords["count"] = fmt.Sprintf("%d", i)

        // Re-register to update (triggers re-announcement)
        resp.RegisterService(ctx, svc)
    }
}
```

**Key Differences**:
- **Dynamic Updates**: Beacon supports in-place updates via re-registration
- **No Restart Required**: Service stays online during updates
- **Cleaner API**: No server shutdown/recreation needed

---

## Migration Checklist

### Pre-Migration

- [ ] Review Beacon's [README](../../README.md) and [Getting Started](../README.md)
- [ ] Check Go version (Beacon requires 1.21+)
- [ ] Review your hashicorp/mdns usage patterns (query vs. responder)
- [ ] Identify TXT record formats that need conversion (`[]string` → `map[string]string`)

### During Migration

- [ ] Replace `mdns.NewClient()` with `querier.New(ctx)`
- [ ] Replace `mdns.Query(params)` with `q.Query(ctx, name, qtype)`
- [ ] Convert TXT records from `[]string` to `map[string]string`
- [ ] Replace `mdns.NewServer()` with `responder.New(ctx)`
- [ ] Replace `server.Register()` with `resp.RegisterService(ctx, svc)`
- [ ] Add `defer Close()` calls for querier and responder
- [ ] Add context handling (`context.WithTimeout`, `context.WithCancel`)
- [ ] Update error handling to use typed errors

### Post-Migration

- [ ] Test querier with `dns-sd -B` or `avahi-browse`
- [ ] Test responder visibility with other mDNS clients
- [ ] Run race detector: `go test -race ./...`
- [ ] Verify graceful shutdown (context cancellation)
- [ ] Check logs for validation errors
- [ ] Performance test (if replacing high-throughput code)

---

## Common Gotchas

### 1. Context Required Everywhere

**hashicorp/mdns**:
```go
client.Query(params) // No context
```

**Beacon**:
```go
q.Query(ctx, name, qtype) // Context required
```

**Solution**: Create contexts at the start of your functions:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

---

### 2. TXT Record Format

**hashicorp/mdns**:
```go
[]string{"key1=value1", "key2=value2"}
```

**Beacon**:
```go
map[string]string{"key1": "value1", "key2": "value2"}
```

**Solution**: Convert during migration:
```go
// Old
txtRecords := []string{"version=1.0", "status=ready"}

// New
txtRecords := map[string]string{
    "version": "1.0",
    "status":  "ready",
}
```

---

### 3. Query Results (Channel vs Slice)

**hashicorp/mdns**:
```go
entriesCh := make(chan *mdns.ServiceEntry, 10)
go mdns.Query(&mdns.QueryParam{Entries: entriesCh})
for entry := range entriesCh {
    // Process entry
}
```

**Beacon**:
```go
records, err := q.Query(ctx, "_http._tcp.local", querier.TypePTR)
for _, record := range records {
    // Process record
}
```

**Solution**: Replace channel loops with slice iteration.

---

### 4. Service Deregistration

**hashicorp/mdns**:
```go
server.Shutdown() // Shuts down ALL services
```

**Beacon**:
```go
resp.DeregisterService(ctx, instanceName, serviceType) // Granular
// OR
resp.Close() // Deregisters all services
```

**Solution**: Use granular deregistration for multi-service responders.

---

## Performance Considerations

### hashicorp/mdns
- Allocates ~9KB per query receive
- No buffer pooling
- Limited concurrency controls

### Beacon
- Buffer pooling reduces allocations by 99%
- Concurrent query support (100+ queries)
- RFC 6762 §6.2 rate limiting (per-interface)

**Recommendation**: If migrating high-throughput code, benchmark before and after:
```bash
go test -bench=. -benchmem ./...
```

---

## Getting Help

- **Documentation**: [Beacon Docs](../README.md)
- **Examples**: [Basic Examples](../../examples/basic/), [Intermediate Examples](../../examples/intermediate/)
- **Issues**: [GitHub Issues](https://github.com/joshuafuller/beacon/issues)
- **RFC References**: [RFC 6762 (mDNS)](https://www.rfc-editor.org/rfc/rfc6762.html), [RFC 6763 (DNS-SD)](https://www.rfc-editor.org/rfc/rfc6763.html)

---

## Further Reading

- [Beacon Architecture](../api/README.md) - Internal design and layer boundaries
- [RFC Compliance Guide](../RFC_COMPLIANCE_GUIDE.md) - How Beacon implements mDNS spec
- [Testing Guide](../CI_PIPELINE.md) - Test coverage and quality gates
- [Migration from grandcat/zeroconf](./from-grandcat.md) - Another popular mDNS library

---

**Last Updated**: 2026-01-06
**Beacon Version**: v0.1.0
**hashicorp/mdns Version**: v1.0.5 (as of 2024-01)
