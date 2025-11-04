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

## Coming Soon

The following examples are planned:

### 📢 basic-responder/ - Service Announcement

**What it does**: Announces a service on the network

**Topics**:
- Creating a responder
- Defining a service
- Registration and probing
- Graceful shutdown

### 🔄 multi-service/ - Multiple Services

**What it does**: Announces multiple services from one responder

**Topics**:
- Registering multiple services
- Service enumeration
- Updating service metadata

### 🏭 production/ - Production-Ready Example

**What it does**: Production-ready service discovery with error handling

**Topics**:
- Comprehensive error handling
- Logging and observability
- Resource management
- Graceful shutdown

### 🤝 coexistence/ - Avahi/Bonjour Coexistence

**What it does**: Demonstrates running alongside system mDNS services

**Topics**:
- SO_REUSEPORT behavior
- Verifying coexistence
- Platform-specific notes

---

## Learning Path

**New to mDNS?** Follow this order:

1. **[discover/](discover/)** - Understand service discovery basics
2. **basic-responder/** (coming soon) - Learn to announce services
3. **multi-service/** (coming soon) - Work with multiple services
4. **production/** (coming soon) - Production patterns

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

## Running Examples with Docker

If you want to test in an isolated environment:

```bash
# Build container
docker build -t beacon-examples -f examples/Dockerfile .

# Run discovery example
docker run --rm --network host beacon-examples discover
```

**Note**: `--network host` is required for mDNS multicast to work.

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
