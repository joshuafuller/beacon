---
name: Bug Report
about: Report a bug in Beacon
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

**Clear and concise description of the bug**:


## Environment

- **Go version**: (run `go version`)
- **Beacon version**: (check `go.mod` or `go list -m github.com/joshuafuller/beacon`)
- **Operating System**: (e.g., Ubuntu 22.04, macOS 14.0, Windows 11)
- **Architecture**: (e.g., amd64, arm64)

## Steps to Reproduce

1.
2.
3.

## Expected Behavior

**What you expected to happen**:


## Actual Behavior

**What actually happened**:


## Minimal Reproduction Code

```go
package main

import (
    "context"
    "github.com/joshuafuller/beacon/querier"
)

func main() {
    // Minimal code that reproduces the issue
}
```

## Error Messages / Logs

```
Paste any error messages or relevant log output here
```

## Network Information (if relevant)

- **Network setup**: (e.g., WiFi, Ethernet, Docker, VM)
- **VPN active**: (yes/no)
- **Firewall**: (enabled/disabled)
- **Other mDNS services running**: (e.g., Avahi, Bonjour)

**Output of network diagnostic** (if applicable):
```bash
# Linux
sudo ss -ulnp 'sport = :5353'
ip addr show

# macOS
sudo lsof -i :5353
ifconfig

# tcpdump (if packet capture available)
sudo tcpdump -i any port 5353 -c 10
```

## Additional Context

**Anything else that might be helpful**:
- Does it work with system mDNS tools (avahi-browse, dns-sd)?
- Packet capture (attach .pcap file if available)
- Screenshots (if UI-related)

## Checklist

- [ ] I have checked [existing issues](https://github.com/joshuafuller/beacon/issues) for duplicates
- [ ] I have read the [Troubleshooting Guide](https://github.com/joshuafuller/beacon/blob/main/docs/guides/troubleshooting.md)
- [ ] I have provided a minimal reproduction case
- [ ] I have included environment details
