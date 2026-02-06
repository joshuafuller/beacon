# Graceful Shutdown

**Category**: Basic
**Estimated Time**: 5 minutes
**Prerequisites**: Go 1.21+, understanding of signal handling

## What This Example Demonstrates

Proper service lifecycle management including signal handling and RFC 6762 §10.1 goodbye packets.

**Key Concepts**:
- Signal handling (SIGINT, SIGTERM)
- Goodbye packet transmission (TTL=0)
- Clean resource cleanup
- Production-ready shutdown pattern

## Why This Matters

In production environments (Docker, Kubernetes), services receive SIGTERM before being killed. Proper shutdown ensures:
- Other devices are immediately notified the service is gone (goodbye packets)
- Resources are released cleanly (no leaked goroutines, file handles)
- No partial state or corrupted data

Without goodbye packets, other devices wait for TTL expiration (~75 minutes default) before removing the service from cache.

## How to Run

```bash
cd beacon/examples/basic/graceful-shutdown
make run

# In another terminal, verify service is visible:
# macOS: dns-sd -B _http._tcp
# Linux: avahi-browse -t _http._tcp

# Press Ctrl+C to trigger shutdown
# Service should disappear immediately from discovery
```

## Expected Output

```
=== Graceful Shutdown Example ===
This example demonstrates proper service lifecycle management.
Press Ctrl+C to trigger graceful shutdown with goodbye packets.

✓ Service registered: Graceful Service._http._tcp.local
  Service is now visible on the network
  Waiting for shutdown signal...

[Press Ctrl+C]

✓ Received signal: interrupt
  Starting graceful shutdown...
  Sending goodbye packets (TTL=0)...
  ✓ Goodbye packets sent
  ✓ Service removed from network

=== Shutdown complete ===
```

## Key Concepts

### Goodbye Packets (RFC 6762 §10.1)

When a service is shutting down, it MUST send goodbye packets:
- Same records as announcements, but with **TTL=0**
- Tells other devices to immediately remove service from cache
- Without goodbye: devices wait 75+ minutes for TTL expiration

**Beacon handles this automatically in `Close()`** - you just need to call it.

### Signal Handling

Production services must handle:
- **SIGINT** (Ctrl+C): User termination
- **SIGTERM**: Orchestrator shutdown (Docker, Kubernetes)

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan // Block until signal received
```

### Defer Pattern

Always defer `Close()` to ensure it runs even if code panics:

```go
r, err := responder.New(ctx)
if err != nil {
	return err
}
defer r.Close() // ✓ Runs even if panic occurs
```

## Production Best Practices

### 1. Always Handle Signals

```go
// ✓ GOOD: Catches both Ctrl+C and orchestrator shutdown
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// ✗ BAD: Only handles Ctrl+C, ignores SIGTERM
signal.Notify(sigChan, os.Interrupt)
```

### 2. Use Context for Coordination

```go
// ✓ GOOD: Cancel context, then close resources
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

<-sigChan
cancel()        // Stop all operations
r.Close()       // Send goodbye packets
time.Sleep(250 * time.Millisecond) // Wait for network I/O
```

### 3. Verify Goodbye Packets

Test with network sniffer:
```bash
# Terminal 1: Run service
go run main.go

# Terminal 2: Capture mDNS traffic
sudo tcpdump -i any -n port 5353

# Terminal 1: Press Ctrl+C
# Terminal 2: Look for records with TTL=0
```

## Troubleshooting

### Problem: Service still visible after shutdown
**Symptom**: `dns-sd` shows service even after Ctrl+C
**Solution**:
- Ensure `r.Close()` is called
- Verify goodbye packets sent (use tcpdump)
- Wait 250ms after Close() for network I/O

### Problem: Process killed immediately (no goodbye)
**Symptom**: Docker/K8s kills service before goodbye packets sent
**Solution**: Set `terminationGracePeriodSeconds: 5` in pod spec (Kubernetes) or add sleep after Close()

## Next Steps

- [Multi-Service](../multi-service/) - Register multiple services with coordinated shutdown
- [Error Handling](../error-handling/) - Handle errors during shutdown
