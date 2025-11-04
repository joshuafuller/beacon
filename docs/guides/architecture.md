# Beacon Architecture Overview

**Audience**: Users and contributors who want to understand how Beacon works
**Reading time**: 10 minutes

---

## High-Level Architecture

Beacon is organized into two main public APIs and several internal layers:

```
┌─────────────────────────────────────────────────────────┐
│                    Your Application                      │
└────────────┬────────────────────────────────┬───────────┘
             │                                │
             ▼                                ▼
┌────────────────────────┐    ┌────────────────────────────┐
│   Querier (Discovery)  │    │  Responder (Announcement)  │
│  • Query()             │    │  • Register()              │
│  • Close()             │    │  • Unregister()            │
└────────────┬───────────┘    └──────────┬─────────────────┘
             │                           │
             └───────────┬───────────────┘
                         ▼
          ┌──────────────────────────────┐
          │      Internal Layers         │
          ├──────────────────────────────┤
          │  Transport (Network I/O)     │
          │  Message (DNS Encoding)      │
          │  Protocol (RFC Constants)    │
          │  Security (Validation)       │
          └──────────────────────────────┘
                         │
                         ▼
          ┌──────────────────────────────┐
          │   Network (UDP Multicast)    │
          │   224.0.0.251:5353 (IPv4)    │
          └──────────────────────────────┘
```

---

## Core Components

### 1. Querier (Service Discovery)

**Purpose**: Find services on the local network

**How it works**:
1. Sends multicast DNS queries to 224.0.0.251:5353
2. Listens for responses from devices on the network
3. Parses DNS responses and returns structured results

**Key features**:
- Context-aware (supports cancellation and timeouts)
- Thread-safe (multiple concurrent queries)
- Buffer pooling (99% allocation reduction for performance)

**Example**:
```go
q, _ := querier.New()
results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
```

**RFC Compliance**: Implements RFC 6762 §5 (Query Transmission)

---

### 2. Responder (Service Announcement)

**Purpose**: Announce services so others can discover them

**How it works**:
1. **Probing** - Checks if service name is already in use (RFC 6762 §8.1)
2. **Announcing** - Broadcasts service availability (RFC 6762 §8.3)
3. **Responding** - Answers queries from other devices (RFC 6762 §6)
4. **Conflict Resolution** - Handles name conflicts automatically (RFC 6762 §8.2)

**State Machine**:
```
         Register()
             │
             ▼
   ┌─────────────────┐
   │   PROBING       │  (250ms, checks for conflicts)
   │  3 probe queries│
   └────────┬────────┘
            │ No conflict
            ▼
   ┌─────────────────┐
   │  ANNOUNCING     │  (Sends 2 announcements, 1 second apart)
   └────────┬────────┘
            │
            ▼
   ┌─────────────────┐
   │   ANNOUNCED     │  (Responds to queries)
   └─────────────────┘
```

**Key features**:
- Automatic conflict detection and resolution
- Rate limiting (max 1 response/sec per record)
- Known-answer suppression (doesn't respond if querier already knows)
- Multi-service support (register many services per responder)

**Example**:
```go
r, _ := responder.New(ctx)
svc := &responder.Service{
    Instance: "My Service",
    Service:  "_http._tcp",
    Port:     8080,
}
r.Register(ctx, svc)
```

**RFC Compliance**: Implements RFC 6762 §6-8 (Responding, Conflict Resolution)

---

## Internal Architecture

Beacon follows **Clean Architecture** principles with strict layer boundaries.

### Layer 1: Transport (Network Abstraction)

**File**: `internal/transport/`

**Purpose**: Abstract network I/O from protocol logic

**Key insight**: By abstracting the network layer, we can:
- Support multiple network types (IPv4 now, IPv6 future)
- Write unit tests without real sockets (using MockTransport)
- Implement platform-specific optimizations

**Interface**:
```go
type Transport interface {
    Send(ctx context.Context, packet []byte, dest net.Addr) error
    Receive(ctx context.Context) ([]byte, net.Addr, error)
    Close() error
}
```

**Implementations**:
- `UDPv4Transport` - Production IPv4 multicast (224.0.0.251:5353)
- `MockTransport` - Test double for unit testing

**Performance optimization**: Uses `sync.Pool` for 9KB receive buffers (99% allocation reduction)

**See**: [ADR-001: Transport Interface Abstraction](../internals/architecture/decisions/001-transport-interface-abstraction.md)

---

### Layer 2: Message (DNS Protocol)

**File**: `internal/message/`

**Purpose**: Build and parse DNS messages (RFC 1035 wire format)

**Responsibilities**:
- **Builder** - Construct DNS query/response packets
- **Parser** - Parse incoming DNS packets
- **Name Encoding** - DNS name compression (RFC 1035 §4.1.4)

**Why separate?** DNS message handling is complex (compression, binary encoding, validation). Isolating it makes the code testable and maintainable.

**Example**:
```go
// Build a query
packet := builder.BuildQuery("_http._tcp.local", QueryTypePTR)

// Parse a response
records, _ := parser.ParseResponse(packet)
```

---

### Layer 3: Protocol (RFC Constants)

**File**: `internal/protocol/`

**Purpose**: Centralize all mDNS/DNS constants

**Why?** Hard-coding magic numbers (like 5353, 224.0.0.251) throughout the code is error-prone. Centralizing them:
- Makes code self-documenting
- Ensures consistency
- Simplifies RFC compliance verification

**Constants**:
```go
const (
    MDNSPort      = 5353
    MDNSGroupIPv4 = "224.0.0.251"

    // Query types
    TypePTR = 12
    TypeSRV = 33
    TypeTXT = 16
    TypeA   = 1

    // Class codes
    ClassIN          = 1
    ClassCacheFlush  = 0x8001  // High bit set (RFC 6762 §10.2)
)
```

---

### Layer 4: Security (Validation & Rate Limiting)

**File**: `internal/security/`

**Purpose**: Protect against malformed input and abuse

**Validation** (RFC 6763 §4):
- Service names ≤63 characters
- Instance names ≤63 characters
- Domain is "local"
- Port 1-65535
- TXT records ≤255 bytes each

**Rate Limiting** (RFC 6762 §6.2):
- Max 1 multicast response per second per record
- Prevents network flooding

**Why critical?** mDNS processes untrusted network input. Validation prevents crashes from malformed packets.

---

### Layer 5: Records (DNS Record Construction)

**File**: `internal/records/`

**Purpose**: Build DNS resource records (PTR, SRV, TXT, A)

**For each service, Beacon creates**:
1. **PTR record** - Points service type → instance
   ```
   _http._tcp.local → My Server._http._tcp.local
   ```

2. **SRV record** - Provides hostname and port
   ```
   My Server._http._tcp.local → myhost.local:8080
   ```

3. **TXT record** - Contains metadata
   ```
   My Server._http._tcp.local → ["version=1.0", "path=/api"]
   ```

4. **A record** - Resolves hostname → IP
   ```
   myhost.local → 192.168.1.100
   ```

**TTL values** (RFC 6762 §10):
- Host records (A): 120 seconds
- Service records (SRV, TXT): 4500 seconds (75 minutes)
- Pointer records (PTR): 4500 seconds

**See**: [ADR-005: DNS-SD TTL Values](../internals/architecture/decisions/005-dns-sd-ttl-values.md)

---

### Layer 6: State Machine (Probing & Announcing)

**File**: `internal/state/`

**Purpose**: Implement RFC 6762 §8 (Probing, Announcing, Conflict Resolution)

**State transitions**:

```
IDLE
  │
  │ Register()
  ▼
PROBING (250ms)
  │ Send 3 probe queries
  │ • Query for our name
  │ • Check if anyone responds
  │ • Wait 250ms between probes
  │
  ├─ Conflict detected ──> Rename & retry
  │                        (append number: "Service (2)")
  │
  │ No conflict
  ▼
ANNOUNCING
  │ Send 2 unsolicited responses
  │ • 1 second apart
  │ • Broadcast our records
  │
  ▼
ANNOUNCED
  │ Normal operation
  │ • Respond to queries
  │ • Maintain registration
  │
  │ Unregister()
  ▼
GOODBYE
  │ Send goodbye packet (TTL=0)
  ▼
IDLE
```

**Why a state machine?** RFC 6762 §8 specifies precise timing and behavior. A state machine makes this explicit and testable.

---

## Design Principles

Beacon is built on five constitutional principles:

### 1. Protocol Compliance First

**Principle**: RFC 6762 compliance is non-negotiable

**How we enforce it**:
- 36 contract tests validate RFC behavior
- 72.2% RFC compliance (91/126 requirements implemented)
- Every feature references specific RFC sections
- Automated Semgrep rules check for violations

### 2. Zero External Dependencies

**Principle**: Standard library only

**Why?**
- No supply chain risk
- Smaller binary size
- Faster builds
- Simpler deployment

**Current dependencies**: None (only `golang.org/x/sys` for platform-specific socket options, considered "extended stdlib")

### 3. Context-Aware Operations

**Principle**: All blocking operations accept `context.Context`

**Why?**
- Enables cancellation
- Supports timeouts
- Integrates with Go patterns

**Examples**:
```go
// Timeout support
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
results, err := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)

// Cancellation support
ctx, cancel := context.WithCancel(context.Background())
go func() {
    results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
}()
cancel()  // Query stops immediately
```

### 4. Clean Architecture

**Principle**: Strict layer boundaries (enforced by import rules)

**Enforcement**:
```bash
# This must return 0 matches
grep -rn "internal/network" querier/
```

**Why?** Prevents circular dependencies, makes code testable, enables future refactoring.

### 5. Test-Driven Development

**Principle**: Tests written FIRST (RED → GREEN → REFACTOR)

**Current metrics**:
- 81.3% test coverage
- 247 tests
- 109,471 fuzz executions (0 crashes)
- 0 known data races

---

## Performance Characteristics

### Benchmarks (Post M1-Refactoring)

| Metric | Value | Context |
|--------|-------|---------|
| Query Latency | 163 ns/op | Time to build+send query |
| Response Latency | 4.8 μs | Time to respond to query |
| Allocations | 48 B/op | Per receive call |
| Concurrent Queries | 100+ | Tested and validated |

### Optimizations Applied

1. **Buffer Pooling** (ADR-002)
   - Problem: 9KB allocation per receive (900KB/sec at 100 req/sec)
   - Solution: `sync.Pool` for buffer reuse
   - Result: 99% allocation reduction (9000 B/op → 48 B/op)

2. **Transport Abstraction** (ADR-001)
   - Problem: Querier tightly coupled to UDP sockets
   - Solution: Transport interface
   - Result: Testable without real network, enables IPv6

3. **DNS Name Compression**
   - Problem: Repeated domain names waste bandwidth
   - Solution: RFC 1035 §4.1.4 name compression
   - Result: Smaller packets, faster transmission

---

## Platform Support

### Linux (Fully Supported ✅)

- All features implemented and tested
- Works with Avahi (SO_REUSEPORT enabled)
- Integration tests run on Linux CI

### macOS (Code Complete ⚠️)

- All code implemented
- SO_REUSEPORT configured for Bonjour coexistence
- Integration tests pending (requires macOS CI runner)

### Windows (Code Complete ⚠️)

- All code implemented
- SO_REUSEPORT configured for mDNS service coexistence
- Integration tests pending (requires Windows CI runner)

**Current limitation**: Socket configuration validated on Linux only. macOS/Windows code is complete but not integration-tested.

---

## Security Architecture

### Input Validation

**All user input is validated** before processing:

```go
// Service validation (responder)
- Instance name: 1-63 chars, valid UTF-8
- Service type: Valid format (_name._tcp or _udp)
- Domain: Must be "local"
- Port: 1-65535
- TXT records: ≤255 bytes each
```

**Network input is never trusted**:
- All DNS packets are validated before parsing
- Malformed packets return `WireFormatError` (never panic)
- No `unsafe` package usage in parsers

### Rate Limiting

**RFC 6762 §6.2 compliance**:
- Max 1 multicast response per second per record
- Prevents network flooding
- Per-interface tracking

### Resource Management

**All resources are bounded**:
- Socket buffers: 9KB (DNS max packet size)
- Context timeouts: User-specified
- Goroutine lifecycle: Tied to responder lifetime

**No resource leaks**:
- 25 Semgrep rules enforce proper resource cleanup
- All tests run with race detector
- Integration tests verify socket cleanup

---

## Coexistence with System Services

Beacon uses **SO_REUSEPORT** to share port 5353 with system mDNS daemons:

### How it Works

```
Port 5353 (Shared)
├─ Avahi/Bonjour (system daemon)
├─ Your Beacon Application #1
├─ Your Beacon Application #2
└─ Other mDNS clients
```

All processes can bind to the same port simultaneously.

### Benefits

- No need to disable system mDNS
- Multiple Beacon applications can run
- Plays nicely with existing services

### Platform Status

| Platform | System Daemon | Status |
|----------|---------------|--------|
| Linux | Avahi | ✅ Tested, works |
| macOS | Bonjour (mDNSResponder) | ⚠️ Code complete, pending tests |
| Windows | mDNS Service | ⚠️ Code complete, pending tests |

**See**: [Test it yourself](../../tests/manual/avahi_coexistence.go)

---

## Future Architecture Plans

### IPv6 Support (v0.2.0)

- Dual-stack operation (IPv4 + IPv6 simultaneously)
- IPv6 multicast group: FF02::FB
- Per-interface transport binding

### Service Browsing (v0.4.0)

- High-level API for browsing service types
- Automatic PTR → SRV → A resolution
- Service change notifications

### Observability (Future)

- Structured logging (F-6 spec planned)
- Metrics and telemetry
- Query/response tracing

---

## Learning More

**For Users**:
- [Getting Started](getting-started.md) - Your first query/responder
- [Querier Guide](querier-guide.md) - Deep dive into discovery
- [Responder Guide](responder-guide.md) - Advanced announcement patterns
- [Advanced Usage](advanced-usage.md) - Performance tuning, production patterns

**For Contributors**:
- [Development Guide](../development/README.md) - Setting up dev environment
- [Contributing Code](../development/contributing-code.md) - How to contribute
- [Architecture Decision Records](../internals/architecture/decisions/) - Why we made key decisions

**For Deep Dives**:
- [RFC Compliance Matrix](../internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md) - Section-by-section implementation status
- [Performance Analysis](../internals/analysis/PERFORMANCE_ANALYSIS.md) - Detailed benchmarks
- [Security Audit](../../specs/006-mdns-responder/SECURITY_AUDIT.md) - Security posture assessment

---

**Questions?** [Open a discussion](https://github.com/joshuafuller/beacon/discussions) or see our [Troubleshooting Guide](troubleshooting.md).
