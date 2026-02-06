# Beacon Examples

Working code examples demonstrating how to use the Beacon mDNS library.

---

## Quick Start

### Run an Example

```bash
# Discover services on your network
cd examples/discover
go run main.go

# Expected output: List of discovered devices and services
```

---

## Available Examples

### 🔍 [discover/](discover/) - Service Discovery

**What it does**: Discovers devices and services on your local network

**Use this to learn**:
- How to create a querier
- How to query for services
- How to parse PTR and SRV records
- How to discover service types

**Run it**:
```bash
cd examples/discover
go run main.go
```

**Expected output**:
```
🔍 Discovering devices on local network...
📡 Found 3 service type(s):
  • _http._tcp.local
  • _ssh._tcp.local
  • _printer._tcp.local

---

🌐 HTTP Services:
  • My Printer._http._tcp.local → printer.local:80
  • Home Server._http._tcp.local → server.local:8080
```

---

## Basic Examples (New!)

The following examples are now available in `examples/basic/`:

### 📢 [hello-responder/](basic/hello-responder/) - Minimal Service Registration

**What it does**: Absolute minimum code to register a service

**Topics**:
- Creating a responder
- Defining a service
- Registration and graceful shutdown

**Run it**:
```bash
cd examples/basic/hello-responder
make run
```

### ⚠️ [error-handling/](basic/error-handling/) - Error Handling Patterns

**What it does**: Demonstrates all error types and handling patterns

**Topics**:
- Validation errors (empty instance name, invalid port)
- Context cancellation
- Production error handling patterns

**Run it**:
```bash
cd examples/basic/error-handling
make run
```

### 🛑 [graceful-shutdown/](basic/graceful-shutdown/) - Proper Shutdown

**What it does**: Clean service termination with goodbye packets

**Topics**:
- Signal handling (SIGINT, SIGTERM)
- RFC 6762 §10.1 goodbye packets
- Resource cleanup

**Run it**:
```bash
cd examples/basic/graceful-shutdown
make run
# Press Ctrl+C to see goodbye packets
```

### 🔄 [multi-service/](basic/multi-service/) - Multiple Services

**What it does**: Register 3+ services with one responder

**Topics**:
- Single responder, multiple services
- Different service types (_http._tcp, _ssh._tcp)
- Coordinated shutdown

**Run it**:
```bash
cd examples/basic/multi-service
make run
```

### 🔍 [browser/](basic/browser/) - Service Discovery

**What it does**: Discover HTTP and SSH services on network

**Topics**:
- Querier API
- PTR record queries
- Processing query results

**Run it**:
```bash
cd examples/basic/browser
make run
```

---

## Intermediate Examples

Production-ready patterns for real-world integration:

### 🌐 [web-server/](intermediate/web-server/) - HTTP Server with mDNS

**What it does**: Integrate mDNS with a standard HTTP server for auto-discovery

**Topics**:
- Dual HTTP + mDNS in single process
- TXT record metadata (path, version)
- Graceful shutdown of multiple subsystems

**Run it**:
```bash
cd examples/intermediate/web-server
make run
# Visit http://localhost:8080 or discover via Safari/dns-sd
```

### 🔄 [service-updates/](intermediate/service-updates/) - Dynamic TXT Records

**What it does**: Update service metadata at runtime (health, load, features)

**Topics**:
- RFC 6762 §10.3 change announcements
- Dynamic TXT record updates
- Use cases: health status, load balancing, feature flags

**Run it**:
```bash
cd examples/intermediate/service-updates
make run
# Watch: dns-sd -L "Status Demo" _http._tcp
```

### 🔧 [custom-service-type/](intermediate/custom-service-type/) - Custom Service Type

**What it does**: Define application-specific service types (_myapp._tcp)

**Topics**:
- RFC 6763 §7 service naming conventions
- Custom TXT schema design
- Application-specific discovery

**Run it**:
```bash
cd examples/intermediate/custom-service-type
make run
# Discover: dns-sd -B _myapp._tcp
```

### 📊 [logging-integration/](intermediate/logging-integration/) - Production Logging

**What it does**: Structured logging with Go's log/slog for observability

**Topics**:
- JSON logging for log aggregation (ELK, Splunk)
- Log level configuration (debug, info, warn, error)
- Service lifecycle event tracking

**Run it**:
```bash
cd examples/intermediate/logging-integration
LOG_LEVEL=debug LOG_FORMAT=json make run
```

### 🌉 [multi-interface-bridge/](intermediate/multi-interface-bridge/) - IoT Bridge (Documentation)

**What it does**: Bridge mDNS across network interfaces (WiFi ↔ Ethernet)

**Topics**:
- RFC 6762 §15 interface-specific addressing
- Query forwarding and response rewriting
- Security filtering (allowlist, subnet exclusion)

**Status**: Design documentation only (requires multi-interface hardware for testing)

---

## Learning Path

**New to mDNS?** Follow this order:

### Basics (~30 minutes)
1. **[basic/hello-responder/](basic/hello-responder/)** - Start here: minimal service registration (~5 min)
2. **[basic/browser/](basic/browser/)** - Discover services on the network (~5 min)
3. **[basic/error-handling/](basic/error-handling/)** - Handle validation and context errors (~10 min)
4. **[basic/graceful-shutdown/](basic/graceful-shutdown/)** - Clean termination with goodbye packets (~5 min)
5. **[basic/multi-service/](basic/multi-service/)** - Register multiple services (~5 min)

### Intermediate (~1 hour)
6. **[intermediate/web-server/](intermediate/web-server/)** - HTTP + mDNS integration (~15 min)
7. **[intermediate/service-updates/](intermediate/service-updates/)** - Dynamic metadata updates (~15 min)
8. **[intermediate/custom-service-type/](intermediate/custom-service-type/)** - Custom service types (~10 min)
9. **[intermediate/logging-integration/](intermediate/logging-integration/)** - Production logging (~15 min)

### Advanced
10. **[discover/](discover/)** - Full service discovery with SRV/TXT records
11. **[multi-interface-demo/](multi-interface-demo/)** - Multi-interface operations
12. **[interface-specific/](interface-specific/)** - Interface-specific addressing (RFC 6762 §15)
13. **[advanced/iot-device/](advanced/iot-device/)** - IoT device service registration

---

## Example Template

Want to contribute an example? Use this template:

```go
// Package main demonstrates [WHAT THE EXAMPLE DOES].
//
// This example shows how to:
// - [Feature 1]
// - [Feature 2]
// - [Feature 3]
//
// Usage:
//
//	go run examples/[NAME]/main.go
//
// Expected output:
//
//	[SAMPLE OUTPUT]
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/joshuafuller/beacon/querier"
)

func main() {
    // 1. Create querier/responder
    // 2. Use API
    // 3. Clean up resources
}
```

**Requirements**:
- Clear, commented code
- Error handling included
- Resource cleanup (defer Close())
- README.md explaining what it does
- Example output shown

---

## Troubleshooting Examples

### "No services found"

**Possible causes**:
- No mDNS services on your network
- Firewall blocking UDP port 5353
- VPN active (mDNS doesn't work over VPNs)

**Solutions**:
- Enable mDNS on a device (iOS, macOS, Linux with Avahi)
- Check firewall: `sudo iptables -L | grep 5353`
- Disconnect VPN and try again

### "Permission denied" error

**Cause**: Some systems require elevated privileges for port 5353

**Solution**:
```bash
# Linux - grant capability
sudo setcap 'cap_net_bind_service=+ep' ./main

# Or run with sudo (not recommended for production)
sudo go run main.go
```

### "Address already in use"

**Cause**: Another program is using port 5353 without SO_REUSEPORT

**Solution**:
```bash
# Check what's using the port
sudo ss -ulnp 'sport = :5353'

# If it's Avahi or Bonjour, that's okay - Beacon uses SO_REUSEPORT
# If it's something else, stop that process
```

---

## Contributing Examples

**Found a bug in an example?** [Open an issue](https://github.com/joshuafuller/beacon/issues)

**Have an idea for an example?** [Start a discussion](https://github.com/joshuafuller/beacon/discussions)

**Want to contribute an example?** See [Contributing Guide](../CONTRIBUTING.md)

**Good example topics**:
- Real-world use cases (IoT, microservices, etc.)
- Platform-specific examples (macOS, Windows, Linux)
- Integration with other libraries
- Performance optimization patterns

---

## Resources

- **[Getting Started Guide](../docs/guides/getting-started.md)** - Tutorial for new users
- **[API Reference](../docs/api/README.md)** - Complete API documentation
- **[Troubleshooting](../docs/guides/troubleshooting.md)** - Common issues

---

**Happy discovering! 🚀**
