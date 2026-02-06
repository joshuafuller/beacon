# Troubleshooting Guide

**Category**: Deployment
**Estimated Time**: Varies by issue
**Prerequisites**: Basic familiarity with Beacon and mDNS concepts

## Overview

This guide provides solutions for the 15 most common issues encountered when deploying Beacon-based mDNS services in production.

---

## Problem Format

Each problem follows this structure:
- **Symptom**: What you observe
- **Diagnosis**: How to identify the issue
- **Solution**: Step-by-step fix
- **Prevention**: How to avoid in the future

---

## Network & Discovery Issues

### 1. Service Not Visible on Network

**Symptom**: Service registers successfully but doesn't appear in `dns-sd -B` or `avahi-browse`

**Diagnosis**:
```bash
# Check if service registered
# Look for "service registered" in logs

# Verify responder is running
ps aux | grep beacon

# Check port 5353 is open
sudo lsof -i :5353
sudo netstat -ulnp | grep 5353

# Monitor multicast traffic
sudo tcpdump -i any udp port 5353 -v
# Should see mDNS packets
```

**Solution**:

1. **Check firewall** (most common cause):
   ```bash
   # Linux (ufw)
   sudo ufw allow 5353/udp
   sudo ufw allow from 224.0.0.0/4

   # Linux (iptables)
   sudo iptables -A INPUT -p udp --dport 5353 -j ACCEPT
   sudo iptables -A OUTPUT -p udp --dport 5353 -j ACCEPT
   sudo iptables -A INPUT -d 224.0.0.251 -j ACCEPT

   # macOS
   # System Preferences → Security & Privacy → Firewall → Firewall Options
   # Add your application to allowed list
   ```

2. **Verify multicast group membership**:
   ```bash
   # Check joined multicast groups
   netstat -g | grep 224.0.0.251

   # If not joined, check interface configuration
   ip maddr show
   ```

3. **Test with Avahi/Bonjour**:
   ```bash
   # If Avahi is running, verify SO_REUSEPORT is working
   # Beacon should coexist with system mDNS
   sudo systemctl status avahi-daemon  # Linux
   ```

**Prevention**:
- Run [Production Checklist](./production-checklist.md) before deployment
- Test discovery from another machine on the same subnet

---

### 2. Service Visible But Can't Connect

**Symptom**: Service appears in discovery but connection attempts fail

**Diagnosis**:
```bash
# Resolve service to IP and port
dns-sd -L "My Service" _http._tcp  # macOS
avahi-resolve -n "My Service._http._tcp.local"  # Linux

# Test TCP connection
nc -zv [resolved_ip] [resolved_port]

# Check if application is listening
sudo netstat -tlnp | grep [port]
```

**Solution**:

1. **Verify application is actually listening**:
   ```go
   // Common mistake: Registering service before HTTP server starts

   // ❌ WRONG - Race condition
   r.Register(svc)
   http.ListenAndServe(":8080", nil) // May not be ready yet

   // ✅ CORRECT - Start server first
   go http.ListenAndServe(":8080", nil)
   time.Sleep(100 * time.Millisecond) // Brief delay
   r.Register(svc)
   ```

2. **Check advertised port matches actual port**:
   ```go
   // Ensure Port field matches where service is listening
   svc := &responder.Service{
       Port: 8080, // Must match HTTP server port
   }
   ```

3. **Verify firewall allows application port**:
   ```bash
   # Port 5353 for mDNS, plus your application port
   sudo ufw allow 8080/tcp
   ```

**Prevention**:
- Add health check endpoint
- Verify service is fully started before registration
- Use integration tests that connect to discovered services

---

### 3. Port Conflicts (EADDRINUSE)

**Symptom**: Error "address already in use" when starting responder

**Diagnosis**:
```bash
# Find what's using port 5353
sudo lsof -i :5353
sudo netstat -tulpn | grep 5353

# Common culprits: Avahi, systemd-resolved, other Beacon instances
```

**Solution**:

**Option 1**: Beacon coexists with system mDNS (recommended)
```go
// Beacon v1.0+ automatically uses SO_REUSEPORT
// No configuration needed - just works with Avahi/Bonjour
r, err := responder.New(ctx)
```

**Option 2**: Stop conflicting service (if not needed)
```bash
# Stop Avahi (Linux)
sudo systemctl stop avahi-daemon
sudo systemctl disable avahi-daemon

# Stop mDNSResponder (macOS - NOT recommended)
# System services depend on it
```

**Option 3**: Use different interfaces
```go
// If running multiple Beacon instances
// Use explicit interface selection (M1.1+)
r, err := responder.New(ctx, responder.WithInterfaces([]string{"eth0"}))
```

**Prevention**:
- Beacon automatically handles SO_REUSEPORT since v1.0
- Use interface selection if running multiple instances
- Test on production-like environment first

---

### 4. Name Conflicts

**Symptom**: Service name changes unexpectedly (e.g., "My Service" becomes "My Service (2)")

**Diagnosis**:
```bash
# Check for existing services with same name
dns-sd -B _http._tcp local | grep "My Service"
avahi-browse -t _http._tcp | grep "My Service"
```

**Solution**:

1. **Use unique instance names**:
   ```go
   import "os"

   // Include hostname to ensure uniqueness
   hostname, _ := os.Hostname()
   svc := &responder.Service{
       InstanceName: fmt.Sprintf("My Service (%s)", hostname),
       // ...
   }
   ```

2. **Leverage automatic conflict resolution**:
   ```go
   // Beacon handles RFC 6762 §8.2 tie-breaking automatically
   // If conflict detected, probing is deferred and retried
   // No manual intervention needed
   ```

3. **Monitor for conflicts**:
   ```go
   // Check logs for conflict resolution messages
   // Beacon logs when it defers due to conflicts
   ```

**Prevention**:
- Use descriptive, unique instance names
- Include hostname or UUID in service names for multi-instance deployments
- Monitor service names in production

---

### 5. Services Disappearing After Network Change

**Symptom**: Services stop being advertised after network interface goes down/up or IP change

**Diagnosis**:
```bash
# Check network interfaces
ip link show

# Monitor interface events
ip monitor link

# Check logs for network errors
journalctl -u your-service -f | grep network
```

**Solution**:

1. **Implement network change detection**:
   ```go
   import "github.com/vishvananda/netlink"

   // Monitor for interface changes
   updates := make(chan netlink.LinkUpdate)
   done := make(chan struct{})

   netlink.LinkSubscribe(updates, done)

   go func() {
       for update := range updates {
           if update.Link.Attrs().Name == "eth0" {
               logger.Info("interface changed",
                   "name", update.Link.Attrs().Name,
                   "state", update.Link.Attrs().OperState,
               )

               // Re-register services
               r.Close()
               r, _ = responder.New(ctx)
               r.Register(svc)
           }
       }
   }()
   ```

2. **Use systemd restart on failure**:
   ```ini
   [Service]
   Restart=on-failure
   RestartSec=5s
   ```

**Prevention**:
- Monitor network interface state
- Implement automatic service re-registration
- Test on networks with dynamic IPs (DHCP)

---

## Application Errors

### 6. Validation Errors

**Symptom**: `svc.Validate()` fails with cryptic error

**Common Validation Errors**:

| Error | Cause | Solution |
|-------|-------|----------|
| `port must be in range 1-65535` | Port is 0 or > 65535 | Set valid port: `svc.Port = 8080` |
| `service type must match "_<service>._tcp.local"` | Invalid format | Use `_http._tcp.local` not `http.tcp.local` |
| `instance name must be 1-63 bytes` | Empty or too long | Keep instance names short and non-empty |
| `TXT record exceeds 1300 bytes` | Too much metadata | Reduce TXT record count or size |

**Solution**:
```go
// Always validate before registration
if err := svc.Validate(); err != nil {
    logger.Error("service validation failed",
        "error", err,
        "instance", svc.InstanceName,
        "port", svc.Port,
    )
    return err
}
```

**Prevention**:
- Call `Validate()` in tests
- Use constants for service types:
  ```go
  const (
      ServiceTypeHTTP = "_http._tcp.local"
      ServiceTypeSSH  = "_ssh._tcp.local"
  )
  ```

---

### 7. Context Timeout Too Short

**Symptom**: Registration fails with "context deadline exceeded"

**Diagnosis**:
```go
// Check logs for context errors
// Error: context deadline exceeded during probing
```

**Solution**:

1. **Increase timeout for probing** (default 250ms per probe × 3 = 750ms):
   ```go
   // Allow time for full probing sequence (250ms × 3 probes = 750ms)
   ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
   defer cancel()

   r, err := responder.New(ctx)
   ```

2. **Use background context for long-running services**:
   ```go
   // For services that run indefinitely
   ctx := context.Background()
   r, err := responder.New(ctx)
   ```

**Prevention**:
- Use 2-5 second timeout for `responder.New()`
- Use background context for production daemons
- Test on slow networks

---

### 8. Goodbye Packets Not Sent on Shutdown

**Symptom**: Services remain in DNS cache after shutdown (stale records)

**Diagnosis**:
```bash
# Monitor mDNS traffic during shutdown
sudo tcpdump -i any udp port 5353 -v

# Look for TTL=0 records (goodbye packets)
# Should see records with TTL 0 when r.Close() is called
```

**Solution**:

1. **Always call Close()**:
   ```go
   r, err := responder.New(ctx)
   if err != nil {
       return err
   }
   defer r.Close() // Ensures cleanup even on panic

   // Or with signal handling
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

   go func() {
       <-sigChan
       logger.Info("shutting down")
       r.Close() // Sends goodbye packets
       os.Exit(0)
   }()
   ```

2. **Verify goodbye packets in tests**:
   ```go
   // Check that Close() sends TTL=0 records
   r.Close()
   // Verify via packet capture or mock transport
   ```

**Prevention**:
- Always use `defer r.Close()`
- Handle SIGTERM/SIGINT for clean shutdown
- Test shutdown behavior in integration tests

---

## Docker & Container Issues

### 9. Multicast Not Working in Docker

**Symptom**: Service works on host but not in Docker container

**Diagnosis**:
```bash
# Check network mode
docker inspect beacon-service | grep NetworkMode
# Should show: "host"

# Check if container can reach multicast group
docker exec beacon-service ping -c 3 224.0.0.251
```

**Solution**:

**Required**: Use `network_mode: "host"`
```yaml
services:
  beacon-service:
    network_mode: "host"  # Not optional for multicast
```

**Why it's required**:
- Bridge networking uses NAT → breaks multicast
- Multicast group membership is per-interface
- mDNS is link-local only (RFC 6762 §11)

**Prevention**:
- Always use host networking for mDNS containers
- See [Docker Deployment Guide](./docker.md) for details

---

### 10. Container Health Check Failing

**Symptom**: Container marked unhealthy, restarts frequently

**Diagnosis**:
```bash
# Check health status
docker ps
# HEALTH column shows "unhealthy"

# View health check logs
docker inspect beacon-service | grep -A 10 Health
```

**Solution**:

1. **Verify endpoint is reachable**:
   ```bash
   docker exec beacon-service wget --quiet --tries=1 --spider http://localhost:8080/health
   # Should return 0 (success)
   ```

2. **Increase start-period** (service needs time to initialize):
   ```yaml
   healthcheck:
       start_period: 10s  # Increase if probing takes time
   ```

3. **Check health endpoint implementation**:
   ```go
   http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
       // Simple check - always returns OK if server is running
       w.WriteHeader(http.StatusOK)
       fmt.Fprintf(w, "OK")
   })
   ```

**Prevention**:
- Test health endpoint before deploying
- Use generous `start_period` for services with probing
- Monitor health check logs

---

## Performance Issues

### 11. High CPU Usage

**Symptom**: Beacon process consuming excessive CPU

**Diagnosis**:
```bash
# Check CPU usage
top -p $(pgrep beacon)

# Profile CPU
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Check for tight loops in logs
journalctl -u beacon -f | grep -i loop
```

**Common Causes**:

1. **Tight receive loop** (rare, should be handled in Beacon):
   - Solution: Ensure using latest Beacon version with buffer pooling

2. **Too many queries per second**:
   - Solution: Enable rate limiting (on by default)
   ```go
   q, err := querier.New(querier.WithRateLimitThreshold(50))
   ```

3. **Logging too verbosely**:
   - Solution: Use log sampling
   ```go
   logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
       Level: slog.LevelInfo, // Not LevelDebug in production
   }))
   ```

**Prevention**:
- Set resource limits in Docker/Kubernetes
- Monitor CPU usage in production
- Use pprof in staging to catch issues early

---

### 12. Memory Leaks

**Symptom**: Memory usage grows unbounded over time

**Diagnosis**:
```bash
# Monitor memory
top -p $(pgrep beacon)

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Check for goroutine leaks
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

**Common Causes**:

1. **Not calling Close()**:
   ```go
   // ❌ WRONG - Leaks goroutines
   r, _ := responder.New(ctx)
   // ... program continues without r.Close()

   // ✅ CORRECT
   r, _ := responder.New(ctx)
   defer r.Close()
   ```

2. **Context not cancelled**:
   ```go
   // ❌ WRONG - Context never cancelled
   ctx, cancel := context.WithCancel(context.Background())
   // ... never calls cancel()

   // ✅ CORRECT
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   ```

**Prevention**:
- Always use `defer r.Close()` and `defer cancel()`
- Monitor goroutine count in production
- Run with `-race` detector in testing

---

## Advanced Issues

### 13. Multi-Interface Confusion

**Symptom**: Service advertises wrong IP address on multi-homed hosts

**Diagnosis**:
```bash
# List all interfaces
ip addr show

# Check which IP is advertised
dns-sd -L "My Service" _http._tcp
# Verify IP matches desired interface
```

**Solution**:

1. **Explicit interface selection** (M1.1+):
   ```go
   r, err := responder.New(ctx,
       responder.WithInterfaces([]string{"eth0"}),
   )
   ```

2. **Interface filtering** (M1.1+):
   ```go
   r, err := responder.New(ctx,
       responder.WithInterfaceFilter(func(iface net.Interface) bool {
           // Only use interfaces starting with "eth"
           return strings.HasPrefix(iface.Name, "eth")
       }),
   )
   ```

3. **RFC 6762 §15 interface-specific addressing** (007+):
   - Beacon automatically uses correct IP per interface
   - Multi-interface hosts advertise different IPs on each network

**Prevention**:
- Use explicit interface selection on multi-homed hosts
- Test on machines with multiple interfaces
- Verify IP addresses match expectations

---

### 14. VPN Interference

**Symptom**: mDNS stops working when VPN connects

**Diagnosis**:
```bash
# Check routing table
ip route show

# Check which interface has multicast route
ip route show | grep 224.0.0.0

# Common issue: VPN becomes default route
```

**Solution**:

1. **Exclude VPN interfaces**:
   ```go
   r, err := responder.New(ctx,
       responder.WithInterfaceFilter(func(iface net.Interface) bool {
           // Exclude VPN interfaces
           return !strings.HasPrefix(iface.Name, "tun") &&
                  !strings.HasPrefix(iface.Name, "wg")
       }),
   )
   ```

2. **Add multicast route to physical interface**:
   ```bash
   # Add explicit route for multicast
   sudo ip route add 224.0.0.0/4 dev eth0
   ```

**Prevention**:
- Configure VPN to not route multicast traffic
- Use split tunneling
- Test with VPN connected

---

### 15. Subnet Isolation

**Symptom**: Services visible on one subnet but not another (common in enterprise networks)

**Diagnosis**:
```bash
# Check if multicast is forwarded between subnets
# This usually requires router configuration

# Test from each subnet
dns-sd -B _http._tcp local  # Run on both subnets
```

**Solution**:

**mDNS is link-local only** (RFC 6762 §11) - this is by design.

**Options**:

1. **Accept limitation** (recommended):
   - mDNS only works on same subnet/VLAN
   - Use DNS-SD with unicast DNS for cross-subnet

2. **IGMP snooping/querier** (network admin task):
   - Configure switch/router to forward multicast
   - Not guaranteed to work (violates RFC)

3. **Bridge subnets** (IoT use case):
   - Use intermediate service to relay mDNS between subnets
   - See [examples/intermediate/multi-interface-bridge](../../examples/intermediate/multi-interface-bridge/)

**Prevention**:
- Understand mDNS is link-local by design
- Use unicast DNS-SD for cross-subnet discovery
- Deploy services on same subnet as clients

---

## Debugging Tools

### tcpdump Filters

```bash
# Capture all mDNS traffic
sudo tcpdump -i any udp port 5353 -w mdns.pcap

# View specific queries
sudo tcpdump -i any udp port 5353 -A | grep "_http._tcp"

# Monitor goodbye packets (TTL=0)
sudo tcpdump -i any udp port 5353 -v | grep "ttl 0"
```

### Wireshark

1. Open `mdns.pcap` in Wireshark
2. Filter: `dns` (mDNS uses DNS wire format)
3. Look for:
   - Query messages (QR=0)
   - Response messages (QR=1)
   - Authority section (probing)
   - TTL=0 (goodbye packets)

### System Tools

```bash
# List multicast group memberships
netstat -g

# Show multicast addresses per interface
ip maddr show

# Monitor interface state
ip monitor link

# Check open UDP sockets
sudo lsof -i UDP:5353
```

---

## Getting Help

If you're still stuck:

1. **Check logs**: Enable debug logging to see detailed mDNS operations
2. **Capture packets**: Use tcpdump/Wireshark to see actual mDNS traffic
3. **Review checklist**: Re-run [Production Checklist](./production-checklist.md)
4. **Simplify**: Test with minimal example from [examples/basic/](../../examples/basic/)
5. **File issue**: Include logs, packet captures, and system info in GitHub issue

## References

- [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) - Multicast DNS
- [Production Checklist](./production-checklist.md) - Pre-deployment validation
- [Monitoring Guide](./monitoring.md) - Observability setup
- [Beacon Examples](../../examples/) - Working code samples
