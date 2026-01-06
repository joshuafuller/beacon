# Deployment Guides - Contracts

**Date**: 2026-01-06
**Feature**: 008-documentation-production-polish
**Priority**: P0 (Critical for v1.0)

## Purpose

This document specifies the interface and content contracts for the 4 deployment guides. These guides target production deployments and provide operational knowledge for running Beacon services reliably.

---

## Guide 1: Production Checklist

### Contract Specification

**Path**: `docs/deployment/production-checklist.md`
**Estimated Time**: 2 hours implementation
**Target Audience**: DevOps engineers preparing for production

### Purpose
Provide a comprehensive pre-deployment validation checklist to prevent common production issues.

### Content Contract

**Structure**:
1. **Overview** (1 paragraph)
2. **Pre-Deployment Checklist** (bulleted checklist with checkboxes)
3. **Validation Commands** (how to verify each item)
4. **Common Gotchas** (things that look like they work but don't)
5. **Next Steps** (links to monitoring, troubleshooting)

### Checklist Categories

#### Network Configuration
- [ ] Port 5353 (UDP) open on firewall (inbound and outbound)
- [ ] Multicast routing enabled on network interfaces
- [ ] Multicast address 224.0.0.251 not blocked by switch/router
- [ ] No NAT between service and clients (mDNS is link-local only)
- [ ] VPN interfaces excluded from mDNS (if applicable)
- [ ] Docker network configured with `network_mode: "host"` (if using Docker)

**Validation Commands**:
```bash
# Check port availability
sudo netstat -ulnp | grep 5353

# Test multicast reachability
ping -c 3 224.0.0.251

# Verify no NAT
ip route show | grep 224.0.0.0
```

#### Interface Selection
- [ ] Correct network interface selected (not loopback)
- [ ] Interface has valid IPv4 address
- [ ] Interface supports multicast (check `ip link show [interface]`)
- [ ] Multiple interfaces handled correctly (if applicable)
- [ ] Interface-specific IP addressing enabled (RFC 6762 §15)

**Validation Commands**:
```bash
# List network interfaces
ip addr show

# Check multicast support
ip maddr show [interface]

# Verify IPv4 address
ip -4 addr show [interface]
```

#### Service Configuration
- [ ] Service instance name is unique on network
- [ ] Service type follows RFC 6763 naming (`_<service>._tcp.local`)
- [ ] Port number is correct and service is listening
- [ ] TXT records are valid (key=value pairs, no spaces in keys)
- [ ] TTL values appropriate (default: 120 seconds for unicast, 4500 for multicast)

**Validation Commands**:
```bash
# Check for name conflicts (macOS)
dns-sd -B _http._tcp local

# Check for name conflicts (Linux)
avahi-browse -t _http._tcp

# Verify service is listening
netstat -tuln | grep [port]
```

#### Resource Limits
- [ ] Sufficient file descriptors (ulimit -n ≥ 1024)
- [ ] Network buffer sizes adequate (net.core.rmem_max, wmem_max)
- [ ] No connection tracking limits for multicast (nf_conntrack)
- [ ] Memory limits allow for buffer pooling

**Validation Commands**:
```bash
# Check file descriptor limit
ulimit -n

# Check network buffer sizes
sysctl net.core.rmem_max
sysctl net.core.wmem_max

# Check memory limits (Docker)
docker stats [container]
```

#### Security
- [ ] Rate limiting enabled (RFC 6762 §6.2 - 1 query per second per interface)
- [ ] Source IP filtering configured (no external queries)
- [ ] Service type allowlist defined (for bridge mode)
- [ ] TXT records don't contain sensitive data

#### Monitoring
- [ ] Structured logging configured (JSON format recommended)
- [ ] Health check endpoint available (if web service)
- [ ] Metrics collection enabled
- [ ] Alerting configured for service unavailability

### Success Criteria
- [ ] Checklist covers all P0 production requirements
- [ ] Every checklist item has validation command
- [ ] Guide links to troubleshooting for failed checks
- [ ] Checklist usable as actual pre-deploy runbook

### References
- F-9: Transport Layer Configuration
- F-10: Network Interface Management
- F-11: Security Architecture

---

## Guide 2: Docker Deployment

### Contract Specification

**Path**: `docs/deployment/docker.md`
**Estimated Time**: 2 hours implementation
**Target Audience**: Teams deploying Beacon in containers

### Purpose
Provide working Docker and docker-compose configurations for Beacon services, with multicast requirements explained.

### Content Contract

**Structure**:
1. **Overview** - Why Docker requires special configuration for multicast
2. **Quick Start** - Minimal working example (3-5 commands)
3. **Dockerfile Template** - Production-ready multi-stage build
4. **docker-compose.yml Template** - Complete orchestration
5. **Multicast Configuration** - `network_mode: "host"` rationale and alternatives
6. **Production Considerations** - Resource limits, health checks, logging
7. **Troubleshooting** - Common Docker-specific issues

### Key Sections

#### Multicast in Docker (Critical)

**Problem**: Docker's default bridge network does not support multicast. mDNS requires multicast to 224.0.0.251:5353.

**Solution 1 - Host Network (Recommended for Development)**:
```yaml
version: "3.8"
services:
  beacon-service:
    build: .
    network_mode: "host"  # Required for multicast
    restart: unless-stopped
```

**Trade-offs**:
- ✅ Simple configuration
- ✅ Works immediately
- ❌ No port isolation (service port must not conflict with host)
- ❌ No container network isolation

**Solution 2 - Macvlan Network (Advanced/Production)**:
```yaml
version: "3.8"
networks:
  mdns-net:
    driver: macvlan
    driver_opts:
      parent: eth0  # Host interface
    ipam:
      config:
        - subnet: 192.168.1.0/24
          gateway: 192.168.1.1

services:
  beacon-service:
    build: .
    networks:
      - mdns-net
    restart: unless-stopped
```

**Trade-offs**:
- ✅ Container network isolation
- ✅ Multicast support
- ❌ Complex setup (requires pre-configured Docker network)
- ❌ Containers get own MAC addresses (may require network admin approval)

**Recommendation**: Use `network_mode: "host"` for examples and development. Document macvlan for advanced users.

#### Dockerfile Template (Multi-Stage Build)

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /beacon-service ./main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /beacon-service .

# Expose application port (NOT 5353 - that's mDNS control plane)
EXPOSE 8080

# Health check (if applicable)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the service
CMD ["./beacon-service"]
```

#### docker-compose.yml Template (Complete)

```yaml
version: "3.8"

services:
  beacon-service:
    build:
      context: .
      dockerfile: Dockerfile
    network_mode: "host"  # Required for mDNS multicast
    restart: unless-stopped

    environment:
      # Service configuration
      - SERVICE_NAME=beacon-demo
      - SERVICE_TYPE=_http._tcp
      - SERVICE_PORT=8080

      # Logging
      - LOG_LEVEL=info
      - LOG_FORMAT=json

    volumes:
      # Config file (read-only)
      - ./config:/config:ro

    # Resource limits (production)
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M

    # Health check
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

### Success Criteria
- [ ] Dockerfile builds without errors
- [ ] docker-compose.yml starts service successfully
- [ ] Service discoverable via `dns-sd` or `avahi-browse` from host
- [ ] Guide explains multicast limitation and solutions
- [ ] Production considerations covered (resource limits, health checks)

### References
- research.md - Docker Multicast Configuration (Research Task 4)

---

## Guide 3: Monitoring

### Contract Specification

**Path**: `docs/deployment/monitoring.md`
**Estimated Time**: 2 hours implementation
**Target Audience**: Operations teams monitoring production services

### Purpose
Provide guidance on monitoring Beacon services using structured logging, metrics, and health checks.

### Content Contract

**Structure**:
1. **Overview** - Why monitoring matters for mDNS services
2. **Structured Logging** - `log/slog` integration (links to logging-integration example)
3. **Metrics** - Key metrics to track
4. **Health Checks** - Service health validation
5. **Alerting** - When to alert and thresholds
6. **Log Aggregation** - Integration with ELK, Splunk, Datadog

### Key Sections

#### Structured Logging with slog

**Log Schema**:
```json
{
  "time": "2026-01-06T10:00:00Z",
  "level": "INFO",
  "msg": "Service registered",
  "service": "Demo Service",
  "type": "_http._tcp",
  "port": 8080,
  "instance_id": "abc123"
}
```

**Log Levels**:
- **DEBUG**: Query details, packet inspection, performance metrics
- **INFO**: Service lifecycle (register, unregister, update)
- **WARN**: Rate limiting triggered, validation warnings, retries
- **ERROR**: Network errors, configuration errors, unrecoverable failures

**Example Integration**:
```go
// Configure slog with JSON handler
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Log service registration
logger.Info("Service registered",
    "service", svc.Instance,
    "type", svc.Service,
    "port", svc.Port,
)
```

#### Key Metrics to Track

**Service Availability**:
- Service registration success/failure count
- Time since last successful announcement
- Conflict detection events (RFC 6762 §8.2)

**Network Performance**:
- Query response latency (p50, p95, p99)
- Multicast packet loss rate
- Rate limiting events

**Resource Usage**:
- Network buffer pool utilization
- Active goroutines
- Memory allocation rate

**RFC Compliance**:
- Goodbye packets sent on shutdown
- Probing rounds completed (RFC 6762 §8.1 - should be 3)
- Announcement count (RFC 6762 §8.3 - should be 2 unsolicited)

#### Health Check Endpoint

**Recommendation**: If service exposes HTTP endpoint, add `/health`:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check if mDNS responder is running
    if responder == nil || !responder.IsRunning() {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "reason": "mDNS responder not running",
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "mdns": "running",
    })
})
```

#### Alerting Thresholds

**Critical Alerts** (page on-call):
- Service registration failed (can't announce to network)
- Network errors on multicast socket (port 5353 unreachable)
- Health check failing for >5 minutes

**Warning Alerts** (notify team):
- Rate limiting triggered frequently (>10 events/minute)
- High query response latency (p95 >100ms)
- Conflict detection events (may indicate name collision)

### Success Criteria
- [ ] Log schema documented with all fields
- [ ] Metrics list covers availability, performance, resources
- [ ] Health check example provided
- [ ] Alerting thresholds justified
- [ ] Guide links to logging-integration example (P1)

### References
- examples/intermediate/logging-integration (P1)

---

## Guide 4: Troubleshooting

### Contract Specification

**Path**: `docs/deployment/troubleshooting.md`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers and operators debugging issues

### Purpose
Document real-world failure scenarios with symptoms, diagnosis steps, and solutions.

### Content Contract

**Structure**:
1. **Overview** - How to use this guide
2. **Diagnostic Tools** - Required tools for troubleshooting (dns-sd, Wireshark, etc.)
3. **Problem Matrix** - Quick lookup table (symptom → problem)
4. **Detailed Scenarios** - 10+ scenarios with diagnosis and solutions
5. **Getting Help** - Where to report bugs, ask questions

### Problem Format (Template)

```markdown
### Problem: [Name]

**Symptom**: [What user observes]
**Common Causes**: [List of potential causes]

**Diagnosis**:
1. [First diagnostic step with command]
   ```bash
   command here
   ```
   Expected output: [...]

2. [Second diagnostic step]
   ...

**Solution**:
- [Primary solution]
- [Alternative solution if primary fails]

**Prevention**: [How to avoid this in future]
```

### Required Scenarios (Minimum 10)

#### Scenario 1: Service Not Visible on Network

**Symptom**: Service registered successfully (no errors in logs) but not visible in `dns-sd -B` or Safari's Bonjour browser.

**Common Causes**:
- Port 5353 blocked by firewall
- Multicast routing disabled
- Wrong network interface selected
- Service on different subnet (mDNS is link-local only)

**Diagnosis**:
1. Check firewall rules:
   ```bash
   # Linux (iptables)
   sudo iptables -L -n | grep 5353

   # macOS
   sudo pfctl -s rules | grep 5353
   ```

2. Test multicast reachability:
   ```bash
   # Send multicast ping
   ping -c 3 224.0.0.251
   ```

3. Capture mDNS traffic with Wireshark:
   - Filter: `udp.port == 5353`
   - Look for announcement packets from your service

**Solution**:
- Open port 5353 UDP (inbound and outbound):
  ```bash
  # Linux (ufw)
  sudo ufw allow 5353/udp

  # macOS (add to /etc/pf.conf)
  pass in proto udp from any to any port 5353
  ```

**Prevention**: Include firewall check in pre-deployment checklist

---

#### Scenario 2: Service Visible But Can't Connect

**Symptom**: Service appears in discovery tools with correct port, but clients can't connect to the service.

**Common Causes**:
- Service not actually listening on advertised port
- Port blocked by firewall (different from mDNS port 5353)
- IP address mismatch (advertising wrong IP)
- Service crashed after registration

**Diagnosis**:
1. Verify service is listening:
   ```bash
   netstat -tuln | grep [port]
   # Should show LISTEN state
   ```

2. Check advertised IP address:
   ```bash
   dns-sd -L "Service Name" _http._tcp
   # Look for IP address in output
   ```

3. Test connection manually:
   ```bash
   telnet [ip] [port]
   # Should connect if service is running
   ```

**Solution**:
- Start service on correct port before registering with mDNS
- Verify IP address matches interface (RFC 6762 §15)

---

#### Scenario 3: Port 5353 Already in Use

**Symptom**: Error: "bind: address already in use" when starting Beacon responder.

**Common Causes**:
- Avahi daemon running (Linux)
- Bonjour service running (macOS - shouldn't conflict)
- Another Beacon instance running
- Test program left running from previous session

**Diagnosis**:
1. Find process using port 5353:
   ```bash
   # Linux
   sudo lsof -i :5353

   # macOS
   sudo lsof -i :5353
   ```

**Solution**:
- **If Avahi**: Stop Avahi (or configure Beacon to use SO_REUSEPORT - see F-9)
  ```bash
  sudo systemctl stop avahi-daemon
  ```

- **If another Beacon instance**: Kill the process or use SO_REUSEPORT option:
  ```go
  r, err := responder.New(
      responder.WithReusePort(true), // Enable SO_REUSEPORT
  )
  ```

**Prevention**: Use SO_REUSEPORT by default (F-9 recommendation)

---

#### Scenario 4: Service Name Conflict

**Symptom**: Service renamed automatically (e.g., "My Service" → "My Service (2)").

**Common Causes**:
- Another device on network with same instance name
- Previous instance of your service still announcing (didn't send goodbye packet)
- Probing detected existing service (RFC 6762 §8.1)

**Diagnosis**:
1. Search for conflicting services:
   ```bash
   dns-sd -B _http._tcp
   # Look for duplicate instance names
   ```

2. Check Beacon logs for conflict detection:
   ```
   WARN: Conflict detected for "My Service._http._tcp.local"
   INFO: Renamed to "My Service (2)._http._tcp.local"
   ```

**Solution**:
- Use unique instance names (include hostname or UUID)
- Ensure previous instances send goodbye packets on shutdown

**Prevention**: Implement graceful shutdown (see graceful-shutdown example)

---

#### Scenario 5: High CPU Usage

**Symptom**: Beacon process consuming excessive CPU (>50% on idle).

**Common Causes**:
- Query flood (malicious or misconfigured client)
- Rate limiting not enabled
- Infinite loop in query processing (bug)
- Large network with many services (query storms)

**Diagnosis**:
1. Profile CPU usage:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile
   ```

2. Check query rate in logs:
   ```bash
   grep "Query received" /var/log/beacon.log | wc -l
   # Count queries per second
   ```

**Solution**:
- Enable rate limiting (RFC 6762 §6.2 - 1 query/sec per interface):
  ```go
  r, err := responder.New(
      responder.WithRateLimit(1.0), // 1 query/sec
  )
  ```

**Prevention**: Always enable rate limiting in production (F-11)

---

#### (Additional scenarios 6-10 would follow same format)

### Diagnostic Tools Reference

**Required Tools**:
| Tool | Platform | Purpose | Install |
|------|----------|---------|---------|
| `dns-sd` | macOS | Bonjour browser/query | Built-in |
| `avahi-browse` | Linux | mDNS service discovery | `apt install avahi-utils` |
| `Wireshark` | All | Packet capture and analysis | `apt install wireshark` |
| `netstat` | All | Port and connection status | Built-in |
| `lsof` | All | List open files/ports | Built-in |

**Optional Tools**:
- `tcpdump` - Command-line packet capture
- `nmap` - Network scanning
- `mtr` - Network path analysis

### Success Criteria
- [ ] At least 10 real-world scenarios documented
- [ ] Each scenario follows Problem Format template
- [ ] Diagnostic commands provided with expected outputs
- [ ] Solutions tested and verified
- [ ] Common tools documented with installation steps

### References
- RFC 6762 - mDNS specification (for protocol-specific issues)
- F-11 - Security Architecture (for rate limiting)

---

## Cross-Guide Consistency Requirements

All deployment guides MUST adhere to these standards:

### Command Examples
- [ ] All commands tested on Linux and macOS
- [ ] Expected outputs shown (not just "command here")
- [ ] Platform-specific commands labeled clearly
- [ ] Commands safe to run (no destructive operations without warning)

### Structure
- [ ] Quick Start section (≤5 commands to working state)
- [ ] Detailed explanations follow Quick Start
- [ ] Troubleshooting section in each guide
- [ ] Next Steps links to related guides

### Production Focus
- [ ] Security considerations highlighted
- [ ] Performance implications explained
- [ ] Resource requirements documented
- [ ] Real-world trade-offs discussed

### Cross-References
- [ ] Link to F-specs when discussing features (F-9, F-10, F-11)
- [ ] Link to examples when demonstrating concepts
- [ ] Link between guides (checklist → docker → monitoring → troubleshooting)

---

## Integration with P0 Tasks

| Guide | Tasks | Primary Output |
|-------|-------|----------------|
| **Production Checklist** | T032 | production-checklist.md |
| **Docker Deployment** | T033, T035 | docker.md + working docker-compose.yml |
| **Monitoring** | T034 | monitoring.md |
| **Troubleshooting** | T036 | troubleshooting.md |

**Checkpoint**: After T036, all 4 deployment guides must be complete, tested, and cross-linked.

---

**Status**: Contract specification complete, ready for implementation
