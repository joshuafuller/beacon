# Beacon: Next Phase Implementation Plan

**Date**: 2025-11-05
**Current Status**: M2 (mDNS Responder) 94.6% Complete
**Purpose**: Research and planning for post-M2 development

---

## Executive Summary

This document outlines the next development phase for Beacon following the near-completion of M2 (mDNS Responder). Based on RFC analysis and current implementation gaps, **M3 (Service Discovery)** is recommended as the next priority, with **M4 (IPv6 Support)** as a parallel or subsequent milestone.

**Recommendation**: Implement **DNS-SD Client-Side Features** (browsing + resolution) as the highest-priority next phase.

---

## Current State Analysis

### What's Complete ✅

| Milestone | Status | Key Features |
|-----------|--------|--------------|
| **M1** | ✅ Complete | mDNS Querier (query-only, A/PTR/SRV/TXT records) |
| **M1-Refactoring** | ✅ Complete | Clean architecture, 99% allocation reduction, transport abstraction |
| **M1.1** | ✅ Complete | Socket configuration (SO_REUSEPORT), interface management, security features |
| **M2** | ✅ 94.6% (122/129) | Service registration, probing, announcing, conflict resolution, query response |

### Current Capabilities

**Responder (Server-Side) - M2**:
- ✅ Register mDNS services with probing and announcing (RFC 6762 §8)
- ✅ Automatic conflict resolution with lexicographic tie-breaking (RFC 6762 §8.2)
- ✅ Respond to queries with PTR/SRV/TXT/A records (RFC 6762 §6)
- ✅ Known-answer suppression (RFC 6762 §7.1)
- ✅ Per-interface rate limiting (RFC 6762 §6.2)
- ✅ Multi-service support and service enumeration (RFC 6763 §9)

**Querier (Query-Only) - M1**:
- ✅ Send one-shot mDNS queries
- ✅ Parse responses (A/PTR/SRV/TXT records)
- ✅ Response validation and error handling
- ✅ Basic deduplication

### Critical Gaps ⚠️

**Missing Features**:
1. ❌ **Service Discovery (Client-Side DNS-SD)** - Cannot browse or continuously monitor services
2. ❌ **IPv6 Support** - IPv4 only (224.0.0.251), no FF02::FB or AAAA records
3. ⚠️ **Goodbye Packets** - Logic exists, TTL=0 wire format missing
4. ❌ **Unicast Response (QU bit)** - Always multicast
5. ❌ **Service Subtypes** - Cannot narrow service browsing
6. ❌ **Structured Logging** - Basic logging only (F-6 spec exists, not implemented)

---

## RFC Compliance Gaps

### RFC 6763: DNS-Based Service Discovery

| Section | Feature | Status | Priority |
|---------|---------|--------|----------|
| **§4** | Service Instance Enumeration (Browsing) | ❌ NOT IMPLEMENTED | **P0** |
| §4.1 | Structured service instance names | ✅ Validation exists | - |
| §4.2 | User interface presentation | ❌ NOT IMPLEMENTED | P3 |
| §4.3 | Internal name handling | ✅ Implemented | - |
| **§5** | Service Instance Resolution | ❌ NOT IMPLEMENTED | **P0** |
| | SRV record resolution | ❌ NOT IMPLEMENTED | P0 |
| | TXT record retrieval | ✅ Responder only | P0 |
| | Hostname resolution (A/AAAA) | ❌ NOT IMPLEMENTED | P0 |
| §7.1 | Service subtypes | ❌ NOT IMPLEMENTED | P2 |
| §11 | Domain enumeration | ❌ NOT IMPLEMENTED | P2 |

### RFC 6762: Multicast DNS

| Section | Feature | Status | Priority |
|---------|---------|--------|----------|
| **§5.4** | Unicast response support (QU bit) | ❌ NOT IMPLEMENTED | P2 |
| **§7.3** | Intelligent cache with TTL | ⚠️ PARTIAL (deduplication only) | P1 |
| **§9.4** | Goodbye packets (TTL=0) | ⚠️ PARTIAL (logic exists) | P1 |
| **§20** | IPv6 Considerations | ❌ NOT IMPLEMENTED | **P1** |
| | IPv6 multicast (FF02::FB) | ❌ NOT IMPLEMENTED | P1 |
| | AAAA records | ❌ NOT IMPLEMENTED | P1 |
| | Dual-stack operation | ❌ NOT IMPLEMENTED | P1 |

**Overall Compliance**:
- RFC 6762: 72.2% (13/18 sections)
- RFC 6763: 65% (major server-side features only)

**Target for M3+M4**: 85-90% combined RFC compliance

---

## Priority 0: Service Discovery (DNS-SD Client)

### Rationale

**Why P0:**
1. **Complements M2 Responder**: Beacon can register services but not discover them (incomplete workflow)
2. **Core DNS-SD Functionality**: RFC 6763 §3 Design Goals lists browsing + resolution as primary requirements
3. **Production Blocker**: Cannot build complete mDNS applications without client-side discovery
4. **User Demand**: Expected by anyone familiar with Avahi/Bonjour APIs

**Independent from IPv6**: Can be implemented entirely on IPv4 first, IPv6 can follow

### RFC Requirements Analysis

#### RFC 6763 §4: Service Instance Enumeration (Browsing)

**Core Concept**: PTR query to `<Service>.<Domain>` returns list of service instances

**Example Flow**:
```
Query:  PTR _http._tcp.local
Answer: PTR records pointing to:
        - "My Web Server._http._tcp.local"
        - "John's Server._http._tcp.local"
        - "Test API._http._tcp.local"
```

**Functional Requirements**:

**FR-M3-001**: System MUST support PTR queries for service type enumeration (RFC 6763 §4)
- **Input**: Service type (e.g., "_http._tcp.local")
- **Output**: List of service instance names
- **Mechanism**: PTR query to multicast address

**FR-M3-002**: System MUST support continuous service browsing (RFC 6762 §5.1)
- **Mechanism**: Source port 5353 for continuous queries
- **Behavior**: Receive all responses on the link, not just directed to this query
- **Use Case**: Live monitoring for service changes (add/remove)

**FR-M3-003**: System MUST handle PTR response with multiple instances
- **Parsing**: Multiple PTR records in single response
- **Deduplication**: Across multiple responses
- **TTL Tracking**: Cache instances with TTL management

**FR-M3-004**: System MUST detect service removal via goodbye packets (TTL=0)
- **Mechanism**: Receive PTR record with TTL=0
- **Action**: Remove from cache immediately (after 1s grace period)
- **Notification**: Notify application of removal event

**FR-M3-005**: System SHOULD support service enumeration metadata query (RFC 6763 §9)
- **Query**: PTR _services._dns-sd._udp.local
- **Response**: List of all advertised service types
- **Use Case**: Discover what kinds of services exist on network

#### RFC 6763 §5: Service Instance Resolution

**Core Concept**: Given a service instance name, resolve to connection info (host:port + metadata)

**Example Flow**:
```
Input:  "My Web Server._http._tcp.local"

Queries (parallel):
  1. SRV query → host=webserver.local, port=8080
  2. TXT query → metadata={path=/api, version=2.0}
  3. A query → IP=192.168.1.100

Output: {
  host: "webserver.local",
  ip: "192.168.1.100",
  port: 8080,
  txt: {path: "/api", version: "2.0"}
}
```

**Functional Requirements**:

**FR-M3-006**: System MUST support SRV record queries for service resolution (RFC 6763 §5)
- **Input**: Service instance name (e.g., "My Printer._ipp._tcp.local")
- **Output**: Priority, weight, port, target hostname
- **Parsing**: SRV record format per RFC 2782

**FR-M3-007**: System MUST support TXT record queries for service metadata (RFC 6763 §5)
- **Input**: Service instance name
- **Output**: Map of key=value pairs
- **Parsing**: TXT record format per RFC 6763 §6

**FR-M3-008**: System MUST support A record queries for hostname resolution
- **Input**: Target hostname from SRV record
- **Output**: IPv4 address
- **Already Exists**: M1 querier supports A records

**FR-M3-009**: System SHOULD query SRV, TXT, A in parallel for efficiency
- **Mechanism**: Send 3 queries simultaneously
- **Optimization**: Reduce total resolution time from ~300ms to ~100ms
- **RFC Guidance**: RFC 6763 §5 recommends parallel queries

**FR-M3-010**: System MUST support additional record processing (RFC 6763 §12)
- **Optimization**: When PTR response includes SRV/TXT/A in additional section, use them
- **Benefit**: Avoid separate queries (1 query instead of 3-4)
- **Behavior**: Most responders include SRV+TXT+A when answering PTR queries

**FR-M3-011**: System MUST handle missing or incomplete resolution data gracefully
- **Case 1**: No SRV record → return error, cannot connect
- **Case 2**: No TXT record → empty metadata (valid per RFC 6763 §6)
- **Case 3**: No A record → return error, cannot connect
- **Case 4**: Multiple SRV records → select by priority/weight (RFC 2782)

**FR-M3-012**: System SHOULD cache resolved service information with TTL
- **Mechanism**: Store {instance name → {host, port, txt}} with TTL
- **Refresh**: Query at 80%, 85%, 90%, 95% of TTL (RFC 6762 §10)
- **Invalidation**: Remove on TTL expiry or goodbye packet

#### Architecture Requirements

**FR-M3-013**: System MUST provide a `Browser` type for service discovery
- **Public API**: `browser/` package (new)
- **Usage**: `browser.Browse(serviceType) → stream of instances`
- **Pattern**: Similar to querier, but continuous monitoring

**FR-M3-014**: System MUST provide a `Resolver` type for service resolution
- **Public API**: `resolver/` package (new) or extend `browser/`
- **Usage**: `resolver.Resolve(instanceName) → ServiceInfo`
- **Pattern**: One-shot query for connection info

**FR-M3-015**: System SHOULD provide unified `ServiceInfo` type
- **Fields**: InstanceName, ServiceType, Hostname, Port, IPAddress, TXTRecords, TTL
- **Location**: `service/` package (new) or `browser/types.go`
- **Shared**: Between Browser and Resolver

**FR-M3-016**: System MUST support context-aware operations (F-9 compliance)
- **Browser.Browse(ctx, serviceType)**: Cancellation support
- **Resolver.Resolve(ctx, instanceName)**: Timeout support
- **Pattern**: All blocking operations accept context.Context

**FR-M3-017**: System SHOULD support event-driven browsing API
- **Mechanism**: Channel or callback for service add/remove events
- **Events**: ServiceAdded, ServiceRemoved, ServiceUpdated (TXT change)
- **Pattern**: `for event := range browser.Events() { ... }`

**FR-M3-018**: System MUST use existing M1 querier for low-level queries
- **Reuse**: `querier.Query()` for PTR/SRV/TXT/A queries
- **Abstraction**: Browser/Resolver wrap querier with DNS-SD logic
- **Benefit**: No duplication, leverage existing code

### User Stories

**US-M3-1: Service Browsing**
- **As a** developer building a networked application
- **I want to** discover all instances of a service type on the local network
- **So that** I can present a list of available services to the user

**Acceptance Criteria**:
1. Call `browser.Browse("_http._tcp.local")` returns list of web servers
2. New services appear in list within 2 seconds of announcement
3. Removed services disappear from list within 2 seconds of goodbye packet
4. Continuously monitor for changes without re-querying manually

**US-M3-2: Service Resolution**
- **As a** developer
- **I want to** resolve a service instance name to connection information
- **So that** I can establish a connection to the selected service

**Acceptance Criteria**:
1. Call `resolver.Resolve("My Printer._ipp._tcp.local")` returns host, port, IP, metadata
2. Resolution completes within 100ms (parallel queries)
3. Cached results used when available (avoid redundant queries)
4. Errors reported clearly (missing SRV/A records)

**US-M3-3: Live Service Monitoring**
- **As a** developer
- **I want to** be notified when services appear, disappear, or change
- **So that** I can keep my UI or application state in sync with network

**Acceptance Criteria**:
1. Browser emits ServiceAdded event when new service announces
2. Browser emits ServiceRemoved event on goodbye packet (TTL=0)
3. Browser emits ServiceUpdated event when TXT records change
4. Events delivered within 2 seconds of network change

### Success Criteria

**SC-M3-001**: Beacon can discover services registered by Avahi/Bonjour
- **Test**: Browse "_http._tcp.local", find Apache/nginx servers
- **Validation**: Cross-platform interoperability

**SC-M3-002**: Beacon can resolve service instances to connection info
- **Test**: Resolve Avahi-registered printer, connect successfully
- **Validation**: End-to-end workflow

**SC-M3-003**: Test coverage ≥80% for browser and resolver packages
- **Requirement**: Constitution mandate
- **Validation**: `make test-coverage`

**SC-M3-004**: Contract tests for RFC 6763 §4-5 compliance
- **Test**: `tests/contract/rfc6763_browsing_test.go`
- **Validation**: 10+ RFC compliance test cases

**SC-M3-005**: Browser can monitor 50+ services without performance degradation
- **Benchmark**: CPU <5%, memory <10MB
- **Validation**: Load test with Avahi daemons

### Implementation Estimate

**Complexity**: Medium (builds on M1 querier + M2 responder patterns)

**Phases**:
1. **Browser Implementation** (1-2 weeks)
   - PTR queries for browsing
   - Continuous monitoring (source port 5353)
   - Event-driven API
   - TTL-based cache

2. **Resolver Implementation** (1 week)
   - SRV/TXT/A parallel queries
   - ServiceInfo aggregation
   - Additional record optimization
   - Error handling

3. **Integration & Polish** (1 week)
   - Contract tests (RFC 6763 §4-5)
   - Interoperability tests (Avahi/Bonjour)
   - Documentation and examples
   - Performance tuning

**Total**: 3-4 weeks for full M3 implementation

---

## Priority 1: IPv6 Support

### Rationale

**Why P1 (not P0):**
1. **Modern Networking Requirement**: IPv6 adoption growing, especially in enterprise/IoT
2. **RFC Compliance**: Required for full RFC 6762 compliance (§20)
3. **Future-Proof**: Prepare for IPv6-only networks
4. **Can Parallelize**: Mostly orthogonal to M3 (service discovery)

**Interdependencies**: None with M3, can be developed in parallel or after

### RFC Requirements Analysis

#### RFC 6762 §20: IPv6 Considerations

**Core Concept**: Dual-stack hosts participate in *two* .local zones (IPv4 and IPv6)

**Key Insights from RFC**:
> "An IPv4-only host and an IPv6-only host behave as 'ships that pass in the night'. Even if they are on the same Ethernet, neither is aware of the other's traffic."

> "A dual-stack (v4/v6) host can participate in both '.local.' zones, and should register its name(s) and perform its lookups both using IPv4 and IPv6."

**Functional Requirements**:

**FR-M4-001**: System MUST support IPv6 multicast (FF02::FB) for queries and responses
- **Address**: `FF02::FB` (link-local scope)
- **Port**: 5353 (same as IPv4)
- **Implementation**: Dual transport layer (IPv4 + IPv6)

**FR-M4-002**: System MUST support AAAA records for IPv6 addresses
- **Record Type**: AAAA (type 28)
- **Parsing**: 16-byte IPv6 addresses
- **Already Partial**: Parser likely supports AAAA (check `internal/message/parser.go`)

**FR-M4-003**: System MUST register services on both IPv4 and IPv6 (dual-stack)
- **Behavior**: Send probes/announcements to both 224.0.0.251 and FF02::FB
- **Records**: Both A and AAAA records for hostname
- **NSEC**: Indicate both A and AAAA in NSEC records

**FR-M4-004**: System MUST query on both IPv4 and IPv6 (dual-stack)
- **Behavior**: Send queries to both multicast addresses
- **Aggregation**: Merge responses from both networks
- **Deduplication**: Handle same service appearing in both zones

**FR-M4-005**: System SHOULD support IPv6-only mode
- **Configuration**: `WithIPv6Only()` option
- **Use Case**: IPv6-only networks (growing in IoT/enterprise)
- **Validation**: Disable IPv4 transport entirely

**FR-M4-006**: System SHOULD support IPv4-only mode (default)
- **Configuration**: Default behavior, or `WithIPv4Only()` option
- **Benefit**: Backward compatibility with current code
- **Validation**: Existing behavior unchanged

**FR-M4-007**: System MUST handle dual-stack default (IPv4 + IPv6)
- **Configuration**: `WithDualStack()` (default for dual-stack hosts)
- **Behavior**: Parallel operation on both networks
- **Optimization**: Prefer IPv6 when available (modern best practice)

#### Architecture Requirements

**FR-M4-008**: System MUST abstract transport layer to support multiple IP versions
- **Already Exists**: M1-Refactoring created transport abstraction
- **Extension**: Add `UDPv6Transport` alongside `UDPv4Transport`
- **Interface**: Same `Transport` interface for both

**FR-M4-009**: System MUST extend record builders to generate AAAA records
- **Location**: `internal/records/record_set.go`
- **Function**: `BuildAAAARecord(hostname, ipv6addr, ttl)`
- **Integration**: Include AAAA in service registration (M2 responder)

**FR-M4-010**: System SHOULD detect host's IP version support
- **Mechanism**: Check for IPv6 interfaces at startup
- **Behavior**: Auto-enable dual-stack if IPv6 available
- **Fallback**: IPv4-only if no IPv6

**FR-M4-011**: System MUST support per-interface IPv6 configuration
- **Already Exists**: M1.1 interface management (F-10)
- **Extension**: Handle IPv6 link-local addresses per interface
- **Complexity**: IPv6 link-local scope requires interface binding

### User Stories

**US-M4-1: Dual-Stack Service Registration**
- **As a** service provider on a dual-stack network
- **I want to** register my service on both IPv4 and IPv6
- **So that** both IPv4-only and IPv6-only clients can discover me

**Acceptance Criteria**:
1. Service registered with both A and AAAA records
2. Probes/announcements sent to both 224.0.0.251 and FF02::FB
3. Service discoverable by IPv4-only and IPv6-only browsers
4. Performance: <2ms overhead for dual-stack vs IPv4-only

**US-M4-2: IPv6-Only Client Discovery**
- **As a** developer on an IPv6-only network
- **I want to** discover services using FF02::FB
- **So that** I can use mDNS without IPv4

**Acceptance Criteria**:
1. Browser discovers services on FF02::FB
2. Resolution returns AAAA records instead of A records
3. No IPv4 traffic generated
4. IPv6-only mode configurable via option

### Success Criteria

**SC-M4-001**: Services registered on dual-stack host discoverable by both IPv4 and IPv6 clients
- **Test**: Run Avahi browser in IPv4-only and IPv6-only mode
- **Validation**: Same service appears in both

**SC-M4-002**: Beacon operates correctly in IPv6-only environment
- **Test**: Disable IPv4, register + browse services
- **Validation**: All operations functional on FF02::FB

**SC-M4-003**: RFC 6762 §20 compliance validated
- **Test**: Contract test `tests/contract/rfc6762_ipv6_test.go`
- **Validation**: Dual-stack behavior per RFC

**SC-M4-004**: Test coverage ≥80% for IPv6 code paths
- **Requirement**: Constitution mandate
- **Validation**: `make test-coverage`

### Implementation Estimate

**Complexity**: Medium-High (network layer changes, platform differences)

**Phases**:
1. **Transport Layer Extension** (1 week)
   - `UDPv6Transport` implementation
   - Socket configuration for IPv6
   - Interface binding for link-local

2. **Record Type Support** (3-4 days)
   - AAAA record parsing/building
   - NSEC record updates (indicate A+AAAA)
   - Integration with M2 responder

3. **Dual-Stack Orchestration** (1 week)
   - Parallel IPv4/IPv6 operation
   - Response aggregation/deduplication
   - Mode selection (IPv4-only, IPv6-only, dual-stack)

4. **Testing & Validation** (1 week)
   - Contract tests (RFC 6762 §20)
   - Platform testing (Linux IPv6 well-supported, macOS/Windows TBD)
   - Interoperability with Avahi/Bonjour

**Total**: 3-4 weeks for full M4 implementation

**Platform Concerns**:
- **Linux**: IPv6 support mature, likely smooth
- **macOS**: IPv6 support good, Bonjour already dual-stack
- **Windows**: IPv6 support varies, testing required

---

## Priority 2: Advanced Features

### Lower-Priority Enhancements

**These can be deferred to M5+ or addressed based on user demand.**

#### 1. Service Subtypes (RFC 6763 §7.1)

**Purpose**: Narrow service browsing to specific subsets

**Example**:
```
Browse: _printer._sub._http._tcp.local
  (instead of all _http._tcp services)
```

**Complexity**: Low (syntactic sugar on existing browsing)
**Value**: Medium (useful for specialized clients)

**Functional Requirements**:
- FR-M5-001: Support browsing service subtypes via `_<subtype>._sub._<service>._<proto>`
- FR-M5-002: Responder should advertise subtypes via additional PTR records

#### 2. Unicast Response Support (RFC 6762 §5.4 - QU Bit)

**Purpose**: Request unicast response instead of multicast (reduce network traffic)

**Mechanism**: Set top bit of QCLASS (0x8000) in query

**Already Supported**: Responder can detect QU bit (M2), but always multicasts
**Missing**: Actually send unicast response to querier's source address

**Complexity**: Low (responder change only)
**Value**: Medium (optimization, not functional requirement)

**Functional Requirements**:
- FR-M5-003: Responder sends unicast response when QU bit set in query
- FR-M5-004: Exception: Still multicast if TTL <1/4 of correct value (RFC 6762 §5.4)

#### 3. Goodbye Packets (RFC 6762 §9.4 - Wire Format)

**Status**: Logic exists in M2, but TTL=0 packets not sent on wire

**Complexity**: Very Low (already 90% done)
**Value**: High (graceful service removal)

**Functional Requirements**:
- FR-M5-005: Send TTL=0 for all records when service unregistered
- FR-M5-006: Responder sends goodbye on graceful shutdown (Close())

**Recommendation**: Complete this in M3 or as a quick M2.1 patch before M3

#### 4. Structured Logging (F-6 Spec)

**Status**: F-6 spec exists, not implemented (basic Go log package used)

**Complexity**: Medium (new dependency decision)
**Value**: High for production deployments

**Functional Requirements**:
- FR-M5-007: Support structured logging (JSON format)
- FR-M5-008: Log levels: DEBUG, INFO, WARN, ERROR
- FR-M5-009: Contextual logging (request ID, service name, etc.)

**Dependency Dilemma**: Requires logging library (slog in Go 1.21+, or third-party)
**Constitution**: Minimize external dependencies (Principle V)
**Recommendation**: Use Go 1.21+ `log/slog` (standard library)

#### 5. Domain Enumeration (RFC 6763 §11)

**Purpose**: Discover browsing/registration domains (beyond .local)

**Complexity**: Medium
**Value**: Low (mostly for Unicast DNS-SD, rare in mDNS)

**Recommendation**: Defer to M6+ or wait for user requests

---

## Recommended Roadmap

### Immediate Next Steps (Post-M2)

**Short-Term (Next 1-2 months)**:

1. **Complete M2 Documentation** (1 week)
   - T123-T126: Update CLAUDE.md, RFC_COMPLIANCE_MATRIX.md, examples
   - Optional: T116-T117 if macOS/Avahi environments available

2. **M2.1: Quick Wins** (1 week)
   - Goodbye packets wire format (FR-M5-005, FR-M5-006)
   - QU bit unicast response (FR-M5-003, FR-M5-004)
   - Fixes any issues found in M2 documentation review

3. **M3: Service Discovery** (3-4 weeks)
   - Browser: PTR queries, continuous monitoring, event-driven API
   - Resolver: SRV/TXT/A resolution, parallel queries, ServiceInfo
   - Contract tests: RFC 6763 §4-5 compliance
   - Interoperability: Avahi/Bonjour integration tests

**Medium-Term (3-5 months)**:

4. **M4: IPv6 Support** (3-4 weeks)
   - UDPv6Transport: FF02::FB multicast
   - AAAA records: Parsing + generation
   - Dual-stack: Parallel IPv4/IPv6 operation
   - Contract tests: RFC 6762 §20 compliance

5. **M5: Production Hardening** (2-3 weeks)
   - Structured logging (F-6 implementation with log/slog)
   - Observability: Metrics, traces (optional)
   - Performance tuning: Cache improvements, TTL refresh
   - Platform testing: macOS, Windows validation

**Long-Term (6+ months)**:

6. **M6: Advanced Features** (as needed)
   - Service subtypes (RFC 6763 §7.1)
   - Domain enumeration (RFC 6763 §11)
   - Additional record types (NSEC, etc.)
   - Community-requested features

### Milestone Dependencies

```
M1 (Querier) ────┐
M1-Refactoring ──┤
M1.1 (Hardening) ┤
M2 (Responder) ──┴──► M2.1 (Quick Wins) ──┬──► M3 (Service Discovery) ─┬──► M5 (Production)
                                            │                            │
                                            └──► M4 (IPv6 Support) ──────┘
```

**Parallelization Opportunities**:
- M3 and M4 can be developed concurrently (mostly independent)
- M2.1 can be done while M3 is being planned
- M5 can overlap with M3/M4 (logging is orthogonal)

---

## Implementation Approach

### Constitution Compliance

All milestones MUST follow [Beacon Constitution v1.1.0](/.specify/memory/constitution.md):

1. **RFC Compliant (Principle I)**: All features validated against RFC 6762/6763
2. **Spec-Driven (Principle II)**: Use `/speckit.specify` for M3/M4 specifications
3. **TDD (Principle III)**: RED → GREEN → REFACTOR for all features
4. **Phased (Principle IV)**: M3/M4 delivered incrementally with working code
5. **Minimal Dependencies (Principle V)**: Use stdlib (log/slog for logging), golang.org/x/* only when justified

### Spec Kit Workflow

For each milestone (M3, M4, M5):

1. **Specify**: `/speckit.specify` → Create `specs/00X-milestone/spec.md`
   - User stories (US-M3-1, US-M3-2, etc.)
   - Functional requirements (FR-M3-001 through FR-M3-NNN)
   - Success criteria (SC-M3-001, etc.)

2. **Plan**: `/speckit.plan` → Generate `specs/00X-milestone/plan.md`
   - Architecture decisions (ADRs if needed)
   - Implementation phases
   - Task breakdown

3. **Implement**: `/speckit.implement` or manual TDD
   - Write tests first (RED)
   - Implement features (GREEN)
   - Refactor for quality (REFACTOR)

4. **Validate**: Completion report
   - All SC-XXX criteria met
   - All FR-XXX requirements implemented
   - Test coverage ≥80%
   - RFC compliance validated

### Testing Strategy

**Contract Tests** (RFC Compliance):
- `tests/contract/rfc6763_browsing_test.go` - §4 service enumeration
- `tests/contract/rfc6763_resolution_test.go` - §5 service resolution
- `tests/contract/rfc6762_ipv6_test.go` - §20 IPv6 support

**Integration Tests** (Interoperability):
- `tests/integration/avahi_browsing_test.go` - Browse Avahi services
- `tests/integration/bonjour_browsing_test.go` - Browse Bonjour services (macOS)
- `tests/integration/dual_stack_test.go` - IPv4+IPv6 coexistence

**Unit Tests** (Component Behavior):
- `browser/browser_test.go` - Browser API, PTR queries, event handling
- `resolver/resolver_test.go` - Resolver API, SRV/TXT/A aggregation
- `internal/transport/udpv6_test.go` - IPv6 transport

**Benchmark Tests** (Performance):
- `browser/browser_bench_test.go` - Browse 50+ services
- `resolver/resolver_bench_test.go` - Resolve with cache
- `internal/transport/transport_bench_test.go` - IPv4 vs IPv6 overhead

### Documentation Requirements

For each milestone:

1. **API Documentation**: Godoc for all exported types (browser.Browser, resolver.Resolver, etc.)
2. **User Guide**: `docs/guides/service-discovery.md`, `docs/guides/ipv6-support.md`
3. **Examples**: `examples/browser/`, `examples/resolver/`
4. **CLAUDE.md Updates**: Add M3/M4 sections with API usage, common patterns
5. **RFC_COMPLIANCE_MATRIX.md**: Update with new implemented sections

---

## Open Questions

### Technical Decisions Needed

1. **Browser API Design**: Channel-based or callback-based for events?
   - Option A: `for event := range browser.Events() { ... }`
   - Option B: `browser.OnServiceAdded(func(svc ServiceInfo) { ... })`
   - Recommendation: Option A (more idiomatic Go)

2. **Resolver Caching**: Should Resolver have built-in cache or rely on Browser cache?
   - Option A: Resolver caches independently (simpler)
   - Option B: Share cache with Browser (more efficient, complex)
   - Recommendation: Option A for MVP, Option B for optimization later

3. **IPv6 Default Behavior**: Enable dual-stack by default or require explicit opt-in?
   - Option A: Auto-detect and enable if IPv6 available (more magical)
   - Option B: Explicit `WithDualStack()` option (more explicit)
   - Recommendation: Option A (better UX, matches OS behavior)

4. **Logging Library**: Use Go 1.21+ log/slog or third-party?
   - Option A: log/slog (stdlib, requires Go 1.21+)
   - Option B: Third-party like zap/zerolog (more features, adds dependency)
   - Recommendation: Option A (Constitution Principle V: minimal dependencies)

### Research Needed

1. **IPv6 Platform Support**: Test IPv6 sockets on macOS/Windows
   - Verify SO_REUSEPORT works with FF02::FB
   - Check link-local scope binding requirements

2. **Avahi/Bonjour Behavior**: How do they handle dual-stack?
   - Do they query both IPv4 and IPv6 simultaneously?
   - How do they deduplicate responses?

3. **Cache Design**: What's the optimal TTL refresh strategy?
   - RFC 6762 §10 says 80%, 85%, 90%, 95%
   - Should we do all 4 or just 80% + 95%?

4. **Performance Impact**: Dual-stack overhead measurement
   - Latency increase with parallel IPv4/IPv6 queries
   - Memory overhead for dual transport

---

## Success Metrics

### M3 (Service Discovery) Success

**Functional**:
- ✅ 100% of Avahi-registered services discoverable via Browser
- ✅ ServiceInfo resolution completes in <100ms (parallel queries)
- ✅ Live monitoring detects add/remove within 2 seconds
- ✅ 10+ RFC 6763 §4-5 contract tests PASS

**Quality**:
- ✅ Test coverage ≥80%
- ✅ Zero data races (`go test -race`)
- ✅ RFC 6763 compliance: 80%+ (up from 65%)
- ✅ Documentation complete (API docs, examples, guides)

### M4 (IPv6) Success

**Functional**:
- ✅ Services registered on FF02::FB discoverable by IPv6-only clients
- ✅ Dual-stack operation: same service in both IPv4 and IPv6 zones
- ✅ AAAA records parsed and generated correctly
- ✅ 5+ RFC 6762 §20 contract tests PASS

**Quality**:
- ✅ Test coverage ≥80%
- ✅ Zero data races (`go test -race`)
- ✅ RFC 6762 compliance: 80%+ (up from 72.2%)
- ✅ Platform validation: Linux ✅, macOS ⚠️ (test), Windows ⚠️ (test)

### Combined M3+M4 Success

- ✅ RFC 6762 compliance: 85-90%
- ✅ RFC 6763 compliance: 80-85%
- ✅ Production-ready: Complete service discovery + dual-stack support
- ✅ Interoperability: Works with Avahi, Bonjour, and other mDNS implementations

---

## Conclusion

**Recommendation**: Proceed with **M3 (Service Discovery)** as the next priority development phase.

**Rationale**:
1. **Completes DNS-SD Workflow**: M2 provides server-side (responder), M3 provides client-side (browser/resolver)
2. **High User Value**: Cannot build full mDNS applications without service discovery
3. **RFC Compliance**: Critical gap in RFC 6763 (§4-5 not implemented)
4. **Foundation for M4**: IPv6 can build on M3's browser/resolver abstractions

**Timeline**: M3 can begin immediately after M2 documentation completion (~1 week). Estimated 3-4 weeks for full M3 implementation.

**Next Action**: Create M3 specification via `/speckit.specify` with user stories, functional requirements, and success criteria based on this research.

---

## Appendices

### A. RFC 6763 Section Summary

| Section | Title | M2 Status | M3 Scope |
|---------|-------|-----------|----------|
| §4 | Service Instance Enumeration (Browsing) | ❌ | ✅ Browser |
| §4.1 | Structured Service Instance Names | ✅ | - |
| §4.2 | User Interface Presentation | ❌ | ❌ (UI layer) |
| §4.3 | Internal Handling of Names | ✅ | - |
| §5 | Service Instance Resolution | ❌ | ✅ Resolver |
| §6 | Data Syntax for DNS-SD TXT Records | ✅ | - |
| §7 | Service Names | ✅ | - |
| §7.1 | Subtypes (Selective Enumeration) | ❌ | ❌ (M5) |
| §9 | Service Type Enumeration | ✅ | Enhance |
| §10 | Populating DNS (Service Registration) | ✅ | - |
| §11 | Domain Enumeration | ❌ | ❌ (M5) |
| §12 | Additional Record Generation | ✅ | Use in Resolver |

### B. Functional Requirements Summary

**M3 (Service Discovery)**:
- FR-M3-001 through FR-M3-018: 18 functional requirements
- Focus: Browser (PTR queries, continuous monitoring, events)
- Focus: Resolver (SRV/TXT/A resolution, parallel queries, cache)

**M4 (IPv6)**:
- FR-M4-001 through FR-M4-011: 11 functional requirements
- Focus: UDPv6Transport (FF02::FB multicast)
- Focus: AAAA records (parsing + generation)
- Focus: Dual-stack operation (IPv4 + IPv6 parallel)

**M5 (Advanced)**:
- FR-M5-001 through FR-M5-009: 9 functional requirements
- Focus: Service subtypes, unicast response (QU bit), goodbye packets, structured logging

### C. Reference Documents

**RFCs**:
- [RFC 6762: Multicast DNS](RFC Docs/RFC-6762-Multicast-DNS.txt) - §4, §5, §7, §20
- [RFC 6763: DNS-Based Service Discovery](RFC Docs/RFC-6763-DNS-SD.txt) - §4, §5, §6, §7, §9, §12

**Internal Specs**:
- [Beacon Constitution v1.1.0](.specify/memory/constitution.md)
- [BEACON_FOUNDATIONS v1.1](.specify/specs/BEACON_FOUNDATIONS.md)
- [RFC Compliance Matrix](docs/internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md)
- [M2 Completion Report](specs/006-mdns-responder/COMPLETION_REPORT.md)

**Architecture Specs**:
- [F-2: Package Structure](.specify/specs/F-2-package-structure.md)
- [F-3: Error Handling](.specify/specs/F-3-error-handling.md)
- [F-6: Logging & Observability](.specify/specs/F-6-logging-observability.md)
- [F-9: Transport Layer](.specify/specs/F-9-transport-layer-socket-configuration.md)
- [F-10: Network Interface Management](.specify/specs/F-10-network-interface-management.md)

---

**Document Version**: 1.0
**Last Updated**: 2025-11-05
**Status**: Research Complete, Ready for M3 Specification
