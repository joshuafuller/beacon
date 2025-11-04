# responder - mDNS Service Announcement

Package `responder` provides RFC 6762 compliant service announcement with automatic conflict resolution, rate limiting, and known-answer suppression.

## Usage

```go
import "github.com/joshuafuller/beacon/responder"

// Create a new responder
r, err := responder.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer r.Close()

// Register a service
svc := &responder.Service{
    Instance: "My Web Server",
    Service:  "_http._tcp",
    Domain:   "local",
    Port:     8080,
    TXT:      []string{"path=/", "version=1.0"},
}

if err := r.Register(ctx, svc); err != nil {
    log.Fatal(err)
}

// Service is now announced on the network
// Responder handles probing, announcing, and responding to queries
```

## Features

### RFC 6762 §8 Compliance
- **§8.1 Probing**: 3 probe queries with 250ms intervals before claiming a name
- **§8.2 Conflict Resolution**: Automatic lexicographic tie-breaking and instance name renaming
- **§8.3 Announcing**: 2 unsolicited multicast announcements after successful probing

### RFC 6762 §6-7 Query Response
- **§6.2 Rate Limiting**: 1 response per second per record per interface
- **§7.1 Known-Answer Suppression**: Suppresses responses when client already has fresh data (TTL ≥50%)

### Additional Features
- **Multi-service support**: Register multiple services per responder
- **Context-aware**: All operations respect context cancellation
- **Thread-safe**: Safe for concurrent use
- **SO_REUSEPORT**: Coexists with Avahi/Bonjour system services

## Service Registration

```go
svc := &responder.Service{
    Instance: "My Printer",      // Human-readable instance name
    Service:  "_ipp._tcp",        // Service type (DNS-SD)
    Domain:   "local",            // Domain (always "local" for mDNS)
    Port:     631,                // TCP/UDP port number
    Host:     "printer",          // Hostname (optional, defaults to system hostname)
    TXT: []string{               // Service metadata (optional)
        "txtvers=1",
        "pdl=application/postscript",
    },
}
```

### TXT Record Validation

TXT records are validated per RFC 6763 §6:
- Maximum 255 bytes per record
- Key-value format: `key=value`
- Keys must be ASCII, case-insensitive
- Values may contain binary data

## Conflict Resolution

When a name conflict is detected during probing:

1. **Lexicographic Comparison**: Compare proposed names
2. **Higher Wins**: Tie-breaker selects lexicographically later name
3. **Automatic Rename**: Losers rename with suffix (e.g., "Server (2)")
4. **Max Attempts**: 10 rename attempts before giving up

## Multi-Service Support

```go
services := []*responder.Service{
    {Instance: "Web Server", Service: "_http._tcp", Port: 80},
    {Instance: "SSH Server", Service: "_ssh._tcp", Port: 22},
    {Instance: "File Server", Service: "_smb._tcp", Port: 445},
}

for _, svc := range services {
    if err := r.Register(ctx, svc); err != nil {
        log.Printf("Failed to register %s: %v", svc.Instance, err)
    }
}
```

## Documentation

- **GoDoc**: https://pkg.go.dev/github.com/joshuafuller/beacon/responder
- **Examples**: See [Quick Start](../README.md#quick-start) in main README
- **RFC 6762**: [Multicast DNS](https://www.rfc-editor.org/rfc/rfc6762.html)
- **RFC 6763**: [DNS-SD](https://www.rfc-editor.org/rfc/rfc6763.html)

## Performance

- Response latency: **4.8μs** (20,833x under 100ms requirement)
- Conflict detection: **35ns** (zero allocations)
- Memory footprint: **750 bytes per service**
- Throughput: **602,595 ops/sec** (response builder)

## Limitations

- **IPv4 only**: IPv6 support planned for v0.2.0
- **Multicast only**: Unicast responses not yet supported
- **No goodbye packets**: TTL=0 announcements planned for v0.2.0
- **Linux optimized**: Full support on Linux, basic support on macOS/Windows

## Security

- **Rate limiting**: RFC 6762 §6.2 compliant (prevents amplification attacks)
- **Input validation**: All DNS parsing uses WireFormatError for malformed packets
- **Source filtering**: Link-local source IP validation (Linux only)
- **Fuzz tested**: 109,471 fuzz executions, 0 crashes
