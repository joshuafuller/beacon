# BEACON Production Polish - RFC Features FIRST

**⚠️ CRITICAL**: Missing RFC 6762 features (72.2% → 88.9% compliance target)

Track progress with [x] for completed items.

---

## P0: CRITICAL RFC FEATURES (Week 1) 🔥 START HERE

### T1: Goodbye Packets (RFC 6762 §9.4)
- [ ] Write test: TestUnregister_SendsGoodbyePackets
- [ ] Implement BuildGoodbyeRecords() in internal/records/record_set.go
- [ ] Update Unregister() in responder/responder.go:266
- [ ] Verify: grep "BuildGoodbyeRecords" internal/records/record_set.go

### T2: Source Address Validation (RFC 6762 §6.4)
- [ ] Write test: TestHandleQuery_RejectsWrongSubnet
- [ ] Implement validateSourceAddress() function
- [ ] Add subnet validation in handleQuery()
- [ ] Verify: grep "validateSourceAddress" responder/responder.go

### T3: TC Bit Truncation (RFC 6762 §6.5)
- [ ] Write test: TestBuildResponse_SetsTCBitWhenTruncated
- [ ] Set TC bit when response >9KB in BuildResponse()
- [ ] Verify: grep "0x02.*TC" internal/responder/response_builder.go

### T4: QU Bit + Unicast Responses (RFC 6762 §5.4)
- [ ] Write test: TestHandleQuery_QUBitUnicastResponse
- [ ] Parse QU bit from question
- [ ] Implement unicast vs multicast logic
- [ ] Verify: grep "QU.*bit" responder/responder.go

---

## Completed
- [x] M1: mDNS Querier (100% complete)
- [x] M2: mDNS Responder (94.6% complete)
- [x] 007: Interface-Specific Addressing (100% complete)

## Notes
- TDD methodology: Write test FIRST (RED), then implement (GREEN), then refactor
- All code must pass: make test && make test-race
- Verify implementations with grep commands after each task
