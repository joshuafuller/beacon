# Migrating from grandcat/zeroconf to Beacon

This guide helps users migrate from the [grandcat/zeroconf](https://github.com/grandcat/zeroconf) library to Beacon.

## Why Migrate to Beacon?

### Advantages of Beacon

- **Production-Ready**: Comprehensive test coverage (>85%), zero flaky tests, race detector clean
- **RFC Compliance**: Full RFC 6762/6763 compliance with documented validation
- **Better Performance**: Buffer pooling reduces allocations by 99% (9000 B/op → 48 B/op)
- **Context Support**: All operations accept `context.Context` for proper cancellation
- **Clean Architecture**: Strict layer boundaries, typed errors, testable design
- **Multi-Interface Support**: Proper interface-specific IP addressing (RFC 6762 §15)
- **Active Maintenance**: Regular updates, responsive issue tracking, clear roadmap

### When to Stay with grandcat/zeroconf

- **Legacy Codebase**: If you need Go 1.16 or earlier
- **Established Codebase**: If zeroconf is working well and you don't need new features
- **Smaller Footprint**: zeroconf has fewer dependencies

---

## API Comparison Table

| **Feature** | **grandcat/zeroconf** | **Beacon** | **Notes** |
|-------------|----------------------|-----------|-----------|
| **Browsing** | `zeroconf.Browse(ctx, service, domain, entries)` | `q.Query(ctx, name, querier.TypePTR)` | Beacon uses explicit PTR queries |
| **Lookup** | `zeroconf.Lookup(ctx, instance, service, domain, entries)` | `q.Query(ctx, fqdn, querier.TypeANY)` | Beacon uses general query API |
| **Service Registration** | `zeroconf.Register(name, service, domain, port, text, ifaces)` | `responder.New(ctx)` + `RegisterService(ctx, svc)` | Beacon separates responder creation and registration |
| **Service Type** | `"_http._tcp"` | `"_http._tcp.local"` | Beacon requires `.local` suffix |
| **TXT Records** | `[]string{"key=value"}` | `map[string]string{"key": "value"}` | Beacon uses key-value map |
| **Results** | `<-chan *ServiceEntry` | `[]querier.ResourceRecord` | Beacon returns slices, not channels |
| **Context** | ✅ Supported | ✅ Required | Both use context, but Beacon requires it everywhere |
| **Error Handling** | Generic errors | Typed errors | Beacon: `NetworkError`, `ValidationError`, etc. |
| **IPv6** | Partial | Full RFC 6762 | Beacon supports dual-stack |
| **Shutdown** | `server.Shutdown()` | `resp.Close()` or `resp.DeregisterService()` | Beacon supports granular deregistration |

---

## Migration Examples

### Example 1: Service Discovery (Browse)

#### grandcat/zeroconf
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/grandcat/zeroconf"
)

func main() {
    resolver, err := zeroconf.NewResolver(nil)
    if err != nil {
        panic(err)
    }

    entries := make(chan *zeroconf.ServiceEntry)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Browse for HTTP services
    err = resolver.Browse(ctx, "_http._tcp", "local", entries)
    if err != nil {
        panic(err)
    }

    // Collect results
    for entry := range entries {
        fmt.Printf("Found: %s at %s:%d\n",
            entry.Instance, entry.AddrIPv4[0], entry.Port)
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

    // Browse for HTTP services (PTR query)
    records, err := q.Query(ctx, "_http._tcp.local", querier.TypePTR)
    if err != nil {
        panic(err)
    }

    // Process results
    for _, record := range records {
        fmt.Printf("Found: %s\n", record.Name)
        // Additional queries needed for SRV/A records to get IP:port
    }
}
```

**Key Differences**:
- **Setup**: Beacon requires `querier.New(ctx)` instead of `NewResolver(nil)`
- **Service Type**: Beacon requires `.local` suffix
- **Results**: Slice instead of channel
- **Details**: PTR records return service names; additional queries needed for IP/port

---

### Example 2: Service Lookup (Resolve Instance)

#### grandcat/zeroconf
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/grandcat/zeroconf"
)

func main() {
    resolver, err := zeroconf.NewResolver(nil)
    if err != nil {
        panic(err)
    }

    entries := make(chan *zeroconf.ServiceEntry)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Lookup specific instance
    err = resolver.Lookup(ctx, "My Web Server", "_http._tcp", "local", entries)
    if err != nil {
        panic(err)
    }

    // Get first result
    entry := <-entries
    fmt.Printf("Instance: %s\n", entry.Instance)
    fmt.Printf("Address: %s:%d\n", entry.AddrIPv4[0], entry.Port)
    fmt.Printf("TXT: %v\n", entry.Text)
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

    q, err := querier.New(ctx)
    if err != nil {
        panic(err)
    }
    defer q.Close()

    // Lookup specific instance (ANY query for all record types)
    fqdn := "My Web Server._http._tcp.local"
    records, err := q.Query(ctx, fqdn, querier.TypeANY)
    if err != nil {
        panic(err)
    }

    // Process different record types
    for _, record := range records {
        switch record.Type {
        case querier.TypeSRV:
            fmt.Printf("Port: %d\n", record.Port)
        case querier.TypeA:
            fmt.Printf("Address: %s\n", record.IP)
        case querier.TypeTXT:
            fmt.Printf("TXT: %s\n", record.TXT)
        }
    }
}
```

**Key Differences**:
- **FQDN**: Beacon requires fully-qualified domain name
- **Record Types**: Beacon separates SRV, A, TXT records
- **Iteration**: Process records by type instead of unified entry

---

### Example 3: Service Registration

#### grandcat/zeroconf
```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/grandcat/zeroconf"
)

func main() {
    server, err := zeroconf.Register(
        "My Web Server",     // instance name
        "_http._tcp",        // service type
        "local",            // domain
        8080,               // port
        []string{"version=1.0", "status=ready"}, // TXT records
        nil,                // network interfaces (all)
    )
    if err != nil {
        panic(err)
    }
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
}
```

**Key Differences**:
- **Two-Step**: Beacon separates responder creation and service registration
- **Service Type**: Beacon requires `.local` suffix
- **TXT Records**: Map instead of `[]string`
- **Deregistration**: `Close()` handles cleanup, or use `DeregisterService()` for granular control

---

### Example 4: Dynamic Service Updates

#### grandcat/zeroconf
```go
package main

import (
    "fmt"
    "time"

    "github.com/grandcat/zeroconf"
)

func main() {
    // zeroconf requires shutdown and re-registration for updates
    for i := 0; i < 10; i++ {
        txt := []string{fmt.Sprintf("count=%d", i)}

        server, err := zeroconf.Register(
            "Counter Service",
            "_http._tcp",
            "local",
            8080,
            txt,
            nil,
        )
        if err != nil {
            panic(err)
        }

        time.Sleep(5 * time.Second)
        server.Shutdown()
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
        resp.RegisterService(ctx, svc) // Re-register to update
    }
}
```

**Key Differences**:
- **No Restart**: Beacon supports in-place updates
- **Cleaner API**: No shutdown/re-registration needed
- **Efficiency**: Service stays online during updates

---

## Migration Checklist

### Pre-Migration

- [ ] Review Beacon's [README](../../README.md) and [Getting Started](../README.md)
- [ ] Check Go version (Beacon requires 1.21+)
- [ ] Identify channel-based result handling (needs conversion to slices)
- [ ] Review TXT record formats (`[]string` → `map[string]string`)
- [ ] Note `.local` suffix requirements

### During Migration

- [ ] Replace `zeroconf.NewResolver()` with `querier.New(ctx)`
- [ ] Replace `resolver.Browse()` with `q.Query(ctx, name, querier.TypePTR)`
- [ ] Replace `resolver.Lookup()` with `q.Query(ctx, fqdn, querier.TypeANY)`
- [ ] Add `.local` suffix to all service types
- [ ] Convert TXT records from `[]string` to `map[string]string`
- [ ] Replace `zeroconf.Register()` with `responder.New(ctx)` + `RegisterService()`
- [ ] Replace channel result handling with slice iteration
- [ ] Add `defer Close()` calls
- [ ] Update error handling to use typed errors

### Post-Migration

- [ ] Test querier with `dns-sd -B` or `avahi-browse`
- [ ] Test responder visibility with other mDNS clients
- [ ] Run race detector: `go test -race ./...`
- [ ] Verify graceful shutdown
- [ ] Check for validation errors in logs
- [ ] Performance test (if replacing high-throughput code)

---

## Common Gotchas

### 1. `.local` Suffix Required

**grandcat/zeroconf**:
```go
resolver.Browse(ctx, "_http._tcp", "local", entries)
```

**Beacon**:
```go
q.Query(ctx, "_http._tcp.local", querier.TypePTR) // .local required
```

---

### 2. Channel vs Slice Results

**grandcat/zeroconf**:
```go
entries := make(chan *zeroconf.ServiceEntry)
resolver.Browse(ctx, "_http._tcp", "local", entries)
for entry := range entries {
    fmt.Println(entry.Instance)
}
```

**Beacon**:
```go
records, err := q.Query(ctx, "_http._tcp.local", querier.TypePTR)
for _, record := range records {
    fmt.Println(record.Name)
}
```

---

### 3. TXT Record Format

**grandcat/zeroconf**:
```go
[]string{"key1=value1", "key2=value2"}
```

**Beacon**:
```go
map[string]string{"key1": "value1", "key2": "value2"}
```

---

### 4. Two-Step Service Registration

**grandcat/zeroconf**:
```go
server, _ := zeroconf.Register("name", "_http._tcp", "local", 8080, txt, nil)
```

**Beacon**:
```go
resp, _ := responder.New(ctx)
svc := &responder.Service{...}
resp.RegisterService(ctx, svc)
```

---

### 5. Separate Record Types

**grandcat/zeroconf** returns unified `ServiceEntry`:
```go
entry := <-entries
fmt.Println(entry.AddrIPv4[0], entry.Port, entry.Text)
```

**Beacon** returns separate records:
```go
for _, record := range records {
    switch record.Type {
    case querier.TypeA:
        fmt.Println(record.IP)
    case querier.TypeSRV:
        fmt.Println(record.Port)
    case querier.TypeTXT:
        fmt.Println(record.TXT)
    }
}
```

---

## Performance Considerations

### grandcat/zeroconf
- Good baseline performance
- Allocates per-query buffers
- Channel-based concurrency

### Beacon
- **99% allocation reduction** via buffer pooling
- **100+ concurrent queries** supported
- **RFC 6762 §6.2 rate limiting** (per-interface)

**Recommendation**: Benchmark high-throughput code:
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

- [Beacon Architecture](../api/README.md) - Internal design
- [RFC Compliance Guide](../RFC_COMPLIANCE_GUIDE.md) - mDNS implementation details
- [Migration from hashicorp/mdns](./from-hashicorp-mdns.md) - Another popular library

---

**Last Updated**: 2026-01-06
**Beacon Version**: v0.1.0
**grandcat/zeroconf Version**: v1.0.0 (as of 2024-01)
