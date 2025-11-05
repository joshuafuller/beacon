# Quickstart Guide: Writing Beacon Feature Specifications

**Purpose**: Help specification writers create Beacon feature specs that properly reference the architectural foundation and comply with constitutional principles.

**Audience**: Specification writers, developers creating feature specifications, contributors

**Last Updated**: 2025-11-01

---

## 📚 Documentation Hierarchy

Before writing a feature specification, understand the Beacon documentation hierarchy:

### 1. RFC 6762 & RFC 6763 - **PRIMARY TECHNICAL AUTHORITY** ⭐

**Location**: `/RFC%20Docs/`

- **[RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)** (184 KB, 1,410 lines)
  - Authoritative specification for mDNS protocol
  - Covers: Message format, queries, responses, probing, announcing, caching, conflict resolution, timing

- **[RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt)** (125 KB, 969 lines)
  - Authoritative specification for DNS-SD
  - Covers: Service naming, browsing, resolution, TXT records, service types, subtypes

**When to Use**: Any feature that involves protocol behavior MUST reference specific RFC sections. These RFCs are the definitive sources of truth.

**Critical**: Constitution Principle I states **"RFC requirements override all other concerns"**. When RFCs conflict with any other documentation, RFCs take precedence.

### 2. Beacon Constitution v1.0.0 - **PROJECT GOVERNANCE**

**Location**: [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

- Defines 7 non-negotiable principles (I-VII)
- Amendment process and compliance enforcement
- Supersedes all documentation **except RFCs**

**When to Use**: Every feature spec MUST include a Constitutional Alignment section demonstrating compliance with relevant principles.

### 3. BEACON_FOUNDATIONS v1.1 - **COMMON FOUNDATIONAL KNOWLEDGE**

**Location**: [`.specify/specs/BEACON_FOUNDATIONS.md`](../../.specify/specs/BEACON_FOUNDATIONS.md)

- Common knowledge base for all users, developers, AI agents, and contributors
- Extracts and explains concepts from RFCs 6762 and 6763
- Provides: DNS fundamentals, mDNS essentials, DNS-SD concepts, terminology glossary (§5), reference tables (§7)

**When to Use**:
- To understand RFC concepts in accessible format
- To find consistent terminology (use BEACON_FOUNDATIONS glossary terms)
- To reference timing values, default configuration, common requirements

**Important**: BEACON_FOUNDATIONS **does NOT replace RFCs** - it provides accessible explanations. For authoritative protocol requirements, always consult the RFCs.

### 4. F-Series Architecture Specifications - **IMPLEMENTATION PATTERNS**

**Location**: [`.specify/specs/`](../../.specify/specs/)

- **[F-2: Package Structure & Layering](../../.specify/specs/F-2-package-structure.md)** - Package organization and import rules
- **[F-3: Error Handling Strategy](../../.specify/specs/F-3-error-handling.md)** - 8 error categories and RFC-specific patterns
- **[F-4: Concurrency Model](../../.specify/specs/F-4-concurrency-model.md)** - Goroutine patterns and RFC-compliant timing
- **[F-5: Configuration & Defaults](../../.specify/specs/F-5-configuration.md)** - RFC MUST vs configurable separation
- **[F-6: Logging & Observability](../../.specify/specs/F-6-logging-observability.md)** - Hot path definition and TXT redaction
- **[F-7: Resource Management](../../.specify/specs/F-7-resource-management.md)** - Cleanup patterns and leak prevention
- **[F-8: Testing Strategy](../../.specify/specs/F-8-testing-strategy.md)** - RFC traceability matrix and TDD requirements

**When to Use**: Reference F-series specs when your feature needs to follow architectural patterns for error handling, concurrency, configuration, logging, resource management, or testing.

**Important**: F-series specs must align with RFCs. If conflict exists, RFCs take precedence per Constitution Principle I.

### 5. Feature Specifications - **SPECIFIC FEATURES**

**Location**: `specs/###-feature-name/spec.md`

- Individual feature requirements and user scenarios
- Must reference RFCs for protocol behavior
- Must comply with Constitution and F-series patterns

---

## ✍️ Writing a Feature Specification

### Step 1: Use `/speckit.specify` Command

```bash
/speckit.specify <your feature description>
```

This command:
1. Creates a new feature branch (`###-feature-name`)
2. Initializes `specs/###-feature-name/spec.md` from template
3. Generates user scenarios, requirements, and success criteria
4. Validates specification quality

### Step 2: Add References Section

Every feature spec MUST include a References section organized by the documentation hierarchy:

```markdown
## References

### Technical Sources of Truth (RFCs)

**PRIMARY AUTHORITY for all protocol behavior:**

- **[RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)** - Authoritative mDNS specification
  - Referenced sections: [List specific sections, e.g., §5, §8.1, §10]
- **[RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt)** - Authoritative DNS-SD specification
  - Referenced sections: [List specific sections, e.g., §4, §6, §7]

**Critical Note**: Constitution Principle I states "RFC requirements override all other concerns".

### Project Governance

- **[Beacon Constitution v1.0.0](../../.specify/memory/constitution.md)** - Project governance

### Foundational Knowledge

- **[BEACON_FOUNDATIONS v1.1](../../.specify/specs/BEACON_FOUNDATIONS.md)** - Common knowledge for all contributors

### Architecture Specifications (F-Series)

- [F-2: Package Structure](../../.specify/specs/F-2-package-structure.md) - [if applicable]
- [F-3: Error Handling](../../.specify/specs/F-3-error-handling.md) - [if applicable]
- [F-4: Concurrency Model](../../.specify/specs/F-4-concurrency-model.md) - [if applicable]
- [F-5: Configuration](../../.specify/specs/F-5-configuration.md) - [if applicable]
- [F-6: Logging](../../.specify/specs/F-6-logging-observability.md) - [if applicable]
- [F-7: Resource Management](../../.specify/specs/F-7-resource-management.md) - [if applicable]
- [F-8: Testing Strategy](../../.specify/specs/F-8-testing-strategy.md) - [if applicable]
```

**Tip**: Only include F-series specs that are relevant to your feature. Remove unused references.

### Step 3: Cite RFCs in User Scenarios

When writing acceptance scenarios that involve protocol behavior, cite specific RFC sections:

```markdown
**Acceptance Scenarios**:

1. **Given** a user wants to query for a service, **When** they send an mDNS query to `_http._tcp.local.`, **Then** the query follows RFC 6762 §5 format with QR=0 and OPCODE=0
2. **Given** a responder receives a query, **When** it needs to announce ownership, **Then** it sends 2 announcements 1 second apart per RFC 6762 §8.3
3. **Given** a service instance name is created, **When** the name contains spaces, **Then** spaces are escaped as `\032` per RFC 6763 §4.3
```

**Format**: `RFC #### §X.Y` where #### is the RFC number and §X.Y is the section

**Examples**:
- `RFC 6762 §5` - Queries
- `RFC 6762 §8.1` - Probing (3 probes, 250ms apart)
- `RFC 6762 §8.3` - Announcing (2+ announcements, 1s apart)
- `RFC 6762 §10` - Resource record TTL values
- `RFC 6763 §4` - Service instance naming
- `RFC 6763 §6` - TXT record format
- `RFC 6763 §7` - Service instance enumeration (PTR queries)

### Step 4: Add Constitutional Alignment Section

Every feature spec MUST include a Constitutional Alignment section that demonstrates compliance with all 7 principles:

```markdown
## Constitutional Alignment

This specification demonstrates alignment with Beacon Constitution v1.0.0:

### Principle I: RFC Compliant
- ✅ **[Evidence 1]**: [How this feature enforces or complies with RFC requirements]
- ✅ **[Evidence 2]**: [Specific FR/SC that reference RFC sections]
- ✅ **[Evidence 3]**: [Validation approach for RFC compliance]

### Principle II: Spec-Driven Development
- ✅ **This Specification Exists**: [Explain that spec comes before implementation]
- ✅ **[Evidence]**: [How spec enables deliberate design]

### Principle III: Test-Driven Development
- ✅ **Testable Acceptance Scenarios**: [Reference user scenarios section]
- ✅ **[Evidence]**: [How acceptance tests define RED phase]
- ✅ **Independent Testability**: [How each user story can be tested independently]

### Principle IV: Phased Approach
- ✅ **Priority-Based**: [Explain P1, P2, P3 user stories]
- ✅ **MVP Viability**: [How each priority delivers standalone value]
- ✅ **[Evidence]**: [Milestone mapping if applicable]

### Principle V: Open Source
- ✅ **Public Documentation**: [Confirm spec is publicly available]
- ✅ **[Evidence]**: [MIT License, transparent process]

### Principle VI: Maintained
- ✅ **Versioning**: [Version control approach]
- ✅ **[Evidence]**: [Backward compatibility considerations]

### Principle VII: Excellence
- ✅ **Best Practices**: [How feature follows industry standards]
- ✅ **Quality Enforcement**: [FR/SC that enforce quality]
- ✅ **[Evidence]**: [Specific excellence measures]
```

**Tip**: Not all principles apply equally to every feature. Focus on the most relevant principles and provide specific evidence (FR/SC numbers, RFC sections, etc.).

### Step 5: Use BEACON_FOUNDATIONS Terminology

Use consistent terminology from [BEACON_FOUNDATIONS §5 (Terminology Glossary)](../../.specify/specs/BEACON_FOUNDATIONS.md#5-terminology-glossary):

**Common Terms**:
- **Querier**: Component that sends queries (not "client" or "requester")
- **Responder**: Component that answers queries (not "server" or "answerer")
- **Probe**: Verify name availability before claiming (RFC 6762 §8.1)
- **Announce**: Declare ownership of a name (RFC 6762 §8.3)
- **Cache-Flush**: Signal to replace cached records (top bit of RRCLASS)
- **QU Bit**: Unicast-response preference (top bit of QCLASS)
- **Goodbye Packet**: Record with TTL=0 indicating removal
- **One-Shot Query**: Single query from ephemeral port
- **Continuous Query**: Ongoing monitoring using source port 5353
- **Link-Local**: Scope limited to single network segment

**Tip**: When in doubt, search BEACON_FOUNDATIONS §5 for the correct term.

---

## 📋 Example: Minimal Feature Spec Snippet

Here's a minimal example showing proper references and alignment:

```markdown
# Feature Specification: mDNS Query Sender

## User Scenarios & Testing

### User Story 1 - Send Basic Query (Priority: P1)

As a **network administrator**, I need to query for services on the local network using mDNS, so that I can discover available services without requiring a DNS server.

**Acceptance Scenarios**:

1. **Given** a service type `_http._tcp.local.`, **When** I send an mDNS query, **Then** the query packet follows RFC 6762 §5 format with QR=0, OPCODE=0, and is sent to 224.0.0.251:5353
2. **Given** a query is sent, **When** responses are received, **Then** I can parse PTR records per RFC 6763 §7

## Requirements

### Functional Requirements

- **FR-001**: Querier MUST send mDNS queries to multicast address 224.0.0.251:5353 per RFC 6762 §5
- **FR-002**: Querier MUST construct DNS message format per RFC 6762 §18
- **FR-003**: Querier MUST implement Known-Answer Suppression per RFC 6762 §7.1

## References

### Technical Sources of Truth (RFCs)

- **[RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)** - §5 (Queries), §7.1 (Known-Answer), §18 (Message Format)
- **[RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt)** - §7 (Service Enumeration)

### Project Governance

- **[Beacon Constitution v1.0.0](../../.specify/memory/constitution.md)**

### Foundational Knowledge

- **[BEACON_FOUNDATIONS v1.1](../../.specify/specs/BEACON_FOUNDATIONS.md)** - Terminology (§5), Reference Tables (§7)

### Architecture Specifications

- [F-2: Package Structure](../../.specify/specs/F-2-package-structure.md) - Use `beacon/querier` package
- [F-4: Concurrency Model](../../.specify/specs/F-4-concurrency-model.md) - Query timeout patterns
- [F-8: Testing Strategy](../../.specify/specs/F-8-testing-strategy.md) - RFC compliance testing

## Constitutional Alignment

### Principle I: RFC Compliant
- ✅ **RFC References**: FR-001, FR-002, FR-003 all cite specific RFC sections
- ✅ **Validation**: Acceptance scenarios validate RFC 6762 §5 query format
- ✅ **No Deviations**: All MUST requirements from RFC 6762 are implemented

### Principle II: Spec-Driven Development
- ✅ **Spec Exists**: This specification defines the feature before implementation

### Principle III: Test-Driven Development
- ✅ **Testable Scenarios**: 2 acceptance scenarios define test cases
- ✅ **RED Phase**: Tests verify RFC 6762 §5 compliance before implementation
```

---

## ✅ Validation Checklist

Before submitting your specification, verify:

- [ ] **References Section**: Organized by hierarchy (RFCs → Constitution → BEACON_FOUNDATIONS → F-series)
- [ ] **RFC Citations**: All protocol behaviors cite specific RFC sections (e.g., "RFC 6762 §8.1")
- [ ] **Constitutional Alignment**: Addresses all 7 principles with specific evidence
- [ ] **BEACON_FOUNDATIONS Terminology**: Uses consistent terms from §5 glossary
- [ ] **Success Criteria**: Technology-agnostic and measurable (no Go packages, implementation details)
- [ ] **Acceptance Scenarios**: Written in Given/When/Then format
- [ ] **F-Series References**: Only includes relevant architecture specs (remove unused)
- [ ] **RFC Authority**: Documents that RFCs are PRIMARY TECHNICAL AUTHORITY

---

## 🚀 Next Steps

After writing your specification:

1. **Validate with `/speckit.clarify`** (optional) - Ask clarifying questions if needed
2. **Generate plan with `/speckit.plan`** - Create implementation plan with Constitutional Check
3. **Generate tasks with `/speckit.tasks`** - Break down implementation into TDD tasks
4. **Implement with `/speckit.implement`** - Execute tasks following RED → GREEN → REFACTOR

---

## 📚 Additional Resources

- [Spec Kit Documentation](https://docs.claude.com/en/docs/claude-code/speckit) - Full Spec Kit workflow
- [Beacon Constitution v1.0.0](../../.specify/memory/constitution.md) - Governance and principles
- [BEACON_FOUNDATIONS v1.1](../../.specify/specs/BEACON_FOUNDATIONS.md) - Common knowledge and terminology
- [ROADMAP](../../ROADMAP.md) - Strategic context and Phase 0 completion
- [RFC 6762 (Full Text)](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - Multicast DNS specification
- [RFC 6763 (Full Text)](../../RFC%20Docs/RFC-6763-DNS-SD.txt) - DNS-SD specification

---

**Questions?** Consult BEACON_FOUNDATIONS or ask in project discussions. The documentation hierarchy is designed so you can find answers without needing to ask maintainers.

**Remember**: RFCs are the ultimate technical authority. When in doubt, consult RFC 6762 or RFC 6763 directly.
