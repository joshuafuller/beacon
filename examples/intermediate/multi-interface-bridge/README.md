# Example 8: Multi-Interface mDNS Bridge (IoT Gateway)

**Difficulty**: Advanced Intermediate
**Target Audience**: IoT developers building multi-network gateways
**Estimated Time**: 30 minutes

## Overview

This example demonstrates how to bridge mDNS queries between network interfaces, enabling service discovery across isolated subnets. This is critical for IoT edge gateways that connect different network segments.

**Real-World Scenario**: Raspberry Pi acting as an IoT gateway with:
- **WiFi** (wlan0, 192.168.1.0/24) - Connected to home network  
- **Ethernet** (eth0, 10.0.0.0/24) - Connected to isolated IoT sensor subnet

## Problem Statement

mDNS is **link-local** by design - queries don't cross router boundaries. This creates a problem:
- Sensors on IoT subnet (10.0.0.0/24) can't discover services on home network (192.168.1.0/24)
- Home automation apps on WiFi can't discover sensors on Ethernet

**Solution**: An mDNS bridge that forwards queries between interfaces while maintaining security and RFC compliance.

## What This Example Demonstrates

- **RFC 6762 §15 Compliance**: Multi-interface operations and forwarding rules
- **Interface-Specific Addressing**: Responses contain correct IP for each interface (007 feature)
- **Service Type Filtering**: Security control (only bridge allowed service types)
- **Subnet Exclusion**: Ignore Docker/VPN interfaces to avoid loops

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Raspberry Pi (Bridge Host)                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  wlan0 (192.168.1.10)              eth0 (10.0.0.1)         │
│       │                                  │                 │
│       │  ┌───────────────────────────┐  │                 │
│       └──┤  mDNS Bridge              ├──┘                 │
│          │  - Query forwarding       │                    │
│          │  - Service filtering      │                    │
│          │  - IP rewriting           │                    │
│          └───────────────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
       │                                      │
       │ WiFi Network                         │ IoT Subnet
       │ (192.168.1.0/24)                     │ (10.0.0.0/24)
       │                                      │
   ┌───┴────┐                            ┌────┴──────┐
   │ Laptop │                            │ Temp      │
   │        │                            │ Sensor    │
   └────────┘                            └───────────┘
```

## Key Concepts

### 1. RFC 6762 §15: Multi-Interface Operations

Per RFC 6762 §15.1:
> A host with multiple interfaces MUST send separate mDNS packets on each interface

**Bridge Behavior**:
- Listen for mDNS queries on BOTH interfaces
- Forward queries between interfaces (not just respond locally)
- Rewrite IP addresses to match target interface

### 2. Interface-Specific IP Addressing (007 Feature)

When forwarding responses, the bridge MUST provide interface-specific IPs:
```
Query on wlan0 → Forward to eth0 → Response contains 192.168.1.10 (wlan0 IP)
Query on eth0 → Forward to wlan0 → Response contains 10.0.0.1 (eth0 IP)
```

## Configuration

Edit `config.yaml`:

```yaml
bridge:
  interfaces:
    - wlan0
    - eth0
  
  # Only forward these service types (security)
  allowed_services:
    - _http._tcp
    - _homekit._tcp
    - _ipp._tcp
  
  # Exclude these subnets (avoid loops)
  exclude_subnets:
    - 172.17.0.0/16  # Docker
    - 10.8.0.0/24    # VPN
```

## Running the Example

**Prerequisites**: Multi-interface system (WiFi + Ethernet or two NICs)

```bash
cd examples/intermediate/multi-interface-bridge
make run
```

**Expected Output**:
```
=== mDNS Multi-Interface Bridge ===
Bridge started: wlan0 ↔ eth0
Allowed services: _http._tcp, _homekit._tcp
Excluding subnets: 172.17.0.0/16, 10.8.0.0/24

[wlan0 → eth0] Forwarding query: _http._tcp
[eth0 → wlan0] Forwarding response: Web Server (192.168.1.50)
```

## Security Considerations

### 1. Service Type Allowlist (CRITICAL)

**Why**: Without filtering, bridge exposes ALL services across networks - security risk!

```yaml
# ✅ GOOD: Explicit allowlist
allowed_services:
  - _http._tcp      # Public web services OK
  - _homekit._tcp   # IoT devices OK

# ❌ BAD: Wildcard (forwards everything!)
# allowed_services: ["*"]
```

### 2. Subnet Exclusions

Prevent bridging Docker/VPN traffic to avoid:
- Routing loops
- Exposing internal Docker services
- VPN traffic leaking to local network

```yaml
exclude_subnets:
  - 172.17.0.0/16  # Docker default
  - 10.8.0.0/24    # OpenVPN default
```

## RFC References

- **RFC 6762 §15**: Multi-Interface Operations
- **RFC 6762 §15.1**: Packet Transmission on Multiple Interfaces
- **RFC 6762 §6.7**: Legacy Unicast Responses (for bridging)

## Limitations

- **Performance**: Forwarding adds latency (~5-10ms per query)
- **Single-Homed Services**: Services only on one interface won't be bridged
- **No Routing**: This is NOT a router - only forwards mDNS, not general traffic

## Troubleshooting

| Problem | Diagnosis | Solution |
|---------|-----------|----------|
| No forwarding | Check interface names match system | Run `ip addr` to verify wlan0/eth0 exist |
| Services not visible | Service type not in allowlist | Add to `allowed_services` in config.yaml |
| High CPU usage | Too many services forwarded | Reduce allowlist, add more subnet exclusions |

## Related Examples

- **Example 7** (Intermediate): `service-updates` - Dynamic TXT records
- **F-10**: Network Interface Management - Interface selection spec

---

**Status**: Educational example (requires multi-interface hardware for testing)
**Last Updated**: 2026-01-06
**Beacon Version**: v1.0+
