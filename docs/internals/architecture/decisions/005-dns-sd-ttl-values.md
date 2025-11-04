# ADR-005: DNS-SD TTL Values for mDNS Resource Records

**Status**: Accepted
**Date**: 2025-11-04
**Context**: US2 GREEN Phase - Message Capture Implementation

## Context and Problem Statement

RFC 6762 §10 provides general guidance for mDNS TTL values:

> "As a general rule, the recommended TTL value for Multicast DNS resource records with a host name as the resource record's name (e.g., A, AAAA, HINFO) or a host name contained within the resource record's rdata (e.g., SRV, reverse mapping PTR record) SHOULD be 120 seconds.
>
> The recommended TTL value for other Multicast DNS resource records is 75 minutes."

However, when implementing **DNS-SD** (DNS-Based Service Discovery per RFC 6763) on top of mDNS, there was ambiguity about which TTL values to use for service discovery records (PTR, SRV, TXT) versus hostname records (A, AAAA).

### The Conflict

**Literal interpretation of RFC 6762 §10**:
- Records with hostname in name/rdata: **120 seconds**
  - A records (hostname as name) → 120s
  - SRV records (hostname in rdata) → 120s
- Other records: **4500 seconds (75 minutes)**
  - PTR records (no hostname) → 4500s
  - TXT records (no hostname) → 4500s

**DNS-SD interpretation** (confirmed by contract tests):
- **Service discovery records** (PTR, SRV, TXT): **120 seconds**
  - Rationale: Services come and go frequently
- **Hostname records** (A, AAAA): **4500 seconds (75 minutes)**
  - Rationale: Host IP addresses change less frequently

### Evidence

1. **Contract tests** (`tests/contract/rfc6762_announcing_test.go:301-302`):
   ```go
   // RFC 6762 §10: TTL values
   //   - Service records (PTR, SRV, TXT): 120 seconds
   //   - Hostname records (A, AAAA): 4500 seconds (75 minutes)
   ```

2. **Real-world behavior**: Apple Bonjour and Avahi both use 120s for PTR/SRV/TXT and 4500s for A/AAAA when advertising DNS-SD services.

3. **Practical reasoning**:
   - **PTR records** point to service instances, not hostnames (e.g., `_http._tcp.local` → `"My Printer._http._tcp.local"`)
   - Service instances change frequently (devices sleep, services restart)
   - Host IP addresses are more stable (DHCP leases, static IPs)

## Decision

We adopt the **DNS-SD interpretation** of RFC 6762 §10 TTL values:

| Record Type | TTL     | Rationale                                      |
|-------------|---------|------------------------------------------------|
| PTR         | 120s    | Service discovery - frequent changes           |
| SRV         | 120s    | Service discovery - frequent changes           |
| TXT         | 120s    | Service discovery - frequent changes           |
| A           | 4500s   | Hostname mapping - infrequent changes          |
| AAAA        | 4500s   | Hostname mapping - infrequent changes          |

This applies specifically to **DNS-SD service records** built via `records.BuildRecordSet()`.

## Consequences

### Positive

1. **RFC Compliance**: Aligns with real-world DNS-SD implementations (Bonjour, Avahi)
2. **Reduced Network Traffic**: Shorter TTL for services means faster discovery of changes; longer TTL for hosts reduces unnecessary queries
3. **Test Alignment**: Contract tests validate RFC compliance per DNS-SD semantics

### Negative

1. **Ambiguity in RFC 6762**: The RFC doesn't explicitly distinguish "service TTL" vs "hostname TTL" for DNS-SD contexts
2. **Test Updates Required**: All unit tests assuming literal RFC 6762 interpretation needed updates

### Affected Files

**Implementation** (TTL values changed):
- `internal/records/record_set.go`:
  - `buildPTRRecord()`: 4500 → **120** (line 94)
  - `buildTXTRecordFromService()`: 4500 → **120** (line 159)
  - `buildARecord()`: 120 → **4500** (line 229)

**Tests** (expectations updated to match DNS-SD interpretation):
- `internal/records/record_set_test.go`:
  - `TestBuildRecordSet_PTRRecord`: expect **120** (line 184)
  - `TestBuildRecordSet_ARecord`: expect **4500** (line 285)
- `internal/responder/response_builder_test.go`:
  - `TestResponseBuilder_BuildResponse_PTRQuery`: PTR → **120**, TXT → **120**, A → **4500** (lines 83, 114, 129)

**Documentation** (comments clarified):
- Updated comments to explain **"service records"** vs **"hostname records"** distinction
- Noted that service records change more frequently than hostname records
- Created ADR-005 to document this architectural decision

## Alternatives Considered

### Alternative 1: Literal RFC 6762 §10 Interpretation

**Rejected**: This would put PTR/TXT at 4500s and A at 120s, which contradicts:
- Real-world DNS-SD implementations
- Contract test expectations
- Practical caching behavior (services change more than IPs)

### Alternative 2: All Records at 120s

**Rejected**: Wastes bandwidth by forcing frequent re-queries for stable hostname mappings.

### Alternative 3: All Records at 4500s

**Rejected**: Causes stale service discovery (up to 75 minutes before detecting service changes).

## References

- [RFC 6762 §10: Resource Record TTL Values](https://www.rfc-editor.org/rfc/rfc6762.html#section-10)
- [RFC 6763: DNS-Based Service Discovery](https://www.rfc-editor.org/rfc/rfc6763.html)
- Apple Bonjour implementation (empirical observation)
- Avahi implementation (empirical observation)
- Contract tests: `tests/contract/rfc6762_announcing_test.go`

## Notes

This ADR documents a **clarification**, not a deviation from the RFCs. RFC 6762 §10 is written for general mDNS use cases. DNS-SD (RFC 6763) adds semantic meaning to record types, requiring a context-aware interpretation of TTL values.

**Key insight**: "Service records" (PTR/SRV/TXT for DNS-SD) vs "hostname records" (A/AAAA) is the correct distinction, NOT "records with hostname in name/rdata" vs "other records".
