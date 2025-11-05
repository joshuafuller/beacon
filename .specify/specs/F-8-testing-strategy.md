# F-8: Testing Strategy

**Spec ID**: F-8
**Type**: Architecture
**Status**: RFC Validated (2025-11-01)
**Dependencies**: F-2 (Package Structure), F-3 (Error Handling), F-4 (Concurrency Model)
**References**: BEACON_FOUNDATIONS v1.1, Beacon Constitution v1.0.0
**RFC Compliance**: Validated against RFC 6762 (mDNS) and RFC 6763 (DNS-SD) on 2025-11-01

**Revision Notes**:
- **2025-11-01**:
  - Updated to align with Beacon Constitution v1.0.0 (ratified 2025-11-01)
  - Updated references to BEACON_FOUNDATIONS v1.1
  - Validated against RFC 6762 (Multicast DNS) and RFC 6763 (DNS-SD)
  - Added RFC compliance testing section with comprehensive test matrix
  - Added protocol test categories (probing, announcing, truncation, etc.)
  - Added interoperability testing strategy (Avahi, Bonjour)
  - Updated build tags to Go 1.17+ format (`//go:build`)
  - Formalized constitutional compliance requirements (TDD, coverage ≥80%, race detection)

---

## Overview

This specification defines Beacon's comprehensive testing strategy, aligned with the Beacon Constitution v1.0.0's commitment to Test-Driven Development (TDD). Testing ensures correctness, prevents regressions, enables confident refactoring, and validates RFC compliance.

**Constitutional Alignment**: This specification implements Constitution Principle III (Test-Driven Development) and supports Principle I (RFC Compliance) through comprehensive RFC validation testing.

---

## Requirements

### REQ-F8-1: Test-Driven Development
All code MUST be developed using the TDD cycle: RED → GREEN → REFACTOR.

**Rationale**: Per Beacon Constitution v1.0.0 Principle III, TDD is a non-negotiable principle. Tests catch bugs early, enable confident refactoring, and serve as executable documentation.

### REQ-F8-2: Test Coverage
Code coverage MUST be ≥80% for unit and integration tests.

**Rationale**: Per Beacon Constitution v1.0.0 Principle III, coverage ≥80% is mandatory. High coverage catches bugs and ensures thorough testing of all code paths.

### REQ-F8-3: No Tests Without Specs
Tests MUST be written from specifications, not from implementation.

**Rationale**: Per Beacon Constitution v1.0.0 Principle II (Spec-Driven Development), specifications MUST define acceptance tests before implementation. Tests verify requirements, not implementation details.

### REQ-F8-4: Tests Pass Before Merge
All tests MUST pass before code is merged.

**Rationale**: Per Beacon Constitution v1.0.0 Principle III enforcement, no merge without passing tests. Main branch must always be in working state.

### REQ-F8-5: Race Detector
All tests MUST pass with `-race` flag.

**Rationale**: Per Beacon Constitution v1.0.0 Principle III enforcement, all tests MUST pass with `-race` flag to detect data races. See F-4 Concurrency Model for concurrency requirements.

### REQ-F8-6: RFC Compliance Testing
All protocol behavior MUST be validated against RFC 6762 (mDNS) and RFC 6763 (DNS-SD) requirements.

**Rationale**: Per Beacon Constitution v1.0.0 Principle I (RFC Compliant), strict adherence to RFCs is non-negotiable. Tests MUST verify MUST requirements from authoritative RFCs to ensure interoperability.

---

## Test Organization

### Test Levels

```
Unit Tests         → Test individual functions/methods in isolation
Integration Tests  → Test component interactions
System Tests       → Test end-to-end scenarios
Interop Tests      → Test with external implementations (Avahi, Bonjour)
```

### Directory Structure

```
beacon/
├── querier/
│   ├── querier.go
│   ├── querier_test.go          # Unit tests (package querier_test)
│   └── internal_test.go         # Internal tests (package querier)
├── internal/
│   ├── message/
│   │   ├── message.go
│   │   └── message_test.go      # Unit tests
│   └── protocol/
│       ├── protocol.go
│       └── protocol_test.go     # Unit tests
└── integration/
    ├── query_test.go            # Integration tests
    ├── response_test.go
    └── discovery_test.go
```

### Test Package Naming

**External tests** (`package foo_test`):
- Test public API from user perspective
- Black-box testing
- Cannot access internal/private members

```go
package querier_test

import (
    "testing"
    "github.com/joshuafuller/beacon/querier"
)

func TestQuery(t *testing.T) {
    q, _ := querier.New()
    // Test public API
}
```

**Internal tests** (`package foo`):
- Test internal functions
- White-box testing
- Access to private members

```go
package querier

import "testing"

func TestInternalHelper(t *testing.T) {
    // Test internal function
    result := internalHelper()
}
```

---

## TDD Workflow

### RED → GREEN → REFACTOR Cycle

**Phase 1: RED (Write Failing Test)**
```go
// Step 1: Write test from spec
func TestQueryReturnsRecords(t *testing.T) {
    q, _ := querier.New()
    records, err := q.Query(context.Background(), "myhost.local", TypeA)

    if err != nil {
        t.Fatalf("Query() error = %v", err)
    }

    if len(records) == 0 {
        t.Error("expected records, got none")
    }
}

// Step 2: Run test - FAILS (not implemented yet)
// $ go test ./querier
// FAIL: TestQueryReturnsRecords
```

**Phase 2: GREEN (Make Test Pass)**
```go
// Step 3: Write minimum code to pass
func (q *Querier) Query(ctx context.Context, name string, qtype RecordType) ([]Record, error) {
    // Minimal implementation
    return []Record{{Name: name, Type: qtype}}, nil
}

// Step 4: Run test - PASSES
// $ go test ./querier
// PASS
```

**Phase 3: REFACTOR (Improve Code)**
```go
// Step 5: Refactor for quality
func (q *Querier) Query(ctx context.Context, name string, qtype RecordType) ([]Record, error) {
    query := buildQuery(name, qtype)
    response, err := q.transport.Send(query)
    if err != nil {
        return nil, err
    }
    return parseRecords(response), nil
}

// Step 6: Run test - STILL PASSES
// $ go test ./querier
// PASS
```

**Repeat**: Write next test, repeat cycle.

### Commit Discipline

- **RED**: Don't commit failing tests (work in progress)
- **GREEN**: Commit when tests pass
- **REFACTOR**: Commit refactoring separately from new features

---

## Unit Testing

### Test Structure

Use table-driven tests:

```go
func TestValidateDomainName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid name",
            input:   "myhost.local",
            wantErr: false,
        },
        {
            name:    "empty name",
            input:   "",
            wantErr: true,
        },
        {
            name:    "name too long",
            input:   strings.Repeat("a", 256),
            wantErr: true,
        },
        {
            name:    "label too long",
            input:   strings.Repeat("a", 64) + ".local",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateDomainName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateDomainName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Benefits**:
- Easy to add test cases
- Clear test names
- Parallel execution possible (`t.Parallel()`)

### Testing Error Handling

```go
func TestQueryError(t *testing.T) {
    q, _ := querier.New(querier.WithTransport(&mockTransportErr{}))

    _, err := q.Query(context.Background(), "test.local", TypeA)

    if err == nil {
        t.Fatal("expected error, got nil")
    }

    var netErr *NetworkError
    if !errors.As(err, &netErr) {
        t.Errorf("expected NetworkError, got %T", err)
    }
}
```

### Testing Concurrency

```go
func TestConcurrentQueries(t *testing.T) {
    q, _ := querier.New()

    var wg sync.WaitGroup
    errors := make(chan error, 100)

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            _, err := q.Query(context.Background(), fmt.Sprintf("host-%d.local", id), TypeA)
            if err != nil {
                errors <- err
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Errorf("concurrent query error: %v", err)
    }
}
```

### Testing Time-Dependent Code

Use time mocking:

```go
type Clock interface {
    Now() time.Time
    Sleep(d time.Duration)
}

type realClock struct{}

func (c *realClock) Now() time.Time {
    return time.Now()
}

func (c *realClock) Sleep(d time.Duration) {
    time.Sleep(d)
}

// Test clock
type testClock struct {
    now time.Time
}

func (c *testClock) Now() time.Time {
    return c.now
}

func (c *testClock) Sleep(d time.Duration) {
    c.now = c.now.Add(d)
}

// Usage in tests
func TestCacheExpiry(t *testing.T) {
    clock := &testClock{now: time.Now()}
    cache := NewCache(WithClock(clock))

    cache.Add(Record{Name: "test.local", TTL: 10})

    // Advance time
    clock.now = clock.now.Add(11 * time.Second)

    _, found := cache.Get("test.local")
    if found {
        t.Error("expected record to be expired")
    }
}
```

---

## Mocking

### Interface-Based Mocking

Define interfaces for dependencies:

```go
type Transport interface {
    Send(msg []byte, addr string) error
    Receive() ([]byte, net.Addr, error)
    Close() error
}

// Mock implementation
type mockTransport struct {
    sendFunc    func([]byte, string) error
    receiveFunc func() ([]byte, net.Addr, error)
}

func (m *mockTransport) Send(msg []byte, addr string) error {
    if m.sendFunc != nil {
        return m.sendFunc(msg, addr)
    }
    return nil
}

func (m *mockTransport) Receive() ([]byte, net.Addr, error) {
    if m.receiveFunc != nil {
        return m.receiveFunc()
    }
    return nil, nil, errors.New("no data")
}

// Usage in tests
func TestQueryUsesMock(t *testing.T) {
    mock := &mockTransport{
        sendFunc: func(msg []byte, addr string) error {
            // Verify send was called correctly
            if addr != "224.0.0.251:5353" {
                t.Errorf("wrong addr: %s", addr)
            }
            return nil
        },
    }

    q, _ := querier.New(querier.WithTransport(mock))
    q.Query(context.Background(), "test.local", TypeA)
}
```

### Test Helpers

Create reusable test helpers:

```go
// testutil/helpers.go
package testutil

func MustParseDomainName(t *testing.T, name string) *DomainName {
    t.Helper()
    dn, err := ParseDomainName(name)
    if err != nil {
        t.Fatalf("ParseDomainName(%q) error: %v", name, err)
    }
    return dn
}

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func AssertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

---

## Integration Testing

### Integration Test Organization

```go
// integration/query_response_test.go
//go:build integration

package integration

import (
    "testing"
    "github.com/joshuafuller/beacon/querier"
    "github.com/joshuafuller/beacon/responder"
)

func TestQueryResponse(t *testing.T) {
    // Setup responder
    r, _ := responder.New(responder.WithHostName("testhost.local"))
    r.AddRecord(responder.A("testhost.local", net.ParseIP("192.168.1.100")))
    r.Start(context.Background())
    defer r.Stop()

    // Setup querier
    q, _ := querier.New()

    // Query
    records, err := q.Query(context.Background(), "testhost.local", TypeA)

    // Verify
    if err != nil {
        t.Fatalf("Query error: %v", err)
    }

    if len(records) != 1 {
        t.Fatalf("expected 1 record, got %d", len(records))
    }

    if records[0].Data != "192.168.1.100" {
        t.Errorf("wrong IP: %s", records[0].Data)
    }
}
```

### Running Integration Tests

```bash
# Unit tests only
go test ./...

# Integration tests
go test -tags=integration ./integration/...

# All tests
go test -tags=integration ./...
```

---

## Test Data

### Test Fixtures

```go
// testdata/fixtures.go
package testdata

var (
    ValidQuery = []byte{
        // DNS message bytes
        0x00, 0x00, // ID
        0x00, 0x00, // Flags
        0x00, 0x01, // Questions: 1
        // ...
    }

    ValidResponse = []byte{
        // DNS response bytes
        // ...
    }
)

func LoadTestPacket(t *testing.T, filename string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    if err != nil {
        t.Fatalf("loading test packet: %v", err)
    }
    return data
}
```

### Test Data Files

```
beacon/
└── testdata/
    ├── query-a-record.bin
    ├── response-a-record.bin
    ├── query-ptr.bin
    └── malformed-packet.bin
```

---

## Benchmarking

### Benchmark Tests

```go
func BenchmarkParseMessage(b *testing.B) {
    data := testdata.ValidQuery

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := ParseMessage(data)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkBuildMessage(b *testing.B) {
    msg := &Message{
        Header: Header{Questions: 1},
        Questions: []Question{
            {Name: "myhost.local", Type: TypeA, Class: ClassIN},
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := msg.Serialize()
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Performance Regression Testing

```bash
# Baseline
go test -bench=. -benchmem > old.txt

# After changes
go test -bench=. -benchmem > new.txt

# Compare
benchstat old.txt new.txt
```

---

## Test Coverage

### Measuring Coverage

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage (terminal)
go tool cover -func=coverage.out

# View coverage (HTML)
go tool cover -html=coverage.out

# Coverage by package
go test -coverprofile=coverage.out ./... && \
    go tool cover -func=coverage.out | grep total
```

### Coverage Requirements

- **Unit tests**: ≥80% coverage per package
- **Integration tests**: ≥70% end-to-end coverage
- **Critical paths**: 100% coverage (error handling, parsing, protocol)

### Coverage Exclusions

Some code may be excluded from coverage requirements:
- Generated code
- Test helpers
- Debug/development-only code

Mark with comment:
```go
// coverage:ignore
func debugDump() {
    // ...
}
```

---

## Continuous Integration

### CI Pipeline

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi

      - name: Integration tests
        run: go test -tags=integration -race ./integration/...

      - name: Lint
        run: |
          go install golang.org/x/lint/golint@latest
          golint -set_exit_status ./...

      - name: Vet
        run: go vet ./...

      - name: Static check
        run: |
          go install honnef.co/go/tools/cmd/staticcheck@latest
          staticcheck ./...
```

---

## Testing Anti-Patterns

### Don't: Test Implementation Details

```go
// BAD - tests internal structure
func TestCacheUsesMap(t *testing.T) {
    c := NewCache()
    if c.records == nil {
        t.Error("cache doesn't use map")
    }
}

// GOOD - tests behavior
func TestCacheStoresAndRetrieves(t *testing.T) {
    c := NewCache()
    c.Add(record)
    got, found := c.Get(record.Name)
    if !found || got != record {
        t.Error("cache didn't store record")
    }
}
```

### Don't: Flaky Tests

```go
// BAD - timing-dependent, flaky
func TestTimeout(t *testing.T) {
    start := time.Now()
    _, err := queryWithTimeout(100 * time.Millisecond)
    duration := time.Since(start)

    if duration < 100*time.Millisecond {
        t.Error("timeout too short")
    }
}

// GOOD - use mock clock or allow tolerance
func TestTimeout(t *testing.T) {
    clock := &testClock{now: time.Now()}
    q := NewQueryWithClock(clock)

    done := make(chan error)
    go func() {
        _, err := q.Query(ctx, "test.local")
        done <- err
    }()

    clock.now = clock.now.Add(200 * time.Millisecond)

    err := <-done
    if !errors.Is(err, ErrTimeout) {
        t.Errorf("expected timeout, got %v", err)
    }
}
```

### Don't: Tests That Don't Test

```go
// BAD - always passes
func TestSomething(t *testing.T) {
    result := doSomething()
    // No assertions!
}

// GOOD - verifies behavior
func TestSomething(t *testing.T) {
    result := doSomething()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

---

## RFC Compliance Testing

**Constitutional Mandate**: Per Beacon Constitution v1.0.0 Principle I (RFC Compliant), strict adherence to RFC 6762 (mDNS) and RFC 6763 (DNS-SD) is non-negotiable. RFC compliance is the foundation of interoperability and enterprise-grade quality.

RFC compliance testing verifies that Beacon correctly implements all MUST requirements from the authoritative RFCs. This section was validated against the full text of RFC 6762 and RFC 6763 on 2025-11-01.

### RFC Compliance Test Matrix

| RFC Section | Requirement | Test Category | Test Name Pattern |
|-------------|-------------|---------------|-------------------|
| **RFC 6762 §6** | Response delay 20-120ms (shared records) | Protocol | `TestResponseDelaySharedRecords` |
| **RFC 6762 §6** | No delay for unique records | Protocol | `TestResponseNoDelayUniqueRecords` |
| **RFC 6762 §6** | Rate limiting (1s minimum) | Protocol | `TestQueryRateLimiting` |
| **RFC 6762 §7.2** | TC bit delay 400-500ms | Protocol | `TestTCBitDelay` |
| **RFC 6762 §7.2** | Requery without known-answers on TC | Protocol | `TestRequeryOnTruncation` |
| **RFC 6762 §8.1** | Probe count MUST be 3 | Protocol | `TestProbeCount` |
| **RFC 6762 §8.1** | Probe interval MUST be 250ms | Protocol | `TestProbeInterval` |
| **RFC 6762 §8.1** | Conflict detection during probing | Protocol | `TestProbeConflictDetection` |
| **RFC 6762 §8.2** | Simultaneous probe tiebreaking | Protocol | `TestSimultaneousProbe` |
| **RFC 6762 §8.3** | Announce count ≥ 2 | Protocol | `TestAnnounceCount` |
| **RFC 6762 §8.3** | Announce interval ≥ 1s | Protocol | `TestAnnounceInterval` |
| **RFC 6762 §10** | Host TTL 120s | Validation | `TestHostTTL` |
| **RFC 6762 §10** | Service TTL 4500s | Validation | `TestServiceTTL` |
| **RFC 6762 §17** | Message size ≤ 9000 bytes | Validation | `TestMaxMessageSize` |
| **RFC 6762 §18** | Name compression loop detection | Security | `TestNameCompressionLoop` |
| **RFC 6762 §18** | Label length ≤ 63 bytes | Validation | `TestLabelLength` |
| **RFC 6762 §18** | Domain name ≤ 255 bytes | Validation | `TestDomainNameLength` |
| **RFC 6763 §6** | TXT record ≤ 1300 bytes | Validation | `TestTXTRecordMaxSize` |
| **RFC 6763 §6** | TXT pair ≤ 255 bytes | Validation | `TestTXTPairLength` |
| **RFC 6763 §6** | TXT empty forms (3 valid forms) | Validation | `TestTXTEmptyForms` |
| **RFC 6763 §7** | Service name ≤ 15 characters | Validation | `TestServiceNameLength` |
| **RFC 6763 §7** | Service type format `_svc._tcp` | Validation | `TestServiceTypeFormat` |

### Protocol Test Categories

#### 1. Probing Tests (RFC 6762 §8.1)

```go
//go:build integration

func TestProbeSequence(t *testing.T) {
    t.Run("sends exactly 3 probes", func(t *testing.T) {
        r := responder.New()
        probes := captureProbes(t, r, "testhost.local")

        if len(probes) != 3 {
            t.Errorf("expected 3 probes, got %d", len(probes))
        }
    })

    t.Run("250ms intervals between probes", func(t *testing.T) {
        r := responder.New()
        times := captureProbeTimestamps(t, r, "testhost.local")

        for i := 1; i < len(times); i++ {
            interval := times[i].Sub(times[i-1])
            // Allow small tolerance (±10ms) for scheduling
            if interval < 240*time.Millisecond || interval > 260*time.Millisecond {
                t.Errorf("probe %d interval %v not ~250ms", i, interval)
            }
        }
    })

    t.Run("detects conflict during probing", func(t *testing.T) {
        // Setup conflicting responder
        r1 := responder.New(responder.WithHostName("testhost.local"))
        r1.Start(ctx)
        defer r1.Stop()

        // Try to probe same name
        r2 := responder.New(responder.WithHostName("testhost.local"))
        err := r2.Start(ctx)

        var conflictErr *ConflictError
        if !errors.As(err, &conflictErr) {
            t.Errorf("expected ConflictError, got %T: %v", err, err)
        }
    })
}
```

#### 2. Announcing Tests (RFC 6762 §8.3)

```go
func TestAnnounceSequence(t *testing.T) {
    t.Run("sends minimum 2 announcements", func(t *testing.T) {
        r := responder.New()
        r.AddRecord(responder.A("testhost.local", net.ParseIP("192.168.1.1")))

        announcements := captureAnnouncements(t, r, "testhost.local")

        if len(announcements) < 2 {
            t.Errorf("expected ≥2 announcements, got %d", len(announcements))
        }
    })

    t.Run("1 second minimum interval", func(t *testing.T) {
        r := responder.New()
        r.AddRecord(responder.A("testhost.local", net.ParseIP("192.168.1.1")))

        times := captureAnnouncementTimestamps(t, r, "testhost.local")

        for i := 1; i < len(times); i++ {
            interval := times[i].Sub(times[i-1])
            if interval < 1*time.Second {
                t.Errorf("announcement %d interval %v < 1s", i, interval)
            }
        }
    })
}
```

#### 3. Response Delay Tests (RFC 6762 §6)

```go
func TestResponseDelays(t *testing.T) {
    t.Run("shared records delayed 20-120ms", func(t *testing.T) {
        r := responder.New()
        r.AddRecord(responder.PTR("_services._dns-sd._udp.local", "_http._tcp.local"))

        queryTime := time.Now()
        sendQuery(t, "_services._dns-sd._udp.local", TypePTR)
        responseTime := captureResponseTime(t, r)

        delay := responseTime.Sub(queryTime)
        if delay < 20*time.Millisecond || delay > 120*time.Millisecond {
            t.Errorf("response delay %v not in 20-120ms range", delay)
        }
    })

    t.Run("unique records no delay", func(t *testing.T) {
        r := responder.New()
        r.AddRecord(responder.A("testhost.local", net.ParseIP("192.168.1.1")))

        queryTime := time.Now()
        sendQuery(t, "testhost.local", TypeA)
        responseTime := captureResponseTime(t, r)

        delay := responseTime.Sub(queryTime)
        // Unique records should respond immediately (< 10ms overhead)
        if delay > 10*time.Millisecond {
            t.Errorf("unique record delay %v > 10ms (should be immediate)", delay)
        }
    })
}
```

#### 4. Truncation Tests (RFC 6762 §7.2)

```go
func TestTruncation(t *testing.T) {
    t.Run("TC bit set when message truncated", func(t *testing.T) {
        r := responder.New()
        // Add many records to force truncation
        for i := 0; i < 100; i++ {
            r.AddRecord(responder.PTR("_services._dns-sd._udp.local",
                fmt.Sprintf("_http-%d._tcp.local", i)))
        }

        response := sendQueryAndGetResponse(t, "_services._dns-sd._udp.local", TypePTR)

        if !response.Header.TC {
            t.Error("expected TC bit set for truncated response")
        }
    })

    t.Run("requery without known-answers on TC", func(t *testing.T) {
        // First query with known-answers that triggers truncation
        response1 := sendQueryWithKnownAnswers(t, "test.local", knownAnswers)

        if !response1.Header.TC {
            t.Skip("expected TC bit, can't test requery")
        }

        // Second query without known-answers
        response2 := sendQueryWithoutKnownAnswers(t, "test.local")

        if response2.Header.TC {
            t.Error("second query still truncated after removing known-answers")
        }
    })
}
```

#### 5. Rate Limiting Tests (RFC 6762 §6)

```go
func TestRateLimiting(t *testing.T) {
    t.Run("queries rate limited to 1/second", func(t *testing.T) {
        q := querier.New()

        // First query should succeed
        _, err1 := q.Query(ctx, "test.local", TypeA)
        if err1 != nil {
            t.Fatalf("first query failed: %v", err1)
        }

        // Immediate second query should be rate limited
        _, err2 := q.Query(ctx, "test.local", TypeA)

        var protocolErr *ProtocolError
        if !errors.As(err2, &protocolErr) {
            t.Errorf("expected ProtocolError for rate limit, got %T", err2)
        }
    })

    t.Run("probes NOT rate limited", func(t *testing.T) {
        r := responder.New()

        // Probes at 250ms intervals should not be rate limited
        err := r.probe(ctx, "testhost.local")
        if err != nil {
            t.Errorf("probing failed (should not be rate limited): %v", err)
        }
    })
}
```

### Interoperability Testing

**Constitutional Requirement**: Per Beacon Constitution v1.0.0 Principle VII (Excellence), interoperability testing against Avahi and Bonjour is required to validate best-in-class implementation quality.

Test against real-world implementations to ensure Beacon interoperates correctly with existing mDNS/DNS-SD ecosystems:

```go
//go:build interop

package interop

// TestAvahiDiscovery verifies Beacon can discover services published by Avahi
func TestAvahiDiscovery(t *testing.T) {
    if !avahiAvailable() {
        t.Skip("Avahi not available")
    }

    // Publish service with Avahi
    publishAvahiService(t, "_http._tcp", "testservice", 8080)
    defer unpublishAvahiService(t, "_http._tcp", "testservice")

    // Discover with Beacon
    b := browser.New()
    services := b.Browse(ctx, "_http._tcp.local")

    found := false
    for _, svc := range services {
        if svc.Instance == "testservice" && svc.Port == 8080 {
            found = true
            break
        }
    }

    if !found {
        t.Error("Beacon failed to discover Avahi-published service")
    }
}

// TestBonjourDiscovery verifies Beacon can discover services published by Bonjour
func TestBonjourDiscovery(t *testing.T) {
    if runtime.GOOS != "darwin" {
        t.Skip("Bonjour only available on macOS")
    }

    // Publish service with Bonjour (dns-sd command)
    publishBonjourService(t, "_http._tcp", "testservice", 8080)
    defer unpublishBonjourService(t)

    // Discover with Beacon
    b := browser.New()
    services := b.Browse(ctx, "_http._tcp.local")

    found := false
    for _, svc := range services {
        if svc.Instance == "testservice" && svc.Port == 8080 {
            found = true
            break
        }
    }

    if !found {
        t.Error("Beacon failed to discover Bonjour-published service")
    }
}

// TestCrossImplementationQuery tests Beacon querier with Avahi responder
func TestCrossImplementationQuery(t *testing.T) {
    if !avahiAvailable() {
        t.Skip("Avahi not available")
    }

    // Setup Avahi responder for testhost.local
    publishAvahiHost(t, "testhost", "192.168.1.100")
    defer unpublishAvahiHost(t, "testhost")

    // Query with Beacon
    q := querier.New()
    records, err := q.Query(ctx, "testhost.local", TypeA)

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    if len(records) == 0 {
        t.Fatal("expected records from Avahi responder")
    }

    if records[0].Data != "192.168.1.100" {
        t.Errorf("wrong IP: got %s, want 192.168.1.100", records[0].Data)
    }
}
```

**Running Interop Tests**:
```bash
# Requires Avahi or Bonjour installed
go test -tags=interop ./interop/...
```

### RFC Requirements Traceability

Maintain a traceability matrix from spec requirements to tests:

```go
// internal/testing/rfc_matrix.go
package testing

// RFCRequirement maps RFC requirements to test functions
type RFCRequirement struct {
    RFC         string   // e.g., "RFC 6762 §8.1"
    Requirement string   // e.g., "MUST send 3 probes at 250ms intervals"
    TestNames   []string // Test functions that verify this
}

var RFCMatrix = []RFCRequirement{
    {
        RFC:         "RFC 6762 §8.1",
        Requirement: "MUST send 3 probes at 250ms intervals",
        TestNames: []string{
            "TestProbeSequence/sends_exactly_3_probes",
            "TestProbeSequence/250ms_intervals_between_probes",
        },
    },
    {
        RFC:         "RFC 6762 §8.3",
        Requirement: "MUST send at least 2 announcements, 1s apart",
        TestNames: []string{
            "TestAnnounceSequence/sends_minimum_2_announcements",
            "TestAnnounceSequence/1_second_minimum_interval",
        },
    },
    {
        RFC:         "RFC 6762 §6",
        Requirement: "Shared records MUST delay response 20-120ms",
        TestNames: []string{
            "TestResponseDelays/shared_records_delayed_20-120ms",
        },
    },
    // ... more requirements
}

// VerifyRFCCoverage checks that all RFC requirements have tests
func VerifyRFCCoverage(t *testing.T) {
    for _, req := range RFCMatrix {
        if len(req.TestNames) == 0 {
            t.Errorf("RFC requirement not covered: %s - %s", req.RFC, req.Requirement)
        }
    }
}
```

**Generate Coverage Report**:
```bash
go test -v ./... | grep -E "RFC [0-9]+" > rfc_coverage.txt
```

---

## Test Documentation

### Documenting Test Intent

```go
// TestQueryHandlesTimeout verifies that queries return ErrTimeout
// when the context deadline is exceeded. This ensures graceful
// handling of network delays per RFC 6762 §5.
func TestQueryHandlesTimeout(t *testing.T) {
    // ...
}
```

### Test Names

Use descriptive names that explain what's being tested:

```go
// Good
func TestQueryReturnsErrorWhenNetworkUnreachable(t *testing.T)
func TestCacheEvictsExpiredRecords(t *testing.T)
func TestProbeDetectsNameConflict(t *testing.T)

// Bad
func TestQuery1(t *testing.T)
func TestCache(t *testing.T)
func TestProbe(t *testing.T)
```

---

## Open Questions

**Q1**: Should we use testify/assert for assertions?
- **Pro**: More readable assertions
- **Con**: External dependency
- **Decision**: Start with standard library, reconsider if assertions become unwieldy

**Q2**: Property-based testing (rapid)?
- **Pro**: Finds edge cases
- **Con**: Complex, slower
- **Decision**: Use for critical parsers/algorithms, not everything

**Q3**: Mutation testing?
- **Pro**: Verifies test quality
- **Con**: Slow, not standard in Go
- **Decision**: Not initially, manual review of test quality

---

## Success Criteria

This specification is considered complete when:

- [ ] TDD workflow established (RED-GREEN-REFACTOR) per Constitution Principle III
- [ ] ≥80% code coverage achieved per Constitution Principle III
- [ ] All tests pass with `-race` flag per Constitution Principle III
- [ ] Unit, integration, and system tests defined
- [ ] Mocking strategy in place
- [ ] CI pipeline running tests on all commits
- [ ] Benchmarks for critical paths
- [ ] RFC compliance tests implemented for all MUST requirements per Constitution Principle I
- [ ] Interoperability tests with Avahi and Bonjour per Constitution Principle VII
- [ ] RFC traceability matrix maintained

---

## Constitutional Compliance

This specification implements and enforces the [Beacon Constitution v1.0.0](../memory/constitution.md):

### Principle I: RFC Compliant
**Status**: ✅ **ENFORCED**

- ✅ **RFC compliance testing matrix**: REQ-F8-6 mandates RFC 6762 and RFC 6763 validation
- ✅ **All RFC MUST requirements tested**: Tests verify MUST requirements from authoritative RFCs to ensure interoperability
- ✅ **Interoperability testing**: Success criteria requires testing against Avahi and Bonjour
- ✅ **RFC traceability**: Requirements include RFC compliance testing with traceability matrix

**Evidence**:
- REQ-F8-6: "All protocol behavior MUST be validated against RFC 6762 (mDNS) and RFC 6763 (DNS-SD) requirements"
- Section "RFC Compliance Testing" (lines 761-1136) defines comprehensive test coverage for RFC MUST requirements
- RFC traceability matrix ensures no RFC requirement is untested

### Principle II: Spec-Driven Development
**Status**: ✅ **ENFORCED**

- ✅ **Tests from specs, not implementation**: REQ-F8-3 mandates "Tests MUST be written from specifications, not from implementation"
- ✅ **Specifications define acceptance tests**: Tests verify requirements, not implementation details
- ✅ **Architecture specification exists**: This F-8 specification defines testing strategy before implementation

**Evidence**:
- REQ-F8-3: "Tests MUST be written from specifications, not from implementation"
- TDD workflow (lines 136-200) ensures tests are written from user scenarios and acceptance criteria
- Acceptance tests defined in feature specifications drive implementation

### Principle III: Test-Driven Development
**Status**: ✅ **MANDATED** (NON-NEGOTIABLE)

- ✅ **TDD cycle mandatory**: REQ-F8-1 enforces RED → GREEN → REFACTOR cycle
- ✅ **Coverage ≥80% mandatory**: REQ-F8-2 requires ≥80% code coverage for unit and integration tests
- ✅ **Race detection mandatory**: REQ-F8-5 requires all tests pass with `-race` flag
- ✅ **Tests before merge**: REQ-F8-4 mandates all tests MUST pass before code is merged
- ✅ **No implementation without tests**: TDD workflow defines test-first development

**Evidence**:
- REQ-F8-1: "All code MUST be developed using the TDD cycle: RED → GREEN → REFACTOR"
- REQ-F8-2: "Code coverage MUST be ≥80% for unit and integration tests"
- REQ-F8-3: "Tests MUST be written from specifications, not from implementation"
- REQ-F8-4: "All tests MUST pass before code is merged"
- REQ-F8-5: "All tests MUST pass with `-race` flag"
- REQ-F8-6: "All protocol behavior MUST be validated against RFC 6762 and RFC 6763"

**Principle III is the foundation of this entire specification** - TDD is non-negotiable and enforced through explicit requirements and CI gates.

### Principle IV: Phased Approach
**Status**: ✅ **SUPPORTED**

- ✅ **Incremental test development**: Testing strategy supports milestone-based development
- ✅ **Test coverage grows with features**: Each milestone adds tests for new functionality
- ✅ **Regression prevention**: Tests prevent regressions as system evolves

**Evidence**: TDD workflow enables incremental development with confidence that existing functionality remains correct.

### Principle V: Open Source
**Status**: ✅ **COMPLIANT**

- ✅ **Public test suite**: All tests publicly available in repository
- ✅ **Transparent testing**: Test strategy, coverage reports, and CI results are public
- ✅ **Community contributions**: Testing patterns enable external contributors to write tests

### Principle VI: Maintained
**Status**: ✅ **SUPPORTED**

- ✅ **Regression prevention**: Comprehensive test suite prevents regressions during maintenance
- ✅ **Refactoring confidence**: High test coverage enables confident refactoring
- ✅ **Long-term stability**: Tests ensure behavior remains correct over time
- ✅ **Documentation**: Tests serve as executable documentation of expected behavior

**Evidence**: ≥80% coverage (REQ-F8-2) and comprehensive test suite provide safety net for long-term maintenance.

### Principle VII: Excellence
**Status**: ✅ **ENFORCED**

- ✅ **Comprehensive test strategy**: Unit, integration, system, and interoperability tests
- ✅ **Industry best practices**: Follows Go testing best practices (table-driven tests, test organization)
- ✅ **RFC compliance validation**: Interoperability testing against Avahi and Bonjour ensures best-in-class compatibility
- ✅ **Benchmarking**: Performance testing ensures excellence in critical paths
- ✅ **Test quality**: Anti-patterns section (lines 683-760) guides developers away from poor practices

**Evidence**:
- REQ-F8-6: RFC compliance testing ensures interoperability excellence
- Section "RFC Compliance Testing" (lines 761-1136) demonstrates commitment to excellence
- Interoperability tests validate "best enterprise-grade implementation" mission
- Benchmarking (lines 541-590) ensures performance excellence

**Overall Assessment**: This specification is the **implementation of Constitution Principle III** and enforces TDD, RFC compliance testing, and excellence across all Beacon development. Without this testing strategy, constitutional compliance would be impossible to verify.

---

## References

### Technical Sources of Truth (RFCs) - PRIMARY AUTHORITY for Protocol Compliance Testing

**Note**: RFC 6762 and RFC 6763 are the **PRIMARY TECHNICAL AUTHORITY** for Beacon. All RFC MUST requirements MUST have corresponding test coverage to ensure interoperability and protocol compliance.

**Critical**: Per Constitution Principle I, RFC requirements override all other concerns. All RFC MUST requirements from RFC 6762 and RFC 6763 MUST be validated through comprehensive test coverage (REQ-F8-6).

- [RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt) - PRIMARY AUTHORITY for mDNS protocol testing
  - **Validation Date**: 2025-11-01
  - **Test Coverage Required**: §6 (Response timing), §7.2 (Truncation), §8.1 (Probing), §8.2 (Tiebreaking), §8.3 (Announcing), §10 (TTL values), §17 (Message size), §18 (Security)
  - **Interoperability**: Tests MUST validate against Avahi and Bonjour implementations

- [RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt) - PRIMARY AUTHORITY for DNS-SD testing
  - **Validation Date**: 2025-11-01
  - **Test Coverage Required**: §4 (Browsing), §5 (Resolution), §6 (TXT records), §7 (Service names)
  - **Interoperability**: Tests MUST validate service discovery compatibility

### Project Governance

- [Beacon Constitution v1.0.0](../memory/constitution.md) - Ratified 2025-11-01
  - **Principle I**: RFC Compliant (NON-NEGOTIABLE) - All RFC MUST requirements tested
  - **Principle II**: Spec-Driven Development (NON-NEGOTIABLE) - Tests from specs, not implementation
  - **Principle III**: Test-Driven Development (NON-NEGOTIABLE) - TDD cycle mandatory, coverage ≥80%, race detection required
  - **Principle VII**: Excellence - Interoperability testing against Avahi and Bonjour

### Foundational Knowledge

- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md) - Terminology and test scenarios

### Architecture Specifications

- [F-2: Package Structure](./F-2-package-structure.md) - Test organization and package boundaries
- [F-3: Error Handling Strategy](./F-3-error-handling.md) - Error testing patterns
- [F-4: Concurrency Model](./F-4-concurrency-model.md) - Race detection requirements and timing tests

### Go Testing Resources

- [Go Testing Tutorial](https://go.dev/doc/tutorial/add-a-test) - Basic testing introduction
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests) - Go testing best practice
- [Advanced Testing with Subtests](https://go.dev/blog/subtests) - Test organization patterns
- [Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks) - Performance testing
- [Race Detector](https://go.dev/doc/articles/race_detector) - Data race detection (mandatory per REQ-F8-5)
- [Testing Comments](https://go.dev/wiki/TestComments) - Test documentation guidance

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 2.0 | 2025-11-01 | System | Updated for Constitution v1.0.0 and Foundations v1.1; added RFC validation; formalized constitutional compliance |
| 1.1 | 2025-11-01 | System | Added RFC compliance testing, interop testing, protocol test categories |
| 1.0 | 2025-10-31 | System | Initial specification |

---

**Governance**: This specification is governed by the Beacon Constitution v1.0.0. All implementations must comply with constitutional principles, particularly Principle I (RFC Compliance), Principle II (Spec-Driven Development), and Principle III (Test-Driven Development).
