# Production Deployment Checklist

**Category**: Deployment
**Estimated Time**: 15 minutes
**Prerequisites**: Working Beacon application ready for production

## Overview

This checklist ensures your Beacon-based mDNS service is ready for production deployment. Complete all items before deploying to production environments.

## Pre-Deployment Validation

### Network Configuration

- [ ] **Multicast enabled on target network**
  ```bash
  # Linux: Check multicast routing
  ip route show | grep 224.0.0.0
  # Should show: 224.0.0.0/4 dev <interface> scope link

  # macOS: Check multicast interface
  netstat -rn | grep 224.0.0
  ```

- [ ] **Port 5353/UDP accessible (mDNS)**
  ```bash
  # Test UDP port 5353 is not blocked
  sudo lsof -i :5353
  # OR
  sudo netstat -ulpn | grep 5353
  ```

- [ ] **Firewall allows multicast to 224.0.0.251**
  ```bash
  # Linux (ufw)
  sudo ufw allow from 224.0.0.0/4
  sudo ufw allow to 224.0.0.251

  # Linux (iptables)
  sudo iptables -A INPUT -d 224.0.0.251 -j ACCEPT
  sudo iptables -A OUTPUT -d 224.0.0.251 -j ACCEPT
  ```

- [ ] **Network interfaces configured correctly**
  ```bash
  # List interfaces
  ip link show
  # Verify target interfaces are UP and support multicast (MULTICAST flag)
  ```

### Service Configuration

- [ ] **Service names follow RFC 6763 conventions**
  - Instance name: 1-63 bytes, UTF-8
  - Service type: `_service._proto.local` format
  - Valid protocols: `_tcp` or `_udp`
  ```go
  // Good
  svc := &responder.Service{
      InstanceName: "My API Server",
      ServiceType:  "_http._tcp.local",
      Port:         8080,
  }

  // Bad: Missing underscore, wrong domain
  svc := &responder.Service{
      InstanceName: "My API Server",
      ServiceType:  "http.tcp.localdomain", // ❌
      Port:         8080,
  }
  ```

- [ ] **Port numbers in valid range (1-65535)**
  ```go
  if svc.Port < 1 || svc.Port > 65535 {
      return fmt.Errorf("invalid port")
  }
  ```

- [ ] **TXT records under 1300 bytes total**
  ```go
  // Calculate TXT record size
  totalSize := 0
  for k, v := range txtRecords {
      totalSize += len(k) + len(v) + 1 // +1 for '=' separator
  }
  if totalSize > 1300 {
      log.Warn("TXT records exceed recommended 1300 byte limit")
  }
  ```

- [ ] **Service validation passes**
  ```go
  if err := svc.Validate(); err != nil {
      log.Fatal("Service validation failed:", err)
  }
  ```

### Resource Management

- [ ] **Context timeout configured appropriately**
  ```go
  // Production: Use context with timeout
  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()

  r, err := responder.New(ctx)
  ```

- [ ] **Graceful shutdown implemented**
  ```go
  // Handle SIGTERM/SIGINT for clean shutdown
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

  go func() {
      <-sigChan
      r.Close() // Sends goodbye packets (TTL=0)
      os.Exit(0)
  }()
  ```

- [ ] **Resource cleanup in error paths**
  ```go
  r, err := responder.New(ctx)
  if err != nil {
      return err
  }
  defer r.Close() // Ensure cleanup even on errors
  ```

### Security

- [ ] **Rate limiting enabled (default: 100 req/sec)**
  ```go
  // Querier rate limiting is enabled by default
  q, err := querier.New()

  // Adjust if needed
  q, err := querier.New(querier.WithRateLimitThreshold(50))
  ```

- [ ] **Source IP filtering configured**
  ```go
  // Rate limiter tracks queries per source IP
  // Automatically drops packets after threshold exceeded
  ```

- [ ] **Running as non-root user (if possible)**
  ```bash
  # Create dedicated user
  sudo useradd -r -s /bin/false mdns-service

  # Grant CAP_NET_RAW capability (Linux)
  sudo setcap cap_net_raw+ep /path/to/beacon-service
  ```

### Monitoring & Observability

- [ ] **Structured logging configured**
  ```go
  import "log/slog"

  logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
  logger.Info("service registered",
      "instance", svc.InstanceName,
      "service", svc.ServiceType,
      "port", svc.Port,
  )
  ```

- [ ] **Health check endpoint available**
  ```go
  // HTTP health check
  http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
      if responder.IsRunning() {
          w.WriteHeader(http.StatusOK)
          w.Write([]byte("OK"))
      } else {
          w.WriteHeader(http.StatusServiceUnavailable)
      }
  })
  ```

- [ ] **Metrics collection planned**
  - Service registration success/failure rate
  - Query response latency
  - Active service count
  - Network errors

### Testing

- [ ] **Service discoverable from another machine**
  ```bash
  # From another machine on same subnet
  dns-sd -B _http._tcp local          # macOS
  avahi-browse -t _http._tcp          # Linux
  ```

- [ ] **Service resolves to correct IP/port**
  ```bash
  dns-sd -L "My Service" _http._tcp   # macOS
  avahi-resolve -n "My Service._http._tcp.local"  # Linux
  ```

- [ ] **Goodbye packets sent on shutdown**
  ```bash
  # Monitor with tcpdump during shutdown
  sudo tcpdump -i any udp port 5353 -v
  # Look for TTL=0 records on responder.Close()
  ```

- [ ] **No name conflicts with existing services**
  ```bash
  # Check for existing services with same name
  dns-sd -B _http._tcp | grep "My Service"
  ```

## Environment-Specific Checks

### Docker/Container Deployment

- [ ] **network_mode: "host" configured** (multicast requirement)
- [ ] **SO_REUSEPORT enabled** (automatic in Beacon v1.0+)
- [ ] **Container has CAP_NET_RAW capability**
  ```yaml
  cap_add:
    - NET_RAW
  ```

### Kubernetes Deployment

- [ ] **hostNetwork: true configured** (multicast requirement)
- [ ] **DaemonSet used** (one responder per node)
- [ ] **Pod security context allows NET_RAW**

### systemd Service

- [ ] **Restart policy configured**
  ```ini
  [Service]
  Restart=on-failure
  RestartSec=5s
  ```

- [ ] **Environment variables set**
  ```ini
  [Service]
  Environment="LOG_LEVEL=info"
  Environment="SERVICE_PORT=8080"
  ```

## Common Pitfalls

❌ **Don't**: Run multiple responders for same service on one host
✅ **Do**: Use SO_REUSEPORT (Beacon handles this automatically)

❌ **Don't**: Forget to call `responder.Close()` on shutdown
✅ **Do**: Use `defer r.Close()` and signal handlers

❌ **Don't**: Assume multicast works across VLANs
✅ **Do**: Test on target network infrastructure

❌ **Don't**: Use production hostnames in service names
✅ **Do**: Use descriptive instance names (e.g., "API Server v2.1")

## Troubleshooting

If validation fails, consult:
- [Troubleshooting Guide](./troubleshooting.md) for common issues
- [Monitoring Guide](./monitoring.md) for observability setup
- [Docker Deployment Guide](./docker.md) for container-specific concerns

## References

- [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) - Multicast DNS
- [RFC 6763](https://www.rfc-editor.org/rfc/rfc6763.html) - DNS-Based Service Discovery
- [Beacon README](../../README.md) - Quick start and examples
