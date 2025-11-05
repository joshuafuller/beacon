# RFC Compliance Guide

**Purpose**: This document provides a comprehensive mapping of Beacon's implementation to RFC 6762 (Multicast DNS) and RFC 1035 (DNS) requirements, making the codebase extremely auditable.

**Last Updated**: 2025-11-05

---

## Overview

Beacon implements Multicast DNS (mDNS) per RFC 6762 for local network service discovery. Every piece of code in this project ties back to specific RFC requirements, functional requirements (FRs), or architectural decisions (ADRs).

**Primary Technical Authorities**:
- **RFC 6762**: Multicast DNS (PRIMARY - all mDNS behavior)
- **RFC 1035**: Domain Names - Implementation and Specification (DNS wire format)
- **RFC 6763**: DNS-Based Service Discovery (service naming, TXT records)
- **RFC 2782**: DNS SRV Records (service location)

---

## Package-by-Package RFC Mapping

###  `querier/` - mDNS Query Public API

**RFC Compliance**: RFC 6762 §§5-6 (Querying), RFC 1035 §§4.1-4.2 (DNS Messages)

**Purpose**: Provides high-level API for discovering services on the local network using mDNS queries.

**Key RFC Requirements**:
- **RFC 6762 §5**: Multicast group 224.0.0.251:5353 for IPv4
- **RFC 6762 §6**: Query message format (QR=0, OPCODE=0, etc.)
- **RFC 6762 §18**: mDNS-specific header field requirements
- **RFC 1035 §4.1**: DNS message wire format (header + question + answer)

**Functional Requirements**:
- FR-001 through FR-022: Query construction, sending, receiving, parsing
- NFR-001: Query processing overhead <100ms
- NFR-002: Support ≥100 concurrent queries

**Files**:
- `querier.go`: Main Querier type, Query() method, response aggregation
- `records.go`: Public ResourceRecord types, type-safe accessors (AsA, AsPTR, etc.)
- `options.go`: Functional options (WithTimeout, WithInterfaces, etc.)
- `doc.go`: Package documentation with usage examples

---

### `responder/` - mDNS Responder Public API

**RFC Compliance**: RFC 6762 §§8-10 (Responder Behavior), RFC 6763 (Service Discovery)

**Purpose**: Provides high-level API for registering services and responding to queries.

**Key RFC Requirements**:
- **RFC 6762 §8**: Probing (§8.1), Conflict Resolution (§8.2), Announcing (§8.3)
- **RFC 6762 §6**: Query response construction
- **RFC 6762 §7.1**: Known-Answer Suppression
- **RFC 6762 §10**: TTL values (120s for services, 4500s for hostnames)
- **RFC 6763 §4**: Service Instance Names (UTF-8, spaces allowed)
- **RFC 6763 §6**: Resource Record format (PTR, SRV, TXT, A)

**Functional Requirements**:
- US1: Service registration with probing/announcing
- US2: Conflict resolution with lexicographic tie-breaking
- US3: Query response with record construction
- US4: Cache coherency via known-answer suppression
- US5: Multi-service support

**Files**:
- `responder.go`: Main Responder type, Register/Unregister/Close methods
- `service.go`: Service definition, validation, renaming on conflict
- `conflict_detector.go`: RFC 6762 §8.2 tie-breaking algorithm
- `options.go`: Functional options (WithHostname, etc.)

---

### `internal/message/` - DNS Wire Format

**RFC Compliance**: RFC 1035 §§3-4 (DNS Message Format), RFC 6762 §18 (mDNS Extensions)

**Purpose**: Implements DNS message parsing and construction per RFC 1035 wire format.

**Key RFC Requirements**:
- **RFC 1035 §4.1.1**: DNS Header (12 bytes: ID, Flags, counts)
- **RFC 1035 §4.1.2**: Question Section (QNAME, QTYPE, QCLASS)
- **RFC 1035 §4.1.3**: Answer/Authority/Additional Sections (NAME, TYPE, CLASS, TTL, RDLENGTH, RDATA)
- **RFC 1035 §4.1.4**: Message Compression (pointer format 0xC0, max 256 jumps)
- **RFC 1035 §3.1**: Domain Name Encoding (label length prefix, max 63 bytes per label, max 255 bytes total)
- **RFC 6762 §18**: mDNS-specific header requirements (QR, AA, OPCODE, RCODE)
- **RFC 6763 §4.3**: Service instance name encoding (UTF-8, spaces allowed in instance name)

**Why This Exists**: DNS uses a compact binary wire format for efficiency. All mDNS messages must conform to this format for interoperability with other mDNS implementations (Avahi, Bonjour, etc.).

**Files**:
- `message.go`: DNSHeader, Question, Answer, DNSMessage types
- `builder.go`: BuildQuery(), BuildResponse() - wire format serialization
- `parser.go`: ParseMessage(), parseHeader(), parseQuestion(), parseAnswer() - wire format deserialization
- `name.go`: EncodeName(), DecodeName(), EncodeServiceInstanceName() - DNS name encoding/decoding

**Critical Implementation Details**:
- **Big-endian encoding**: All multi-byte fields use big-endian per RFC 1035
- **Compression pointers**: High 2 bits = 11 (0xC0) indicates pointer, remaining 14 bits = offset
- **Label encoding**: Each label prefixed by length byte (1-63), terminated by 0x00
- **Service names**: RFC 6763 allows UTF-8/spaces in instance portion, but not in service type portion

---

### `internal/protocol/` - mDNS Protocol Constants

**RFC Compliance**: RFC 6762 §§5,10,18 (Protocol Constants), RFC 1035 §3.2 (DNS Constants)

**Purpose**: Centralized source of truth for all mDNS protocol constants and validation logic.

**Key RFC Requirements**:
- **RFC 6762 §5**: Port 5353, multicast address 224.0.0.251 (IPv4)
- **RFC 6762 §8.1**: Probe interval 250ms, 3 probes
- **RFC 6762 §8.3**: Announcement interval 1s, 2 announcements
- **RFC 6762 §10**: TTL values (120s service, 4500s hostname)
- **RFC 6762 §18**: Header flags (QR, AA, TC, RD, OPCODE, RCODE)
- **RFC 1035 §3.2.2-3.2.4**: Record types (A=1, PTR=12, TXT=16, SRV=33), Classes (IN=1)
- **RFC 1035 §3.1**: Name constraints (max label 63 bytes, max name 255 bytes)

**Why This Exists**: Protocol constants are defined in one place to prevent hardcoding throughout the codebase. This makes RFC compliance auditable and makes it impossible to accidentally violate RFC requirements.

**Files**:
- `mdns.go`: All mDNS protocol constants (port, multicast address, TTLs, timings, flags)
- `validator.go`: Input validation per RFC requirements (ValidateName, ValidateRecordType, ValidateResponse)

**Design Decision**: All RFC timing constants (ProbeInterval, etc.) are defined here and referenced throughout the codebase. Per Constitutional Principle I, RFC MUST requirements cannot be configurable - they are hard-coded from protocol package.

---

### `internal/state/` - State Machine for Registration

**RFC Compliance**: RFC 6762 §8 (Probing and Announcing)

**Purpose**: Implements the service registration state machine per RFC 6762 §8.

**Key RFC Requirements**:
- **RFC 6762 §8.1**: Probing phase (3 probes, 250ms apart, ~750ms total)
- **RFC 6762 §8.2**: Conflict resolution via lexicographic comparison
- **RFC 6762 §8.3**: Announcing phase (2 announcements, 1s apart, ~1s total)
- **RFC 6762 §8**: State flow: Initial → Probing → Announcing → Established (or ConflictDetected)

**Why This Exists**: RFC 6762 requires a specific probing/announcing sequence before a service is considered "established." This state machine enforces that sequence and handles conflicts.

**State Flow**:
```
Initial → Probing (3 probes, 250ms apart)
       ↓
       ├─ No conflict → Announcing (2 announcements, 1s apart)
       │                    ↓
       │                Established
       │
       └─ Conflict detected → ConflictDetected
                              (caller renames and retries)
```

**Files**:
- `machine.go`: State machine orchestration, Run() method, state transitions
- `prober.go`: Probing phase implementation (3 probes, 250ms intervals)
- `announcer.go`: Announcing phase implementation (2 announcements, 1s intervals)
- `state.go`: State type definitions (Initial, Probing, Announcing, Established, ConflictDetected)

---

### `internal/records/` - Resource Record Construction

**RFC Compliance**: RFC 6762 §10 (TTLs), RFC 6763 §6 (Record Format), RFC 1035 §3.2-3.4 (RR Format)

**Purpose**: Constructs the complete set of DNS resource records (PTR, SRV, TXT, A) for a registered service.

**Key RFC Requirements**:
- **RFC 6763 §6**: Service instance advertised with PTR, SRV, TXT, and A/AAAA records
- **RFC 6762 §10**: TTL values (120s for service records, 4500s for hostname records)
- **RFC 6762 §10.2**: Cache-flush bit for unique records (SRV, TXT, A), not for shared records (PTR)
- **RFC 6763 §6**: TXT record encoding (length-prefixed key=value strings, mandatory 0x00 if empty)
- **RFC 2782**: SRV record format (priority, weight, port, target)
- **RFC 1035 §3.4.1**: A record format (4-byte IPv4 address)

**Why This Exists**: A registered service must announce multiple record types for full discovery. This package constructs the complete record set with correct TTLs, cache-flush bits, and wire format encoding.

**Files**:
- `record_set.go`: BuildRecordSet(), buildPTRRecord(), buildSRVRecord(), buildTXTRecord(), buildARecord()
- `ttl.go`: TTL constants and helpers (RFC 6762 §10 values)

**Record Set Example**:
```
Service: "My Printer" on "_ipp._tcp.local", port 631, IP 192.168.1.100

Generated records:
1. PTR:  _ipp._tcp.local → My Printer._ipp._tcp.local (TTL=120s, shared)
2. SRV:  My Printer._ipp._tcp.local → myhost.local:631 (TTL=120s, cache-flush)
3. TXT:  My Printer._ipp._tcp.local → ["version=1.0"] (TTL=120s, cache-flush)
4. A:    myhost.local → 192.168.1.100 (TTL=4500s, cache-flush)
```

---

### `internal/security/` - Input Validation and Rate Limiting

**RFC Compliance**: RFC 6762 §6.2 (Rate Limiting), RFC 1035 (Input Validation)

**Purpose**: Protects against malformed input and multicast storms.

**Key RFC Requirements**:
- **RFC 6762 §6.2**: "A Multicast DNS responder MUST NOT multicast a given resource record on a given interface until at least one second has elapsed since the last time that resource record was multicast on that particular interface."
- **RFC 6762 §6.2**: Exception for probe defense: 250ms minimum instead of 1 second
- **RFC 1035 §3.1**: DNS name constraints (labels ≤63 bytes, total name ≤255 bytes)
- **RFC 6762 §2**: Link-local scope (packets from public IPs should be rejected)

**Why This Exists**: mDNS is vulnerable to multicast storms and malformed input attacks. Rate limiting and input validation prevent resource exhaustion.

**Files**:
- `validation.go`: Input validation (DNS name format, record types, TTLs)
- `rate_limiter.go`: Per-record, per-interface rate limiting (1s minimum between multicasts)
- `source_filter.go`: Link-local source IP validation (reject public IPs)

---

### `internal/responder/` - Responder Internal Logic

**RFC Compliance**: RFC 6762 §§6-7 (Query Response), RFC 6763 §6 (Record Construction)

**Purpose**: Internal implementation of query handling and response construction.

**Key RFC Requirements**:
- **RFC 6762 §6**: Query response format (AA=1, QR=1, answer records + additional records)
- **RFC 6762 §7.1**: Known-Answer Suppression (don't send records client already has)
- **RFC 6762 §6.2**: Per-record, per-interface rate limiting

**Why This Exists**: Separation of public API (responder/) from internal implementation (internal/responder/) per F-2 Architecture Layers.

**Files**:
- `registry.go`: Thread-safe service registry (sync.RWMutex for concurrent access)
- `response_builder.go`: Query response construction (answer + additional sections)
- `known_answer.go`: Known-Answer Suppression logic (RFC 6762 §7.1)

---

### `internal/transport/` - Network Abstraction

**RFC Compliance**: RFC 6762 §5 (Multicast Group), Platform-specific socket options

**Purpose**: Abstracts UDP multicast socket operations for testability and IPv6 support.

**Key RFC Requirements**:
- **RFC 6762 §5**: Join multicast group 224.0.0.251:5353 (IPv4)
- Platform-specific: SO_REUSEPORT for Avahi/Bonjour coexistence (M1.1)

**Why This Exists**: Decouples querier/responder from network implementation (ADR-001). Enables IPv6 support (M2) and test doubles (MockTransport).

**Files**:
- `transport.go`: Transport interface (Send, Receive, Close)
- `udp.go`: UDPv4Transport implementation (production IPv4 multicast)
- `buffer_pool.go`: Buffer pooling (ADR-002, 99% allocation reduction)
- `mock_transport.go`: Test double for unit testing

---

## RFC Section Quick Reference

### RFC 6762 (Multicast DNS) - Most Referenced Sections

| Section | Title | Purpose | Files |
|---------|-------|---------|-------|
| §5 | Multicast DNS Message Format | Multicast group, port 5353 | querier/, transport/ |
| §6 | Querying | Query construction, response handling | querier/, internal/responder/ |
| §6.2 | Rate Limiting | 1s minimum between record multicasts | internal/security/, internal/records/ |
| §7.1 | Known-Answer Suppression | Don't send records client has | internal/responder/known_answer.go |
| §8 | Probing and Announcing | Service registration flow | internal/state/ |
| §8.1 | Probing | 3 probes, 250ms apart | internal/state/prober.go |
| §8.2 | Conflict Resolution | Lexicographic tie-breaking | responder/conflict_detector.go |
| §8.3 | Announcing | 2 announcements, 1s apart | internal/state/announcer.go |
| §10 | Resource Record TTL Values | 120s service, 4500s hostname | internal/protocol/mdns.go, internal/records/ |
| §10.2 | Cache-flush Bit | Unique records set bit 15 of class | internal/records/, internal/message/ |
| §17 | Maximum Packet Size | 9000 bytes | querier/querier.go (validation) |
| §18 | Header Fields and Flags | QR, AA, OPCODE, RCODE requirements | internal/message/, internal/protocol/ |

### RFC 1035 (DNS Wire Format) - Most Referenced Sections

| Section | Title | Purpose | Files |
|---------|-------|---------|-------|
| §3.1 | Name Space Definitions | Label encoding, length limits | internal/message/name.go, internal/protocol/ |
| §3.2 | RR Definitions | Record types (A, PTR, SRV, TXT) | internal/protocol/mdns.go |
| §4.1 | Message Format | Header + sections structure | internal/message/message.go |
| §4.1.1 | Header Section Format | 12-byte header, flags layout | internal/message/message.go, builder.go |
| §4.1.2 | Question Section Format | QNAME, QTYPE, QCLASS | internal/message/builder.go, parser.go |
| §4.1.3 | Resource Record Format | NAME, TYPE, CLASS, TTL, RDLENGTH, RDATA | internal/message/parser.go, builder.go |
| §4.1.4 | Message Compression | Pointer format (0xC0 + offset) | internal/message/name.go, parser.go |

### RFC 6763 (DNS-SD) - Most Referenced Sections

| Section | Title | Purpose | Files |
|---------|-------|---------|-------|
| §4 | Service Instance Names | UTF-8, spaces allowed in instance name | responder/service.go, internal/message/name.go |
| §4.3 | DNS Name Encoding | Service instance name encoding | internal/message/name.go |
| §6 | Service Resource Records | PTR, SRV, TXT, A record format | internal/records/record_set.go |
| §6 | TXT Record Format | key=value encoding, mandatory 0x00 if empty | internal/records/record_set.go |
| §9 | Service Enumeration | "_services._dns-sd._udp.local" | internal/responder/registry.go |

---

## Functional Requirement (FR) Index

Functional requirements are defined in `specs/*/spec.md` files. Here's a mapping:

### Querier FRs (specs/002-mdns-querier/spec.md)
- **FR-001 to FR-022**: Query construction, sending, parsing, validation
- **FR-023**: Response construction (responder)
- **FR-026 to FR-034**: Responder-specific requirements

### Responder FRs (specs/006-mdns-responder/spec.md)
- **US1**: Service registration with probing/announcing
- **US2**: Conflict resolution
- **US3**: Query response
- **US4**: Cache coherency
- **US5**: Multi-service support

---

## Architecture Decision Records (ADRs)

ADRs document WHY we made key architectural choices:

- **ADR-001**: Transport Interface Abstraction (`docs/decisions/001-transport-interface-abstraction.md`)
  - WHY: Decouples querier from network, enables IPv6 and testability
  - WHERE: `internal/transport/transport.go`

- **ADR-002**: Buffer Pooling Pattern (`docs/decisions/002-buffer-pooling-pattern.md`)
  - WHY: Eliminates 900KB/sec allocations (9KB per receive call)
  - WHERE: `internal/transport/buffer_pool.go`

- **ADR-003**: Integration Test Timing Tolerance (`docs/decisions/003-integration-test-timing-tolerance.md`)
  - WHY: Real mDNS traffic has variable timing
  - WHERE: `tests/integration/`

---

## How to Verify RFC Compliance

### 1. Code → RFC Tracing
Every significant function has comments referencing RFC sections:
```go
// buildPTRRecord constructs a PTR record per RFC 6763 §6.
//
// PTR record format:
//   - Name: _service._proto.local (e.g., "_http._tcp.local")
//   - Type: PTR (12)
//   - Class: IN (1)
//   - TTL: 120 seconds (service record per RFC 6762 §10)
//   - RDATA: instance._service._proto.local
//   - CacheFlush: false (PTR is a shared record per RFC 6762 §10.2)
```

### 2. RFC → Code Tracing
Use grep to find all references to a specific RFC section:
```bash
# Find all RFC 6762 §8.2 references
grep -r "RFC 6762 §8.2" .

# Find all references to conflict detection
grep -r "conflict" . | grep -i "rfc"
```

### 3. Contract Tests
`tests/contract/` contains RFC compliance tests:
- Each test validates a specific RFC requirement
- Test names reference RFC sections (e.g., `TestRFC6762_Section8_1_Probing`)
- Tests use actual mDNS wire format messages

### 4. Fuzz Tests
`tests/fuzz/` validates handling of malformed input:
- Fuzzes DNS message parsing (RFC 1035 compliance)
- Fuzzes name encoding/decoding
- Ensures no crashes on malformed packets

---

## Common RFC Compliance Questions

### Q: Why do we use 250ms for probing intervals?
**A**: RFC 6762 §8.1 REQUIRES 250ms: "the host should first verify that the hardware address is ready by sending a standard ARP Request for the desired IP address and then wait 250 milliseconds."

This is a MUST requirement, not configurable.

**Code**: `internal/protocol/mdns.go:ProbeInterval`, used by `internal/state/prober.go`

### Q: Why do we set the cache-flush bit on SRV records but not PTR records?
**A**: RFC 6762 §10.2 distinguishes between unique and shared records:
- **Unique records** (SRV, TXT, A): Only one instance exists for a given name → set cache-flush bit
- **Shared records** (PTR): Multiple instances can exist (multiple services of same type) → don't set cache-flush bit

**Code**: `internal/records/record_set.go:buildPTRRecord()` (CacheFlush: false), `buildSRVRecord()` (CacheFlush: true)

### Q: Why do we use 120s TTL for service records and 4500s for hostname records?
**A**: RFC 6762 §10 recommends:
- Service records (PTR, SRV, TXT) change more frequently → 120s
- Hostname records (A, AAAA) change less frequently → 4500s (75 minutes)

This balances responsiveness (short TTL for services) with network efficiency (long TTL for stable hostnames).

**Code**: `internal/protocol/mdns.go:TTLService`, `TTLHostname`, used by `internal/records/record_set.go`

### Q: Why do we validate that DNS labels are ≤63 bytes?
**A**: RFC 1035 §3.1 REQUIRES: "labels are restricted to 63 octets or less."

This is a fundamental DNS constraint, not specific to mDNS.

**Code**: `internal/message/name.go:EncodeName()`, `internal/protocol/mdns.go:MaxLabelLength`

### Q: Why do we use QR=0 for queries and QR=1 for responses?
**A**: RFC 6762 §18.2 REQUIRES:
- "In query messages the QR bit MUST be zero"
- "In response messages the QR bit MUST be one"

This distinguishes queries from responses at the header level.

**Code**: `internal/message/builder.go:buildQueryHeader()` (QR=0), `buildResponseHeader()` (QR=1)

---

## Compliance Verification Checklist

Before releasing any change:

- [ ] All RFC references in comments are accurate (section numbers, quotes)
- [ ] All timing constants match RFC requirements (250ms probing, 1s announcing, etc.)
- [ ] All TTL values match RFC 6762 §10 (120s service, 4500s hostname)
- [ ] All header flags match RFC 6762 §18 (QR, AA, OPCODE, RCODE)
- [ ] All DNS name encoding follows RFC 1035 §3.1 (label length, total length)
- [ ] All resource record formats follow RFC 1035 §3.2-3.4
- [ ] Cache-flush bit set correctly per RFC 6762 §10.2 (unique vs shared)
- [ ] Known-Answer suppression implemented per RFC 6762 §7.1
- [ ] Rate limiting implemented per RFC 6762 §6.2 (1s minimum, 250ms for probe defense)
- [ ] Conflict resolution follows RFC 6762 §8.2 (lexicographic comparison)
- [ ] All contract tests pass (`make test-contract`)
- [ ] All fuzz tests pass without crashes (`make test-fuzz`)

---

## Conclusion

Every line of code in Beacon traces back to an RFC requirement, functional requirement, or architectural decision. This guide provides the mapping to make the codebase extremely auditable.

**When adding new features**:
1. Identify the RFC requirement
2. Reference the RFC section in package and function comments
3. Add contract tests that validate RFC compliance
4. Update this guide with the new mapping

**When debugging**:
1. Find the RFC section that governs the behavior
2. Trace code references to that section
3. Verify behavior matches RFC requirements
4. Check contract tests for that section

This makes Beacon's RFC compliance transparent, auditable, and maintainable.
