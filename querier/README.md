# querier - mDNS Service Discovery

Package `querier` provides a simple, context-aware interface for discovering services on the local network using multicast DNS (mDNS).

## Usage

```go
import "github.com/joshuafuller/beacon/querier"

// Create a new querier
q, err := querier.New()
if err != nil {
    log.Fatal(err)
}
defer q.Close()

// Query for HTTP services
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
if err != nil {
    log.Fatal(err)
}

// Process results
for _, rr := range results {
    fmt.Printf("Found: %s -> %s (TTL: %ds)\n", rr.Name, rr.Data, rr.TTL)
}
```

## Features

- **RFC 6762 §5 Compliant**: Proper multicast query transmission
- **Context-Aware**: All operations respect context cancellation and timeouts
- **Thread-Safe**: Safe for concurrent use
- **Zero Allocations**: Optimized receive path with buffer pooling

## Query Types

- `QueryTypePTR`: Service instance enumeration (e.g., `_http._tcp.local`)
- `QueryTypeSRV`: Service location (hostname and port)
- `QueryTypeTXT`: Service metadata
- `QueryTypeA`: IPv4 address resolution

## Configuration

```go
q, err := querier.New(
    querier.WithPort(5353),              // Custom port (default: 5353)
    querier.WithInterface("eth0"),       // Specific interface
    querier.WithMulticastAddress("224.0.0.251"), // Custom multicast address
)
```

## Documentation

- **GoDoc**: https://pkg.go.dev/github.com/joshuafuller/beacon/querier
- **Examples**: [../examples/discover/](../examples/discover/)
- **RFC 6762**: [Multicast DNS](https://www.rfc-editor.org/rfc/rfc6762.html)

## Performance

- Query latency: ~163 ns/op
- Buffer allocation: 48 B/op (99% reduction via pooling)
- Concurrent query support: 100+ simultaneous queries

## Limitations

- **IPv4 only**: IPv6 support planned for v0.2.0
- **Multicast only**: Unicast responses (QU bit) not yet supported
- **Linux optimized**: Full support on Linux, basic support on macOS/Windows
