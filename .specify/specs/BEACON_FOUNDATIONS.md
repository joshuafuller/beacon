# Beacon Foundations

**Version**: 1.1
**Purpose**: Shared context and reference for all Beacon specifications and implementations
**Audience**: Spec writers, developers, testers, contributors
**Governance**: Development governed by [Beacon Constitution v1.0.0](../memory/constitution.md)

This document establishes the foundational knowledge required to understand, specify, and implement Beacon. All other specifications reference this document rather than re-explaining basics.

**Note**: Beacon development follows strict governance under the Beacon Constitution, which mandates RFC compliance, spec-driven development, TDD, and phased implementation. All specifications and implementations must align with constitutional principles.

---

## Table of Contents

1. [DNS Fundamentals](#1-dns-fundamentals)
2. [Multicast DNS Essentials](#2-multicast-dns-essentials)
3. [DNS-SD Concepts](#3-dns-sd-concepts)
4. [Beacon Architecture](#4-beacon-architecture)
5. [Terminology Glossary](#5-terminology-glossary)
6. [Common Requirements](#6-common-requirements)
7. [Reference Tables](#7-reference-tables)

---

## 1. DNS Fundamentals

### 1.1 Domain Names and Labels

**Domain Name**: A hierarchical name consisting of labels separated by dots.
- Example: `myhost.local.`
- Root domain: `.` (single dot)
- Fully Qualified Domain Name (FQDN): Domain name ending with root (trailing dot)

**Label**: A single component of a domain name.
- Example: In `myhost.local.`, the labels are `myhost` and `local`
- Maximum length: 63 bytes
- Total domain name maximum: 255 bytes

**Label Encoding**:
- Traditional DNS: ASCII or Punycode for internationalized names
- mDNS: Direct UTF-8 encoding (no Punycode)
- Case-insensitive for ASCII letters (a-z matches A-Z)
- Case-preserving for storage and display

### 1.2 DNS Records

**Resource Record (RR)**: The fundamental unit of DNS data.

**Structure**:
```
NAME:   Domain name this record belongs to
TYPE:   Record type (A, AAAA, PTR, SRV, TXT, etc.)
CLASS:  Typically IN (Internet), value 1
TTL:    Time To Live in seconds (how long to cache)
RDLENGTH: Length of RDATA in bytes
RDATA:  Type-specific data
```

**Common Record Types**:

**A Record** (Type 1):
- Maps domain name to IPv4 address
- RDATA: 4 bytes (IPv4 address)
- Example: `myhost.local. 120 IN A 192.168.1.100`

**AAAA Record** (Type 28):
- Maps domain name to IPv6 address
- RDATA: 16 bytes (IPv6 address)
- Example: `myhost.local. 120 IN AAAA fe80::1`

**PTR Record** (Type 12):
- Pointer to another domain name
- Used for reverse DNS and service enumeration
- RDATA: Domain name
- Example: `_http._tcp.local. PTR myservice._http._tcp.local.`

**SRV Record** (Type 33):
- Service locator (host and port)
- RDATA: Priority, Weight, Port, Target
- Example: `myservice._http._tcp.local. SRV 0 0 8080 myhost.local.`
- Format: `priority weight port target`

**TXT Record** (Type 16):
- Arbitrary text data, used for metadata
- RDATA: One or more text strings
- Example: `myservice._http._tcp.local. TXT "path=/api" "version=1.0"`

**NSEC Record** (Type 47):
- Negative response in mDNS (proves non-existence)
- RDATA: Next domain name, type bitmap
- Used differently in mDNS than DNSSEC (for negative responses, not authentication)

### 1.3 DNS Messages

**Message Structure**:
```
+------------------+
| Header           | (12 bytes)
+------------------+
| Question Section | (variable)
+------------------+
| Answer Section   | (variable)
+------------------+
| Authority Section| (variable)
+------------------+
| Additional Section| (variable)
+------------------+
```

**Header Fields** (12 bytes):
```
ID:      16-bit query identifier
FLAGS:   16-bit flags (QR, OPCODE, AA, TC, RD, RA, Z, RCODE)
QDCOUNT: Number of questions
ANCOUNT: Number of answer records
NSCOUNT: Number of authority records
ARCOUNT: Number of additional records
```

**Key Header Flags**:
- **QR** (Query/Response): 0 = Query, 1 = Response
- **OPCODE**: Operation code (0 = standard query)
- **AA** (Authoritative Answer): Set by authoritative servers
- **TC** (Truncation): Message was truncated
- **RD** (Recursion Desired): Not used in mDNS
- **RA** (Recursion Available): Not used in mDNS
- **RCODE**: Response code (0 = no error)

**Question Format**:
```
QNAME:  Domain name being queried
QTYPE:  Record type requested (or ANY for all types)
QCLASS: Query class (IN for Internet)
```

**Query vs Response**:
- **Query**: QR=0, contains questions, typically no answers
- **Response**: QR=1, contains answers (may repeat questions in unicast DNS, not in mDNS)

### 1.4 Name Compression

To reduce packet size, DNS uses **name compression**:
- Labels can reference earlier labels via pointers
- Pointer format: Top 2 bits set (0xC0), lower 14 bits are offset
- Example: `myservice._http._tcp.local.` and `other._http._tcp.local.` share `_http._tcp.local.`

**Compression Rules**:
- Can compress in any section (questions, answers, authority, additional)
- Can compress within RDATA for types that contain domain names (SRV, PTR)
- MUST support parsing compressed names
- SHOULD generate compressed names to save space

---

## 2. Multicast DNS Essentials

### 2.1 What is Multicast DNS?

**Definition**: mDNS allows DNS-like queries and responses over IP multicast on a local link, without requiring a central DNS server.

**Key Differences from Unicast DNS**:
- Uses **multicast** (one-to-many) instead of unicast (one-to-one)
- Operates on **link-local scope** (single network segment)
- Uses port **5353** instead of 53
- Queries sent to multicast group, all hosts receive
- Any host can respond authoritatively for its own records
- No delegation, no central authority

### 2.2 The .local Domain

**Special-Use Domain**: `.local` is reserved for mDNS use only.

**Semantics**:
- Any query for a `.local` name MUST be sent to mDNS (not unicast DNS)
- Link-local scope only (does not cross routers without special config)
- No global uniqueness guarantees (name conflicts possible)
- No delegation (no NS records, no SOA records)

**Examples**:
- Host name: `myhost.local.`
- Service: `myservice._http._tcp.local.`

**Reverse Mappings**:
- IPv4 link-local: `169.254.in-addr.arpa.`
- IPv6 link-local: Reverse domain for `fe80::/10`

### 2.3 Multicast Addressing

**IPv4**:
- Multicast address: `224.0.0.251`
- Port: `5353`
- Sent to: `224.0.0.251:5353`

**IPv6**:
- Multicast address: `FF02::FB`
- Port: `5353`
- Sent to: `[FF02::FB]:5353`

**Link-Local Scope**: Multicast packets do not cross routers (TTL=255, scope=link).

### 2.4 Query Types

**One-Shot Query**:
- Sent to discover information once
- Source port: Ephemeral (random high port)
- Destination: Multicast address:5353
- Waits for responses, then stops

**Continuous Query**:
- Ongoing monitoring for changes
- Source port: **5353** (required for continuous queries)
- Destination: Multicast address:5353
- Receives all responses on the link (not just directed to this query)
- Used for service browsing, live updates

**Query ID**: Typically set to 0 in mDNS queries (unlike unicast DNS).

### 2.5 Unicast-Response Preference (QU Bit)

**Purpose**: Request that responder sends unicast response instead of multicast.

**Mechanism**: Top bit of QCLASS field in question.
- QCLASS = 0x0001 (IN): Multicast response requested
- QCLASS = 0x8001 (IN | 0x8000): Unicast response requested (QU bit set)

**Use Cases**:
- Startup queries (reduce multicast traffic)
- Known-Answer suppression not needed
- Network with many devices (reduce load)

**Responder Behavior**:
- If QU set and responder has answer, MAY send unicast response
- Unicast response sent to querier's source address:port
- Still MAY multicast if beneficial to other listeners

### 2.6 Cache-Flush Bit

**Purpose**: Signal that cached records for this name should be replaced (not supplemented).

**Mechanism**: Top bit of RRCLASS field in resource record.
- RRCLASS = 0x0001 (IN): Normal record
- RRCLASS = 0x8001 (IN | 0x8000): Cache-flush record

**Semantics**:
- Receiver should flush cached records for this name/type that are >1 second old
- Only applies to **unique records** (not shared records like PTR)
- Enables efficient updates without waiting for TTL expiry

**Example**:
```
myhost.local. 120 IN|0x8000 A 192.168.1.100
                     ^^^^^^
                     Cache-flush bit set
```

### 2.7 Record Ownership

**Authoritative Records**:
- Records a host owns and can answer for
- Host's own A/AAAA records
- Services the host publishes
- MUST respond authoritatively

**Cached Records**:
- Records learned from other hosts' responses
- Stored for TTL duration
- MUST NOT be sent in responses (only authoritative records)

### 2.8 Shared vs Unique Records

**Shared Records**:
- Multiple hosts may have same name/type with different RDATA
- Example: PTR records (many services of same type)
- No conflict if multiple exist

**Unique Records**:
- Only one host should have a given name/type combination
- Example: A/AAAA records (hostname → IP)
- Conflict if another host claims the same
- MUST probe before claiming
- Use cache-flush bit when announcing

### 2.9 Probing and Announcing

**Probing**: Before claiming a unique name, MUST verify no one else is using it.
- Send 3 probe queries (type ANY for the name)
- Wait 250ms between probes
- If no conflicting response, name is claimed

**Announcing**: After successful probing, announce ownership.
- Send unsolicited response with records
- Send at least 2 announcements, 1 second apart
- Use cache-flush bit for unique records

**Conflict Detection**: Continuously monitor responses for conflicts.
- If conflict detected, MUST resolve (typically rename)

### 2.10 Goodbye Packets

**Purpose**: Signal that records are no longer valid.

**Mechanism**: Send record with **TTL=0**.
- Receivers should delete record from cache
- Actually set TTL=1 internally, delete after 1 second (grace period)

**Use Case**: Host shutting down, service stopping.

### 2.11 Timing and Rate Limiting

**Query Intervals**:
- Minimum 1 second between queries for same name/type
- Exponential backoff (double interval each retry)
- Maximum interval: 60 minutes

**Response Delays**:
- Shared records: Random delay 20-120ms
- Unique records (sole responder): No delay
- With TC bit: 400-500ms delay
- Exception: Probe responses (defending unique records) should have minimal delay (time-critical)

**Rate Limiting**:
- MUST NOT multicast same record more than once per second
- Exception: Probes (250ms intervals)

**Cache Maintenance**:
- Query at 80%, 85%, 90%, 95% of TTL to refresh
- If no response, consider record expired

---

## 3. DNS-SD Concepts

### 3.1 What is DNS-SD?

**DNS-Based Service Discovery** (RFC 6763): A convention for using DNS record types (PTR, SRV, TXT) to discover services on a network.

**Not a new protocol**: Uses existing DNS infrastructure (or mDNS) with specific naming conventions.

**Two Main Operations**:
1. **Browsing**: Discover available service instances of a type
2. **Resolving**: Get connection info (host, port, metadata) for a service instance

### 3.2 Service Naming Structure

**Service Instance Name**: Identifies a specific service instance.

**Format**: `<Instance>.<Service>.<Domain>`

**Components**:

**Instance**: User-visible service name (e.g., "My Printer", "John's Web Server")
- Single DNS label (but may contain dots, escaped with backslash)
- UTF-8 encoded
- Maximum 63 bytes
- User-configurable, human-readable

**Service**: Service type identifier (e.g., `_http._tcp`, `_printer._tcp`)
- Format: `_<servicename>._<proto>`
- `<servicename>`: Service name (≤15 chars, letters/digits/hyphens)
- `<proto>`: `_tcp` or `_udp`
- Registered with IANA

**Domain**: DNS domain where service is advertised (e.g., `local.`)
- For mDNS: Typically `local.`
- For unicast DNS: Organization's domain

**Examples**:
```
Instance:        "My Printer"
Service:         _printer._tcp
Domain:          local.
Service Instance Name: My\032Printer._printer._tcp.local.
                       ^^^^
                       Space escaped as \032 in wire format

Instance:        "Bob's Music Server"
Service:         _http._tcp
Domain:          local.
Service Instance Name: Bob's\032Music\032Server._http._tcp.local.
```

### 3.3 Service Types

**Service Type Format**: `_<servicename>._tcp` or `_<servicename>._udp`

**Common Examples**:
- `_http._tcp` - Web servers
- `_https._tcp` - Secure web servers
- `_ssh._tcp` - SSH servers
- `_printer._tcp` - Printers
- `_airplay._tcp` - AirPlay devices

**Service Name Rules**:
- Maximum 15 characters (excluding underscore)
- Letters, digits, hyphens only
- Must begin and end with letter or digit
- Must contain at least one letter
- Case-insensitive

**Protocol Suffix**:
- `_tcp` for TCP services
- `_udp` for all non-TCP services (including UDP, SCTP, or others)

### 3.4 Browsing (Service Enumeration)

**Goal**: Discover all instances of a service type.

**Mechanism**: PTR query.

**Query**:
```
QNAME:  _http._tcp.local.
QTYPE:  PTR
QCLASS: IN
```

**Response**: PTR records pointing to service instance names.
```
_http._tcp.local. PTR My\032Web\032Server._http._tcp.local.
_http._tcp.local. PTR John's\032Server._http._tcp.local.
```

**Continuous Browsing**:
- Use continuous query (source port 5353)
- Receive live updates as services appear/disappear
- Goodbye packets (TTL=0) indicate service removed

**User Interface**:
- Display only `<Instance>` portion to user
- Hide `<Service>.<Domain>` (implicit from browse operation)
- Show "My Web Server", "John's Server" (not full names)

### 3.5 Resolution (Service Instance Resolution)

**Goal**: Get connection information for a specific service instance.

**Mechanism**: SRV and TXT queries.

**Queries** (parallel):
```
QNAME:  My\032Web\032Server._http._tcp.local.
QTYPE:  SRV
QCLASS: IN

QNAME:  My\032Web\032Server._http._tcp.local.
QTYPE:  TXT
QCLASS: IN
```

**SRV Response**: Provides host and port.
```
My\032Web\032Server._http._tcp.local. SRV 0 0 8080 webserver.local.
                                           ^   ^   ^    ^
                                         Pri Wgt Port Target
```
- Priority: 0 (for single instance)
- Weight: 0 (for single instance)
- Port: Service port number
- Target: Host name (resolve to A/AAAA records)

**TXT Response**: Provides metadata as key/value pairs.
```
My\032Web\032Server._http._tcp.local. TXT "path=/admin" "version=2.0"
```

**Resolution Result**:
- Host: `webserver.local.` (resolve to IP via A/AAAA query)
- Port: `8080`
- Metadata: `{path: "/admin", version: "2.0"}`

### 3.6 TXT Record Format

**Purpose**: Carry service metadata as key/value pairs.

**Format**: One or more strings, each `key=value`.

**Rules**:
- Minimum (no data): RFC 6763 §6.1 defines three valid forms:
  1. A TXT record containing a single zero byte (RECOMMENDED)
  2. An empty (zero-length) TXT record (not strictly legal but should be accepted)
  3. No TXT record (NXDOMAIN or no-error-no-answer response)
- Keys: Printable ASCII (0x20-0x7E), excluding '='
- Keys: Case-insensitive, ≥1 character, SHOULD be ≤9 characters
- Values: Opaque bytes (often UTF-8 text)
- Missing '=' means boolean attribute (key present, no value)

**Size Limits**:
- SHOULD be ≤200 bytes total
- Preferably ≤400 bytes
- NOT RECOMMENDED >1300 bytes

**Examples**:
```
TXT "txtvers=1" "path=/api" "secure"
    ^^^^^^^^^^^ ^^^^^^^^^^^  ^^^^^^
    Version     Path value   Boolean (no =)
```

**Special Keys**:
- `txtvers=<number>`: Version of TXT record format (optional, should be first)

**Conflicts**:
- If same key appears multiple times, use FIRST occurrence
- Silently ignore unknown keys
- Ignore strings beginning with '='

### 3.7 Service Subtypes

**Purpose**: Narrow browsing to specific subset of a service type.

**Format**: `_<subtype>._sub._<service>._<proto>.<Domain>`

**Example**:
```
Service Type:  _http._tcp.local.
Subtype:       _printer._sub._http._tcp.local.
               ^^^^^^^^
               Subtype for printers with web interface
```

**Mechanism**: Additional PTR record.
```
# Base service
_http._tcp.local. PTR My\032Printer._http._tcp.local.

# Subtype (same target)
_printer._sub._http._tcp.local. PTR My\032Printer._http._tcp.local.
```

**Use Cases**:
- Client wants specific subset (not all instances)
- Examples: `_printer._sub._http._tcp` (only printers), `_scanner._sub._http._tcp` (only scanners)

### 3.8 Flagship Naming

**Purpose**: Coordinate unique names across related protocols.

**Example**: A printer might support multiple protocols.
- IPP (flagship): `My\032Printer._ipp._tcp.local.`
- LPR: `My\032Printer._printer._tcp.local.`
- Want same instance name across both

**Mechanism**: Non-flagship protocols create placeholder SRV record.
```
My\032Printer._printer._tcp.local. SRV 0 0 0 printer.local.
                                           ^     ^^^^^^^^^^^^
                                        Port=0   Target=hostname
```
- No PTR record (not browsable)
- Port=0 (placeholder)
- Target=hostname (for conflict detection)
- Ensures name coordination

---

## 4. Beacon Architecture

**Status**: Architecture specifications (F-2 through F-8) completed and validated as of 2025-11-01. All architecture specs have undergone RFC validation per Constitution requirements.

**Architecture Specifications**:
- F-2: Package Structure & Layering
- F-3: Error Handling Strategy
- F-4: Concurrency Model
- F-5: Configuration & Defaults
- F-6: Logging & Observability
- F-7: Resource Management
- F-8: Testing Strategy

See `.specify/specs/` directory for complete specifications.

### 4.1 System Layers

Beacon is organized into logical layers:

```
┌─────────────────────────────────────┐
│     Application Layer               │  User code using Beacon
│   (User's Go application)           │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     Public API Layer                │  beacon/querier
│   (Querier, Responder, Browser,    │  beacon/responder
│    Resolver, Publisher interfaces)  │  beacon/service
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     Service Layer                   │  Orchestration logic
│   (Lifecycle, Conflict, Browse,     │  State management
│    Resolution logic)                │  Event coordination
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     Protocol Layer                  │  beacon/internal/protocol
│   (Message parse/build, Query/      │  RFC 6762/6763 compliance
│    Response construction, Cache)    │  Wire format handling
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     Transport Layer                 │  beacon/internal/transport
│   (UDP multicast, Socket mgmt,      │  Network I/O
│    Interface monitoring)            │  OS integration
└─────────────────────────────────────┘
```

### 4.2 Key Components

**Querier**: Sends mDNS queries, receives responses.

**Responder**: Answers mDNS queries for owned records.

**Browser**: DNS-SD service browsing (PTR queries).

**Resolver**: DNS-SD service resolution (SRV/TXT queries).

**Publisher**: DNS-SD service advertisement.

**Lifecycle Manager**: Probing, announcing, conflict detection.

**Cache**: Stores learned records with TTL management.

**Message Parser/Builder**: Serializes/deserializes DNS messages.

**Transport**: UDP multicast socket operations.

### 4.3 Package Structure (Planned)

```
github.com/joshuafuller/beacon/
├── querier/          # Public API: Querying
├── responder/        # Public API: Responding
├── service/          # Public API: DNS-SD browsing/publishing
├── internal/
│   ├── protocol/     # Protocol logic (RFC compliance)
│   ├── transport/    # Network I/O
│   ├── cache/        # Record caching
│   └── message/      # DNS message format
└── examples/         # Usage examples
```

**Import Rules**:
- Public packages: Can be imported by users
- Internal packages: Cannot be imported externally (Go convention)
- Internal can import other internal
- Public can import internal
- Avoid circular dependencies

### 4.4 Data Flow

**Query Flow**:
```
User → Querier.Query() → Protocol.BuildQuery() → Transport.Send()
  ↓
  ← Querier.Results() ← Protocol.ParseResponse() ← Transport.Receive()
```

**Response Flow**:
```
Transport.Receive() → Protocol.ParseQuery() → Responder.HandleQuery()
  ↓
  → Protocol.BuildResponse() → Transport.Send()
```

**Browse Flow**:
```
User → Browser.Browse(serviceType) → Querier.Query(PTR)
  ↓
  ← Browser.Instances() ← Parse PTR responses
```

**Resolve Flow**:
```
User → Resolver.Resolve(instance) → Querier.Query(SRV + TXT)
  ↓
  ← Resolver.ServiceInfo() ← Parse SRV/TXT, resolve A/AAAA
```

---

## 5. Terminology Glossary

**Authoritative Record**: A resource record that a host owns and can answer for.

**Browsing**: Discovering available service instances of a given type (DNS-SD).

**Cache-Flush**: Mechanism to signal cached records should be replaced (top bit of RRCLASS).

**Continuous Query**: Ongoing monitoring query using source port 5353.

**Domain**: Hierarchical namespace (e.g., `local.`).

**FQDN**: Fully Qualified Domain Name, ending with root dot.

**Flagship Protocol**: Primary protocol in a family for name coordination.

**Goodbye Packet**: Record with TTL=0 indicating removal.

**Instance**: User-visible service name (first label of service instance name).

**Known-Answer Suppression**: Including known records in query to suppress redundant responses.

**Label**: Single component of domain name (between dots).

**Link-Local**: Scope limited to single network segment (no routing).

**Multicast**: One-to-many communication model.

**One-Shot Query**: Single query from ephemeral port.

**Probing**: Verifying name availability before claiming.

**PTR Record**: Pointer record, used for service enumeration.

**QU Bit**: Unicast-response preference (top bit of QCLASS).

**Querier**: Component that sends queries.

**Resolution**: Obtaining connection info for a service instance (DNS-SD).

**Resource Record (RR)**: Fundamental DNS data unit (name, type, class, TTL, rdata).

**Responder**: Component that answers queries.

**Service Instance Name**: Full name of service instance (`<Instance>.<Service>.<Domain>`).

**Service Type**: Category of service (e.g., `_http._tcp`).

**Shared Record**: Record type where multiple hosts may have same name/type (e.g., PTR).

**SRV Record**: Service locator record (priority, weight, port, target).

**Subtype**: Narrowing mechanism for service browsing.

**TTL**: Time To Live, cache duration in seconds.

**TXT Record**: Text record carrying key/value metadata.

**Unicast**: One-to-one communication model.

**Unique Record**: Record type where only one host should have name/type (e.g., A/AAAA).

---

## 6. Common Requirements

### 6.1 Character Encoding

**mDNS Names**: UTF-8 (NOT Punycode)
- Direct UTF-8 in DNS labels
- Precomposed form (NFC normalization)
- No Byte Order Mark (BOM)

**Case Sensitivity**:
- ASCII letters (a-z, A-Z): Case-insensitive
- Non-ASCII: Case-preserving, byte-wise comparison
- No special equivalences (é ≠ e)

### 6.2 TTL Values

**Host Name Records** (A/AAAA):
- SHOULD: 120 seconds

**Other Records** (SRV/TXT/PTR):
- SHOULD: 4500 seconds (75 minutes)

**Goodbye**:
- 0 seconds (processed as TTL=1 internally)

### 6.3 Timing Values

**Probing**:
- Initial delay: 0-250ms (random)
- Between probes: 250ms
- Number of probes: 3

**Announcing**:
- Between announcements: 1 second
- Minimum announcements: 2

**Query Intervals**:
- Minimum: 1 second
- Backoff: 2x each retry
- Maximum: 3600 seconds (60 minutes)

**Response Delays**:
- Shared records: 20-120ms (random)
- Unique records (sole responder): 0ms
- With TC bit: 400-500ms

**Cache Refresh**:
- Query at: 80%, 85%, 90%, 95% of TTL

### 6.4 Size Limits

**DNS Label**: 63 bytes maximum

**Domain Name**: 255 bytes maximum

**Service Name**: 15 characters maximum (excluding underscore)

**DNS Message**:
- Recommended: ≤ 1500 bytes (typical MTU)
- Maximum: 9000 bytes (multicast)
- Fragmentation: Avoid if possible

**TXT Record**:
- SHOULD: ≤200 bytes
- Preferably: ≤400 bytes
- NOT RECOMMENDED: >1300 bytes

### 6.5 Network Values

**IPv4 Multicast Address**: `224.0.0.251`

**IPv6 Multicast Address**: `FF02::FB`

**Port**: `5353` (both UDP source and destination for continuous queries)

**IP TTL**: 255 (max value, link-local only)

**Multicast Scope**: Link-local (no routing)

---

## 7. Reference Tables

### 7.1 DNS Record Types

| Type | Code | Name | Purpose |
|------|------|------|---------|
| A | 1 | IPv4 Address | Map name to IPv4 |
| AAAA | 28 | IPv6 Address | Map name to IPv6 |
| PTR | 12 | Pointer | Service enumeration, reverse DNS |
| SRV | 33 | Service | Port and target host |
| TXT | 16 | Text | Key/value metadata |
| NSEC | 47 | Next Secure | Negative responses (mDNS-specific use) |
| ANY | 255 | All | Query for all record types |

### 7.2 DNS Classes

| Class | Code | Name | Purpose |
|-------|------|------|---------|
| IN | 1 | Internet | Standard class |
| IN \| 0x8000 | 0x8001 | Internet + QU/Cache-Flush | Top bit set |

### 7.3 DNS Response Codes

| Code | Name | Meaning |
|------|------|---------|
| 0 | NOERROR | No error |
| 3 | NXDOMAIN | Name does not exist |

*Note: mDNS typically uses RCODE=0 even for negative responses (uses NSEC instead).*

### 7.4 Header Flags

| Flag | Bit | Meaning |
|------|-----|---------|
| QR | 15 | 0=Query, 1=Response |
| OPCODE | 14-11 | 0=Standard Query |
| AA | 10 | Authoritative Answer |
| TC | 9 | Truncation |
| RD | 8 | Recursion Desired (not used in mDNS) |
| RA | 7 | Recursion Available (not used in mDNS) |
| Z | 6-4 | Reserved (must be 0) |
| RCODE | 3-0 | Response code |

### 7.5 Default Configuration Values

| Parameter | Default Value | Rationale |
|-----------|---------------|-----------|
| Host Name TTL | 120s | Hosts may change IPs (DHCP, mobility) |
| Service TTL | 4500s (75m) | Services more stable than IPs |
| Probe Count | 3 | Balance speed and reliability |
| Probe Interval | 250ms | Fast startup, avoid collisions |
| Announce Count | 2+ | Ensure delivery on lossy links |
| Announce Interval | 1s | Allow caching, avoid storms |
| Max Query Interval | 3600s (60m) | Balance freshness and traffic |
| Cache Refresh Threshold | 80% of TTL | Proactive refresh before expiry |

---

## 8. Document Conventions

**Throughout Beacon specifications**:

**MUST, SHOULD, MAY**: Per RFC 2119 definitions.
- MUST: Absolute requirement
- SHOULD: Recommended, exceptions require justification
- MAY: Optional

**References to this document**:
- "See FOUNDATIONS §2.3" = Section 2.3 of this document
- "See FOUNDATIONS Table 7.1" = Table 7.1 of this document

**References to RFCs**:
- "RFC 6762 §5" = RFC 6762, Section 5
- "RFC 6763 §4.1" = RFC 6763, Section 4.1

**RFC Compliance**:
- See [RFC Compliance Matrix](../../docs/RFC_COMPLIANCE_MATRIX.md) for detailed section-by-section compliance status

**Code Examples**:
- Pseudocode uses Go-like syntax
- Not actual implementation (specs define behavior, not code)

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.1 | 2025-11-01 | Added constitution governance reference, updated architecture section to reflect completed F-series specs and RFC validation |
| 1.0 | 2025-10-31 | Initial version |

---

**This document is a living reference. It will be updated as Beacon evolves, but maintains backwards-compatible terminology and concepts.**
