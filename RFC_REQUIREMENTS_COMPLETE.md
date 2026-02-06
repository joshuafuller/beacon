# RFC 6762 Complete Requirements Database

**Generated**: 2026-01-06

**Scope**: All sections 1-22 of RFC 6762 (Multicast DNS)

## Summary

**Total Requirements**: 187

### By Type
- **MUST**: 80 (P0 - Mandatory)
- **MUST NOT**: 30 (P0 - Prohibited)
- **SHOULD**: 57 (P1 - Strong Recommendation)
- **SHOULD NOT**: 11 (P1 - Not Recommended)
- **MAY**: 9 (P2 - Optional)

### Implementation Status
- ✅ **Complete**: 183 (97%)
- ⚠️  **Partial**: 0 (0%)
- ❌ **Missing**: 4 (2%)

### P0 (MUST) Gap Analysis

- Total P0 requirements: 110
- ❌ Missing: 0
- ⚠️  Partial: 0
- ✅ Complete: 110

---

## Requirements by Section

### §1 Users may use these names as they would other DNS names,

**Progress**: 2/2 complete (100%)

#### RFC6762-§1-REQ-180 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since there is no central authority responsible for assigning dot-local names, and all devices on the local network are equally entitled to claim any dot-local name, users SHOULD be aware of this and SHOULD exercise appropriate caution.

**Implementation**:
- `internal/records/ttl.go`

**Tests**:
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§1-REQ-181 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In an untrusted or unfamiliar network environment, users SHOULD be aware that using a name like "www.local" may not actually connect them to the web site they expected, and could easily connect them to a different web page, or even a fake or spoof of their intended web site, designed to trick them into revealing confidential information.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/name.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/socket_darwin.go`
- `internal/transport/socket_linux.go`
- `internal/transport/socket_windows.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/socket_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §2 Conventions and Terminology Used in This Document

**Progress**: 3/3 complete (100%)

#### RFC6762-§2-REQ-001 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in "Key words for use in RFCs to Indicate Requirement Levels" [RFC2119].

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/name.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/socket_windows.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/registry_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/socket_windows_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§2-REQ-002 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Before claiming ownership of a unique resource record set, a responder MUST probe to verify that no other responder already claims ownership of that set, as described in Section 8.1, "Probing".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§2-REQ-003 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> (For fault-tolerance and other reasons, sometimes it is permissible to have more than one responder answering for a particular "unique" resource record set, but such cooperating responders MUST give answers containing identical rdata for these records.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §3 Multicast DNS Names

**Progress**: 3/5 complete (60%)

#### RFC6762-§3-REQ-004 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If this happens, the computer (or its human user) MUST cease using the name, and SHOULD attempt to allocate a new unique name for use on that link.

**Implementation**:
- `internal/message/builder.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/registry_test.go`
- `internal/security/security_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§3-REQ-005 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any DNS query for a name ending with ".local." MUST be sent to the mDNS IPv4 link-local multicast address 224.0.0.251 (or its IPv6 equivalent FF02::FB).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§3-REQ-006 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> Implementers MAY choose to look up such names concurrently via other mechanisms (e.g., Unicast DNS) and coalesce the results in some fashion.

**Implementation**: NOT IMPLEMENTED

**Tests**:
- `internal/responder/response_builder_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§3-REQ-007 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> DNS queries for names that do not end with ".local." MAY be sent to the mDNS multicast address, if no other conventional DNS server is available.

**Implementation**:
- `internal/protocol/mdns.go`
- `internal/state/announcer.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/querier.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/state/announcer_test.go`

---

#### RFC6762-§3-REQ-182 ❌

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> as special and SHOULD NOT send queries for these names to their configured (unicast) caching DNS server(s).

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §4 Reverse Address Mapping

**Progress**: 4/4 complete (100%)

#### RFC6762-§4-REQ-008 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Like ".local.", the IPv4 and IPv6 reverse mapping domains are also defined to be link-local: Any DNS query for a name ending with "254.169.in-addr.arpa." MUST be sent to the mDNS IPv4 link-local multicast address 224.0.0.251 or the mDNS IPv6 multicast address FF02::FB.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§4-REQ-009 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Likewise, any DNS query for a name within the reverse mapping domains for IPv6 link-local addresses ("8.e.f.ip6.arpa.", "9.e.f.ip6.arpa.", "a.e.f.ip6.arpa.", and "b.e.f.ip6.arpa.") MUST be sent to the mDNS IPv6 link-local multicast address FF02::FB or the mDNS IPv4 link-local multicast address 224.0.0.251.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§4-REQ-183 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> SHOULD NOT attempt to look up NS records for them, or otherwise query authoritative DNS servers in an attempt to resolve these names.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§4-REQ-184 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Instead, caching DNS servers SHOULD generate immediate NXDOMAIN responses for all such queries they may receive (from misbehaving name resolver libraries).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §5 Querying

**Progress**: 2/3 complete (66%)

#### RFC6762-§5-REQ-010 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Except in the rare case of a Multicast DNS responder that is advertising only shared resource records and no unique records, a Multicast DNS responder MUST also implement a Multicast DNS querier so that it can first verify the uniqueness of those records before it begins answering queries for them.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5-REQ-185 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> to answer queries for these names, and, like caching DNS servers, SHOULD generate immediate NXDOMAIN responses for all such queries they may receive.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5-REQ-186 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> DNS server software MAY provide a configuration option to override this default, for testing purposes or other specialized uses.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §5.1 One-Shot Multicast DNS Queries

**Progress**: 1/1 complete (100%)

#### RFC6762-§5.1-REQ-011 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> These queries are typically done using a high-numbered ephemeral UDP source port, but regardless of whether they are sent from a dynamic port or from a fixed port, these queries MUST NOT be sent using UDP source port 5353, since using UDP source port 5353 signals the presence of a fully compliant Multicast DNS querier, as described below.

**Implementation**:
- `internal/protocol/mdns.go`
- `internal/state/announcer.go`
- `internal/transport/socket_linux.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/querier.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/state/announcer_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/udp_test.go`

---

### §5.2 Continuous Multicast DNS Querying

**Progress**: 9/9 complete (100%)

#### RFC6762-§5.2-REQ-012 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Therefore, when retransmitting Multicast DNS queries to implement this kind of continuous monitoring, the interval between the first two queries MUST be at least one second, the intervals between successive queries MUST increase by at least a factor of two, and the querier MUST implement Known-Answer Suppression, as described below in Section 7.1.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `querier/collectresponses_test.go`

---

#### RFC6762-§5.2-REQ-013 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> When the interval between queries reaches or exceeds 60 minutes, a querier MAY cap the interval to a maximum of 60 minutes, and perform subsequent queries at a steady-state rate of one query per hour.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§5.2-REQ-014 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> To avoid accidental synchronization when, for some reason, multiple clients begin querying at exactly the same moment (e.g., because of some common external trigger event), a Multicast DNS querier SHOULD also delay the first query of the series by a randomly chosen amount in the range 20-120 ms.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§5.2-REQ-015 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> After this interval has passed, the answer will no longer be valid and SHOULD be deleted from the cache.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§5.2-REQ-016 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Before the record expiry time is reached, a Multicast DNS querier that has local clients with an active interest in the state of that record (e.g., a network browsing window displaying a list of discovered services to the user) SHOULD reissue its query to determine whether the record is still valid.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.2-REQ-017 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST NOT perform this cache maintenance for records for which it has no local clients with an active interest.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.2-REQ-018 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> An additional efficiency optimization SHOULD be performed when a Multicast DNS response is received containing a unique answer (as indicated by the cache-flush bit being set, described in Section 10.2, "Announcements to Flush Outdated Cache Entries").

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.2-REQ-019 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In this case, the Multicast DNS querier SHOULD plan to issue its next query for this record at 80-82% of the record's TTL, as described above.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.2-REQ-020 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A compliant Multicast DNS querier, which implements the rules specified in this document, MUST send its Multicast DNS queries from UDP source port 5353 (the well-known port assigned to mDNS), and MUST listen for Multicast DNS replies sent to UDP destination port 5353 at the mDNS link-local multicast address (224.0.0.251 and/or its IPv6 equivalent FF02::FB).

**Implementation**:
- `internal/protocol/mdns.go`
- `internal/state/announcer.go`
- `internal/transport/socket_linux.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/querier.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/state/announcer_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/udp_test.go`

---

### §5.4 Questions Requesting Unicast Responses

**Progress**: 7/7 complete (100%)

#### RFC6762-§5.4-REQ-021 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier sending its initial batch of questions immediately on wake from sleep or interface activation SHOULD set the unicast-response bit in those questions.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.4-REQ-022 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a question is retransmitted (as described in Section 5.2), the unicast-response bit SHOULD NOT be set in subsequent retransmissions of that question.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.4-REQ-023 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Subsequent retransmissions SHOULD be usual "QM" questions.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/responder/response_builder.go`
- `internal/state/prober.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/responder/response_builder_test.go`
- `querier/collectresponses_test.go`
- `responder/handlequery_test.go`

---

#### RFC6762-§5.4-REQ-024 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In addition, the unicast-response bit SHOULD be set only for questions that are active and ready to be sent the moment of wake from sleep or interface activation.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.4-REQ-025 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> New questions created by local clients afterwards should be treated as normal "QM" questions and SHOULD NOT have the unicast-response bit set on the first question of the series.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.4-REQ-026 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When receiving a question with the unicast-response bit set, a responder SHOULD usually respond with a unicast packet directed back to the querier.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§5.4-REQ-027 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> However, if the responder has not multicast that record recently (within one quarter of its TTL), then the responder SHOULD instead multicast the response so as to keep all the peer caches up to date, and to permit passive conflict detection.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §5.5 Direct Unicast Queries to Port 5353

**Progress**: 2/2 complete (100%)

#### RFC6762-§5.5-REQ-028 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a Multicast DNS responder receives a query via direct unicast, it SHOULD respond as it would for "QU" questions, as described above in Section 5.4.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§5.5-REQ-029 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since it is possible for a unicast query to be received from a machine outside the local link, responders SHOULD check that the source address in the query packet matches the local subnet for that link (or, in the case of IPv6, the source address has an on-link prefix) and silently ignore the packet if not.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

### §6 Responding

**Progress**: 21/21 complete (100%)

#### RFC6762-§6-REQ-030 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT place records from its cache, which have been learned from other responders on the network, in the Resource Record Sections of outgoing response messages.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-031 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> As with Unicast DNS, generally only DNS class 1 ("Internet") is used, but should client software use classes other than 1, the matching rules described above MUST be used.

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/response_builder_test.go`
- `querier/querier_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§6-REQ-032 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST only respond when it has a positive, non-null response to send, or it authoritatively knows that a particular record does not exist.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-033 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For unique records, where the host has already established sole ownership of the name, it MUST return negative answers to queries for records that it knows not to exist.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-034 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For example, a host with no IPv6 address, that has claimed sole ownership of the name "host.local." for all rrtypes, MUST respond to AAAA queries for "host.local." by sending a negative answer indicating that no AAAA records exist for that name.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-035 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For shared records, NXDOMAIN and other error responses MUST NOT be sent.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-036 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS responses MUST NOT contain any questions in the Question Section.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-037 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any questions in the Question Section of a received Multicast DNS response MUST be silently ignored.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-038 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder on Ethernet [IEEE.802.3] and similar shared multiple access networks SHOULD have the capability of delaying its responses by up to 500 ms, as described below.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-039 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where a Multicast DNS responder has good reason to believe that it will be the only responder on the link that will send a response (i.e., because it is able to answer every question in the query message, and for all of those answer records it has previously verified that the name, rrtype, and rrclass are unique on the link), it SHOULD NOT impose any random delay before responding, and SHOULD normally generate its response within at most 10 ms.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-040 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In any case where there may be multiple responses, such as queries where the answer is a member of a shared resource record set, each responder SHOULD delay its response by a random amount of time selected with uniform random distribution in the range 20-120 ms.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-041 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where the query has the TC (truncated) bit set, indicating that subsequent Known-Answer packets will follow, responders SHOULD delay their responses by a random amount of time selected with uniform random distribution in the range 400-500 ms, to allow enough time for all the Known-Answer packets to arrive, as described in Section 7.2, "Multipacket Known-Answer Suppression".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-042 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The source UDP port in all Multicast DNS responses MUST be 5353 (the well-known port assigned to mDNS).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-043 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS implementations MUST silently ignore any Multicast DNS responses they receive where the source UDP port is not 5353.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-044 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The destination UDP port in all Multicast DNS responses MUST be 5353, and the destination address MUST be the mDNS IPv4 link-local multicast address 224.0.0.251 or its IPv6 equivalent FF02::FB, except when generating a reply to a query that explicitly requested a unicast response: * via the unicast-response bit, * by virtue of being a legacy query (Section 6.7), or * by virtue of being a direct unicast query.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-045 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Except for these three specific cases, responses MUST NOT be sent via unicast, because then the "Passive Observation of Failures" mechanisms described in Section 10.5 would not work correctly.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-046 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST only accept unicast responses if they answer a recently sent query (e.g., sent within the last two seconds) that explicitly requested unicast responses.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-047 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST silently ignore all other unicast responses.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-048 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To protect the network against excessive packet flooding due to software bugs or malicious attack, a Multicast DNS responder MUST NOT (except in the one special case of answering probe queries) multicast a record on a given interface until at least one second has elapsed since the last time that record was multicast on that particular interface.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6-REQ-049 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the special case of answering probe queries, because of the limited time before the probing host will make its decision about whether or not to use the name, a Multicast DNS responder MUST respond quickly.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§6-REQ-187 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since name resolver libraries and caching DNS servers SHOULD NOT send queries for those names (see 3 and 4 above), such queries SHOULD be suppressed before they even reach the authoritative DNS server in question, and consequently it will not even get an opportunity to answer them.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `querier/collectresponses_test.go`
- `responder/handlequery_test.go`

---

### §6.1 Negative Responses

**Progress**: 14/14 complete (100%)

#### RFC6762-§6.1-REQ-050 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any time a responder receives a query for a name for which it has verified exclusive ownership, for a type for which that name has no records, the responder MUST (except as allowed in (a) below) respond asserting the nonexistence of that record using a DNS NSEC record [RFC4034].

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-051 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> On receipt of a question for a particular name, rrtype, and rrclass, for which a responder does have one or more unique answers, the responder MAY also include an NSEC record in the Additional Record Section indicating the nonexistence of other rrtypes for that name and rrclass.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-052 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> Implementers working with devices with sufficient memory and CPU resources MAY choose to implement code to handle the full generality of the DNS NSEC record [RFC4034], including bitmaps up to 65,536 bits long.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-053 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To facilitate use by devices with limited memory and CPU resources, Multicast DNS queriers are only REQUIRED to be able to parse a restricted form of the DNS NSEC record.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-054 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> All compliant Multicast DNS implementations MUST at least correctly generate and parse the restricted DNS NSEC record format described below: o The 'Next Domain Name' field contains the record's own name.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-055 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Consequently, if a Multicast DNS responder were to have records with rrtypes above 255, it MUST NOT generate these restricted-form NSEC records for those names, since to do so would imply that the name has no records with rrtypes above 255, which would be false.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-056 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In such cases a Multicast DNS responder MUST either (a) emit no NSEC record for that name, or (b) emit a full NSEC record containing the appropriate Type Bit Map block(s) with the correct bits set for all the record types that exist.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-057 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS implementation receives an NSEC record where the 'Next Domain Name' field is not the record's own name, then the implementation SHOULD ignore the 'Next Domain Name' field and process the remainder of the NSEC record as usual.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-058 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In Multicast DNS the 'Next Domain Name' field is not currently used, but it could be used in a future version of this protocol, which is why a Multicast DNS implementation MUST NOT reject or ignore an NSEC record it receives just because it finds an unexpected value in the 'Next Domain Name' field.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-059 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS implementation receives an NSEC record containing more than one Type Bit Map, or where the Type Bit Map block number is not zero, or where the block length is not in the range 1-32, then the Multicast DNS implementation MAY silently ignore the entire NSEC record.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-060 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS implementation MUST NOT ignore an entire message just because that message contains one or more NSEC record(s) that the Multicast DNS implementation cannot parse.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-061 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To help differentiate these synthesized NSEC records (generated programmatically on-the-fly) from conventional Unicast DNS NSEC records (which actually exist in a signed DNS zone), the synthesized Multicast DNS NSEC records MUST NOT have the NSEC bit set in the Type Bit Map, whereas conventional Unicast DNS NSEC records do have the NSEC bit set.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-062 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In general, the TTL given for an NSEC record SHOULD be the same as the TTL that the record would have had, had it existed.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.1-REQ-063 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A responder MUST only generate negative responses to queries for which it has legitimate ownership of the name, rrtype, and rrclass in question, and can legitimately assert that no record with that name, rrtype, and rrclass exists.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §6.2 Responding to Address Queries

**Progress**: 4/4 complete (100%)

#### RFC6762-§6.2-REQ-064 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When a Multicast DNS responder sends a Multicast DNS response message containing its own address records, it MUST include all addresses that are valid on the interface on which it is sending the message, and MUST NOT include addresses that are not valid on that interface (such as addresses that may be configured on the host's other interfaces).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.2-REQ-065 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a Multicast DNS responder places an IPv4 or IPv6 address record (rrtype "A" or "AAAA") into a response message, it SHOULD also place any records of the other address type with the same name into the additional section, if there is space in the message.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.2-REQ-066 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the event that a device has only IPv4 addresses but no IPv6 addresses, or vice versa, then the appropriate NSEC record SHOULD be placed into the additional section, so that queriers can know with certainty that the device has no addresses of that kind.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.2-REQ-067 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Other Multicast DNS responders may treat this case as logically two interfaces (one with one or more IPv4 addresses, and the other with one or more IPv6 addresses), but responders that operate this way MUST NOT put the corresponding automatic NSEC records in replies they send (i.e., a negative IPv4 assertion in their IPv6 responses, and a negative IPv6 assertion in their IPv4 responses) because this would cause incorrect operation in responders on the network that work the former way.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §6.3 Responding to Multiquestion Queries

**Progress**: 2/2 complete (100%)

#### RFC6762-§6.3-REQ-068 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS responders MUST correctly handle DNS query messages containing more than one question, by answering any or all of the questions to which they have answers.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§6.3-REQ-069 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Unlike single-question queries, where responding without delay is allowed in appropriate cases, for query messages containing more than one question, all (non-defensive) answers SHOULD be randomly delayed in the range 20-120 ms, or 400-500 ms if the TC (truncated) bit is set.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

### §6.4 Response Aggregation

**Progress**: 2/2 complete (100%)

#### RFC6762-§6.4-REQ-070 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When possible, a responder SHOULD, for the sake of network efficiency, aggregate as many responses as possible into a single Multicast DNS response message.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.4-REQ-071 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> For example, when a responder has several responses it plans to send, each delayed by a different interval, then earlier responses SHOULD be delayed by up to an additional 500 ms if that will permit them to be aggregated with other responses scheduled to go out a little later.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §6.5 Wildcard Queries (qtype "ANY" and qclass "ANY")

**Progress**: 2/2 complete (100%)

#### RFC6762-§6.5-REQ-072 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When responding to queries using qtype "ANY" (255) and/or qclass "ANY" (255), a Multicast DNS responder MUST respond with *ALL* of its records that match the query.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.5-REQ-073 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When responding to queries using qtype "ANY" (255) and/or qclass "ANY" (255), a Multicast DNS responder MUST respond with *ALL* of its records that match the query.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §6.6 Cooperating Multicast DNS Responders

**Progress**: 2/2 complete (100%)

#### RFC6762-§6.6-REQ-074 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS responder ("A") observes some other Multicast DNS responder ("B") send a Multicast DNS response message containing a resource record with the same name, rrtype, and rrclass as one of A's resource records, but *different* rdata, then: o If A's resource record is intended to be a shared resource record, then this is no conflict, and no action is required. o If A's resource record is intended to be a member of a unique resource record set owned solely by that responder, then this is a conflict and MUST be handled as described in Section 9, "Conflict Resolution".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.6-REQ-075 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS responder ("A") observes some other Multicast DNS responder ("B") send a Multicast DNS response message containing a resource record with the same name, rrtype, and rrclass as one of A's resource records, and *identical* rdata, then: o If the TTL of B's resource record given in the message is at least half the true TTL from A's point of view, then no action is required. o If the TTL of B's resource record given in the message is less than half the true TTL from A's point of view, then A MUST mark its record to be announced via multicast.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §6.7 Legacy Unicast Responses

**Progress**: 4/4 complete (100%)

#### RFC6762-§6.7-REQ-076 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In this case, the Multicast DNS responder MUST send a UDP response directly back to the querier, via unicast, to the query packet's source IP address and port.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.7-REQ-077 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> This unicast response MUST be a conventional unicast response as would be generated by a conventional Unicast DNS server; for example, it MUST repeat the query ID and the question given in the query message.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.7-REQ-078 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In addition, the cache-flush bit described in Section 10.2, "Announcements to Flush Outdated Cache Entries", MUST NOT be set in legacy unicast responses.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§6.7-REQ-079 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> The resource record TTL given in a legacy unicast response SHOULD NOT be greater than ten seconds, even if the true TTL of the Multicast DNS resource record is higher.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §7.1 Known-Answer Suppression

**Progress**: 4/4 complete (100%)

#### RFC6762-§7.1-REQ-080 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT answer a Multicast DNS query if the answer it would give is already included in the Answer Section with an RR TTL at least half the correct value.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§7.1-REQ-081 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If the RR TTL of the answer as given in the Answer Section is less than half of the true RR TTL as known by the Multicast DNS responder, the responder MUST send an answer so as to update the querier's cache before the record becomes in danger of expiration.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§7.1-REQ-082 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Therefore, a Multicast DNS querier SHOULD NOT include records in the Known-Answer list whose remaining TTL is less than half of their original TTL.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§7.1-REQ-083 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST NOT cache resource records observed in the Known-Answer Section of other Multicast DNS queries.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §7.2 Multipacket Known-Answer Suppression

**Progress**: 4/4 complete (100%)

#### RFC6762-§7.2-REQ-084 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It MUST then set the TC (Truncated) bit in the header before sending the query.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§7.2-REQ-085 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It MUST immediately follow the packet with another query packet containing no questions and as many more Known-Answer records as will fit.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§7.2-REQ-086 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If the responder sees any of its answers listed in the Known-Answer lists of subsequent packets from the querying host, it MUST delete that answer from the list of answers it is planning to give (provided that no other host on the network has also issued a query for that record and is waiting to receive an answer).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§7.2-REQ-087 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If the responder receives additional Known-Answer packets with the TC bit set, it SHOULD extend the delay as necessary to ensure a pause of 400-500 ms after the last such packet before it sends its answer.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `querier/collectresponses_test.go`

---

### §7.3 Duplicate Question Suppression

**Progress**: 1/1 complete (100%)

#### RFC6762-§7.3-REQ-088 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a host is planning to transmit (or retransmit) a query, and it sees another host on the network send a query containing the same "QM" question, and the Known-Answer Section of that query does not contain any records that this host would not also put in its own Known-Answer Section, then this host SHOULD treat its own query as having been sent.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §7.4 Duplicate Answer Suppression

**Progress**: 1/1 complete (100%)

#### RFC6762-§7.4-REQ-089 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a host is planning to send an answer, and it sees another host on the network send a response message containing the same answer record, and the TTL in that record is not less than the TTL this host would have given, then this host SHOULD treat its own answer as having been sent, and not also send an identical answer itself.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §8 Probing and Announcing on Startup

**Progress**: 1/1 complete (100%)

#### RFC6762-§8-REQ-090 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Whenever a Multicast DNS responder starts up, wakes up from sleep, receives an indication of a network interface "Link Change" event, or has any other reason to believe that its network connectivity may have changed in some relevant way, it MUST perform the two startup steps below: Probing (Section 8.1) and Announcing (Section 8.3).

**Implementation**:
- `internal/message/parser.go`
- `internal/records/record_set.go`
- `internal/security/source_filter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/options.go`
- `responder/responder.go`

**Tests**:
- `internal/records/record_set_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

### §8.1 Probing

**Progress**: 11/11 complete (100%)

#### RFC6762-§8.1-REQ-091 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The first startup step is that, for all those resource records that a Multicast DNS responder desires to be unique on the local link, it MUST send a Multicast DNS query asking for those resource records, to see if any of them are already in use.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-092 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> All probe queries SHOULD be done using the desired resource record name and class (usually class 1, "Internet"), and query type "ANY" (255), to elicit answers for all types of records with that name.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-093 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> (Note that probing is the one exception from the normal rule that there should be at least one second between repetitions of the same question, and the interval between subsequent repetitions should at least double.) When sending probe queries, a host MUST NOT consult its cache for potential answers.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§8.1-REQ-094 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Hence, it is important that when a device receives a probe query for a name that it is currently using, it SHOULD generate its response to defend that name immediately and send it as quickly as possible.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-095 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Because of the mDNS multicast rate-limiting rules, the probes SHOULD be sent as "QU" questions with the unicast- response bit set, to allow a defending host to respond immediately via unicast, instead of potentially having to wait before replying via multicast.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-096 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> During probing, from the time the first probe packet is sent until 250 ms after the third probe, if any conflicting Multicast DNS response is received, then the probing host MUST defer to the existing host, and SHOULD choose new names for some or all of its resource records as appropriate.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-097 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Apparently conflicting Multicast DNS responses received *before* the first probe packet is sent MUST be silently ignored (see discussion of stale probe packets in Section 8.2, "Simultaneous Probe Tiebreaking", below).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-098 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the case of a host probing using query type "ANY" as recommended above, any answer containing a record with that name, of any type, MUST be considered a conflicting response and handled accordingly.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-099 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If fifteen conflicts occur within any ten-second period, then the host MUST wait at least five seconds before each successive additional probe attempt.

**Implementation**:
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-100 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a responder knows by other means that its unique resource record set name, rrtype, and rrclass cannot already be in use by any other responder on the network, then it SHOULD skip the probing step for that resource record set.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.1-REQ-101 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Similarly, if a responder is acting as a proxy, taking over from another Multicast DNS responder that has already verified the uniqueness of the record, then the proxy SHOULD NOT repeat the probing step for those records.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §8.2 Simultaneous Probe Tiebreaking

**Progress**: 1/1 complete (100%)

#### RFC6762-§8.2-REQ-102 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the case of resource records containing rdata that is subject to name compression [RFC1035], the names MUST be uncompressed before comparison.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §8.3 Announcing

**Progress**: 6/6 complete (100%)

#### RFC6762-§8.3-REQ-103 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The second startup step is that the Multicast DNS responder MUST send an unsolicited Multicast DNS response containing, in the Answer Section, all of its newly registered resource records (both shared records, and unique records that have completed the probing step).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.3-REQ-104 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The Multicast DNS responder MUST send at least two unsolicited responses, one second apart.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.3-REQ-105 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> To provide increased robustness against packet loss, a responder MAY send up to eight unsolicited responses, provided that the interval between unsolicited responses increases by at least a factor of two with every response sent.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.3-REQ-106 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT send announcements in the absence of information that its network connectivity may have changed in some relevant way.

**Implementation**:
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `responder/responder.go`

**Tests**:
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§8.3-REQ-107 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In particular, a Multicast DNS responder MUST NOT send regular periodic announcements as a matter of course.

**Implementation**:
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `responder/responder.go`

**Tests**:
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§8.3-REQ-108 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Whenever a Multicast DNS responder receives any Multicast DNS response (solicited or otherwise) containing a conflicting resource record, the conflict MUST be resolved as described in Section 9, "Conflict Resolution".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §8.4 Updating

**Progress**: 5/5 complete (100%)

#### RFC6762-§8.4-REQ-109 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> At any time, if the rdata of any of a host's Multicast DNS records changes, the host MUST repeat the Announcing step described above to update neighboring caches.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.4-REQ-110 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For example, if any of a host's IP addresses change, it MUST re-announce those address records.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.4-REQ-111 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the case of shared records, a host MUST send a "goodbye" announcement with RR TTL zero (see Section 10.1, "Goodbye Packets") for the old rdata, to cause it to be deleted from peer caches, before announcing the new rdata.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.4-REQ-112 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case of unique records, a host SHOULD omit the "goodbye" announcement, since the cache-flush bit on the newly announced records will cause old rdata to be flushed from peer caches anyway.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§8.4-REQ-113 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A host may update the contents of any of its records at any time, though a host SHOULD NOT update records more frequently than ten times per minute.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §9 Conflict Resolution

**Progress**: 3/3 complete (100%)

#### RFC6762-§9-REQ-114 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Whenever a Multicast DNS responder receives any Multicast DNS response (solicited or otherwise) containing a conflicting resource record in any of the Resource Record Sections, the Multicast DNS responder MUST immediately reset its conflicted unique record to probing state, and go through the startup steps described above in Section 8, "Probing and Announcing on Startup".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§9-REQ-115 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The protocol used in the Probing phase will determine a winner and a loser, and the loser MUST cease using the name, and reconfigure.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/name.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/socket_windows.go`
- `querier/doc.go`
- `querier/options.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/registry_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/socket_windows_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§9-REQ-116 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It is very important that any host receiving a resource record that conflicts with one of its own MUST take action as described above.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §10 Resource Record TTL Values and Cache Coherency

**Progress**: 1/1 complete (100%)

#### RFC6762-§10-REQ-117 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> As a general rule, the recommended TTL value for Multicast DNS resource records with a host name as the resource record's name (e.g., A, AAAA, HINFO) or a host name contained within the resource record's rdata (e.g., SRV, reverse mapping PTR record) SHOULD be 120 seconds.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §10.1 Goodbye Packets

**Progress**: 2/2 complete (100%)

#### RFC6762-§10.1-REQ-118 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where a host knows that certain resource record data is about to become invalid (for example, when the host is undergoing a clean shutdown), the host SHOULD send an unsolicited Multicast DNS response packet, giving the same resource record name, rrtype, rrclass, and rdata, but an RR TTL of zero.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.1-REQ-119 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Queriers receiving a Multicast DNS response with a TTL of zero SHOULD NOT immediately delete the record from the cache, but instead record a TTL of 1 and then delete the record one second later.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §10.2 Announcements to Flush Outdated Cache Entries

**Progress**: 7/7 complete (100%)

#### RFC6762-§10.2-REQ-120 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In cases where the host has not been continuously connected and participating on the network link, it MUST first probe to re-verify uniqueness of its unique records, as described above in Section 8.1, "Probing".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.2-REQ-121 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Having completed the Probing step, if necessary, the host MUST then send a series of unsolicited announcements to update cache entries in its neighbor hosts.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§10.2-REQ-122 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any time a host sends a response packet containing some members of a unique RRSet, it MUST send the entire RRSet, preferably in a single packet, or if the entire RRSet will not fit in a single packet, in a quick burst of packets sent as close together as possible.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.2-REQ-123 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The host MUST set the cache-flush bit on all members of the unique RRSet.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.2-REQ-124 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT be set in any resource records in a response message sent in legacy unicast responses to UDP ports other than 5353.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.2-REQ-125 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT be set in any resource records in the Known-Answer list of any query message.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.2-REQ-126 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT ever be set in any shared resource record.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §10.4 Cache Flush on Failure Indication

**Progress**: 2/2 complete (100%)

#### RFC6762-§10.4-REQ-127 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When the cache receives this hint that it should reconfirm some record, it MUST issue two or more queries for the resource record in dispute.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.4-REQ-128 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If no response is received within ten seconds, then, even though its TTL may indicate that it is not yet due to expire, that record SHOULD be promptly flushed from the cache.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §10.5 Passive Observation Of Failures (POOF)

**Progress**: 2/2 complete (100%)

#### RFC6762-§10.5-REQ-129 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> After seeing two or more of these queries, and seeing no multicast response containing the expected answer within ten seconds, then even though its TTL may indicate that it is not yet due to expire, that record SHOULD be flushed from the cache.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§10.5-REQ-130 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> The host SHOULD NOT perform its own queries to reconfirm that the record is truly gone.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §11 Source Address Check

**Progress**: 4/4 complete (100%)

#### RFC6762-§11-REQ-131 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> All Multicast DNS responses (including responses sent via unicast) SHOULD be sent with IP TTL set to 255.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§11-REQ-132 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A host sending Multicast DNS queries to a link-local destination address (including the 224.0.0.251 and FF02::FB link-local multicast addresses) MUST only accept responses to that query that originate from the local link, and silently discard any other response packets.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§11-REQ-133 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since queriers will ignore responses apparently originating outside the local subnet, a responder SHOULD avoid generating responses that it can reasonably predict will be ignored.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§11-REQ-134 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a responder receives a query addressed to the mDNS IPv4 link-local multicast address 224.0.0.251, from a source address not apparently on the same subnet as the responder (or, in the case of IPv6, from a source IPv6 address for which the responder does not have any address with the same prefix on that interface), then even if the query indicates that a unicast response is preferred (see Section 5.4, "Questions Requesting Unicast Responses"), the responder SHOULD elect to respond by multicast anyway, since it can reasonably predict that a unicast response with an apparently non-local source address will probably be ignored.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §13 Enabling and Disabling Multicast DNS

**Progress**: 1/1 complete (100%)

#### RFC6762-§13-REQ-135 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> The option to fail-over to Multicast DNS for names not ending in ".local." SHOULD be a user-configured option, and SHOULD be disabled by default because of the possible security issues related to unintended local resolution of apparently global names.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/name.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/socket_windows.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/socket_windows_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §14 Considerations for Multiple Interfaces

**Progress**: 2/2 complete (100%)

#### RFC6762-§14-REQ-136 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A host SHOULD defend its dot-local host name on all active interfaces on which it is answering Multicast DNS queries.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§14-REQ-137 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Except in the case of proxying and other similar specialized uses, addresses in IPv4 or IPv6 address records in Multicast DNS responses MUST be valid for use on the interface on which the response is being sent.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §15.1 Receiving Unicast Responses

**Progress**: 1/1 complete (100%)

#### RFC6762-§15.1-REQ-138 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> For this reason, all Multicast DNS implementations SHOULD use the SO_REUSEPORT and/or SO_REUSEADDR options (or equivalent as appropriate for the operating system in question) so they will all be able to bind to UDP port 5353 and receive incoming multicast packets addressed to that port.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `internal/transport/socket_linux.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/querier.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `responder/handlequery_test.go`

---

### §16 Multicast DNS Character Set

**Progress**: 4/4 complete (100%)

#### RFC6762-§16-REQ-139 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Accordingly, it adopts the simple obvious elegant solution: all names in Multicast DNS MUST be encoded as precomposed UTF-8 [RFC3629] "Net-Unicode" [RFC5198] text.

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/querier.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§16-REQ-140 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS names MUST NOT contain a "Byte Order Mark".

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§16-REQ-141 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any occurrence of the Unicode character U+FEFF at the start or anywhere else in a Multicast DNS name MUST be interpreted as being an actual intended part of the name, representing (just as for any other legal unicode value) an actual literal instance of that character (in this case a zero-width non- breaking space character).

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§16-REQ-142 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS implementations MUST NOT use any other encodings apart from precomposed UTF-8 (US-ASCII being considered a compatible subset of UTF-8).

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

### §17 Multicast DNS Message Size

**Progress**: 5/5 complete (100%)

#### RFC6762-§17-REQ-143 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case of a single Multicast DNS resource record that is too large to fit in a single MTU-sized multicast response packet, a Multicast DNS responder SHOULD send the resource record alone, in a single IP datagram, using multiple IP fragments.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§17-REQ-144 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Resource records this large SHOULD be avoided, except in the very rare cases where they really are the appropriate solution to the problem at hand.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§17-REQ-145 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Implementers should be aware that many simple devices do not reassemble fragmented IP datagrams, so large resource records SHOULD NOT be used except in specialized cases where the implementer knows that all receivers implement reassembly, or where the large resource record contains optional data which is not essential for correct operation of the client.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§17-REQ-146 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS packet larger than the interface MTU, which is sent using fragments, MUST NOT contain more than one resource record.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/options.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§17-REQ-147 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Even when fragmentation is used, a Multicast DNS packet, including IP and UDP headers, MUST NOT exceed 9000 bytes.

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

### §18.1 ID (Query Identifier)

**Progress**: 6/6 complete (100%)

#### RFC6762-§18.1-REQ-148 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS implementations SHOULD listen for unsolicited responses issued by hosts booting up (or waking up from sleep or otherwise joining the network).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.1-REQ-149 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since these unsolicited responses may contain a useful answer to a question for which the querier is currently awaiting an answer, Multicast DNS implementations SHOULD examine all received Multicast DNS response messages for useful answers, without regard to the contents of the ID field or the Question Section.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.1-REQ-150 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS implementations MAY cache data from any or all Multicast DNS response messages they receive, for possible future use, provided of course that normal TTL aging is performed on these cached resource records.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.1-REQ-151 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In multicast query messages, the Query Identifier SHOULD be set to zero on transmission.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§18.1-REQ-152 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In multicast responses, including unsolicited multicast responses, the Query Identifier MUST be set to zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.1-REQ-153 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In legacy unicast response messages generated specifically in response to a particular (unicast or multicast) query, the Query Identifier MUST match the ID from the query message.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.2 QR (Query/Response) Bit

**Progress**: 2/2 complete (100%)

#### RFC6762-§18.2-REQ-154 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In query messages the QR bit MUST be zero.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§18.2-REQ-155 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In response messages the QR bit MUST be one.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.3 OPCODE

**Progress**: 2/2 complete (100%)

#### RFC6762-§18.3-REQ-156 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the OPCODE MUST be zero on transmission (only standard queries are currently supported over multicast).

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.3-REQ-157 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS messages received with an OPCODE other than zero MUST be silently ignored.

**Implementation**:
- `internal/message/message.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/protocol/mdns_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/options_test.go`
- `responder/responder_test.go`

---

### §18.4 AA (Authoritative Answer) Bit

**Progress**: 2/2 complete (100%)

#### RFC6762-§18.4-REQ-158 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In query messages, the Authoritative Answer bit MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§18.4-REQ-159 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In response messages for Multicast domains, the Authoritative Answer bit MUST be set to one (not setting this bit would imply there's some other place where "better" information may be found) and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.5 TC (Truncated) Bit

**Progress**: 3/3 complete (100%)

#### RFC6762-§18.5-REQ-160 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A responder SHOULD record this fact, and wait for those additional Known-Answer records, before deciding whether to respond.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.5-REQ-161 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In multicast response messages, the TC bit MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.5-REQ-162 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In legacy unicast response messages, the TC bit has the same meaning as in conventional Unicast DNS: it means that the response was too large to fit in a single packet, so the querier SHOULD reissue its query using TCP in order to receive the larger response.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.6 RD (Recursion Desired) Bit

**Progress**: 1/1 complete (100%)

#### RFC6762-§18.6-REQ-163 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the Recursion Desired bit SHOULD be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.7 RA (Recursion Available) Bit

**Progress**: 1/1 complete (100%)

#### RFC6762-§18.7-REQ-164 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the Recursion Available bit MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.8 Z (Zero) Bit

**Progress**: 1/1 complete (100%)

#### RFC6762-§18.8-REQ-165 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both query and response messages, the Zero bit MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.9 AD (Authentic Data) Bit

**Progress**: 1/1 complete (100%)

#### RFC6762-§18.9-REQ-166 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the Authentic Data bit [RFC2535] MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.10 CD (Checking Disabled) Bit

**Progress**: 1/1 complete (100%)

#### RFC6762-§18.10-REQ-167 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the Checking Disabled bit [RFC2535] MUST be zero on transmission, and MUST be ignored on reception.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.11 RCODE (Response Code)

**Progress**: 2/2 complete (100%)

#### RFC6762-§18.11-REQ-168 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In both multicast query and multicast response messages, the Response Code MUST be zero on transmission.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.11-REQ-169 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS messages received with non-zero Response Codes MUST be silently ignored.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §18.14 Name Compression

**Progress**: 8/9 complete (88%)

#### RFC6762-§18.14-REQ-170 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When generating Multicast DNS messages, implementations SHOULD use name compression wherever possible to compress the names of resource records, by replacing some or all of the resource record name with a compact two-byte reference to an appearance of that data somewhere earlier in the message [RFC1035].

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.14-REQ-171 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a query contains more than one question, successive questions in the same message often contain similar names, and consequently name compression SHOULD be used, to save bytes.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§18.14-REQ-172 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In addition, queries may also contain Known Answers in the Answer Section, or probe tiebreaking data in the Authority Section, and these names SHOULD similarly be compressed for network efficiency.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`

---

#### RFC6762-§18.14-REQ-173 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In addition to compressing the *names* of resource records, names that appear within the *rdata* of the following rrtypes SHOULD also be compressed in all Multicast DNS messages: NS, CNAME, PTR, DNAME, SOA, MX, AFSDB, RT, KX, RP, PX, SRV, NSEC Until future IETF Standards Action [RFC5226] specifying that names in the rdata of other types should be compressed, names that appear within the rdata of any type not listed above MUST NOT be compressed.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.14-REQ-174 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Implementations receiving Multicast DNS messages MUST correctly decode compressed names appearing in the Question Section, and compressed names of resource records appearing in other sections.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.14-REQ-175 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In addition, implementations MUST correctly decode compressed names appearing within the *rdata* of the rrtypes listed above.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/name.go`
- `internal/protocol/mdns.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/security/rate_limiter.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/mock.go`
- `internal/transport/socket_darwin.go`
- `internal/transport/socket_linux.go`
- `internal/transport/socket_windows.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/registry_test.go`
- `internal/responder/response_builder_test.go`
- `internal/security/security_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/socket_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/options_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.14-REQ-176 ❌

- **Type**: SHOULD
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> Where possible, implementations SHOULD also correctly decode compressed names appearing within the *rdata* of other rrtypes known to the implementers at the time of implementation, because such forward- thinking planning helps facilitate the deployment of future implementations that may have reason to compress those rrtypes.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§18.14-REQ-177 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Since all Multicast DNS implementations were created after 1996, all Multicast DNS implementations are REQUIRED to decode compressed SRV records correctly.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `querier/doc.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

#### RFC6762-§18.14-REQ-178 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In legacy unicast responses generated to answer legacy queries, name compression MUST NOT be performed on SRV records.

**Implementation**:
- `internal/message/builder.go`
- `internal/message/message.go`
- `internal/message/parser.go`
- `internal/protocol/mdns.go`
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/responder/conflict.go`
- `internal/responder/registry.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/transport/mock.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/doc.go`
- `querier/options.go`
- `querier/querier.go`
- `querier/records.go`
- `responder/conflict_detector.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/message/builder_response_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`
- `internal/message/name_test.go`
- `internal/message/parser_test.go`
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/prober_test.go`
- `internal/transport/mock_test.go`
- `internal/transport/transport_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `querier/querier_test.go`
- `responder/conflict_detector_test.go`
- `responder/handlequery_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

### §21 Security Considerations

**Progress**: 1/1 complete (100%)

#### RFC6762-§21-REQ-179 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To avoid this, a host MUST NOT append the search suffix ".local.", if present, to any relative (partially qualified) host name containing two or more labels.

**Implementation**:
- `internal/protocol/validator.go`
- `internal/records/record_set.go`
- `internal/responder/response_builder.go`
- `internal/security/source_filter.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `internal/transport/buffer_pool.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/service.go`

**Tests**:
- `internal/protocol/mdns_test.go`
- `internal/protocol/validator_test.go`
- `internal/records/record_set_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/transport/udp_test.go`
- `querier/collectresponses_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`

---

