# Data Model: Interface-Specific IP Address Advertising

**Feature**: 007-interface-specific-addressing
**Date**: 2025-11-06
**Status**: Complete

## Overview

This document defines the data structures and entities required for interface-specific IP address advertising in mDNS responses, implementing RFC 6762 §15 compliance.

---

## Core Entities

### 1. InterfaceContext

**Purpose**: Represents the network interface that received an mDNS query, providing context for building interface-specific responses.

**Location**: Internal to transport and responder layers (not public API)

**Structure**:

```go
// InterfaceContext contains metadata about which network interface received a query.
//
// RFC 6762 §15: Responses MUST include addresses valid on the receiving interface
type InterfaceContext struct {
    // Index is the OS-assigned interface index (from ipv4.ControlMessage.IfIndex)
    // Zero indicates unknown or unspecified interface
    Index int

    // Name is the interface name for debugging (e.g., "eth0", "wlan0")
    // Empty string if interface lookup failed
    Name string

    // IPv4Address is the cached IPv4 address for this interface
    // Nil if not yet looked up or if interface has no IPv4
    IPv4Address net.IP
}
```

**Lifecycle**:
1. Created when `Transport.Receive()` extracts interface index from control message
2. Populated with interface name and IP via `getIPv4ForInterface()`
3. Passed to query handler and response builder
4. Discarded after response sent (no persistence)

**Validation Rules**:
- `Index > 0` (interface index must be positive)
- `IPv4Address != nil` when used for response building
- `IPv4Address.To4() != nil` (must be valid IPv4)

**Relationships**:
- Used by: `Transport.Receive()`, `Responder.handleQuery()`, `ResponseBuilder`
- References: `net.Interface` (via index), `net.IP` (IPv4 address)

---

### 2. Enhanced Transport Interface

**Purpose**: Extend the `Transport` interface to return interface context alongside received packets.

**Location**: `internal/transport/transport.go`

**Structure**:

```go
// Transport abstracts network I/O for mDNS queries and responses.
//
// M1-Refactoring: Interface decouples querier from network implementation
// 007-interface-specific: Added interfaceIndex return value for RFC 6762 §15
type Transport interface {
    // Send transmits packet to destination address.
    Send(ctx context.Context, packet []byte, dest net.Addr) error

    // Receive waits for incoming packet, returning packet data, source address,
    // and the network interface index that received the packet.
    //
    // RFC 6762 §15: Interface index enables building responses with
    // addresses valid on the receiving interface.
    //
    // Returns:
    //   - packet: Received DNS message bytes
    //   - src: Source address of the packet
    //   - ifIndex: OS interface index (from IP_PKTINFO/IP_RECVIF)
    //   - error: NetworkError on timeout or receive failure
    Receive(ctx context.Context) (packet []byte, src net.Addr, ifIndex int, err error)

    // Close releases network resources.
    Close() error
}
```

**Breaking Change Analysis**:
- ✅ Internal interface - no public API impact
- ⚠️ Requires updates to:
  - `UDPv4Transport` implementation
  - `MockTransport` test double
  - All test callsites using `Transport.Receive()`

**Migration**:
- All implementers must return interface index
- Return `0` if interface unknown (graceful degradation)
- Callers must handle new return value

---

### 3. Enhanced UDPv4Transport

**Purpose**: Concrete implementation of `Transport` that extracts interface context from IPv4 control messages.

**Location**: `internal/transport/udp.go`

**Structure**:

```go
// UDPv4Transport implements Transport with interface context extraction.
type UDPv4Transport struct {
    rawConn  net.PacketConn      // Underlying UDP connection
    ipv4Conn *ipv4.PacketConn    // Wrapper for control message access
}
```

**New Fields**:
- `ipv4Conn`: Wraps `rawConn` to enable `IP_PKTINFO`/`IP_RECVIF` control messages

**Modified Methods**:

```go
// NewUDPv4Transport creates transport with control message support.
func NewUDPv4Transport() (*UDPv4Transport, error) {
    // ... existing socket creation ...

    // Wrap connection for control message access
    ipv4Conn := ipv4.NewPacketConn(rawConn)

    // Enable interface index in control messages
    err = ipv4Conn.SetControlMessage(ipv4.FlagInterface, true)
    if err != nil {
        rawConn.Close()
        return nil, &errors.NetworkError{
            Operation: "enable control messages",
            Err:       err,
            Details:   "failed to set IP_PKTINFO/IP_RECVIF",
        }
    }

    return &UDPv4Transport{
        rawConn:  rawConn,
        ipv4Conn: ipv4Conn,
    }, nil
}

// Receive extracts interface index from control messages.
func (t *UDPv4Transport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
    // ... context deadline handling ...

    buffer := GetBuffer()
    defer PutBuffer(buffer)

    // Read with control messages
    n, cm, src, err := t.ipv4Conn.ReadFrom(*buffer)
    if err != nil {
        return nil, nil, 0, handleReceiveError(err)
    }

    // Extract interface index from control message
    interfaceIndex := 0  // Default: unknown interface
    if cm != nil {
        interfaceIndex = cm.IfIndex
    }

    result := make([]byte, n)
    copy(result, (*buffer)[:n])

    return result, src, interfaceIndex, nil
}
```

**State Transitions**:
1. **Initialization**: `NewUDPv4Transport()` creates and configures connection
2. **Control Message Setup**: `SetControlMessage(FlagInterface, true)`
3. **Receive Loop**: Each `Receive()` extracts interface index
4. **Shutdown**: `Close()` releases both `ipv4Conn` and `rawConn`

---

### 4. Interface Address Resolver

**Purpose**: Look up the IPv4 address assigned to a network interface.

**Location**: `responder/responder.go` (private helper function)

**Function Signature**:

```go
// getIPv4ForInterface returns the IPv4 address assigned to the specified interface.
//
// RFC 6762 §15: Responses MUST include addresses valid on the receiving interface.
//
// Parameters:
//   - ifIndex: Network interface index (from Transport.Receive or ipv4.ControlMessage)
//
// Returns:
//   - []byte: IPv4 address (4 bytes) in network byte order
//   - error: NetworkError if interface lookup fails, ValidationError if no IPv4
//
// Edge Cases:
//   - Interface not found → NetworkError
//   - Interface has no IPv4 address → ValidationError
//   - Interface has multiple IPs → returns first IPv4 (IPv6 out of scope)
func getIPv4ForInterface(ifIndex int) ([]byte, error)
```

**Algorithm**:

```go
func getIPv4ForInterface(ifIndex int) ([]byte, error) {
    // Step 1: Look up interface by index
    iface, err := net.InterfaceByIndex(ifIndex)
    if err != nil {
        return nil, &errors.NetworkError{
            Operation: "lookup interface",
            Err:       err,
            Details:   fmt.Sprintf("interface index %d not found", ifIndex),
        }
    }

    // Step 2: Get all addresses for this interface
    addrs, err := iface.Addrs()
    if err != nil {
        return nil, &errors.NetworkError{
            Operation: "get interface addresses",
            Err:       err,
            Details:   fmt.Sprintf("failed to get addresses for %s", iface.Name),
        }
    }

    // Step 3: Find first IPv4 address
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipv4 := ipnet.IP.To4(); ipv4 != nil {
                return ipv4, nil
            }
        }
    }

    // Step 4: No IPv4 found
    return nil, &errors.ValidationError{
        Field:   "interface",
        Value:   iface.Name,
        Message: "no IPv4 address found on interface",
    }
}
```

**Performance**:
- `InterfaceByIndex`: O(n) lookup through system interfaces (~100ns)
- `Interface.Addrs()`: Syscall to get addresses (~150ns)
- **Total**: ~250ns per call

**Caching**: Not implemented in fast-track fix (deferred to M4)

---

### 5. Query Context

**Purpose**: Bundle query metadata (packet, source, interface) for passing through responder pipeline.

**Location**: Internal to `Responder.listenForQueries()` and `handleQuery()`

**Structure**:

```go
// queryContext contains all metadata about an incoming mDNS query.
//
// This is an internal structure used within the responder to pass
// query information through the handling pipeline.
type queryContext struct {
    // Packet is the raw DNS query bytes
    Packet []byte

    // SourceAddr is the address of the querier
    SourceAddr net.Addr

    // InterfaceIndex is the index of the interface that received the query
    // Zero indicates unknown interface (use getLocalIPv4 fallback)
    InterfaceIndex int
}
```

**Usage**:

```go
// In listenForQueries goroutine
for {
    packet, src, ifIndex, err := r.transport.Receive(r.ctx)
    if err != nil {
        continue
    }

    qctx := queryContext{
        Packet:         packet,
        SourceAddr:     src,
        InterfaceIndex: ifIndex,
    }

    go r.handleQuery(qctx)  // Process in goroutine
}
```

**Alternative Approach** (NOT chosen):
Pass three separate parameters to `handleQuery()`. Rejected because:
- Less extensible (adding metadata requires signature change)
- More verbose callsites
- Harder to add optional fields later

---

## Data Flow

### Query Reception → Response Generation

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Transport.Receive()                                          │
│    - Read packet from socket                                    │
│    - Extract interface index from ipv4.ControlMessage.IfIndex   │
│    - Return (packet, srcAddr, ifIndex, nil)                     │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Responder.listenForQueries()                                 │
│    - Bundle into queryContext{packet, srcAddr, ifIndex}         │
│    - Dispatch to handleQuery(qctx) in goroutine                 │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Responder.handleQuery(qctx)                                  │
│    - Parse query message                                        │
│    - Extract service type from PTR query                        │
│    - Look up matching service in registry                       │
│    - Call getIPv4ForInterface(qctx.InterfaceIndex) ◄─ NEW      │
│    - Create ServiceWithIP with interface-specific IP            │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. getIPv4ForInterface(ifIndex)                                 │
│    - net.InterfaceByIndex(ifIndex) → iface                      │
│    - iface.Addrs() → []net.Addr                                 │
│    - Filter for first IPv4 address                              │
│    - Return ipv4 (4 bytes) or error                             │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. ResponseBuilder.BuildResponse()                              │
│    - Create A record with interface-specific IP                 │
│    - Create PTR, SRV, TXT records (unchanged)                   │
│    - Return DNS response message                                │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Transport.Send(response, multicastAddr)                      │
│    - Send response to 224.0.0.251:5353                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Error Handling

### Error Types and Recovery

#### 1. Interface Lookup Failure

**Scenario**: `net.InterfaceByIndex(ifIndex)` returns error

**Cause**: Interface removed after query received (USB Ethernet unplugged)

**Error Type**: `NetworkError`

**Recovery**:
```go
ipv4, err := getIPv4ForInterface(ifIndex)
if err != nil {
    // Log error for debugging
    log.Warnf("Interface lookup failed for index %d: %v", ifIndex, err)

    // Skip this query (don't send response)
    // Alternative: Fall back to getLocalIPv4() (degraded mode)
    return nil
}
```

**User Impact**: Query goes unanswered for this interface (client will retry)

#### 2. No IPv4 Address on Interface

**Scenario**: Interface exists but has no IPv4 address (IPv6-only interface)

**Cause**: Link-local IPv6, no IPv4 DHCP lease yet

**Error Type**: `ValidationError`

**Recovery**: Same as interface lookup failure (skip response)

**Future**: M4 will handle IPv6 with AAAA records

#### 3. Control Message Unavailable

**Scenario**: `ipv4.ControlMessage` is nil or `IfIndex == 0`

**Cause**: Platform doesn't support `IP_PKTINFO`, or socket option failed

**Error Type**: None (graceful degradation)

**Recovery**:
```go
interfaceIndex := 0
if cm != nil {
    interfaceIndex = cm.IfIndex
}

// In handleQuery:
if interfaceIndex == 0 {
    // Fall back to global IP lookup (pre-fix behavior)
    ipv4, err = getLocalIPv4()
} else {
    // Use interface-specific lookup
    ipv4, err = getIPv4ForInterface(interfaceIndex)
}
```

**User Impact**: Degrades to current behavior (incorrect IP on multi-interface hosts)

#### 4. Interface Index Out of Range

**Scenario**: Control message returns invalid `IfIndex` (corrupt, race condition)

**Cause**: Kernel bug, memory corruption

**Error Type**: `NetworkError` from `InterfaceByIndex()`

**Recovery**: Skip response, log warning

**Likelihood**: Extremely low (kernel-level issue)

---

## State Transitions

### UDPv4Transport Lifecycle

```
┌──────────┐
│  INIT    │  NewUDPv4Transport() called
└──────────┘
     │
     │ Create socket, join multicast group
     ▼
┌──────────┐
│ SETUP CM │  SetControlMessage(FlagInterface, true)
└──────────┘
     │
     │ Success
     ▼
┌──────────┐
│  READY   │  Receive() can extract interface index
└──────────┘
     │
     │ Close() called
     ▼
┌──────────┐
│  CLOSED  │  Resources released
└──────────┘
```

**Invariants**:
- Control messages enabled before first `Receive()`
- `ipv4Conn != nil` while in READY state
- `Close()` idempotent (safe to call multiple times)

---

## Validation Rules

### Interface Index Validation

```go
// Valid interface index: positive integer
func isValidInterfaceIndex(ifIndex int) bool {
    return ifIndex > 0
}

// Usage:
if !isValidInterfaceIndex(ifIndex) {
    // Fall back to getLocalIPv4() or skip response
}
```

### IPv4 Address Validation

```go
// Valid IPv4: 4-byte address, not loopback, not unspecified
func isValidIPv4ForResponse(ip net.IP) bool {
    ipv4 := ip.To4()
    if ipv4 == nil {
        return false  // Not IPv4
    }

    if ipv4.IsLoopback() {
        return false  // 127.0.0.0/8
    }

    if ipv4.IsUnspecified() {
        return false  // 0.0.0.0
    }

    return true
}
```

### Response Building Validation

```go
// Before building response, verify we have valid IP
if !isValidIPv4ForResponse(ipv4) {
    return nil, &errors.ValidationError{
        Field:   "ipv4",
        Value:   ipv4.String(),
        Message: "invalid IPv4 address for mDNS response",
    }
}
```

---

## Testing Data Structures

### MockTransport Interface Context

```go
// MockTransport simulates multi-interface scenarios for testing.
type MockTransport struct {
    // ReceiveResponses contains pre-defined responses with interface indices
    ReceiveResponses []ReceiveResponse

    receiveIndex int
}

type ReceiveResponse struct {
    Packet         []byte
    Source         net.Addr
    InterfaceIndex int  // ◄─ NEW: Simulate interface context
    Error          error
}

// Receive returns next mock response with interface index.
func (m *MockTransport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
    if m.receiveIndex >= len(m.ReceiveResponses) {
        return nil, nil, 0, io.EOF
    }

    resp := m.ReceiveResponses[m.receiveIndex]
    m.receiveIndex++

    return resp.Packet, resp.Source, resp.InterfaceIndex, resp.Error
}
```

### Test Scenario Data

```go
// Multi-interface test scenario
type InterfaceTestCase struct {
    Description    string
    InterfaceIndex int
    ExpectedIP     net.IP
    ShouldError    bool
}

var multiInterfaceTests = []InterfaceTestCase{
    {
        Description:    "Query on eth0 returns eth0 IP",
        InterfaceIndex: 1,  // Assume eth0 is index 1
        ExpectedIP:     net.ParseIP("10.0.0.5"),
        ShouldError:    false,
    },
    {
        Description:    "Query on wlan0 returns wlan0 IP",
        InterfaceIndex: 2,  // Assume wlan0 is index 2
        ExpectedIP:     net.ParseIP("192.168.1.100"),
        ShouldError:    false,
    },
    {
        Description:    "Invalid interface index returns error",
        InterfaceIndex: 999,
        ExpectedIP:     nil,
        ShouldError:    true,
    },
}
```

---

## Backwards Compatibility

### Maintaining getLocalIPv4()

**Why keep it?**:
- Still used during `Register()` (service registration time)
- Fallback when interface index unavailable (control message failure)
- Compatibility with single-interface machines

**Deprecation Path**:
```go
// getLocalIPv4 gets the first non-loopback IPv4 address.
//
// DEPRECATED for response building: Use getIPv4ForInterface() instead.
// Still required for service registration until M4 per-interface binding.
//
// RFC 6762 §15 violation: This function returns arbitrary IP, not interface-specific.
func getLocalIPv4() ([]byte, error) {
    // ... existing implementation ...
}
```

**M4 Migration**:
- Remove `getLocalIPv4()` entirely
- Service registration requires explicit interface selection
- Per-interface transports eliminate need for global IP

---

## Summary

### New Entities

1. **InterfaceContext** - Interface metadata (index, name, IP)
2. **Enhanced Transport.Receive()** - Returns interface index
3. **UDPv4Transport.ipv4Conn** - Control message wrapper
4. **getIPv4ForInterface()** - Interface-specific IP resolver
5. **queryContext** - Query metadata bundle

### Modified Entities

1. **Transport interface** - Added `ifIndex` return value
2. **UDPv4Transport** - Control message extraction
3. **Responder.handleQuery()** - Interface-aware IP lookup

### Relationships

```
Transport.Receive()
    → queryContext{packet, src, ifIndex}
        → getIPv4ForInterface(ifIndex)
            → net.InterfaceByIndex(ifIndex)
                → net.Interface.Addrs()
                    → IPv4 address
                        → ResponseBuilder.BuildResponse()
```

### Validation Gates

- ✅ Interface index > 0
- ✅ IPv4 address is valid (4 bytes, not loopback, not unspecified)
- ✅ Interface exists and has IPv4 address
- ✅ Response contains only interface-specific IP (RFC 6762 §15)

---

**Data Model Complete**: 2025-11-06
**Next Phase**: Contracts (API specifications)
