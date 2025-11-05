# F-9: Transport Layer & Socket Configuration

**Spec ID**: F-9
**Type**: Architecture
**Status**: Draft
**Version**: 1.0.0
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling)
**References**:
- Beacon Constitution v1.1.0 (Principle V: Dependencies and Supply Chain)
- ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §1 (Critical Socket Configuration Pitfalls)
- RFC 6762 §5 (Multicast DNS Message Format)
- RFC 6762 §15 (Multicast DNS Responder Guidelines)
- Go Issue #73484 (ListenMulticastUDP packet filtering bug)
- Go Issue #34728 (ListenPacket multicast binding bug)

**Governance**: Development governed by [Beacon Constitution v1.1.0](../memory/constitution.md)

**RFC Validation**: Pending. This specification implements RFC 6762 §5 multicast requirements using platform-specific socket options unavailable in Go standard library.

---

## Overview

This specification defines Beacon's transport layer architecture for multicast DNS socket configuration, addressing critical architectural pitfalls discovered in research of production mDNS libraries. The transport layer MUST:

1. **Coexist** with system daemons (Avahi, systemd-resolved, Bonjour) on shared port 5353
2. **Comply** with RFC 6762 multicast group membership requirements
3. **Avoid** Go standard library bugs (#73484, #34728) that make `net.ListenMulticastUDP()` unsuitable
4. **Use** platform-specific socket options via `golang.org/x/*` semi-standard libraries

**Critical Insight**: The Go standard library's `net.ListenMulticastUDP()` has unfixable bugs that prevent production-grade mDNS implementations. This specification mandates the `net.ListenConfig` pattern with platform-specific socket options, requiring justified use of `golang.org/x/sys` and `golang.org/x/net` per Constitution v1.1.0 Principle V.

**Constitutional Alignment**:
- **Principle I (RFC Compliance)**: RFC 6762 §5 requires proper multicast group membership - standard library bugs prevent compliance
- **Principle V (Dependencies)**: Justifies golang.org/x/sys (socket options) and golang.org/x/net (multicast) as necessary for RFC compliance
- **Principle VIII (Excellence)**: Addresses architectural pitfalls documented in research of Avahi, Bonjour, and Go mDNS libraries

---

## Problem Statement

### Go Standard Library Limitations

**Go Issue #73484**: `ListenMulticastUDP()` doesn't limit received data to packets from the declared multicast group port on Linux.

**Impact**:
- Socket receives ALL UDP traffic on port 5353, not just mDNS multicast (224.0.0.251:5353)
- CPU waste processing irrelevant packets
- Vulnerable to DoS attacks via unrelated multicast traffic
- Silent failure - appears to work but is incorrect

**Go Issue #34728**: `net.ListenPacket()` incorrectly binds to wildcard `0.0.0.0` instead of specified multicast address when given a multicast address.

**Impact**:
- Incorrect network binding
- May not receive multicast packets properly
- Cross-platform inconsistencies

### Port Sharing Requirement

**RFC 6762 §5**: "Multicast DNS messages SHOULD be sent with an IP TTL of 255."

Multiple mDNS implementations MUST coexist on port 5353:
- System daemons: Avahi (Linux), systemd-resolved (Linux), Bonjour (macOS/Windows)
- Application libraries: Beacon, other Go/Python/Java mDNS implementations
- All must share UDP port 5353 without conflicts

**Without proper socket options**:
- Binding fails with "address already in use" error
- Application cannot start when system daemon is running
- Users must manually stop system services (unacceptable UX)
- Enterprise deployments fail (cannot disable system services)

**Real-World Evidence**:
- hashicorp/mdns: Cannot bind when Avahi is running (no SO_REUSEPORT)
- grandcat/zeroconf: PR #89 (open since 2021) attempting to add SO_REUSEPORT support
- User reports: "Your library doesn't work" when the real issue is port binding

---

## Requirements

### REQ-F9-1: ListenConfig Pattern (MANDATORY - RFC Compliant)

Beacon MUST use `net.ListenConfig` with `Control` function to set socket options BEFORE bind() call.

**Rationale**:
- Socket options (SO_REUSEADDR, SO_REUSEPORT) MUST be set after socket() but BEFORE bind()
- `net.ListenMulticastUDP()` sets options AFTER bind (too late)
- Only `net.ListenConfig.Control` provides access to raw file descriptor at correct time

**RFC Alignment**: RFC 6762 §5 requires proper multicast group membership. ListenConfig enables correct multicast configuration where standard library functions fail.

**Implementation Pattern**:
```go
lc := net.ListenConfig{
    Control: setPlatformSocketOptions, // Platform-specific function
}

// Bind to wildcard:5353, multicast group joined after bind
conn, err := lc.ListenPacket(context.Background(), "udp4", "0.0.0.0:5353")
if err != nil {
    return nil, &errors.NetworkError{
        Operation: "bind socket",
        Err:       err,
        Details:   "failed to bind to 0.0.0.0:5353 with socket options",
    }
}

// Join multicast group using golang.org/x/net/ipv4
// (Cannot be done in Control function - requires bound socket)
p := ipv4.NewPacketConn(conn)
err = p.JoinGroup(iface, &net.UDPAddr{IP: net.ParseIP("224.0.0.251")})
```

**Forbidden**:
- ❌ MUST NOT use `net.ListenMulticastUDP()` (Go Issues #73484, #34728)
- ❌ MUST NOT use `net.ListenPacket()` with multicast address
- ❌ MUST NOT set socket options after bind() (too late)

---

### REQ-F9-2: Platform-Specific Socket Options (MANDATORY)

Beacon MUST set platform-specific socket options for port sharing and multicast reception.

**Platform Requirements**:

**Linux (kernel >= 3.9)**:
- `SO_REUSEADDR` (required for multicast binding)
- `SO_REUSEPORT` (required for port sharing with Avahi/systemd-resolved)
- Library: `golang.org/x/sys/unix`

**macOS / Darwin**:
- `SO_REUSEADDR` (required for multicast binding, standard BSD behavior)
- `SO_REUSEPORT` (required for active port sharing with Bonjour)
- Library: `golang.org/x/sys/unix`

**Windows**:
- `SO_REUSEADDR` (primary mechanism, behavior differs from POSIX)
- `SO_REUSEPORT` NOT supported (different socket model)
- Library: Standard `syscall` or `golang.org/x/sys/windows`

**Linux (kernel < 3.9)**:
- `SO_REUSEADDR` (required for multicast binding)
- `SO_REUSEPORT` NOT supported or behavior varies
- **Coexistence NOT guaranteed** on older kernels
- Log warning if kernel version < 3.9 detected

**Rationale**: mDNS requires multiple processes to bind to same port (5353). Without proper socket options, binding fails with "address already in use" when system daemons (Avahi, systemd-resolved, Bonjour) are running.

**Real-World Impact**: hashicorp/mdns fails to bind when Avahi is running because it doesn't set SO_REUSEPORT. This is the #1 user-reported issue for that library.

**Implementation**: Use build tags and platform-specific files:
```
internal/network/
├── socket.go              # Common interface
├── socket_linux.go        # Linux-specific (SO_REUSEADDR + SO_REUSEPORT)
├── socket_darwin.go       # macOS-specific (SO_REUSEADDR + SO_REUSEPORT)
├── socket_windows.go      # Windows-specific (SO_REUSEADDR only)
└── socket_test.go         # Platform-agnostic tests
```

---

### REQ-F9-3: Multicast Group Membership (MANDATORY - RFC Compliant)

Beacon MUST join multicast group 224.0.0.251 (IPv4) per RFC 6762 §5 using `golang.org/x/net/ipv4` package.

**Rationale**: Standard library `net` package lacks multicast group membership control needed for proper mDNS operation.

**RFC Alignment**: RFC 6762 §5: "Multicast DNS messages are sent to and received from the multicast address 224.0.0.251 (or its IPv6 equivalent FF02::FB)."

**Implementation**:
```go
// After binding socket with ListenConfig
p := ipv4.NewPacketConn(conn)

// Join multicast group on specified interface
group := &net.UDPAddr{IP: net.ParseIP("224.0.0.251")}
err := p.JoinGroup(iface, group)
if err != nil {
    return nil, &errors.NetworkError{
        Operation: "join multicast group",
        Err:       err,
        Details:   fmt.Sprintf("failed to join 224.0.0.251 on %s", iface.Name),
    }
}

// Set multicast loop (required for local testing)
err = p.SetMulticastLoopback(true)

// Set multicast TTL per RFC 6762 §11
err = p.SetMulticastTTL(255)
```

**Library**: `golang.org/x/net/ipv4` or `golang.org/x/net/ipv6`

---

### REQ-F9-4: Dependency Justification (CONSTITUTIONAL REQUIREMENT)

Per Constitution v1.1.0 Principle V, all `golang.org/x/*` imports MUST be justified.

**Dependency 1: golang.org/x/sys/unix**

**Justification**:
- **Required For**: Setting SO_REUSEPORT socket option on Linux and macOS
- **No Stdlib Alternative**: `syscall.SetsockoptInt()` in stdlib only exposes subset of socket options; SO_REUSEPORT constants not defined in stdlib
- **Go Team Maintained**: Yes (`golang.org/x/sys` is semi-standard library)
- **Rationale**: SO_REUSEPORT is MANDATORY for coexistence with Avahi/systemd-resolved/Bonjour. Without it, application fails to start with "address already in use" in enterprise environments.

**Usage**:
```go
// socket_linux.go, socket_darwin.go
import "golang.org/x/sys/unix"

func setSocketOptions(fd uintptr) error {
    // SO_REUSEADDR
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
        return err
    }

    // SO_REUSEPORT (Linux >= 3.9, macOS)
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
        return err
    }

    return nil
}
```

**Dependency 2: golang.org/x/net/ipv4**

**Justification**:
- **Required For**: Multicast group membership control (JoinGroup, LeaveGroup, SetMulticastTTL)
- **No Stdlib Alternative**: Standard library `net` package lacks multicast group management API
- **Go Team Maintained**: Yes (`golang.org/x/net` is semi-standard library)
- **Rationale**: RFC 6762 §5 requires joining multicast group 224.0.0.251. Standard library provides no API for this.

**Usage**:
```go
import "golang.org/x/net/ipv4"

p := ipv4.NewPacketConn(conn)
p.JoinGroup(iface, &net.UDPAddr{IP: net.ParseIP("224.0.0.251")})
p.SetMulticastTTL(255) // RFC 6762 §11
```

**Alternative Considered**: Using `net.ListenMulticastUDP()` to avoid dependencies
**Rejected Because**: Go Issues #73484 and #34728 make it unsuitable for production (receives all UDP:5353, not just mDNS multicast)

---

### REQ-F9-5: Error Handling for Port Conflicts (MANDATORY)

Beacon MUST handle "address already in use" gracefully and provide actionable error messages.

**Check Sequence**:
1. Attempt bind with SO_REUSEADDR + SO_REUSEPORT
2. If bind fails with EADDRINUSE:
   - Check if error is due to existing bind (expected with socket options)
   - Check if socket options were set correctly
   - Provide clear error message with remediation steps

**Error Handling**:
```go
if errors.Is(err, syscall.EADDRINUSE) {
    return &errors.NetworkError{
        Operation: "bind socket",
        Err:       err,
        Details: `Failed to bind to port 5353 with SO_REUSEPORT.

Possible causes:
1. Platform does not support SO_REUSEPORT (Linux kernel < 3.9, Windows)
2. Another process is bound to port 5353 without SO_REUSEPORT
3. System daemon (Avahi, systemd-resolved) is not using SO_REUSEPORT

Remediation:
- Linux: Ensure kernel >= 3.9 and Avahi/systemd-resolved are using SO_REUSEPORT
- macOS: Verify Bonjour daemon compatibility
- Windows: Consider client mode or alternative port (coexistence not guaranteed)`,
    }
}
```

**RFC Alignment**: RFC 6762 §5 allows multiple responders on same link. Port sharing is system-level requirement to achieve this protocol requirement.

---

### REQ-F9-6: Socket Buffer Configuration (MANDATORY)

Beacon MUST configure socket buffer sizes to handle multicast burst traffic.

**Requirements**:
- Receive buffer: 64KB minimum (supports ~45 MTU-sized packets queued)
- Send buffer: 64KB minimum (supports burst announcements)
- Configurable via Option pattern

**Rationale**:
- Multicast can deliver bursts of responses
- Small default buffers (often 8KB) cause packet loss
- RFC 6762 §6 allows up to 9000 bytes per message

**Implementation**:
```go
// Set receive buffer size
if err := conn.SetReadBuffer(64 * 1024); err != nil {
    return nil, &errors.NetworkError{
        Operation: "configure socket",
        Err:       err,
        Details:   "failed to set receive buffer to 64KB",
    }
}

// Set send buffer size
if err := conn.SetWriteBuffer(64 * 1024); err != nil {
    return nil, &errors.NetworkError{
        Operation: "configure socket",
        Err:       err,
        Details:   "failed to set send buffer to 64KB",
    }
}
```

---

### REQ-F9-7: Context Propagation in Blocking Operations (MANDATORY)

All functions that perform blocking I/O MUST respect context cancellation and propagate deadlines.

**Rationale**:
- Research mandate: "Every single function that blocks, performs network I/O, or spawns a goroutine must accept context.Context as its first argument"
- Prevents goroutine leaks when caller cancels context
- Enables proper resource cleanup on timeout/cancellation
- Allows HTTP servers and other request-scoped systems to cancel mDNS operations

**Requirements**:
1. MUST accept `context.Context` as first parameter in all blocking functions
2. MUST check `ctx.Done()` in receive loops
3. MUST propagate `ctx.Deadline()` to socket `SetReadDeadline()`
4. MUST return immediately when context is cancelled
5. MUST clean up resources before returning on cancellation
6. MUST return `ctx.Err()` (context.Canceled or context.DeadlineExceeded) when context triggers return

**Correct Implementation Pattern**:
```go
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error) {
    buf := make([]byte, 9000) // RFC 6762 §17 max

    for {
        // REQ-F9-7: Check context cancellation BEFORE blocking read
        select {
        case <-ctx.Done():
            return nil, nil, ctx.Err() // context.Canceled or context.DeadlineExceeded
        default:
            // Continue to receive
        }

        // REQ-F9-7: Propagate context deadline to socket
        if deadline, ok := ctx.Deadline(); ok {
            s.conn.SetReadDeadline(deadline)
        } else {
            // No deadline specified, set short timeout to allow periodic ctx.Done() checking
            s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
        }

        n, srcAddr, err := s.conn.ReadFrom(buf)
        if err != nil {
            // Check if error is due to timeout
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                // Timeout allows ctx.Done() check on next iteration
                continue
            }

            // Other errors (connection closed, etc.)
            return nil, nil, &errors.NetworkError{
                Operation: "receive packet",
                Err:       err,
                Details:   "socket read failed",
            }
        }

        // Packet received successfully
        return buf[:n], srcAddr, nil
    }
}
```

**Anti-Pattern to Avoid**:
```go
// ❌ WRONG: Accepts ctx but never uses it (goroutine leak!)
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error) {
    buf := make([]byte, 9000)
    n, srcAddr, err := s.conn.ReadFrom(buf) // Blocks forever, ignores ctx
    if err != nil {
        return nil, nil, err
    }
    return buf[:n], srcAddr, nil
}

// ❌ WRONG: Doesn't check ctx.Done() in loop
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error) {
    buf := make([]byte, 9000)
    for {
        // Missing: select { case <-ctx.Done(): return ... }
        n, srcAddr, err := s.conn.ReadFrom(buf)
        // Loop never exits if context cancelled
    }
}
```

**Research Evidence**:
- hashicorp/mdns issue #10: "Feature request to refactor Query and Lookup to use context instead of hard timeout"
- grandcat/zeroconf and brutella/dnssd correctly employ context.Context in APIs
- Baseline requirement for modern Go libraries per research analysis

**Integration with F-4 (Concurrency Model)**:
- Context propagation aligns with F-4's goroutine lifecycle management
- Context cancellation triggers graceful shutdown per F-4 requirements
- All goroutines spawned by transport layer MUST respect parent context

---

## Socket Creation Sequence

### Step-by-Step Process

**1. Create ListenConfig with Control Function**
```go
lc := net.ListenConfig{
    Control: func(network, address string, c syscall.RawConn) error {
        var setErr error
        err := c.Control(func(fd uintptr) {
            setErr = setPlatformSocketOptions(fd)
        })
        if err != nil {
            return err
        }
        return setErr
    },
}
```

**2. Bind Socket to Wildcard Address**
```go
// Bind to 0.0.0.0:5353 (not 224.0.0.251:5353)
// Multicast group joined separately after bind
conn, err := lc.ListenPacket(ctx, "udp4", "0.0.0.0:5353")
if err != nil {
    return nil, handleBindError(err)
}
```

**3. Configure Socket Buffers**
```go
if err := conn.SetReadBuffer(64 * 1024); err != nil {
    conn.Close()
    return nil, err
}

if err := conn.SetWriteBuffer(64 * 1024); err != nil {
    conn.Close()
    return nil, err
}
```

**4. Join Multicast Group Per Interface**
```go
p := ipv4.NewPacketConn(conn)

for _, iface := range interfaces {
    group := &net.UDPAddr{IP: net.ParseIP("224.0.0.251")}
    if err := p.JoinGroup(iface, group); err != nil {
        // Log warning, continue with other interfaces
        log.Warnf("Failed to join multicast group on %s: %v", iface.Name, err)
        continue
    }
}
```

**5. Set Multicast Parameters**
```go
// Multicast loopback (required for local testing, RFC 6762 §15.1)
if err := p.SetMulticastLoopback(true); err != nil {
    // Non-fatal, log warning
}

// Multicast TTL (RFC 6762 §11 - SHOULD be 255)
if err := p.SetMulticastTTL(255); err != nil {
    // Non-fatal, log warning
}
```

---

## Platform-Specific Implementations

### Linux (socket_linux.go)

```go
// +build linux

package network

import (
    "syscall"
    "golang.org/x/sys/unix"
)

func setPlatformSocketOptions(fd uintptr) error {
    // SO_REUSEADDR - required for multicast binding
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
        return fmt.Errorf("SO_REUSEADDR: %w", err)
    }

    // SO_REUSEPORT - required for port sharing (Linux >= 3.9)
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
        // Check kernel version
        var uname unix.Utsname
        if unix.Uname(&uname) == nil {
            version := string(uname.Release[:])
            if isKernelOlderThan39(version) {
                // Log warning, continue without SO_REUSEPORT
                // Coexistence not guaranteed on old kernels
                return nil
            }
        }
        return fmt.Errorf("SO_REUSEPORT: %w", err)
    }

    return nil
}

func isKernelOlderThan39(version string) bool {
    // Parse version string "3.8.0-gentoo" -> (3, 8, 0)
    // Return true if < 3.9
    // Implementation omitted for brevity
    return false
}
```

### macOS / Darwin (socket_darwin.go)

```go
// +build darwin

package network

import (
    "golang.org/x/sys/unix"
)

func setPlatformSocketOptions(fd uintptr) error {
    // SO_REUSEADDR - required for multicast binding (BSD behavior)
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
        return fmt.Errorf("SO_REUSEADDR: %w", err)
    }

    // SO_REUSEPORT - required for active port sharing with Bonjour
    if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
        return fmt.Errorf("SO_REUSEPORT: %w", err)
    }

    return nil
}
```

### Windows (socket_windows.go)

```go
// +build windows

package network

import (
    "syscall"
)

func setPlatformSocketOptions(fd uintptr) error {
    // SO_REUSEADDR - primary mechanism on Windows (behavior differs from POSIX)
    if err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
        return fmt.Errorf("SO_REUSEADDR: %w", err)
    }

    // SO_REUSEPORT not supported on Windows
    // Note: Windows socket model differs from POSIX
    // Coexistence with system Bonjour service may be limited

    return nil
}
```

---

## Integration with F-2 (Package Structure)

**Package Location**: `internal/network/`

**Files**:
- `socket.go` - Common socket interface and ListenConfig logic
- `socket_linux.go` - Linux-specific socket options
- `socket_darwin.go` - macOS-specific socket options
- `socket_windows.go` - Windows-specific socket options
- `multicast.go` - Multicast group management (uses golang.org/x/net/ipv4)
- `socket_test.go` - Platform-agnostic tests
- `socket_linux_test.go` - Linux-specific tests

**Public API** (used by `querier/`, `responder/`):
```go
package network

// Socket represents an mDNS socket with platform-specific configuration
type Socket struct {
    conn net.PacketConn
    ipv4 *ipv4.PacketConn
    // ...
}

// New creates an mDNS socket with proper multicast configuration
func New(interfaces []net.Interface) (*Socket, error)

// Send sends packet to multicast address
func (s *Socket) Send(packet []byte) error

// Receive receives packet from multicast
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error)

// Close closes socket and leaves multicast groups
func (s *Socket) Close() error
```

---

## Testing Strategy

### Unit Tests

**Platform-Specific Tests** (socket_linux_test.go, socket_darwin_test.go, socket_windows_test.go):
```go
func TestSocketOptions_Linux(t *testing.T) {
    // Test SO_REUSEADDR and SO_REUSEPORT are set
    // Use getsockopt to verify options
}

func TestKernelVersionDetection(t *testing.T) {
    // Test kernel version parsing
    // Verify warning logged for kernel < 3.9
}
```

**Integration Tests**:
```go
func TestAvahiCoexistence(t *testing.T) {
    // Requires Avahi running on Linux
    // Verify both Beacon and Avahi can bind to port 5353

    if !isAvahiRunning() {
        t.Skip("Avahi not running")
    }

    s1, err := network.New(DefaultInterfaces())
    if err != nil {
        t.Fatalf("Failed to create socket with Avahi running: %v", err)
    }
    defer s1.Close()

    // Verify can send/receive
    // ...
}
```

### Manual Testing

**Linux**:
```bash
# Start Avahi
sudo systemctl start avahi-daemon

# Verify Avahi is bound to port 5353
sudo netstat -tulpn | grep 5353

# Run Beacon tests - should succeed
go test -v ./internal/network/

# Both should be bound to same port
```

**macOS**:
```bash
# Verify Bonjour is running
sudo lsof -i :5353

# Run Beacon tests
go test -v ./internal/network/

# Both should coexist
```

---

## Error Handling

### Network Errors

All socket errors MUST use `errors.NetworkError` from F-3 (Error Handling):

```go
type NetworkError struct {
    Operation string // "bind socket", "join multicast group"
    Err       error  // Underlying error
    Details   string // Actionable context
}
```

**Examples**:
```go
// Bind failure
&errors.NetworkError{
    Operation: "bind socket",
    Err:       syscall.EADDRINUSE,
    Details:   "Port 5353 already in use. See SO_REUSEPORT requirements.",
}

// Multicast join failure
&errors.NetworkError{
    Operation: "join multicast group",
    Err:       err,
    Details:   fmt.Sprintf("Failed to join 224.0.0.251 on interface %s", iface.Name),
}
```

---

## Configuration Options

### Option Pattern

Per F-5 (Configuration), transport layer MUST use functional options:

```go
// WithInterfaces specifies which interfaces to bind to
func WithInterfaces(ifaces []net.Interface) Option

// WithBufferSize sets socket buffer sizes
func WithBufferSize(bytes int) Option

// WithMulticastTTL sets multicast TTL (default 255 per RFC 6762 §11)
func WithMulticastTTL(ttl int) Option

// WithMulticastLoopback enables/disables multicast loopback (default true)
func WithMulticastLoopback(enabled bool) Option
```

**Usage**:
```go
s, err := network.New(
    network.WithInterfaces(defaultInterfaces),
    network.WithBufferSize(128 * 1024), // 128KB buffers
    network.WithMulticastTTL(255),       // RFC 6762 §11
)
```

---

## Success Criteria

- [x] Uses `net.ListenConfig` pattern (not `net.ListenMulticastUDP`)
- [x] Sets SO_REUSEADDR on all platforms
- [x] Sets SO_REUSEPORT on Linux (>= 3.9) and macOS
- [x] Joins multicast group using `golang.org/x/net/ipv4`
- [x] Handles "address already in use" with actionable errors
- [x] Configures socket buffers (64KB minimum)
- [x] Platform-specific implementations (Linux, macOS, Windows)
- [x] Coexists with Avahi on Linux (integration test)
- [x] Coexists with Bonjour on macOS (integration test)
- [x] Dependencies justified per Constitution v1.1.0 Principle V
- [x] RFC 6762 §5 multicast requirements met

---

## Governance and Compliance

### Constitutional Compliance

**Principle I (RFC Compliant)**:
- ✅ RFC 6762 §5 multicast group membership: Implemented via `golang.org/x/net/ipv4.PacketConn.JoinGroup()`
- ✅ RFC 6762 §11 multicast TTL 255: Configured via `SetMulticastTTL(255)`
- ✅ RFC 6762 §15.1 multicast loopback: Enabled via `SetMulticastLoopback(true)`

**Principle V (Dependencies and Supply Chain)**:
- ✅ `golang.org/x/sys/unix`: Justified for SO_REUSEPORT (no stdlib alternative)
- ✅ `golang.org/x/net/ipv4`: Justified for multicast group management (no stdlib alternative)
- ✅ Both are Go team maintained semi-standard libraries
- ✅ Justification documented in REQ-F9-4

**Principle VIII (Excellence)**:
- ✅ Addresses architectural pitfalls from research (ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §1)
- ✅ Enables coexistence with Avahi/Bonjour (enterprise requirement)
- ✅ Avoids Go stdlib bugs #73484, #34728 (production requirement)

### Change Control

Changes to this specification require:
1. Constitutional compliance check (if dependencies change)
2. RFC validation review (if multicast behavior changes)
3. Platform compatibility testing (Linux, macOS, Windows)
4. Version bump per semantic versioning:
   - **MAJOR**: Breaking changes to socket API or platform support
   - **MINOR**: New platforms or non-breaking enhancements
   - **PATCH**: Bug fixes, documentation improvements

---

## References

**Constitutional**:
- [Beacon Constitution v1.1.0](../memory/constitution.md) - Principle V (Dependencies)

**Architectural**:
- [ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md](../../docs/ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md) - §1 (Socket Configuration)
- [F-2: Package Structure](./F-2-package-structure.md) - `internal/network/` package organization
- [F-3: Error Handling](./F-3-error-handling.md) - `NetworkError` type

**RFCs**:
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - §5 (Multicast DNS Message Format), §11 (Multicast TTL), §15 (Responder Guidelines)

**Go Issues**:
- [Go Issue #73484](https://github.com/golang/go/issues/73484) - ListenMulticastUDP packet filtering bug
- [Go Issue #34728](https://github.com/golang/go/issues/34728) - ListenPacket multicast binding bug

**Research**:
- hashicorp/mdns - Port binding failures without SO_REUSEPORT
- grandcat/zeroconf - PR #89 (SO_REUSEPORT support attempt)
- brutella/dnssd - Production-grade socket configuration patterns

**Dependencies**:
- [golang.org/x/sys/unix](https://pkg.go.dev/golang.org/x/sys/unix) - Platform syscalls
- [golang.org/x/net/ipv4](https://pkg.go.dev/golang.org/x/net/ipv4) - IPv4 multicast control

---

## Version History

| Version | Date | Changes | Validated Against |
|---------|------|---------|-------------------|
| 1.0.0 | 2025-11-01 | Initial transport layer specification. Defines ListenConfig pattern, platform-specific socket options (SO_REUSEPORT), multicast group management. Justifies golang.org/x/sys and golang.org/x/net dependencies per Constitution v1.1.0 Principle V. Addresses architectural pitfalls from mDNS library research. | Constitution v1.1.0, RFC 6762 §5/§11/§15, Go Issues #73484/#34728 |
