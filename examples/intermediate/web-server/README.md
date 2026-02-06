# Example 6: Web Server with mDNS

**Difficulty**: Intermediate
**Target Audience**: Web developers adding service discovery
**Estimated Time**: 15 minutes

## Overview

This example demonstrates integrating an mDNS responder with a standard HTTP server, making your web application discoverable on the local network without manual configuration. Users can discover and access your web service using tools like Safari's Bonjour browser or `dns-sd`.

**Real-World Use Case**: Ideal for local development servers, embedded web interfaces (IoT devices), and internal tools where users need automatic discovery without DNS configuration.

## What This Example Demonstrates

- **Dual Functionality**: Running HTTP server + mDNS responder in a single process
- **TXT Record Metadata**: Publishing service metadata (path, version) for clients
- **Goroutine Coordination**: Managing multiple subsystems concurrently
- **Graceful Shutdown**: Cleanly stopping both HTTP and mDNS on interrupt

## Code Structure

```
web-server/
├── README.md         # This file
├── main.go           # HTTP server + mDNS integration (~100 lines)
├── go.mod            # Go module definition
└── Makefile          # Build and run commands
```

## How It Works

### 1. HTTP Server Setup
```go
// Standard HTTP server listening on port 8080
server := &http.Server{
    Addr:    ":8080",
    Handler: http.HandlerFunc(handleRequest),
}
```

### 2. mDNS Service Registration
```go
// Register HTTP service with TXT metadata
service := responder.ServiceDefinition{
    ServiceInstanceName: "Web Demo",
    ServiceType:         "_http._tcp",
    Port:                8080,
    TXTRecords:          []string{"path=/", "version=1.0"},
}
```

### 3. TXT Record Usage (RFC 6763 §6)

TXT records provide additional metadata to clients:
- `path=/` - HTTP endpoint path (RFC 6763 §4.1.3)
- `version=1.0` - API version for compatibility checking

**Why TXT Records Matter**: Clients can read this metadata before connecting, enabling intelligent routing (e.g., Safari automatically constructs `http://[ip]:[port]/` from the `path=` TXT record).

### 4. Concurrent Execution
```go
// Start HTTP server in background goroutine
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("HTTP server error: %v", err)
    }
}()

// Main goroutine blocks on signal (Ctrl+C)
```

### 5. Graceful Shutdown Pattern
```go
// Trap Ctrl+C
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan

// Shutdown HTTP server with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)

// Close mDNS responder
resp.Close()
```

## Running the Example

### 1. Start the Server
```bash
cd examples/intermediate/web-server
make run
```

**Expected Output**:
```
HTTP server listening on :8080
mDNS service registered: Web Demo._http._tcp.local
TXT records: path=/, version=1.0
Press Ctrl+C to stop
```

### 2. Discover the Service

**Option A: macOS Safari**
1. Open Safari
2. Bookmarks → Show All Bookmarks
3. Look under "Bonjour" section
4. Click "Web Demo" → Opens `http://[ip]:8080/`

**Option B: Command Line**
```bash
# Browse for HTTP services
dns-sd -B _http._tcp

# Expected output:
# Browsing for _http._tcp
# Timestamp     A/R Flags  Domain   Service Type  Instance Name
# 10:00:00.000  Add     2  local.   _http._tcp.   Web Demo

# Resolve service details
dns-sd -L "Web Demo" _http._tcp

# Expected output shows:
# - Hostname: your-machine.local
# - Port: 8080
# - TXT: path=/, version=1.0
```

### 3. Access the Web Server
```bash
curl http://localhost:8080
# Response: Hello from mDNS-discoverable server!
```

## Integration Patterns

### Pattern 1: Existing Web Frameworks (Gin, Echo, Chi)

```go
// Gin example
router := gin.Default()
router.GET("/", func(c *gin.Context) {
    c.String(200, "Hello from Gin + mDNS!")
})

server := &http.Server{
    Addr:    ":8080",
    Handler: router,
}

// Register mDNS (same as example)
service := responder.ServiceDefinition{
    ServiceInstanceName: "Gin API",
    ServiceType:         "_http._tcp",
    Port:                8080,
    TXTRecords:          []string{"framework=gin", "api_version=v1"},
}
```

### Pattern 2: TLS/HTTPS Services

For HTTPS servers, use `_https._tcp` service type:
```go
service := responder.ServiceDefinition{
    ServiceType: "_https._tcp",  // Note: _https, not _http
    Port:        8443,
    TXTRecords:  []string{"path=/", "tls=1.3"},
}
```

### Pattern 3: Multiple HTTP Endpoints

Advertise multiple services with different paths:
```go
// Main API
serviceAPI := responder.ServiceDefinition{
    ServiceInstanceName: "My App API",
    ServiceType:         "_http._tcp",
    Port:                8080,
    TXTRecords:          []string{"path=/api", "version=2.0"},
}

// Admin Dashboard
serviceAdmin := responder.ServiceDefinition{
    ServiceInstanceName: "My App Admin",
    ServiceType:         "_http._tcp",
    Port:                8080,
    TXTRecords:          []string{"path=/admin", "auth=required"},
}
```

## What Could Go Wrong?

### Issue 1: Service Not Visible
**Symptoms**: `dns-sd -B _http._tcp` doesn't show your service
**Causes**:
- Firewall blocking UDP port 5353 (mDNS)
- VPN active (mDNS is link-local only, doesn't route over VPN)
- Docker network isolation (requires bridge configuration)

**Fix**:
```bash
# Linux: Allow mDNS through firewall
sudo ufw allow 5353/udp

# macOS: Check firewall settings
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate
```

### Issue 2: HTTP Server Not Starting
**Symptoms**: `listen tcp :8080: bind: address already in use`
**Fix**: Another process is using port 8080. Find and stop it:
```bash
# Linux
sudo lsof -i :8080
sudo kill <PID>

# macOS
lsof -i :8080
kill <PID>
```

### Issue 3: Shutdown Hangs
**Symptoms**: Program doesn't exit after Ctrl+C
**Cause**: HTTP server has active connections exceeding 30-second timeout
**Fix**: Reduce timeout or force-kill connections:
```go
// Use shorter timeout for dev environments
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

## Production Deployment Considerations

### 1. Firewall Configuration
Ensure UDP port 5353 is open for mDNS traffic:
```bash
# Linux (ufw)
sudo ufw allow 5353/udp

# Linux (iptables)
sudo iptables -A INPUT -p udp --dport 5353 -j ACCEPT

# macOS: No action needed (mDNS enabled by default)
```

### 2. systemd Service
Deploy as a systemd service for automatic startup:
```ini
# /etc/systemd/system/web-server-mdns.service
[Unit]
Description=Web Server with mDNS Discovery
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/web-server
ExecStart=/opt/web-server/web-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable web-server-mdns
sudo systemctl start web-server-mdns
```

### 3. Docker Deployment
When running in Docker, use host networking mode:
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o web-server .

FROM debian:bookworm-slim
COPY --from=builder /app/web-server /usr/local/bin/
EXPOSE 8080
CMD ["web-server"]
```

```bash
# CRITICAL: Use --network host for mDNS to work
docker run --network host web-server
```

**Note**: Bridge networking (`docker run -p 8080:8080`) will NOT work for mDNS because mDNS requires multicast group membership on the host interface.

### 4. Multi-Interface Hosts
On machines with multiple network interfaces (WiFi + Ethernet), mDNS announces on all interfaces by default. To restrict to specific interfaces, see F-10 Network Interface Management specification.

### 5. Security Considerations
- **No Authentication**: This example has no authentication. Add authentication for production (e.g., basic auth, JWT)
- **Rate Limiting**: Consider rate-limiting HTTP endpoints to prevent abuse
- **TXT Metadata**: Don't expose sensitive information in TXT records (they're publicly visible)

## RFC References

- **RFC 6763 §6**: TXT Record Construction - Defines key=value format
- **RFC 6763 §4.1.3**: Standard TXT Keys - `path=` and URL construction
- **RFC 6762 §10**: TTL Values - Service records use 120-second TTL

## Related Examples

- **Example 1** (Basic): `hello-responder` - Basic mDNS responder without HTTP
- **Example 7** (Intermediate): `service-updates` - Updating TXT records dynamically
- **Example 10** (Intermediate): `logging-integration` - Production logging patterns

## Next Steps

1. **Add TLS Support**: Modify to use `_https._tcp` and `server.ListenAndServeTLS()`
2. **Dynamic TXT Updates**: See Example 7 for updating TXT records at runtime
3. **Multiple Services**: Register both API and admin interfaces with different TXT records
4. **Health Checks**: Add `/health` endpoint and advertise in TXT records

---

**Status**: Production-ready intermediate example
**Last Updated**: 2026-01-06
**Beacon Version**: v1.0+
