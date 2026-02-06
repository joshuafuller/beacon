# Basic Examples - Contracts

**Date**: 2026-01-06
**Feature**: 008-documentation-production-polish
**Priority**: P0 (Critical for v1.0)

## Purpose

This document specifies the interface and behavior contracts for the 5 basic examples. Basic examples target developers new to mDNS/Beacon and focus on fundamental operations with minimal complexity.

---

## Example 1: Hello Responder

### Contract Specification

**Path**: `examples/basic/hello-responder/`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers new to mDNS

### Purpose
Demonstrate the absolute minimum code needed to register a service and make it discoverable on the local network.

### Interface Contract

**Inputs**:
- None (hardcoded service definition)

**Outputs**:
- Console message: "Service registered: [instance].[service].[domain]"
- Console message: "Press Ctrl+C to exit"
- mDNS announcement on network (multicast to 224.0.0.251:5353)

**Behavior**:
1. Create responder with default options
2. Define service: Instance="Hello World", Service="_http._tcp", Domain="local", Port=8080
3. Register service (triggers RFC 6762 §8.1 probing + §8.3 announcing)
4. Print success message
5. Wait for interrupt signal (Ctrl+C)
6. Close responder gracefully

### Key Concepts Demonstrated
- Responder creation and lifecycle
- Service structure (Instance, Service, Domain, Port)
- Graceful shutdown pattern

### Expected File Structure
```
hello-responder/
├── README.md         # What/Why/How/Output/Troubleshooting
├── main.go           # ~50 lines
├── go.mod            # With replace directive
└── Makefile          # Standard targets
```

### Success Criteria
- [ ] Compiles with `go build main.go`
- [ ] Runs with `go run main.go` (no errors)
- [ ] Service visible via `dns-sd -B _http._tcp` (macOS) or `avahi-browse -t _http._tcp` (Linux)
- [ ] Graceful shutdown on Ctrl+C
- [ ] README follows template exactly

### RFC References
- RFC 6762 §8.1 (Probing)
- RFC 6762 §8.3 (Announcing)

---

## Example 2: Error Handling

### Contract Specification

**Path**: `examples/basic/error-handling/`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers learning error patterns

### Purpose
Demonstrate all error types that Beacon can return and how to handle them properly.

### Interface Contract

**Inputs**:
- None (intentionally trigger errors via invalid configurations)

**Outputs**:
- Console messages showing each error type
- Proper error handling patterns (typed errors, error wrapping)

**Behavior**:
1. **Scenario 1 - Validation Error**: Attempt to register service with invalid instance name (empty string)
   - Catch `*errors.ValidationError`
   - Print: "ValidationError: [message]"

2. **Scenario 2 - Network Error**: Attempt to create responder on port already in use
   - Catch `*errors.NetworkError`
   - Print: "NetworkError: [message]"

3. **Scenario 3 - Context Cancellation**: Cancel context during registration
   - Catch `context.Canceled`
   - Print: "Operation canceled: [message]"

4. **Scenario 4 - Proper Error Handling**: Show correct pattern for production code
   - Use typed error assertions
   - Log errors with structured logging
   - Graceful degradation

### Key Concepts Demonstrated
- Beacon error types (`ValidationError`, `NetworkError`)
- Go error handling patterns (type assertions, error wrapping)
- Context cancellation handling
- Structured error logging

### Expected File Structure
```
error-handling/
├── README.md         # Error patterns and handling strategies
├── main.go           # ~80 lines (4 error scenarios)
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] All 4 error scenarios trigger correctly
- [ ] Each error type is caught and handled
- [ ] README explains each error type with recovery strategies
- [ ] Code demonstrates production-ready error handling patterns

### RFC References
- None (error handling is implementation-specific)

---

## Example 3: Graceful Shutdown

### Contract Specification

**Path**: `examples/basic/graceful-shutdown/`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers deploying to production

### Purpose
Demonstrate proper service lifecycle management, including goodbye packets and clean resource cleanup.

### Interface Contract

**Inputs**:
- Signal (SIGINT from Ctrl+C or SIGTERM from container orchestrator)

**Outputs**:
- Console message: "Service registered: [details]"
- Console message: "Shutting down gracefully..."
- Goodbye packet sent (RFC 6762 §10.1 - PTR record with TTL=0)
- Console message: "Goodbye packet sent"
- Console message: "Shutdown complete"

**Behavior**:
1. Register service (as in hello-responder)
2. Set up signal handling for SIGINT and SIGTERM
3. Wait for interrupt signal
4. On signal received:
   - Print "Shutting down gracefully..."
   - Call `responder.Unregister()` (sends goodbye packet)
   - Print "Goodbye packet sent"
   - Call `responder.Close()`
   - Wait 250ms for goodbye packet propagation
   - Print "Shutdown complete"
   - Exit with code 0

### Key Concepts Demonstrated
- Signal handling (os.Signal, signal.Notify)
- Goodbye packet transmission (RFC 6762 §10.1)
- Resource cleanup order (unregister → close → exit)
- Timing considerations (250ms goodbye propagation)

### Expected File Structure
```
graceful-shutdown/
├── README.md         # Shutdown patterns and timing rationale
├── main.go           # ~60 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] Responds to both SIGINT (Ctrl+C) and SIGTERM (kill)
- [ ] Goodbye packet visible in Wireshark (PTR record with TTL=0)
- [ ] Service disappears from `dns-sd -B` within 1 second of Ctrl+C
- [ ] No resource leaks (goroutines, file descriptors)
- [ ] README explains RFC 6762 §10.1 goodbye packet requirement

### RFC References
- RFC 6762 §10.1 (Goodbye Packets - TTL=0)

---

## Example 4: Multi-Service

### Contract Specification

**Path**: `examples/basic/multi-service/`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers running multi-protocol services

### Purpose
Demonstrate registering multiple services from a single application (e.g., web server exposing HTTP, HTTPS, and SSH).

### Interface Contract

**Inputs**:
- None (hardcoded 3 services)

**Outputs**:
- Console message: "Registered service 1/3: [instance]._http._tcp.local"
- Console message: "Registered service 2/3: [instance]._ssh._tcp.local"
- Console message: "Registered service 3/3: [instance]._custom._tcp.local"
- Console message: "All services registered. Press Ctrl+C to exit."

**Behavior**:
1. Create responder
2. Define 3 services:
   - **HTTP Service**: Instance="Multi-Service Demo", Service="_http._tcp", Port=8080
   - **SSH Service**: Instance="Multi-Service Demo", Service="_ssh._tcp", Port=22
   - **Custom Service**: Instance="Multi-Service Demo", Service="_myapp._tcp", Port=9000
3. Register each service sequentially
4. Print confirmation after each registration
5. Wait for interrupt
6. Unregister all services (goodbye packets for each)

### Key Concepts Demonstrated
- Multiple service registration from single responder
- Different service types (_http, _ssh, custom)
- Shared instance name across services
- Bulk unregistration on shutdown

### Expected File Structure
```
multi-service/
├── README.md         # Multi-service patterns and use cases
├── main.go           # ~70 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] All 3 services visible simultaneously in `dns-sd -B`
- [ ] Each service has correct port and TXT records
- [ ] Unregistering one service doesn't affect others
- [ ] All services send goodbye packets on shutdown
- [ ] README explains common multi-service scenarios

### RFC References
- RFC 6763 §7 (Service Names - _<service>._<proto>.<domain>)

---

## Example 5: Service Browser

### Contract Specification

**Path**: `examples/basic/browser/`
**Estimated Time**: 2 hours implementation
**Target Audience**: Developers discovering services on network

### Purpose
Demonstrate querying for available services using RFC 6763 §9 service enumeration.

### Interface Contract

**Inputs**:
- None (queries `_services._dns-sd._udp.local`)

**Outputs**:
- Console message: "Browsing for available services..."
- List of discovered services:
  ```
  Found service: _http._tcp.local
  Found service: _ssh._tcp.local
  Found service: _airplay._tcp.local
  ...
  ```
- Console message: "Total services found: [count]"

**Behavior**:
1. Create querier with default options
2. Query for `_services._dns-sd._udp.local` (RFC 6763 §9)
3. Parse PTR responses (each PTR record points to a service type)
4. Deduplicate service types
5. Print each unique service type
6. Print total count
7. Exit after 2-second query timeout

### Key Concepts Demonstrated
- Querier usage (vs. Responder)
- Service enumeration (RFC 6763 §9)
- PTR record parsing
- Query timeout handling

### Expected File Structure
```
browser/
├── README.md         # Service discovery patterns
├── main.go           # ~70 lines
├── go.mod
└── Makefile
```

### Success Criteria
- [ ] Discovers all services on local network
- [ ] No duplicate service types in output
- [ ] Handles empty responses gracefully (no services found)
- [ ] Query completes within 2 seconds
- [ ] README explains RFC 6763 §9 service enumeration

### RFC References
- RFC 6763 §9 (Service Type Enumeration - `_services._dns-sd._udp`)

---

## Cross-Example Consistency Requirements

All basic examples MUST adhere to these standards:

### Code Style
- [ ] Use `context.Background()` for simple examples (not `context.TODO()`)
- [ ] Always `defer responder.Close()` or `querier.Close()`
- [ ] Use `log.Fatal()` for unrecoverable errors (before main loop starts)
- [ ] Use structured error handling in main loop
- [ ] Comment all non-obvious code

### README Style
- [ ] What/Why/How/Expected Output/Troubleshooting sections
- [ ] Quick start takes ≤3 commands
- [ ] Expected output shows actual console output
- [ ] Troubleshooting covers ≥2 common issues
- [ ] Links to next steps (related examples, docs)

### Testing
- [ ] Compiles with zero warnings (`go build`)
- [ ] Runs successfully on Linux and macOS
- [ ] Compatible with Avahi (Linux) and Bonjour (macOS)
- [ ] Works with firewall enabled (port 5353 open)

### Documentation
- [ ] RFC references cited where applicable
- [ ] Technical terms defined on first use
- [ ] Code snippets <20 lines in README
- [ ] Link to full example code, not inline paste

---

## Integration with P0 Tasks

| Example | Tasks | Files |
|---------|-------|-------|
| **Hello Responder** | T007-T011 | README, main.go, go.mod, Makefile |
| **Error Handling** | T012-T016 | README, main.go, go.mod, Makefile |
| **Graceful Shutdown** | T017-T021 | README, main.go, go.mod, Makefile |
| **Multi-Service** | T022-T026 | README, main.go, go.mod, Makefile |
| **Service Browser** | T027-T031 | README, main.go, go.mod, Makefile |

**Checkpoint**: After T031, all 5 basic examples must compile, run, and pass manual testing.

---

**Status**: Contract specification complete, ready for implementation
