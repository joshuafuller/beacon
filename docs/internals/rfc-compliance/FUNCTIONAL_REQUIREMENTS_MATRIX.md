# Beacon Functional Requirements Matrix

**Last Updated**: 2025-11-02
**Project Phase**: Foundation Complete (M1 + M1-Refactoring + M1.1)
**Total Functional Requirements**: 61 (22 M1 + 4 M1-R + 35 M1.1)

---

## Purpose

This matrix provides **complete traceability** from functional requirements (FRs) to implementation code, test evidence, and RFC 6762/6763 compliance. It aggregates all FRs from the Foundation phase (M1, M1-Refactoring, M1.1) into a single reference document.

### Milestone-Prefixed FR-IDs

To preserve traceability to source code comments, git commits, and original specification checklists, FRs use **milestone-prefixed identifiers**:

- **FR-M1-XXX**: M1 Basic mDNS Querier requirements (22 FRs)
- **FR-M1R-XXX**: M1-Refactoring architectural requirements (4 FRs)
- **FR-M1.1-XXX**: M1.1 Architectural Hardening requirements (35 FRs)

This prevents renumbering risk and maintains bidirectional links between documentation and implementation.

---

## Column Definitions

| Column | Description |
|--------|-------------|
| **FR-ID** | Milestone-prefixed functional requirement identifier |
| **Description** | MUST/SHOULD statement defining the requirement |
| **Status** | ✅ Implemented, ⚠️ Partial, ❌ Not Implemented, 🔄 In Progress, 📋 Planned |
| **Milestone** | Which milestone implemented this FR (M1, M1-R, M1.1) |
| **Functional Area** | Category: Query Construction, Socket Config, Security, etc. |
| **Implementation** | Source file(s) where requirement is implemented |
| **RFC Reference** | RFC 6762/6763 section(s) this FR satisfies |
| **Test Evidence** | Test file(s) validating the requirement |

---

## Status Legend

- ✅ **Implemented**: Requirement fully implemented with test coverage
- ⚠️ **Partial**: Partially implemented or platform-specific limitations
- ❌ **Not Implemented**: Planned but not yet started
- 🔄 **In Progress**: Currently being implemented
- 📋 **Planned**: Future milestone work

### Platform Status Notation

Where platform support varies:
- **Linux ✅**: Fully validated on Linux with integration tests
- **macOS ⚠️**: Code-complete with platform-specific build tags, untested (no macOS CI)
- **Windows ⚠️**: Code-complete with platform-specific build tags, untested (no Windows CI)

---

## Summary Statistics

| Milestone | Total FRs | Implemented | Partial | Not Implemented | Completion % |
|-----------|-----------|-------------|---------|-----------------|--------------|
| **M1** | 22 | 22 | 0 | 0 | 100% |
| **M1-Refactoring** | 4 | 4 | 0 | 0 | 100% |
| **M1.1** | 35 | 34 | 1 | 0 | 97.1% |
| **Foundation Total** | **61** | **60** | **1** | **0** | **98.4%** |

**Note**: The single partial FR is FR-M1.1-020 (per-interface transport binding), which is deferred to M2 IPv6 implementation. Interface filtering is implemented; per-interface transports require M2 architecture.

---

## Functional Requirements by Milestone

### M1: Basic mDNS Querier (22 FRs)

#### Query Construction (4 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-001 | System MUST construct valid mDNS query messages per RFC 1035 wire format | ✅ | `internal/message/builder.go` (BuildQuery) | RFC 6762 §5.1, RFC 1035 §4 | `tests/integration/TestQuery` |
| FR-M1-002 | System MUST set QR=0, OPCODE=0 (Standard Query), AA=0 for queries | ✅ | `internal/message/builder.go` (BuildQuery) | RFC 6762 §18.1 | `internal/message/builder_test.go` |
| FR-M1-003 | System MUST construct QNAME with DNS name compression support | ✅ | `internal/message/builder.go` (encodeName) | RFC 1035 §4.1.4 | `internal/message/builder_test.go` |
| FR-M1-004 | System MUST support QTYPE: A (1), PTR (12), SRV (33), TXT (16) | ✅ | `querier/types.go` (RecordType constants) | RFC 1035 §3.2.2 | `querier/querier_test.go` |

#### Query Execution (4 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-005 | System MUST send queries to 224.0.0.251:5353 (mDNS multicast group) | ✅ | `internal/protocol/constants.go`, `internal/transport/udp.go` | RFC 6762 §5 | `tests/integration/TestQuery` |
| FR-M1-006 | System MUST accept context.Context for cancellation and timeout | ✅ | `querier/querier.go` (Query signature) | Go best practices | `querier/querier_test.go` (TestQuery_Timeout) |
| FR-M1-007 | System MUST handle context cancellation during query execution | ✅ | `querier/querier.go` (select with ctx.Done()) | Go best practices | `querier/querier_test.go` (TestQuery_ContextCancellation) |
| FR-M1-008 | System MUST return empty results on timeout (not error) | ✅ | `querier/querier.go` (Query) | RFC 6762 §5.2 | `querier/querier_test.go` (TestQuery_Timeout) |

#### Response Handling (4 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-009 | System MUST parse DNS wire format responses per RFC 1035 | ✅ | `internal/message/parser.go` (Parse) | RFC 1035 §4 | `internal/message/parser_test.go` |
| FR-M1-010 | System MUST decompress DNS names in responses | ✅ | `internal/message/parser.go` (parseName with compression) | RFC 1035 §4.1.4 | `internal/message/parser_test.go` |
| FR-M1-011 | System MUST extract resource records from Answer/Additional sections | ✅ | `internal/message/parser.go` (parseResourceRecord) | RFC 1035 §3.2.1 | `internal/message/parser_test.go` |
| FR-M1-012 | System MUST convert parsed records to public ResourceRecord type | ✅ | `querier/querier.go` (convertToResourceRecords) | N/A | `querier/querier_test.go` |

#### Error Handling (4 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-013 | System MUST return NetworkError for transport failures | ✅ | `internal/errors/errors.go` (NetworkError) | N/A | `querier/querier_test.go` |
| FR-M1-014 | System MUST return ValidationError for invalid inputs | ✅ | `internal/errors/errors.go` (ValidationError) | N/A | `querier/querier_test.go` |
| FR-M1-015 | System MUST return WireFormatError for malformed packets | ✅ | `internal/errors/errors.go` (WireFormatError) | RFC 6762 §18.3 | `internal/message/parser_test.go` |
| FR-M1-016 | System MUST NOT panic on malformed packets | ✅ | `internal/message/parser.go` (bounds checking) | RFC 6762 §21 | `tests/fuzz/message_test.go` (114K iterations) |

#### Resource Management (3 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-017 | System MUST provide Close() method to release resources | ✅ | `querier/querier.go` (Close) | Go best practices | `querier/querier_test.go` |
| FR-M1-018 | System MUST close UDP socket on Close() | ✅ | `querier/querier.go` (Close propagates to transport) | N/A | `querier/querier_test.go` |
| FR-M1-019 | System MUST support ≥100 concurrent queries | ✅ | `querier/querier.go` (goroutine-safe) | NFR-002 | `tests/integration/TestConcurrentQueries` |

#### RFC 6762 Compliance (3 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1-020 | System MUST use QU bit (unicast response) for one-shot queries | ✅ | `internal/message/builder.go` (QCLASS = 0x8001) | RFC 6762 §5.4 | `internal/message/builder_test.go` |
| FR-M1-021 | System MUST enforce 9000 byte maximum packet size | ✅ | `internal/protocol/constants.go` (MaxPacketSize) | RFC 6762 §17 | `internal/message/parser_test.go` |
| FR-M1-022 | System MUST ignore malformed packets without crashing | ✅ | `internal/message/parser.go` (error returns) | RFC 6762 §18.3 | `tests/fuzz/message_test.go` |

---

### M1-Refactoring: Architectural Improvements (4 FRs)

#### Transport Abstraction (1 FR)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1R-001 | System MUST abstract network transport via Transport interface | ✅ | `internal/transport/transport.go` (Transport interface) | ADR-001 | `internal/transport/udp_test.go`, `querier/querier_test.go` (MockTransport) |

#### Performance Optimization (1 FR)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1R-002 | System MUST use buffer pooling to reduce allocations by ≥80% in receive path | ✅ | `internal/transport/buffer_pool.go` (sync.Pool) | ADR-002 | `internal/transport/buffer_pool_test.go`, benchmarks (99% reduction: 9000 B/op → 48 B/op) |

#### Error Propagation (1 FR)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1R-003 | System MUST propagate all errors without swallowing | ✅ | `internal/transport/udp.go` (Close returns error) | FR-004 | Code review (grep -rn "_ =" confirms no swallowed errors) |

#### Layer Boundaries (1 FR)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1R-004 | System MUST enforce layer boundaries (querier MUST NOT import internal/network) | ✅ | `querier/querier.go` (imports only internal/transport) | F-2 Architecture Layers | `grep -rn "internal/network" querier/` (0 matches) |

---

### M1.1: Architectural Hardening (35 FRs)

#### Socket Configuration (10 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1.1-001 | System MUST set SO_REUSEADDR on all platforms for address reuse | ✅ | `internal/transport/socket_linux.go`, `socket_darwin.go`, `socket_windows.go` | RFC 6762 §15 | `tests/integration/TestSocketOptions` |
| FR-M1.1-002 | System SHOULD set SO_REUSEPORT on Linux (kernel >= 3.9) for port sharing | ✅ | `internal/transport/socket_linux.go` (setSockoptInt SO_REUSEPORT) | RFC 6762 §15 | `tests/integration/TestAvahiCoexistence` (Linux ✅) |
| FR-M1.1-003 | System SHOULD set SO_REUSEPORT on macOS for port sharing | ✅ | `internal/transport/socket_darwin.go` (setSockoptInt SO_REUSEPORT) | RFC 6762 §15 | Code review (macOS ⚠️ untested) |
| FR-M1.1-004 | System MUST use SO_REUSEADDR on Windows (no SO_REUSEPORT needed) | ✅ | `internal/transport/socket_windows.go` (SO_REUSEADDR only) | Windows sockets behavior | Code review (Windows ⚠️ untested) |
| FR-M1.1-005 | System MUST use platform-specific build tags for socket options | ✅ | `//go:build linux` tags in socket files | Go build system | Build validation (`go build ./...` on Linux) |
| FR-M1.1-006 | System MUST bind to 0.0.0.0:5353 to receive multicast traffic | ✅ | `internal/transport/udp.go` (net.ListenUDP) | RFC 6762 §5 | `tests/integration/TestQuery` (Linux ✅) |
| FR-M1.1-007 | System MUST join 224.0.0.251 multicast group on selected interfaces | ✅ | `internal/transport/udp.go` (JoinGroup) | RFC 6762 §5 | `tests/integration/TestQuery` (Linux ✅) |
| FR-M1.1-008 | System MUST set IP_MULTICAST_TTL=255 for link-local scope | ✅ | `internal/transport/udp.go` (SetMulticastTTL) | RFC 6762 §11 | `tests/integration/TestSocketOptions` (Linux ✅) |
| FR-M1.1-009 | System MUST set IP_MULTICAST_LOOP=true to receive own packets | ✅ | `internal/transport/udp.go` (SetMulticastLoopback) | RFC 6762 §15 | `tests/integration/TestSocketOptions` (Linux ✅) |
| FR-M1.1-010 | System MUST leave multicast group on Close() | ✅ | `internal/transport/udp.go` (LeaveGroup in Close) | Network cleanup | `internal/transport/udp_test.go` (TestClose) |

#### Network Interface Management (12 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1.1-011 | System MUST provide DefaultInterfaces() to auto-select suitable interfaces | ✅ | `internal/network/interfaces.go` (DefaultInterfaces) | F-10 | `internal/network/interfaces_test.go` (Linux ✅) |
| FR-M1.1-012 | System MUST exclude loopback interfaces (lo) from DefaultInterfaces() | ✅ | `internal/network/interfaces.go` (FlagLoopback check) | F-10 | `internal/network/interfaces_test.go` (Linux ✅) |
| FR-M1.1-013 | System MUST exclude down interfaces from DefaultInterfaces() | ✅ | `internal/network/interfaces.go` (FlagUp check) | F-10 | `internal/network/interfaces_test.go` (Linux ✅) |
| FR-M1.1-014 | System MUST exclude interfaces without multicast support | ✅ | `internal/network/interfaces.go` (FlagMulticast check) | RFC 6762 §5 | `internal/network/interfaces_test.go` (Linux ✅) |
| FR-M1.1-015 | System MUST exclude VPN interfaces (utun*, tun*, ppp*) from DefaultInterfaces() | ✅ | `internal/network/interfaces.go` (isVPNInterface) | F-10 Privacy | `internal/network/interfaces_test.go` (TestDefaultInterfaces_VPN) |
| FR-M1.1-016 | System MUST exclude Docker interfaces (docker0, veth*, br-*) from DefaultInterfaces() | ✅ | `internal/network/interfaces.go` (isDockerInterface) | F-10 Performance | `internal/network/interfaces_test.go` (TestDefaultInterfaces_Docker) |
| FR-M1.1-017 | System MUST provide WithInterfaces([]net.Interface) option for explicit interface selection | ✅ | `querier/options.go` (WithInterfaces) | F-10 | `querier/querier_test.go` |
| FR-M1.1-018 | System MUST provide WithInterfaceFilter(func) option for custom filtering | ✅ | `querier/options.go` (WithInterfaceFilter) | F-10 | `querier/querier_test.go` |
| FR-M1.1-019 | System MUST apply user-provided interface filter AFTER DefaultInterfaces() filtering | ✅ | `querier/querier.go` (applyOptions) | F-10 | Code review |
| FR-M1.1-020 | System SHOULD support per-interface transport binding (one socket per interface) | ⚠️ | Deferred to M2 | RFC 6762 §14 | M2 IPv6 architecture required |
| FR-M1.1-021 | System MUST validate that at least one interface remains after filtering | ✅ | `querier/querier.go` (New returns error if empty) | F-10 | `querier/querier_test.go` (TestNew_NoInterfaces) |
| FR-M1.1-022 | System MUST log selected interfaces during Querier creation | ✅ | `querier/querier.go` (debug logging) | F-6 (future) | Manual testing (Linux ✅) |

#### Security Features (13 FRs)

| FR-ID | Description | Status | Implementation | RFC Reference | Test Evidence |
|-------|-------------|--------|----------------|---------------|---------------|
| FR-M1.1-023 | System MUST implement SourceFilter to validate packet source IPs | ✅ | `internal/security/source_filter.go` (SourceFilter struct) | RFC 6762 §11 | `internal/security/source_filter_test.go` (Linux ✅, macOS/Windows ⚠️ code-complete) |
| FR-M1.1-024 | System MUST reject packets from non-link-local sources (not 169.254.0.0/16) | ✅ | `internal/security/source_filter.go` (IsLinkLocal check) | RFC 6762 §11 | `internal/security/source_filter_test.go` (TestSourceFilter_RejectNonLinkLocal) |
| FR-M1.1-025 | System MUST drop source-filtered packets BEFORE parsing (performance) | ✅ | `querier/querier.go` (receive loop checks AllowSource first) | F-11 | Code review |
| FR-M1.1-026 | System MUST implement per-source-IP rate limiting with sliding window (1 second) | ✅ | `internal/security/rate_limiter.go` (RateLimiter struct) | RFC 6762 §6 | `internal/security/rate_limiter_test.go` (TestRateLimiter_SlidingWindow) |
| FR-M1.1-027 | System MUST use 100 queries-per-second (qps) default threshold for rate limiting | ✅ | `querier/options.go` (DefaultRateLimitThreshold = 100) | F-11 | `internal/security/rate_limiter_test.go` |
| FR-M1.1-028 | System MUST use 60-second cooldown period for rate-limited sources | ✅ | `querier/options.go` (DefaultRateLimitCooldown = 60s) | F-11 | `internal/security/rate_limiter_test.go` |
| FR-M1.1-029 | System MUST drop rate-limited packets BEFORE parsing (performance) | ✅ | `querier/querier.go` (receive loop checks AllowPacket first) | F-11 | Code review |
| FR-M1.1-030 | System MUST provide WithRateLimit(bool) option to enable/disable rate limiting | ✅ | `querier/options.go` (WithRateLimit) | F-11 | `querier/querier_test.go` |
| FR-M1.1-031 | System MUST provide WithRateLimitThreshold(int) option for custom qps threshold | ✅ | `querier/options.go` (WithRateLimitThreshold) | F-11 | `querier/querier_test.go` |
| FR-M1.1-032 | System MUST provide WithRateLimitCooldown(time.Duration) option for custom cooldown | ✅ | `querier/options.go` (WithRateLimitCooldown) | F-11 | `querier/querier_test.go` |
| FR-M1.1-033 | System MUST enforce 9000-byte packet size limit BEFORE parsing | ✅ | `querier/querier.go` (receive loop checks len(packet) <= MaxPacketSize) | RFC 6762 §17 | `querier/querier_test.go` (TestReceive_OversizedPacket) |
| FR-M1.1-034 | System MUST maintain zero-crash guarantee on malformed packets (fuzz tested) | ✅ | `internal/message/parser.go` (bounds checking) | RFC 6762 §21 | `tests/fuzz/message_test.go` (114K iterations, zero crashes) |
| FR-M1.1-035 | System MUST maintain zero race conditions (race detector clean) | ✅ | All concurrent code (querier, transport, security) | Go concurrency | `go test ./... -race` (0 races detected) |

---

## Cross-References

### RFC 6762 Section → Functional Requirements

For bidirectional traceability, see [RFC Compliance Matrix](./RFC_COMPLIANCE_MATRIX.md) which maps RFC sections to implementing FRs.

**Key Mappings**:
- **§5 (Multicast DNS Queries)**: FR-M1-001 through FR-M1-008, FR-M1.1-006, FR-M1.1-007
- **§11 (Source Address Check)**: FR-M1.1-008 (TTL=255), FR-M1.1-023, FR-M1.1-024
- **§14 (Multiple Interfaces)**: FR-M1.1-011 through FR-M1.1-022
- **§15 (Multiple Responders)**: FR-M1.1-001 through FR-M1.1-005 (SO_REUSEPORT)
- **§17 (Packet Size Limit)**: FR-M1-021, FR-M1.1-033
- **§21 (Security Considerations)**: FR-M1-016, FR-M1.1-026 through FR-M1.1-035

### Functional Area Index

**Query Construction**: FR-M1-001 to FR-M1-004
**Query Execution**: FR-M1-005 to FR-M1-008
**Response Handling**: FR-M1-009 to FR-M1-012
**Error Handling**: FR-M1-013 to FR-M1-016
**Resource Management**: FR-M1-017 to FR-M1-019
**RFC Compliance**: FR-M1-020 to FR-M1-022
**Transport Abstraction**: FR-M1R-001
**Performance Optimization**: FR-M1R-002
**Error Propagation**: FR-M1R-003
**Layer Boundaries**: FR-M1R-004
**Socket Configuration**: FR-M1.1-001 to FR-M1.1-010
**Interface Management**: FR-M1.1-011 to FR-M1.1-022
**Security Features**: FR-M1.1-023 to FR-M1.1-035

---

## Quality Metrics

- **Total Tasks Across Foundation**: 210+ tasks (M1: 50, M1-R: 97, M1.1: 94)
- **Test Coverage**: 80.0% (post-M1.1)
- **Race Condition Tests**: ✅ PASS (`go test ./... -race`)
- **Fuzz Testing**: ✅ 114K iterations, zero crashes
- **Flaky Tests**: 0 (eliminated in M1-Refactoring)
- **Integration Tests**: 10/10 PASS (Linux ✅, Avahi coexistence validated)
- **Benchmark Performance**: 99% allocation reduction (9000 B/op → 48 B/op)

---

## Related Documentation

- **[Compliance Dashboard](./COMPLIANCE_DASHBOARD.md)** - Single-page project status overview
- **[RFC Compliance Matrix](./RFC_COMPLIANCE_MATRIX.md)** - Section-by-section RFC 6762 implementation status
- **[Foundation Completion Report](./FOUNDATION_COMPLETE.md)** - Narrative of M1→M1-R→M1.1 journey
- **[ROADMAP](../ROADMAP.md)** - Milestone plan (M1-M6)
- **[Beacon Constitution](../.specify/memory/constitution.md)** - Project principles

---

**Matrix Version**: 1.0
**Foundation Phase**: ✅ Complete (M1 + M1-R + M1.1)
**Next Milestone**: M2 (mDNS Responder)
