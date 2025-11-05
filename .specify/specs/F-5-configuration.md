# F-5: Configuration & Defaults

**Spec ID**: F-5
**Type**: Architecture
**Status**: Validated (2025-11-01)
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling)
**References**: BEACON_FOUNDATIONS v1.1 §6
**Governance**: [Beacon Constitution v1.0.0](../memory/constitution.md)
**RFC Validation**: Validated against RFC 6762 and RFC 6763 (2025-11-01)

**Revision Notes**:
- **2025-11-01**: Aligned with Constitution v1.0.0 and BEACON_FOUNDATIONS v1.1
  - Updated governance references to Constitution v1.0.0
  - Updated foundation references to v1.1
  - RFC compliance validation completed (RFC 6762 §8.1, §10; RFC 6763 §6.2, §7)
  - Removed probe count/interval configurability (RFC 6762 §8.1 MUST requirements)
  - Added TXT record size validation (RFC 6763 §6.2)
  - Added missing timing constants (TC delay, initial probe delay)
  - Clarified RFC MUST vs configurable defaults per Constitution Principle I

---

## Overview

This specification defines Beacon's configuration strategy, including the functional options pattern, default values, validation rules, and rationale for defaults. Good defaults enable zero-configuration operation while allowing customization when needed.

**Constitutional Alignment**: This specification implements Constitution Principle I (RFC Compliance) by distinguishing between RFC MUST requirements (non-configurable) and recommended defaults (configurable). All configuration options are validated against RFC 6762 (mDNS) and RFC 6763 (DNS-SD) to ensure compliance.

---

## Requirements

### REQ-F5-1: Functional Options Pattern
Public constructors MUST use the functional options pattern for configuration.

**Rationale**: Idiomatic Go, allows optional parameters, backwards compatible.

### REQ-F5-2: Sensible Defaults
Components MUST have sensible defaults allowing zero-configuration operation.

**Rationale**: Users should be able to `New()` and have it work for common cases.

### REQ-F5-3: Early Validation
Configuration MUST be validated at construction time, not at runtime.

**Rationale**: Fail fast, catch errors before operation begins.

### REQ-F5-4: Immutability
Configuration SHOULD be immutable after construction.

**Rationale**: Thread-safety, predictability.

### REQ-F5-5: Documentation
All configuration options MUST be documented with rationale and valid ranges.

---

## Functional Options Pattern

### Basic Pattern

```go
// Option configures a Querier.
type Option func(*config)

type config struct {
    timeout       time.Duration
    maxRetries    int
    interfaces    []net.Interface
    logger        Logger
}

// New creates a new Querier with the given options.
func New(opts ...Option) (*Querier, error) {
    cfg := defaultConfig()

    for _, opt := range opts {
        opt(&cfg)
    }

    if err := cfg.validate(); err != nil {
        return nil, &ValidationError{
            Field:   "configuration",
            Message: err.Error(),
        }
    }

    return newQuerier(cfg)
}

func defaultConfig() config {
    return config{
        timeout:    5 * time.Second,
        maxRetries: 3,
        interfaces: nil, // nil means all interfaces
        logger:     nil, // nil means no logging
    }
}

func (c *config) validate() error {
    if c.timeout < 0 {
        return errors.New("timeout must be non-negative")
    }
    if c.maxRetries < 0 {
        return errors.New("maxRetries must be non-negative")
    }
    return nil
}
```

### Option Functions

```go
// WithTimeout sets the query timeout.
// Default: 5 seconds.
func WithTimeout(d time.Duration) Option {
    return func(c *config) {
        c.timeout = d
    }
}

// WithMaxRetries sets the maximum number of retries.
// Default: 3.
func WithMaxRetries(n int) Option {
    return func(c *config) {
        c.maxRetries = n
    }
}

// WithInterfaces sets specific network interfaces to use.
// Default: all interfaces.
func WithInterfaces(ifaces []net.Interface) Option {
    return func(c *config) {
        c.interfaces = ifaces
    }
}

// WithLogger sets a custom logger.
// Default: no logging.
func WithLogger(logger Logger) Option {
    return func(c *config) {
        c.logger = logger
    }
}
```

### Usage

```go
// Zero-config (uses defaults)
q, err := querier.New()

// Custom configuration
q, err := querier.New(
    querier.WithTimeout(10*time.Second),
    querier.WithMaxRetries(5),
)

// All options
q, err := querier.New(
    querier.WithTimeout(10*time.Second),
    querier.WithMaxRetries(5),
    querier.WithInterfaces(selectedIfaces),
    querier.WithLogger(myLogger),
)
```

---

## Default Values

### Timing Defaults

Based on BEACON_FOUNDATIONS v1.1 §6.3:

```go
const (
    // DefaultQueryTimeout is the default timeout for queries.
    // Configurable via WithTimeout().
    DefaultQueryTimeout = 5 * time.Second

    // --- RFC MANDATED (NOT CONFIGURABLE) ---

    // ProbeInterval is the time between probe packets.
    // RFC 6762 §8.1: MUST be 250 milliseconds.
    // NOT CONFIGURABLE - RFC requirement.
    ProbeInterval = 250 * time.Millisecond

    // ProbeCount is the number of probe packets to send.
    // RFC 6762 §8.1: MUST be 3 probes.
    // NOT CONFIGURABLE - RFC requirement.
    ProbeCount = 3

    // InitialProbeDelayMin is the minimum random delay before first probe.
    // RFC 6762 §8.1: 0-250ms random delay.
    InitialProbeDelayMin = 0
    // InitialProbeDelayMax is the maximum random delay before first probe.
    InitialProbeDelayMax = 250 * time.Millisecond

    // ResponseDelayMin is the minimum random delay for shared records.
    // RFC 6762 §6: 20-120ms.
    ResponseDelayMin = 20 * time.Millisecond
    // ResponseDelayMax is the maximum random delay for shared records.
    ResponseDelayMax = 120 * time.Millisecond

    // TCDelayMin is the minimum delay when TC bit set.
    // RFC 6762 §7.2: 400-500ms.
    TCDelayMin = 400 * time.Millisecond
    // TCDelayMax is the maximum delay when TC bit set.
    TCDelayMax = 500 * time.Millisecond

    // --- CONFIGURABLE WITH CONSTRAINTS ---

    // DefaultAnnounceInterval is the time between announcements.
    // RFC 6762 §8.3: MUST be at least 1 second.
    // Configurable, but minimum enforced.
    DefaultAnnounceInterval = 1 * time.Second

    // MinAnnounceCount is the minimum number of announcements.
    // RFC 6762 §8.3: MUST be at least 2.
    // Configurable via WithAnnounceCount(), but minimum enforced.
    MinAnnounceCount = 2

    // DefaultAnnounceCount is the default number of announcements.
    DefaultAnnounceCount = 2
)
```

### TTL Defaults

Based on BEACON_FOUNDATIONS v1.1 §6.2:

```go
const (
    // DefaultHostTTL is the TTL for host name records (A/AAAA).
    // RFC 6762 §10: 120 seconds.
    DefaultHostTTL = 120 * time.Second

    // DefaultServiceTTL is the TTL for service records (PTR/SRV/TXT).
    // RFC 6762 §10: 4500 seconds (75 minutes).
    DefaultServiceTTL = 4500 * time.Second

    // DefaultCacheRefreshThreshold is when to refresh cached records.
    // RFC 6762 §5.2: 80% of TTL.
    DefaultCacheRefreshThreshold = 0.80
)
```

### Network Defaults

```go
const (
    // DefaultMulticastIPv4 is the mDNS multicast address for IPv4.
    DefaultMulticastIPv4 = "224.0.0.251"

    // DefaultMulticastIPv6 is the mDNS multicast address for IPv6.
    DefaultMulticastIPv6 = "FF02::FB"

    // DefaultPort is the mDNS port.
    DefaultPort = 5353

    // DefaultMaxMessageSize is the maximum mDNS message size.
    // RFC 6762 §17: Up to 9000 bytes for multicast.
    DefaultMaxMessageSize = 9000
)
```

### DNS-SD Defaults

```go
const (
    // TXTRecordRecommendedMax is the recommended maximum TXT record size.
    // RFC 6763 §6.2: SHOULD be ≤200 bytes.
    TXTRecordRecommendedMax = 200

    // TXTRecordPreferredMax is the preferred maximum TXT record size.
    // RFC 6763 §6.2: Preferably ≤400 bytes.
    TXTRecordPreferredMax = 400

    // TXTRecordAbsoluteMax is the absolute maximum TXT record size.
    // RFC 6763 §6.2: >1300 bytes not recommended (fragmentation).
    TXTRecordAbsoluteMax = 1300

    // ServiceNameMaxLength is the maximum service name length.
    // RFC 6763 §7: ≤15 characters.
    ServiceNameMaxLength = 15
)
```

### Resource Limits

```go
const (
    // DefaultCacheSize is the maximum number of cached records.
    // 0 means unlimited (rely on TTL for eviction).
    DefaultCacheSize = 0

    // DefaultMaxConcurrentQueries is the maximum concurrent queries.
    // 0 means unlimited.
    DefaultMaxConcurrentQueries = 100

    // DefaultReceiveBufferSize is the UDP receive buffer size.
    DefaultReceiveBufferSize = 65536 // 64 KB

    // DefaultSendBufferSize is the UDP send buffer size.
    DefaultSendBufferSize = 65536 // 64 KB
)
```

---

## Configuration Options by Component

### Querier Options

```go
// Querier-specific options

// WithTimeout sets the query timeout.
func WithTimeout(d time.Duration) Option

// WithRetries sets the maximum number of retries.
func WithRetries(n int) Option

// WithInterval sets the retry interval (doubles each retry).
func WithInterval(d time.Duration) Option

// WithUnicastResponse requests unicast responses (QU bit).
func WithUnicastResponse(enabled bool) Option

// WithKnownAnswerSuppression includes known answers in queries.
func WithKnownAnswerSuppression(enabled bool) Option
```

### Responder Options

```go
// Responder-specific options

// WithHostName sets the host name to advertise.
func WithHostName(name string) Option

// WithConflictHandler sets a custom conflict resolution handler.
func WithConflictHandler(handler ConflictHandler) Option

// WithAnnounceCount sets the number of announcements.
// RFC 6762 §8.3: MUST be at least 2.
// Default: 2. Values < 2 will be rejected.
func WithAnnounceCount(n int) Option
```

### Browser/Publisher Options (DNS-SD)

```go
// Service-specific options

// WithDomain sets the DNS-SD domain (default "local.").
func WithDomain(domain string) Option

// WithSubtypes sets service subtypes.
func WithSubtypes(subtypes []string) Option

// WithTXTRecords sets TXT record key/value pairs.
func WithTXTRecords(txt map[string]string) Option
```

### Common Options (All Components)

```go
// Common options

// WithInterfaces sets specific network interfaces.
// Default: all multicast-capable interfaces.
func WithInterfaces(ifaces []net.Interface) Option

// WithIPv4 enables/disables IPv4 (default true).
func WithIPv4(enabled bool) Option

// WithIPv6 enables/disables IPv6 (default true).
func WithIPv6(enabled bool) Option

// WithLogger sets a custom logger.
func WithLogger(logger Logger) Option

// WithContext sets the context for lifecycle.
func WithContext(ctx context.Context) Option
```

---

## Validation Rules

### Network Configuration

```go
func (c *config) validateNetwork() error {
    if !c.ipv4 && !c.ipv6 {
        return errors.New("at least one of IPv4 or IPv6 must be enabled")
    }

    if c.port < 1 || c.port > 65535 {
        return fmt.Errorf("invalid port: %d (must be 1-65535)", c.port)
    }

    for _, iface := range c.interfaces {
        if iface.Flags&net.FlagMulticast == 0 {
            return fmt.Errorf("interface %s does not support multicast", iface.Name)
        }
    }

    return nil
}
```

### Timing Configuration

```go
func (c *config) validateTiming() error {
    if c.timeout < 0 {
        return errors.New("timeout cannot be negative")
    }

    // Announce count validation (RFC 6762 §8.3: MUST be at least 2)
    if c.announceCount < MinAnnounceCount {
        return fmt.Errorf("announceCount must be at least %d (RFC 6762 §8.3)", MinAnnounceCount)
    }

    // Announce interval validation (RFC 6762 §8.3: MUST be at least 1 second)
    if c.announceInterval < DefaultAnnounceInterval {
        return fmt.Errorf("announceInterval must be at least %v (RFC 6762 §8.3)", DefaultAnnounceInterval)
    }

    return nil
}
```

### Name Validation

```go
func (c *config) validateNames() error {
    if c.hostName != "" {
        if err := validateDomainName(c.hostName); err != nil {
            return fmt.Errorf("invalid hostName: %w", err)
        }
    }

    if c.serviceType != "" {
        if err := validateServiceType(c.serviceType); err != nil {
            return fmt.Errorf("invalid serviceType: %w", err)
        }
    }

    return nil
}

func validateDomainName(name string) error {
    if len(name) == 0 || len(name) > 255 {
        return errors.New("domain name length must be 1-255 bytes")
    }

    labels := strings.Split(name, ".")
    for _, label := range labels {
        if len(label) == 0 || len(label) > 63 {
            return fmt.Errorf("label %q exceeds 63 bytes", label)
        }
    }

    return nil
}

func validateServiceType(serviceType string) error {
    // Format: _servicename._tcp or _servicename._udp
    parts := strings.Split(serviceType, ".")
    if len(parts) != 2 {
        return errors.New("must be '_servicename._tcp' or '_servicename._udp'")
    }

    if !strings.HasPrefix(parts[0], "_") {
        return errors.New("service name must start with underscore")
    }

    if parts[1] != "_tcp" && parts[1] != "_udp" {
        return errors.New("protocol must be '_tcp' or '_udp'")
    }

    serviceName := strings.TrimPrefix(parts[0], "_")
    if len(serviceName) > ServiceNameMaxLength {
        return fmt.Errorf("service name exceeds %d characters (RFC 6763 §7)", ServiceNameMaxLength)
    }

    return nil
}
```

### DNS-SD Validation

```go
func (c *config) validateDNSSD() error {
    // Validate TXT records
    if c.txtRecords != nil {
        totalSize := 0
        for key, value := range c.txtRecords {
            // Each key=value pair: 1 byte length + len(key) + 1 ('=') + len(value)
            pairSize := 1 + len(key) + 1 + len(value)
            totalSize += pairSize

            if pairSize > 255 {
                return fmt.Errorf("TXT record pair %q=%q exceeds 255 bytes", key, value)
            }
        }

        // Warn if exceeds recommended size
        if totalSize > TXTRecordRecommendedMax {
            // This should be a warning, not an error
            // In implementation, log a warning but don't reject
            if totalSize > TXTRecordAbsoluteMax {
                return fmt.Errorf("TXT record total size %d bytes exceeds %d bytes (RFC 6763 §6.2: fragmentation)",
                    totalSize, TXTRecordAbsoluteMax)
            }
        }
    }

    return nil
}
```

---

## Configuration Helpers

### Detecting Network Interfaces

```go
// DefaultInterfaces returns all multicast-capable interfaces.
func DefaultInterfaces() ([]net.Interface, error) {
    allIfaces, err := net.Interfaces()
    if err != nil {
        return nil, err
    }

    var multicastIfaces []net.Interface
    for _, iface := range allIfaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // Skip down interfaces
        }
        if iface.Flags&net.FlagMulticast == 0 {
            continue // Skip non-multicast interfaces
        }
        multicastIfaces = append(multicastIfaces, iface)
    }

    return multicastIfaces, nil
}
```

### Hostname Detection

```go
// DefaultHostName returns the system hostname + ".local."
func DefaultHostName() (string, error) {
    hostname, err := os.Hostname()
    if err != nil {
        return "", err
    }

    // Ensure .local suffix
    if !strings.HasSuffix(hostname, ".local.") {
        if strings.HasSuffix(hostname, ".local") {
            hostname += "."
        } else {
            hostname += ".local."
        }
    }

    return hostname, nil
}
```

---

## Defaults Rationale

### RFC MUST Requirements (Not Configurable)

**Probe Count = 3, Probe Interval = 250ms**:
- **RFC 6762 §8.1**: "A host probes to see if anyone else is using a name by sending three probe queries, 250 milliseconds apart"
- **MUST requirement**: Not configurable - violates RFC compliance
- **Total time**: 750ms (reasonable for startup)
- **Reliability**: 3 probes catch intermittent issues

**Response Delays (20-120ms, 400-500ms)**:
- **RFC 6762 §6, §7.2**: Specific delays mandated
- **MUST requirement**: Not configurable

**Minimum Announce Count ≥ 2**:
- **RFC 6762 §8.3**: "MUST send at least two unsolicited announcements"
- **MUST requirement**: Enforced via validation, can be increased

### Configurable Defaults

**Why 5-second query timeout?**
- **Long enough**: Allows for network delays, retries
- **Short enough**: Doesn't hang UI
- **Common**: Used by many DNS clients

**Why 120s TTL for host names?**
- **RFC recommendation**: RFC 6762 §10
- **Rationale**: Hosts change IPs (DHCP, mobile)
- **Not too short**: Reduces query traffic

**Why 75-minute TTL for services?**
- **RFC recommendation**: RFC 6762 §10
- **Rationale**: Services more stable than host IPs
- **Long enough**: Reduces network traffic
- **Short enough**: Changes propagate reasonably fast

**Why no cache size limit?**
- **TTL-based eviction**: Records expire naturally
- **Memory**: Modern systems have ample RAM
- **Simplicity**: No need for LRU logic
- **Override**: Users can set limit if needed

**Why 100 concurrent queries?**
- **Reasonable limit**: Prevents resource exhaustion
- **High enough**: Supports aggressive scanning
- **Low enough**: Bounds resource usage

---

## Environment Variables (Optional)

For debugging/testing, support environment variables:

```go
// Optional environment variable overrides (for debugging)

func init() {
    if timeout := os.Getenv("BEACON_QUERY_TIMEOUT"); timeout != "" {
        if d, err := time.ParseDuration(timeout); err == nil {
            DefaultQueryTimeout = d
        }
    }

    if logLevel := os.Getenv("BEACON_LOG_LEVEL"); logLevel != "" {
        DefaultLogLevel = parseLogLevel(logLevel)
    }
}
```

**Use sparingly**: Environment variables are global mutable state. Prefer explicit options.

---

## Configuration Examples

### Zero Configuration

```go
// Use all defaults
q, err := querier.New()
if err != nil {
    log.Fatal(err)
}
```

### Custom Timeout

```go
q, err := querier.New(
    querier.WithTimeout(10*time.Second),
)
```

### Specific Interface

```go
iface, _ := net.InterfaceByName("eth0")
q, err := querier.New(
    querier.WithInterfaces([]net.Interface{*iface}),
)
```

### Full Configuration

```go
q, err := querier.New(
    querier.WithTimeout(10*time.Second),
    querier.WithRetries(5),
    querier.WithInterval(2*time.Second),
    querier.WithUnicastResponse(true),
    querier.WithLogger(myLogger),
    querier.WithIPv4(true),
    querier.WithIPv6(false),
)
```

### DNS-SD Service

```go
p, err := publisher.New(
    publisher.WithDomain("local."),
    publisher.WithTXTRecords(map[string]string{
        "version": "1.0",
        "path":    "/api",
    }),
    publisher.WithSubtypes([]string{"_printer"}),
)
```

---

## Testing Configuration

### Test Validation

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        opts    []Option
        wantErr bool
    }{
        {
            name: "valid config",
            opts: []Option{
                WithTimeout(5 * time.Second),
                WithRetries(3),
            },
            wantErr: false,
        },
        {
            name: "negative timeout",
            opts: []Option{
                WithTimeout(-1 * time.Second),
            },
            wantErr: true,
        },
        {
            name: "invalid interface",
            opts: []Option{
                WithInterfaces([]net.Interface{{Name: "invalid"}}),
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := New(tt.opts...)
            if (err != nil) != tt.wantErr {
                t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Defaults

```go
func TestDefaults(t *testing.T) {
    q, err := New()
    if err != nil {
        t.Fatal(err)
    }

    // Verify defaults
    if q.timeout != DefaultQueryTimeout {
        t.Errorf("timeout = %v, want %v", q.timeout, DefaultQueryTimeout)
    }

    if q.maxRetries != 3 {
        t.Errorf("maxRetries = %d, want 3", q.maxRetries)
    }
}
```

---

## Open Questions

**Q1**: Should configuration be hot-reloadable?
- **Pro**: Change without restart
- **Con**: Complexity, thread-safety issues
- **Decision**: No, immutable after construction. Restart to reconfigure.

**Q2**: Configuration file support (YAML/JSON)?
- **Pro**: Easier for ops teams
- **Con**: Additional dependency, complexity
- **Decision**: Not initially. Users can parse and pass options.

**Q3**: Default hostname with uniqueness suffix?
- **Example**: `hostname-abc123.local.`
- **Pro**: Avoids conflicts
- **Con**: Ugly names
- **Decision**: No suffix by default. Let conflict resolution handle it.

---

## Constitution Compliance

This specification aligns with the [Beacon Constitution v1.0.0](../memory/constitution.md) as follows:

**Principle I - RFC Compliant (NON-NEGOTIABLE)**:
- ✅ All RFC MUST requirements identified and enforced as non-configurable constants
- ✅ RFC 6762 §8.1: Probe count (3) and interval (250ms) are constants, not configurable
- ✅ RFC 6762 §8.3: Minimum announce count (2) and interval (1s) enforced via validation
- ✅ RFC 6762 §10: Default TTL values follow RFC recommendations
- ✅ RFC 6763 §6.2: TXT record size constraints validated
- ✅ RFC 6763 §7: Service name length limits (15 chars) enforced
- ✅ Validation reports configuration errors referencing specific RFC sections

**Principle II - Spec-Driven Development (NON-NEGOTIABLE)**:
- ✅ Complete specification provided before implementation
- ✅ Configuration patterns, defaults, and validation rules fully specified
- ✅ Examples and test strategies included

**Principle III - Test-Driven Development (NON-NEGOTIABLE)**:
- ✅ Test validation examples provided (§Testing Configuration)
- ✅ Validation logic specified for TDD implementation
- ✅ Edge cases identified (negative timeouts, invalid interfaces, RFC violations)

**Principle IV - Phased Approach**:
- ✅ Configuration system supports incremental implementation
- ✅ Defaults allow early milestones to work without full customization

**Principle V - Open Source**:
- ✅ Public specification in repository
- ✅ Clear documentation for contributors

**Principle VI - Maintained**:
- ✅ Immutable configuration after construction ensures stability
- ✅ Functional options pattern supports backwards-compatible additions

**Principle VII - Excellence**:
- ✅ Idiomatic Go patterns (functional options)
- ✅ Comprehensive validation and clear error messages
- ✅ Well-documented rationale for all defaults

**RFC Validation Status**: ✅ PASSED (2025-11-01)
- All configuration options reviewed against RFC 6762 and RFC 6763
- MUST requirements identified and made non-configurable
- SHOULD requirements used as configurable defaults with validation
- No RFC compliance issues identified

---

## Success Criteria

- [ ] Functional options pattern implemented
- [ ] All defaults documented with rationale
- [ ] Validation catches invalid configurations
- [ ] Zero-config works for common cases
- [ ] Configuration immutable after construction
- [ ] Tests verify defaults and validation

---

## References

### Governance
- [Beacon Constitution v1.0.0](../memory/constitution.md) - Project governance and principles
- Constitution Principle I: RFC Compliance (NON-NEGOTIABLE)
- Constitution Principle II: Spec-Driven Development

### Technical References
- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md) §6 (Common Requirements)
- RFC 6762 (Multicast DNS) - Default TTLs, timing values, MUST requirements
  - §8.1: Probing (3 probes, 250ms intervals - MUST requirements)
  - §8.3: Announcing (minimum 2 announcements, 1 second intervals)
  - §10: Resource Record TTL values
- RFC 6763 (DNS-Based Service Discovery)
  - §6.2: TXT Record size constraints
  - §7: Service Name length limits (15 characters)
- Go Wiki: [Functional Options](https://github.com/uber-go/guide/blob/master/style.md#functional-options)
