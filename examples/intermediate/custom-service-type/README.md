# Example 9: Custom Service Type

**Difficulty**: Intermediate
**Target Audience**: Developers creating custom protocols
**Estimated Time**: 10 minutes

## Overview

Demonstrates defining a custom mDNS service type (`_myapp._tcp`) for application-specific discovery with rich TXT metadata.

## Running

```bash
cd examples/intermediate/custom-service-type
make run
```

Discover with: `dns-sd -B _myapp._tcp`

## RFC References
- RFC 6763 §7: Service Names (`_<service>._<proto>.<domain>`)
- RFC 6763 §6: TXT Record Construction
