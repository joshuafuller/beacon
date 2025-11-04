# Troubleshooting Guide

**Last Updated**: 2025-11-04

Common issues, solutions, and debugging techniques for Beacon.

---

## Table of Contents

- [Service Discovery Issues](#service-discovery-issues)
- [Service Announcement Issues](#service-announcement-issues)
- [Network Issues](#network-issues)
- [Performance Issues](#performance-issues)
- [Platform-Specific Issues](#platform-specific-issues)
- [Debugging Tools](#debugging-tools)

---

## Service Discovery Issues

### "No services found" but I know they exist

**Symptoms**:
```go
results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
fmt.Println(len(results))  // Outputs: 0
```

**Common Causes**:

#### 1. Firewall Blocking mDNS Traffic

mDNS uses UDP port 5353 and multicast address 224.0.0.251.

**Check**:
```bash
# Linux
sudo iptables -L -n | grep 5353

# View all firewall rules
sudo iptables -L -v
```

**Fix**:
```bash
# Linux - Allow mDNS traffic
sudo iptables -A INPUT -p udp --dport 5353 -j ACCEPT
sudo iptables -A OUTPUT -p udp --dport 5353 -j ACCEPT

# macOS - Check firewall
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# Windows - Allow mDNS in Windows Firewall
netsh advfirewall firewall add rule name="mDNS" dir=in action=allow protocol=UDP localport=5353
```

#### 2. VPN Active

mDNS is **link-local only** and doesn't work over VPNs (by design).

**Check**:
```bash
# Linux - List network interfaces
ip addr show

# Look for VPN interfaces (tun0, vpn0, etc.)
```

**Fix**:
- Disconnect VPN for local network discovery
- OR use Beacon on the physical interface only

#### 3. Different Subnet

mDNS only works on the same local network segment.

**Check**:
```bash
# Verify you're on the same subnet as the service
ip addr show  # Your IP
ping <service-host-ip>  # Can you reach it?
```

**Fix**: Ensure devices are on the same network (same WiFi, same VLAN, etc.)

#### 4. Query Timeout Too Short

Devices may take time to respond.

**Check**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)  // Too short!
```

**Fix**:
```go
// Increase timeout to 3-5 seconds
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

#### 5. Querying Wrong Service Type

Service types must match exactly (case-sensitive).

**Check**:
```go
// ❌ WRONG
q.Query(ctx, "_HTTP._tcp.local", querier.QueryTypePTR)  // Wrong case

// ❌ WRONG
q.Query(ctx, "http._tcp.local", querier.QueryTypePTR)   // Missing underscore

// ✅ CORRECT
q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
```

**Fix**: Verify service type name with system tools (see [Debugging Tools](#debugging-tools))

---

### Query Returns Partial Results

**Symptom**: Some services found, but missing expected services

**Causes**:

1. **Response Lost** - UDP is unreliable; packets can be dropped
   - **Fix**: Retry the query 2-3 times

2. **Known-Answer Suppression** - Service thinks you already know
   - **Fix**: Wait 1 second and query again (cache will be stale)

3. **Service Timing** - Service is still probing/announcing
   - **Fix**: Wait 1-2 seconds after service starts, then query

---

## Service Announcement Issues

### "Address already in use" error

**Symptom**:
```
Failed to create responder: listen udp :5353: bind: address already in use
```

**Causes**:

#### 1. SO_REUSEPORT Not Supported (Unlikely)

Beacon uses SO_REUSEPORT to share port 5353. This should work on all modern systems.

**Check**:
```bash
# Linux - Check kernel version (needs 3.9+)
uname -r

# macOS - Should work on all versions
# Windows - Should work on Windows 10+
```

**Fix**: Update OS if running ancient kernel/Windows version

#### 2. Multiple Beacon Responders Without SO_REUSEPORT

If you're creating multiple responders in the same process:

**Check**:
```go
// ❌ WRONG - Creates multiple responders
r1, _ := responder.New(ctx)
r2, _ := responder.New(ctx)  // May fail
```

**Fix**:
```go
// ✅ CORRECT - Reuse one responder for multiple services
r, _ := responder.New(ctx)
r.Register(ctx, service1)
r.Register(ctx, service2)
r.Register(ctx, service3)
```

#### 3. Non-SO_REUSEPORT Process Using Port 5353

Some older software binds exclusively to port 5353.

**Check**:
```bash
# Linux - See what's using port 5353
sudo ss -ulnp 'sport = :5353'

# macOS
sudo lsof -i :5353

# Windows
netstat -ano | findstr :5353
```

**Fix**: Stop the conflicting process or configure it to use a different port

---

### Service Name Keeps Changing

**Symptom**:
```
Registered: My Service._http._tcp.local
But appears as: My Service (2)._http._tcp.local
```

**Cause**: Another device on the network is using the same instance name

**Explanation**: This is **normal RFC 6762 behavior**! When Beacon detects a conflict during probing:
1. It appends " (2)" to the name
2. Probes again with the new name
3. If still conflicts, tries " (3)", etc.

**Solutions**:

1. **Use unique names** (recommended):
```go
hostname, _ := os.Hostname()
svc := &responder.Service{
    Instance: fmt.Sprintf("My Service (%s)", hostname),
    Service:  "_http._tcp",
    Port:     8080,
}
```

2. **Include MAC address**:
```go
import "net"

func getUniqueID() string {
    ifaces, _ := net.Interfaces()
    for _, iface := range ifaces {
        if len(iface.HardwareAddr) > 0 {
            return iface.HardwareAddr.String()[:8]  // First 8 chars
        }
    }
    return "unknown"
}

svc := &responder.Service{
    Instance: fmt.Sprintf("My Service (%s)", getUniqueID()),
    // ...
}
```

3. **Accept the rename** - It's working as designed!

---

### Service Not Discovered After Registration

**Symptom**:
```go
r.Register(ctx, svc)  // Returns no error
// But queries don't find it
```

**Causes**:

#### 1. Querying Too Soon After Registration

Registration involves probing (250ms) + announcing (1 second).

**Fix**:
```go
r.Register(ctx, svc)
time.Sleep(2 * time.Second)  // Wait for probing + announcing
// Now query
```

#### 2. Querying from Same Process

Beacon may suppress responses if the query comes from the same socket.

**Fix**: Query from a different process or machine

#### 3. Responder Closed

**Check**:
```go
r, _ := responder.New(ctx)
r.Register(ctx, svc)
r.Close()  // ❌ Service is now unregistered!
```

**Fix**:
```go
// Keep responder open
r, _ := responder.New(ctx)
defer r.Close()  // Only close on shutdown

r.Register(ctx, svc)
// Service remains announced while r is open
```

---

## Network Issues

### mDNS Traffic Not Visible on Network

**Check if mDNS packets are flowing**:

```bash
# Capture mDNS traffic
sudo tcpdump -i any port 5353 -v

# You should see packets like:
# 12:34:56.789 IP 192.168.1.100.5353 > 224.0.0.251.5353: ... (query)
# 12:34:56.790 IP 192.168.1.50.5353 > 224.0.0.251.5353: ... (response)
```

**If no traffic**:
1. Firewall is blocking
2. Wrong network interface
3. Multicast not enabled on interface

---

### Multicast Not Working

**Symptoms**: No packets sent/received

**Check multicast group membership**:
```bash
# Linux
ip maddr show

# Look for 224.0.0.251 on your interface
```

**Check multicast routing**:
```bash
# Linux
ip route show

# Should have a multicast route:
# 224.0.0.0/4 dev eth0 scope link
```

**Fix**:
```bash
# Linux - Enable multicast on interface
sudo ip link set eth0 multicast on

# Add multicast route
sudo route add -net 224.0.0.0 netmask 240.0.0.0 dev eth0
```

---

### "Permission Denied" When Creating Querier/Responder

**Symptom**:
```
Failed to create querier: listen udp :5353: bind: permission denied
```

**Cause**: On some systems, binding to port 5353 requires elevated privileges

**Fix**:
```bash
# Linux - Run with sudo (NOT recommended for production)
sudo ./myapp

# OR grant CAP_NET_BIND_SERVICE capability (better)
sudo setcap 'cap_net_bind_service=+ep' ./myapp

# OR use SO_REUSEPORT and let system daemon handle it
# (Beacon already does this)
```

---

## Performance Issues

### High CPU Usage

**Likely causes**:

1. **Receiving Too Many Queries** - Busy network
   - **Fix**: Rate limiting is built-in (RFC 6762 §6.2 compliance)
   - **Check**: Are you on a network with many mDNS devices?

2. **Tight Query Loop**
   ```go
   // ❌ WRONG - Hammers the network
   for {
       results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
   }
   ```
   **Fix**: Add delays between queries
   ```go
   ticker := time.NewTicker(5 * time.Second)
   for range ticker.C {
       results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
   }
   ```

3. **Large Number of Registered Services**
   - **Check**: How many services are registered?
   - **Limitation**: Beacon is optimized for 1-10 services per responder

---

### High Memory Usage

**Check allocations**:
```bash
# Run with memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze
go tool pprof mem.prof
```

**Likely causes**:

1. **Querier/Responder Not Closed**
   ```go
   // ❌ WRONG - Leaks sockets
   q, _ := querier.New()
   // Never closed!
   ```
   **Fix**: Always defer Close()
   ```go
   q, _ := querier.New()
   defer q.Close()
   ```

2. **Context Leaks**
   ```go
   // ❌ WRONG - Cancel never called
   ctx, cancel := context.WithCancel(context.Background())
   ```
   **Fix**:
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   ```

---

### Slow Query Response

**Expected**: mDNS queries typically complete in 1-3 seconds

**If slower than 5 seconds**:

1. **Network congestion** - Too much traffic
2. **Underpowered hardware** - Raspberry Pi, embedded device
3. **Context timeout too short** - Not giving devices time to respond

**Benchmark your setup**:
```bash
# Check query latency
go test -bench=BenchmarkQuery ./querier
```

**Expected result**: ~163 ns/op (query construction)
**Response time**: Depends on network (typically 1-3 seconds for devices to respond)

---

## Platform-Specific Issues

### Linux

#### Avahi Conflicts

**Symptom**: Beacon works but Avahi stops responding

**Cause**: Misconfigured Avahi

**Fix**:
```bash
# Check Avahi config
cat /etc/avahi/avahi-daemon.conf

# Should have:
# [server]
# allow-interfaces=eth0  # Your interface
# enable-dbus=yes

# Restart Avahi
sudo systemctl restart avahi-daemon
```

#### Systemd-Resolved Conflicts

**Symptom**: Queries fail with "connection refused"

**Cause**: systemd-resolved may be interfering

**Fix**:
```bash
# Check if systemd-resolved is running
systemctl status systemd-resolved

# Disable mDNS in systemd-resolved
sudo nano /etc/systemd/resolved.conf
# Set: MulticastDNS=no

sudo systemctl restart systemd-resolved
```

---

### macOS

#### Bonjour Interference

**Status**: Code complete, integration tests pending

**Known issue**: Not integration-tested on macOS yet

**Workaround**: Should work via SO_REUSEPORT, but not validated

**Report**: If you encounter issues on macOS, please [open an issue](https://github.com/joshuafuller/beacon/issues)

---

### Windows

#### Windows Firewall Blocking

**Symptom**: No traffic sent/received

**Fix**:
```powershell
# Run PowerShell as Administrator
New-NetFirewallRule -DisplayName "mDNS Beacon" -Direction Inbound -Protocol UDP -LocalPort 5353 -Action Allow
New-NetFirewallRule -DisplayName "mDNS Beacon" -Direction Outbound -Protocol UDP -LocalPort 5353 -Action Allow
```

#### mDNS Service Conflicts

**Status**: Code complete, integration tests pending

**Known issue**: Not integration-tested on Windows yet

**Report**: If you encounter issues on Windows, please [open an issue](https://github.com/joshuafuller/beacon/issues)

---

## Debugging Tools

### System mDNS Tools

**Linux (Avahi)**:
```bash
# Browse all services
avahi-browse -a -t

# Browse specific service type
avahi-browse -r _http._tcp

# Resolve hostname
avahi-resolve -n myserver.local
```

**macOS (dns-sd)**:
```bash
# Browse all services
dns-sd -B _services._dns-sd._udp

# Browse specific service type
dns-sd -B _http._tcp

# Lookup service instance
dns-sd -L "My Service" _http._tcp

# Resolve hostname
dns-sd -G v4 myserver.local
```

**Windows (Bonjour SDK)**:
Download from Apple Developer

---

### Network Packet Capture

**tcpdump** (Linux/macOS):
```bash
# Capture mDNS traffic
sudo tcpdump -i any port 5353 -v -X

# Save to file for analysis
sudo tcpdump -i any port 5353 -w mdns.pcap

# Analyze with Wireshark
wireshark mdns.pcap
```

**Wireshark filters**:
```
# All mDNS traffic
mdns

# Only queries
mdns.flags.response == 0

# Only responses
mdns.flags.response == 1

# Specific service type
mdns.ptr.domain_name contains "_http._tcp"
```

---

### Beacon Debug Logging

**Enable verbose output** (coming in future release):
```go
// Future API (not yet implemented)
querier.New(querier.WithDebug(true))
```

**Current workaround**: Use packet capture

---

### Testing Connectivity

**Verify basic mDNS**:

1. **Start a responder**:
```go
// test-responder.go
r, _ := responder.New(context.Background())
svc := &responder.Service{
    Instance: "Test Service",
    Service:  "_http._tcp",
    Domain:   "local",
    Port:     8080,
    TXT:      []string{"test=true"},
}
r.Register(context.Background(), svc)
select {}  // Run forever
```

2. **Query from another terminal**:
```bash
# Linux
avahi-browse -r _http._tcp

# macOS
dns-sd -B _http._tcp
```

3. **Should see "Test Service"** in results

If this works but Beacon queries don't:
- Issue is in Beacon query code
- [Open an issue](https://github.com/joshuafuller/beacon/issues) with details

---

## Getting Help

### Before Opening an Issue

1. **Check this guide** - Common issues covered above
2. **Test with system tools** - Verify mDNS works on your network
3. **Packet capture** - Collect tcpdump/Wireshark output
4. **Minimal reproduction** - Create smallest program that shows the issue

### Opening an Issue

**Include**:
- Go version: `go version`
- Beacon version: Check `go.mod`
- OS and version: `uname -a` (Linux/macOS), `winver` (Windows)
- Network setup: WiFi? Ethernet? VPN?
- Code snippet: Minimal reproduction
- Error messages: Full output
- Packet capture: If available

**GitHub Issues**: https://github.com/joshuafuller/beacon/issues

### Community Support

**GitHub Discussions**: https://github.com/joshuafuller/beacon/discussions
- Ask questions
- Share use cases
- Get help from community

---

## Additional Resources

- [Getting Started Guide](getting-started.md) - Installation and basic usage
- [Querier Guide](querier-guide.md) - Advanced discovery patterns
- [Responder Guide](responder-guide.md) - Advanced announcement patterns
- [Platform Notes](platform-notes.md) - Platform-specific information
- [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) - mDNS specification

---

**Still stuck?** [Open a discussion](https://github.com/joshuafuller/beacon/discussions) - we're here to help!
