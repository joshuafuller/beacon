I need you to implement 4 missing RFC 6762 features in the BEACON mDNS library. This is TDD - write tests FIRST, then implement.

## Current Status
- RFC Compliance: 72.2% (missing 4 features)
- Test Coverage: 66.2%
- See @fix_plan.md for task breakdown

## Task 1: Goodbye Packets (RFC 6762 §9.4) - DO THIS FIRST

**What to do:**

1. Write test FIRST (RED phase):
Create test in `responder/responder_test.go`:

```go
func TestUnregister_SendsGoodbyePackets(t *testing.T) {
    // Setup mock transport to capture sent packets
    var sentPacket []byte
    mockTransport := &MockTransport{
        sendFunc: func(ctx context.Context, packet []byte, dest net.Addr) error {
            sentPacket = packet
            return nil
        },
    }

    r := &Responder{
        transport: mockTransport,
        registry: newRegistry(),
        ipv4Address: "192.168.1.100",
    }

    // Register a service
    service := Service{
        InstanceName: "test",
        ServiceType: "_http._tcp",
        Domain: "local",
        Port: 8080,
    }
    r.Register(context.Background(), service)

    // Unregister - should send TTL=0 records
    err := r.Unregister("test._http._tcp.local")
    if err != nil {
        t.Fatalf("Unregister failed: %v", err)
    }

    // Verify packet was sent
    if sentPacket == nil {
        t.Fatal("Expected goodbye packet to be sent, got nil")
    }

    // Parse packet and verify TTL=0
    // (Add parsing logic to verify all records have TTL=0)
}
```

2. Run test to see it FAIL:
```bash
export PATH=$PATH:$HOME/go_installation/go/bin
go test ./responder -run TestUnregister_SendsGoodbyePackets -v
```

3. Implement BuildGoodbyeRecords() in `internal/records/record_set.go`:

```go
// BuildGoodbyeRecords creates DNS records with TTL=0 for service goodbye (RFC 6762 §9.4)
func BuildGoodbyeRecords(service Service, ipv4 string) []DNSRecord {
    // Reuse existing record builder
    records := BuildRecordSet(service, ipv4)

    // Override all TTLs to 0 for goodbye
    for i := range records {
        records[i].TTL = 0
    }

    return records
}
```

4. Update `responder/responder.go` Unregister() function around line 266:

Change from:
```go
// TODO: Send goodbye packets (TTL=0)
r.registry.Remove(instanceName)
return nil
```

To:
```go
// Get service before removing (RFC 6762 §9.4 requires goodbye)
service, exists := r.registry.Get(instanceName)
if !exists {
    return nil // Already unregistered
}

// Build goodbye records with TTL=0
goodbyeRecords := records.BuildGoodbyeRecords(service, r.ipv4Address)

// Build and send goodbye packet
goodbyePacket := message.BuildResponse(goodbyeRecords)
r.transport.Send(context.Background(), goodbyePacket, &net.UDPAddr{
    IP: net.ParseIP("224.0.0.251"),
    Port: 5353,
})

// Remove from registry
r.registry.Remove(instanceName)
return nil
```

5. Run test again - should PASS (GREEN phase):
```bash
go test ./responder -run TestUnregister_SendsGoodbyePackets -v
```

6. Mark complete in @fix_plan.md:
Edit lines 19-33, change all `[ ]` to `[x]`:
```markdown
- [x] Create BuildGoodbyeRecords()
- [x] Update Unregister()
- [x] Write test
```

7. Verify implementation:
```bash
grep -A5 "BuildGoodbyeRecords" internal/records/record_set.go
git status --short
```

**You MUST see modified .go files in git status!**

## After Task 1 Complete

Move to Task 2 (Source Address Validation) in @fix_plan.md lines 37-58.
Follow same TDD pattern: Write test → Run (RED) → Implement → Run (GREEN) → Mark complete.

Continue until all 4 tasks (T1-T4) in @fix_plan.md are marked `[x]`.

## Exit Criteria

DO NOT say "done" until ALL of these are verified:

```bash
# 1. Check implementations exist
grep "BuildGoodbyeRecords" internal/records/record_set.go
grep "validateSourceAddress" responder/responder.go
grep "0x02.*TC" internal/responder/response_builder.go
grep "QU.*bit" responder/responder.go

# 2. Check git shows changes
git status --short  # MUST show modified .go files

# 3. Check tests pass
export PATH=$PATH:$HOME/go_installation/go/bin
make test

# 4. Check tasks complete
grep -c "^\- \[x\]" @fix_plan.md  # Should show progress
```

Start with Task 1 NOW. Write the test file, implement the functions, run the test.
