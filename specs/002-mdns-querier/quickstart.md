# Quickstart Guide: Basic mDNS Querier (M1)

**Feature**: 002-mdns-querier
**Target Audience**: Developers implementing or using Beacon M1
**Date**: 2025-11-01

---

## What is M1 Basic mDNS Querier?

M1 is the first milestone of Beacon: a **one-shot mDNS querier** for discovering hosts and services on the local network without requiring DNS servers or manual IP configuration.

**Use Cases**:
- Discover printers, scanners, smart home devices on .local domain
- Find network services (HTTP servers, SSH, printer services)
- Resolve hostnames to IP addresses for local network communication

**What M1 Does** ✅:
- Query for A (IPv4 address), PTR (service instance), SRV (service location), TXT (metadata) records
- Send one query, collect all responses within timeout (default: 1 second)
- Parse DNS responses with full RFC 6762 compliance
- Handle errors gracefully (network failures, malformed packets)

**What M1 Doesn't Do** ❌ (Deferred to M2+):
- Response caching (M2: Cache Manager)
- Continuous service browsing (M2: Service Browser)
- Probing/announcing (Responder - separate feature)
- IPv6 support (M4: Dual-stack)
- Multi-interface queries (M4: Multi-homing)

---

## Getting Started (User Perspective)

### Prerequisites

**Platform**: Linux (M1 only)
**Go Version**: 1.21 or later
**Permissions**: Root privileges or `CAP_NET_RAW` capability (required to bind to port 5353)

**Grant Capability** (alternative to running as root):
```bash
# Build your application
go build -o myapp main.go

# Grant CAP_NET_RAW capability
sudo setcap cap_net_raw+ep ./myapp

# Now you can run without sudo
./myapp
```

### Installation

```bash
# Add Beacon as dependency to your project
go get github.com/joshuafuller/beacon/querier

# Import in your code
import "github.com/joshuafuller/beacon/querier"
```

### Hello World: Discover a Printer

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/joshuafuller/beacon/querier"
)

func main() {
    // Create querier
    q, err := querier.New()
    if err != nil {
        log.Fatalf("Failed to create querier: %v", err)
    }
    defer q.Close()  // Always close when done

    // Query for printer.local
    ctx := context.Background()
    resp, err := q.Query(ctx, "printer.local", querier.RecordTypeA)
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    // Print results
    if len(resp.Records) == 0 {
        fmt.Println("Printer not found")
    } else {
        for _, record := range resp.Records {
            if ip, err := record.AsA(); err == nil {
                fmt.Printf("Printer found at %s\n", ip)
            }
        }
    }
}
```

**Run**:
```bash
# With root
sudo go run main.go

# Or with capability
go build -o discover-printer main.go
sudo setcap cap_net_raw+ep ./discover-printer
./discover-printer
```

**Expected Output**:
```
Printer found at 192.168.1.100
```

---

## Common Use Cases

### 1. Discover All HTTP Services (Service Discovery)

```go
func discoverHTTPServices(q *querier.Querier) error {
    ctx := context.Background()

    // Step 1: Query for all _http._tcp.local service instances (PTR records)
    resp, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
    if err != nil {
        return fmt.Errorf("PTR query failed: %w", err)
    }

    if len(resp.Records) == 0 {
        fmt.Println("No HTTP services found")
        return nil
    }

    fmt.Printf("Found %d HTTP service(s):\n", len(resp.Records))

    // Step 2: For each service instance, query SRV and TXT records
    for _, record := range resp.Records {
        instanceName, err := record.AsPTR()
        if err != nil {
            continue
        }

        fmt.Printf("\nService: %s\n", instanceName)

        // Query SRV record for connection details
        srvResp, err := q.Query(ctx, instanceName, querier.RecordTypeSRV)
        if err == nil && len(srvResp.Records) > 0 {
            if srv, err := srvResp.Records[0].AsSRV(); err == nil {
                fmt.Printf("  Host: %s\n", srv.Target)
                fmt.Printf("  Port: %d\n", srv.Port)
            }
        }

        // Query TXT record for metadata
        txtResp, err := q.Query(ctx, instanceName, querier.RecordTypeTXT)
        if err == nil && len(txtResp.Records) > 0 {
            if txt, err := txtResp.Records[0].AsTXT(); err == nil {
                fmt.Printf("  Metadata: %v\n", txt)
            }
        }
    }

    return nil
}
```

**Example Output**:
```
Found 2 HTTP service(s):

Service: My Web Server._http._tcp.local
  Host: server.local
  Port: 8080
  Metadata: [version=1.0 path=/api]

Service: Home Assistant._http._tcp.local
  Host: homeassistant.local
  Port: 8123
  Metadata: [version=2023.12]
```

### 2. Scan Network for All Devices (Host Discovery)

```go
func scanNetwork(q *querier.Querier, hostnames []string) map[string]net.IP {
    ctx := context.Background()
    devices := make(map[string]net.IP)

    for _, hostname := range hostnames {
        resp, err := q.Query(ctx, hostname, querier.RecordTypeA)
        if err != nil {
            continue  // Skip errors, keep scanning
        }

        for _, record := range resp.Records {
            if ip, err := record.AsA(); err == nil {
                devices[hostname] = ip
            }
        }
    }

    return devices
}

// Usage
hostnames := []string{
    "printer.local",
    "scanner.local",
    "nas.local",
    "homeassistant.local",
}
devices := scanNetwork(q, hostnames)

for hostname, ip := range devices {
    fmt.Printf("%s -> %s\n", hostname, ip)
}
```

### 3. Custom Timeout for Slow Networks

```go
// Create querier with 2-second timeout (default is 1 second)
q, err := querier.New(querier.WithTimeout(2 * time.Second))
if err != nil {
    log.Fatal(err)
}
defer q.Close()

// Or override per-query using context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := q.Query(ctx, "slow-device.local", querier.RecordTypeA)
```

### 4. Handle Errors Gracefully

```go
resp, err := q.Query(ctx, "device.local", querier.RecordTypeA)
if err != nil {
    // Type-assert error for specific handling
    switch e := err.(type) {
    case *querier.NetworkError:
        fmt.Printf("Network error: %s\n", e)
        fmt.Println("Troubleshooting:")
        fmt.Println("  - Check if network interface is up: ip link")
        fmt.Println("  - Check multicast route: ip route | grep 224.0.0.0")
        fmt.Println("  - Check permissions: run with sudo or grant CAP_NET_RAW")

    case *querier.ValidationError:
        fmt.Printf("Invalid input: %s\n", e)
        fmt.Println("Troubleshooting:")
        fmt.Println("  - Ensure name is not empty")
        fmt.Println("  - Ensure name is ≤255 bytes")
        fmt.Println("  - Ensure record type is A, PTR, SRV, or TXT")

    default:
        fmt.Printf("Unknown error: %s\n", err)
    }
    return
}

// Check if no responders (not an error, just no devices found)
if len(resp.Records) == 0 {
    fmt.Println("No devices responded (they may not exist or timeout too short)")
}
```

---

## Developer Onboarding (Implementation Perspective)

### Repository Structure

```
beacon/
├── querier/               # Public API package (YOU WORK HERE)
│   ├── querier.go         # Main Querier implementation
│   ├── options.go         # Functional options (WithTimeout, etc.)
│   ├── records.go         # ResourceRecord, RecordType, SRVData types
│   ├── querier_test.go    # Public API tests
│   └── doc.go             # Package documentation
│
├── internal/              # Internal implementation (NOT importable by users)
│   ├── message/           # DNS message parsing (RFC 1035 wire format)
│   │   ├── message.go
│   │   ├── parser.go
│   │   ├── builder.go
│   │   └── name.go        # DNS name compression (RFC 1035 §4.1.4)
│   │
│   ├── protocol/          # mDNS protocol (RFC 6762 compliance)
│   │   ├── mdns.go        # Constants (port 5353, multicast 224.0.0.251)
│   │   ├── validator.go   # Name and response validation
│   │   └── types.go       # Record type enums
│   │
│   ├── network/           # Network I/O (UDP multicast)
│   │   ├── socket.go
│   │   ├── sender.go
│   │   └── receiver.go
│   │
│   └── errors/            # Error types (NetworkError, ValidationError, WireFormatError)
│       └── errors.go
│
└── tests/                 # Integration and contract tests
    ├── integration/       # Full query/response cycle tests
    ├── contract/          # API contract tests
    └── fuzz/              # Fuzz testing for parser
```

### Development Workflow (TDD per F-8)

**Step 1: Write Test First** (RED phase)
```go
// querier/querier_test.go
func TestQuery_EmptyName_ReturnsValidationError(t *testing.T) {
    q, err := New()
    if err != nil {
        t.Fatal(err)
    }
    defer q.Close()

    ctx := context.Background()
    _, err = q.Query(ctx, "", RecordTypeA)  // Empty name

    if err == nil {
        t.Fatal("expected error, got nil")
    }

    var valErr *ValidationError
    if !errors.As(err, &valErr) {
        t.Fatalf("expected ValidationError, got %T", err)
    }

    if !strings.Contains(err.Error(), "name cannot be empty") {
        t.Errorf("unexpected error message: %s", err)
    }
}
```

**Step 2: Run Test (verify it fails)**
```bash
go test ./querier -v -run TestQuery_EmptyName
# Expected: FAIL (because Query() doesn't validate yet)
```

**Step 3: Implement Minimal Code** (GREEN phase)
```go
// querier/querier.go
func (q *Querier) Query(ctx context.Context, name string, recordType RecordType) (*Response, error) {
    // Validate name (FR-003, FR-014)
    if name == "" {
        return nil, &ValidationError{message: "name cannot be empty", field: "name"}
    }

    // ... rest of implementation
}
```

**Step 4: Run Test (verify it passes)**
```bash
go test ./querier -v -run TestQuery_EmptyName
# Expected: PASS
```

**Step 5: Refactor** (if needed, then re-run tests)

**Step 6: Repeat** for each FR requirement (FR-001 through FR-022)

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector (mandatory per FR-019)
go test -race ./...

# Run with coverage (≥80% required per SC-010)
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run fuzz tests (10,000 iterations per NFR-003)
go test -fuzz=FuzzMessageParser -fuzztime=10000x ./tests/fuzz
```

### Key Implementation Notes

**RFC Compliance** (Constitution Principle I):
- Every MUST requirement in spec.md maps to an FR requirement
- Every FR requirement maps to a test case
- No RFC deviations without explicit documentation

**Package Visibility** (per F-2):
- `beacon/querier/`: Public API (importable by users)
- `internal/`: Implementation details (not importable externally)

**Error Handling** (per F-3):
- Use `NetworkError` for socket/I/O failures (FR-013)
- Use `ValidationError` for invalid inputs (FR-014)
- Use `WireFormatError` for malformed packets (FR-015)
- Always include actionable context in error messages (NFR-006)

**Concurrency** (per F-4):
- Use `context.Context` for all I/O operations (FR-008)
- Track goroutines with `sync.WaitGroup` (FR-017)
- Clean up resources on context cancellation (FR-018)

**Resource Management** (per F-7):
- Close all sockets in `Close()` (FR-017)
- Wait for goroutines to terminate (FR-018)
- Zero leaks verified with `go test -race` (FR-019)

---

## Troubleshooting

### "permission denied" when binding to port 5353

**Cause**: Port 5353 requires root privileges or CAP_NET_RAW capability.

**Solutions**:
1. Run with sudo: `sudo ./myapp`
2. Grant capability: `sudo setcap cap_net_raw+ep ./myapp`
3. Use Docker with `--cap-add=NET_RAW`

### "no network interfaces available"

**Cause**: System has no active network interfaces.

**Solutions**:
1. Check interfaces: `ip link`
2. Bring up interface: `sudo ip link set eth0 up`
3. Check multicast route: `ip route | grep 224.0.0.0`

### "multicast not supported on interface"

**Cause**: Default network interface doesn't support multicast.

**Solutions**:
1. Enable multicast: `sudo ip link set eth0 multicast on`
2. Add multicast route: `sudo ip route add 224.0.0.0/4 dev eth0`

### Query times out with no results

**Possible Causes**:
- No devices on network responding to the queried name
- Firewall blocking multicast traffic (UDP 5353)
- Timeout too short for slow network

**Solutions**:
1. Test with known device (e.g., avahi-browse on Linux)
2. Check firewall: `sudo iptables -L | grep 5353`
3. Increase timeout: `querier.New(querier.WithTimeout(2*time.Second))`
4. Verify multicast traffic: `sudo tcpdump -i any port 5353`

### Tests fail with race detector

**Cause**: Data race detected (violates FR-019).

**Solution**: Fix race condition using proper synchronization:
- Use `sync.Mutex` for shared data
- Use channels for goroutine communication
- Follow F-4 concurrency patterns

---

## Next Steps

**For Users**:
1. ✅ Read this quickstart
2. ✅ Try "Hello World" example
3. ✅ Explore common use cases
4. 📖 Read API contract: `contracts/querier-api.md`
5. 🚀 Integrate into your application

**For Developers**:
1. ✅ Read this quickstart
2. ✅ Understand repository structure
3. ✅ Set up development environment
4. 📖 Read F-series architecture specs (F-2, F-3, F-4, F-5, F-7, F-8)
5. 📖 Read RFC 6762 (mDNS) and RFC 1035 (DNS wire format)
6. 📖 Review research.md and data-model.md
7. 🔨 Follow TDD workflow (RED → GREEN → REFACTOR)
8. ✅ Run `/speckit.tasks` to generate implementation tasks
9. 🚀 Implement M1 per tasks.md

**Resources**:
- Spec: `specs/002-mdns-querier/spec.md`
- Plan: `specs/002-mdns-querier/plan.md`
- Research: `specs/002-mdns-querier/research.md`
- Data Model: `specs/002-mdns-querier/data-model.md`
- API Contract: `specs/002-mdns-querier/contracts/querier-api.md`
- RFC 6762: `RFC%20Docs/RFC-6762-Multicast-DNS.txt`
- Constitution: `.specify/memory/constitution.md`

---

**Quickstart Status**: ✅ **COMPLETE**
**Ready for**: `/speckit.tasks` to generate implementation tasks
