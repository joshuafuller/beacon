# Feature Specification: Basic mDNS Querier (M1)

**Feature Branch**: `002-mdns-querier`
**Created**: 2025-11-01
**Status**: Draft
**Input**: User description: "Basic mDNS querier"

## Overview

This specification defines Milestone 1 (M1) of the Beacon project: a **Basic mDNS Querier** that implements core mDNS query functionality per RFC 6762. M1 provides one-shot query capabilities for discovering hosts and services on the local network using multicast DNS.

**Scope**: M1 focuses on the fundamental query/response cycle without caching, continuous browsing, or responder functionality. These advanced features are deferred to subsequent milestones (M2+).

## User Scenarios & Testing

### User Story 1 - Discover Host by Name (Priority: P1)

As a **network application developer**, I need to resolve a local hostname (e.g., "myprinter.local") to an IP address using mDNS, so that my application can connect to devices on the local network without requiring manual IP configuration.

**Why this priority**: This is the fundamental capability of mDNS - resolving .local domain names to IP addresses. Without this, no other mDNS functionality is useful. This delivers immediate value by enabling basic host discovery.

**Independent Test**: Can be fully tested by querying a known hostname (e.g., "test-device.local") and verifying the returned IP address matches the expected value. Requires only a test mDNS responder on the network.

**Acceptance Scenarios**:

1. **Given** a local device "printer.local" exists on the network, **When** I query for "printer.local" A record, **Then** I receive the device's IPv4 address within 1 second
2. **Given** a hostname that doesn't exist, **When** I query for "nonexistent.local", **Then** I receive no results within the timeout period (1 second default)
3. **Given** multiple devices respond with the same hostname, **When** I query for that hostname, **Then** I receive all IP addresses from all responding devices
4. **Given** a query timeout of 500ms is configured, **When** I query for a slow-responding device, **Then** the query completes (with or without results) within 500ms

---

### User Story 2 - Query Service Records (Priority: P2)

As a **network application developer**, I need to query for service records (PTR, SRV, TXT) on the local network, so that I can discover available services (e.g., "_http._tcp.local") and their connection details (port, host, metadata).

**Why this priority**: This extends basic host discovery (P1) to enable service discovery, which is the primary use case for DNS-SD. Depends on P1's query infrastructure but adds service-specific record handling.

**Independent Test**: Can be tested independently by querying for a known service (e.g., "_test._tcp.local") and verifying the returned PTR, SRV, and TXT records match expected values. Requires a test service responder.

**Acceptance Scenarios**:

1. **Given** a service "_http._tcp.local" is advertised on the network, **When** I query for PTR records for "_http._tcp.local", **Then** I receive a list of service instances
2. **Given** a service instance "MyWebServer._http._tcp.local", **When** I query for its SRV record, **Then** I receive the target hostname and port number
3. **Given** a service instance with TXT metadata, **When** I query for its TXT record, **Then** I receive key-value pairs (e.g., "version=1.0", "path=/api")
4. **Given** I query for a service type that doesn't exist, **When** the query timeout expires, **Then** I receive an empty result set with no errors

---

### User Story 3 - Handle Network and Protocol Errors (Priority: P3)

As a **network application developer**, I need clear error reporting when mDNS queries fail due to network issues or malformed responses, so that I can diagnose problems and provide meaningful feedback to users.

**Why this priority**: Error handling is critical for production use but can be developed after core functionality (P1, P2) is working. This ensures robustness without blocking initial development.

**Independent Test**: Can be tested independently by simulating network failures (disconnected interface, unreachable multicast group) and malformed packets (truncated messages, invalid DNS names). Does not require working query functionality.

**Acceptance Scenarios**:

1. **Given** the network interface is down, **When** I attempt to send an mDNS query, **Then** I receive a NetworkError with details about the interface state
2. **Given** a response with a malformed DNS name (invalid characters), **When** I parse the response, **Then** I receive a ValidationError and the response is discarded
3. **Given** a response with truncated message data, **When** I parse the response, **Then** I receive a WireFormatError and the response is discarded
4. **Given** insufficient permissions to bind to port 5353, **When** I create a querier, **Then** I receive a NetworkError indicating permission issues

---

### Edge Cases

**Query Edge Cases**:
- What happens when querying for an empty name ("")? → ValidationError: "name cannot be empty"
- What happens when querying for a name longer than 255 bytes? → ValidationError: "name exceeds maximum length (255 bytes)"
- What happens when querying for a name with invalid characters (e.g., spaces, underscores in hostnames)? → ValidationError: "invalid characters in hostname"
- What happens when querying for a record type not supported in M1 (e.g., AAAA for IPv6)? → ValidationError: "unsupported record type: AAAA"

**Network Edge Cases**:
- What happens when the default network interface has no multicast route? → NetworkError: "multicast not supported on interface"
- What happens when the system has no network interfaces? → NetworkError: "no network interfaces available"
- What happens when a firewall blocks outbound multicast traffic? → Query succeeds but times out with no responses (expected behavior)
- What happens when a firewall blocks inbound multicast traffic? → Query succeeds but times out with no responses (expected behavior)

**Protocol Edge Cases**:
- What happens when a response contains compressed DNS names (RFC 1035 §4.1.4)? → Parse correctly using name decompression algorithm
- What happens when a response contains the TC (truncation) bit set? → M1: Log warning, return partial results. M2: Implement TCP fallback (out of scope)
- What happens when a response contains multiple Answer sections for the same query? → Aggregate all answers into result set
- What happens when a response arrives after the query timeout? → Discard response (query already completed)
- What happens when a response contains records in Additional section but no Answer section? → M1: Process Answer section only, ignore Additional (out of scope for cache-less querier)

**Resource Edge Cases**:
- What happens when 100 concurrent queries are executed? → All queries succeed independently with no resource leaks (per F-7 requirements)
- What happens when a query is cancelled via context.Context? → Query terminates immediately, resources cleaned up, returns context.Canceled error
- What happens when the querier is closed while queries are in flight? → All active queries are cancelled, sockets closed, goroutines terminated within graceful shutdown timeout (per F-7)

## Requirements

### Functional Requirements

**Query Construction**:
- **FR-001**: System MUST construct valid mDNS query messages per RFC 6762 with QU (unicast response) bit clear for standard multicast queries
- **FR-002**: System MUST support querying for A, PTR, SRV, and TXT record types
- **FR-003**: System MUST validate queried names follow DNS naming rules (labels ≤63 bytes, total name ≤255 bytes, valid characters)
- **FR-004**: System MUST use mDNS port 5353 and multicast address 224.0.0.251 for IPv4 queries

**Query Execution**:
- **FR-005**: System MUST send mDNS queries to the multicast group on the default network interface
- **FR-006**: System MUST listen for mDNS responses on port 5353 for the duration of the query timeout
- **FR-007**: System MUST accept configurable query timeout (default: 1 second, range: 100ms to 10 seconds)
- **FR-008**: System MUST support context-based cancellation of queries per F-4 concurrency patterns

**Response Handling**:
- **FR-009**: System MUST parse mDNS response messages per RFC 6762 wire format specification
- **FR-010**: System MUST extract Answer, Authority, and Additional sections from responses (M1: Answer section only; Additional section deferred to M2 for caching support)
- **FR-011**: System MUST validate response message format and discard malformed packets
- **FR-012**: System MUST decompress DNS names per RFC 1035 §4.1.4 (message compression)

**Error Handling**:
- **FR-013**: System MUST return NetworkError for socket creation, binding, or I/O failures per F-3 error types
- **FR-014**: System MUST return ValidationError for invalid query names or unsupported record types per F-3 error types
- **FR-015**: System MUST return WireFormatError for malformed response packets per F-3 error types
- **FR-016**: System MUST log malformed packets at DEBUG level without failing the query (continue collecting valid responses)

**Resource Management**:
- **FR-017**: System MUST clean up all sockets and goroutines when query completes per F-7 resource management patterns
- **FR-018**: System MUST support graceful shutdown that cancels active queries and releases resources within configured timeout per F-7
- **FR-019**: System MUST pass `go test -race` with zero race conditions per F-8 testing requirements

**RFC 6762 Compliance**:
- **FR-020**: System MUST set DNS header fields per RFC 6762 §18 (QR=0 per §18.2, OPCODE=0 per §18.3, AA=0 per §18.4, TC=0 per §18.5, RD=0 per §18.6)
- **FR-021**: System MUST validate received responses have QR=1 (response bit set) per RFC 6762 §18.2
- **FR-022**: System MUST ignore responses with RCODE != 0 (error responses) per RFC 6762 §18.11

### Non-Functional Requirements

**Performance**:
- **NFR-001**: Query processing overhead MUST be under 100ms (time from response receipt to parsed records available)
- **NFR-002**: System MUST handle 100 concurrent queries without memory leaks or goroutine leaks per F-7

**Reliability**:
- **NFR-003**: System MUST handle malformed packets without crashes or panics (verified via fuzz testing with 10,000 random packets)
- **NFR-004**: System MUST achieve ≥80% test coverage per F-8 requirements

**Usability**:
- **NFR-005**: API MUST provide idiomatic Go interfaces with context.Context support per F-4 patterns
- **NFR-006**: Error messages MUST include actionable context (e.g., "failed to bind to 224.0.0.251:5353: permission denied (requires root or CAP_NET_RAW)")

### Key Entities

- **Querier**: The core component that sends mDNS queries and collects responses. Encapsulates socket management, query construction, and response parsing.
- **Query**: Represents a single mDNS query operation with a name, record type (A/PTR/SRV/TXT), and timeout configuration.
- **Response**: Contains parsed DNS resource records (Answer section) returned from one or more mDNS responders on the network.
- **ResourceRecord**: A single DNS resource record with name, type, class, TTL, and type-specific data (e.g., IP address for A records, hostname/port for SRV records).

## Success Criteria

**Functionality**:
- **SC-001**: Developers can resolve .local hostnames to IP addresses with a single function call
- **SC-002**: System successfully discovers 95% of responding devices on the local network within 1 second
- **SC-003**: All RFC 6762 MUST requirements for query format and response handling are validated with test coverage

**Performance**:
- **SC-004**: Query processing overhead is under 100ms (time from response receipt to parsed records available)
- **SC-005**: System handles 100 concurrent queries without memory leaks or goroutine leaks

**Reliability**:
- **SC-006**: Zero crashes or panics when processing malformed packets (verified via fuzz testing with 10,000 random packets)
- **SC-007**: 100% of tests pass with race detector enabled (`go test -race`)

**Usability**:
- **SC-008**: Developers can configure query timeout in a single line of code using functional options pattern (per F-5)
- **SC-009**: Error messages clearly distinguish between network errors, validation errors, and protocol errors (per F-3 error types)

**Code Quality**:
- **SC-010**: Test coverage is ≥80% for all querier code per F-8 requirements
- **SC-011**: All public API functions have godoc comments with usage examples

## Scope & Constraints

### In Scope (M1)
✅ One-shot mDNS queries (send query, collect responses, return results)
✅ A, PTR, SRV, TXT record type queries
✅ Query timeout configuration
✅ Basic response parsing and record extraction
✅ Error handling for network and protocol errors
✅ IPv4 multicast (224.0.0.251:5353)
✅ Default network interface only

### Out of Scope (Future Milestones)
❌ Continuous service browsing (M2: Service Browser)
❌ Response caching and TTL management (M2: Cache Manager)
❌ Probing and announcing (Responder - separate feature)
❌ Known-Answer suppression (M2: Cache integration required)
❌ Multi-interface queries (M4: Multi-homing)
❌ IPv6 support (M4: Dual-stack)
❌ QU bit (unicast response) queries (M3: Advanced Queries)
❌ Additional section processing for cache pre-population (M2: Cache Manager)

### Constraints

**Platform**:
- M1 targets **Linux only** (per Constitution Principle IV: Phased Approach)
- Requires **Go 1.21 or later**
- Requires **root privileges or CAP_NET_RAW capability** for binding to port 5353

**RFC Compliance**:
- **Constitution Principle I**: RFC 6762 compliance is mandatory and overrides all other concerns
- Any deviation from RFC MUST requirements requires explicit documentation and justification

**Architecture**:
- MUST follow F-2 package structure: `beacon/querier/` (public API), `internal/message/`, `internal/protocol/` (implementation)
- MUST use F-3 error types: NetworkError, ValidationError, WireFormatError
- MUST follow F-4 concurrency patterns: context propagation, goroutine lifecycle
- MUST meet F-7 resource management requirements: no leaks, graceful shutdown
- MUST follow F-8 testing requirements: TDD cycle, ≥80% coverage, race detection

## Dependencies

### Architecture Specifications

- [F-2: Package Structure](../../.specify/specs/F-2-package-structure.md) - Defines `beacon/querier/` public API package and `internal/message/`, `internal/protocol/` implementation packages
- [F-3: Error Handling](../../.specify/specs/F-3-error-handling.md) - Defines NetworkError, ValidationError, WireFormatError types
- [F-4: Concurrency Model](../../.specify/specs/F-4-concurrency-model.md) - Defines context propagation, goroutine lifecycle, timeout patterns
- [F-5: Configuration & Defaults](../../.specify/specs/F-5-configuration.md) - Defines functional options pattern for query timeout configuration
- [F-7: Resource Management](../../.specify/specs/F-7-resource-management.md) - Defines cleanup patterns, graceful shutdown, no-leak requirements
- [F-8: Testing Strategy](../../.specify/specs/F-8-testing-strategy.md) - Defines TDD cycle, coverage requirements (≥80%), race detection

### Project Governance

- [Beacon Constitution v1.0.0](../../.specify/memory/constitution.md) - Governs all development (Principles I-VII)
- [BEACON_FOUNDATIONS v1.1](../../.specify/specs/BEACON_FOUNDATIONS.md) - Common knowledge base, terminology glossary (§5), architecture overview (§4)

### Technical Authority

**PRIMARY TECHNICAL AUTHORITY** per Constitution Principle I:

- [RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - PRIMARY AUTHORITY for mDNS protocol
  - §5.4: Questions (QM bit, QU bit)
  - §6: Responding
  - §18: Security Considerations (message validation, header fields)
  - §18.1: Query Header Fields
  - §18.3: Response Header Fields

- [RFC 1035: Domain Names - Implementation and Specification](https://tools.ietf.org/html/rfc1035) - DNS message format and name compression
  - §4.1.4: Message compression (for DNS name decompression in responses)

**Critical Note**: Per Constitution Principle I, RFC requirements override all other concerns. Any conflict between this specification and RFC 6762/1035 MUST be resolved in favor of the RFC.

## References

### Go Resources

- [Go net package](https://pkg.go.dev/net) - Network I/O, UDP sockets
- [Go context package](https://pkg.go.dev/context) - Context-based cancellation and timeouts
- [Go testing package](https://pkg.go.dev/testing) - Testing framework for TDD per F-8

---

**Next Phase**: After this specification is approved, proceed to `/speckit.plan` to generate the implementation plan based on F-series architecture patterns.
