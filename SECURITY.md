# Security Policy

## Supported Versions

The following versions of Beacon are currently being supported with security updates:

| Version | Supported          | Status |
| ------- | ------------------ | ------ |
| 0.1.x   | :white_check_mark: | Current |
| < 0.1.0 | :x:                | Pre-release, not supported |

## Security Model

Beacon is designed with security as a core principle:

### Input Validation

**All user input is validated** before processing:
- Service names (RFC 6763 §4 compliance)
- Port numbers (1-65535 range)
- TXT record sizes (≤255 bytes)
- Domain names (must be "local")

**All network input is treated as untrusted**:
- DNS packets are validated before parsing
- Malformed packets return errors (never panic)
- No `unsafe` package usage in packet parsers
- Maximum packet size enforced (9KB, DNS limit)

### Resource Protection

**Rate Limiting** (RFC 6762 §6.2):
- Maximum 1 multicast response per second per record
- Prevents network flooding and abuse

**Resource Bounds**:
- Buffer sizes capped at 9KB (DNS maximum packet size)
- Context timeouts enforced
- Goroutine lifecycle tied to responder/querier lifetime

### Testing

**Security validation**:
- 109,471 fuzz executions (0 crashes)
- 25 Semgrep rules enforce security patterns
- Zero `panic()` on malformed input (verified by fuzz tests)
- STRONG security grade ([Security Audit](specs/006-mdns-responder/SECURITY_AUDIT.md))

## Reporting a Vulnerability

**We take security vulnerabilities seriously.** If you discover a security issue in Beacon, please report it responsibly.

### How to Report

**Email**: joshuafuller@gmail.com

**Please include**:
1. **Description** - What is the vulnerability?
2. **Impact** - What can an attacker do?
3. **Reproduction** - Steps to reproduce the issue
4. **Version** - Which version of Beacon is affected?
5. **Proof of Concept** - Code or packet capture demonstrating the issue (if applicable)

### What to Expect

1. **Acknowledgment** - We will acknowledge receipt within **48 hours**
2. **Assessment** - We will assess the vulnerability and determine severity
3. **Timeline** - We will provide an estimated fix timeline
4. **Fix** - We will develop and test a fix
5. **Disclosure** - We will coordinate disclosure with you

**Typical timeline**:
- **Critical** (RCE, DoS, data corruption): Fix within 7 days
- **High** (information disclosure, privilege escalation): Fix within 14 days
- **Medium** (resource exhaustion, denial of service): Fix within 30 days
- **Low** (minor issues): Fix in next release

### Disclosure Policy

**We follow coordinated disclosure**:

1. **Private disclosure** - Report sent to maintainers only
2. **Fix developed** - Patch created and tested
3. **Security advisory** - Published to GitHub Security Advisories
4. **Public disclosure** - After fix is released (typically 7-14 days)

**We will credit you** in the security advisory (unless you request anonymity).

### Scope

**In scope**:
- Vulnerabilities in Beacon code (`querier`, `responder`, `internal/*`)
- Protocol violations that enable attacks
- Input validation bypasses
- Resource exhaustion attacks
- Information disclosure
- Denial of service vulnerabilities

**Out of scope**:
- Theoretical attacks without proof of concept
- Vulnerabilities in dependencies (report to upstream projects)
- Social engineering attacks
- Physical access attacks
- Issues already reported

## Known Limitations

These are **not security vulnerabilities** but documented limitations:

### 1. mDNS Protocol Limitations

**No authentication** (RFC 6762 design):
- mDNS has no built-in authentication or encryption
- Any device on the local network can send mDNS packets
- Service announcements can be spoofed

**Mitigation**: Use mDNS only on trusted networks. For untrusted networks, use authenticated service discovery (e.g., DNS-SD over TLS).

**No message integrity** (RFC 6762 design):
- mDNS packets are not signed
- Responses can be forged

**Mitigation**: Beacon implements RFC 6762 §8.2 conflict detection to detect name conflicts. Use application-level authentication (e.g., TLS) to verify service identity.

### 2. Local Network Only

**mDNS is link-local** (by design):
- Only works on the same network segment
- Does not route over VPNs or across subnets

**Mitigation**: This is a feature, not a bug. It prevents external attackers from reaching mDNS services.

### 3. UDP Protocol Limitations

**Packet loss** (UDP design):
- Queries and responses can be lost
- No guaranteed delivery

**Mitigation**: Beacon follows RFC 6762 recommendations for retries and timing.

**Packet spoofing** (UDP design):
- Source IP addresses can be spoofed

**Mitigation**: Use application-level authentication to verify service identity.

## Security Best Practices

### For Users

1. **Use mDNS only on trusted networks** - Don't expose mDNS on public WiFi
2. **Validate service identity** - Use TLS to authenticate services after discovery
3. **Set appropriate context timeouts** - Prevent indefinite blocking
4. **Close resources** - Always `defer querier.Close()` and `responder.Close()`
5. **Validate discovered services** - Don't blindly trust service metadata

### Example: Secure Service Discovery

```go
// 1. Discover service
q, _ := querier.New()
defer q.Close()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

results, _ := q.Query(ctx, "_https._tcp.local", querier.QueryTypePTR)

// 2. Validate service exists
if len(results) == 0 {
    return errors.New("service not found")
}

// 3. Extract service details (requires additional queries for SRV, A records)
// ...

// 4. Connect with TLS to verify identity
conn, err := tls.Dial("tcp", serviceAddr, &tls.Config{
    ServerName: serviceName,
    // Verify certificate
})
```

### For Contributors

1. **Never use `panic()` on network input** - Return errors instead
2. **Validate all input** - User input and network input
3. **Use `defer` for cleanup** - Prevent resource leaks
4. **Write fuzz tests** - For all parsers and validators
5. **Run Semgrep** - `make semgrep-check` before committing

See [Development Security Guide](docs/development/security.md) for details.

## Security Audits

### M2 Responder Implementation (2025-11-04)

**Grade**: STRONG

**Key findings**:
- ✅ No `panic()` on malformed input (verified by 109,471 fuzz executions)
- ✅ All inputs validated per RFC 6763 §4
- ✅ Rate limiting implemented (RFC 6762 §6.2)
- ✅ No `unsafe` package usage in parsers
- ✅ Proper resource cleanup (25 Semgrep rules enforced)

**See**: [SECURITY_AUDIT.md](specs/006-mdns-responder/SECURITY_AUDIT.md)

## Security Contact

**Primary contact**: joshuafuller@gmail.com

**PGP key**: Available on request

**GitHub**: [@joshuafuller](https://github.com/joshuafuller)

## References

- [RFC 6762 Security Considerations](https://www.rfc-editor.org/rfc/rfc6762.html#section-16)
- [RFC 6763 Security Considerations](https://www.rfc-editor.org/rfc/rfc6763.html#section-8)
- [Beacon Security Audit](specs/006-mdns-responder/SECURITY_AUDIT.md)
- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)

---

**Thank you for helping keep Beacon secure!**
