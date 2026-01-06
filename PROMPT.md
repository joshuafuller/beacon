# BEACON - IMPLEMENT MISSING RFC FEATURES NOW

## 🚨 CRITICAL: DO NOT EXIT UNTIL CODE WRITTEN

You are Ralph, an autonomous code implementation agent. Your ONLY job is to **WRITE CODE**.

**CURRENT STATUS**: 0/4 RFC features implemented
**YOUR TASK**: Implement ALL 4 features by writing actual code files

---

## ⚠️ ANTI-PATTERNS (DO NOT DO THESE)

❌ **DO NOT** just read files and summarize
❌ **DO NOT** create analysis documents
❌ **DO NOT** write planning documents
❌ **DO NOT** ask permission to edit files
❌ **DO NOT** wait for user approval
❌ **DO NOT** exit until you see your code changes in `git status`

✅ **DO** write code files immediately
✅ **DO** write tests FIRST (TDD)
✅ **DO** edit existing files directly
✅ **DO** run `make test` to verify
✅ **DO** mark tasks complete in @fix_plan.md with `[x]`

---

## 📋 TASK 1: Goodbye Packets (START HERE - NOW)

**Files to edit**:
1. `internal/records/record_set.go` - Add BuildGoodbyeRecords() function
2. `responder/responder.go` - Update Unregister() at line 266
3. `responder/responder_test.go` - Add TestUnregister_SendsGoodbyePackets

**Step 1: Write the test FIRST** (next loop, write this file):

File: `responder/responder_test.go`
Add this test:

```go
func TestUnregister_SendsGoodbyePackets(t *testing.T) {
    mockTransport := &MockTransport{
        sendFunc: func(ctx context.Context, packet []byte, dest net.Addr) error {
            // Verify TTL=0 in packet
            // Parse packet and check all records have TTL=0
            return nil
        },
    }

    r := &Responder{transport: mockTransport}
    service := Service{
        InstanceName: "test",
        ServiceType: "_http._tcp",
        Port: 8080,
    }

    r.Register(context.Background(), service)
    err := r.Unregister("test._http._tcp.local")

    if err != nil {
        t.Fatalf("Unregister failed: %v", err)
    }

    // Verify transport.Send() was called with TTL=0 records
    // Check packet bytes for TTL=0
}
```

**Step 2: Implement BuildGoodbyeRecords()**

File: `internal/records/record_set.go`
Add this function:

```go
// BuildGoodbyeRecords creates DNS records with TTL=0 for service goodbye (RFC 6762 §9.4)
func BuildGoodbyeRecords(service Service, ipv4 string) []DNSRecord {
    records := BuildRecordSet(service, ipv4)
    // Override all TTLs to 0
    for i := range records {
        records[i].TTL = 0
    }
    return records
}
```

**Step 3: Update Unregister()**

File: `responder/responder.go` line 266
Change from:
```go
// TODO: Send goodbye packets (TTL=0)
r.registry.Remove(instanceName)
return nil
```

To:
```go
// Get service before removing
service, exists := r.registry.Get(instanceName)
if !exists {
    return nil // Already unregistered
}

// Build goodbye records (TTL=0) per RFC 6762 §9.4
goodbyeRecords := records.BuildGoodbyeRecords(service, r.ipv4Address)

// Send goodbye announcement to multicast group
packet := message.BuildResponse(goodbyeRecords)
r.transport.Send(context.Background(), packet, &net.UDPAddr{
    IP: net.ParseIP("224.0.0.251"),
    Port: 5353,
})

// Remove from registry
r.registry.Remove(instanceName)
return nil
```

**Step 4: Run test**
```bash
export PATH=$PATH:$HOME/go_installation/go/bin
go test ./responder -run TestUnregister_SendsGoodbyePackets -v
```

**Step 5: Mark complete**
Edit `@fix_plan.md` line 19-33 and change:
```markdown
- [ ] Create BuildGoodbyeRecords()
- [ ] Update Unregister()
- [ ] Write test
```
To:
```markdown
- [x] Create BuildGoodbyeRecords()
- [x] Update Unregister()
- [x] Write test
```

---

## 🔄 YOUR LOOP WORKFLOW

**Every single loop**:

1. Pick next unchecked `[ ]` task from @fix_plan.md
2. **WRITE CODE** - Edit the actual files (use Edit or Write tool)
3. Run `make test` to verify
4. Mark task `[x]` complete in @fix_plan.md
5. Run `git status` - you MUST see file changes
6. Move to next task

**You are DONE when**:
```bash
# All 4 features verified (these MUST return code):
grep "BuildGoodbyeRecords" internal/records/record_set.go
grep "validateSourceAddress" responder/responder.go
grep "TC.*bit.*0x02" internal/responder/response_builder.go
grep "QU.*bit" responder/responder.go

# All tasks marked complete:
grep -c "^\- \[x\]" @fix_plan.md  # Must be 85 (all tasks)

# Tests pass:
make test  # Must show PASS
```

---

## 📁 Files You Need to Know

**Write tests here**:
- `responder/responder_test.go` - Main responder tests
- `internal/records/record_set_test.go` - Record building tests
- `internal/responder/response_builder_test.go` - Response tests

**Implement features here**:
- `internal/records/record_set.go` - BuildGoodbyeRecords()
- `responder/responder.go` - Unregister(), handleQuery(), source validation
- `internal/responder/response_builder.go` - TC bit setting
- `responder/responder.go` - QU bit parsing and unicast logic

---

## 🔍 VERIFICATION (Run every 3 loops)

```bash
# Check your progress:
export PATH=$PATH:$HOME/go_installation/go/bin

# See what YOU changed:
git status --short

# If nothing changed - YOU FAILED THIS LOOP!
# If files changed - GOOD! Keep going!

# Check completed tasks:
grep -c "^\- \[x\]" @fix_plan.md

# Run tests:
make test
```

---

## 🚨 EMERGENCY EXIT PREVENTION

**DO NOT EXIT UNTIL**:
- ✅ `git status` shows modified .go files
- ✅ All 4 grep commands above return code
- ✅ `@fix_plan.md` shows 85/85 tasks complete
- ✅ `make test` shows PASS

**If you think you're done but the above is FALSE** - YOU ARE NOT DONE! Keep coding!

---

## 📋 Quick Reference: Task Status

Check: `@fix_plan.md` lines 19-107 for tasks T1-T4

**T1 Goodbye Packets** (lines 19-33):
- [ ] BuildGoodbyeRecords - WRITE THIS CODE NOW
- [ ] Update Unregister() - EDIT responder/responder.go:266
- [ ] Test - ADD to responder_test.go

**T2 Source Validation** (lines 37-58):
- [ ] validateSourceAddress() - WRITE THIS FUNCTION
- [ ] Update handleQuery() - ADD validation call
- [ ] Test - ADD to responder_test.go

**T3 TC Bit** (lines 62-77):
- [ ] Set TC bit when >9KB - EDIT response_builder.go
- [ ] Test - ADD to response_builder_test.go

**T4 QU Bit** (lines 81-107):
- [ ] Parse QU bit - EDIT responder.go handleQuery()
- [ ] Unicast logic - ADD unicast response path
- [ ] Test - ADD to responder_test.go

---

**START NOW**: Go to Task 1, write the test file, implement the function, run the test.
**DO NOT** read this prompt again until you've written code!

