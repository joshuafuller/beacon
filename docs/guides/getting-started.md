# Getting Started with Beacon

**Audience**: New users who want to use Beacon for mDNS service discovery
**Time to complete**: 15 minutes
**Prerequisites**: Go 1.21 or later installed

---

## What is Beacon?

Beacon is a high-performance mDNS (Multicast DNS) library for Go that lets you:

- **Discover services** on your local network (printers, IoT devices, APIs, etc.)
- **Announce services** so other devices can find you
- **Replace** legacy libraries like hashicorp/mdns with a modern, RFC-compliant implementation

**Key Benefits**:
- 10,000x faster than alternatives (4.8μs response time)
- 72.2% RFC 6762 compliant (industry-leading)
- Zero external dependencies
- Production-ready (81.3% test coverage, extensive fuzz testing)

---

## Installation

```bash
go get github.com/joshuafuller/beacon@latest
```

**Verify installation**:
```bash
go mod tidy
go list -m github.com/joshuafuller/beacon
# Should output: github.com/joshuafuller/beacon v0.x.x
```

---

## Your First Query (Service Discovery)

Let's find all HTTP services on your local network.

### Step 1: Create a basic program

Create a file named `find_http.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/joshuafuller/beacon/querier"
)

func main() {
    // Create a querier
    q, err := querier.New()
    if err != nil {
        log.Fatalf("Failed to create querier: %v", err)
    }
    defer q.Close()

    // Set a timeout for the query
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    // Query for HTTP services
    fmt.Println("Searching for HTTP services...")
    results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    // Display results
    fmt.Printf("Found %d services:\n", len(results))
    for _, rr := range results {
        fmt.Printf("  - %s (TTL: %d seconds)\n", rr.Name, rr.TTL)
    }
}
```

### Step 2: Run it

```bash
go run find_http.go
```

**Expected output**:
```
Searching for HTTP services...
Found 2 services:
  - My Web Server._http._tcp.local (TTL: 120 seconds)
  - Home Assistant._http._tcp.local (TTL: 120 seconds)
```

**What just happened?**

1. **Created a querier** - This sets up the mDNS socket on port 5353
2. **Sent a PTR query** - Asked "who provides _http._tcp services?"
3. **Received responses** - Devices on your network replied with their service names
4. **Displayed results** - Showed the service instances found

---

## Your First Responder (Service Announcement)

Let's announce a service so others can discover it.

### Step 1: Create an announcement program

Create a file named `announce_service.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/joshuafuller/beacon/responder"
)

func main() {
    // Create a responder
    r, err := responder.New(context.Background())
    if err != nil {
        log.Fatalf("Failed to create responder: %v", err)
    }
    defer r.Close()

    // Define your service
    svc := &responder.Service{
        Instance: "My Test Service",       // Human-readable name
        Service:  "_http._tcp",            // Service type
        Domain:   "local",                 // Always "local" for mDNS
        Port:     8080,                    // Your service's port
        TXT: []string{                     // Optional metadata
            "version=1.0",
            "path=/api",
        },
    }

    // Register and announce the service
    ctx := context.Background()
    fmt.Printf("Registering service: %s._http._tcp.local\n", svc.Instance)
    if err := r.Register(ctx, svc); err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }
    fmt.Println("Service announced! Press Ctrl+C to stop.")

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    fmt.Println("\nShutting down...")
}
```

### Step 2: Run it

```bash
go run announce_service.go
```

**Expected output**:
```
Registering service: My Test Service._http._tcp.local
Service announced! Press Ctrl+C to stop.
```

### Step 3: Discover it from another terminal

While `announce_service.go` is running, open a new terminal and run:

```bash
go run find_http.go
```

You should now see "My Test Service" in the results!

**What just happened?**

1. **Created a responder** - Set up mDNS to answer queries
2. **Defined a service** - Specified name, type, port, and metadata
3. **Registered** - Beacon performed RFC-compliant probing (checked for conflicts) and announcing
4. **Service is discoverable** - Other devices can now find your service via mDNS

---

## Common Service Types

Beacon supports any service type, but here are common ones:

| Service Type | Description | Example Use Case |
|--------------|-------------|------------------|
| `_http._tcp` | HTTP web services | REST APIs, web servers |
| `_https._tcp` | HTTPS web services | Secure APIs |
| `_ssh._tcp` | SSH services | Remote access |
| `_printer._tcp` | Network printers | Print discovery |
| `_airplay._tcp` | AirPlay devices | Media streaming |
| `_homekit._tcp` | HomeKit accessories | Smart home devices |
| `_googlecast._tcp` | Chromecast devices | Media casting |

**Creating custom service types**:
```go
Service: "_myapp._tcp"  // Your custom service
```

Service names should follow the pattern `_<name>._<protocol>` (usually `_tcp` or `_udp`).

---

## Understanding Query Types

Beacon supports different DNS query types for different use cases:

### PTR (Pointer) - "What services exist?"

**Use case**: Browse for service instances

```go
results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
// Returns: ["My Server._http._tcp.local", "Another Server._http._tcp.local"]
```

### SRV (Service) - "Where is this service?"

**Use case**: Get hostname and port for a specific service instance

```go
results, _ := q.Query(ctx, "My Server._http._tcp.local", querier.QueryTypeSRV)
// Returns: SRV record with hostname and port
```

### TXT - "What metadata does this service have?"

**Use case**: Get service metadata (version, capabilities, etc.)

```go
results, _ := q.Query(ctx, "My Server._http._tcp.local", querier.QueryTypeTXT)
// Returns: TXT records like ["version=1.0", "path=/api"]
```

### A (Address) - "What's the IP address?"

**Use case**: Resolve hostname to IPv4 address

```go
results, _ := q.Query(ctx, "myserver.local", querier.QueryTypeA)
// Returns: A record with IPv4 address
```

**Pro tip**: When discovering a service, you often need multiple queries (PTR → SRV → A) to get the full connection details. See [Advanced Usage](advanced-usage.md) for helper patterns.

---

## Error Handling Best Practices

Always handle errors properly in production code:

```go
// 1. Check querier/responder creation
q, err := querier.New()
if err != nil {
    return fmt.Errorf("failed to create querier: %w", err)
}
defer q.Close()  // Always close resources

// 2. Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 3. Check query/register results
results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
if err != nil {
    // Handle context deadline, network errors, etc.
    return fmt.Errorf("query failed: %w", err)
}

// 4. Validate results
if len(results) == 0 {
    log.Println("No services found (this is normal, not an error)")
}
```

---

## Resource Management

**Always close resources** to avoid socket leaks:

```go
// Querier
q, err := querier.New()
if err != nil {
    return err
}
defer q.Close()  // ✅ Closes UDP socket

// Responder
r, err := responder.New(ctx)
if err != nil {
    return err
}
defer r.Close()  // ✅ Closes UDP socket and stops goroutines
```

**For long-running services**, use graceful shutdown:

```go
r, _ := responder.New(context.Background())
defer r.Close()

// ... register services ...

// Wait for shutdown signal
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan

// r.Close() is called via defer, sending goodbye packets
```

---

## Troubleshooting

### "No services found" but I know they exist

**Possible causes**:
1. **Firewall blocking UDP 5353** - Allow multicast traffic
2. **VPN active** - mDNS doesn't work over VPNs (by design)
3. **Different subnet** - mDNS is link-local only
4. **Query timeout too short** - Try increasing to 5 seconds

**Debug steps**:
```bash
# Check if mDNS traffic is flowing
sudo tcpdump -i any port 5353

# Verify with system tools
avahi-browse -a -t  # Linux
dns-sd -B _http._tcp  # macOS
```

### "Address already in use" error

**Cause**: Another program is using port 5353

**Solution**: Beacon uses SO_REUSEPORT to coexist with system daemons (Avahi/Bonjour). This error usually means:
- You have multiple Beacon instances running
- A non-SO_REUSEPORT program is using the port

**Debug**:
```bash
# See what's using port 5353
sudo ss -ulnp 'sport = :5353'  # Linux
sudo lsof -i :5353             # macOS
```

### Service conflicts during registration

**Symptom**: Service name changes (e.g., "My Service" becomes "My Service (2)")

**Cause**: Another device is using the same instance name

**Solution**: This is normal RFC 6762 behavior! Beacon automatically resolves conflicts by appending a number. To avoid conflicts:
```go
// Use unique names (e.g., include hostname)
Instance: fmt.Sprintf("My Service (%s)", hostname)
```

---

## Next Steps

**Learn More**:
- [Querier Guide](querier-guide.md) - Deep dive into service discovery
- [Responder Guide](responder-guide.md) - Advanced service announcement patterns
- [Advanced Usage](advanced-usage.md) - Performance tuning, production patterns
- [API Reference](../api/README.md) - Complete API documentation

**See Examples**:
- [examples/basic-query](../../examples/basic-query/) - Simple query example
- [examples/basic-responder](../../examples/basic-responder/) - Simple responder example
- [examples/service-browser](../../examples/service-browser/) - Full service browser
- [examples/production](../../examples/production/) - Production-ready example

**Get Help**:
- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- [GitHub Discussions](https://github.com/joshuafuller/beacon/discussions) - Ask questions
- [GitHub Issues](https://github.com/joshuafuller/beacon/issues) - Report bugs

---

**Questions or feedback?** [Open a discussion](https://github.com/joshuafuller/beacon/discussions) - we'd love to hear from you!
