# BEACON Project Overview

## Purpose
BEACON is a lightweight, high-performance mDNS (Multicast DNS) library for Go, implementing RFC 6762 for service discovery on local networks.

**Status**: Production Ready (94% complete, final polish in progress via Ralph)

## Key Features
- **mDNS Querier**: Query for services on local network (RFC 6762)
- **mDNS Responder**: Announce services with probing and conflict resolution (RFC 6762 §8)
- **DNS-SD Support**: Service discovery (RFC 6763)
- **Interface-Specific Addressing**: Multi-interface IP resolution (RFC 6762 §15)
- **High Performance**: 4.8μs response time, 99% allocation reduction via buffer pooling

## Architecture Philosophy
- **Clean Architecture**: Strict layer boundaries (F-2 spec)
- **RFC Compliance First**: All behavior must comply with RFC 6762/6763
- **Test-Driven Development**: Tests written FIRST (RED → GREEN → REFACTOR)
- **Minimal Dependencies**: Standard library + golang.org/x/net + golang.org/x/sys only
- **Context-Aware**: All blocking operations accept context.Context

## Milestones Completed
- ✅ M1: mDNS Querier (100%)
- ✅ M1.1: Architectural Hardening (100%)
- ✅ M2: mDNS Responder (129/129 tasks, 100%)
- ✅ 007: Interface-Specific Addressing (116/116 tasks, 100%)

## Current Work (Ralph Autonomous Loop)
- Test coverage increase: 64.2% → 80%+
- Semgrep static analysis
- Documentation polish
- TODO cleanup

## Package Structure
- `querier/` - Public API for mDNS queries
- `responder/` - Public API for mDNS service announcements
- `internal/transport/` - Network transport abstraction (IPv4/IPv6)
- `internal/message/` - DNS message parsing/building
- `internal/responder/` - Responder implementation
- `internal/state/` - State machine (probing/announcing)
- `internal/records/` - DNS record construction
- `internal/security/` - Validation and rate limiting
- `tests/contract/` - RFC compliance tests
- `tests/integration/` - Integration tests
- `tests/fuzz/` - Fuzz tests
- `examples/` - Usage examples
- `specs/` - Feature specifications (Spec Kit framework)

## Tech Stack
- **Language**: Go 1.21+ (currently 1.23.5)
- **Testing**: Standard go test + contract tests + fuzz tests
- **Static Analysis**: Semgrep
- **Standards**: RFC 6762 (mDNS), RFC 6763 (DNS-SD), RFC 1035 (DNS)
- **Methodology**: Specification-Driven Development (Spec Kit)
