# Beacon Documentation

Welcome to the Beacon documentation! This guide will help you find exactly what you need.

---

## 📚 Documentation by Audience

### 👤 **I want to USE Beacon** (Users)

**Start here**: [Getting Started Guide](guides/getting-started.md) ⭐

**User Guides**:
- [Getting Started](guides/getting-started.md) - Installation, first query/response (15 min)
- [Architecture Overview](guides/architecture.md) - How Beacon works (10 min read)
- [Troubleshooting Guide](guides/troubleshooting.md) - Common issues and solutions

**API Reference**:
- [API Overview](api/README.md) - Quick reference and common patterns
- [Querier API](api/querier.md) - Service discovery (coming soon)
- [Responder API](api/responder.md) - Service announcement (coming soon)
- [Common Types](api/types.md) - Shared types and constants (coming soon)

**Examples**:
- [Code Examples](../examples/) - Working code samples

---

### 🛠️ **I want to CONTRIBUTE** (Contributors)

**Start here**: [Contributing Guide](../CONTRIBUTING.md) ⭐

**Development Guides**:
- [Development Setup](development/setup.md) - Environment setup (coming soon)
- [Testing Guide](development/testing.md) - How to write tests (coming soon)
- [Contributing Code](development/contributing-code.md) - PR workflow (coming soon)

**Community**:
- [Code of Conduct](../CODE_OF_CONDUCT.md) - Community standards
- [Security Policy](../SECURITY.md) - Reporting vulnerabilities

---

### 🔬 **I want to UNDERSTAND the internals** (Researchers/Architects)

**Technical Documentation**:
- [RFC Compliance Matrix](internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md) - 72.2% RFC 6762/6763 compliant
- [Functional Requirements Matrix](internals/rfc-compliance/FUNCTIONAL_REQUIREMENTS_MATRIX.md) - All 61 FRs with traceability
- [Compliance Dashboard](internals/rfc-compliance/COMPLIANCE_DASHBOARD.md) - Quick status overview

**Architecture**:
- [Architecture Decisions (ADRs)](internals/architecture/decisions/) - Why we made key decisions
  - [ADR-001: Transport Interface Abstraction](internals/architecture/decisions/001-transport-interface-abstraction.md)
  - [ADR-002: Buffer Pooling Pattern](internals/architecture/decisions/002-buffer-pooling-pattern.md)
  - [ADR-003: Integration Test Timing Tolerance](internals/architecture/decisions/003-integration-test-timing-tolerance.md)
  - [ADR-004: Coverage Enforcement in CI Not Hooks](internals/architecture/decisions/004-coverage-enforcement-in-ci-not-hooks.md)
  - [ADR-005: DNS-SD TTL Values](internals/architecture/decisions/005-dns-sd-ttl-values.md)
- [Architectural Pitfalls & Mitigations](internals/architecture/ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md) - Security and resilience

**Analysis**:
- [hashicorp/mdns Comparison](internals/analysis/HASHICORP_COMPARISON.md) - Why Beacon is 10,000x faster
- [Security Audit](../specs/006-mdns-responder/SECURITY_AUDIT.md) - STRONG security grade
- [Performance Analysis](../specs/006-mdns-responder/PERFORMANCE_ANALYSIS.md) - Benchmark results (A+ grade)
- [Code Review](../specs/006-mdns-responder/CODE_REVIEW.md) - Code quality assessment (A grade)

**Shipping**:
- [Shipping Guide](internals/SHIPPING_GUIDE.md) - Using Beacon in production

---

## 🚀 Quick Links

### Essentials

- **[Getting Started](guides/getting-started.md)** - New to Beacon? Start here
- **[API Reference](api/README.md)** - Quick API lookup
- **[Troubleshooting](guides/troubleshooting.md)** - Having issues? Check here
- **[Examples](../examples/)** - Working code samples
- **[Contributing](../CONTRIBUTING.md)** - Want to help? Read this

### Project Info

- **[README](../README.md)** - Project overview and quick start
- **[CHANGELOG](../CHANGELOG.md)** - Version history and changes
- **[ROADMAP](../ROADMAP.md)** - Future plans (M2-M6)
- **[LICENSE](../LICENSE)** - MIT License

### Protocol References

- **[RFC 6762](../RFC%20Docs/rfc6762.txt)** - Multicast DNS specification
- **[RFC 6763](../RFC%20Docs/rfc6763.txt)** - DNS-Based Service Discovery

---

## 📖 Documentation Structure

```
docs/
├── README.md (you are here)         # Documentation hub
│
├── guides/                          # 👤 USER-FACING GUIDES
│   ├── getting-started.md           # Installation, first query/response
│   ├── architecture.md              # High-level architecture overview
│   └── troubleshooting.md           # Common issues and solutions
│
├── api/                             # 📚 API REFERENCE
│   ├── README.md                    # API overview and common patterns
│   ├── querier.md                   # Querier API reference (coming soon)
│   ├── responder.md                 # Responder API reference (coming soon)
│   └── types.md                     # Common types and constants (coming soon)
│
├── development/                     # 🛠️ DEVELOPER DOCUMENTATION
│   ├── README.md                    # Development hub (coming soon)
│   ├── setup.md                     # Dev environment setup (coming soon)
│   ├── testing.md                   # Testing guide (coming soon)
│   └── contributing-code.md         # How to contribute code (coming soon)
│
└── internals/                       # 🔬 INTERNAL/TECHNICAL DOCS
    ├── rfc-compliance/              # RFC compliance tracking
    │   ├── RFC_COMPLIANCE_MATRIX.md
    │   ├── FUNCTIONAL_REQUIREMENTS_MATRIX.md
    │   └── COMPLIANCE_DASHBOARD.md
    ├── architecture/                # Architecture decisions
    │   ├── decisions/               # ADRs
    │   └── ARCHITECTURAL_PITFALLS_AND_MITIGATIONS.md
    ├── analysis/                    # Performance & security analysis
    │   └── HASHICORP_COMPARISON.md
    └── SHIPPING_GUIDE.md            # Production deployment guide
```

---

## 🎯 Common Tasks

### "I want to discover services on my network"

1. Read: [Getting Started - Your First Query](guides/getting-started.md#your-first-query-service-discovery)
2. See: [Querier API Reference](api/README.md#querier-api)
3. Try: [Basic Query Example](../examples/basic-query/)

### "I want to announce a service"

1. Read: [Getting Started - Your First Responder](guides/getting-started.md#your-first-responder-service-announcement)
2. See: [Responder API Reference](api/README.md#responder-api)
3. Try: [Basic Responder Example](../examples/basic-responder/)

### "I'm getting an error"

1. Check: [Troubleshooting Guide](guides/troubleshooting.md)
2. Search: [GitHub Issues](https://github.com/joshuafuller/beacon/issues)
3. Ask: [GitHub Discussions](https://github.com/joshuafuller/beacon/discussions)

### "I want to contribute code"

1. Read: [Contributing Guide](../CONTRIBUTING.md)
2. Setup: [Development Environment](development/setup.md) (coming soon)
3. Learn: [Testing Guide](development/testing.md) (coming soon)

### "I want to understand RFC compliance"

1. Overview: [Compliance Dashboard](internals/rfc-compliance/COMPLIANCE_DASHBOARD.md)
2. Details: [RFC Compliance Matrix](internals/rfc-compliance/RFC_COMPLIANCE_MATRIX.md)
3. Traceability: [Functional Requirements Matrix](internals/rfc-compliance/FUNCTIONAL_REQUIREMENTS_MATRIX.md)

### "I want to deploy to production"

1. Read: [Shipping Guide](internals/SHIPPING_GUIDE.md)
2. Review: [Security Audit](../specs/006-mdns-responder/SECURITY_AUDIT.md)
3. Check: [Performance Analysis](../specs/006-mdns-responder/PERFORMANCE_ANALYSIS.md)

---

## 🆘 Getting Help

### Self-Service

1. **Search documentation** - Use your browser's find feature (Ctrl+F / Cmd+F)
2. **Check troubleshooting** - [Troubleshooting Guide](guides/troubleshooting.md)
3. **Browse examples** - [Examples Directory](../examples/)
4. **Read the code** - Beacon is well-documented with godoc comments

### Community Support

- **[GitHub Discussions](https://github.com/joshuafuller/beacon/discussions)** - Ask questions, share ideas
- **[GitHub Issues](https://github.com/joshuafuller/beacon/issues)** - Report bugs, request features
- **Email**: joshuafuller@gmail.com - Direct support

### Before Asking

Please provide:
- **Go version**: `go version`
- **Beacon version**: Check `go.mod`
- **OS**: Linux/macOS/Windows + version
- **Minimal reproduction**: Smallest code that shows the issue
- **What you've tried**: Steps you've already taken

**Good question example**:

```
Title: "Query returns empty results on Ubuntu 22.04"

Environment:
- Go 1.21.3
- Beacon v0.1.0
- Ubuntu 22.04

Issue:
When I query for _http._tcp.local services, I get an empty result set,
but `avahi-browse -a` shows 3 services.

Code:
[paste minimal reproduction]

What I've tried:
- Increased context timeout to 10 seconds
- Checked firewall (sudo iptables -L shows no blocks)
- Verified network interface is up
- Ran tcpdump, see attached pcap showing responses

Attached:
- mdns-capture.pcap
```

---

## 📊 Documentation Status

### ✅ Complete

- Getting Started Guide
- Architecture Overview
- Troubleshooting Guide
- API Overview
- CODE_OF_CONDUCT.md
- SECURITY.md
- CONTRIBUTING.md
- RFC Compliance Matrix
- Functional Requirements Matrix
- Architecture Decision Records (5 ADRs)

### 🚧 Coming Soon

- Querier API Reference (detailed)
- Responder API Reference (detailed)
- Common Types Reference
- Development Setup Guide
- Testing Guide
- Contributing Code Guide
- Platform-Specific Guides (Linux/macOS/Windows)
- Migration Guide (from hashicorp/mdns)
- Advanced Usage Guide

### 📝 Future

- Performance Tuning Guide
- Observability Guide (when logging is implemented)
- IPv6 Guide (when IPv6 is implemented)
- Service Browsing Guide (when implemented)

---

## 🔄 Documentation Feedback

**Found an issue?** Documentation bugs are still bugs!

- **Typos or errors**: [Open an issue](https://github.com/joshuafuller/beacon/issues/new)
- **Unclear explanations**: [Start a discussion](https://github.com/joshuafuller/beacon/discussions)
- **Missing topics**: [Request in discussions](https://github.com/joshuafuller/beacon/discussions)

**Want to contribute?** Documentation PRs are welcome! See [Contributing Guide](../CONTRIBUTING.md#improve-documentation).

---

## 📜 License

All documentation is licensed under [MIT License](../LICENSE), same as the code.

Feel free to use, copy, modify, and distribute with attribution.

---

**Last Updated**: 2025-11-04
**Documentation Version**: 2.0 (Post-M2 Documentation Overhaul)

---

**Happy coding with Beacon! 🚀**
