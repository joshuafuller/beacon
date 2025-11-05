# Beacon Next Phase - Executive Summary

**Date**: 2025-11-05
**Prepared By**: Research and Planning Analysis
**Status**: Ready for Implementation

---

## Quick Summary

After comprehensive analysis of RFC 6762/6763 requirements and current implementation status, we recommend:

### **Primary Recommendation: M3 - Service Discovery (DNS-SD Client)**

**Priority**: P0 (Highest)
**Duration**: 3-4 weeks
**RFC Coverage**: RFC 6763 §4-5 (Service Browsing + Resolution)

---

## What's Complete Today

| Milestone | Status | Key Features |
|-----------|--------|--------------|
| M1 | ✅ Complete | mDNS Querier (query-only) |
| M1-Refactoring | ✅ Complete | Clean architecture, transport abstraction |
| M1.1 | ✅ Complete | Socket config, security, interface management |
| M2 | ✅ 94.6% | Service registration, probing, announcing, conflict resolution |

**Current RFC Compliance**:
- RFC 6762 (mDNS): 72.2% (13/18 sections)
- RFC 6763 (DNS-SD): 65% (server-side only)

---

## Critical Gap: Client-Side Service Discovery

### The Problem

Beacon can **register** services (M2 responder) but **cannot discover** them:
- ❌ Cannot browse available services on the network
- ❌ Cannot resolve service instances to connection info (host:port)
- ❌ Cannot monitor for service changes (add/remove)

This is like building a web server without a web browser - incomplete workflow.

### The Solution: M3 - Service Discovery

Implement **Browser** and **Resolver** for client-side DNS-SD:

**Browser**:
- Discover all instances of a service type (e.g., find all printers)
- Continuous monitoring for service changes
- Event-driven API (ServiceAdded, ServiceRemoved, ServiceUpdated)

**Resolver**:
- Resolve service instance to connection info (host, port, IP, metadata)
- Parallel SRV/TXT/A queries for efficiency (<100ms)
- Cache results with TTL management

### User Impact

**Before M3** (Current):
```go
// ❌ Cannot discover services
// Must manually know service names
```

**After M3**:
```go
// ✅ Discover all HTTP services
browser := browser.Browse("_http._tcp.local")
for event := range browser.Events() {
    fmt.Printf("Found: %s at %s:%d\n",
        event.Instance, event.Host, event.Port)
}

// ✅ Resolve specific service
info := resolver.Resolve("My Printer._ipp._tcp.local")
conn := net.Dial("tcp", fmt.Sprintf("%s:%d", info.IP, info.Port))
```

---

## Secondary Priority: M4 - IPv6 Support

**Priority**: P1 (High, after M3)
**Duration**: 3-4 weeks
**RFC Coverage**: RFC 6762 §20 (IPv6 Considerations)

### Why IPv6?

1. **Modern Networking**: IPv6 adoption growing (enterprise, IoT, mobile)
2. **RFC Compliance**: Required for full RFC 6762 compliance
3. **Future-Proof**: Prepare for IPv6-only networks

### What's Involved

- **FF02::FB Multicast**: IPv6 multicast address for mDNS
- **AAAA Records**: IPv6 address records (like A for IPv4)
- **Dual-Stack**: Participate in both IPv4 and IPv6 .local zones
- **UDPv6Transport**: Extend transport layer for IPv6

### Can Parallelize

M3 and M4 are mostly independent - can develop concurrently if resources available.

---

## Recommended Roadmap

### Immediate (Next 1-2 Months)

```
M2 Completion (1 week)
  └─► M2.1: Quick Wins (1 week)
       ├─► Goodbye packets (TTL=0 wire format)
       └─► QU bit unicast response

  └─► M3: Service Discovery (3-4 weeks)
       ├─► Browser: PTR queries, continuous monitoring
       ├─► Resolver: SRV/TXT/A resolution
       └─► Contract tests: RFC 6763 §4-5
```

### Medium-Term (3-5 Months)

```
M4: IPv6 Support (3-4 weeks)
  ├─► UDPv6Transport: FF02::FB
  ├─► AAAA records
  └─► Dual-stack operation

M5: Production Hardening (2-3 weeks)
  ├─► Structured logging (log/slog)
  ├─► Observability (metrics, traces)
  └─► Platform testing (macOS, Windows)
```

### Long-Term (6+ Months)

```
M6: Advanced Features (as needed)
  ├─► Service subtypes (RFC 6763 §7.1)
  ├─► Domain enumeration (RFC 6763 §11)
  └─► Community requests
```

---

## Key Functional Requirements

### M3: Service Discovery (18 Requirements)

**Browsing (RFC 6763 §4)**:
- FR-M3-001: PTR queries for service type enumeration
- FR-M3-002: Continuous browsing (source port 5353)
- FR-M3-003: Handle multiple PTR responses
- FR-M3-004: Detect service removal (goodbye packets)
- FR-M3-005: Service enumeration metadata (_services._dns-sd._udp)

**Resolution (RFC 6763 §5)**:
- FR-M3-006: SRV record queries (host:port)
- FR-M3-007: TXT record queries (metadata)
- FR-M3-008: A record queries (IPv4 address)
- FR-M3-009: Parallel queries for efficiency
- FR-M3-010: Additional record processing (optimization)
- FR-M3-011: Handle missing/incomplete data gracefully
- FR-M3-012: Cache with TTL management

**Architecture**:
- FR-M3-013: Browser type (public API)
- FR-M3-014: Resolver type (public API)
- FR-M3-015: ServiceInfo type (shared)
- FR-M3-016: Context-aware operations
- FR-M3-017: Event-driven API
- FR-M3-018: Reuse M1 querier

### M4: IPv6 Support (11 Requirements)

**Core**:
- FR-M4-001: FF02::FB multicast support
- FR-M4-002: AAAA record support
- FR-M4-003: Register services on both IPv4 and IPv6
- FR-M4-004: Query on both IPv4 and IPv6

**Modes**:
- FR-M4-005: IPv6-only mode
- FR-M4-006: IPv4-only mode (default)
- FR-M4-007: Dual-stack mode (auto-detect)

**Architecture**:
- FR-M4-008: Abstract transport layer (already done in M1-R)
- FR-M4-009: Extend record builders for AAAA
- FR-M4-010: Auto-detect IP version support
- FR-M4-011: Per-interface IPv6 config

---

## Success Criteria

### M3 Success

**Functional**:
- ✅ Discover 100% of Avahi-registered services
- ✅ Resolution completes in <100ms
- ✅ Live monitoring detects changes within 2 seconds
- ✅ 10+ RFC 6763 contract tests PASS

**Quality**:
- ✅ Test coverage ≥80%
- ✅ Zero data races
- ✅ RFC 6763 compliance: 80%+ (up from 65%)

### M4 Success

**Functional**:
- ✅ Services on FF02::FB discoverable by IPv6-only clients
- ✅ Dual-stack: same service in both zones
- ✅ AAAA records work correctly
- ✅ 5+ RFC 6762 §20 contract tests PASS

**Quality**:
- ✅ Test coverage ≥80%
- ✅ Zero data races
- ✅ RFC 6762 compliance: 80%+ (up from 72.2%)

### Combined Success

- ✅ RFC 6762 compliance: 85-90%
- ✅ RFC 6763 compliance: 80-85%
- ✅ Production-ready complete service discovery

---

## Implementation Approach

All milestones follow [Beacon Constitution v1.1.0]:

1. **Spec-Driven**: Use `/speckit.specify` to create detailed specs
2. **TDD**: Write tests first (RED → GREEN → REFACTOR)
3. **RFC Compliant**: Validate against RFC 6762/6763
4. **Phased**: Deliver incrementally with working code
5. **Minimal Dependencies**: Use stdlib (log/slog), golang.org/x/* only when justified

### Workflow for M3

```
1. /speckit.specify → specs/007-service-discovery/spec.md
2. /speckit.plan    → specs/007-service-discovery/plan.md
3. /speckit.tasks   → specs/007-service-discovery/tasks.md
4. Implementation   → TDD (RED → GREEN → REFACTOR)
5. Validation       → Contract tests, interoperability
6. Documentation    → API docs, examples, guides
```

---

## Open Questions

### Technical Decisions

1. **Browser API**: Channel-based events or callbacks?
   - **Recommendation**: Channel-based (`for event := range browser.Events()`)
   - **Rationale**: More idiomatic Go

2. **Resolver Caching**: Independent cache or share with Browser?
   - **Recommendation**: Independent for MVP, shared for optimization later
   - **Rationale**: Simpler architecture

3. **IPv6 Default**: Auto-enable dual-stack or require opt-in?
   - **Recommendation**: Auto-detect and enable if available
   - **Rationale**: Better UX, matches OS behavior

4. **Logging Library**: log/slog or third-party?
   - **Recommendation**: log/slog (requires Go 1.21+)
   - **Rationale**: Constitution Principle V (minimal dependencies)

### Research Needed

1. **IPv6 Platform Support**: Test on macOS/Windows
2. **Avahi/Bonjour Dual-Stack**: How do they handle it?
3. **Cache Strategy**: Optimal TTL refresh (80%? 95%?)
4. **Performance Impact**: Dual-stack overhead measurement

---

## Quick Wins Before M3

### M2.1: Polish (1 Week)

These can be done quickly before starting M3:

1. **Goodbye Packets** (1-2 days)
   - Logic exists, just send TTL=0 on wire
   - High value: graceful service removal

2. **QU Bit Unicast Response** (1-2 days)
   - Responder detects QU bit, just send unicast
   - Medium value: network traffic optimization

3. **M2 Documentation** (2-3 days)
   - Update CLAUDE.md
   - Update examples
   - Validate quickstart.md

---

## Timeline Summary

| Milestone | Duration | Start | Key Deliverables |
|-----------|----------|-------|------------------|
| M2 Completion | 1 week | Now | Documentation, examples |
| M2.1 Quick Wins | 1 week | +1 week | Goodbye, QU bit |
| **M3 Service Discovery** | **3-4 weeks** | **+2 weeks** | **Browser, Resolver, RFC 6763 §4-5** |
| M4 IPv6 Support | 3-4 weeks | +6 weeks | UDPv6, AAAA, dual-stack |
| M5 Production | 2-3 weeks | +10 weeks | Logging, metrics, platform tests |

**Total to Production-Ready**: ~3 months (12 weeks)

---

## Resources

### Full Plan
- [NEXT_PHASE_PLAN.md](./NEXT_PHASE_PLAN.md) - Detailed analysis with all FRs, RFCs, and implementation notes

### Current Status
- [M2 Completion Report](./specs/006-mdns-responder/COMPLETION_REPORT.md) - 94.6% complete, 122/129 tasks
- [RFC Compliance Matrix](./docs/internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md) - Current 72.2% RFC 6762, 65% RFC 6763

### RFCs
- [RFC 6762: Multicast DNS](./RFC%20Docs/RFC-6762-Multicast-DNS.txt)
- [RFC 6763: DNS-Based Service Discovery](./RFC%20Docs/RFC-6763-DNS-SD.txt)

### Project Governance
- [Beacon Constitution v1.1.0](./.specify/memory/constitution.md) - Principles and non-negotiables
- [BEACON_FOUNDATIONS v1.1](./.specify/specs/BEACON_FOUNDATIONS.md) - DNS/mDNS/DNS-SD concepts

---

## Next Action

**Immediate**: Create M3 specification

```bash
# Use Spec Kit to create detailed M3 spec
/speckit.specify

# Input: "Implement DNS-SD Service Discovery (client-side) with
# Browser for continuous service monitoring and Resolver for
# service instance resolution per RFC 6763 §4-5"
```

This will generate `specs/007-service-discovery/spec.md` with:
- User stories (US-M3-1, US-M3-2, US-M3-3)
- Functional requirements (FR-M3-001 through FR-M3-018)
- Success criteria (SC-M3-001 through SC-M3-005)

Then proceed with `/speckit.plan` and `/speckit.tasks` for implementation.

---

**Status**: ✅ Research Complete, Ready for Implementation
**Recommended Start Date**: After M2 documentation complete (~1 week)
**Expected Completion (M3)**: 4-5 weeks from start
