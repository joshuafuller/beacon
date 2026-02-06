# RFC 6762 Critical Requirements Quick Reference

**Project**: Beacon mDNS Library
**Generated**: 2026-01-06

---

## Purpose

This document provides a **quick reference** for the most critical RFC 6762 requirements that developers must understand when working on Beacon.

For comprehensive requirements, see `RFC_REQUIREMENTS_COMPLETE.md`.

---

## Critical P0 (MUST) Requirements

### 🌐 Network Layer

#### Multicast Addressing (§3, §4)
```
MUST: Send queries for .local to 224.0.0.251:5353 (IPv4) or FF02::FB (IPv6)
MUST: Send reverse queries for 169.254/16 to mDNS multicast address
MUST: Send reverse queries for IPv6 link-local to mDNS multicast address
```

**Implementation**: `internal/transport/udp.go`, `internal/protocol/mdns.go`
```go
const (
    MulticastAddrIPv4 = "224.0.0.251"
    MulticastAddrIPv6 = "FF02::FB"
    Port = 5353
)
```

**Test**: `internal/transport/udp_test.go:TestUDPv4Transport_SendToMulticast`

---

#### Source Port Rules (§5.1)

```
MUST NOT: Use UDP source port 5353 for one-shot queries
MUST: Use UDP source port 5353 for continuous queries (signals full mDNS querier)
```

**Implementation**: `querier/querier.go`
```go
// One-shot queries use ephemeral port (handled by OS)
// Continuous queries bind to port 5353
```

**Test**: `querier/querier_test.go`

---

### 🔍 Probing & Conflict Detection (§8)

#### Probing Before Announcing (§8.1)

```
MUST: Probe before claiming ownership of unique resource record set
MUST: Send 3 probe queries, each at least 250ms apart
MUST: Suppress probing on interface where address is already in use
```

**Implementation**: `internal/state/prober.go`
```go
const (
    ProbeCount = 3
    ProbeInterval = 250 * time.Millisecond
)

func (p *Prober) Start(ctx context.Context) error {
    for i := 0; i < ProbeCount; i++ {
        // Send probe query
        // Wait for responses
        // Check for conflicts
        time.Sleep(ProbeInterval)
    }
}
```

**Test**: `internal/state/prober_test.go:TestProber_ThreeProbeSequence`

---

#### Simultaneous Probe Tie-Breaking (§8.2)

```
MUST: Compare probe data lexicographically when simultaneous probes detected
MUST: Defer if probe data is lexicographically later
MUST: Continue if probe data is lexicographically earlier
```

**Implementation**: `responder/conflict_detector.go`
```go
func (cd *ConflictDetector) Compare(ours, theirs []byte) int {
    // Lexicographic comparison of rdata
    return bytes.Compare(ours, theirs)
}
```

**Test**: `responder/conflict_detector_test.go:TestConflictDetector_Tiebreaking`

---

#### Announcing (§8.3)

```
MUST: Announce after successful probing
MUST: Send unsolicited announcement containing all records
MUST: Send announcement twice, at least 1 second apart
```

**Implementation**: `internal/state/announcer.go`
```go
const (
    AnnouncementCount = 2
    AnnouncementInterval = 1 * time.Second
)

func (a *Announcer) Announce(ctx context.Context) error {
    for i := 0; i < AnnouncementCount; i++ {
        // Send announcement
        time.Sleep(AnnouncementInterval)
    }
}
```

**Test**: `internal/state/announcer_test.go:TestAnnouncer_TwoAnnouncements`

---

### 📡 Responding (§6)

#### Query Response Generation

```
MUST: Respond to queries for advertised records
MUST: Include all relevant records (PTR, SRV, TXT, A)
MUST: Delay response by random 20-120ms to avoid collisions
MUST NOT: Answer queries for records not owned
```

**Implementation**: `internal/responder/response_builder.go`
```go
func (rb *ResponseBuilder) BuildResponse(query *Query) (*Response, error) {
    // Check ownership
    // Build answer section (PTR, SRV, TXT, A)
    // Add additional records
    // Apply random delay
    delay := randomDelay(20*time.Millisecond, 120*time.Millisecond)
    time.Sleep(delay)
    return response, nil
}
```

**Test**: `internal/responder/response_builder_test.go`

---

#### QU (Unicast-Response) Bit Handling (§5.4)

```
MUST: Check QU bit in questions
SHOULD: Respond via unicast if QU bit set AND record not recently multicast
SHOULD: Respond via multicast if record not multicast in last quarter TTL
```

**Implementation**: `responder/responder.go`
```go
func (r *Responder) handleQuery(query *Query) error {
    if query.QU && !r.recentlyMulticast(record) {
        // Send unicast response
        return r.sendUnicastResponse(query.Source, response)
    }
    // Send multicast response
    return r.sendMulticastResponse(response)
}
```

**Test**: `responder/responder_test.go:TestResponder_QUBitHandling`

---

### 🗄️ Cache Coherency (§10)

#### Cache-Flush Bit (§10.2)

```
MUST: Set cache-flush bit (0x8000) in class field for unique records
MUST NOT: Set cache-flush bit for shared records
MUST: Flush cache entries when receiving record with cache-flush bit set
```

**Implementation**: `internal/records/record_set.go`
```go
const CacheFlushBit = 0x8000

func (rs *RecordSet) BuildA(unique bool) *ResourceRecord {
    class := uint16(ClassIN)
    if unique {
        class |= CacheFlushBit  // Set 0x8000
    }
    return &ResourceRecord{
        Class: class,
        // ... other fields
    }
}
```

**Test**: `internal/records/record_set_test.go:TestRecordSet_CacheFlushBit`

---

#### TTL Values (§10)

```
MUST: Use 120 seconds for host name (A/AAAA) records
MUST: Use 75 minutes (4500s) for service (PTR/SRV/TXT) records
MUST: Use TTL=0 for goodbye packets (service shutdown)
```

**Implementation**: `internal/records/ttl.go`
```go
const (
    HostRecordTTL    = 120      // 2 minutes
    ServiceRecordTTL = 4500     // 75 minutes
    GoodbyeTTL       = 0        // Goodbye packet
)
```

**Test**: `internal/records/ttl_test.go`

---

### 🚦 Traffic Reduction (§7)

#### Known-Answer Suppression (§7.1)

```
MUST: Implement Known-Answer Suppression for continuous queries
MUST: Include known answers in query's Answer Section
MUST: Suppress response if answer matches known answer
MUST: Increase query interval by at least factor of 2
```

**Implementation**: `internal/responder/known_answer.go`
```go
func (ka *KnownAnswerSuppressor) ShouldSuppress(answer *RR, knownAnswers []*RR) bool {
    for _, known := range knownAnswers {
        if answer.Name == known.Name &&
           answer.Type == known.Type &&
           answer.Class == known.Class &&
           bytes.Equal(answer.RData, known.RData) {
            return true  // Suppress
        }
    }
    return false
}
```

**Test**: `internal/responder/known_answer_test.go`

---

#### Continuous Query Timing (§5.2)

```
MUST: First query interval ≥ 1 second
MUST: Increase intervals by at least factor of 2
MAY: Cap interval at 60 minutes
SHOULD: Add random 20-120ms delay to initial query
```

**Implementation**: `querier/querier.go`
```go
func (q *Querier) continuousQuery(ctx context.Context) {
    interval := 1 * time.Second
    maxInterval := 60 * time.Minute

    for {
        q.sendQuery()
        time.Sleep(interval)

        interval *= 2  // Exponential backoff
        if interval > maxInterval {
            interval = maxInterval
        }
    }
}
```

**Test**: `querier/querier_test.go:TestQuerier_ExponentialBackoff`

---

### 🔒 Security (§11, §21)

#### Source Address Validation (§11)

```
MUST: Verify source address is on local link
MUST: Ignore packets from non-link-local sources
SHOULD: Check source address subnet matches interface subnet
```

**Implementation**: `internal/security/source_filter.go`
```go
func (sf *SourceFilter) ValidateSource(srcAddr net.IP, iface *net.Interface) bool {
    // Check link-local
    if !srcAddr.IsLinkLocalUnicast() && !srcAddr.IsLinkLocalMulticast() {
        return false
    }

    // Check subnet match
    addrs, _ := iface.Addrs()
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipnet.Contains(srcAddr) {
                return true
            }
        }
    }

    return false
}
```

**Test**: `internal/security/source_filter_test.go`

---

#### Rate Limiting (§6.2)

```
MUST: Rate limit responses per interface
SHOULD: Limit to 1 response per second per interface for same question
```

**Implementation**: `internal/security/rate_limiter.go`
```go
const ResponseRateLimit = 1 * time.Second

func (rl *RateLimiter) AllowResponse(ifaceIndex int, question *Question) bool {
    key := fmt.Sprintf("%d:%s:%d", ifaceIndex, question.Name, question.Type)

    rl.mu.Lock()
    defer rl.mu.Unlock()

    lastSent, exists := rl.lastResponse[key]
    if exists && time.Since(lastSent) < ResponseRateLimit {
        return false  // Rate limited
    }

    rl.lastResponse[key] = time.Now()
    return true
}
```

**Test**: `internal/security/rate_limiter_test.go`

---

### 🔄 Conflict Resolution (§9)

#### Ongoing Conflict Detection

```
MUST: Continuously monitor for conflicts
MUST: Defend name by responding to queries
MUST: Cease using name if simultaneous probe with lower lexicographic data
MUST: Wait 5 seconds before retrying after losing conflict
```

**Implementation**: `internal/responder/conflict.go`
```go
const ConflictBackoff = 5 * time.Second

func (c *ConflictHandler) HandleConflict(record *RR, conflictData []byte) error {
    ours := record.RData
    theirs := conflictData

    if bytes.Compare(ours, theirs) > 0 {
        // We lose - cease using name
        c.unregister(record)

        // Wait before retry
        time.Sleep(ConflictBackoff)

        // Rename and re-probe
        return c.renameAndReprobe(record)
    }

    // We win - defend
    return c.defend(record)
}
```

**Test**: `internal/responder/conflict_test.go`

---

### 🌍 Multi-Interface (§14, §15)

#### Interface-Specific Addressing (§15)

```
MUST: Use interface-specific IP address when responding
MUST NOT: Advertise same IP on all interfaces
```

**Implementation**: `responder/responder.go` (007-interface-specific-addressing)
```go
func getIPv4ForInterface(interfaceIndex int) (net.IP, error) {
    iface, err := net.InterfaceByIndex(interfaceIndex)
    if err != nil {
        return defaultIP, nil  // Graceful fallback
    }

    addrs, _ := iface.Addrs()
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ip4 := ipnet.IP.To4(); ip4 != nil {
                return ip4, nil  // Interface-specific IP
            }
        }
    }

    return defaultIP, nil
}
```

**Test**: `responder/responder_test.go:TestGetIPv4ForInterface`

---

## Quick Checklist for New Features

When implementing new mDNS functionality:

- [ ] Queries use 224.0.0.251:5353
- [ ] One-shot queries avoid port 5353
- [ ] Unique records probed before announcing
- [ ] Probing includes 3 queries at 250ms intervals
- [ ] Tie-breaking uses lexicographic comparison
- [ ] Announcing sends 2 packets, 1 second apart
- [ ] Responses include cache-flush bit for unique records
- [ ] TTL values match RFC (120s host, 4500s service)
- [ ] Known-Answer Suppression implemented
- [ ] Source address validated (link-local only)
- [ ] Rate limiting per interface
- [ ] QU bit handled correctly
- [ ] Conflict detection ongoing
- [ ] Interface-specific IPs used

---

## Common Pitfalls

### ❌ Wrong: Announcing without probing
```go
// WRONG
func Register(service *Service) error {
    return r.announce(service)  // RFC violation!
}
```

### ✅ Correct: Probe then announce
```go
// CORRECT
func Register(service *Service) error {
    if err := r.probe(service); err != nil {
        return err
    }
    return r.announce(service)
}
```

---

### ❌ Wrong: No cache-flush bit
```go
// WRONG
rr.Class = ClassIN  // Missing cache-flush bit!
```

### ✅ Correct: Set cache-flush for unique
```go
// CORRECT
rr.Class = ClassIN | CacheFlushBit  // 0x8001
```

---

### ❌ Wrong: Fixed response timing
```go
// WRONG
func respond(query *Query) {
    time.Sleep(50 * time.Millisecond)  // Fixed delay
    sendResponse()
}
```

### ✅ Correct: Random delay
```go
// CORRECT
func respond(query *Query) {
    delay := randomDelay(20*time.Millisecond, 120*time.Millisecond)
    time.Sleep(delay)
    sendResponse()
}
```

---

## Testing Critical Requirements

### Contract Tests
Location: `tests/contract/rfc6762_test.go`

Example:
```go
func TestRFC6762_Section8_1_Probing(t *testing.T) {
    // Verify 3-probe sequence
    // Verify 250ms intervals
    // Verify probe format
}
```

### Integration Tests
Location: `tests/integration/`

Example:
```go
func TestProbeAnnounceSequence(t *testing.T) {
    // Full probe → announce flow
    // Verify network packets
    // Verify timing
}
```

---

## References

- **Complete Database**: `RFC_REQUIREMENTS_COMPLETE.md`
- **Compliance Summary**: `RFC_COMPLIANCE_SUMMARY.md`
- **Index**: `RFC_REQUIREMENTS_INDEX.md`
- **RFC 6762 Text**: `RFC Docs/RFC-6762-Multicast-DNS.txt`

---

**Last Updated**: 2026-01-06
**Status**: 100% P0 Compliance ✅
