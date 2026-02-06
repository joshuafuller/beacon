# Service Browser

**Category**: Basic
**Estimated Time**: 5 minutes
**Prerequisites**: Go 1.21+

## What This Example Demonstrates

Discovering mDNS services on the local network using the querier API.

**Key Concepts**:
- Service discovery with Query()
- Querier lifecycle
- Processing query results

## How to Run

```bash
cd beacon/examples/basic/browser
make run
```

## Expected Output

```
=== Service Browser Example ===
Discovering mDNS services on the local network...

Searching for HTTP services (_http._tcp.local)...
  Found 2 HTTP service(s):
    1. Web Server._http._tcp.local (TTL: 4500s)
    2. API Server._http._tcp.local (TTL: 4500s)

Searching for SSH services (_ssh._tcp.local)...
  Found 1 SSH service(s):
    1. SSH Server._ssh._tcp.local (TTL: 4500s)

=== Discovery complete ===
```

## Why This Matters

Service discovery enables:
- Zero-configuration networking (no hardcoded IPs)
- Dynamic service location (find printers, IoT devices, APIs)
- Network inventory (what services exist?)

## Key Concepts

### Querier API

```go
q, _ := querier.New(ctx)
defer q.Close()

results, err := q.Query(ctx, "_http._tcp.local")
for _, r := range results {
	fmt.Println(r.Name)
}
```

### Context Timeout

Always use timeout to prevent hanging:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### Service Types

Common service types:
- `_http._tcp.local` - HTTP servers
- `_ssh._tcp.local` - SSH servers
- `_printer._tcp.local` - Printers
- `_airplay._tcp.local` - AirPlay devices

## Next Steps

- [Hello Responder](../hello-responder/) - Announce your own services
- [Multi-Service](../multi-service/) - Register multiple services simultaneously
