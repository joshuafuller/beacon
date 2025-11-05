# Beacon Compliance Dashboard

**Last Updated**: 2025-11-02
**Status**: Foundation Complete (M1 + M1-Refactoring + M1.1)
**Next Milestone**: M2 (mDNS Responder)

---

## Quick Status

| Milestone | Status | Completion Date | Key Achievement |
|-----------|--------|-----------------|-----------------|
| **M1** (Basic Querier) | ✅ Complete | 2025-10-XX | Query-only mDNS, 4 record types (A/PTR/SRV/TXT) |
| **M1-Refactoring** | ✅ Complete | 2025-11-01 | Clean architecture, 99% allocation reduction, transport abstraction |
| **M1.1** (Architectural Hardening) | ✅ Complete | 2025-11-02 | Socket configuration, interface management, security features |
| **Foundation Phase** | ✅ Complete | 2025-11-02 | Production-ready query-only mDNS library |

**Current RFC 6762 Compliance**: 52.8% (9.5/18 core sections implemented)
**Test Coverage**: 80.0%
**Functional Requirements Tracked**: 61 FRs across 3 milestones
**Next Up**: M2 - mDNS Responder (service registration, probing, announcing)

---

## What Works Today

Beacon is a **production-ready query-only mDNS library** with the following capabilities:

###  Core Functionality
- **mDNS Queries**: Send multicast queries to discover devices/services on .local domain
- **Record Types**: A (IPv4 addresses), PTR (service enumeration), SRV (service location), TXT (service metadata)
- **Response Parsing**: Full RFC 1035 DNS wire format support including name compression
- **Context-Aware**: All operations accept `context.Context` for cancellation and timeouts

### Production Features (M1.1)
- **System Daemon Coexistence**: Works alongside Avahi (Linux) and Bonjour (macOS) using SO_REUSEPORT
- **VPN Privacy Protection**: Automatically excludes VPN interfaces (utun*, tun*, ppp*) to prevent query leakage
- **Docker Interface Exclusion**: Filters out Docker virtual interfaces (docker0, veth*, br-*)
- **Multicast Storm Protection**: Rate limiting (100 qps threshold per source IP, 60s cooldown)
- **Link-Local Source Validation**: Drops non-link-local packets before parsing (RFC 6762 §11 compliance)
- **Packet Size Validation**: Enforces 9000 byte limit per RFC 6762 §17

### Architecture & Performance
- **Zero External Dependencies**: Standard library only (except `golang.org/x/sys`, `golang.org/x/net` for socket control)
- **Clean Architecture**: Transport abstraction enables IPv6 (M2), strict layer boundaries (F-2 compliant)
- **High Performance**: 99% allocation reduction via buffer pooling, <100ms query overhead
- **Robustness**: Zero crashes on malformed packets (fuzz tested with 114K iterations), zero race conditions

### Platform Support
- **Linux**: ✅ Fully validated (all features tested, Avahi coexistence confirmed)
- **macOS**: ⚠️ Code-complete, untested (socket options implemented, Bonjour coexistence expected to work)
- **Windows**: ⚠️ Code-complete, untested (SO_REUSEADDR implemented, no SO_REUSEPORT needed)

---

## Known Limitations

### Out of Scope (Planned for Future Milestones)
- **No mDNS Responder** (M2): Cannot register services or announce presence - query-only
- **No Service Browsing** (M2): Cannot continuously monitor for service changes - one-shot queries only
- **No IPv6** (M2): IPv4 multicast only (224.0.0.251), IPv6 (FF02::FB) in M2
- **No Response Caching** (M2): Every query hits the network - no TTL-based cache
- **No Probing/Announcing** (M2 Responder): Cannot claim names or advertise services
- **No Known-Answer Suppression** (M3): Cannot optimize repeat queries

### Platform Validation Status
- **macOS & Windows**: Socket configuration code complete with platform-specific build tags, but **untested**
  - Expected to work based on standard socket API behavior
  - Integration tests need macOS/Windows CI runners
  - Community testing welcome (see How to Contribute)

### Logging Infrastructure
- **No Structured Logging** (F-6 spec exists, not implemented): Debug logging uses basic Go `log` package
  - F-6 (Logging & Observability) spec defines structured logging, metrics, tracing
  - Deferred to maintain zero dependencies for M1.X
  - Planned for M2 or M3 based on user demand

---

## Navigation

**Compliance & Requirements**:
- [RFC Compliance Matrix](./RFC_COMPLIANCE_MATRIX.md) - Section-by-section RFC 6762/6763 implementation status
- [Functional Requirements Matrix](./FUNCTIONAL_REQUIREMENTS_MATRIX.md) - All 61 Foundation FRs with traceability
- Foundation completion narrative (publication pending)

**Project Governance**:
- [Beacon Constitution v1.1.0](../.specify/memory/constitution.md) - Project principles and non-negotiables
- [ROADMAP](../ROADMAP.md) - Milestone plan (M1-M6: Basic Querier → Production Ready)

**Architecture Specifications** (F-Series):
- [F-2: Package Structure](../.specify/specs/F-2-package-structure.md) - Layer boundaries and clean architecture
- [F-3: Error Handling](../.specify/specs/F-3-error-handling.md) - Error propagation (NetworkError, ValidationError, WireFormatError)
- [F-9: Transport Layer Configuration](../.specify/specs/F-9-transport-layer-socket-configuration.md) - Socket options, multicast configuration
- [F-10: Network Interface Management](../.specify/specs/F-10-network-interface-management.md) - Interface filtering, VPN exclusion
- [F-11: Security Architecture](../.specify/specs/F-11-security-architecture.md) - Rate limiting, source IP filtering

**Feature Specifications**:
- [M1: Basic mDNS Querier](../specs/002-mdns-querier/spec.md) - Original query-only spec
- [M1-Refactoring](../specs/003-m1-refactoring/spec.md) - Architectural improvements
- [M1.1: Architectural Hardening](../specs/004-m1-1-architectural-hardening/spec.md) - Production readiness

**Reference**:
- [RFC 6762: Multicast DNS](../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - PRIMARY TECHNICAL AUTHORITY
- [RFC 6763: DNS-SD](../RFC%20Docs/RFC-6763-DNS-SD.txt) - Service Discovery (M3+)
- [BEACON_FOUNDATIONS](../.specify/specs/BEACON_FOUNDATIONS.md) - DNS/mDNS/DNS-SD conceptual foundation

---

## How to Contribute

###  Testing & Validation
- **macOS Testing Needed**: Help validate socket configuration and Bonjour coexistence
  - Run: `go test ./... -v` on macOS with Bonjour active
  - Report results in [GitHub Issues](https://github.com/joshuafuller/beacon/issues)
- **Windows Testing Needed**: Validate SO_REUSEADDR behavior on Windows
  - Test: Create querier, verify no port conflicts
  - Report: Socket errors, binding issues, query behavior

### Open Issues & Roadmap
- [GitHub Issues](https://github.com/joshuafuller/beacon/issues) - Bug reports, feature requests
- [M2 Planning](../specs/) - mDNS Responder specification (coming soon)
- [Spec Kit Workflow](https://docs.claude.com/en/docs/claude-code/speckit) - How features are specified and planned

### Development Process
Beacon follows [Specification-Driven Development](../.specify/memory/constitution.md):
1. **Spec First**: Features start with `/speckit.specify` (user stories, FRs, success criteria)
2. **Plan**: `/speckit.plan` generates implementation plan (architecture, phases, tasks)
3. **TDD**: Write tests first (RED → GREEN → REFACTOR)
4. **Validate**: All FRs must be testable, all SCs must be measurable

**Want to contribute?** Start by reading:
- [Beacon Constitution](../.specify/memory/constitution.md) - Project principles
- [CLAUDE.md](../CLAUDE.md) - Development guidelines for AI-assisted development
- [ROADMAP](../ROADMAP.md) - Understand where the project is heading

---

**Dashboard Version**: 1.0
**Foundation Phase**: ✅ Complete (M1 + M1-R + M1.1)
**Ready for**: M2 (mDNS Responder)
