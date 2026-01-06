Implement Task 1 from @fix_plan.md RIGHT NOW. Do not ask for permission. Write the code files directly.

## Task 1: Goodbye Packets - IMPLEMENT NOW

Step 1: Add this test to responder/responder_test.go:

```go
func TestUnregister_SendsGoodbyePackets(t *testing.T) {
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

    service := Service{
        InstanceName: "test",
        ServiceType: "_http._tcp",
        Domain: "local",
        Port: 8080,
    }
    r.Register(context.Background(), service)

    err := r.Unregister("test._http._tcp.local")
    if err != nil {
        t.Fatalf("Unregister failed: %v", err)
    }

    if sentPacket == nil {
        t.Fatal("Expected goodbye packet to be sent, got nil")
    }
}
```

Step 2: Add this function to internal/records/record_set.go:

```go
// BuildGoodbyeRecords creates DNS records with TTL=0 for service goodbye (RFC 6762 §9.4)
func BuildGoodbyeRecords(service Service, ipv4 string) []DNSRecord {
    records := BuildRecordSet(service, ipv4)
    for i := range records {
        records[i].TTL = 0
    }
    return records
}
```

Step 3: Update Unregister() in responder/responder.go (around line 266):

Replace the TODO comment with:

```go
service, exists := r.registry.Get(instanceName)
if !exists {
    return nil
}

goodbyeRecords := records.BuildGoodbyeRecords(service, r.ipv4Address)
goodbyePacket := message.BuildResponse(goodbyeRecords)
r.transport.Send(context.Background(), goodbyePacket, &net.UDPAddr{
    IP: net.ParseIP("224.0.0.251"),
    Port: 5353,
})

r.registry.Remove(instanceName)
return nil
```

Step 4: Run the test:
```bash
export PATH=$PATH:$HOME/go_installation/go/bin
go test ./responder -run TestUnregister_SendsGoodbyePackets -v
```

Step 5: Mark complete in @fix_plan.md - change the 4 checkboxes under T1 from [ ] to [x]

After completing Task 1, move to Task 2 (Source Address Validation) and implement it the same way.

DO NOT ask "would you like me to" or "shall I proceed". Just write the code files.
