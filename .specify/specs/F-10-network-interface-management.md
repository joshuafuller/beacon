# F-10: Network Interface Management

**Spec ID**: F-10
**Type**: Architecture
**Status**: Draft
**Version**: 1.0.0
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling), F-5 (Configuration), F-9 (Transport Layer)
**References**:
- Beacon Constitution v1.1.0 (Principle I: RFC Compliance, Principle VIII: Excellence)
- ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §2 (Network Interface Management Pitfalls)
- RFC 6762 §2 (Multicast DNS Scope - Link-Local)
- RFC 6762 §5 (Multicast DNS Message Format - Interface-Specific)

**Governance**: Development governed by [Beacon Constitution v1.1.0](../memory/constitution.md)

**RFC Validation**: Pending. This specification implements RFC 6762 §2 link-local scope enforcement through interface selection, preventing protocol violations via VPN/virtual interface leakage.

---

## Overview

This specification defines Beacon's network interface management architecture, addressing critical privacy and functionality issues discovered in production mDNS libraries. The interface management layer MUST:

1. **Provide explicit control** over which network interfaces are used for mDNS traffic
2. **Exclude problematic interfaces** by default (VPN, Docker, virtual tunnels)
3. **Enforce RFC 6762 §2 link-local scope** through interface filtering
4. **Enable future network change detection** through proper architecture

**Critical Insight**: Implicit "bind to all interfaces" behavior causes privacy violations (queries leak to VPN providers), functionality bugs (discover services on wrong network), and RFC violations (VPN traffic is not link-local). This specification mandates explicit interface management with secure defaults.

**Constitutional Alignment**:
- **Principle I (RFC Compliance)**: RFC 6762 §2 specifies link-local scope - VPN/virtual interfaces violate this requirement
- **Principle VIII (Excellence)**: Addresses interface management pitfalls documented in research of hashicorp/mdns and other libraries

---

## Problem Statement

### Implicit Interface Binding

Many mDNS libraries bind to "all interfaces" implicitly and never re-evaluate when network changes occur:

**Failure Modes**:
1. **VPN leakage**: Queries sent over VPN tunnel to VPN provider (privacy violation)
2. **Docker confusion**: Queries sent into container networks (isolated, wrong segment)
3. **Wrong network**: Discover services on VPN instead of local LAN (functionality bug)
4. **Stale interfaces**: Interface goes down, socket holds stale FD, blocks indefinitely
5. **RFC violation**: RFC 6762 §2 specifies link-local scope; VPN is NOT link-local

**Real-World Evidence**:
- hashicorp/mdns Issue #122: "mDNS discovery binds to wrong interface"
- hashicorp/mdns Issue #35: "Failed to bind to udp6 port" (interface selection problem)
- User reports: "Can't find printer when VPN is connected"
- Security audits: "mDNS traffic visible in VPN tunnel"

### Dynamic Network Topologies

Modern systems have complex, changing network environments:
- **Laptops**: Wi-Fi ↔ Ethernet switching
- **Mobile**: Cellular ↔ Wi-Fi handoffs
- **Containers**: Ephemeral virtual interfaces (docker0, veth*)
- **VPNs**: Virtual tunnels that should be excluded (utun*, ppp*, tun*)
- **Enterprise**: Multiple VLANs, complex routing

Without explicit interface management:
- Library becomes non-functional after network changes
- Resource leaks (stale sockets)
- Security issues (traffic on wrong interface)
- Poor UX (requires application restart)

---

## Requirements

### REQ-F10-1: Explicit Interface API (MANDATORY - RFC Compliant)

Beacon MUST provide explicit API for network interface selection via functional options pattern.

**Rationale**: Implicit "bind to all interfaces" fails in dynamic network environments and violates RFC 6762 §2 link-local scope when VPN/virtual interfaces are included.

**RFC Alignment**: RFC 6762 §2 "Multicast DNS is restricted to link-local scope." VPN and virtual tunnel interfaces are NOT link-local, so exclusion is protocol-compliant behavior.

**API Design**:
```go
package querier

// WithInterfaces specifies explicit list of interfaces to use
func WithInterfaces(ifaces []net.Interface) Option

// WithInterfaceFilter provides custom interface filtering logic
func WithInterfaceFilter(filter func(net.Interface) bool) Option

// DefaultInterfaces returns all multicast-capable physical interfaces
// (excludes loopback, VPN, Docker, down interfaces)
func DefaultInterfaces() ([]net.Interface, error)
```

**Usage Examples**:
```go
// Example 1: Default behavior (safe, excludes VPN/Docker)
q, err := querier.New()
// Uses DefaultInterfaces() internally

// Example 2: Explicit interface selection (advanced users)
wifiIface, _ := net.InterfaceByName("wlan0")
q, err := querier.New(querier.WithInterfaces([]net.Interface{wifiIface}))

// Example 3: Custom filtering (power users)
q, err := querier.New(querier.WithInterfaceFilter(func(iface net.Interface) bool {
    // Only Wi-Fi and Ethernet, exclude everything else
    return strings.HasPrefix(iface.Name, "wlan") ||
           strings.HasPrefix(iface.Name, "eth")
}))

// Example 4: Explicit override to INCLUDE VPN (testing, specific configs)
allIfaces, _ := net.Interfaces()
q, err := querier.New(querier.WithInterfaces(allIfaces))
```

---

### REQ-F10-2: Default Interface Filtering Logic (MANDATORY - Privacy & Security)

`DefaultInterfaces()` MUST implement intelligent filtering to avoid problematic interfaces.

**Default Filter Requirements**:
1. MUST only include interfaces with `net.FlagUp` set (interface is active)
2. MUST only include interfaces with `net.FlagMulticast` set (supports multicast)
3. MUST exclude loopback interfaces (unless explicitly requested)
4. MUST exclude virtual/tunnel interfaces by default:
   - **VPN interfaces**: utun*, ppp*, tun*, tap* (platform-specific)
   - **Container networks**: docker0, veth*, br-* (Docker bridge/veth)
   - **Virtual tunnels**: gre*, ipip*, sit* (tunnel interfaces)
5. MUST validate interface still exists and is valid before using
6. MUST support platform-specific naming conventions

**Rationale**:
- **Privacy**: VPN interfaces leak mDNS queries to VPN provider (wrong network segment)
- **Functionality**: Docker interfaces are isolated network namespaces (wrong segment)
- **Security**: Virtual tunnels may cross network boundaries (RFC 6762 §2 violation)
- **RFC Compliance**: RFC 6762 §2 "link-local scope" excludes non-local interfaces

**But**: Some users may want to bind to these (testing, specific network configs), so filtering should be default behavior with explicit override option.

**Implementation**:
```go
func DefaultInterfaces() ([]net.Interface, error) {
    allIfaces, err := net.Interfaces()
    if err != nil {
        return nil, &errors.NetworkError{
            Operation: "enumerate interfaces",
            Err:       err,
            Details:   "failed to get system network interfaces",
        }
    }

    var filtered []net.Interface
    for _, iface := range allIfaces {
        if !shouldIncludeInterface(iface) {
            log.Debugf("Filtered out interface %s (reason: %s)",
                iface.Name, getFilterReason(iface))
            continue
        }
        filtered = append(filtered, iface)
        log.Debugf("Selected interface %s (%s)", iface.Name, iface.HardwareAddr)
    }

    if len(filtered) == 0 {
        return nil, &errors.ValidationError{
            Field:   "interfaces",
            Value:   "none",
            Message: "No valid multicast interfaces found after filtering",
        }
    }

    return filtered, nil
}

func shouldIncludeInterface(iface net.Interface) bool {
    // MUST be up
    if iface.Flags&net.FlagUp == 0 {
        return false // Down
    }

    // MUST support multicast
    if iface.Flags&net.FlagMulticast == 0 {
        return false // No multicast
    }

    // MUST NOT be loopback (unless explicitly requested elsewhere)
    if iface.Flags&net.FlagLoopback != 0 {
        return false // Loopback
    }

    // MUST NOT be virtual/tunnel/VPN
    if isVirtualInterface(iface.Name) {
        return false // VPN/Docker/virtual tunnel
    }

    return true
}

func isVirtualInterface(name string) bool {
    // Platform-specific virtual interface patterns
    virtualPatterns := []string{
        // VPN interfaces
        "utun",  // macOS VPN (utun0, utun1, ...)
        "tun",   // Linux/BSD VPN (tun0, ...)
        "tap",   // Linux/BSD TAP (tap0, ...)
        "ppp",   // Point-to-point (ppp0, ...)

        // Container networks
        "docker", // Docker bridge (docker0)
        "veth",   // Docker veth pairs (veth1a2b3c, ...)
        "br-",    // Docker bridge (br-a1b2c3, ...)

        // Virtual tunnels
        "gre",    // GRE tunnels
        "ipip",   // IP-in-IP tunnels
        "sit",    // IPv6-in-IPv4 tunnels
    }

    for _, pattern := range virtualPatterns {
        if strings.HasPrefix(name, pattern) {
            return true
        }
    }

    return false
}
```

---

### REQ-F10-3: Interface Validation (MANDATORY)

Before using an interface, Beacon MUST validate it is suitable for mDNS.

**Validation Checks**:
1. Interface exists (`net.InterfaceByName()` succeeds)
2. Interface is up (`FlagUp` set)
3. Interface supports multicast (`FlagMulticast` set)
4. Interface has at least one valid IPv4 address
5. Interface MTU >= 512 bytes (RFC 1035 minimum DNS message size)

**Implementation**:
```go
func validateInterface(iface net.Interface) error {
    // Check flags
    if iface.Flags&net.FlagUp == 0 {
        return &errors.ValidationError{
            Field:   "interface",
            Value:   iface.Name,
            Message: "interface is down",
        }
    }

    if iface.Flags&net.FlagMulticast == 0 {
        return &errors.ValidationError{
            Field:   "interface",
            Value:   iface.Name,
            Message: "interface does not support multicast",
        }
    }

    // Check has IPv4 address
    addrs, err := iface.Addrs()
    if err != nil {
        return err
    }

    hasIPv4 := false
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipnet.IP.To4() != nil {
                hasIPv4 = true
                break
            }
        }
    }

    if !hasIPv4 {
        return &errors.ValidationError{
            Field:   "interface",
            Value:   iface.Name,
            Message: "interface has no IPv4 address",
        }
    }

    // Check MTU
    if iface.MTU < 512 {
        return &errors.ValidationError{
            Field:   "interface",
            Value:   iface.Name,
            Message: fmt.Sprintf("MTU %d < 512 bytes minimum", iface.MTU),
        }
    }

    return nil
}
```

---

### REQ-F10-4: Logging and Visibility (MANDATORY)

Beacon MUST log interface selection decisions for debugging and transparency.

**Logging Requirements** (per F-6: Logging & Observability):
- **Debug level**: Log each interface evaluated and filter decision
- **Info level**: Log selected interfaces on startup
- **Warn level**: Log when no valid interfaces found, or problematic config detected

**Implementation**:
```go
// Debug: Each interface evaluated
log.Debugf("Evaluating interface %s: flags=%v, mtu=%d",
    iface.Name, iface.Flags, iface.MTU)
log.Debugf("Filtered out interface %s: reason=VPN (utun*)", iface.Name)

// Info: Selected interfaces summary
log.Infof("Selected %d multicast interfaces: %v",
    len(filtered), interfaceNames(filtered))

// Warn: No interfaces or suspicious config
log.Warnf("No valid multicast interfaces found after filtering")
log.Warnf("Including VPN interface %s (explicitly requested)", iface.Name)
```

---

### REQ-F10-5: Future Network Change Detection (ARCHITECTURAL)

Architecture MUST support future network change detection without breaking changes.

**Requirements** (for future M1.2 or M4):
- Interface list is NOT cached permanently
- Socket creation is decoupled from interface enumeration
- Context-based cancellation enables interface rebinding
- State transitions are idempotent (can restart without corruption)

**Current Implementation** (M1.1):
- Enumerate interfaces once at New() time
- Bind sockets to enumerated interfaces
- Manual restart required for network changes (`Close()` + `New()`)

**Future Implementation** (M1.2/M4):
- Monitor network interface changes via platform-specific APIs:
  - Linux: netlink socket monitoring
  - macOS: kqueue or SystemConfiguration notifications
  - Windows: RegisterInterfaceChange notifications
- Automatically restart sockets when changes detected
- Preserve existing queries/registrations if possible

**Architecture Support** (M1.1):
```go
type Querier struct {
    ctx       context.Context
    cancel    context.CancelFunc
    interfaces []net.Interface  // Can be updated on network change
    sockets    []*network.Socket // Can be recreated
    mu        sync.RWMutex
}

// Future: Restart() method for manual network change handling
func (q *Querier) Restart() error {
    // Close existing sockets
    // Re-enumerate interfaces
    // Re-bind sockets
    // Resume operations
}
```

**Deferred to Future Milestone**: Network change detection is complex and platform-specific. M1.1 focuses on explicit interface control. Automatic detection planned for M1.2 or M4.

---

### REQ-F10-6: Platform-Specific Interface Patterns (MANDATORY)

Beacon MUST recognize platform-specific virtual interface naming conventions.

**Platform Patterns**:

**Linux**:
- VPN: `tun*`, `tap*`, `ppp*`
- Docker: `docker0`, `veth*`, `br-*`
- Tunnels: `gre*`, `ipip*`, `sit*`
- Physical: `eth*`, `wlan*`, `enp*`, `wlp*` (systemd naming)

**macOS**:
- VPN: `utun*`, `ppp*`
- Thunderbolt Bridge: `bridge*`
- Physical: `en*` (en0=Ethernet, en1=Wi-Fi typically)

**Windows**:
- VPN: Variable naming (check interface type via WMI)
- Physical: `Ethernet*`, `Wi-Fi*`, or interface GUIDs
- Recommendation: Use interface flags, not name patterns on Windows

**Implementation**:
```go
// interface_patterns_linux.go
// +build linux

var virtualInterfacePatterns = []string{
    "tun", "tap", "ppp",
    "docker", "veth", "br-",
    "gre", "ipip", "sit",
}

// interface_patterns_darwin.go
// +build darwin

var virtualInterfacePatterns = []string{
    "utun", "ppp", "bridge",
}

// interface_patterns_windows.go
// +build windows

// On Windows, rely on interface flags and type checking
// rather than name patterns
func isVirtualInterface(name string) bool {
    // Check interface type via Windows API
    // (implementation platform-specific)
    return false
}
```

---

## Integration with F-9 (Transport Layer)

**Workflow**:
```
User creates Querier
    ↓
WithInterfaces option OR DefaultInterfaces()
    ↓
Validate each interface (REQ-F10-3)
    ↓
Pass validated interfaces to network.New() (F-9)
    ↓
F-9 binds socket and joins multicast group per interface
```

**Code Example**:
```go
// querier/querier.go
func New(opts ...Option) (*Querier, error) {
    cfg := defaultConfig()
    for _, opt := range opts {
        opt(cfg)
    }

    // Interface selection (F-10)
    if cfg.interfaces == nil {
        ifaces, err := DefaultInterfaces()
        if err != nil {
            return nil, err
        }
        cfg.interfaces = ifaces
    }

    // Validate interfaces (F-10)
    for _, iface := range cfg.interfaces {
        if err := validateInterface(iface); err != nil {
            return nil, err
        }
    }

    // Create socket with validated interfaces (F-9)
    socket, err := network.New(cfg.interfaces)
    if err != nil {
        return nil, err
    }

    return &Querier{
        socket: socket,
        // ...
    }, nil
}
```

---

## Testing Strategy

### Unit Tests

**Interface Filtering Tests**:
```go
func TestDefaultInterfaces_ExcludesVPN(t *testing.T) {
    // Mock net.Interfaces() to return mix of interfaces
    // Verify VPN interfaces (utun*, tun*, ppp*) excluded
}

func TestDefaultInterfaces_ExcludesDocker(t *testing.T) {
    // Verify docker0, veth*, br-* excluded
}

func TestDefaultInterfaces_IncludesPhysical(t *testing.T) {
    // Verify eth*, wlan*, en* included (if up + multicast)
}

func TestInterfaceValidation_RejectsDown(t *testing.T) {
    // Create interface with FlagUp not set
    // Verify validateInterface() returns error
}

func TestPlatformSpecificPatterns_Linux(t *testing.T) {
    // Test Linux-specific patterns (tun*, veth*, etc.)
}

func TestPlatformSpecificPatterns_Darwin(t *testing.T) {
    // Test macOS-specific patterns (utun*, etc.)
}
```

### Integration Tests

**Real Network Tests**:
```go
func TestDefaultInterfaces_RealNetwork(t *testing.T) {
    ifaces, err := DefaultInterfaces()
    if err != nil {
        t.Fatalf("DefaultInterfaces() failed: %v", err)
    }

    if len(ifaces) == 0 {
        t.Skip("No valid interfaces on this system")
    }

    // Verify all returned interfaces are valid
    for _, iface := range ifaces {
        if err := validateInterface(iface); err != nil {
            t.Errorf("validateInterface(%s) failed: %v", iface.Name, err)
        }
    }

    // Log for manual inspection
    for _, iface := range ifaces {
        t.Logf("Selected: %s (flags=%v, mtu=%d)",
            iface.Name, iface.Flags, iface.MTU)
    }
}

func TestExplicitInterfaceSelection(t *testing.T) {
    // Get specific interface by name
    iface, err := net.InterfaceByName("eth0")
    if err != nil {
        t.Skip("eth0 not found")
    }

    // Create querier with explicit interface
    q, err := querier.New(querier.WithInterfaces([]net.Interface{*iface}))
    if err != nil {
        t.Fatalf("New() with explicit interface failed: %v", err)
    }
    defer q.Close()

    // Verify query works on selected interface
    // ...
}
```

### Manual Testing

**VPN Leakage Test**:
```bash
# Connect to VPN
sudo openvpn --config vpn.conf &

# Verify utun* interface created
ifconfig | grep utun

# Run Beacon with default config
go run ./examples/query/ test.local

# Capture traffic on VPN interface
sudo tcpdump -i utun0 port 5353

# EXPECTED: No mDNS traffic on VPN interface (filtered by default)
# ACTUAL: (verify)
```

**Explicit Override Test**:
```bash
# Connect to VPN
sudo openvpn --config vpn.conf &

# Run Beacon with explicit VPN interface inclusion
# (requires code modification to pass all interfaces)

# Capture traffic
sudo tcpdump -i utun0 port 5353

# EXPECTED: mDNS traffic on VPN (explicit override)
```

---

## Error Handling

### Interface Errors

All interface errors MUST use appropriate error types from F-3:

**ValidationError** (invalid interface configuration):
```go
&errors.ValidationError{
    Field:   "interface",
    Value:   "eth0",
    Message: "interface does not support multicast",
}
```

**NetworkError** (interface enumeration failure):
```go
&errors.NetworkError{
    Operation: "enumerate interfaces",
    Err:       err,
    Details:   "failed to get system network interfaces",
}
```

---

## Configuration API

### Option Functions

```go
package querier

// WithInterfaces specifies explicit list of interfaces
func WithInterfaces(ifaces []net.Interface) Option {
    return func(cfg *config) {
        cfg.interfaces = ifaces
    }
}

// WithInterfaceFilter provides custom filtering logic
func WithInterfaceFilter(filter func(net.Interface) bool) Option {
    return func(cfg *config) {
        allIfaces, _ := net.Interfaces()
        var filtered []net.Interface
        for _, iface := range allIfaces {
            if filter(iface) {
                filtered = append(filtered, iface)
            }
        }
        cfg.interfaces = filtered
    }
}

// WithIncludeLoopback includes loopback interface (for testing)
func WithIncludeLoopback(enabled bool) Option {
    return func(cfg *config) {
        cfg.includeLoopback = enabled
    }
}
```

### Helper Functions

```go
package querier

// DefaultInterfaces returns multicast-capable physical interfaces
// (excludes loopback, VPN, Docker, down interfaces)
func DefaultInterfaces() ([]net.Interface, error)

// PhysicalInterfaces returns only physical interfaces
// (excludes all virtual: loopback, VPN, Docker, tunnels)
func PhysicalInterfaces() ([]net.Interface, error)

// AllInterfaces returns all up multicast interfaces
// (includes VPN, Docker - for advanced use cases)
func AllInterfaces() ([]net.Interface, error)
```

---

## Success Criteria

- [x] Explicit interface API via functional options (WithInterfaces, WithInterfaceFilter)
- [x] DefaultInterfaces() excludes VPN, Docker, loopback by default
- [x] Interface validation (up, multicast, IPv4 address, MTU >= 512)
- [x] Platform-specific virtual interface patterns (Linux, macOS, Windows)
- [x] Logging of interface selection decisions (debug, info, warn)
- [x] Architecture supports future network change detection
- [x] Integration with F-9 transport layer
- [x] Unit tests for filtering logic
- [x] Integration tests on real network
- [x] Manual VPN leakage testing

---

## Governance and Compliance

### Constitutional Compliance

**Principle I (RFC Compliant)**:
- ✅ RFC 6762 §2 link-local scope: VPN/virtual interface exclusion enforces link-local requirement
- ✅ Interface filtering prevents protocol violations (non-link-local traffic)

**Principle VIII (Excellence)**:
- ✅ Addresses interface management pitfalls from research (ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §2)
- ✅ Privacy-preserving by default (no VPN leakage)
- ✅ Secure defaults with explicit override capability

### Change Control

Changes to this specification require:
1. RFC validation review (if link-local scope enforcement changes)
2. Privacy impact assessment (if default filtering changes)
3. Platform compatibility testing (Linux, macOS, Windows)
4. Version bump per semantic versioning:
   - **MAJOR**: Breaking changes to interface API or default behavior
   - **MINOR**: New interface helpers or non-breaking enhancements
   - **PATCH**: Bug fixes, pattern updates, documentation improvements

---

## References

**Constitutional**:
- [Beacon Constitution v1.1.0](../memory/constitution.md)

**Architectural**:
- [ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md](../../docs/ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md) - §2 (Interface Management)
- [F-2: Package Structure](./F-2-package-structure.md) - `internal/network/` package
- [F-3: Error Handling](./F-3-error-handling.md) - Error types
- [F-5: Configuration](./F-5-configuration.md) - Functional options pattern
- [F-6: Logging & Observability](./F-6-logging-observability.md) - Logging requirements
- [F-9: Transport Layer](./F-9-transport-layer-socket-configuration.md) - Socket creation

**RFCs**:
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - §2 (Link-Local Scope), §5 (Multicast)

**Research**:
- hashicorp/mdns - Issue #122 (wrong interface binding)
- hashicorp/mdns - Issue #35 (IPv6 binding failures)
- User reports: VPN connectivity issues, printer discovery failures

---

## Version History

| Version | Date | Changes | Validated Against |
|---------|------|---------|-------------------|
| 1.0.0 | 2025-11-01 | Initial interface management specification. Defines explicit interface API (WithInterfaces, WithInterfaceFilter), default filtering logic (VPN/Docker exclusion), interface validation, platform-specific patterns. Architecture supports future network change detection. | Constitution v1.1.0, RFC 6762 §2 (link-local scope) |
