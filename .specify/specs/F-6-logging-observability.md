# F-6: Logging & Observability

**Spec ID**: F-6
**Type**: Architecture
**Status**: Validated (2025-11-01)
**Dependencies**: F-2 (Package Structure), F-5 (Configuration)
**References**: BEACON_FOUNDATIONS v1.1
**Governance**: Beacon Constitution v1.0.0
**RFC Validation**: Validated against RFC 6762 and RFC 6763 (2025-11-01)

**Revision Notes**:
- **2025-11-01**: Updated to align with Constitution v1.0.0 and Foundations v1.1
  - Added RFC validation status
  - Added governance reference to Constitution v1.0.0
  - Updated Foundations reference to v1.1
  - Added explicit hot path definition
  - Added probe/announce lifecycle event logging
  - Added timing metadata for RFC-critical operations
  - Clarified TXT record redaction policy
  - Added mDNS-specific bit logging (TC, cache-flush, QU)

---

## Overview

This specification defines Beacon's logging and observability strategy, including log levels, structured logging, what to log, performance considerations, and metrics for monitoring. Effective logging enables debugging without overwhelming users; observability enables production monitoring.

---

## Constitutional Compliance

This specification aligns with the [Beacon Constitution v1.0.0](../memory/constitution.md):

**I. RFC Compliant**: Logging captures RFC-critical events (probing, announcing, cache-flush, TC bit, QU bit) with timing metadata to verify compliance with RFC 6762 timing requirements (§8.1-8.3). TXT record handling follows RFC 6763 §6.1 security guidance by redacting sensitive values.

**II. Spec-Driven Development**: This architecture specification governs logging patterns across all Beacon components, ensuring consistent observability before implementation begins.

**III. Test-Driven Development**: Logging is testable via TestLogger interface, enabling verification of log output in unit tests. Success criteria include test logger implementation.

**VII. Excellence**: Structured logging using Go's standard `log/slog` follows best practices. Performance considerations prevent logging overhead in hot paths, maintaining production quality.

**Validation**: This specification has been validated against RFC 6762 and RFC 6763 to ensure logged events correspond to RFC-mandated protocol behaviors, with no conflicts identified.

---

## Requirements

### REQ-F6-1: Optional Logging
Logging MUST be optional. `nil` logger means no logging.

**Rationale**: Not all users want logging overhead. Default to silent operation.

### REQ-F6-2: Structured Logging
Logs SHOULD be structured (key/value pairs) for machine parsing.

**Rationale**: Enables log aggregation, filtering, analysis.

### REQ-F6-3: Log Levels
Beacon MUST support standard log levels: Debug, Info, Warn, Error.

### REQ-F6-4: No Logging in Hot Paths
MUST NOT log in performance-critical code paths at Info/Debug levels.

**Rationale**: Logging has overhead. Reserve hot-path logging for errors.

**Hot Paths Defined**:
- Packet parsing (message decode/encode)
- Name compression/decompression
- Cache lookups (per-packet basis)
- Record comparison (duplicate suppression)
- Tight loops (processing large record sets)

**Allowed in Hot Paths**:
- Error logging (always allowed)
- Warn logging (sparingly, for actual issues)

**Not Allowed in Hot Paths**:
- Debug logging
- Info logging (except at boundaries)

### REQ-F6-5: Context in Logs
Logs SHOULD include relevant context (operation, name, address).

**Rationale**: Enables diagnosis without reading code.

### REQ-F6-6: TXT Record Redaction
TXT record values MUST NOT be logged (keys only). TXT records may contain sensitive data.

**Rationale**: TXT records often contain API keys, passwords, tokens. Log only keys for debugging, never values.

**Examples**:
```go
// ✅ Good - logs keys only
logger.Debug("TXT record", "keys", []string{"version", "api_key"})

// ❌ Bad - logs sensitive values
logger.Debug("TXT record", "data", map[string]string{
    "api_key": "secret123", // NEVER LOG THIS!
})
```

---

## Log Levels

### Debug
**Purpose**: Detailed diagnostic information for troubleshooting.

**When to use**:
- Packet parsing details
- State transitions
- Algorithm steps
- Cache operations

**Examples**:
```
level=debug msg="parsed query" name=myhost.local qtype=A questions=1
level=debug msg="cache lookup" name=myhost.local found=true ttl=95s
level=debug msg="sending probe" name=myhost.local attempt=1/3
```

**Enabled by**: Users opt-in for debugging.

### Info
**Purpose**: Normal operation events of interest.

**When to use**:
- Component start/stop
- Service discovered/removed
- Name conflicts resolved
- Successful operations

**Examples**:
```
level=info msg="querier started" interfaces=[eth0,wlan0]
level=info msg="service discovered" instance="My Printer" type=_printer._tcp
level=info msg="name conflict resolved" old=myhost.local new=myhost-2.local
```

**Enabled by**: Default for most users.

### Warn
**Purpose**: Unexpected but recoverable situations.

**When to use**:
- Malformed packets received
- Timeouts (but retrying)
- Configuration oddities
- Deprecation warnings

**Examples**:
```
level=warn msg="malformed packet received" from=192.168.1.100 error="invalid header flags"
level=warn msg="query timeout" name=myhost.local attempt=2/3
level=warn msg="cache full" size=10000 evicting=100
```

**Enabled by**: Always (minimal logging).

### Error
**Purpose**: Operation failures requiring attention.

**When to use**:
- Network errors (can't bind socket, permission denied)
- Fatal errors (component can't start)
- Unrecoverable failures

**Examples**:
```
level=error msg="failed to join multicast group" addr=224.0.0.251 error="permission denied"
level=error msg="responder stopped" error="context cancelled"
level=error msg="query failed" name=myhost.local error="network unreachable"
```

**Enabled by**: Always.

---

## Logger Interface

### Standard Interface

Use standard library `log/slog` (Go 1.21+):

```go
import "log/slog"

// Logger is the interface for logging in Beacon.
// Implementations should be safe for concurrent use.
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

**Rationale**: `log/slog` is standard library (Go 1.21+), structured, performant.

### Adapter for slog

```go
// SlogLogger adapts slog.Logger to Beacon's Logger interface.
type SlogLogger struct {
    logger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) *SlogLogger {
    return &SlogLogger{logger: logger}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
    l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
    l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
    l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
    l.logger.Error(msg, args...)
}
```

### Usage

```go
// User provides logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

q, err := querier.New(
    querier.WithLogger(NewSlogLogger(logger)),
)

// Or no logging
q, err := querier.New() // nil logger, no logging
```

---

## What to Log

### Component Lifecycle

**Start**:
```go
logger.Info("querier started",
    "interfaces", interfaceNames(cfg.interfaces),
    "ipv4", cfg.ipv4,
    "ipv6", cfg.ipv6,
)
```

**Stop**:
```go
logger.Info("querier stopped")
```

### Queries

**Sent**:
```go
logger.Debug("query sent",
    "name", query.Name,
    "qtype", query.Type,
    "addr", destAddr,
)
```

**Received Response**:
```go
logger.Debug("response received",
    "name", query.Name,
    "records", len(response.Answers),
    "from", sourceAddr,
)
```

**Timeout**:
```go
logger.Warn("query timeout",
    "name", query.Name,
    "attempt", attempt,
    "maxAttempts", maxAttempts,
)
```

### Responses (Responder)

**Query Received**:
```go
logger.Debug("query received",
    "name", question.Name,
    "qtype", question.Type,
    "from", sourceAddr,
)
```

**Response Sent**:
```go
logger.Debug("response sent",
    "name", question.Name,
    "records", len(response.Answers),
    "to", destAddr,
)
```

### Service Discovery

**Service Discovered**:
```go
logger.Info("service discovered",
    "instance", instance.Name,
    "type", instance.Type,
    "host", serviceInfo.Host,
    "port", serviceInfo.Port,
)
```

**Service Removed**:
```go
logger.Info("service removed",
    "instance", instance.Name,
    "type", instance.Type,
    "reason", "goodbye packet",
)
```

### Conflicts

**Conflict Detected**:
```go
logger.Warn("name conflict detected",
    "name", record.Name,
    "type", record.Type,
    "conflictingData", record.Data,
)
```

**Conflict Resolved**:
```go
logger.Info("name conflict resolved",
    "oldName", oldName,
    "newName", newName,
)
```

### Probing & Announcing (RFC 6762 §8)

**Probe Start**:
```go
logger.Info("probing started",
    "name", name,
    "count", 3,
    "interval", "250ms",
)
```

**Probe Sent**:
```go
logger.Debug("probe sent",
    "name", name,
    "attempt", attempt,
    "timestamp", time.Now(), // Timing metadata for verification
)
```

**Probe Complete (No Conflict)**:
```go
logger.Info("probing completed",
    "name", name,
    "duration", duration, // Total probing time
    "result", "success",
)
```

**Announce Start**:
```go
logger.Info("announcing started",
    "name", name,
    "count", count,
    "interval", "1s",
)
```

**Announce Sent**:
```go
logger.Debug("announcement sent",
    "name", name,
    "attempt", attempt,
    "records", recordCount,
    "timestamp", time.Now(), // Timing metadata
)
```

**Announce Complete**:
```go
logger.Info("announcing completed",
    "name", name,
    "announcements", count,
    "duration", duration,
)
```

**Goodbye Packet Sent**:
```go
logger.Info("goodbye sent",
    "name", name,
    "reason", reason, // e.g., "service stopping", "name change"
)
```

### mDNS Protocol Bits

**TC Bit (Truncation)**:
```go
logger.Debug("message truncated",
    "name", name,
    "tc", true,
    "size", actualSize,
    "maxSize", maxSize,
    "delay", delay, // 400-500ms delay applied
)
```

**Cache-Flush Bit**:
```go
logger.Debug("cache-flush record",
    "name", record.Name,
    "type", record.Type,
    "cache_flush", true, // 0x8000 bit set
)
```

**QU Bit (Unicast Response Preferred)**:
```go
logger.Debug("unicast response requested",
    "name", question.Name,
    "qu", true, // 0x8000 bit set in QCLASS
)
```

### TXT Record Handling

**TXT Record Logged (Redacted)**:
```go
// Per REQ-F6-6: Log TXT keys only, not values (may contain secrets)
logger.Debug("TXT record",
    "name", record.Name,
    "keys", txtKeys, // e.g., ["version", "path", "api_key"]
    "size", totalSize,
    // NEVER log: "values", txtValues (may contain secrets)
)
```

**TXT Record Size Warning**:
```go
logger.Warn("TXT record exceeds recommended size",
    "name", record.Name,
    "size", actualSize,
    "recommended", 200,
    "preferred", 400,
)
```

### Errors

**Network Error**:
```go
logger.Error("network error",
    "op", "send query",
    "addr", addr,
    "error", err,
)
```

**Protocol Error**:
```go
logger.Warn("protocol error",
    "op", "parse message",
    "from", sourceAddr,
    "error", err,
)
```

---

## Logging Patterns

### Pattern 1: Log at Boundaries

Log at public API entry points:

```go
func (q *Querier) Query(ctx context.Context, name string) ([]Record, error) {
    q.logger.Debug("query started", "name", name)
    defer q.logger.Debug("query completed", "name", name)

    records, err := q.performQuery(ctx, name)
    if err != nil {
        q.logger.Error("query failed", "name", name, "error", err)
        return nil, err
    }

    q.logger.Info("query succeeded", "name", name, "records", len(records))
    return records, nil
}
```

### Pattern 2: Conditional Logging

Only log if logger is not nil:

```go
func (q *Querier) logDebug(msg string, args ...any) {
    if q.logger != nil {
        q.logger.Debug(msg, args...)
    }
}

// Usage
q.logDebug("cache lookup", "name", name, "found", found)
```

### Pattern 3: Structured Context

Include relevant context:

```go
// Good - includes context
logger.Error("query failed",
    "name", name,
    "qtype", qtype,
    "timeout", timeout,
    "attempt", attempt,
    "error", err,
)

// Bad - missing context
logger.Error("query failed", "error", err)
```

### Pattern 4: Avoid Logging in Loops

```go
// Bad - logs every iteration (spam!)
for _, record := range records {
    logger.Debug("processing record", "name", record.Name)
    process(record)
}

// Good - log summary
logger.Debug("processing records", "count", len(records))
for _, record := range records {
    process(record)
}
logger.Debug("records processed", "count", len(records))
```

---

## Metrics & Observability

### Metrics to Expose

Use a metrics interface for flexibility:

```go
// Metrics collects operational metrics.
type Metrics interface {
    // Query metrics
    IncQueriesSent()
    IncResponsesReceived()
    IncQueryTimeouts()
    ObserveQueryDuration(duration time.Duration)

    // Response metrics
    IncQueriesReceived()
    IncResponsesSent()

    // Cache metrics
    IncCacheHits()
    IncCacheMisses()
    ObserveCacheSize(size int)

    // Error metrics
    IncNetworkErrors()
    IncProtocolErrors()
    IncConflicts()
}
```

### Prometheus Example

```go
import "github.com/prometheus/client_golang/prometheus"

type PrometheusMetrics struct {
    queriesSent      prometheus.Counter
    responsesRcvd    prometheus.Counter
    queryTimeouts    prometheus.Counter
    queryDuration    prometheus.Histogram
    cacheHits        prometheus.Counter
    cacheMisses      prometheus.Counter
    networkErrors    prometheus.Counter
    protocolErrors   prometheus.Counter
    conflicts        prometheus.Counter
}

func (m *PrometheusMetrics) IncQueriesSent() {
    m.queriesSent.Inc()
}

func (m *PrometheusMetrics) ObserveQueryDuration(d time.Duration) {
    m.queryDuration.Observe(d.Seconds())
}

// ... other methods
```

### Metrics Configuration

```go
// Optional metrics
q, err := querier.New(
    querier.WithMetrics(myMetrics),
)
```

---

## Performance Considerations

### Logging Overhead

**Guideline**: Debug/Info logging should have minimal overhead when disabled.

```go
// Bad - always allocates even if logger is nil
q.logger.Debug("msg", "data", expensiveOperation())

// Good - short-circuit if no logger
if q.logger != nil {
    q.logger.Debug("msg", "data", expensiveOperation())
}

// Better - lazy evaluation
func (q *Querier) debugf(format string, fn func() []any) {
    if q.logger != nil {
        q.logger.Debug(fmt.Sprintf(format, fn()...))
    }
}
```

### Structured Logging Performance

Use `slog` efficiently:

```go
// Efficient - no allocation if level disabled
logger.Debug("message", "key", value)

// Less efficient - always formats
logger.Debug(fmt.Sprintf("message: %v", value))
```

### Sampling for High-Volume Events

For very high-frequency events, sample:

```go
type sampledLogger struct {
    logger   Logger
    counter  atomic.Uint64
    interval uint64
}

func (l *sampledLogger) Debug(msg string, args ...any) {
    if l.counter.Add(1)%l.interval == 0 {
        l.logger.Debug(msg, args...)
    }
}
```

---

## Testing with Logging

### Test Logger

```go
// TestLogger captures logs for verification.
type TestLogger struct {
    mu   sync.Mutex
    logs []LogEntry
}

type LogEntry struct {
    Level string
    Msg   string
    Args  map[string]any
}

func (l *TestLogger) Debug(msg string, args ...any) {
    l.append("debug", msg, args)
}

func (l *TestLogger) append(level, msg string, args []any) {
    l.mu.Lock()
    defer l.mu.Unlock()

    entry := LogEntry{
        Level: level,
        Msg:   msg,
        Args:  argsToMap(args),
    }
    l.logs = append(l.logs, entry)
}

func (l *TestLogger) Logs() []LogEntry {
    l.mu.Lock()
    defer l.mu.Unlock()
    return append([]LogEntry{}, l.logs...)
}

// Usage in tests
func TestQueryLogging(t *testing.T) {
    logger := &TestLogger{}
    q, _ := querier.New(querier.WithLogger(logger))

    q.Query(ctx, "myhost.local")

    logs := logger.Logs()
    if len(logs) == 0 {
        t.Error("expected logs, got none")
    }

    // Verify specific log
    found := false
    for _, log := range logs {
        if log.Msg == "query started" && log.Args["name"] == "myhost.local" {
            found = true
            break
        }
    }
    if !found {
        t.Error("expected 'query started' log")
    }
}
```

---

## Debugging Aids

### Packet Dumping

For troubleshooting, support packet dumping:

```go
// WithPacketDump enables hexdump of packets.
func WithPacketDump(enabled bool) Option

// Internal
func (t *Transport) logPacket(direction string, packet []byte) {
    if !t.packetDump {
        return
    }
    t.logger.Debug("packet dump",
        "direction", direction,
        "size", len(packet),
        "data", hex.EncodeToString(packet),
    )
}
```

### State Dumping

```go
// DumpState returns current internal state for debugging.
func (q *Querier) DumpState() map[string]any {
    return map[string]any{
        "queries":   q.stats.queriesSent,
        "responses": q.stats.responsesReceived,
        "cache":     q.cache.Size(),
    }
}
```

---

## Documentation

### Log Format Documentation

Document log schema:

```markdown
# Log Fields

All logs include:
- `level`: debug, info, warn, error
- `msg`: Human-readable message
- `time`: Timestamp (RFC3339)

Common fields:
- `name`: Domain name
- `qtype`: Query type (A, AAAA, PTR, SRV, TXT)
- `from`: Source address
- `to`: Destination address
- `error`: Error message
- `attempt`: Retry attempt number
```

### Log Examples

Provide examples in documentation:

```markdown
# Example Logs

## Query
```
json
{
  "level": "debug",
  "time": "2025-10-31T12:00:00Z",
  "msg": "query sent",
  "name": "myhost.local",
  "qtype": "A",
  "addr": "224.0.0.251:5353"
}
```

## Error
```json
{
  "level": "error",
  "time": "2025-10-31T12:00:01Z",
  "msg": "query failed",
  "name": "myhost.local",
  "error": "network unreachable"
}
```
```

---

## Open Questions

**Q1**: Should we use context.Context for request-scoped logging?
- **Pro**: Automatic correlation (trace IDs)
- **Con**: Complexity, not idiomatic for libraries
- **Decision**: No, keep logging simple. Applications can correlate.

**Q2**: Log redaction for sensitive data?
- **Example**: TXT records may contain secrets
- **Pro**: Security
- **Con**: Harder debugging
- **Decision**: ✅ DECIDED - Per REQ-F6-6, NEVER log TXT values, only keys. This is now a MUST requirement.

**Q3**: Sampling rate for high-volume logs?
- **Decision**: Make configurable via WithLogSampling(rate int)

---

## Success Criteria

- [ ] Logger interface defined
- [ ] slog adapter implemented
- [ ] Log levels documented
- [ ] What to log specified
- [ ] Performance impact minimized
- [ ] Test logger for testing
- [ ] Metrics interface defined

---

## References

- [Beacon Constitution v1.0.0](../memory/constitution.md)
- [BEACON_FOUNDATIONS v1.1](./BEACON_FOUNDATIONS.md)
- [RFC 6762: Multicast DNS](../../RFC%20Docs/RFC-6762-Multicast-DNS.txt)
- [RFC 6763: DNS-Based Service Discovery](../../RFC%20Docs/RFC-6763-DNS-SD.txt)
- Go Blog: [Structured Logging with slog](https://go.dev/blog/slog)
- `log/slog` package documentation
- Prometheus: [Best Practices](https://prometheus.io/docs/practices/)

---

## Version History

| Version | Date | Changes | Validated Against RFCs |
|---------|------|---------|------------------------|
| 1.1 | 2025-11-01 | Aligned with Constitution v1.0.0 and Foundations v1.1; added Constitutional Compliance section; added RFC validation status; updated governance references | Yes (RFC 6762, RFC 6763) |
| 1.0 | 2025-11-01 | Initial architecture specification with hot path definitions, probe/announce logging, timing metadata, TXT redaction, mDNS bit logging | Partial |
