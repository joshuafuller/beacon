# Documentation Standards

**Purpose**: This document defines the documentation standards for the Beacon project to ensure extreme auditability and RFC traceability.

**Last Updated**: 2025-11-05

---

## Core Principles

1. **RFC Attribution**: Every significant function must reference the RFC section it implements
2. **Explain WHY, Not Just WHAT**: Comments should explain the purpose and rationale, not just restate the code
3. **Functional Requirement Ties**: Link code to functional requirements (FRs) from spec.md files
4. **Task Traceability**: Reference task IDs (T001, T002, etc.) for implementation tracking
5. **Examples Over Prose**: Show usage examples for public APIs

---

## Package-Level Documentation

Every package must have comprehensive package-level documentation explaining:

### Required Elements:
1. **WHY THIS PACKAGE EXISTS**: The problem it solves and RFC requirement it addresses
2. **PRIMARY TECHNICAL AUTHORITY**: Which RFC sections govern this package
3. **DESIGN RATIONALE**: Key architectural decisions and tradeoffs
4. **RFC COMPLIANCE**: List of RFC requirements implemented
5. **KEY CONCEPTS**: Domain-specific terminology and concepts

### Template:

```go
// Package <name> implements <functionality> per RFC <number> §<section>.
//
// WHY THIS PACKAGE EXISTS:
// <Explain the problem this solves and the RFC requirement it addresses>
//
// DESIGN RATIONALE:
// - <Key decision 1 and why>
// - <Key decision 2 and why>
//
// RFC COMPLIANCE:
// - RFC <number> §<section>: <requirement>
// - RFC <number> §<section>: <requirement>
//
// PRIMARY TECHNICAL AUTHORITY: RFC <number> §§<sections>
package <name>
```

### Example:

```go
// Package state implements the mDNS service registration state machine per RFC 6762 §8.
//
// WHY THIS PACKAGE EXISTS:
// RFC 6762 §8 mandates a specific sequence before a service can be considered "established":
// 1. Probing phase (~750ms): Send 3 probe queries to detect naming conflicts
// 2. Announcing phase (~1s): Send 2 unsolicited announcements to inform the network
// 3. Established: Service is now discoverable by other mDNS clients
//
// DESIGN RATIONALE:
// - Goroutine-per-service: Each service runs independently (R001)
// - Context-aware: All operations respect context cancellation (F-9)
// - Testable: Decoupled from transport for unit testing
//
// RFC COMPLIANCE:
// - RFC 6762 §8.1: Probing (3 probes, 250ms apart)
// - RFC 6762 §8.2: Conflict detection
// - RFC 6762 §8.3: Announcing (2 announcements, 1s apart)
//
// PRIMARY TECHNICAL AUTHORITY: RFC 6762 §§8-9
package state
```

---

## Type Documentation

### Required Elements for Exported Types:
1. **Purpose**: What the type represents (per RFC if applicable)
2. **RFC Mapping**: Which RFC section defines this type
3. **Wire Format**: For protocol types, show the wire format
4. **Functional Requirements**: Which FRs this type satisfies
5. **Usage Examples**: For complex types

### Template:

```go
// <Type> represents <description> per RFC <number> §<section>.
//
// <Detailed explanation of purpose and usage>
//
// Wire format (if applicable):
//
//	<ASCII diagram of wire format>
//
// RFC Compliance:
// - FR-XXX: <requirement this type addresses>
//
// Example:
//
//	<usage example>
type <Type> struct {
    // <Field> is <purpose> per RFC <number> §<section>.
    //
    // <Additional context, constraints, or examples>
    <Field> <type>
}
```

### Example:

```go
// DNSHeader represents the DNS message header per RFC 1035 §4.1.1.
//
// The header is always 12 bytes and contains metadata about the message.
//
// Wire format (big-endian):
//
//	                                1  1  1  1  1  1
//	  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      ID                       |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    QDCOUNT                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// FR-020: System MUST set DNS header fields per RFC 6762 §18
type DNSHeader struct {
    // ID is the transaction ID (16 bits).
    //
    // RFC 6762 §18.1: mDNS messages SHOULD use ID = 0 for one-shot queries.
    ID uint16

    // Flags contains bit-packed header flags (16 bits).
    //
    // RFC 6762 §18: QR=0 for queries, QR=1 for responses
    Flags uint16
}
```

---

## Function Documentation

### Required Elements for Exported Functions:
1. **Purpose**: What the function does (one line)
2. **RFC Section**: Which RFC section this implements
3. **WHY Comment**: Explain the purpose and RFC requirement
4. **Parameters**: Document each parameter with purpose and constraints
5. **Returns**: Document each return value
6. **Algorithm**: For complex functions, explain the algorithm steps
7. **RFC Quotes**: Include relevant RFC quotes for critical behavior
8. **Examples**: Show usage for public APIs
9. **Functional Requirements**: Which FRs this function satisfies
10. **Task Reference**: Which task(s) implemented this

### Template:

```go
// <Function> <one-line description> per RFC <number> §<section>.
//
// WHY: <Explain the RFC requirement and why this function exists>
//
// OPERATION: (for complex functions)
// <Step-by-step explanation of the algorithm>
//
// RFC <number> §<section> states:
// "<Relevant RFC quote if critical>"
//
// Parameters:
//   - <param>: <purpose and constraints>
//
// Returns:
//   - <return>: <meaning and possible values>
//
// Example:
//
//	<usage example>
//
// FR-XXX: <Functional requirement this addresses>
// T0XX: <Task that implemented this>
func <Function>(<params>) <returns> {
    ...
}
```

### Example:

```go
// BuildQuery constructs an mDNS query message per RFC 6762 §18.
//
// WHY: mDNS queries must conform to specific header field requirements
// (QR=0, OPCODE=0, AA=0, etc.) to be recognized by mDNS responders.
// This function ensures all queries are RFC-compliant.
//
// OPERATION:
//   1. Validate record type is supported (A, PTR, SRV, TXT)
//   2. Encode DNS name per RFC 1035 §3.1 (label-prefixed)
//   3. Build 12-byte header with RFC 6762 §18 flag requirements
//   4. Build question section (QNAME + QTYPE + QCLASS)
//   5. Return combined wire format message
//
// RFC 6762 §18 Query Requirements:
//   §18.2: QR bit MUST be zero (query)
//   §18.3: OPCODE MUST be zero (standard query)
//   §18.4: AA bit MUST be zero
//   §18.6: RD bit SHOULD be zero (M1 enforces MUST)
//
// Parameters:
//   - name: The DNS name to query (e.g., "printer.local")
//   - recordType: The DNS record type (A=1, PTR=12, TXT=16, SRV=33)
//
// Returns:
//   - query: The wire format DNS query message
//   - error: ValidationError if name or recordType is invalid
//
// Example:
//
//	query, err := BuildQuery("printer.local", protocol.RecordTypeA)
//	if err != nil {
//	    return err
//	}
//	// query is now ready to send via UDP multicast
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-020: System MUST set DNS header fields per RFC 6762 §18
// T020: Implement BuildQuery with RFC 6762 §18 compliance
func BuildQuery(name string, recordType uint16) ([]byte, error) {
    ...
}
```

---

## Inline Comments

### When to Write Inline Comments:

1. **WHY Comments**: Explain non-obvious decisions
```go
// Use crypto/rand for ID generation (not math/rand) to prevent
// predictable query IDs which could enable spoofing attacks (G404)
id, err := rand.Int(rand.Reader, big.NewInt(65536))
```

2. **RFC Requirement Comments**: Explain mandatory behavior
```go
// RFC 6762 §8.1 REQUIRES 250ms interval between probes
// This is not configurable per Constitutional Principle I
time.Sleep(250 * time.Millisecond)
```

3. **Bounds Check Comments**: Explain G115 suppressions
```go
// G115: rand.Int bounds upper limit to 65536, so result fits in uint16
id := uint16(idBig.Uint64() % 65536) //nolint:gosec
```

4. **Error Handling Comments**: Explain intentional error swallowing
```go
// Error impossible: regex guarantees suffixStr contains only digits
suffix, _ := strconv.Atoi(suffixStr) // nosemgrep: beacon-error-swallowing
```

5. **Algorithm Comments**: Explain complex logic steps
```go
// RFC 6762 §8.2: Compare records lexicographically
// Step 1: Compare class (excluding cache-flush bit)
// Step 2: Compare type
// Step 3: Compare RDATA bytewise (UNSIGNED comparison)
```

### When NOT to Write Comments:

❌ Don't restate obvious code:
```go
// Bad: i++ increments i
i++
```

❌ Don't document every line:
```go
// Bad: Commented every line
x := 10 // Set x to 10
y := 20 // Set y to 20
z := x + y // Add x and y
```

✅ Do explain WHY:
```go
// Good: Explains the non-obvious reason
// Pre-allocate 512 bytes to avoid reallocation for typical DNS messages
// (most mDNS messages are <512 bytes per RFC 1035 traditional limit)
buffer := make([]byte, 0, 512)
```

---

## Constants Documentation

### Required Elements:
1. **Purpose**: What the constant represents
2. **RFC Section**: Where this value is defined in the RFC
3. **WHY**: Why this specific value (not just "per RFC")
4. **Functional Requirement**: Which FR this constant supports

### Template:

```go
// <Constant> is <description> per RFC <number> §<section>.
//
// WHY: <Explain why this specific value>
//
// FR-XXX: <Functional requirement>
const <Constant> = <value>
```

### Example:

```go
// ProbeInterval is the interval between probe packets - 250 milliseconds per RFC 6762 §8.1.
//
// WHY: This specific timing ensures network convergence while minimizing delay.
// Too short (e.g., 50ms) → insufficient time for responses to arrive
// Too long (e.g., 1s) → unnecessarily slow registration (would take 3s for probing alone)
//
// RFC 6762 §8.1: "When ready to send its Multicast DNS probe packet(s) the host should
// first verify that the hardware address is ready by sending a standard ARP Request for
// the desired IP address and then wait 250 milliseconds."
//
// FR-XXX: System MUST use RFC-mandated timing for probing
// Constitutional Principle I: RFC MUST requirements are not configurable
const ProbeInterval = 250 * time.Millisecond
```

---

## Error Documentation

### Required Elements for Custom Errors:
1. **Purpose**: What error condition this represents
2. **When Returned**: Scenarios that trigger this error
3. **RFC Context**: If applicable, which RFC violation this indicates

### Example:

```go
// ValidationError indicates that input failed RFC-mandated validation.
//
// This error is returned when:
//   - DNS names exceed 255 bytes (RFC 1035 §3.1 violation)
//   - DNS labels exceed 63 bytes (RFC 1035 §3.1 violation)
//   - Record types are unsupported (FR-002: only A, PTR, SRV, TXT)
//   - Service names have invalid format (RFC 6763 §4 violation)
//
// WHY: Validating input upfront prevents malformed packets that would be
// rejected by other mDNS implementations (Avahi, Bonjour).
type ValidationError struct {
    Field   string      // Field that failed validation
    Value   interface{} // Invalid value
    Message string      // Human-readable explanation
}
```

---

## Test Documentation

### Required Elements for Test Functions:
1. **Purpose**: What RFC behavior this test validates
2. **RFC Section**: Specific RFC requirement being tested
3. **Test Scenario**: What scenario is being tested
4. **Expected Outcome**: What result indicates compliance

### Template:

```go
// Test<Functionality>_<Scenario> validates RFC <number> §<section> compliance.
//
// RFC Requirement:
// "<Relevant RFC quote>"
//
// Test Scenario:
// <Description of test setup and actions>
//
// Expected Outcome:
// <What result indicates RFC compliance>
//
// FR-XXX: <Functional requirement being validated>
func Test<Functionality>_<Scenario>(t *testing.T) {
    ...
}
```

### Example:

```go
// TestProber_ConflictDetection_LexicographicTiebreak validates RFC 6762 §8.2 compliance.
//
// RFC 6762 §8.2 states:
// "The two records are compared and the lexicographically later data wins."
//
// Test Scenario:
// Simultaneous probe from two devices for "myservice.local":
//   - Our record: A 192.168.1.50
//   - Their record: A 192.168.1.100
//
// Expected Outcome:
// Conflict detected (192.168.1.100 > 192.168.1.50 lexicographically)
// Machine transitions to ConflictDetected state
// Caller renames service and retries
//
// FR-002 (US2): System MUST detect conflicts via lexicographic comparison
// T059: ConflictDetector integration with Prober
func TestProber_ConflictDetection_LexicographicTiebreak(t *testing.T) {
    ...
}
```

---

## Documentation Review Checklist

Before submitting code for review:

### Package Level:
- [ ] Package has WHY THIS PACKAGE EXISTS section
- [ ] Package lists PRIMARY TECHNICAL AUTHORITY (RFC sections)
- [ ] Package explains DESIGN RATIONALE
- [ ] Package lists all RFC COMPLIANCE requirements

### Types:
- [ ] All exported types have purpose statement
- [ ] Protocol types include wire format diagram
- [ ] Types reference RFC sections where applicable
- [ ] Complex types include usage examples

### Functions:
- [ ] All exported functions have one-line purpose
- [ ] Functions reference RFC sections
- [ ] WHY comment explains RFC requirement and rationale
- [ ] Parameters and returns are documented
- [ ] Complex algorithms have step-by-step explanation
- [ ] Critical behavior includes RFC quotes
- [ ] Public APIs include usage examples
- [ ] Functions reference FR and task IDs

### Constants:
- [ ] All constants reference RFC sections
- [ ] Constants explain WHY this specific value
- [ ] RFC MUST requirements noted as non-configurable

### Inline Comments:
- [ ] WHY comments for non-obvious decisions
- [ ] RFC requirement comments for mandatory behavior
- [ ] Bounds check explanations for G115 suppressions
- [ ] Error handling explanations for intentional swallowing
- [ ] Algorithm comments for complex logic

### Tests:
- [ ] Test names indicate functionality and scenario
- [ ] Tests reference RFC sections being validated
- [ ] Test comments explain RFC requirement
- [ ] Test comments describe scenario and expected outcome

---

## Examples of Good Documentation

See these files for exemplary documentation:

- **Package Level**: `internal/state/machine.go` (comprehensive WHY and RFC mapping)
- **Type Documentation**: `internal/message/message.go` (wire format diagrams)
- **Function Documentation**: `internal/message/builder.go:BuildQuery()` (complete RFC attribution)
- **Constant Documentation**: `internal/protocol/mdns.go` (WHY for each value)
- **Test Documentation**: `tests/contract/` (RFC section validation)

---

## Anti-Patterns to Avoid

### ❌ Restating Code:
```go
// Bad: Just restates the code
// Increment counter
counter++
```

### ❌ Vague RFC References:
```go
// Bad: Which section? What requirement?
// Per RFC 6762
const ProbeInterval = 250 * time.Millisecond
```

### ❌ Missing WHY:
```go
// Bad: Doesn't explain why 250ms
// ProbeInterval is the interval between probes
const ProbeInterval = 250 * time.Millisecond
```

### ✅ Good Documentation:
```go
// Good: Explains RFC section, requirement, and WHY
// ProbeInterval is the interval between probe packets - 250 milliseconds per RFC 6762 §8.1.
//
// WHY: This balances network convergence speed with response arrival time.
// RFC 6762 §8.1 states: "the host should...wait 250 milliseconds"
//
// Constitutional Principle I: RFC MUST requirements are not configurable
const ProbeInterval = 250 * time.Millisecond
```

---

## Documentation Maintenance

### When to Update Documentation:

1. **RFC Reference Changes**: If implementation changes to satisfy different RFC section
2. **Algorithm Changes**: If logic flow changes significantly
3. **New Requirements**: If new FRs are added to specs
4. **Deprecations**: If functionality is deprecated or replaced

### Documentation Debt:

When making quick fixes, mark documentation TODO items:

```go
// TODO DOC: Update RFC reference when implementing full compression support
// Current: Partial RFC 1035 §4.1.4 compliance (no compression)
// Target: Full compression support in M2
```

Track documentation debt in spec.md files and address in dedicated cleanup tasks.

---

## Conclusion

Documentation is not an afterthought—it's a critical component of Beacon's RFC compliance and auditability.

**Key Principles**:
1. Explain **WHY**, not just WHAT
2. Reference **RFC sections explicitly**
3. Tie code to **Functional Requirements**
4. Provide **Examples** for public APIs
5. Show **RFC quotes** for critical behavior

This makes Beacon's implementation transparent, auditable, and maintainable.

When in doubt, ask: "Could someone verify RFC compliance from this documentation alone?"
