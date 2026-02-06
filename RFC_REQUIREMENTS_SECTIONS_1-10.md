# RFC 6762 Requirements Database (Sections 1-10)

**Generated**: 2026-01-06

## Summary

- **Total Requirements**: 137
  - MUST/MUST NOT: 79 (P0)
  - SHOULD/SHOULD NOT: 39 (P1)
  - MAY/OPTIONAL: 8 (P2)

- **Implementation Status**:
  - ✅ Complete: 110
  - ⚠️  Partial: 2
  - ❌ Missing: 25

## Requirements by Section

### §1 Users may use these names as they would other DNS names,

#### RFC6762-§1-REQ-130 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since there is no central authority responsible for assigning dot-local names, and all devices on the local network are equally entitled to claim any dot-local name, users SHOULD be aware of this and SHOULD exercise appropriate caution.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§1-REQ-131 ❌

- **Type**: SHOULD
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> In an untrusted or unfamiliar network environment, users SHOULD be aware that using a name like "www.local" may not actually connect them to the web site they expected, and could easily connect them to a different web page, or even a fake or spoof of their intended web site, designed to trick them into revealing confidential information.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §2 Conventions and Terminology Used in This Document

#### RFC6762-§2-REQ-001 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in "Key words for use in RFCs to Indicate Requirement Levels" [RFC2119].

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§2-REQ-002 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Before claiming ownership of a unique resource record set, a responder MUST probe to verify that no other responder already claims ownership of that set, as described in Section 8.1, "Probing". (For fault-tolerance and other reasons, sometimes it is permissible to have more than one responder answering for a particular "unique" resource record set, but such cooperating responders MUST give answers containing identical rdata for these records.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

### §3 Multicast DNS Names

#### RFC6762-§3-REQ-003 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If this happens, the computer (or its human user) MUST cease using the name, and SHOULD attempt to allocate a new unique name for use on that link.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§3-REQ-004 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any DNS query for a name ending with ".local." MUST be sent to the mDNS IPv4 link-local multicast address 224.0.0.251 (or its IPv6 equivalent FF02::FB).

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§3-REQ-005 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> Implementers MAY choose to look up such names concurrently via other mechanisms (e.g., Unicast DNS) and coalesce the results in some fashion.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§3-REQ-006 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> DNS queries for names that do not end with ".local." MAY be sent to the mDNS multicast address, if no other conventional DNS server is available.

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§3-REQ-132 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> as special and SHOULD NOT send queries for these names to their configured (unicast) caching DNS server(s).

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §4 Reverse Address Mapping

#### RFC6762-§4-REQ-007 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Like ".local.", the IPv4 and IPv6 reverse mapping domains are also defined to be link-local: Any DNS query for a name ending with "254.169.in-addr.arpa." MUST be sent to the mDNS IPv4 link-local multicast address 224.0.0.251 or the mDNS IPv6 multicast address FF02::FB.

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§4-REQ-008 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Likewise, any DNS query for a name within the reverse mapping domains for IPv6 link-local addresses ("8.e.f.ip6.arpa.", "9.e.f.ip6.arpa.", "a.e.f.ip6.arpa.", and "b.e.f.ip6.arpa.") MUST be sent to the mDNS IPv6 link-local multicast address FF02::FB or the mDNS IPv4 link-local multicast address 224.0.0.251.

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§4-REQ-133 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> SHOULD NOT attempt to look up NS records for them, or otherwise query authoritative DNS servers in an attempt to resolve these names.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§4-REQ-134 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Instead, caching DNS servers SHOULD generate immediate NXDOMAIN responses for all such queries they may receive (from misbehaving name resolver libraries).

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §5 Querying

#### RFC6762-§5-REQ-009 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Except in the rare case of a Multicast DNS responder that is advertising only shared resource records and no unique records, a Multicast DNS responder MUST also implement a Multicast DNS querier so that it can first verify the uniqueness of those records before it begins answering queries for them.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§5-REQ-135 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> to answer queries for these names, and, like caching DNS servers, SHOULD generate immediate NXDOMAIN responses for all such queries they may receive.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5-REQ-136 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> DNS server software MAY provide a configuration option to override this default, for testing purposes or other specialized uses.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §5.1 One-Shot Multicast DNS Queries

#### RFC6762-§5.1-REQ-010 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> These queries are typically done using a high-numbered ephemeral UDP source port, but regardless of whether they are sent from a dynamic port or from a fixed port, these queries MUST NOT be sent using UDP source port 5353, since using UDP source port 5353 signals the presence of a fully compliant Multicast DNS querier, as described below.

**Implementation**:
- `querier/querier.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/socket_linux.go`
- `internal/state/announcer.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/transport/mock_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`

---

### §5.2 Continuous Multicast DNS Querying

#### RFC6762-§5.2-REQ-011 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Therefore, when retransmitting Multicast DNS queries to implement this kind of continuous monitoring, the interval between the first two queries MUST be at least one second, the intervals between successive queries MUST increase by at least a factor of two, and the querier MUST implement Known-Answer Suppression, as described below in Section 7.1.

**Implementation**:
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.2-REQ-012 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> When the interval between queries reaches or exceeds 60 minutes, a querier MAY cap the interval to a maximum of 60 minutes, and perform subsequent queries at a steady-state rate of one query per hour.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.2-REQ-013 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> To avoid accidental synchronization when, for some reason, multiple clients begin querying at exactly the same moment (e.g., because of some common external trigger event), a Multicast DNS querier SHOULD also delay the first query of the series by a randomly chosen amount in the range 20-120 ms.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.2-REQ-014 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> After this interval has passed, the answer will no longer be valid and SHOULD be deleted from the cache.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§5.2-REQ-015 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Before the record expiry time is reached, a Multicast DNS querier that has local clients with an active interest in the state of that record (e.g., a network browsing window displaying a list of discovered services to the user) SHOULD reissue its query to determine whether the record is still valid.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.2-REQ-016 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST NOT perform this cache maintenance for records for which it has no local clients with an active interest.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§5.2-REQ-017 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> An additional efficiency optimization SHOULD be performed when a Multicast DNS response is received containing a unique answer (as indicated by the cache-flush bit being set, described in Section 10.2, "Announcements to Flush Outdated Cache Entries").

**Implementation**:
- `querier/records.go`
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/responder/registry.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§5.2-REQ-018 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In this case, the Multicast DNS querier SHOULD plan to issue its next query for this record at 80-82% of the record's TTL, as described above.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§5.2-REQ-019 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A compliant Multicast DNS querier, which implements the rules specified in this document, MUST send its Multicast DNS queries from UDP source port 5353 (the well-known port assigned to mDNS), and MUST listen for Multicast DNS replies sent to UDP destination port 5353 at the mDNS link-local multicast address (224.0.0.251 and/or its IPv6 equivalent FF02::FB).

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/transport/socket_linux.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/transport/mock_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

### §5.4 Questions Requesting Unicast Responses

#### RFC6762-§5.4-REQ-020 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier sending its initial batch of questions immediately on wake from sleep or interface activation SHOULD set the unicast-response bit in those questions.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-021 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a question is retransmitted (as described in Section 5.2), the unicast-response bit SHOULD NOT be set in subsequent retransmissions of that question.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-022 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Subsequent retransmissions SHOULD be usual "QM" questions.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-023 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In addition, the unicast-response bit SHOULD be set only for questions that are active and ready to be sent the moment of wake from sleep or interface activation.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-024 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> New questions created by local clients afterwards should be treated as normal "QM" questions and SHOULD NOT have the unicast-response bit set on the first question of the series.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-025 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When receiving a question with the unicast-response bit set, a responder SHOULD usually respond with a unicast packet directed back to the querier.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.4-REQ-026 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> However, if the responder has not multicast that record recently (within one quarter of its TTL), then the responder SHOULD instead multicast the response so as to keep all the peer caches up to date, and to permit passive conflict detection.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/responder/conflict.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §5.5 Direct Unicast Queries to Port 5353

#### RFC6762-§5.5-REQ-027 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> When a Multicast DNS responder receives a query via direct unicast, it SHOULD respond as it would for "QU" questions, as described above in Section 5.4.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§5.5-REQ-028 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since it is possible for a unicast query to be received from a machine outside the local link, responders SHOULD check that the source address in the query packet matches the local subnet for that link (or, in the case of IPv6, the source address has an on-link prefix) and silently ignore the packet if not.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §6 Responding

#### RFC6762-§6-REQ-029 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT place records from its cache, which have been learned from other responders on the network, in the Resource Record Sections of outgoing response messages.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6-REQ-030 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> As with Unicast DNS, generally only DNS class 1 ("Internet") is used, but should client software use classes other than 1, the matching rules described above MUST be used.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-031 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> A Multicast DNS responder MUST only respond when it has a positive, non-null response to send, or it authoritatively knows that a particular record does not exist.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-032 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For unique records, where the host has already established sole ownership of the name, it MUST return negative answers to queries for records that it knows not to exist.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6-REQ-033 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For example, a host with no IPv6 address, that has claimed sole ownership of the name "host.local." for all rrtypes, MUST respond to AAAA queries for "host.local." by sending a negative answer indicating that no AAAA records exist for that name.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-034 ⚠️

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: PARTIAL

**Requirement**:
> For shared records, NXDOMAIN and other error responses MUST NOT be sent.

**Implementation**:
- `responder/conflict_detector.go`
- `internal/records/record_set.go`

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-035 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS responses MUST NOT contain any questions in the Question Section.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-036 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any questions in the Question Section of a received Multicast DNS response MUST be silently ignored.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-037 ⚠️

- **Type**: SHOULD
- **Priority**: P1
- **Status**: PARTIAL

**Requirement**:
> A Multicast DNS responder on Ethernet [IEEE.802.3] and similar shared multiple access networks SHOULD have the capability of delaying its responses by up to 500 ms, as described below.

**Implementation**:
- `responder/conflict_detector.go`
- `internal/records/record_set.go`

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-038 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where a Multicast DNS responder has good reason to believe that it will be the only responder on the link that will send a response (i.e., because it is able to answer every question in the query message, and for all of those answer records it has previously verified that the name, rrtype, and rrclass are unique on the link), it SHOULD NOT impose any random delay before responding, and SHOULD normally generate its response within at most 10 ms.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6-REQ-039 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In any case where there may be multiple responses, such as queries where the answer is a member of a shared resource record set, each responder SHOULD delay its response by a random amount of time selected with uniform random distribution in the range 20-120 ms.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-040 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where the query has the TC (truncated) bit set, indicating that subsequent Known-Answer packets will follow, responders SHOULD delay their responses by a random amount of time selected with uniform random distribution in the range 400-500 ms, to allow enough time for all the Known-Answer packets to arrive, as described in Section 7.2, "Multipacket Known-Answer Suppression".

**Implementation**:
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-041 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> The source UDP port in all Multicast DNS responses MUST be 5353 (the well-known port assigned to mDNS).

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-042 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> Multicast DNS implementations MUST silently ignore any Multicast DNS responses they receive where the source UDP port is not 5353.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-043 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The destination UDP port in all Multicast DNS responses MUST be 5353, and the destination address MUST be the mDNS IPv4 link-local multicast address 224.0.0.251 or its IPv6 equivalent FF02::FB, except when generating a reply to a query that explicitly requested a unicast response: * via the unicast-response bit, * by virtue of being a legacy query (Section 6.7), or * by virtue of being a direct unicast query.

**Implementation**:
- `querier/querier.go`
- `querier/options.go`
- `responder/responder.go`
- `internal/transport/transport.go`
- `internal/transport/udp.go`
- `internal/transport/ipv6_stub.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `querier/querier_test.go`
- `responder/responder_test.go`
- `internal/transport/udp_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/records/record_set_test.go`
- `internal/protocol/mdns_test.go`
- `internal/message/builder_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§6-REQ-044 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> Except for these three specific cases, responses MUST NOT be sent via unicast, because then the "Passive Observation of Failures" mechanisms described in Section 10.5 would not work correctly.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6-REQ-045 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST only accept unicast responses if they answer a recently sent query (e.g., sent within the last two seconds) that explicitly requested unicast responses.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-046 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST silently ignore all other unicast responses.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-047 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To protect the network against excessive packet flooding due to software bugs or malicious attack, a Multicast DNS responder MUST NOT (except in the one special case of answering probe queries) multicast a record on a given interface until at least one second has elapsed since the last time that record was multicast on that particular interface.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-048 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the special case of answering probe queries, because of the limited time before the probing host will make its decision about whether or not to use the name, a Multicast DNS responder MUST respond quickly.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6-REQ-137 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Since name resolver libraries and caching DNS servers SHOULD NOT send queries for those names (see 3 and 4 above), such queries SHOULD be suppressed before they even reach the authoritative DNS server in question, and consequently it will not even get an opportunity to answer them.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §6.1 Negative Responses

#### RFC6762-§6.1-REQ-049 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any time a responder receives a query for a name for which it has verified exclusive ownership, for a type for which that name has no records, the responder MUST (except as allowed in (a) below) respond asserting the nonexistence of that record using a DNS NSEC record [RFC4034].

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.1-REQ-050 ✅

- **Type**: MAY
- **Priority**: P2
- **Status**: COMPLETE

**Requirement**:
> On receipt of a question for a particular name, rrtype, and rrclass, for which a responder does have one or more unique answers, the responder MAY also include an NSEC record in the Additional Record Section indicating the nonexistence of other rrtypes for that name and rrclass.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6.1-REQ-051 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> Implementers working with devices with sufficient memory and CPU resources MAY choose to implement code to handle the full generality of the DNS NSEC record [RFC4034], including bitmaps up to 65,536 bits long.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-052 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> To facilitate use by devices with limited memory and CPU resources, Multicast DNS queriers are only REQUIRED to be able to parse a restricted form of the DNS NSEC record.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.1-REQ-053 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> All compliant Multicast DNS implementations MUST at least correctly generate and parse the restricted DNS NSEC record format described below: o The 'Next Domain Name' field contains the record's own name.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-054 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Consequently, if a Multicast DNS responder were to have records with rrtypes above 255, it MUST NOT generate these restricted-form NSEC records for those names, since to do so would imply that the name has no records with rrtypes above 255, which would be false.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.1-REQ-055 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> In such cases a Multicast DNS responder MUST either (a) emit no NSEC record for that name, or (b) emit a full NSEC record containing the appropriate Type Bit Map block(s) with the correct bits set for all the record types that exist.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-056 ❌

- **Type**: SHOULD
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> If a Multicast DNS implementation receives an NSEC record where the 'Next Domain Name' field is not the record's own name, then the implementation SHOULD ignore the 'Next Domain Name' field and process the remainder of the NSEC record as usual.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-057 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> In Multicast DNS the 'Next Domain Name' field is not currently used, but it could be used in a future version of this protocol, which is why a Multicast DNS implementation MUST NOT reject or ignore an NSEC record it receives just because it finds an unexpected value in the 'Next Domain Name' field.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-058 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> If a Multicast DNS implementation receives an NSEC record containing more than one Type Bit Map, or where the Type Bit Map block number is not zero, or where the block length is not in the range 1-32, then the Multicast DNS implementation MAY silently ignore the entire NSEC record.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-059 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> A Multicast DNS implementation MUST NOT ignore an entire message just because that message contains one or more NSEC record(s) that the Multicast DNS implementation cannot parse.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-060 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> To help differentiate these synthesized NSEC records (generated programmatically on-the-fly) from conventional Unicast DNS NSEC records (which actually exist in a signed DNS zone), the synthesized Multicast DNS NSEC records MUST NOT have the NSEC bit set in the Type Bit Map, whereas conventional Unicast DNS NSEC records do have the NSEC bit set.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.1-REQ-061 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In general, the TTL given for an NSEC record SHOULD be the same as the TTL that the record would have had, had it existed.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§6.1-REQ-062 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A responder MUST only generate negative responses to queries for which it has legitimate ownership of the name, rrtype, and rrclass in question, and can legitimately assert that no record with that name, rrtype, and rrclass exists.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §6.2 Responding to Address Queries

#### RFC6762-§6.2-REQ-063 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> When a Multicast DNS responder sends a Multicast DNS response message containing its own address records, it MUST include all addresses that are valid on the interface on which it is sending the message, and MUST NOT include addresses that are not valid on that interface (such as addresses that may be configured on the host's other interfaces).

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.2-REQ-064 ❌

- **Type**: SHOULD
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> When a Multicast DNS responder places an IPv4 or IPv6 address record (rrtype "A" or "AAAA") into a response message, it SHOULD also place any records of the other address type with the same name into the additional section, if there is space in the message.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.2-REQ-065 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the event that a device has only IPv4 addresses but no IPv6 addresses, or vice versa, then the appropriate NSEC record SHOULD be placed into the additional section, so that queriers can know with certainty that the device has no addresses of that kind.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.2-REQ-066 ❌

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> Other Multicast DNS responders may treat this case as logically two interfaces (one with one or more IPv4 addresses, and the other with one or more IPv6 addresses), but responders that operate this way MUST NOT put the corresponding automatic NSEC records in replies they send (i.e., a negative IPv4 assertion in their IPv6 responses, and a negative IPv6 assertion in their IPv4 responses) because this would cause incorrect operation in responders on the network that work the former way.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §6.3 Responding to Multiquestion Queries

#### RFC6762-§6.3-REQ-067 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Multicast DNS responders MUST correctly handle DNS query messages containing more than one question, by answering any or all of the questions to which they have answers.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.3-REQ-068 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Unlike single-question queries, where responding without delay is allowed in appropriate cases, for query messages containing more than one question, all (non-defensive) answers SHOULD be randomly delayed in the range 20-120 ms, or 400-500 ms if the TC (truncated) bit is set.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §6.4 Response Aggregation

#### RFC6762-§6.4-REQ-069 ❌

- **Type**: SHOULD
- **Priority**: P1
- **Status**: MISSING

**Requirement**:
> When possible, a responder SHOULD, for the sake of network efficiency, aggregate as many responses as possible into a single Multicast DNS response message.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§6.4-REQ-070 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> For example, when a responder has several responses it plans to send, each delayed by a different interval, then earlier responses SHOULD be delayed by up to an additional 500 ms if that will permit them to be aggregated with other responses scheduled to go out a little later.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §6.5 Wildcard Queries (qtype "ANY" and qclass "ANY")

#### RFC6762-§6.5-REQ-071 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When responding to queries using qtype "ANY" (255) and/or qclass "ANY" (255), a Multicast DNS responder MUST respond with *ALL* of its records that match the query.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.5-REQ-072 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When responding to queries using qtype "ANY" (255) and/or qclass "ANY" (255), a Multicast DNS responder MUST respond with *ALL* of its records that match the query.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §6.6 Cooperating Multicast DNS Responders

#### RFC6762-§6.6-REQ-073 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS responder ("A") observes some other Multicast DNS responder ("B") send a Multicast DNS response message containing a resource record with the same name, rrtype, and rrclass as one of A's resource records, but *different* rdata, then: o If A's resource record is intended to be a shared resource record, then this is no conflict, and no action is required. o If A's resource record is intended to be a member of a unique resource record set owned solely by that responder, then this is a conflict and MUST be handled as described in Section 9, "Conflict Resolution".

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6.6-REQ-074 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If a Multicast DNS responder ("A") observes some other Multicast DNS responder ("B") send a Multicast DNS response message containing a resource record with the same name, rrtype, and rrclass as one of A's resource records, and *identical* rdata, then: o If the TTL of B's resource record given in the message is at least half the true TTL from A's point of view, then no action is required. o If the TTL of B's resource record given in the message is less than half the true TTL from A's point of view, then A MUST mark its record to be announced via multicast.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §6.7 Legacy Unicast Responses

#### RFC6762-§6.7-REQ-075 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In this case, the Multicast DNS responder MUST send a UDP response directly back to the querier, via unicast, to the query packet's source IP address and port.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.7-REQ-076 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> This unicast response MUST be a conventional unicast response as would be generated by a conventional Unicast DNS server; for example, it MUST repeat the query ID and the question given in the query message.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§6.7-REQ-077 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In addition, the cache-flush bit described in Section 10.2, "Announcements to Flush Outdated Cache Entries", MUST NOT be set in legacy unicast responses.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§6.7-REQ-078 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> The resource record TTL given in a legacy unicast response SHOULD NOT be greater than ten seconds, even if the true TTL of the Multicast DNS resource record is higher.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §7.1 Known-Answer Suppression

#### RFC6762-§7.1-REQ-079 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT answer a Multicast DNS query if the answer it would give is already included in the Answer Section with an RR TTL at least half the correct value.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§7.1-REQ-080 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If the RR TTL of the answer as given in the Answer Section is less than half of the true RR TTL as known by the Multicast DNS responder, the responder MUST send an answer so as to update the querier's cache before the record becomes in danger of expiration.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§7.1-REQ-081 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Therefore, a Multicast DNS querier SHOULD NOT include records in the Known-Answer list whose remaining TTL is less than half of their original TTL.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§7.1-REQ-082 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS querier MUST NOT cache resource records observed in the Known-Answer Section of other Multicast DNS queries.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

### §7.2 Multipacket Known-Answer Suppression

#### RFC6762-§7.2-REQ-083 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It MUST then set the TC (Truncated) bit in the header before sending the query.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§7.2-REQ-084 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It MUST immediately follow the packet with another query packet containing no questions and as many more Known-Answer records as will fit.

**Implementation**:
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§7.2-REQ-085 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If the responder sees any of its answers listed in the Known-Answer lists of subsequent packets from the querying host, it MUST delete that answer from the list of answers it is planning to give (provided that no other host on the network has also issued a query for that record and is waiting to receive an answer).

**Implementation**:
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§7.2-REQ-086 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If the responder receives additional Known-Answer packets with the TC bit set, it SHOULD extend the delay as necessary to ensure a pause of 400-500 ms after the last such packet before it sends its answer.

**Implementation**:
- `internal/responder/response_builder.go`

**Tests**:
- `internal/responder/known_answer_test.go`

---

### §7.3 Duplicate Question Suppression

#### RFC6762-§7.3-REQ-087 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a host is planning to transmit (or retransmit) a query, and it sees another host on the network send a query containing the same "QM" question, and the Known-Answer Section of that query does not contain any records that this host would not also put in its own Known-Answer Section, then this host SHOULD treat its own query as having been sent.

**Implementation**:
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §7.4 Duplicate Answer Suppression

#### RFC6762-§7.4-REQ-088 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a host is planning to send an answer, and it sees another host on the network send a response message containing the same answer record, and the TTL in that record is not less than the TTL this host would have given, then this host SHOULD treat its own answer as having been sent, and not also send an identical answer itself.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §8 Probing and Announcing on Startup

#### RFC6762-§8-REQ-089 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> Whenever a Multicast DNS responder starts up, wakes up from sleep, receives an indication of a network interface "Link Change" event, or has any other reason to believe that its network connectivity may have changed in some relevant way, it MUST perform the two startup steps below: Probing (Section 8.1) and Announcing (Section 8.3).

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §8.1 Probing

#### RFC6762-§8.1-REQ-090 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The first startup step is that, for all those resource records that a Multicast DNS responder desires to be unique on the local link, it MUST send a Multicast DNS query asking for those resource records, to see if any of them are already in use.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.1-REQ-091 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> All probe queries SHOULD be done using the desired resource record name and class (usually class 1, "Internet"), and query type "ANY" (255), to elicit answers for all types of records with that name.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§8.1-REQ-092 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If, by 250 ms after the third probe, no conflicting Multicast DNS responses have been received, the host may move to the next step, announcing. (Note that probing is the one exception from the normal rule that there should be at least one second between repetitions of the same question, and the interval between subsequent repetitions should at least double.) When sending probe queries, a host MUST NOT consult its cache for potential answers.

**Implementation**:
- `querier/records.go`
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/responder/conflict.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.1-REQ-093 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Hence, it is important that when a device receives a probe query for a name that it is currently using, it SHOULD generate its response to defend that name immediately and send it as quickly as possible.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§8.1-REQ-094 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Because of the mDNS multicast rate-limiting rules, the probes SHOULD be sent as "QU" questions with the unicast- response bit set, to allow a defending host to respond immediately via unicast, instead of potentially having to wait before replying via multicast.

**Implementation**:
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§8.1-REQ-095 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> During probing, from the time the first probe packet is sent until 250 ms after the third probe, if any conflicting Multicast DNS response is received, then the probing host MUST defer to the existing host, and SHOULD choose new names for some or all of its resource records as appropriate.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`

---

#### RFC6762-§8.1-REQ-096 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Apparently conflicting Multicast DNS responses received *before* the first probe packet is sent MUST be silently ignored (see discussion of stale probe packets in Section 8.2, "Simultaneous Probe Tiebreaking", below).

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`

---

#### RFC6762-§8.1-REQ-097 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the case of a host probing using query type "ANY" as recommended above, any answer containing a record with that name, of any type, MUST be considered a conflicting response and handled accordingly.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`

---

#### RFC6762-§8.1-REQ-098 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> If fifteen conflicts occur within any ten-second period, then the host MUST wait at least five seconds before each successive additional probe attempt.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`

---

#### RFC6762-§8.1-REQ-099 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If a responder knows by other means that its unique resource record set name, rrtype, and rrclass cannot already be in use by any other responder on the network, then it SHOULD skip the probing step for that resource record set.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.1-REQ-100 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Similarly, if a responder is acting as a proxy, taking over from another Multicast DNS responder that has already verified the uniqueness of the record, then the proxy SHOULD NOT repeat the probing step for those records.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

### §8.2 Simultaneous Probe Tiebreaking

#### RFC6762-§8.2-REQ-101 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> In the case of resource records containing rdata that is subject to name compression [RFC1035], the names MUST be uncompressed before comparison. (The details of how a particular name is compressed is an artifact of how and where the record is written into the DNS message; it is not an intrinsic property of the resource record itself.) The bytes of the raw uncompressed rdata are compared in turn, interpreting the bytes as eight-bit UNSIGNED values, until a byte is found whose value is greater than that of its counterpart (in which case, the rdata whose byte has the greater value is deemed lexicographically later) or one of the resource records runs out of rdata (in which case, the resource record which still has remaining data first is deemed lexicographically later).

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

### §8.3 Announcing

#### RFC6762-§8.3-REQ-102 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The second startup step is that the Multicast DNS responder MUST send an unsolicited Multicast DNS response containing, in the Answer Section, all of its newly registered resource records (both shared records, and unique records that have completed the probing step).

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.3-REQ-103 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> The Multicast DNS responder MUST send at least two unsolicited responses, one second apart.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§8.3-REQ-104 ❌

- **Type**: MAY
- **Priority**: P2
- **Status**: MISSING

**Requirement**:
> To provide increased robustness against packet loss, a responder MAY send up to eight unsolicited responses, provided that the interval between unsolicited responses increases by at least a factor of two with every response sent.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§8.3-REQ-105 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> A Multicast DNS responder MUST NOT send announcements in the absence of information that its network connectivity may have changed in some relevant way.

**Implementation**:
- `responder/responder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`

**Tests**:
- `responder/responder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`

---

#### RFC6762-§8.3-REQ-106 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In particular, a Multicast DNS responder MUST NOT send regular periodic announcements as a matter of course.

**Implementation**:
- `responder/responder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`

**Tests**:
- `responder/responder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`

---

#### RFC6762-§8.3-REQ-107 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Whenever a Multicast DNS responder receives any Multicast DNS response (solicited or otherwise) containing a conflicting resource record, the conflict MUST be resolved as described in Section 9, "Conflict Resolution".

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`

---

### §8.4 Updating

#### RFC6762-§8.4-REQ-108 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> At any time, if the rdata of any of a host's Multicast DNS records changes, the host MUST repeat the Announcing step described above to update neighboring caches.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.4-REQ-109 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> For example, if any of a host's IP addresses change, it MUST re-announce those address records.

**Implementation**:
- `responder/responder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`

**Tests**:
- `responder/responder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`

---

#### RFC6762-§8.4-REQ-110 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In the case of shared records, a host MUST send a "goodbye" announcement with RR TTL zero (see Section 10.1, "Goodbye Packets") for the old rdata, to cause it to be deleted from peer caches, before announcing the new rdata.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§8.4-REQ-111 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case of unique records, a host SHOULD omit the "goodbye" announcement, since the cache-flush bit on the newly announced records will cause old rdata to be flushed from peer caches anyway.

**Implementation**:
- `querier/records.go`
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/responder/registry.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§8.4-REQ-112 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> A host may update the contents of any of its records at any time, though a host SHOULD NOT update records more frequently than ten times per minute.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

### §9 Conflict Resolution

#### RFC6762-§9-REQ-113 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Whenever a Multicast DNS responder receives any Multicast DNS response (solicited or otherwise) containing a conflicting resource record in any of the Resource Record Sections, the Multicast DNS responder MUST immediately reset its conflicted unique record to probing state, and go through the startup steps described above in Section 8, "Probing and Announcing on Startup".

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§9-REQ-114 ❌

- **Type**: MUST
- **Priority**: P0
- **Status**: MISSING

**Requirement**:
> The protocol used in the Probing phase will determine a winner and a loser, and the loser MUST cease using the name, and reconfigure.

**Implementation**: NOT IMPLEMENTED

**Tests**: NO TEST

---

#### RFC6762-§9-REQ-115 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> It is very important that any host receiving a resource record that conflicts with one of its own MUST take action as described above.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`

---

### §10 Resource Record TTL Values and Cache Coherency

#### RFC6762-§10-REQ-116 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> As a general rule, the recommended TTL value for Multicast DNS resource records with a host name as the resource record's name (e.g., A, AAAA, HINFO) or a host name contained within the resource record's rdata (e.g., SRV, reverse mapping PTR record) SHOULD be 120 seconds.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §10.1 Goodbye Packets

#### RFC6762-§10.1-REQ-117 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> In the case where a host knows that certain resource record data is about to become invalid (for example, when the host is undergoing a clean shutdown), the host SHOULD send an unsolicited Multicast DNS response packet, giving the same resource record name, rrtype, rrclass, and rdata, but an RR TTL of zero.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§10.1-REQ-118 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> Queriers receiving a Multicast DNS response with a TTL of zero SHOULD NOT immediately delete the record from the cache, but instead record a TTL of 1 and then delete the record one second later.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §10.2 Announcements to Flush Outdated Cache Entries

#### RFC6762-§10.2-REQ-119 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> In cases where the host has not been continuously connected and participating on the network link, it MUST first probe to re-verify uniqueness of its unique records, as described above in Section 8.1, "Probing".

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/registry.go`
- `internal/responder/conflict.go`
- `internal/state/machine.go`
- `internal/state/prober.go`
- `internal/state/states.go`
- `internal/records/record_set.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/conflict_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-120 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Having completed the Probing step, if necessary, the host MUST then send a series of unsolicited announcements to update cache entries in its neighbor hosts.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/state/machine.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/announcer_test.go`
- `internal/state/machine_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-121 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> Any time a host sends a response packet containing some members of a unique RRSet, it MUST send the entire RRSet, preferably in a single packet, or if the entire RRSet will not fit in a single packet, in a quick burst of packets sent as close together as possible.

**Implementation**:
- `responder/service.go`
- `responder/responder.go`
- `internal/responder/registry.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-122 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The host MUST set the cache-flush bit on all members of the unique RRSet.

**Implementation**:
- `querier/records.go`
- `responder/service.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/responder/registry.go`
- `internal/state/announcer.go`
- `internal/state/prober.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `responder/service_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/responder/registry_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-123 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT be set in any resource records in a response message sent in legacy unicast responses to UDP ports other than 5353.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-124 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT be set in any resource records in the Known-Answer list of any query message.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.2-REQ-125 ✅

- **Type**: MUST NOT
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> The cache-flush bit MUST NOT ever be set in any shared resource record.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_response_test.go`

---

### §10.4 Cache Flush on Failure Indication

#### RFC6762-§10.4-REQ-126 ✅

- **Type**: MUST
- **Priority**: P0
- **Status**: COMPLETE

**Requirement**:
> When the cache receives this hint that it should reconfirm some record, it MUST issue two or more queries for the resource record in dispute.

**Implementation**:
- `querier/records.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/records/record_set_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`

---

#### RFC6762-§10.4-REQ-127 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> If no response is received within ten seconds, then, even though its TTL may indicate that it is not yet due to expire, that record SHOULD be promptly flushed from the cache.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

### §10.5 Passive Observation Of Failures (POOF)

#### RFC6762-§10.5-REQ-128 ✅

- **Type**: SHOULD
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> After seeing two or more of these queries, and seeing no multicast response containing the expected answer within ten seconds, then even though its TTL may indicate that it is not yet due to expire, that record SHOULD be flushed from the cache.

**Implementation**:
- `querier/records.go`
- `querier/querier.go`
- `responder/responder.go`
- `responder/conflict_detector.go`
- `internal/responder/response_builder.go`
- `internal/state/announcer.go`
- `internal/records/record_set.go`
- `internal/records/ttl.go`
- `internal/protocol/mdns.go`
- `internal/message/message.go`
- `internal/message/builder.go`
- `internal/message/parser.go`

**Tests**:
- `querier/collectresponses_test.go`
- `responder/conflict_detector_test.go`
- `responder/responder_test.go`
- `internal/responder/known_answer_test.go`
- `internal/responder/response_builder_test.go`
- `internal/state/prober_test.go`
- `internal/records/record_set_test.go`
- `internal/records/ttl_test.go`
- `internal/message/builder_test.go`
- `internal/message/builder_response_test.go`
- `internal/message/parser_test.go`
- `internal/message/message_test.go`

---

#### RFC6762-§10.5-REQ-129 ✅

- **Type**: SHOULD NOT
- **Priority**: P1
- **Status**: COMPLETE

**Requirement**:
> The host SHOULD NOT perform its own queries to reconfirm that the record is truly gone.

**Implementation**:
- `responder/responder.go`
- `internal/message/message.go`
- `internal/message/builder.go`

**Tests**:
- `responder/responder_test.go`
- `internal/responder/response_builder_test.go`
- `internal/message/builder_test.go`

---

