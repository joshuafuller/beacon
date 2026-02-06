# Intermediate Examples - Contracts

**Date**: 2026-01-06
**Feature**: 008-documentation-production-polish
**Priority**: P1 (For v1.1)

## Purpose

This document specifies the interface and behavior contracts for the 5 intermediate examples. Intermediate examples target developers with mDNS basics who need real-world integration patterns (web servers, Docker, IoT, custom service types).

---

## Example 6: Web Server with mDNS

### Contract Specification

**Path**: `examples/intermediate/web-server/`
**Estimated Time**: 1 day implementation
**Target Audience**: Web developers adding service discovery

### Purpose
Demonstrate integrating mDNS responder with a standard HTTP server so the web application is discoverable without manual configuration.

### Interface Contract

**Inputs**:
- HTTP requests on port 8080

**Outputs**:
- Console message: "HTTP server listening on :8080"
- Console message: "mDNS service registered: Web Demo._http._tcp.local"
- HTTP responses to requests
- mDNS announcements on network

**Behavior**:
1. Create HTTP server (`http.Server`) listening on `:8080`
2. Register mDNS responder for `_http._tcp` service with TXT records:
   - `path=/` (root endpoint)
   - `version=1.0`
3. Start HTTP server in goroutine
4. Print startup messages
5. Serve HTTP requests at `/` (return "Hello from mDNS-discoverable server!")
6. On shutdown:
   - Gracefully stop HTTP server (30-second timeout)
   - Unregister mDNS service
   - Close responder

### Key Concepts Demonstrated
- Dual functionality (HTTP + mDNS in single process)
- TXT record metadata (path, version)
- Goroutine coordination (HTTP server + mDNS responder)
- Graceful shutdown of multiple subsystems

### Expected File Structure
```
web-server/
├── README.md         # Integration patterns with web frameworks
├── main.go           # ~100 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] HTTP server accessible at `http://localhost:8080`
- [ ] Service visible in Safari (Bonjour browser) or `dns-sd -B _http._tcp`
- [ ] TXT records visible with `dns-sd -L "Web Demo" _http._tcp`
- [ ] Clicking service in Safari opens web page
- [ ] Graceful shutdown stops both HTTP and mDNS cleanly

### RFC References
- RFC 6763 §6 (TXT Record Construction)
- RFC 6763 §4.1.3 (Path and URL TXT Keys)

---

## Example 7: Service Updates

### Contract Specification

**Path**: `examples/intermediate/service-updates/`
**Estimated Time**: 1 day implementation
**Target Audience**: Developers needing dynamic service metadata

### Purpose
Demonstrate updating service TXT records at runtime (e.g., health status, load level, feature flags).

### Interface Contract

**Inputs**:
- Timer ticks (every 5 seconds)

**Outputs**:
- Console message: "Service registered with status=healthy, load=0"
- Console message (every 5s): "Updated service: status=healthy, load=[0-100]"
- mDNS update announcements (TXT record changes)

**Behavior**:
1. Register service with initial TXT records:
   - `status=healthy`
   - `load=0`
   - `features=v1,v2`
2. Every 5 seconds:
   - Simulate load change (random 0-100)
   - Update TXT record `load=[new value]`
   - Call `responder.UpdateService()`
   - Print update confirmation
3. After 30 seconds (6 updates):
   - Gracefully shutdown

### Key Concepts Demonstrated
- Dynamic TXT record updates
- RFC 6762 §10.3 (announcing changes - reannounce with new values)
- Use cases: health status, feature flags, dynamic metadata

### Expected File Structure
```
service-updates/
├── README.md         # Dynamic update patterns and use cases
├── main.go           # ~90 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] TXT records change every 5 seconds (visible in `dns-sd -L`)
- [ ] Observers receive updated TXT values without re-querying
- [ ] README explains RFC 6762 §10.3 update mechanism
- [ ] Example shows real-world use case (health, load, features)

### RFC References
- RFC 6762 §10.3 (Announcing Changes - reannounce updated records)

---

## Example 8: Multi-Interface Subnet Bridge (IoT)

### Contract Specification

**Path**: `examples/intermediate/multi-interface-bridge/`
**Estimated Time**: 1 day implementation
**Target Audience**: IoT developers with multi-network devices

### Purpose
Demonstrate bridging mDNS queries between network interfaces (e.g., WiFi + Ethernet) to enable service discovery across isolated subnets.

**Real-World Scenario**: Raspberry Pi acting as edge gateway with:
- WiFi interface (192.168.1.0/24) - connected to home network
- Ethernet interface (10.0.0.0/24) - connected to isolated IoT subnet with sensors

### Interface Contract

**Inputs**:
- Configuration:
  - Bridge interfaces: `wlan0`, `eth0`
  - Service allowlist: `_http._tcp`, `_homekit._tcp`
  - Subnet exclusion: Docker (`172.17.0.0/16`), VPN (`10.8.0.0/24`)
- mDNS queries on either interface

**Outputs**:
- Console message: "Bridge started: wlan0 ↔ eth0"
- Console message: "Allowed services: _http._tcp, _homekit._tcp"
- Console message (per query): "Forwarding query [service] from wlan0 → eth0"
- Forwarded mDNS queries and responses between interfaces

**Behavior**:
1. **Initialization**:
   - Parse interface names (`wlan0`, `eth0`)
   - Create querier and responder for each interface
   - Load service type allowlist from config

2. **Bridge Logic** (RFC 6762 §15):
   - Listen for queries on **wlan0**:
     - If query matches allowlist → forward to **eth0**
     - Rewrite source interface in responses
   - Listen for queries on **eth0**:
     - If query matches allowlist → forward to **wlan0**
     - Rewrite source interface in responses

3. **Filtering**:
   - **Service Type Filtering**: Only forward allowed service types (security)
   - **Subnet Exclusion**: Ignore queries from Docker/VPN interfaces (F-10 compatibility)

4. **Shutdown**:
   - Stop forwarding on both interfaces
   - Close queriers and responders gracefully

### Key Concepts Demonstrated
- Multi-interface operations (RFC 6762 §15)
- Query forwarding and response rewriting
- Service type filtering (security/performance)
- Subnet exclusion (F-10 compliance - Docker/VPN)
- Interface-specific IP addressing (007-interface-specific-addressing)

### Expected File Structure
```
multi-interface-bridge/
├── README.md         # IoT bridge patterns and security considerations
├── main.go           # ~150 lines (bridge orchestration, signal handling)
├── bridge.go         # ~100 lines (Bridge type, forwarding logic)
├── config.yaml       # Configuration file (interfaces, allowlist)
├── go.mod
└── Makefile
```

### Bridge Type Contract

```go
type Bridge struct {
    // Interface names
    InterfaceA string // e.g., "wlan0"
    InterfaceB string // e.g., "eth0"

    // Filtering
    AllowedServices []string // e.g., ["_http._tcp", "_homekit._tcp"]
    ExcludedSubnets []string // e.g., ["172.17.0.0/16", "10.8.0.0/24"]

    // Internal
    querierA *querier.Querier
    querierB *querier.Querier
    // ... (private fields)
}

func NewBridge(config *BridgeConfig) (*Bridge, error)
func (b *Bridge) Start(ctx context.Context) error
func (b *Bridge) Stop() error
func (b *Bridge) forwardQuery(query *Query, from, to string) error
```

### Success Criteria
- [ ] Queries on wlan0 for `_http._tcp` forwarded to eth0
- [ ] Responses from eth0 devices visible to wlan0 clients
- [ ] Filtered services (not in allowlist) NOT forwarded
- [ ] Docker/VPN queries ignored (subnet exclusion)
- [ ] README explains RFC 6762 §15 (interface-specific addressing)
- [ ] README includes IoT security best practices (allowlist rationale)
- [ ] Example runs on Raspberry Pi (tested on RPi 4)

### RFC References
- RFC 6762 §15 (Multiple Interfaces - interface-specific addressing)
- RFC 6762 §5.5 (Multicast Scope - link-local only, no routing)

### Security Considerations
- **Allowlist Required**: Never forward all services (potential security risk)
- **Subnet Filtering**: Prevent accidental bridging of Docker/VPN traffic
- **No Internet Forwarding**: mDNS is link-local only (RFC 6762 §3)

### Testing Procedure

**Prerequisites**:
- Multi-interface machine (WiFi + Ethernet) OR virtual machine with bridged + NAT network adapters
- Verify interfaces exist: `ip addr` (Linux) or `ifconfig` (macOS) showing both interfaces
- Two separate network subnets (e.g., WiFi on 192.168.1.0/24, Ethernet on 10.0.0.0/24)

**Success Criteria**:

1. **Query Forwarding Verification**:
   - Start packet capture on wlan0: `sudo tcpdump -i wlan0 port 5353 -v`
   - Send query on eth0 (from device on Ethernet subnet)
   - Expected: Query appears in wlan0 packet capture showing forwarded mDNS query

2. **Interface-Specific IP Addressing** (RFC 6762 §15):
   - Check bridge console output for "Forwarding query [service] from eth0 → wlan0"
   - Verify response includes **wlan0's IP address** (not eth0's IP, not 0.0.0.0)
   - Use `dns-sd -L [service-name] [service-type]` from wlan0 subnet to see IP address in response

3. **Service Filtering**:
   - Query for allowed service (`_http._tcp`): Should see "Forwarding query" message
   - Query for non-allowed service (`_ssh._tcp`): Should NOT see forwarding message
   - Verify only allowlisted services cross interface boundary

4. **Subnet Exclusion** (Docker/VPN):
   - Start Docker container and check bridge ignores queries from `172.17.0.0/16`
   - Expected console output: "Ignoring query from excluded subnet 172.17.x.x"

**Testing Commands**:

```bash
# Terminal 1: Start bridge
cd examples/intermediate/multi-interface-bridge
make run

# Terminal 2: Monitor wlan0 traffic
sudo tcpdump -i wlan0 port 5353 -v

# Terminal 3: Send test query from eth0 subnet
# (Run on device connected to Ethernet network)
dns-sd -B _http._tcp

# Verify in Terminal 1 console: "Forwarding query _http._tcp from eth0 → wlan0"
# Verify in Terminal 2: mDNS packet appears on wlan0
```

**Fallback Testing** (No Multi-Interface Machine):

If you don't have a physical multi-interface machine, use Docker network simulation:

```bash
# Create two Docker networks
docker network create --subnet=192.168.1.0/24 net-wifi
docker network create --subnet=10.0.0.0/24 net-ethernet

# Run bridge container attached to both networks
docker run --network net-wifi --network net-ethernet beacon-bridge

# Run test service on net-ethernet
docker run --network net-ethernet beacon-test-service

# Run test client on net-wifi
docker run --network net-wifi beacon-test-client

# Verify client discovers service across bridge
```

Alternatively, rely on unit tests for bridge logic validation:
```bash
cd examples/intermediate/multi-interface-bridge
go test -v ./... -run TestBridge
```

**Expected Output**:

```
Bridge started: wlan0 ↔ eth0
Allowed services: _http._tcp, _homekit._tcp
Excluded subnets: 172.17.0.0/16, 10.8.0.0/24

[Forwarding query _http._tcp.local from eth0 → wlan0]
[Forwarding response from wlan0 → eth0 (IP: 192.168.1.100)]

[Ignoring query _ssh._tcp.local (not in allowlist)]
[Ignoring query from 172.17.0.5 (excluded subnet)]
```

---

## Example 9: Custom Service Type

### Contract Specification

**Path**: `examples/intermediate/custom-service-type/`
**Estimated Time**: 1 day implementation
**Target Audience**: Developers creating custom protocols

### Purpose
Demonstrate defining a custom service type for application-specific discovery (e.g., `_myapp._tcp`).

### Interface Contract

**Inputs**:
- None (custom service definition)

**Outputs**:
- Console message: "Custom service registered: My App Instance._myapp._tcp.local"
- Console message: "TXT records: api_version=2.0, protocol=custom, features=feature1,feature2"
- mDNS announcements for custom service type

**Behavior**:
1. Define custom service type: `_myapp._tcp`
2. Register service with rich TXT metadata:
   - `api_version=2.0`
   - `protocol=custom`
   - `features=feature1,feature2`
   - `endpoint=/api/v2`
3. Print registration confirmation
4. Wait for interrupt
5. Graceful shutdown

### Key Concepts Demonstrated
- Custom service type naming (RFC 6763 §7)
- TXT record schema design
- Application-specific metadata
- Service type registration (no IANA registration needed for local use)

### Expected File Structure
```
custom-service-type/
├── README.md         # Service type naming conventions and best practices
├── main.go           # ~80 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] Service visible with `dns-sd -B _myapp._tcp`
- [ ] TXT records parseable with `dns-sd -L "My App Instance" _myapp._tcp`
- [ ] README explains RFC 6763 §7 naming rules (_<service>._<proto>)
- [ ] README includes TXT schema documentation template

### RFC References
- RFC 6763 §7 (Service Names - `_<service>._<proto>.<domain>`)
- RFC 6763 §6 (TXT Record Construction)

---

## Example 10: Logging Integration

### Contract Specification

**Path**: `examples/intermediate/logging-integration/`
**Estimated Time**: 1 day implementation
**Target Audience**: Production deployments needing observability

### Purpose
Demonstrate integrating Beacon with Go's structured logging (`log/slog`) for production observability.

### Interface Contract

**Inputs**:
- Environment variable `LOG_LEVEL` (debug, info, warn, error)
- Environment variable `LOG_FORMAT` (json, text)

**Outputs**:
- Structured logs (JSON or text) to stdout:
  ```json
  {"time":"2026-01-06T10:00:00Z","level":"INFO","msg":"Service registered","service":"Demo Service","type":"_http._tcp","port":8080}
  {"time":"2026-01-06T10:00:05Z","level":"DEBUG","msg":"Query received","query":"_http._tcp.local","source":"192.168.1.100"}
  ```

**Behavior**:
1. Parse `LOG_LEVEL` and `LOG_FORMAT` from environment
2. Configure `slog` handler (JSON or text)
3. Create responder with logging hooks:
   - Log service registration
   - Log incoming queries
   - Log response sending
   - Log errors with stack traces
4. Demonstrate different log levels:
   - DEBUG: Query details, packet sizes
   - INFO: Service lifecycle events
   - WARN: Rate limiting, validation warnings
   - ERROR: Network errors, configuration errors

### Key Concepts Demonstrated
- Structured logging with `log/slog` (Go 1.21+)
- Log level configuration
- JSON output for log aggregation (ELK, Splunk)
- Production observability patterns

### Expected File Structure
```
logging-integration/
├── README.md         # Logging best practices and log schema
├── main.go           # ~100 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] JSON logs parseable by `jq` or log aggregator
- [ ] Log levels respected (INFO doesn't show DEBUG)
- [ ] All service lifecycle events logged
- [ ] README documents log schema (fields and meanings)
- [ ] Example shows integration with observability stack

### RFC References
- None (logging is implementation-specific, but should log RFC-relevant events)

---

## Cross-Example Consistency Requirements

All intermediate examples MUST adhere to these standards:

### Code Complexity
- [ ] Examples are self-contained (no external services required)
- [ ] Code is ≤200 lines total (split into multiple files if needed)
- [ ] Complex logic extracted into helper functions with clear names

### Production Patterns
- [ ] Configuration via environment variables or config files
- [ ] Graceful shutdown with timeout (30 seconds max)
- [ ] Structured error handling (don't panic in example code)
- [ ] Resource cleanup on all exit paths

### README Quality
- [ ] Real-world use case clearly explained
- [ ] "What could go wrong?" section (common pitfalls)
- [ ] Production deployment considerations
- [ ] Links to relevant F-specs (F-9, F-10, F-11)

### Testing
- [ ] Runs successfully in Docker container
- [ ] Compatible with systemd service deployment
- [ ] Works with firewall enabled (document required ports)
- [ ] Tested on multi-interface machine (WiFi + Ethernet)

---

## Integration with P1 Tasks

| Example | Tasks | Files |
|---------|-------|-------|
| **Web Server** | T062-T066 | README, main.go, go.mod, Makefile |
| **Service Updates** | T067-T071 | README, main.go, go.mod, Makefile |
| **Multi-Interface Bridge** | T072-T077 | README, main.go, bridge.go, config.yaml, go.mod, Makefile |
| **Custom Service Type** | T078-T082 | README, main.go, go.mod, Makefile |
| **Logging Integration** | T083-T087 | README, main.go, go.mod, Makefile |

**Checkpoint**: After T087, all 5 intermediate examples must compile, run, and integrate with production patterns.

---

## Special Focus: Multi-Interface Bridge (User-Requested)

The multi-interface subnet bridge example is a **critical IoT use case** explicitly requested by the user. This example should be:

1. **Production-Ready**: Suitable for deployment on Raspberry Pi or edge gateways
2. **Well-Documented**: Explain RFC 6762 §15 in detail, security implications
3. **Secure by Default**: Allowlist-based filtering, subnet exclusion
4. **F-10 Compliant**: Integrate with existing network interface management (VPN/Docker exclusion)

This example serves as a reference implementation for advanced users building IoT gateways and should showcase Beacon's capabilities in real-world scenarios.

---

**Status**: Contract specification complete, ready for implementation
