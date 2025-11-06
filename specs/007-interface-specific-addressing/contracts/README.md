# API Contracts: Interface-Specific IP Address Advertising

**Feature**: 007-interface-specific-addressing
**Version**: 1.0
**Date**: 2025-11-06

## Overview

This directory contains API contract specifications for the interface-specific addressing feature. These are **specification files**, not compiled code. They document the expected interface changes and serve as contracts for implementation.

## Contract Files

### 1. `transport_interface.go`

**Purpose**: Documents the breaking change to the `Transport` interface

**Key Changes**:
- `Transport.Receive()` signature change: Added `interfaceIndex int` return value
- Enables RFC 6762 §15 compliance by providing interface context

**Impact**:
- ✅ Internal interface - no public API break
- ⚠️ Requires updates to all implementers (UDPv4Transport, MockTransport)
- ⚠️ Requires updates to all test callsites

**Migration Guide**: Included in file

### 2. `interface_resolver.go`

**Purpose**: Documents the new `getIPv4ForInterface()` helper function

**Key Features**:
- Interface index → IPv4 address lookup
- Error handling for missing interfaces and IPv4-less interfaces
- Validation rules for IPv4 addresses in responses

**Policies**:
- `FallbackBehavior`: SkipResponse vs UseGlobalIP
- `InterfaceResolutionPolicy`: Configuration for future extensibility

## Usage

These contract files are **reference documentation** for developers implementing the feature. They:

1. **Define expected behavior** - What each API should do
2. **Document edge cases** - How to handle errors and unusual scenarios
3. **Provide migration guides** - How to update existing code
4. **Specify validation rules** - What constitutes a valid response

## Implementation Checklist

Use these contracts to verify implementation:

### Transport Interface

- [ ] `UDPv4Transport.Receive()` returns 4 values (packet, src, ifIndex, err)
- [ ] Interface index extracted from `ipv4.ControlMessage.IfIndex`
- [ ] Returns 0 if control message unavailable (graceful degradation)
- [ ] `MockTransport` updated to return mock interface index

### Interface Resolver

- [ ] `getIPv4ForInterface(ifIndex)` implemented in `responder/responder.go`
- [ ] Uses `net.InterfaceByIndex()` for interface lookup
- [ ] Uses `iface.Addrs()` for IP address lookup
- [ ] Returns `NetworkError` if interface not found
- [ ] Returns `ValidationError` if no IPv4 address on interface
- [ ] Returns first IPv4 address if multiple IPs present

### Responder Integration

- [ ] `listenForQueries()` extracts interface index from `Receive()`
- [ ] `handleQuery()` calls `getIPv4ForInterface(ifIndex)` instead of `getLocalIPv4()`
- [ ] Falls back to `getLocalIPv4()` if `ifIndex == 0` (graceful degradation)
- [ ] Skips response if interface lookup fails (SkipResponse policy)
- [ ] Logs warnings on interface resolution failures

### Testing

- [ ] Contract test `TestRFC6762_Section15_InterfaceSpecificAddresses` passes
- [ ] Unit test `TestGetIPv4ForInterface` covers all edge cases
- [ ] Integration test validates multi-interface scenarios
- [ ] All existing tests updated to handle 4-value `Receive()` return

## RFC Compliance

These contracts implement:

**RFC 6762 §15 "Responding to Address Queries"** (lines 1020-1024):

> When a Multicast DNS responder sends a Multicast DNS response message
> containing its own address records, it MUST include all addresses
> that are valid on the interface on which it is sending the message,
> and MUST NOT include addresses that are not valid on that interface.

**Compliance Strategy**:

1. ✅ Determine receiving interface via `Transport.Receive()` interface index
2. ✅ Look up IPv4 address for that interface via `getIPv4ForInterface()`
3. ✅ Build A record with ONLY that IP address
4. ✅ Never include IPs from other interfaces (validation rules)

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-06 | Initial contract specification. Transport.Receive() signature change, getIPv4ForInterface() API, validation rules. |

## See Also

- [spec.md](../spec.md) - Feature requirements
- [research.md](../research.md) - Technical research and decisions
- [data-model.md](../data-model.md) - Data structures and entities
- [plan.md](../plan.md) - Implementation plan
