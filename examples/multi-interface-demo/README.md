# Multi-Interface mDNS Demonstration

This example demonstrates the fix for **Issue #27**: Multi-interface hosts now advertise the correct IP address per interface, in compliance with **RFC 6762 §15**.

## The Problem (Before the Fix)

On a machine with multiple network interfaces (e.g., WiFi + Ethernet, or Ethernet + Docker), the mDNS responder would advertise the **same IP address** on all queries, regardless of which interface received the query.

### Example Failure Scenario

```
Machine: eth0 (192.168.1.100) + docker0 (172.17.0.1)

Query from physical network (192.168.1.x):
  → Responder advertises: 192.168.1.100 ✅
  → Client can connect

Query from Docker container (172.17.x.x):
  → Responder advertises: 192.168.1.100 ❌
  → Client CANNOT connect (wrong subnet!)
```

**Impact**: Docker containers, VPN clients, and devices on secondary networks cannot connect to services.

## The Solution (After the Fix)

The responder now extracts the **interface index** from each query's control message and responds with the **interface-specific IP address**.

### Example Success Scenario

```
Machine: eth0 (192.168.1.100) + docker0 (172.17.0.1)

Query from physical network (192.168.1.x):
  → Interface: eth0 (index 2)
  → Responder advertises: 192.168.1.100 ✅
  → Client can connect

Query from Docker container (172.17.x.x):
  → Interface: docker0 (index 3)
  → Responder advertises: 172.17.0.1 ✅
  → Client can connect!
```

**Result**: All clients can connect, regardless of which network they're on! 🎉

## How to Run

### Prerequisites

- A machine with **2+ network interfaces** (e.g., WiFi + Ethernet, Ethernet + Docker)
- Linux, macOS, or Windows
- `avahi-browse` (optional, for testing)

### Run the Demo

```bash
cd examples/multi-interface-demo
go run main.go
```

### Expected Output

```
=== Multi-Interface mDNS Demonstration ===
This example demonstrates RFC 6762 §15 interface-specific IP addressing

📡 Available Network Interfaces:
  • eth0 (index 2)
    → IPv4: 192.168.1.100
  • docker0 (index 3)
    → IPv4: 172.17.0.1

🚀 Starting mDNS Responder...
✅ Service registered: Multi-Interface Demo._http._tcp.local

📋 What's Happening (RFC 6762 §15 Compliance):

The responder is now listening on ALL interfaces (0.0.0.0:5353).
When a query arrives on a specific interface, the responder will:

  1. Extract the interface index from the IP_PKTINFO control message
  2. Resolve the IPv4 address for THAT specific interface
  3. Respond with ONLY that interface's IP address

This ensures clients can actually REACH the advertised IP address!

🔍 Expected Behavior:

  Query on eth0 (index 2):
    → Response will advertise: 192.168.1.100
    → Clients on eth0 network can connect to 192.168.1.100:8080 ✅

  Query on docker0 (index 3):
    → Response will advertise: 172.17.0.1
    → Clients on docker0 network can connect to 172.17.0.1:8080 ✅
```

## Testing the Fix

### Test 1: Query from Same Machine

Open a second terminal and run:

```bash
# Using avahi-browse (Linux)
avahi-browse -r _http._tcp

# Using DNS-SD (macOS)
dns-sd -B _http._tcp

# Using dig
dig @224.0.0.251 -p 5353 "Multi-Interface Demo._http._tcp.local" A
```

**What to Look For**: The A record should contain the IP of the interface where the query was sent.

### Test 2: Query from Docker Container

If you have Docker installed:

```bash
# Start a container with network tools
docker run -it --rm nicolaka/netshoot

# Inside the container:
avahi-browse -r _http._tcp
```

**Expected Result**: The container should see the **docker0 bridge IP** (e.g., `172.17.0.1`), NOT the physical network IP.

### Test 3: Query from Another Machine

On a different machine **on the same physical network**:

```bash
avahi-browse -r _http._tcp
```

**Expected Result**: The remote machine should see the **physical network IP** (e.g., `192.168.1.100`).

### Test 4: Verify Interface-Specific IPs

Run this script to verify the responder is resolving IPs correctly:

```bash
# Check what IP each interface would advertise
for iface in $(ip -o link show | awk -F': ' '{print $2}' | grep -v lo); do
    index=$(ip link show "$iface" | head -1 | awk '{print $1}' | tr -d ':')
    ip=$(ip -4 addr show "$iface" | grep inet | awk '{print $2}' | cut -d/ -f1 | head -1)
    if [ -n "$ip" ]; then
        echo "Interface $iface (index $index) → IP $ip"
    fi
done
```

**Expected Result**: Each interface should map to its own unique IP address.

## Platform Support

| Platform | Status | Behavior |
|----------|--------|----------|
| **Linux** | ✅ Full Support | IP_PKTINFO extracts interface index, full RFC 6762 §15 compliance |
| **macOS/BSD** | ✅ Expected to Work | IP_RECVIF extracts interface index (via `golang.org/x/net/ipv4`) |
| **Windows** | ⚠️ Graceful Degradation | Control messages not supported, falls back to single-interface behavior (no regression) |

## Technical Details

### Implementation

1. **Transport Layer** ([internal/transport/udp.go](../../internal/transport/udp.go)):
   - Wraps UDP connection with `ipv4.PacketConn`
   - Enables `SetControlMessage(ipv4.FlagInterface, true)` to receive interface index
   - Returns `interfaceIndex` from `Receive()` call

2. **Responder Layer** ([responder/responder.go](../../responder/responder.go)):
   - Receives `interfaceIndex` from transport
   - Calls `getIPv4ForInterface(interfaceIndex)` to resolve interface-specific IP
   - Includes ONLY that IP in the A record response

### RFC 6762 §15 Compliance

> **RFC 6762 §15**: "When a Multicast DNS responder sends a Multicast DNS response message containing its own address records in response to a query received on a particular interface, it **MUST include only addresses that are valid on that interface**, and **MUST NOT include addresses configured on other interfaces**."

This implementation fully complies with RFC 6762 §15 on platforms with control message support.

## Troubleshooting

### "Only 1 interface found"

The demo works best with 2+ interfaces. Options:
- Connect to WiFi AND Ethernet
- Install Docker (creates `docker0` bridge)
- Create a VPN connection

### "Service not discoverable"

Check that mDNS is allowed through your firewall:

```bash
# Linux (allow mDNS)
sudo ufw allow 5353/udp

# macOS (usually enabled by default)
# Windows (usually blocked, requires firewall rule)
```

### "Wrong IP advertised"

This likely indicates:
1. Platform doesn't support control messages (Windows) → Expected behavior
2. Query sent to multicast group but not bound to specific interface → Check query source

## Related Documentation

- **Issue**: [#27 - Multi-interface hosts advertise wrong IP address](https://github.com/joshuafuller/beacon/issues/27)
- **Specification**: [specs/007-interface-specific-addressing/spec.md](../../specs/007-interface-specific-addressing/spec.md)
- **Before/After Examples**: [BEFORE_AFTER_EXAMPLES.md](../../specs/007-interface-specific-addressing/BEFORE_AFTER_EXAMPLES.md)
- **RFC 6762 §15**: "Responding to Address Queries"

## Summary

This example demonstrates that multi-interface hosts now correctly advertise interface-specific IP addresses in mDNS responses, ensuring clients can actually connect to the advertised services! 🎉
