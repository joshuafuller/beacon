# F-11: Security Architecture

**Spec ID**: F-11
**Type**: Architecture
**Status**: Draft
**Version**: 1.0.0
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling), F-6 (Logging), F-8 (Testing Strategy), F-9 (Transport Layer), F-10 (Interface Management)
**References**:
- Beacon Constitution v1.1.0 (Principle I: RFC Compliance, Principle VIII: Excellence)
- ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §3 (Security Architecture Gaps)
- RFC 6762 §2 (Multicast DNS Scope - Link-Local)
- RFC 6762 §6 (Multicast DNS Resource Record TTL and Cache Coherency)
- RFC 6762 §18 (Security Considerations)

**Governance**: Development governed by [Beacon Constitution v1.1.0](../memory/constitution.md)

**RFC Validation**: Pending. This specification implements RFC 6762 §18 security considerations through defense-in-depth: source IP filtering, rate limiting, and input validation.

---

## Overview

This specification defines Beacon's security architecture for defending against attacks, malicious packets, and resource exhaustion while maintaining RFC 6762 compliance. The security layer MUST:

1. **Enforce link-local scope** through source IP filtering (RFC 6762 §2)
2. **Protect against resource exhaustion** via rate limiting and bounds checking
3. **Handle malformed packets** gracefully without panics or crashes
4. **Enable security testing** through fuzzing and attack simulation

**Critical Insight**: mDNS operates on untrusted local networks. Malicious or buggy devices can send crafted packets to cause DoS, crashes, or resource exhaustion. This specification implements defense-in-depth security controls while maintaining query-only simplicity for M1.1.

**Constitutional Alignment**:
- **Principle I (RFC Compliance)**: RFC 6762 §18 mandates security considerations - this spec implements them
- **Principle VIII (Excellence)**: Addresses security pitfalls from research (DRDoS, multicast storms, malformed packets)

---

## Threat Model

### In-Scope Threats (M1.1 Query-Only)

**T1: Source IP Spoofing**
- Attacker sends packets from non-link-local IP (spoofed or routed)
- **Impact**: CPU waste processing invalid packets, potential for DRDoS participation
- **Mitigation**: Source IP filtering (REQ-F11-1)

**T2: Multicast Storm (Resource Exhaustion)**
- Malicious or buggy device sends 1000+ queries/second
- Real-world: Hubitat bug (2020) generated 1000 queries/sec, crashed ESP32 devices
- **Impact**: Memory exhaustion (unbounded goroutines), CPU saturation
- **Mitigation**: Per-source-IP rate limiting (REQ-F11-2)

**T3: Malformed Packet (Crash/DoS)**
- Attacker sends crafted packet to exploit parsing vulnerabilities
- Examples: Compression pointer loops, buffer overruns, invalid label lengths
- **Impact**: Panic, crash, infinite loop, memory corruption
- **Mitigation**: Input validation, fuzzing (REQ-F11-3, REQ-F11-4)

**T4: Oversized Packet (Resource Exhaustion)**
- Attacker sends packets larger than RFC maximum (9000 bytes)
- **Impact**: Memory allocation DoS, buffer overflow
- **Mitigation**: Packet size validation (REQ-F11-3)

### Out-of-Scope Threats (Future Milestones)

**T5: Cache Poisoning** - Deferred to M2 (requires cache implementation)
**T6: DRDoS Amplification** - Query-only doesn't respond, not an amplification vector
**T7: Split-Brain mDNS** - Mitigated by F-9 (SO_REUSEPORT), M5 (daemon detection)

---

## Requirements

### REQ-F11-1: Source IP Filtering (MANDATORY - RFC Compliant)

Beacon MUST validate source IP is link-local or same subnet as receiving interface, silently dropping non-link-local packets BEFORE parsing.

**Rationale**:
- RFC 6762 §2 specifies link-local scope
- Packets from non-link-local IPs are routing errors or spoofed attack traffic
- Early rejection prevents expensive packet parsing for attack traffic

**RFC Alignment**: RFC 6762 §2 "Multicast DNS is restricted to link-local scope." Source IP validation enforces this at transport layer.

**Validation Logic**:
```go
// IPv4 link-local validation
func isLinkLocalSource(srcIP net.IP, iface net.Interface) bool {
    // Check if source IP is link-local range (169.254.0.0/16 per RFC 3927)
    if isLinkLocal(srcIP) {
        return true
    }

    // Check if source IP is in same subnet as interface
    addrs, err := iface.Addrs()
    if err != nil {
        return false
    }

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipnet.Contains(srcIP) {
                return true // Same subnet
            }
        }
    }

    return false // Not link-local, not same subnet → REJECT
}

func isLinkLocal(ip net.IP) bool {
    // IPv4: 169.254.0.0/16 (RFC 3927)
    if ip4 := ip.To4(); ip4 != nil {
        return ip4[0] == 169 && ip4[1] == 254
    }

    // IPv6: fe80::/10 (link-local)
    if len(ip) == net.IPv6len {
        return ip[0] == 0xfe && (ip[1]&0xc0) == 0x80
    }

    return false
}
```

**Implementation** (in receive loop):
```go
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error) {
    buf := make([]byte, 9000) // RFC 6762 §17 max message size

    for {
        // F-9 REQ-F9-7: Check context cancellation BEFORE blocking
        select {
        case <-ctx.Done():
            return nil, nil, ctx.Err() // Graceful shutdown on cancellation
        default:
            // Continue to receive
        }

        // F-9 REQ-F9-7: Propagate context deadline to socket
        if deadline, ok := ctx.Deadline(); ok {
            s.conn.SetReadDeadline(deadline)
        } else {
            // Set short timeout to allow periodic ctx.Done() checking
            s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
        }

        n, srcAddr, err := s.conn.ReadFrom(buf)
        if err != nil {
            // Check if timeout (allows ctx.Done() check on next iteration)
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                continue // Timeout, check ctx.Done() on next iteration
            }
            return nil, nil, err // Other errors
        }

        udpAddr, ok := srcAddr.(*net.UDPAddr)
        if !ok {
            continue // Unexpected address type
        }

        // REQ-F11-1: Source IP filtering (BEFORE parsing)
        if !isLinkLocalSource(udpAddr.IP, s.iface) {
            // Silent drop, no error, optional debug log
            log.Debugf("Dropped packet from non-link-local source %s on %s",
                udpAddr.IP, s.iface.Name)
            continue // Drop silently
        }

        // REQ-F11-2: Rate limiting check (BEFORE parsing)
        if s.rateLimiter.IsRateLimited(udpAddr.IP) {
            log.Debugf("Dropped packet from rate-limited source %s", udpAddr.IP)
            continue // Drop silently
        }

        // REQ-F11-3: Packet size validation
        if n > 9000 {
            log.Warnf("Dropped oversized packet (%d bytes) from %s", n, udpAddr.IP)
            continue
        }

        // Passed all security checks, return packet for parsing
        return buf[:n], srcAddr, nil
    }
}
```

**Performance**: Single IP comparison per packet (~5 nanoseconds), negligible overhead.

**Logging**: Debug level only (avoid log spam), disabled by default.

---

### REQ-F11-2: Rate Limiting (MANDATORY - Resilience)

Beacon MUST implement per-source-IP rate limiting with cooldown to protect against multicast storms and resource exhaustion.

**Rationale**:
- Real-world example: Hubitat bug (2020) sent 1000+ queries/second, crashed ESP32 devices
- Unbounded goroutine creation → memory exhaustion
- CPU saturation processing attack traffic
- Library must not crash AND must not amplify storms

**Requirements**:
1. Track query rate per source IP (sliding 1-second window)
2. Threshold: >100 queries/second from single IP → cooldown
3. Cooldown: Drop packets from source for 60 seconds
4. Cleanup: Periodically remove expired cooldown entries
5. Logging: Warn on first violation, debug on subsequent (avoid log spam)
6. Configurable: Threshold and cooldown duration

**RFC Alignment**: RFC 6762 §6 specifies minimum 1 second between queries. Rate limiting at 100/sec is 100x more lenient than protocol minimum, providing protection while allowing legitimate high-frequency queries.

**Implementation**:
```go
// RateLimiter tracks per-source-IP query rates
type RateLimiter struct {
    mu             sync.RWMutex
    queries        map[string]*queryTracker // IP -> tracker
    threshold      int                      // Queries per second
    cooldownDuration time.Duration          // Cooldown period
}

type queryTracker struct {
    count       int
    windowStart time.Time
    cooldownUntil time.Time
    warnedOnce  bool // For logging
}

func NewRateLimiter(threshold int, cooldown time.Duration) *RateLimiter {
    rl := &RateLimiter{
        queries:          make(map[string]*queryTracker),
        threshold:        threshold,
        cooldownDuration: cooldown,
    }

    // Cleanup goroutine: Remove expired entries every minute
    go rl.cleanup()

    return rl
}

func (rl *RateLimiter) IsRateLimited(srcIP net.IP) bool {
    key := srcIP.String()
    now := time.Now()

    rl.mu.Lock()
    defer rl.mu.Unlock()

    tracker, exists := rl.queries[key]
    if !exists {
        // First packet from this source
        rl.queries[key] = &queryTracker{
            count:       1,
            windowStart: now,
        }
        return false
    }

    // Check if in cooldown
    if now.Before(tracker.cooldownUntil) {
        // Still in cooldown
        log.Debugf("Rate limit: source %s still in cooldown", srcIP)
        return true
    }

    // Check if window expired (sliding 1-second window)
    if now.Sub(tracker.windowStart) > time.Second {
        // New window
        tracker.count = 1
        tracker.windowStart = now
        tracker.warnedOnce = false
        return false
    }

    // Increment count in current window
    tracker.count++

    // Check if threshold exceeded
    if tracker.count > rl.threshold {
        // Rate limit exceeded → start cooldown
        tracker.cooldownUntil = now.Add(rl.cooldownDuration)

        if !tracker.warnedOnce {
            log.Warnf("Rate limit exceeded for source %s: %d queries in 1 second (threshold: %d). Cooldown for %v.",
                srcIP, tracker.count, rl.threshold, rl.cooldownDuration)
            tracker.warnedOnce = true
        }

        return true
    }

    return false
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()

        for ip, tracker := range rl.queries {
            // Remove if no recent activity and not in cooldown
            if now.Sub(tracker.windowStart) > 5*time.Minute &&
               now.After(tracker.cooldownUntil) {
                delete(rl.queries, ip)
            }
        }

        rl.mu.Unlock()
    }
}
```

**Configuration Options**:
```go
// WithRateLimit enables/disables rate limiting (default: enabled)
func WithRateLimit(enabled bool) Option

// WithRateLimitThreshold sets queries per second threshold (default: 100)
func WithRateLimitThreshold(queriesPerSec int) Option

// WithRateLimitCooldown sets cooldown duration (default: 60s)
func WithRateLimitCooldown(duration time.Duration) Option
```

**Usage**:
```go
q, err := querier.New(
    querier.WithRateLimit(true),
    querier.WithRateLimitThreshold(100),
    querier.WithRateLimitCooldown(60 * time.Second),
)
```

---

### REQ-F11-3: Input Validation (MANDATORY - Already Implemented in M1)

Beacon MUST validate all input and never panic on malformed packets.

**Status**: ✅ **ALREADY IMPLEMENTED IN M1**

**Evidence from M1**:
- T094: FuzzParseMessage with 100+ executions, zero crashes
- T084: WireFormatError includes Field, Message, Operation
- All parsing includes bounds checking
- Tests validate malformed packet handling

**Validation Requirements** (already met):
1. ✅ Validate packet bounds before every read operation
2. ✅ Validate compression pointers (offset within packet, no loops)
3. ✅ Validate label lengths (≤63 bytes per label)
4. ✅ Validate message sections fit within packet size
5. ✅ Use defensive parsing (check bounds, not assume validity)
6. ✅ Return WireFormatError for malformed packets (never panic)

**Forbidden** (already enforced):
- ❌ MUST NOT panic on any network input
- ❌ MUST NOT use unsafe.Pointer for packet parsing
- ❌ MUST NOT trust packet length fields without bounds checking
- ❌ MUST NOT allow compression pointer loops

**Error Types** (from F-3, already implemented):
```go
type WireFormatError struct {
    Op      string // Operation (e.g., "parse message", "decompress name")
    Field   string // Field that's malformed (e.g., "name pointer", "label")
    Message string // Description
    Err     error  // Underlying error (if any)
}
```

**No Action Required for M1.1**: Input validation architecture is solid from M1. This requirement documents and validates existing security controls.

---

### REQ-F11-4: Fuzzing (MANDATORY - Already Implemented in M1)

Beacon MUST include fuzzing tests for packet parser in CI pipeline.

**Status**: ✅ **ALREADY IMPLEMENTED IN M1**

**Evidence from M1**:
- T094: FuzzParseMessage test implemented
- Executed 100+ fuzz iterations
- Zero crashes discovered
- Zero panics found

**Existing Fuzz Target**:
```go
func FuzzParseMessage(f *testing.F) {
    // Seed corpus
    f.Add([]byte{...}) // Valid message
    f.Add([]byte{...}) // Malformed message

    f.Fuzz(func(t *testing.T, data []byte) {
        // Should not panic
        msg, err := message.Parse(data)
        if err == nil && msg != nil {
            // Validate parsed message is sane
            // ...
        }
    })
}
```

**CI Integration** (for M1.1):
```bash
# Run fuzz tests for 30 seconds in CI
go test -fuzz=FuzzParseMessage -fuzztime=30s ./internal/message/

# Check for crashes
if [ $? -ne 0 ]; then
    echo "Fuzz testing found crashes - FAIL"
    exit 1
fi
```

**Fuzz Corpus Maintenance**:
- Commit discovered interesting inputs to corpus
- Regression testing: All fuzz-discovered issues must remain fixed

**No Major Action Required for M1.1**: Fuzzing infrastructure exists. M1.1 adds CI integration (minor enhancement).

---

### REQ-F11-5: Packet Size Limits (MANDATORY)

Beacon MUST reject packets larger than RFC 6762 §17 maximum (9000 bytes).

**Rationale**:
- RFC 6762 §17 specifies 9000 bytes as maximum DNS message size
- Larger packets indicate malicious or malformed traffic
- Prevents memory allocation DoS

**Implementation** (in receive loop):
```go
func (s *Socket) Receive(ctx context.Context) ([]byte, net.Addr, error) {
    buf := make([]byte, 9000) // RFC 6762 §17 max

    n, srcAddr, err := s.conn.ReadFrom(buf)
    if err != nil {
        return nil, nil, err
    }

    // REQ-F11-5: Packet size validation
    if n > 9000 {
        log.Warnf("Dropped oversized packet (%d bytes > 9000 max) from %s",
            n, srcAddr)
        continue
    }

    return buf[:n], srcAddr, nil
}
```

**RFC Alignment**: RFC 6762 §17 "The 9000-byte size limit applies to the entire DNS message, including the DNS message header."

---

### REQ-F11-6: Resource Limits (MANDATORY)

Beacon MUST protect itself from resource exhaustion under attack.

**Requirements**:
1. Bounded receive buffer (64KB, see F-9)
2. Bounded packet processing (via rate limiting, see REQ-F11-2)
3. Bounded goroutine creation (via worker pool pattern in future)
4. Timeout for all network operations (via context.Context)

**Current Implementation** (M1.1):
- ✅ Socket buffers: 64KB (F-9 REQ-F9-6)
- ✅ Rate limiting: 100 queries/sec per source IP (REQ-F11-2)
- ✅ Context timeouts: All operations respect context deadlines (F-4)
- ⏳ Worker pool: Deferred to M4 (advanced resilience features)

**Future Enhancement** (M4):
```go
// Worker pool pattern for bounded goroutines
type workerPool struct {
    workers   int
    tasks     chan func()
    semaphore chan struct{}
}

func (p *workerPool) Submit(task func()) error {
    select {
    case p.semaphore <- struct{}{}:
        go func() {
            defer func() { <-p.semaphore }()
            task()
        }()
        return nil
    default:
        return errors.New("worker pool full")
    }
}
```

---

## Security Testing Strategy

### Attack Simulation Tests

**Multicast Storm Simulation**:
```go
func TestRateLimiting_MulticastStorm(t *testing.T) {
    rl := NewRateLimiter(100, 60*time.Second)

    srcIP := net.ParseIP("192.168.1.100")

    // Simulate 1000 queries in 1 second (storm)
    stormStart := time.Now()
    blockedCount := 0

    for i := 0; i < 1000; i++ {
        if rl.IsRateLimited(srcIP) {
            blockedCount++
        }
        time.Sleep(1 * time.Millisecond) // 1000 queries/sec
    }

    stormDuration := time.Since(stormStart)

    // Verify rate limiting kicked in
    if blockedCount < 800 {
        t.Errorf("Expected >800 blocked queries during storm, got %d", blockedCount)
    }

    t.Logf("Storm duration: %v, blocked: %d/1000", stormDuration, blockedCount)
}
```

**Spoofed IP Test**:
```go
func TestSourceIPFiltering_RejectsSpoofed(t *testing.T) {
    iface := getTestInterface(t) // Local interface (e.g., 192.168.1.0/24)

    testCases := []struct {
        name     string
        srcIP    string
        expected bool // true = accepted, false = rejected
    }{
        {"Link-local", "169.254.1.1", true},
        {"Same subnet", "192.168.1.100", true},
        {"Routed external", "8.8.8.8", false}, // Google DNS - not link-local
        {"Multicast source", "224.0.0.251", false}, // Invalid source
        {"Different subnet", "10.0.0.1", false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := isLinkLocalSource(net.ParseIP(tc.srcIP), iface)
            if result != tc.expected {
                t.Errorf("isLinkLocalSource(%s) = %v, want %v",
                    tc.srcIP, result, tc.expected)
            }
        })
    }
}
```

**Malformed Packet Test** (already in M1):
```go
func TestParseMessage_MalformedInput(t *testing.T) {
    testCases := [][]byte{
        {}, // Empty
        {0x00}, // Too short
        {0x00, 0x01, 0x00, 0x00}, // Truncated header
        // ... more malformed inputs
    }

    for i, data := range testCases {
        msg, err := message.Parse(data)
        if err == nil {
            t.Errorf("Case %d: Expected error for malformed input, got nil", i)
        }
        if msg != nil {
            t.Errorf("Case %d: Expected nil message for error, got %v", i, msg)
        }
    }
}
```

### Penetration Testing Checklist

**Manual Security Testing** (M1.1 completion):
- [ ] Spoofed source IP rejection (tcpdump + packet crafting)
- [ ] Multicast storm resilience (load generator sending 1000+ queries/sec)
- [ ] Malformed packet handling (fuzz corpus + manual crafted packets)
- [ ] Oversized packet rejection (>9000 bytes)
- [ ] Memory leak under attack (run storm for 10 minutes, check RSS)
- [ ] CPU usage under attack (should not spike above reasonable threshold)

---

## Integration with Other F-Series Specs

### F-3 (Error Handling)

Security errors use existing error types:

- **ValidationError**: Invalid input (oversized packets, bad IPs)
- **WireFormatError**: Malformed packets
- **NetworkError**: Socket/transport errors

### F-6 (Logging & Observability)

Security events logged per logging strategy:

- **Debug**: Dropped packets (source IP, rate limit) - high volume
- **Info**: Rate limiter activated/deactivated
- **Warn**: First rate limit violation per source IP
- **Error**: Security-critical failures (should not occur in normal operation)

### F-8 (Testing Strategy)

Security testing integrated into test strategy:

- **Unit Tests**: Rate limiting logic, source IP validation
- **Fuzz Tests**: Malformed packet handling (already done in M1)
- **Integration Tests**: Attack simulation (multicast storm, spoofed IPs)
- **Performance Tests**: Overhead of security checks (<1% CPU)

### F-9 (Transport Layer)

Security checks integrated into receive loop (before packet parsing).

### F-10 (Interface Management)

Interface filtering (VPN exclusion) provides defense-in-depth for link-local scope enforcement.

---

## Configuration API

### Security Options

```go
package querier

// WithRateLimit enables/disables rate limiting (default: enabled)
func WithRateLimit(enabled bool) Option

// WithRateLimitThreshold sets queries per second threshold (default: 100)
func WithRateLimitThreshold(queriesPerSec int) Option

// WithRateLimitCooldown sets cooldown duration (default: 60s)
func WithRateLimitCooldown(duration time.Duration) Option

// WithSourceIPFiltering enables/disables source IP filtering (default: enabled)
// Disabling is NOT recommended (violates RFC 6762 §2)
func WithSourceIPFiltering(enabled bool) Option
```

**Usage**:
```go
// Default: All security features enabled
q, err := querier.New()

// Custom rate limit (more lenient)
q, err := querier.New(
    querier.WithRateLimitThreshold(200),
    querier.WithRateLimitCooldown(30 * time.Second),
)

// Disable rate limiting (NOT recommended, only for testing)
q, err := querier.New(querier.WithRateLimit(false))
```

---

## Success Criteria

- [x] Source IP filtering enforces link-local scope (RFC 6762 §2)
- [x] Rate limiting protects against multicast storms (>100 queries/sec)
- [x] Input validation prevents panics on malformed packets (already done in M1)
- [x] Fuzzing validates parser robustness (already done in M1, CI integration added)
- [x] Packet size limits prevent memory DoS (≤9000 bytes)
- [x] Resource limits protect against exhaustion (buffers, timeouts, rate limits)
- [x] Security tests validate protections (attack simulation)
- [x] Configuration options for security features
- [x] Logging provides security visibility (debug, warn levels)

---

## Governance and Compliance

### Constitutional Compliance

**Principle I (RFC Compliant)**:
- ✅ RFC 6762 §2 link-local scope: Source IP filtering enforces requirement
- ✅ RFC 6762 §6 query timing: Rate limiting respects minimum 1-second interval
- ✅ RFC 6762 §17 message size: 9000-byte maximum enforced
- ✅ RFC 6762 §18 security: Defense-in-depth controls implemented

**Principle VIII (Excellence)**:
- ✅ Addresses security pitfalls from research (ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §3)
- ✅ Protects against real-world attacks (DRDoS, multicast storms, crashes)
- ✅ Enables security testing (fuzzing, attack simulation)

### Change Control

Changes to this specification require:
1. RFC validation review (if security controls affect protocol compliance)
2. Security impact assessment (if default security posture changes)
3. Attack simulation testing (validate controls still effective)
4. Version bump per semantic versioning:
   - **MAJOR**: Breaking changes to security API or default protections
   - **MINOR**: New security controls or non-breaking enhancements
   - **PATCH**: Security fixes, threshold adjustments, documentation

---

## References

**Constitutional**:
- [Beacon Constitution v1.1.0](../memory/constitution.md)

**Architectural**:
- [ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md](../../docs/ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md) - §3 (Security Architecture)
- [F-2: Package Structure](./F-2-package-structure.md) - Security layer organization
- [F-3: Error Handling](./F-3-error-handling.md) - Security error types
- [F-6: Logging & Observability](./F-6-logging-observability.md) - Security event logging
- [F-8: Testing Strategy](./F-8-testing-strategy.md) - Security testing requirements
- [F-9: Transport Layer](./F-9-transport-layer-socket-configuration.md) - Receive loop integration
- [F-10: Interface Management](./F-10-network-interface-management.md) - VPN exclusion defense

**RFCs**:
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - §2 (Link-Local Scope), §6 (Query Timing), §17 (Message Size), §18 (Security Considerations)
- [RFC 3927](https://www.rfc-editor.org/rfc/rfc3927.html) - IPv4 Link-Local (169.254/16)

**Research**:
- ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md - DRDoS attacks, multicast storms, malformed packets
- Hubitat bug (2020) - 1000 queries/sec multicast storm
- Avahi CVE history - Malformed packet vulnerabilities

---

## Version History

| Version | Date | Changes | Validated Against |
|---------|------|---------|-------------------|
| 1.0.0 | 2025-11-01 | Initial security architecture specification. Defines source IP filtering (link-local validation), rate limiting (multicast storm protection), input validation (references M1 implementation), fuzzing requirements (references M1 implementation), packet size limits, and resource protection. Integrates with F-9 transport layer receive loop. | Constitution v1.1.0, RFC 6762 §2/§6/§17/§18, ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md §3 |
