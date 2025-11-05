# Documentation Structure Research (T008 - R005)

**Task**: Review existing `docs/` directory, identify naming conventions and cross-reference patterns
**Date**: 2025-11-02

---

## Current Documentation Structure

### Existing Files in `docs/`

1. **RFC_COMPLIANCE_MATRIX.md** (Updated 2025-11-01, M1 status)
   - Format: Section-by-section matrix with Status column
   - Legend: ✅ ❌ ⚠️ 🔄 📋 status icons
   - Structure: Organized by RFC (6762, 6763), then by section
   - Last Updated tracking: "Last Updated: YYYY-MM-DD"

2. **CONTEXT_AND_LOGGING_COMPLIANCE_MATRIX.md** (Exists, older)
   - Similar format to RFC matrix
   - May be outdated or superseded

3. **decisions/** (Architecture Decision Records)
   - ADR-001: Transport Interface Abstraction
   - ADR-002: Buffer Pooling Pattern
   - ADR-003: Integration Test Timing Tolerance
   - Format: Standard ADR format (Context, Decision, Consequences)

---

## Naming Conventions

### File Naming
- **Matrices**: `[TOPIC]_COMPLIANCE_MATRIX.md` (UPPERCASE, underscores)
- **ADRs**: `###-kebab-case-title.md` (numbered, kebab-case)
- **Reports**: `[TOPIC]_COMPLETE.md` or `[NAME]_REPORT.md` (UPPERCASE, underscores)
- **Dashboards**: `[NAME]_DASHBOARD.md` (UPPERCASE, underscores)

### Section Naming
- Headers use Title Case
- Status columns use emoji icons consistently (✅ ❌ ⚠️ 🔄 📋)
- Platform notes use format: "Linux ✅, macOS ⚠️, Windows ⚠️"

---

## Cross-Reference Patterns

### Internal Links (Relative Paths)
```markdown
Link Text: ../path/to/file.md
Link to Section: ./file.md#section-name
```

### RFC References
```markdown
RFC 6762 §5.3       (section symbol §)
RFC 1035 §4.1.4     (full RFC number + section)
```

### File Path References
```markdown
internal/message/builder.go:42    (file:line)
querier/querier.go               (file only)
```

### ADR References
```markdown
ADR-001 (short form)
docs/decisions/001-transport-interface-abstraction.md (full path)
```

---

## Matrix Schema Analysis

### RFC Compliance Matrix Columns
| Section | Requirement | Status | Notes |
|---------|-------------|--------|-------|
| Text | Text | Icon | Text |

### Proposed FR Matrix Columns (per spec.md FR-007)
| FR-ID | Description | Status | Milestone | Implementation File(s) | RFC Reference(s) | Test Evidence |
|-------|-------------|--------|-----------|----------------------|-----------------|---------------|
| FR-M1-001 | Text | Text | M1/M1-R/M1.1 | Path | RFC X §Y | Test path |

---

## Link Validation Patterns

### Internal Links
- Use relative paths from repository root
- Example: `../RFC%20Docs/RFC-6762-Multicast-DNS.txt` (note: space encoding)
- Anchor links: `#section-name` (lowercase, hyphens replace spaces)

### External Links
- RFC URLs: `https://www.rfc-editor.org/rfc/rfc6762.html`
- GitHub URLs: Full URLs for cross-repo links

---

## Recommendations for New Documentation

### COMPLIANCE_DASHBOARD.md
- **Location**: `docs/COMPLIANCE_DASHBOARD.md`
- **Format**: Single-page overview with 5 sections (per spec)
- **Links**: Use relative paths to matrices, ROADMAP, Constitution
- **Status Icons**: Consistent with RFC matrix (✅ ❌ ⚠️)

### FUNCTIONAL_REQUIREMENTS_MATRIX.md
- **Location**: `docs/FUNCTIONAL_REQUIREMENTS_MATRIX.md`
- **Format**: Table with 7 columns (per FR-007)
- **Grouping**: Organize by milestone section (M1, M1-R, M1.1)
- **IDs**: Milestone-prefixed (FR-M1-001, FR-M1R-001, FR-M1.1-001)

### FOUNDATION_COMPLETE.md
- **Location**: `docs/FOUNDATION_COMPLETE.md`
- **Format**: Narrative report with 5 sections (per spec)
- **Style**: Similar to `archive/m1-refactoring/reports/REFACTORING_COMPLETE.md`
- **Metrics**: Include quantifiable outcomes (210+ tasks, 80% coverage)

---

## Quality Standards

### All Documentation MUST Have:
1. **Last Updated** date at top or bottom
2. **Status Icons** (if applicable): ✅ ❌ ⚠️ 🔄 📋
3. **Relative Paths** for internal links
4. **Markdown Validation**: No broken tables, lists, headers
5. **Link Validation**: All links resolve correctly

### Tables MUST:
- Use proper markdown pipe syntax (`|`)
- Have header separator row (`|---|---|`)
- Align columns consistently

---

**Generated**: 2025-11-02
**Next**: Use this research in design tasks (T010-T014) to define matrix schemas and dashboard structure
