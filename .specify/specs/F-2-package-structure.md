# F-2: Package Structure & Layering

**Spec ID**: F-2
**Type**: Architecture
**Status**: Validated (RFC Compliant)
**Version**: 1.0.0
**Dependencies**: None
**References**:
- BEACON_FOUNDATIONS v1.1 §4 (Beacon Architecture)
- Beacon Constitution v1.0.0
- RFC 6762 (Multicast DNS)
- RFC 6763 (DNS-Based Service Discovery)

**Governance**: Development governed by [Beacon Constitution v1.0.0](../memory/constitution.md)

**RFC Validation**: Completed 2025-11-01. Package structure validated against RFC 6762 and RFC 6763 requirements for protocol layering and separation of concerns. No blocking issues identified.

---

## Overview

This specification defines Beacon's Go package structure, layering principles, import rules, and organizational guidelines. The package structure balances:
- **Clear boundaries** between public API and internal implementation
- **Testability** through focused, mockable packages
- **Maintainability** through logical organization
- **Extensibility** allowing future growth without breaking changes

**Constitutional Alignment**: This architecture specification upholds:
- **Principle I (RFC Compliance)**: Dedicated protocol layer ensures RFC 6762/6763 compliance is isolated and verifiable
- **Principle II (Spec-Driven)**: Clear package boundaries enable spec-first design for each component
- **Principle III (TDD)**: Focused packages with minimal dependencies maximize testability
- **Principle VII (Excellence)**: Layered architecture follows Go best practices and enables code review at appropriate abstraction levels

---

## Requirements

### REQ-F2-1: Public API Separation (RFC Compliant)
Beacon MUST separate public API packages from internal implementation packages.

**Rationale**: Go's `internal/` convention prevents external imports, allowing internal refactoring without breaking users.

**RFC Alignment**: Separation of concerns aligns with protocol layering principles inherent in RFC 6762 and RFC 6763, ensuring protocol implementation details remain encapsulated.

### REQ-F2-2: Layer Organization (RFC Compliant)
Beacon MUST organize code into distinct layers with clear responsibilities.

**Layers** (top to bottom):
1. **Public API Layer**: User-facing interfaces and types
2. **Service Layer**: Orchestration and business logic
3. **Protocol Layer**: RFC 6762/6763 compliance and wire format
4. **Transport Layer**: Network I/O

**RFC Alignment**: Layer 3 (Protocol) is dedicated exclusively to RFC compliance, ensuring all RFC MUST requirements are verifiable and isolated from business logic (Constitution Principle I).

### REQ-F2-3: No Circular Dependencies
Package imports MUST form a directed acyclic graph (DAG). Circular dependencies are prohibited.

### REQ-F2-4: Focused Packages
Each package SHOULD have a single, well-defined responsibility.

**Anti-pattern**: "util" or "common" packages that become dumping grounds.

### REQ-F2-5: Internal Package Convention
Packages in `internal/` MUST NOT be importable by external code (enforced by Go).

---

## Package Structure

```
github.com/joshuafuller/beacon/
│
├── querier/                    # PUBLIC: mDNS query client
│   ├── querier.go
│   ├── options.go
│   └── querier_test.go
│
├── responder/                  # PUBLIC: mDNS responder server
│   ├── responder.go
│   ├── options.go
│   └── responder_test.go
│
├── service/                    # PUBLIC: DNS-SD browser/publisher
│   ├── browser.go
│   ├── resolver.go
│   ├── publisher.go
│   ├── types.go
│   └── service_test.go
│
├── internal/
│   │
│   ├── message/                # DNS message format
│   │   ├── message.go          # Message structure
│   │   ├── header.go           # Header parsing/building
│   │   ├── question.go         # Question section
│   │   ├── record.go           # Resource records
│   │   ├── name.go             # Name encoding/compression
│   │   ├── parser.go           # Byte stream → Message
│   │   ├── builder.go          # Message → Byte stream
│   │   └── message_test.go
│   │
│   ├── protocol/               # mDNS protocol logic
│   │   ├── query.go            # Query construction
│   │   ├── response.go         # Response construction
│   │   ├── probe.go            # Probing protocol
│   │   ├── announce.go         # Announcing protocol
│   │   ├── conflict.go         # Conflict detection/resolution
│   │   └── protocol_test.go
│   │
│   ├── cache/                  # Record caching
│   │   ├── cache.go            # Cache interface
│   │   ├── store.go            # In-memory storage
│   │   ├── ttl.go              # TTL management
│   │   └── cache_test.go
│   │
│   ├── transport/              # Network I/O
│   │   ├── transport.go        # Transport interface
│   │   ├── multicast.go        # UDP multicast operations
│   │   ├── socket.go           # Socket management
│   │   ├── interface.go        # Network interface monitoring
│   │   └── transport_test.go
│   │
│   └── lifecycle/              # Service lifecycle management
│       ├── manager.go          # Lifecycle orchestration
│       ├── prober.go           # Probing coordinator
│       ├── announcer.go        # Announcing coordinator
│       └── lifecycle_test.go
│
├── examples/                   # Example applications
│   ├── query/
│   │   └── main.go             # Simple query example
│   ├── browse/
│   │   └── main.go             # Service browsing example
│   └── publish/
│       └── main.go             # Service publishing example
│
├── docs/                       # Documentation
│   └── BEACON_FOUNDATIONS.md  # v1.1 - Shared context
├── .specify/
│   ├── memory/
│   │   └── constitution.md    # Constitution v1.0.0
│   └── specs/                 # Feature specifications
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

---

## Package Descriptions

### Public Packages

#### `querier/`
**Purpose**: mDNS query client for discovering hosts and services.

**Exports**:
- `type Querier interface` - Query operations
- `func New(opts ...Option) (*Client, error)` - Constructor
- `type Option func(*config)` - Configuration options
- `type Record struct` - Query results

**Users**: Applications needing to discover mDNS hosts.

**Example**:
```go
import "github.com/joshuafuller/beacon/querier"

q, _ := querier.New()
records, _ := q.Query(ctx, "myhost.local", querier.TypeA)
```

#### `responder/`
**Purpose**: mDNS responder for advertising hosts and records.

**Exports**:
- `type Responder interface` - Response operations
- `func New(opts ...Option) (*Server, error)` - Constructor
- `type Record struct` - Records to advertise
- `func (r *Server) AddRecord(record Record) error`
- `func (r *Server) Start(ctx context.Context) error`

**Users**: Applications advertising mDNS services.

**Example**:
```go
import "github.com/joshuafuller/beacon/responder"

r, _ := responder.New()
r.AddRecord(responder.A("myhost.local", net.ParseIP("192.168.1.100")))
r.Start(ctx)
```

#### `service/`
**Purpose**: DNS-SD service browsing and publishing.

**Exports**:
- `type Browser interface` - Service browsing
- `type Resolver interface` - Instance resolution
- `type Publisher interface` - Service publishing
- `type ServiceInstance struct` - Service identity
- `type ServiceInfo struct` - Resolved service info

**Users**: Applications doing zero-config service discovery.

**Example**:
```go
import "github.com/joshuafuller/beacon/service"

b, _ := service.NewBrowser()
instances := b.Browse(ctx, "_http._tcp")
for inst := range instances {
    info, _ := service.Resolve(ctx, inst)
    fmt.Printf("%s at %s:%d\n", inst.Name, info.Host, info.Port)
}
```

---

### Internal Packages

#### `internal/message/`
**Purpose**: DNS message format parsing and building.

**Exports** (within Beacon):
- `type Message struct` - DNS message
- `type Header struct` - Message header
- `type Question struct` - Question section
- `type Record struct` - Resource record
- `func Parse([]byte) (*Message, error)` - Parse bytes to message
- `func (m *Message) Serialize() ([]byte, error)` - Serialize message

**Used By**: `internal/protocol/`, `querier/`, `responder/`

**Rationale**: Message format is stable but internal. Users work with higher-level APIs.

#### `internal/protocol/`
**Purpose**: mDNS protocol logic (queries, responses, probing, announcing).

**Exports** (within Beacon):
- `func BuildQuery(name, qtype) (*message.Message, error)`
- `func BuildResponse(query, records) (*message.Message, error)`
- `func BuildProbe(records) (*message.Message, error)`
- `func BuildAnnounce(records) (*message.Message, error)`
- `func DetectConflict(incoming, ours) bool`

**Used By**: `querier/`, `responder/`, `internal/lifecycle/`

**Rationale**: Protocol operations are complex and RFC-specific. Isolating enables testing and compliance verification.

#### `internal/cache/`
**Purpose**: Record caching with TTL management.

**Exports** (within Beacon):
- `type Cache interface`
- `func New() Cache`
- `func (c *Cache) Add(record) error`
- `func (c *Cache) Lookup(name, qtype) ([]Record, bool)`
- `func (c *Cache) Remove(record) error`

**Used By**: `querier/`, `internal/protocol/`

**Rationale**: Cache is implementation detail. Different strategies possible (in-memory, persistent, LRU).

#### `internal/transport/`
**Purpose**: Network I/O (UDP multicast, socket management).

**Exports** (within Beacon):
- `type Transport interface`
- `func NewMulticast() (Transport, error)`
- `func (t *Transport) Send(msg, addr) error`
- `func (t *Transport) Receive() ([]byte, net.Addr, error)`

**Used By**: `querier/`, `responder/`

**Rationale**: Network layer abstraction enables testing (mock transport) and platform-specific implementations.

#### `internal/lifecycle/`
**Purpose**: Service lifecycle orchestration (probe, announce, conflict handling).

**Exports** (within Beacon):
- `type Manager interface`
- `func (m *Manager) Probe(records) error`
- `func (m *Manager) Announce(records) error`
- `func (m *Manager) HandleConflict(record) error`

**Used By**: `responder/`, `service/`

**Rationale**: Lifecycle is complex state machine. Isolating enables focused testing.

---

## Import Rules

### RULE-1: Public → Internal (Allowed)
Public packages MAY import internal packages.

**Example**: `querier/` imports `internal/protocol/`, `internal/transport/`

### RULE-2: Internal → Public (Prohibited)
Internal packages MUST NOT import public packages.

**Rationale**: Prevents circular dependencies, keeps internals decoupled from API.

### RULE-3: Internal → Internal (Allowed, with ordering)
Internal packages MAY import other internal packages, respecting layer order.

**Allowed**:
- `internal/protocol/` → `internal/message/`
- `internal/lifecycle/` → `internal/protocol/`
- `internal/cache/` → `internal/message/`

**Prohibited**:
- `internal/message/` → `internal/protocol/` (wrong direction)
- Any circular imports

### RULE-4: Standard Library and Third-Party
All packages MAY import standard library and approved third-party packages.

**Approved Third-Party** (currently none, minimize dependencies):
- None initially
- Future: Consider structured logging, time mocking for tests

### RULE-5: Test Imports
Test files (`*_test.go`) MAY import additional packages for testing:
- `testing`
- `github.com/joshuafuller/beacon/internal/...` (for integration tests)

---

## Layer Boundaries

### Layer 1: Public API
**Packages**: `querier/`, `responder/`, `service/`

**Responsibilities**:
- User-facing interfaces
- Configuration and options
- High-level operations
- Error translation (internal errors → user-friendly errors)

**Dependencies**: Internal packages (protocol, transport, cache, lifecycle)

### Layer 2: Service Logic
**Packages**: `internal/lifecycle/`, portions of public packages

**Responsibilities**:
- Orchestration (coordinating protocol operations)
- State management (probing state, announcing state)
- Business logic (when to probe, how to handle conflicts)

**Dependencies**: Protocol, cache

### Layer 3: Protocol (RFC Compliance Layer)
**Packages**: `internal/protocol/`, `internal/message/`

**Responsibilities**:
- **RFC 6762 compliance**: Multicast DNS protocol operations
- **RFC 6763 compliance**: DNS-SD naming and conventions
- Message construction and parsing
- Wire format encoding/decoding
- Protocol-level validation

**Dependencies**: Message (for protocol), none (for message - foundation)

**Constitutional Mandate**: This layer implements Constitution Principle I (RFC Compliance). All RFC MUST requirements are enforced here. No RFC-defined behavior may be made configurable or optional.

### Layer 4: Transport
**Packages**: `internal/transport/`

**Responsibilities**:
- UDP socket operations
- Multicast group join/leave
- Network interface monitoring
- OS-level integration

**Dependencies**: Standard library only

---

## Design Principles

### PRIN-1: API Stability
Public package APIs SHOULD be stable. Breaking changes require major version bump.

### PRIN-2: Internal Freedom
Internal packages MAY change freely without version impact.

**Rationale**: Users cannot import `internal/`, so refactoring doesn't break them.

### PRIN-3: Dependency Minimization
Minimize third-party dependencies, especially in public packages.

**Rationale**:
- Reduces supply chain risk
- Simplifies builds
- Avoids version conflicts for users

### PRIN-4: Test Organization
Tests SHOULD be colocated with code (`*_test.go` in same package).

**Package-level tests** (`package foo_test`):
- Test public API from user perspective
- Black-box testing

**Internal tests** (`package foo`):
- Test internal functions
- White-box testing

### PRIN-5: Example Code
Examples in `examples/` MUST only import public packages.

**Rationale**: Examples demonstrate how users should use Beacon. If examples need internal packages, API is insufficient.

---

## Migration and Evolution

### Adding New Packages
When adding new internal packages:
1. Define clear responsibility
2. Document purpose and exports
3. Validate no circular dependencies
4. Add to dependency graph

### Promoting Internal to Public
If functionality should be public:
1. Design stable API
2. Move to top-level package (not internal/)
3. Document thoroughly
4. Consider backwards compatibility

### Deprecating Packages
When deprecating:
1. Mark with deprecation comment
2. Provide migration path
3. Maintain for one major version
4. Remove in next major version

---

## Validation

### Build-Time Validation
```bash
# No circular dependencies
go build ./...

# Internal packages not importable externally
# (enforced by Go toolchain)
```

### Tooling
```bash
# Visualize dependencies
go mod graph

# Check for issues
go vet ./...
```

---

## Open Questions

**Q1**: Should we have a top-level `beacon/` package for common types?
- **Pro**: Single import for common types
- **Con**: Becomes dumping ground
- **Decision**: TBD, lean toward no

**Q2**: Should examples live in separate repo?
- **Pro**: Keeps main repo focused
- **Con**: Harder to keep in sync
- **Decision**: Keep in same repo for now

**Q3**: Configuration package (internal/config)?
- **Pro**: Centralized configuration
- **Con**: Shared mutable state
- **Decision**: Use options pattern per package

---

## Implementation Notes

- Use `go:embed` for any embedded resources
- Follow Go standard package layout conventions
- Run `gofmt` and `go vet` as part of CI
- Use `internal/` convention for all non-public code

---

## Success Criteria

- [ ] Package structure defined
- [ ] Import rules documented
- [ ] Layer boundaries clear
- [ ] No circular dependencies
- [ ] Public API minimal and focused
- [ ] Internal packages enable testability
- [ ] Examples use only public API

---

## Governance and Compliance

### Constitutional Compliance

This specification implements the architectural foundation for:

**Principle I (RFC Compliant)**:
- Dedicated `internal/protocol/` package isolates RFC compliance logic
- Protocol layer cannot be bypassed or overridden by higher layers
- All RFC MUST requirements enforced at protocol boundary
- Validation: Architecture review confirmed no RFC behavior is configurable

**Principle II (Spec-Driven Development)**:
- Package boundaries align with specification boundaries
- Each package corresponds to a clear specification domain
- Public API packages map to user-facing specifications
- Internal packages map to implementation specifications

**Principle III (Test-Driven Development)**:
- Focused packages with single responsibilities maximize testability
- `internal/` packages can be tested independently
- Mock interfaces at layer boundaries enable isolated unit tests
- Test files colocated with implementation (`*_test.go`)

**Principle VII (Excellence)**:
- Follows Go community best practices for package organization
- Uses standard `internal/` convention for encapsulation
- Minimizes dependencies to reduce coupling
- Clear separation enables code review at appropriate abstraction levels

### Architecture Validation Record

**Validation Date**: 2025-11-01
**Validator**: RFC Compliance Review
**Validation Method**: Cross-reference against RFC 6762 §§1-22 and RFC 6763 §§1-14

**Findings**:
- ✅ **P0 Issues**: None identified
- ✅ **RFC 6762 Alignment**: Package structure supports all required protocol operations (queries, responses, probing, announcing, conflict detection)
- ✅ **RFC 6763 Alignment**: Service layer can implement DNS-SD browsing and publishing without RFC violations
- ✅ **Separation of Concerns**: Protocol layer properly isolated from transport and business logic

**Conclusion**: Architecture approved for implementation. No blocking issues. Package structure enables RFC-compliant implementation.

### Change Control

Changes to this specification require:
1. RFC validation review (if layer responsibilities change)
2. Constitutional compliance check (if principles affected)
3. Version bump per semantic versioning:
   - **MAJOR**: Breaking changes to layer boundaries or import rules
   - **MINOR**: New packages or non-breaking enhancements
   - **PATCH**: Clarifications, documentation improvements

---

## References

**Constitutional**:
- [Beacon Constitution v1.0.0](../memory/constitution.md)
- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md)

**RFCs**:
- [RFC 6762](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - Multicast DNS
- [RFC 6763](../../RFC%20Docs/RFC-6763-DNS-SD.txt) - DNS-Based Service Discovery

**Go Best Practices**:
- Go Blog: [Internal Packages](https://go.dev/doc/go1.4#internalpackages)
- Effective Go: [Package Names](https://go.dev/doc/effective_go#package-names)
- Go Code Review Comments: [Package Structure](https://github.com/golang/go/wiki/CodeReviewComments)

---

## Version History

| Version | Date | Changes | Validated Against |
|---------|------|---------|-------------------|
| 1.0.0 | 2025-11-01 | Initial architecture specification validated against Constitution v1.0.0 and BEACON_FOUNDATIONS v1.1. RFC compliance validated against RFC 6762 and RFC 6763. | Constitution v1.0.0, BEACON_FOUNDATIONS v1.1, RFC 6762, RFC 6763 |
