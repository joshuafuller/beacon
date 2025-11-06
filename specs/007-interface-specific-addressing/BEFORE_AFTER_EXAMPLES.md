# Before/After Examples: Interface-Specific IP Addressing

**Feature**: 007-interface-specific-addressing
**Issue**: #27 - Multi-interface hosts advertise wrong IP address
**RFC**: RFC 6762 §15 "Responding to Address Queries"

---

## Example 1: Laptop with WiFi + Ethernet

### Scenario Setup

**Hardware**: Developer laptop with two active network interfaces
- **WiFi (wlan0)**: 10.0.0.50 (connected to home network: 10.0.0.0/24)
- **Ethernet (eth0)**: 192.168.1.100 (connected to office network: 192.168.1.0/24)
- **Service**: Development web server registered via mDNS (`myapp._http._tcp.local`)

### Before Fix ❌

**Problem**: All queries receive the same IP (first interface found)

```
System picks: 192.168.1.100 (Ethernet IP) as default

┌─────────────────────────────────────────────────────────────┐
│ Query from WiFi Network (10.0.0.0/24)                      │
├─────────────────────────────────────────────────────────────┤
│ Client: avahi-browse on 10.0.0.25                          │
│ Query sent to: 224.0.0.251:5353 (mDNS multicast)          │
│ Received on: wlan0 (10.0.0.50)                             │
│                                                             │
│ Response from laptop:                                       │
│   myapp._http._tcp.local. 120 IN A 192.168.1.100  ❌      │
│                                                             │
│ Client attempts connection:                                 │
│   curl http://192.168.1.100:8080                           │
│   ❌ FAILS: No route to host (different subnet)            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Ethernet Network (192.168.1.0/24)               │
├─────────────────────────────────────────────────────────────┤
│ Client: avahi-browse on 192.168.1.50                       │
│ Query sent to: 224.0.0.251:5353 (mDNS multicast)          │
│ Received on: eth0 (192.168.1.100)                          │
│                                                             │
│ Response from laptop:                                       │
│   myapp._http._tcp.local. 120 IN A 192.168.1.100  ✅      │
│                                                             │
│ Client attempts connection:                                 │
│   curl http://192.168.1.100:8080                           │
│   ✅ SUCCESS: Connection established                        │
└─────────────────────────────────────────────────────────────┘
```

**Root Cause**: `getLocalIPv4()` returns first non-loopback IPv4 (192.168.1.100), used for ALL responses regardless of receiving interface.

**Impact**: 50% of clients cannot connect (WiFi clients get wrong IP).

### After Fix ✅

**Solution**: Interface-specific IP resolution via RFC 6762 §15

```
System extracts interface index from each query

┌─────────────────────────────────────────────────────────────┐
│ Query from WiFi Network (10.0.0.0/24)                      │
├─────────────────────────────────────────────────────────────┤
│ Client: avahi-browse on 10.0.0.25                          │
│ Query sent to: 224.0.0.251:5353 (mDNS multicast)          │
│ Received on: wlan0 (10.0.0.50)                             │
│                                                             │
│ Control message: IP_PKTINFO.IfIndex = 2 (wlan0)           │
│ Resolution: getIPv4ForInterface(2) → 10.0.0.50            │
│                                                             │
│ Response from laptop:                                       │
│   myapp._http._tcp.local. 120 IN A 10.0.0.50  ✅          │
│                                                             │
│ Client attempts connection:                                 │
│   curl http://10.0.0.50:8080                               │
│   ✅ SUCCESS: Connection established                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Ethernet Network (192.168.1.0/24)               │
├─────────────────────────────────────────────────────────────┤
│ Client: avahi-browse on 192.168.1.50                       │
│ Query sent to: 224.0.0.251:5353 (mDNS multicast)          │
│ Received on: eth0 (192.168.1.100)                          │
│                                                             │
│ Control message: IP_PKTINFO.IfIndex = 3 (eth0)            │
│ Resolution: getIPv4ForInterface(3) → 192.168.1.100        │
│                                                             │
│ Response from laptop:                                       │
│   myapp._http._tcp.local. 120 IN A 192.168.1.100  ✅      │
│                                                             │
│ Client attempts connection:                                 │
│   curl http://192.168.1.100:8080                           │
│   ✅ SUCCESS: Connection established                        │
└─────────────────────────────────────────────────────────────┘
```

**Result**: 100% of clients can connect (each gets correct IP for their network).

---

## Example 2: Multi-NIC Server with VLAN Isolation

### Scenario Setup

**Hardware**: Production server with three NIC cards, isolated VLANs
- **eth0**: 10.0.1.10 (Management VLAN, no routing to others)
- **eth1**: 10.0.2.10 (Application VLAN, no routing to others)
- **eth2**: 10.0.3.10 (Storage VLAN, no routing to others)
- **Service**: Database service registered via mDNS (`db._postgresql._tcp.local`)

### Before Fix ❌

**Problem**: All VLANs receive the same IP, violating VLAN isolation

```
System picks: 10.0.1.10 (Management VLAN IP) as default

┌─────────────────────────────────────────────────────────────┐
│ Query from Management VLAN (10.0.1.0/24)                   │
├─────────────────────────────────────────────────────────────┤
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.1.10   │
│ Connection: ✅ SUCCESS (same VLAN)                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Application VLAN (10.0.2.0/24)                  │
├─────────────────────────────────────────────────────────────┤
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.1.10   │
│ Connection: ❌ FAILS (no inter-VLAN routing)                │
│                                                             │
│ Security Issue: Advertises Management VLAN IP to App VLAN! │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Storage VLAN (10.0.3.0/24)                      │
├─────────────────────────────────────────────────────────────┤
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.1.10   │
│ Connection: ❌ FAILS (no inter-VLAN routing)                │
│                                                             │
│ Security Issue: Advertises Management VLAN IP to Storage!  │
└─────────────────────────────────────────────────────────────┘
```

**Impact**:
- 67% of clients cannot connect
- **Security risk**: Exposes Management VLAN IP to other VLANs

### After Fix ✅

**Solution**: Each VLAN gets its own IP, maintaining VLAN isolation

```
┌─────────────────────────────────────────────────────────────┐
│ Query from Management VLAN (10.0.1.0/24)                   │
├─────────────────────────────────────────────────────────────┤
│ Interface: eth0 (IfIndex=2)                                 │
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.1.10   │
│ Connection: ✅ SUCCESS                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Application VLAN (10.0.2.0/24)                  │
├─────────────────────────────────────────────────────────────┤
│ Interface: eth1 (IfIndex=3)                                 │
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.2.10   │
│ Connection: ✅ SUCCESS                                       │
│                                                             │
│ Security: Only Application VLAN IP advertised ✅            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Storage VLAN (10.0.3.0/24)                      │
├─────────────────────────────────────────────────────────────┤
│ Interface: eth2 (IfIndex=4)                                 │
│ Response: db._postgresql._tcp.local. 120 IN A 10.0.3.10   │
│ Connection: ✅ SUCCESS                                       │
│                                                             │
│ Security: Only Storage VLAN IP advertised ✅                │
└─────────────────────────────────────────────────────────────┘
```

**Result**:
- 100% of clients can connect
- **VLAN isolation preserved**: No cross-VLAN IP exposure

---

## Example 3: Docker Development Environment

### Scenario Setup

**Hardware**: Developer workstation with Docker
- **eth0**: 192.168.1.100 (Physical network: 192.168.1.0/24)
- **docker0**: 172.17.0.1 (Docker bridge: 172.17.0.0/16)
- **Service**: API service registered via mDNS (`api._http._tcp.local`)

### Before Fix ❌

**Problem**: Docker containers receive wrong IP

```
System picks: 192.168.1.100 (Physical interface) as default

┌─────────────────────────────────────────────────────────────┐
│ Query from Physical Network (192.168.1.0/24)               │
├─────────────────────────────────────────────────────────────┤
│ Response: api._http._tcp.local. 120 IN A 192.168.1.100    │
│ Connection: ✅ SUCCESS                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Docker Container (172.17.0.0/16)                │
├─────────────────────────────────────────────────────────────┤
│ Response: api._http._tcp.local. 120 IN A 192.168.1.100    │
│ Connection: ❌ FAILS (container cannot reach physical IP)   │
│                                                             │
│ Expected: Should receive docker0 bridge IP (172.17.0.1)    │
└─────────────────────────────────────────────────────────────┘
```

**Impact**: Docker containers cannot discover/connect to host services.

### After Fix ✅

**Solution**: Docker containers get Docker bridge IP

```
┌─────────────────────────────────────────────────────────────┐
│ Query from Physical Network (192.168.1.0/24)               │
├─────────────────────────────────────────────────────────────┤
│ Interface: eth0 (IfIndex=2)                                 │
│ Response: api._http._tcp.local. 120 IN A 192.168.1.100    │
│ Connection: ✅ SUCCESS                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Query from Docker Container (172.17.0.0/16)                │
├─────────────────────────────────────────────────────────────┤
│ Interface: docker0 (IfIndex=3)                              │
│ Response: api._http._tcp.local. 120 IN A 172.17.0.1       │
│ Connection: ✅ SUCCESS                                       │
│                                                             │
│ Result: Container can reach host via bridge IP ✅           │
└─────────────────────────────────────────────────────────────┘
```

**Result**: Both physical and Docker networks can connect.

---

## Example 4: Windows Graceful Degradation

### Scenario Setup

**Hardware**: Windows 10 machine with WiFi + Ethernet
- **WiFi**: 10.0.0.50
- **Ethernet**: 192.168.1.100
- **Control Messages**: Not supported on Windows

### Before Fix ❌

```
Same behavior as Example 1: Single IP for all queries
```

### After Fix ⚠️ (Graceful Degradation)

**Solution**: Falls back to default IP when control messages unavailable

```
┌─────────────────────────────────────────────────────────────┐
│ Platform: Windows (control messages unavailable)           │
├─────────────────────────────────────────────────────────────┤
│ Control message: cm = nil                                   │
│ Interface index: 0 (unknown)                                │
│ Resolution: Fallback to getLocalIPv4() → 192.168.1.100    │
│                                                             │
│ Response: myapp._http._tcp.local. 120 IN A 192.168.1.100  │
│                                                             │
│ Behavior: Same as before fix (single-interface behavior)   │
│ Status: ⚠️ Best-effort RFC compliance                       │
└─────────────────────────────────────────────────────────────┘
```

**Result**: Windows maintains single-interface behavior (no regression), full RFC 6762 §15 compliance on platforms with control message support.

---

## Code Comparison

### Before Fix

```go
// responder/responder.go (OLD)
func (r *Responder) handleQuery(packet []byte) {
    // ...

    // ❌ Same IP for all interfaces
    ipv4, err := getLocalIPv4()
    if err != nil {
        // Skip response
        return
    }

    // Build response with single IP for all queries
    // ...
}
```

### After Fix

```go
// responder/responder.go (NEW)
func (r *Responder) handleQuery(packet []byte, interfaceIndex int) {
    // ...

    // ✅ Interface-specific IP resolution
    if interfaceIndex == 0 {
        // Graceful degradation (Windows, platform limitations)
        ipv4, err = getLocalIPv4()
    } else {
        // RFC 6762 §15: Use ONLY the IP from receiving interface
        ipv4, err = getIPv4ForInterface(interfaceIndex)
    }

    if err != nil {
        // Skip response if interface lookup fails
        return
    }

    // Build response with interface-specific IP
    // ...
}
```

---

## Summary

| Scenario | Before Fix | After Fix | Impact |
|----------|-----------|-----------|--------|
| **Laptop (WiFi + Ethernet)** | 50% clients connect | 100% clients connect | ✅ +50% connectivity |
| **Multi-NIC Server (VLAN)** | 33% clients connect, security risk | 100% clients connect, VLAN isolation | ✅ +67% connectivity, ✅ Security |
| **Docker Environment** | Containers cannot connect | All networks connect | ✅ Docker compatibility |
| **Windows (no control msgs)** | 50% clients connect | 50% clients connect (graceful degradation) | ⚠️ No regression |

**Overall**: RFC 6762 §15 compliance achieved on all platforms with control message support. Graceful degradation maintains backward compatibility on platforms without control messages.
