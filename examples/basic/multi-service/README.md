# Multi-Service Registration

**Category**: Basic
**Estimated Time**: 5 minutes
**Prerequisites**: Go 1.21+

## What This Example Demonstrates

Registering multiple mDNS services with a single responder instance.

**Key Concepts**:
- Single responder, multiple services
- Different service types (_http._tcp, _ssh._tcp)
- Coordinated shutdown

## How to Run

```bash
cd beacon/examples/basic/multi-service
make run
```

## Expected Output

```
=== Multi-Service Registration Example ===

Registering services...
  ✓ Web Server._http._tcp.local (port 8080)
  ✓ API Server._http._tcp.local (port 8081)
  ✓ SSH Server._ssh._tcp.local (port 22)

✓ All 3 services registered successfully
Press Ctrl+C to exit
```

## Why This Matters

Real applications often expose multiple services:
- Web UI on port 80
- Admin API on port 8080
- Metrics endpoint on port 9090

A single responder can announce all services, simplifying lifecycle management.

## Production Pattern

```go
r, _ := responder.New(ctx)
defer r.Close() // ✓ Single Close() sends goodbye for ALL services

services := []*responder.Service{...}
for _, svc := range services {
	r.Register(svc)
}
```

## Next Steps

- [Service Browser](../browser/) - Discover services on the network
- [Graceful Shutdown](../graceful-shutdown/) - Coordinate multi-service shutdown
