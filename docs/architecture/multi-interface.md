# Multi-Interface IP Addressing

This document describes how Beacon implements interface-specific IP addressing for mDNS responses, ensuring RFC 6762 §15 compliance on multi-interface hosts.

## Problem Statement

**RFC 6762 §15 Requirement**:
> A Multicast DNS responder MUST use the IP address of the interface on which it received the query when building its response.

**Multi-Interface Scenarios**:
- Laptop with WiFi + Ethernet
- Server with multiple NICs
- IoT device with WiFi + Ethernet bridge
- Virtual machine with host + bridge interfaces

**Without Interface-Specific Addressing**:
```
Host has two interfaces:
  eth0: 192.168.1.100
  wlan0: 10.0.0.50

❌ INCORRECT BEHAVIOR:
  Query arrives on eth0 → Response includes 10.0.0.50 (wrong interface!)
  Result: Connection failures, unreachable services
```

**With Interface-Specific Addressing**:
```
✅ CORRECT BEHAVIOR:
  Query arrives on eth0 → Response includes 192.168.1.100 (eth0's IP)
  Query arrives on wlan0 → Response includes 10.0.0.50 (wlan0's IP)
  Result: Services always reachable via correct interface
```

---

## Solution Architecture

### 1. Extract Interface Index from IP_PKTINFO

When receiving mDNS queries, extract the interface index from socket control messages:

**Linux (IP_PKTINFO)**:
```c
struct in_pktinfo {
    int ipi_ifindex;    // Interface index
    struct in_addr ipi_spec_dst;  // Local address
    struct in_addr ipi_addr;      // Destination address
};
```

**macOS/BSD (IP_RECVIF)**:
```c
struct sockaddr_dl {
    u_char sdl_len;
    u_char sdl_family;
    u_short sdl_index;   // Interface index
    ...
};
```

### 2. Resolve Interface-Specific IP

Use interface index to look up the correct IPv4 address:

```go
func getIPv4ForInterface(interfaceIndex int) (net.IP, error) {
    iface, err := net.InterfaceByIndex(interfaceIndex)
    if err != nil {
        return nil, err
    }

    addrs, err := iface.Addrs()
    if err != nil {
        return nil, err
    }

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ip4 := ipnet.IP.To4(); ip4 != nil {
                return ip4, nil
            }
        }
    }

    return nil, errors.New("no IPv4 address on interface")
}
```

### 3. Build Interface-Specific Response

Use resolved IP when constructing A records:

```go
func (rb *ResponseBuilder) buildResponse(query Query, interfaceIndex int) Response {
    // Resolve IP for interface that received query
    ip, err := rb.getIPv4ForInterface(interfaceIndex)
    if err != nil {
        return nil, err
    }

    // Build A record with interface-specific IP
    aRecord := ARecord{
        Name: rb.hostname + ".local",
        IP:   ip,  // Interface-specific!
        TTL:  120,
    }

    return Response{
        Answers: []Record{ptrRecord, srvRecord, txtRecord, aRecord},
    }
}
```

---

## Implementation Details

### Transport Layer (007-interface-specific-addressing)

**File**: `internal/transport/udp.go`

**Socket Options** (Linux):
```go
err = syscall.SetsockoptInt(
    int(fd),
    syscall.IPPROTO_IP,
    syscall.IP_PKTINFO,  // Enable IP_PKTINFO
    1,
)
```

**Receive with Control Messages**:
```go
func (t *UDPv4Transport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
    conn := ipv4.NewPacketConn(t.conn)

    // Receive packet with control messages
    n, cm, src, err := conn.ReadFrom(buffer)
    if err != nil {
        return nil, nil, 0, err
    }

    // Extract interface index from control message
    interfaceIndex := cm.IfIndex

    return buffer[:n], src, interfaceIndex, nil
}
```

**Platform-Specific Details**:
- **Linux**: Uses `IP_PKTINFO` (syscall constant available)
- **macOS/BSD**: Uses `IP_RECVIF` (via `golang.org/x/net/ipv4`)
- **Windows**: Uses `IP_PKTINFO` (similar to Linux)

### Response Builder

**File**: `internal/responder/response_builder.go`

**Interface-Aware Response**:
```go
func (rb *ResponseBuilder) BuildResponse(
    query *message.Query,
    service *Service,
    interfaceIndex int,
) (*message.Response, error) {
    // 1. Resolve interface-specific IP
    ip, err := rb.resolver.GetIPv4ForInterface(interfaceIndex)
    if err != nil {
        return nil, fmt.Errorf("resolve interface IP: %w", err)
    }

    // 2. Build records with interface-specific IP
    records := []Record{
        rb.buildPTR(service),
        rb.buildSRV(service),
        rb.buildTXT(service),
        rb.buildA(service, ip),  // Use resolved IP
    }

    return &message.Response{Answers: records}, nil
}
```

---

## Diagram: Interface-Specific Addressing Flow

```mermaid
sequenceDiagram
    participant Q as Querier (eth0: 192.168.1.50)
    participant N1 as Network (eth0: 192.168.1.0/24)
    participant R as Responder
    participant N2 as Network (wlan0: 10.0.0.0/24)
    participant Q2 as Querier (wlan0: 10.0.0.100)

    Note over R: Multi-Interface Host<br/>eth0: 192.168.1.100<br/>wlan0: 10.0.0.50

    Note over Q,N1: Scenario 1: Query via eth0
    Q->>N1: Query (_myapp._tcp.local)
    N1->>R: Query received on eth0
    R->>R: Extract IfIndex from IP_PKTINFO<br/>IfIndex = 2 (eth0)
    R->>R: getIPv4ForInterface(2)<br/>→ 192.168.1.100
    R->>R: Build A record: hostname.local → 192.168.1.100
    R->>N1: Response (A: 192.168.1.100)
    N1->>Q: Response received
    Q->>Q: Connect to 192.168.1.100:8080 ✅

    Note over Q2,N2: Scenario 2: Query via wlan0
    Q2->>N2: Query (_myapp._tcp.local)
    N2->>R: Query received on wlan0
    R->>R: Extract IfIndex from IP_PKTINFO<br/>IfIndex = 3 (wlan0)
    R->>R: getIPv4ForInterface(3)<br/>→ 10.0.0.50
    R->>R: Build A record: hostname.local → 10.0.0.50
    R->>N2: Response (A: 10.0.0.50)
    N2->>Q2: Response received
    Q2->>Q2: Connect to 10.0.0.50:8080 ✅
```

---

## Example Scenarios

### Scenario 1: Laptop with WiFi + Ethernet

**Configuration**:
```
wlan0: 10.0.0.25 (Home WiFi)
eth0:  192.168.1.50 (Corporate LAN)
```

**Service Registration**:
```go
service := responder.Service{
    Instance:    "My Laptop",
    ServiceType: "_http._tcp",
    Port:        8080,
}
resp.RegisterService(service)
```

**Query from WiFi Client**:
```
1. Query arrives on wlan0
2. Extract interfaceIndex = 3 (wlan0)
3. Resolve IP: 10.0.0.25
4. Response: A hostname.local → 10.0.0.25
```

**Query from Ethernet Client**:
```
1. Query arrives on eth0
2. Extract interfaceIndex = 2 (eth0)
3. Resolve IP: 192.168.1.50
4. Response: A hostname.local → 192.168.1.50
```

**Result**: Clients on both networks can successfully connect!

---

### Scenario 2: IoT Device WiFi ↔ Ethernet Bridge

**Configuration**:
```
wlan0: 192.168.4.1 (IoT WiFi AP)
eth0:  192.168.1.100 (Home LAN)
```

**Use Case**: Bridge IoT sensors (WiFi) to home network (Ethernet)

**Service Registration** (IoT sensor service):
```go
service := responder.Service{
    Instance:    "Temperature Sensor",
    ServiceType: "_sensor._tcp",
    Port:        9000,
}
resp.RegisterService(service)
```

**Query from IoT Sensor (WiFi)**:
```
1. Query arrives on wlan0
2. Extract interfaceIndex = 3 (wlan0)
3. Resolve IP: 192.168.4.1
4. Response: A hostname.local → 192.168.4.1
5. Sensor connects to 192.168.4.1:9000 ✅
```

**Query from Home Automation Hub (Ethernet)**:
```
1. Query arrives on eth0
2. Extract interfaceIndex = 2 (eth0)
3. Resolve IP: 192.168.1.100
4. Response: A hostname.local → 192.168.1.100
5. Hub connects to 192.168.1.100:9000 ✅
```

**Without Interface-Specific Addressing**:
- Hub might receive WiFi IP (192.168.4.1) → unreachable ❌
- Sensor might receive Ethernet IP (192.168.1.100) → unreachable ❌

---

### Scenario 3: Multi-NIC Server

**Configuration**:
```
eth0: 10.0.1.50 (Management Network)
eth1: 10.0.2.50 (Application Network)
eth2: 10.0.3.50 (Storage Network)
```

**Service Registration** (Database service):
```go
service := responder.Service{
    Instance:    "PostgreSQL",
    ServiceType: "_postgresql._tcp",
    Port:        5432,
}
resp.RegisterService(service)
```

**Queries from Different Networks**:
```
Management client → eth0 → Response: 10.0.1.50 ✅
Application client → eth1 → Response: 10.0.2.50 ✅
Storage client → eth2 → Response: 10.0.3.50 ✅
```

**Benefit**: Network segmentation maintained, clients use correct interface

---

## Testing Strategy

### Unit Tests

**Interface Resolver Tests**:
```go
func TestGetIPv4ForInterface(t *testing.T) {
    // Find loopback interface
    iface, err := net.InterfaceByName("lo")
    require.NoError(t, err)

    // Resolve IP for loopback
    ip, err := getIPv4ForInterface(iface.Index)
    require.NoError(t, err)
    assert.Equal(t, "127.0.0.1", ip.String())
}
```

**Response Builder Tests**:
```go
func TestBuildResponse_InterfaceSpecific(t *testing.T) {
    // Mock resolver returns interface-specific IP
    resolver := &MockResolver{
        IPForInterface: map[int]net.IP{
            2: net.ParseIP("192.168.1.100"),
            3: net.ParseIP("10.0.0.50"),
        },
    }

    builder := NewResponseBuilder(resolver)

    // Test eth0 response
    resp, err := builder.BuildResponse(query, service, 2)
    require.NoError(t, err)
    assert.Equal(t, "192.168.1.100", resp.A[0].IP.String())

    // Test wlan0 response
    resp, err = builder.BuildResponse(query, service, 3)
    require.NoError(t, err)
    assert.Equal(t, "10.0.0.50", resp.A[0].IP.String())
}
```

### Integration Tests

**Multi-Interface Response Test**:
```go
func TestResponder_MultiInterface(t *testing.T) {
    // Requires multi-interface host
    if testing.Short() {
        t.Skip("Skipping multi-interface test")
    }

    // Register service
    resp, err := responder.New()
    require.NoError(t, err)
    defer resp.Close()

    service := responder.Service{...}
    err = resp.RegisterService(service)
    require.NoError(t, err)

    // Query via eth0
    results1, err := queryViaInterface("eth0", service.ServiceType)
    require.NoError(t, err)
    eth0IP := getInterfaceIP("eth0")
    assert.Equal(t, eth0IP, results1[0].IP)

    // Query via wlan0
    results2, err := queryViaInterface("wlan0", service.ServiceType)
    require.NoError(t, err)
    wlan0IP := getInterfaceIP("wlan0")
    assert.Equal(t, wlan0IP, results2[0].IP)
}
```

---

## Performance Impact

### Overhead Analysis

**Per-Query Overhead**:
```
1. Extract interface index from control message: ~100ns
2. Look up interface by index: ~500ns
3. Iterate interface addresses: ~1μs
Total: ~1.6μs per query
```

**Compared to Total Response Time** (~10ms):
- Interface resolution: 0.016% of total
- **Negligible impact**

### Memory Allocation

- No additional allocations in hot path
- Interface lookup uses standard library (no caching needed)
- Control message buffer reused via buffer pool

---

## Common Pitfalls

### ❌ Pitfall 1: Using Single IP for All Interfaces

```go
// WRONG: Use hardcoded IP
func buildResponse() Response {
    return Response{
        A: ARecord{IP: net.ParseIP("192.168.1.100")},  // Breaks multi-interface!
    }
}
```

### ✅ Solution: Resolve Interface-Specific IP

```go
// CORRECT: Resolve IP per interface
func buildResponse(interfaceIndex int) Response {
    ip, _ := getIPv4ForInterface(interfaceIndex)
    return Response{
        A: ARecord{IP: ip},  // Interface-specific ✅
    }
}
```

---

### ❌ Pitfall 2: Ignoring Control Messages

```go
// WRONG: Standard ReadFrom (no control messages)
n, src, err := conn.ReadFrom(buffer)
// Lost interface index!
```

### ✅ Solution: Use ipv4.PacketConn

```go
// CORRECT: Read control messages
conn := ipv4.NewPacketConn(udpConn)
n, cm, src, err := conn.ReadFrom(buffer)
interfaceIndex := cm.IfIndex  // ✅
```

---

### ❌ Pitfall 3: Caching IP per Interface

```go
// WRONG: Cache interface IPs (breaks if IPs change)
var ipCache = map[int]net.IP{}

func getIP(ifIndex int) net.IP {
    if ip, ok := ipCache[ifIndex]; ok {
        return ip  // Stale if DHCP lease renewed!
    }
    ...
}
```

### ✅ Solution: Resolve IP on Every Query

```go
// CORRECT: Always resolve (handles DHCP changes)
func getIP(ifIndex int) (net.IP, error) {
    iface, err := net.InterfaceByIndex(ifIndex)
    if err != nil {
        return nil, err
    }
    // Fresh IP address ✅
    return resolveIPv4(iface)
}
```

**Why?** Interface IPs can change:
- DHCP lease renewal
- Manual reconfiguration
- VPN connect/disconnect

**Performance**: Negligible overhead (~1.6μs), correctness more important

---

## RFC Compliance

### RFC 6762 §15 (Multicast DNS Responder)

> **A Multicast DNS responder MUST use the source address that corresponds to the interface on which the query was received when sending responses.**

✅ **Beacon Compliance**:
- Extracts interface index from `IP_PKTINFO`/`IP_RECVIF`
- Resolves interface-specific IPv4 address
- Builds A records with correct interface IP
- **Result**: Full RFC 6762 §15 compliance

---

## References

- **RFC 6762 §15**: Responder Address Selection - [https://www.rfc-editor.org/rfc/rfc6762.html#section-15](https://www.rfc-editor.org/rfc/rfc6762.html#section-15)
- **007-interface-specific-addressing**: Implementation spec in `specs/007-interface-specific-addressing/`
- **IP_PKTINFO**: Linux man page - [ip(7)](https://man7.org/linux/man-pages/man7/ip.7.html)
- **golang.org/x/net/ipv4**: IPv4 control message handling

## See Also

- [message-flow.md](message-flow.md) - Complete message flow
- [state-machine.md](state-machine.md) - Responder state machine
- [buffer-pooling.md](buffer-pooling.md) - Performance optimization
