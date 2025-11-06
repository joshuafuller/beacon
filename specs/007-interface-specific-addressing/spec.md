# Feature Specification: Interface-Specific IP Address Advertising

**Feature Branch**: `007-interface-specific-addressing`
**Created**: 2025-11-06
**Status**: Draft
**Input**: Fix multi-interface hosts to advertise correct IP address per RFC 6762 Section 15

## Context

**Problem**: Multi-interface hosts (WiFi + Ethernet, multi-NIC servers, Docker/VPN) currently advertise the wrong IP address in mDNS responses, causing connectivity failures and violating RFC 6762 §15.

**Current Behavior**: `getLocalIPv4()` in [responder/responder.go:288-308](../../responder/responder.go) returns the FIRST non-loopback IPv4 address found, regardless of which interface received the query. This single IP is used for all responses.

**Example Failure**:
```
Machine: eth0 (10.0.0.5), eth1 (192.168.1.100)
Query arrives on eth0 → Response advertises 192.168.1.100
Client on 10.0.0.0/24 cannot reach service (wrong subnet)
```

**RFC 6762 §15 Requirement** (lines 1020-1024):
> When a Multicast DNS responder sends a Multicast DNS response message containing its own address records, it MUST include all addresses that are valid on the interface on which it is sending the message, and **MUST NOT include addresses that are not valid on that interface**.

**Related Work**:
- M4 (IPv6 Support) already plans per-interface transport binding (FR-M4-003)
- See [docs/planning/NEXT_PHASE_PLAN.md](../../docs/planning/NEXT_PHASE_PLAN.md)
- This fast-track fix implements the minimal changes needed for correctness

## User Scenarios & Testing

### User Story 1 - Laptop with WiFi and Ethernet (Priority: P1)

Developer's laptop has WiFi (192.168.1.100) and Ethernet (10.0.0.5) connected to different networks. They register an mDNS service for their development server. Clients on the 10.0.0.0/24 network query for the service and need to receive the correct 10.0.0.5 address to connect.

**Why this priority**: Most common scenario. Affects all multi-interface development environments (laptops with WiFi + Ethernet, VPN + LAN). Critical for service reachability.

**Independent Test**: Can be fully tested by registering a service on a multi-interface machine, sending queries from each network, and verifying responses contain only the interface-specific IP. Delivers immediate connectivity fix.

**Acceptance Scenarios**:

1. **Given** laptop with WiFi (192.168.1.100) and Ethernet (10.0.0.5), service registered, **When** query arrives on Ethernet interface, **Then** response MUST advertise 10.0.0.5 and MUST NOT advertise 192.168.1.100
2. **Given** laptop with WiFi (192.168.1.100) and Ethernet (10.0.0.5), service registered, **When** query arrives on WiFi interface, **Then** response MUST advertise 192.168.1.100 and MUST NOT advertise 10.0.0.5
3. **Given** single-interface machine (10.0.0.5 only), service registered, **When** query arrives, **Then** response advertises 10.0.0.5 (regression test - single-interface still works)

---

### User Story 2 - Server with Multiple NICs (Priority: P2)

Production server has multiple network interfaces (10.0.1.10, 10.0.2.10, 10.0.3.10) on separate VLANs for different purposes (management, application, storage). Each network has isolated routing. Services registered via mDNS must advertise the correct IP for clients on each VLAN.

**Why this priority**: Enterprise/production scenario. Affects multi-NIC servers, infrastructure services. Important for security (VLAN isolation) and routing correctness.

**Independent Test**: Can be tested by registering service on multi-NIC server, querying from each VLAN, and verifying responses contain only the VLAN-specific IP. Delivers correct routing behavior.

**Acceptance Scenarios**:

1. **Given** server with 3 NICs (10.0.1.10, 10.0.2.10, 10.0.3.10), service registered, **When** query arrives on VLAN1 interface, **Then** response advertises only 10.0.1.10
2. **Given** server with 3 NICs (10.0.1.10, 10.0.2.10, 10.0.3.10), service registered, **When** query arrives on VLAN2 interface, **Then** response advertises only 10.0.2.10
3. **Given** server with isolated VLANs (no inter-VLAN routing), **When** query on VLAN1 receives VLAN2 IP, **Then** connection fails (current buggy behavior - test validates fix)

---

### User Story 3 - Docker and VPN Interfaces (Priority: P3)

Developer machine has physical interface (192.168.1.100), Docker bridge (172.17.0.1), and VPN tunnel (10.8.0.2). Only the physical interface should be used for mDNS service advertising. Docker/VPN interfaces should be excluded from mDNS responses.

**Why this priority**: Specialized scenario. Affects containerized environments, remote work with VPNs. Important but less common than P1/P2.

**Independent Test**: Can be tested by registering service on machine with Docker/VPN, querying from physical network, and verifying response contains only physical interface IP (excludes Docker/VPN). Delivers correct interface selection.

**Acceptance Scenarios**:

1. **Given** machine with physical interface (192.168.1.100), Docker (172.17.0.1), VPN (10.8.0.2), **When** query arrives on physical interface, **Then** response advertises only 192.168.1.100
2. **Given** machine with Docker/VPN interfaces, **When** service registered, **Then** system excludes Docker bridges and VPN tunnels from mDNS advertising (leverages M1.1 interface selection logic from F-10)
3. **Given** Docker container querying host service, **When** query sent to host, **Then** response includes Docker bridge IP (special case - containers need to reach host via bridge)

---

### Edge Cases

- **What happens when interface goes down during service lifetime?** System detects interface down event, stops advertising on that interface, continues on remaining interfaces. (Deferred to M4 - requires interface monitoring)
- **What happens when new interface added while service registered?** New interface not automatically used until service re-registered or system restarted. (Deferred to M4)
- **What happens when query arrives on multiple interfaces simultaneously?** Each interface receives independent response with its own IP address. No coordination needed.
- **What happens when machine has only loopback interface?** Registration fails with error (no valid interfaces). Current behavior preserved.
- **What happens when interface has multiple IPs (IPv4 + IPv6)?** This fix handles IPv4 only. IPv6 deferred to M4.
- **What happens when interface IP changes (DHCP renewal)?** Service continues advertising old IP until re-registered. (Deferred to M4 - requires IP change monitoring)

## Requirements

### Functional Requirements

- **FR-001**: System MUST determine which network interface received each mDNS query
- **FR-002**: System MUST look up the IPv4 address assigned to the receiving interface
- **FR-003**: System MUST include ONLY the interface-specific IPv4 address in A records of the response
- **FR-004**: System MUST NOT include IP addresses from other interfaces in responses (RFC 6762 §15 MUST NOT requirement)
- **FR-005**: System MUST preserve existing single-interface behavior (no regression for simple cases)
- **FR-006**: System MUST handle registration on machines with multiple interfaces without requiring user to specify interface
- **FR-007**: Response builder MUST accept interface context (interface ID or index) when constructing responses
- **FR-008**: System MUST handle errors when interface lookup fails (interface down, invalid index) by returning error or skipping response
- **FR-009**: System MUST work with existing F-10 interface selection logic (exclude Docker, VPN, etc.)
- **FR-010**: Contract tests MUST verify RFC 6762 §15 compliance for multi-interface scenarios

### Non-Functional Requirements

- **NFR-001**: Changes MUST NOT break API compatibility (internal changes only)
- **NFR-002**: Performance impact MUST be acceptable for RFC compliance fix (<10% overhead on response path)
- **NFR-003**: Solution MUST work with current transport architecture (single socket on 0.0.0.0:5353)
- **NFR-004**: Solution MUST NOT block M4 per-interface transport work (compatible architecture)

### Key Entities

- **QueryContext**: Represents incoming query with interface metadata (interface ID/index, source IP, query bytes). Passed through query handling pipeline to response builder.
- **InterfaceAddress**: Mapping of network interface to its assigned IPv4 address. Cached or looked up per-query.
- **ResponseBuilder**: Constructs mDNS response messages. Enhanced to accept interface context and build interface-specific A records.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Multi-interface machines respond to queries with interface-specific IP addresses (100% of test cases pass)
- **SC-002**: RFC 6762 §15 contract test passes (verifies MUST include interface IP, MUST NOT include other IPs)
- **SC-003**: Single-interface machines continue to work without regression (existing integration tests pass)
- **SC-004**: Response construction overhead <10% compared to baseline (benchmark `BenchmarkHandleQuery_WithInterfaceContext`)
- **SC-005**: All existing responder tests pass without modification (API compatibility maintained)
- **SC-006**: Documentation updated (godoc, RFC_COMPLIANCE_GUIDE.md) to reflect interface-specific behavior

## RFC Compliance

**RFC 6762 §15 "Address Records"** (lines 1020-1024):

This feature directly addresses the MUST NOT requirement:

> When a Multicast DNS responder sends a Multicast DNS response message
> containing its own address records, it MUST include all addresses
> that are valid on the interface on which it is sending the message,
> and **MUST NOT include addresses that are not valid on that interface**
> (such as addresses that may be configured on the host's other
> interfaces).

**Compliance Strategy**:
1. Determine receiving interface for each query (FR-001)
2. Look up IP address for that interface (FR-002)
3. Build A record with only that IP (FR-003)
4. Never include IPs from other interfaces (FR-004)

**Contract Test**: `TestRFC6762_Section15_InterfaceSpecificAddresses`

## Scope

### In Scope (Fast-Track Fix)

1. **Interface Context Propagation**:
   - Add `interfaceIndex` parameter to query handler
   - Thread interface context through to response builder
   - Look up interface IP at response time

2. **Response Builder Enhancement**:
   - Accept interface ID/index in `BuildRecordSet()` or similar
   - Replace global `getLocalIPv4()` with per-interface lookup
   - Build A records with interface-specific IP

3. **Transport Layer Minimal Changes**:
   - Determine interface index from received packet
   - Pass interface context to responder query handler
   - No per-interface socket binding (deferred to M4)

4. **Testing**:
   - RFC 6762 §15 contract test
   - Multi-interface unit tests
   - Single-interface regression tests

5. **Documentation**:
   - Update godoc in `responder/responder.go`
   - Update `docs/RFC_COMPLIANCE_GUIDE.md`
   - Note limitations (interface monitoring, IP changes)

### Out of Scope (Deferred to M4)

1. **Per-Interface Socket Binding**: Full M4 architecture (FR-M4-003)
2. **Interface Monitoring**: Detect interface up/down events
3. **IP Change Detection**: Handle DHCP renewals, address changes
4. **IPv6 Support**: AAAA records, IPv6 multicast (FF02::FB)
5. **Multi-Packet Responses**: When interface has multiple IPs (edge case)

## Dependencies

- **F-10 Interface Management**: Leverage existing interface selection logic (exclude Docker, VPN)
- **006-mdns-responder**: Built on top of M2 responder implementation
- **M4 Planning**: Must not conflict with planned per-interface transport architecture

## Assumptions

1. **Single IPv4 per interface**: Each interface has at most one IPv4 address for mDNS purposes (multiple IPs deferred to M4)
2. **Transport compatibility**: Current `internal/transport/udp.go` can determine receiving interface from packet metadata or socket options
3. **Go net package support**: `net.InterfaceByIndex()` and `net.Interface.Addrs()` provide necessary interface → IP lookup
4. **Static configuration**: Interface configuration does not change during service lifetime (monitoring deferred to M4)

## Testing Strategy

### Contract Tests (RFC Compliance)

**Test**: `TestRFC6762_Section15_InterfaceSpecificAddresses`

```go
// Setup: Machine with eth0 (10.0.0.5), eth1 (192.168.1.100)
// Query on eth0 → Response MUST include 10.0.0.5, MUST NOT include 192.168.1.100
// Query on eth1 → Response MUST include 192.168.1.100, MUST NOT include 10.0.0.5
```

Location: `tests/contract/rfc6762_interface_test.go`

### Unit Tests

1. **Interface IP Lookup**: Test `getIPv4ForInterface(index)` with valid/invalid indices
2. **Response Builder**: Test `BuildRecordSet()` with interface context produces correct A records
3. **Query Handler**: Test query pipeline passes interface index correctly
4. **Error Handling**: Test behavior when interface lookup fails

### Integration Tests

1. **Multi-Interface Registration**: Register service on machine with 2+ interfaces, verify correct advertising
2. **Single-Interface Regression**: Verify existing single-interface tests still pass
3. **Interface Selection**: Verify Docker/VPN exclusion logic still works (F-10)

### Benchmarks

**Benchmark**: `BenchmarkBuildResponse_WithInterfaceContext`

Compare:
- Baseline: `BenchmarkBuildResponse` (current)
- Enhanced: `BenchmarkBuildResponse_WithInterfaceContext` (new)

Target: <1% overhead (NFR-002)

## Related Work

- **Issue #27**: GitHub issue tracking this bug ([link](https://github.com/joshuafuller/beacon/issues/27))
- **Bug Report**: `/tmp/multi_interface_bug_report.md` (comprehensive analysis)
- **M4 Plan**: [docs/planning/NEXT_PHASE_PLAN.md](../../docs/planning/NEXT_PHASE_PLAN.md) - Full per-interface transport architecture
- **F-10 Spec**: [.specify/specs/F-10-network-interface-management.md](../../.specify/specs/F-10-network-interface-management.md) - Interface selection logic
- **ADR-001**: [docs/decisions/001-transport-interface-abstraction.md](../../docs/decisions/001-transport-interface-abstraction.md) - Transport abstraction

## Open Questions

*None at this time. Fast-track scope is clear: interface context propagation + interface-specific IP lookup.*
