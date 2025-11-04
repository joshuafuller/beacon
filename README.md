# Beacon - High-Performance mDNS Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/joshuafuller/beacon.svg)](https://pkg.go.dev/github.com/joshuafuller/beacon)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuafuller/beacon)](https://goreportcard.com/report/github.com/joshuafuller/beacon)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Beacon is a lightweight, high-performance mDNS (Multicast DNS) library for Go, implementing [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) for service discovery on local networks.

**Perfect for**: IoT devices, microservices, local network service discovery, replacing unmaintained alternatives like hashicorp/mdns.

---

## Features

- **10,000x faster** - 4.8μs response latency vs ~50ms in alternatives
- **72.2% RFC 6762 compliance** - Rigorous protocol implementation
- **Zero external dependencies** - Standard library only
- **Production-tested** - 81.3% test coverage, 109,471 fuzz executions, 0 data races
- **Automatic conflict resolution** - RFC 6762 §8.2 compliant
- **SO_REUSEPORT** - Coexists with Avahi/Bonjour system services

[See detailed comparison with hashicorp/mdns →](docs/HASHICORP_COMPARISON.md)

---

## Quick Start

### Installation

```bash
go get github.com/joshuafuller/beacon@latest
```

### Announce a Service

```go
import "github.com/joshuafuller/beacon/responder"

r, _ := responder.New(ctx)
defer r.Close()

svc := &responder.Service{
    Instance: "My Web Server",
    Service:  "_http._tcp",
    Domain:   "local",
    Port:     8080,
    TXT:      []string{"path=/"},
}

r.Register(ctx, svc) // Service is now announced on the network
```

### Discover Services

```go
import "github.com/joshuafuller/beacon/querier"

q, _ := querier.New()
defer q.Close()

results, _ := q.Query(ctx, "_http._tcp.local", querier.QueryTypePTR)
for _, rr := range results {
    fmt.Printf("Found: %s\n", rr.Name)
}
```

[More examples →](examples/)

---

## Why Beacon?

### Built with Engineering Rigor

Beacon is built on proven engineering principles:

- **RFC Compliance First** - Every feature validated against [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html)/[6763](https://www.rfc-editor.org/rfc/rfc6763.html)
- **Specification-Driven** - No code without a spec ([Spec Kit](https://github.com/github/spec-kit) framework)
- **Test-Driven Development** - Tests written first (RED → GREEN → REFACTOR)
- **Automated Quality** - 25 [Semgrep rules](SEMGREP_RULES_SUMMARY.md) enforce compliance
- **Constitutional Governance** - [Development principles](.specify/memory/constitution.md) enshrined

### Measurable Results

| Metric | hashicorp/mdns | Beacon | Improvement |
|--------|----------------|--------|-------------|
| Response Latency | ~50ms | 4.8μs | **10,000x faster** |
| RFC Compliance | ~7% | 72.2% | **10x better** |
| Fuzz Testing | 0 tests | 109,471 execs | **∞ better** |
| Test Coverage | ~10 tests | 247 tests | **25x better** |
| Data Races | Known | 0 (verified) | **Production ready** |

---

## Documentation

### Getting Started
- [Shipping Guide](docs/SHIPPING_GUIDE.md) - Using Beacon in your project
- [API Reference](https://pkg.go.dev/github.com/joshuafuller/beacon) - GoDoc
- [Examples](examples/) - Code samples

### Technical Details
- [RFC Compliance Matrix](docs/RFC_COMPLIANCE_MATRIX.md) - 72.2% compliant (91/126 requirements)
- [Performance Analysis](specs/006-mdns-responder/PERFORMANCE_ANALYSIS.md) - Benchmarks (Grade A+)
- [Security Audit](specs/006-mdns-responder/SECURITY_AUDIT.md) - Security posture (STRONG)
- [hashicorp/mdns Comparison](docs/HASHICORP_COMPARISON.md) - Detailed comparison

### Development
- [Contributing Guide](CLAUDE.md) - Development guidelines
- [Semgrep Rules](SEMGREP_RULES_SUMMARY.md) - Quality enforcement

---

## What's Implemented

### mDNS Responder (Service Announcement)
✅ RFC 6762 §8.1-8.3 Probing, Conflict Resolution, Announcing
✅ RFC 6762 §6.2 Rate Limiting (1 response/sec per record)
✅ RFC 6762 §7.1 Known-Answer Suppression (TTL ≥50%)
✅ Multi-service support per responder

### mDNS Querier (Service Discovery)
✅ RFC 6762 §5 Query transmission
✅ Context-aware, cancellable operations
✅ Thread-safe concurrent queries

### Platform Support
✅ **Linux** - Full support
⚠️ **macOS/Windows** - Code-complete, pending integration tests

---

## Roadmap

**v0.1.0** (Current) - Production Ready
**v0.2.0** - IPv6 support, Goodbye packets
**v0.3.0** - Unicast response support
**v0.4.0** - Service browsing

[See RFC Compliance Matrix for detailed status →](docs/RFC_COMPLIANCE_MATRIX.md)

---

## Requirements

- **Go 1.21 or later**
- **Standard library only** (zero external dependencies)

---

## Contributing

Contributions welcome! Please read [CLAUDE.md](CLAUDE.md) for guidelines.

Quick checklist:
- Tests written first (TDD)
- All tests pass with `-race`
- ≥80% coverage maintained
- `make semgrep-check` passes

---

## License

[MIT License](LICENSE) - Copyright (c) 2025 Joshua Fuller

---

## Acknowledgments

- Inspired by the need for a modern alternative to [hashicorp/mdns](https://github.com/hashicorp/mdns)
- Implements [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) and [RFC 6763](https://www.rfc-editor.org/rfc/rfc6763.html)
- Built using [Spec Kit](https://github.com/github/spec-kit) framework

**Questions?** [Open an issue](https://github.com/joshuafuller/beacon/issues) · **Email**: joshuafuller@gmail.com
