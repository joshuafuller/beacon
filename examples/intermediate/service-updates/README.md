# Example 7: Dynamic Service Updates

**Difficulty**: Intermediate
**Target Audience**: Developers needing runtime metadata updates
**Estimated Time**: 15 minutes

## Overview

This example demonstrates how to update service TXT records at runtime, allowing you to publish dynamic metadata like health status, load levels, or feature flags that change over time.

**Real-World Use Case**: Ideal for services that need to publish changing state (server load, resource availability, operational status) without re-registering or requiring clients to re-query.

## What This Example Demonstrates

- **Dynamic TXT Record Updates**: Modifying service metadata at runtime
- **RFC 6762 §10.3 Compliance**: Proper reannouncement of changed records
- **Use Cases**: Health status, load balancing, feature flags, capacity tracking
- **Update Patterns**: Periodic updates vs. event-driven updates

## Code Structure

```
service-updates/
├── README.md         # This file
├── main.go           # Dynamic TXT update demonstration (~90 lines)
├── go.mod            # Go module definition
└── Makefile          # Build and run commands
```

## How It Works

### 1. Initial Service Registration

```go
// Register service with baseline TXT metadata
service := responder.Service{
    InstanceName: "Load Monitor",
    ServiceType:  "_http._tcp",
    Port:         8080,
    TXTRecords: map[string]string{
        "status":   "healthy",
        "load":     "0",
        "features": "v1,v2",
    },
}
resp.Register(&service)
```

### 2. Dynamic Updates (RFC 6762 §10.3)

When TXT records change, the responder must reannounce the updated values:

```go
// Simulate load change
currentLoad := rand.Intn(101)
service.TXTRecords["load"] = fmt.Sprintf("%d", currentLoad)

// Update service - triggers RFC 6762 §10.3 reannouncement
if err := resp.UpdateService(&service); err != nil {
    log.Printf("Update failed: %v", err)
}
```

**RFC 6762 §10.3 Behavior**:
- Responder sends multicast announcements with updated TXT record
- Clients passively listening receive new values without re-querying
- TTL clock resets, ensuring freshness

## Running the Example

### Start the Service

```bash
cd examples/intermediate/service-updates
make run
```

**Expected Output**:
```
Service registered with status=healthy, load=0
Updating TXT records every 5 seconds...

[5s]  Updated service: status=healthy, load=23
[10s] Updated service: status=healthy, load=67
[15s] Updated service: status=healthy, load=42
[20s] Updated service: status=healthy, load=89
[25s] Updated service: status=healthy, load=14
[30s] Shutting down...
```

## RFC References

- **RFC 6762 §10**: Resource Record TTL Values
- **RFC 6762 §10.3**: Announcing Changes - Multicast reannouncements
- **RFC 6763 §6**: TXT Record Construction (key=value format)
- **RFC 6763 §6.2**: TXT Record Size Limits (≤1300 bytes)

---

**Status**: Production-ready intermediate example
**Last Updated**: 2026-01-06
**Beacon Version**: v1.0+
