# Implementation Plan: Foundation Consolidation & Compliance Tracking

**Branch**: `005-foundation-consolidation` | **Date**: 2025-11-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-foundation-consolidation/spec.md`

**Note**: This is a **documentation-only milestone** - no code changes required. The planning workflow is adapted to focus on documentation structure, content extraction, and consistency validation rather than software implementation.

## Summary

**Primary Requirement**: Consolidate M1/M1-Refactoring/M1.1 achievements into comprehensive, interconnected documentation that provides clear project status visibility for all stakeholders.

**Technical Approach**:
1. Audit existing compliance matrices and specs to extract accurate status
2. Aggregate 59 functional requirements from M1/M1-R/M1.1 spec checklists
3. Update RFC compliance matrix with M1.1 completions (target: 50-60% compliance)
4. Create new functional requirements matrix with full traceability
5. Build compliance dashboard as single entry point to all documentation
6. Write Foundation Completion Report narrating the M1 → M1-R → M1.1 progression
7. Integrate all documentation via ROADMAP.md links
8. Optionally prepare v0.5.0 release (decision deferred to execution)

**Deliverables**: 5 documentation files (1 new dashboard, 1 new FR matrix, 1 new completion report, 2 updated matrices/roadmap)

## Technical Context

**Language/Version**: Markdown (GitHub-flavored, CommonMark spec)
**Primary Dependencies**: None (documentation-only, no external tools)
**Storage**: Git repository (`/home/joshuafuller/development/beacon/`)
**Testing**: Manual validation (link checking, markdown rendering, cross-reference accuracy)
**Target Platform**: GitHub repository (rendered via GitHub markdown engine)
**Project Type**: Documentation consolidation (existing codebase, no new implementation)
**Performance Goals**: N/A (documentation)
**Constraints**:
- Must use relative paths (repo-root relative) for all internal links
- Must maintain consistency across 5 documentation files
- Must preserve existing ROADMAP.md structure (add, don't restructure)
- Compliance percentages must use consistent counting methodology
**Scale/Scope**:
- 59 functional requirements to aggregate (M1: 22, M1-R: 4, M1.1: 33)
- ~80 RFC 6762 sections to audit for M1.1 status updates
- 5 user stories, 32 new FRs, 12 success criteria to validate

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: RFC Compliant
**Status**: ✅ **PASS** (N/A for documentation-only milestone)

**Analysis**: This milestone consolidates documentation of existing RFC-compliant code. No new RFC requirements introduced. Compliance matrices will accurately reflect current implementation status (M1.1 addressed critical RFC gaps like socket options, multicast group membership).

**Action**: None required (documentation milestone)

---

### Principle II: Spec-Driven Development
**Status**: ✅ **PASS**

**Analysis**: This milestone follows spec-driven development:
- ✅ Complete specification in `/specs/005-foundation-consolidation/spec.md`
- ✅ 5 user stories with acceptance scenarios
- ✅ 32 functional requirements with testable outcomes
- ✅ 12 success criteria with measurable outcomes
- ✅ Quality checklist validated (all items PASS)

**Action**: None required (compliant)

---

### Principle III: Test-Driven Development
**Status**: ✅ **PASS** (Adapted for documentation)

**Analysis**: TDD cycle adapted for documentation:
- **RED**: Identify outdated/missing documentation (RFC matrix at ~35%, no FR matrix, no dashboard)
- **GREEN**: Create/update documentation to meet success criteria
- **REFACTOR**: Validate cross-references, fix broken links, ensure consistency

**Test Strategy**:
- Manual validation of all internal links (SC-010)
- Cross-reference accuracy checks (SC-012)
- Markdown rendering validation (SC-009)
- Stakeholder usability test (SC-001: "Can answer questions in <2 minutes")

**Action**: Validation checklist in tasks.md will enforce documentation quality gates

---

### Principle IV: Phased Approach
**Status**: ✅ **PASS**

**Analysis**: This milestone is a natural checkpoint between M1.1 (Foundation Complete) and M2 (Responder):
- Delivers working documentation (compliance dashboard, updated matrices)
- Enables informed M2 planning (clear baseline of what's implemented)
- Provides early stakeholder value (status visibility)
- Validates approach before proceeding (are matrices useful?)

**Phases**:
- Phase 0: Research - Audit existing docs, extract FR counts, validate status
- Phase 1: Design - Define matrix structures, dashboard sections, report outline
- Phase 2: Execution - Create/update all documentation files

**Action**: None required (compliant)

---

### Principle V: Dependencies and Supply Chain
**Status**: ✅ **PASS** (N/A for documentation-only milestone)

**Analysis**: No code dependencies introduced. Documentation-only milestone using:
- Git (already in use)
- Markdown (no external renderer dependencies)
- GitHub (existing platform)

**Action**: None required (documentation milestone)

---

### Principle VI: Open Source
**Status**: ✅ **PASS**

**Analysis**: All documentation follows MIT license. Compliance matrices, FR matrices, and dashboards are project artifacts contributed under project license.

**Action**: None required (compliant)

---

### Principle VII: Maintained
**Status**: ✅ **PASS**

**Analysis**: Documentation created in this milestone becomes **living artifacts**:
- RFC matrix updated each milestone as features are completed
- FR matrix updated as new milestones add requirements
- Compliance dashboard updated to reflect current status
- Foundation report becomes historical artifact (snapshot, not updated)

**Maintenance Plan**:
- Update matrices in each milestone's REFACTOR phase
- Dashboard links validated before each milestone completion
- Compliance percentages recalculated when major features complete

**Action**: Document maintenance responsibilities in plan

---

### Principle VIII: Excellence
**Status**: ✅ **PASS**

**Analysis**: This milestone exemplifies excellence through:
- Comprehensive status tracking (59 FRs, 80+ RFC sections)
- Single source of truth (compliance dashboard)
- Clear stakeholder communication (Foundation report)
- Institutional knowledge capture (why 3 milestones, what's implemented)
- Informed decision-making (clear baseline for M2 planning)

**Action**: None required (exemplifies excellence)

---

### Constitutional Gate Summary

| Principle | Status | Notes |
|-----------|--------|-------|
| I. RFC Compliant | ✅ PASS | N/A (documentation milestone) |
| II. Spec-Driven | ✅ PASS | Complete spec with user stories, FRs, SCs |
| III. TDD | ✅ PASS | Adapted RED/GREEN/REFACTOR for docs |
| IV. Phased | ✅ PASS | Natural checkpoint before M2 |
| V. Dependencies | ✅ PASS | N/A (documentation milestone) |
| VI. Open Source | ✅ PASS | MIT licensed artifacts |
| VII. Maintained | ✅ PASS | Living artifacts, maintenance plan |
| VIII. Excellence | ✅ PASS | Comprehensive, clear, valuable |

**Overall**: ✅ **APPROVED** - All constitutional requirements met. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/005-foundation-consolidation/
├── spec.md              # Feature specification (/speckit.specify output) ✅
├── plan.md              # This file (/speckit.plan output) 🔄 IN PROGRESS
├── research.md          # Phase 0: Documentation audit & FR extraction
├── data-model.md        # Phase 1: Documentation structure definitions
├── quickstart.md        # Phase 1: User guide for using compliance matrices
├── contracts/           # Phase 1: Matrix column schemas, validation rules
│   └── matrix-schema.md # Schema for RFC matrix, FR matrix, dashboard
└── tasks.md             # Phase 2: Execution tasks (/speckit.tasks output)
```

### Source Code (repository root)

**Note**: This milestone creates/updates **documentation files only**, no source code changes.

```text
docs/
├── RFC_COMPLIANCE_MATRIX.md          # ✏️ UPDATED (mark M1.1 sections complete)
├── FUNCTIONAL_REQUIREMENTS_MATRIX.md # ➕ NEW (aggregate 59 FRs from M1/M1-R/M1.1)
├── COMPLIANCE_DASHBOARD.md           # ➕ NEW (single entry point)
├── FOUNDATION_COMPLETE.md            # ➕ NEW (M1 → M1-R → M1.1 narrative)
└── CONTEXT_AND_LOGGING_COMPLIANCE_MATRIX.md # ✅ UNCHANGED (separate audit)

ROADMAP.md                             # ✏️ UPDATED (add dashboard link, Foundation report link)

specs/
├── 002-mdns-querier/
│   └── checklists/requirements.md    # 📖 READ (extract 22 M1 FRs)
├── 003-m1-refactoring/
│   └── checklists/requirements.md    # 📖 READ (extract 4 M1-R FRs) - MAY NOT EXIST, validate
└── 004-m1-1-architectural-hardening/
    └── checklists/requirements.md    # 📖 READ (extract 33 M1.1 FRs)

(Optional - only if v0.5.0 release decided)
RELEASE_NOTES_v0.5.0.md               # ➕ NEW (Foundation summary)
README.md                             # ✏️ UPDATED (installation section for v0.5.0)
```

**Structure Decision**: Single-project repository (existing Beacon codebase). This milestone operates purely in the `/docs/` directory and updates cross-cutting files (`ROADMAP.md`). No changes to `/querier/`, `/internal/`, or `/tests/` directories.

## Complexity Tracking

> **Documentation milestone - no constitutional violations to justify.**

All constitutional principles are satisfied. No complexity concerns for a documentation consolidation phase.

---

## Phase 0: Research & Audit

**Objective**: Extract accurate status data from existing specifications and code to populate compliance matrices.

**Prerequisites**: M1.1 complete, all existing specs and checklists available.

### Research Tasks

#### R001: Audit RFC Compliance Matrix for M1.1 Gaps

**Goal**: Identify which RFC 6762 sections changed status due to M1.1 implementation.

**Approach**:
1. Read `docs/RFC_COMPLIANCE_MATRIX.md` (current baseline)
2. Read `specs/004-m1-1-architectural-hardening/spec.md` (M1.1 features)
3. Cross-reference M1.1 FRs (socket options, interface management, rate limiting, source filtering) to RFC sections
4. List RFC sections to mark as ✅ Implemented (e.g., §11 Source Address Check, §15 Multiple Responders, §21 Security)
5. Estimate new compliance percentage (formula: implemented sections / total sections)

**Output**: List of RFC sections to update, rationale, estimated compliance percentage

---

#### R002: Extract Functional Requirements from M1 Checklist

**Goal**: Aggregate all 22 FRs from M1 Basic Querier.

**Approach**:
1. Read `specs/002-mdns-querier/checklists/requirements.md`
2. Extract FR-001 through FR-022 with descriptions, status
3. Convert to milestone-prefixed IDs (FR-001 → FR-M1-001, preserves traceability to source code)
4. Map each FR to implementation file (e.g., FR-M1-006 → querier/options.go:32)
5. Map each FR to RFC section (if applicable, e.g., FR-M1-001 → RFC 6762 §5.3)
6. Note deferred FRs (if any)

**Output**: Table with columns: FR-ID (milestone-prefixed), Description, Status, Milestone, Implementation File, RFC Reference

---

#### R003: Extract Functional Requirements from M1-Refactoring

**Goal**: Aggregate FRs from M1-Refactoring (if checklist exists).

**Approach**:
1. Check if `specs/003-m1-refactoring/checklists/requirements.md` exists
2. If exists, extract FRs (expected: ~4 FRs related to transport interface, buffer pooling, error propagation)
3. If not exists, infer FRs from `archive/m1-refactoring/reports/REFACTORING_COMPLETE.md`
4. Convert to milestone-prefixed IDs (FR-001 → FR-M1R-001, "R" for Refactoring)
5. Map FRs to ADRs (e.g., FR-M1R-001 Transport → docs/decisions/001-transport-interface-abstraction.md)

**Output**: Table of M1-R FRs with milestone-prefixed IDs (or note if checklist doesn't exist, FRs inferred from completion report)

---

#### R004: Extract Functional Requirements from M1.1 Checklist

**Goal**: Aggregate all 33 FRs from M1.1 Architectural Hardening.

**Approach**:
1. Read `specs/004-m1-1-architectural-hardening/checklists/requirements.md`
2. Extract FR-001 through FR-033 (socket options, interface management, rate limiting, source filtering)
3. Convert to milestone-prefixed IDs (FR-001 → FR-M1.1-001, preserves traceability)
4. Note platform-specific FRs (e.g., FR-M1.1-018 socket_linux.go)
5. Note deferred FRs (e.g., logging FRs deferred to F-6 spec)
6. Map to RFC sections (e.g., FR-M1.1-023 source filtering → RFC 6762 §11)

**Output**: Table of M1.1 FRs with milestone-prefixed IDs and platform status (Linux ✅, macOS/Windows ⚠️)

---

#### R005: Documentation Structure Research

**Goal**: Define optimal structure for compliance dashboard and matrices.

**Approach**:
1. Review industry best practices for compliance matrices (NIST, ISO, OWASP)
2. Analyze existing matrix structure (RFC_COMPLIANCE_MATRIX.md) for strengths/weaknesses
3. Determine optimal columns for FR matrix (see FR-007 in spec.md)
4. Design dashboard sections (Quick Status, What Works, Limitations, Contribute)
5. Define linking strategy (relative paths, anchor links, external links)

**Output**: Documentation structure design with section definitions, column schemas, linking patterns

---

#### R006: Compliance Percentage Calculation Methodology

**Goal**: Define consistent methodology for calculating RFC compliance percentage.

**Approach**:
1. Review RFC 6762 structure (20 numbered sections, many with subsections)
2. Decide granularity: count top-level sections, or subsections (§8.1, §8.2), or individual MUSTs?
3. Document rationale for chosen methodology (see Assumption #1 in spec.md)
4. Create formula: `compliance % = (implemented sections / total sections) * 100`
5. Apply formula to current state to validate (should be ~35% pre-M1.1, ~50-60% post-M1.1)

**Output**: Compliance calculation methodology documented, applied to current state for validation

---

**Research Output**: `research.md` with all findings consolidated, ready for Phase 1 design.

---

## Phase 1: Design & Contracts

**Objective**: Define structure, schemas, and content organization for all documentation artifacts.

**Prerequisites**: Phase 0 research complete (all FRs extracted, RFC sections identified, methodology defined).

### Design Artifacts

#### D001: Data Model (data-model.md)

**Purpose**: Define entities and relationships for compliance tracking.

**Entities**:
- **RFC Requirement**: Section number, requirement text, status, evidence links, platform notes
- **Functional Requirement**: FR-ID (milestone-prefixed: FR-M1-XXX, FR-M1R-XXX, FR-M1.1-XXX), description, status, milestone, implementation files, RFC references, test evidence
- **Success Criterion**: SC-ID, description, status, validation method, evidence
- **Milestone**: ID, name, status, completion date, FR count, task count
- **Compliance Status**: Aggregate metrics (RFC %, total FRs, implemented FRs)

**Relationships**:
- RFC Requirement ← (implements) → Functional Requirement (many-to-many)
- Functional Requirement → Milestone (many-to-one)
- Success Criterion → Functional Requirement (many-to-many)

**Output**: Entity definitions with attributes and relationships documented in `data-model.md`

---

#### D002: Matrix Schemas (contracts/matrix-schema.md)

**Purpose**: Define column structure and validation rules for each matrix.

**RFC Compliance Matrix Schema**:
```markdown
| Section | Requirement | Status | Implementation Evidence | Platform Notes |
|---------|-------------|--------|------------------------|----------------|
| RFC 6762 §X.Y | Requirement text | ✅/❌/⚠️/🔄/📋 | Link to code/spec | Linux ✅, macOS ⚠️ |
```

**Functional Requirements Matrix Schema**:
```markdown
| FR-ID | Description | Status | Milestone | Implementation File(s) | RFC Reference(s) | Test Evidence |
|-------|-------------|--------|-----------|----------------------|-----------------|---------------|
| FR-M1-001 | Query transmission | Implemented | M1 | querier/querier.go:45 | RFC 6762 §5.1 | tests/integration/TestQuery |
| FR-M1R-001 | Transport interface abstraction | Implemented | M1-R | transport/transport.go:12 | - | tests/transport/TestUDPv4Transport |
| FR-M1.1-015 | Rate limiting (multicast storm protection) | Implemented | M1.1 | security/rate_limiter.go:23 | RFC 6762 §6 | tests/security/TestRateLimiter |
```

**Note**: Milestone-prefixed FR-IDs preserve traceability to source code comments, git commits, and original spec checklists without renumbering risk.

**Validation Rules**:
- Status values must be from approved set (✅ Implemented, ❌ Not Implemented, ⚠️ Partial, 🔄 In Progress, 📋 Planned)
- All file paths must be relative from repo root
- All RFC references must use §N.M format
- Platform notes required for socket-related items

**Output**: Schema definitions and validation rules in `contracts/matrix-schema.md`

---

#### D003: Compliance Dashboard Structure (contracts/dashboard-sections.md)

**Purpose**: Define sections and content organization for COMPLIANCE_DASHBOARD.md.

**Dashboard Sections**:

1. **Quick Status** (US1 acceptance criterion)
   - Foundation phase summary (M1 ✅, M1-R ✅, M1.1 ✅)
   - Current RFC compliance percentage
   - Next milestone (M2)
   - Last updated date

2. **What Works Today** (US1 acceptance criterion, FR-015)
   - Query-only mDNS (send queries, receive responses)
   - Record types: A, PTR, SRV, TXT
   - Platform-specific socket configuration (Linux)
   - VPN/Docker interface exclusion
   - Rate limiting (multicast storm protection)
   - Source IP filtering (link-local validation)
   - Avahi/Bonjour coexistence (port 5353 sharing)

3. **Known Limitations** (FR-016)
   - No responder/service registration (M2)
   - No IPv6 (M2)
   - Platform validation: Linux only (macOS/Windows code-complete, untested)
   - No logging infrastructure (F-6 spec exists, not implemented)

4. **Navigation** (FR-014)
   - Links to: RFC matrix, FR matrix, ROADMAP, Foundation report, Constitution, F-series specs

5. **How to Contribute** (FR-017)
   - Link to open issues
   - Link to M2 planning (when available)
   - Link to spec kit workflow

**Output**: Section definitions with content guidelines in `contracts/dashboard-sections.md`

---

#### D004: Foundation Completion Report Outline (contracts/foundation-report-outline.md)

**Purpose**: Define structure for FOUNDATION_COMPLETE.md narrative.

**Report Sections** (per FR-018-022):

1. **Executive Summary**
   - Foundation Complete: M1 + M1-R + M1.1
   - 210+ tasks complete across 3 milestones
   - 80% test coverage, zero regressions, zero race conditions

2. **Why Three Milestones?** (FR-019)
   - M1: Basic (query-only, functional)
   - M1-Refactoring: Refactored (clean architecture, buffer pooling, error propagation)
   - M1.1: Production-Ready (socket config, security, interface management)
   - Why not one big milestone? (progressive refinement, validated approach)

3. **What's Implemented** (FR-021 - by functional area)
   - **Querying**: Multicast query transmission, response parsing, deduplication, timeout handling
   - **Socket Configuration**: SO_REUSEPORT, multicast group membership, platform-specific options
   - **Interface Management**: VPN/Docker exclusion, explicit interface selection, filtering API
   - **Security**: Rate limiting, source IP filtering, packet size validation, malformed packet handling
   - **Testing**: Contract tests, integration tests, fuzz tests, race detection, 80% coverage

4. **Quality Metrics** (FR-020)
   - 210+ tasks complete
   - 80.0% test coverage
   - 10/10 packages PASS
   - Zero race conditions
   - Zero regressions (all M1 tests pass)
   - Fuzz tested (114K executions, zero crashes)

5. **What's Next** (FR-022)
   - M2: mDNS Responder (service registration, probing, announcing)
   - M3: DNS-SD Core (service discovery protocol)
   - M4-M6: Advanced features, platform expansion, enterprise readiness

**Output**: Report outline with content guidelines in `contracts/foundation-report-outline.md`

---

#### D005: Quickstart Guide (quickstart.md)

**Purpose**: User guide for navigating and using compliance matrices.

**Content**:
1. **How to Use the Compliance Dashboard**
   - What is it? (single entry point to project status)
   - Who should use it? (stakeholders, contributors, users)
   - How to answer common questions ("Does Beacon support X?", "What's implemented?")

2. **How to Read the RFC Compliance Matrix**
   - What do status icons mean? (✅ ❌ ⚠️ 🔄 📋)
   - How to find if a specific RFC section is implemented
   - How to interpret platform notes (Linux ✅, macOS ⚠️)

3. **How to Read the FR Matrix**
   - What is an FR? (testable requirement)
   - How to trace FR to implementation (file paths column)
   - How to trace FR to RFC section (bidirectional links)
   - How to filter by milestone (M1/M1-R/M1.1)

4. **How to Update Matrices** (for maintainers)
   - When to update (each milestone completion)
   - How to mark FRs complete (change status, add evidence links)
   - How to recalculate compliance percentage
   - How to validate cross-references

**Output**: User-friendly guide in `quickstart.md`

---

**Phase 1 Output**:
- `data-model.md` (entity definitions)
- `contracts/matrix-schema.md` (column schemas, validation rules)
- `contracts/dashboard-sections.md` (dashboard structure)
- `contracts/foundation-report-outline.md` (report structure)
- `quickstart.md` (user guide)

---

## Phase 2: Execution (Tasks)

**Note**: Phase 2 is NOT executed by `/speckit.plan`. It will be generated by `/speckit.tasks` command.

**Approach**: Tasks will be organized by user story (US1-US5) to enable independent testing and delivery.

**Expected Task Structure**:
- **Phase 1: US1 - Compliance Dashboard** (create COMPLIANCE_DASHBOARD.md)
- **Phase 2: US2 - RFC Matrix Update** (update docs/RFC_COMPLIANCE_MATRIX.md)
- **Phase 3: US3 - FR Matrix Creation** (create docs/FUNCTIONAL_REQUIREMENTS_MATRIX.md)
- **Phase 4: US4 - Foundation Report** (create docs/FOUNDATION_COMPLETE.md)
- **Phase 5: US5 - Optional v0.5.0** (decision point, then release prep if approved)
- **Phase 6: Integration** (update ROADMAP.md links, validate cross-references)
- **Phase 7: Validation** (all success criteria, quality gates)

**User Story Independence**: Each phase delivers independently testable value (compliance dashboard works without FR matrix, FR matrix works without Foundation report, etc.).

---

## Validation Strategy

**Documentation Quality Gates** (per SC-009-012):

1. **Markdown Rendering** (SC-009)
   - Preview all files in GitHub markdown renderer
   - Validate tables render correctly (no broken pipes)
   - Validate code blocks use correct syntax highlighting
   - Validate lists and headers render as expected

2. **Link Validation** (SC-010)
   - Test all internal links (relative paths resolve)
   - Test all anchor links (section headers exist)
   - Test all external links (RFC URLs, Go docs)
   - Use link checker tool (or manual validation)

3. **Compliance Percentage** (SC-011)
   - Document calculation methodology in RFC matrix header
   - Apply methodology consistently
   - Show calculation (e.g., "45 implemented / 80 total sections = 56.25%")

4. **Cross-Reference Accuracy** (SC-012)
   - RFC matrix → FR matrix: Verify "Implemented via FR-XXX" links work
   - FR matrix → Implementation: Verify file paths are correct
   - FR matrix → Tests: Verify test evidence links work
   - Dashboard → All docs: Verify all navigation links work

**Stakeholder Validation** (per SC-001):
- Have a non-contributor (or fresh contributor) use dashboard to answer "What does Beacon support?"
- Time how long it takes (target: <2 minutes per SC-001)
- Collect feedback on clarity, usability

---

## Maintenance Plan

**Living Artifacts** (updated each milestone):
- `docs/RFC_COMPLIANCE_MATRIX.md` - Update when RFC sections are implemented
- `docs/FUNCTIONAL_REQUIREMENTS_MATRIX.md` - Append new FRs as milestones add requirements
- `docs/COMPLIANCE_DASHBOARD.md` - Update Quick Status, What Works, Known Limitations

**Snapshot Artifacts** (historical record, not updated):
- `docs/FOUNDATION_COMPLETE.md` - Snapshot of M1.X Foundation phase

**Update Triggers**:
- **Milestone Completion**: Update matrices, dashboard, ROADMAP
- **Major Feature**: Update RFC matrix if new RFC sections implemented
- **Deferred Feature**: Update FR matrix status ("Deferred to MX")

**Validation Cadence**:
- Before each milestone merge: Validate all links, cross-references
- Quarterly: Review compliance percentages for accuracy
- Annually: Full audit of matrix consistency

---

## Success Metrics

From spec.md success criteria, we will measure:

| Metric | Target | Validation Method |
|--------|--------|------------------|
| SC-001: Status answer time | <2 minutes | Stakeholder usability test |
| SC-002: RFC compliance % | 50-60% | Calculation validation |
| SC-003: FR tracking | 59 FRs | Count validation |
| SC-004: Dashboard links | All working | Link checker |
| SC-005: Foundation narrative | Clear explanation | Non-contributor review |
| SC-006: Documentation consistency | 100% match | Cross-reference audit |
| SC-007: Platform status clarity | Clear markers | Visual inspection |
| SC-009: Markdown rendering | No errors | GitHub preview |
| SC-010: Link resolution | 100% resolve | Link checker |
| SC-011: Calculation methodology | Documented | RFC matrix header |
| SC-012: Cross-reference accuracy | 100% accurate | Bidirectional validation |

---

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| FR count mismatch (M1/M1-R/M1.1 checklists incomplete) | Medium | Low | Audit code if checklist missing, infer FRs from specs |
| RFC compliance % calculation dispute | Low | Medium | Document methodology clearly, show calculation |
| Dashboard becomes maintenance burden | Medium | Medium | Keep structure simple, automate validation if possible |
| Links break when files move | High | Low | Use relative paths, validate before each merge |
| v0.5.0 release decision unclear | Low | High | Defer decision, make optional FRs clearly optional |

---

## Next Steps

1. **Review this plan** - Validate approach, structure, success metrics
2. **Run `/speckit.tasks`** - Generate execution tasks with TDD checkpoints
3. **Execute Phase 0** - Research & audit (extract FRs, identify RFC updates)
4. **Execute Phase 1** - Design (data model, schemas, outlines)
5. **Execute Phase 2** - Create/update all documentation files
6. **Validate** - All success criteria, quality gates
7. **Merge** - 005-foundation-consolidation → master
8. **Optional** - Tag v0.5.0 if decided

**Estimated Total Effort**: 6 hours (as spec.md notes)

---

## References

- [Feature Specification](spec.md)
- [Beacon Constitution v1.1.0](../../.specify/memory/constitution.md)
- [ROADMAP.md](../../ROADMAP.md)
- [RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)
- [M1 Completion Summary](../002-mdns-querier/tasks.md)
- [M1.1 INCOMPLETE_TASKS_ANALYSIS](../004-m1-1-architectural-hardening/INCOMPLETE_TASKS_ANALYSIS.md)
