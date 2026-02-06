# Docker Deployment Guide

**Category**: Deployment
**Estimated Time**: 30 minutes
**Prerequisites**: Docker 20.10+, docker-compose 1.29+

## Overview

Deploy Beacon-based mDNS services in Docker containers. This guide covers multicast networking requirements, best practices, and working examples for production deployments.

**Key Challenge**: mDNS uses IP multicast (224.0.0.251), which requires special Docker networking configuration.

---

## Quick Start

```bash
# Clone example
cd docs/deployment/docker-example

# Build and run
docker-compose up --build

# Test from host
dns-sd -B _http._tcp local          # macOS
avahi-browse -t _http._tcp          # Linux
```

---

## Multicast Networking in Docker

### Why network_mode: "host" is Required

Docker's default bridge networking **does not support multicast** between containers and the host network. For mDNS to work, you must use one of these approaches:

#### ✅ Option 1: Host Networking (Recommended)

```yaml
services:
  beacon-service:
    build: .
    network_mode: "host"  # Share host's network stack
```

**Pros**:
- Simplest setup
- Full multicast support
- No port mapping needed
- Production-grade performance

**Cons**:
- Container shares host's network namespace (less isolation)
- Port conflicts possible with host services
- Not available in Docker Swarm mode

#### ⚠️ Option 2: macvlan Network (Advanced)

```yaml
networks:
  mdns:
    driver: macvlan
    driver_opts:
      parent: eth0
    ipam:
      config:
        - subnet: 192.168.1.0/24
          gateway: 192.168.1.1

services:
  beacon-service:
    build: .
    networks:
      - mdns
```

**Pros**:
- Container gets its own MAC address
- Full network isolation
- Works in complex network topologies

**Cons**:
- Requires promiscuous mode on host interface
- Complex configuration
- May not work on cloud providers
- Container can't communicate with host (by design)

#### ❌ Option 3: Bridge Network (Does NOT Work)

```yaml
# ❌ This will NOT work for mDNS
services:
  beacon-service:
    build: .
    ports:
      - "5353:5353/udp"  # mDNS won't work through NAT
```

**Why it fails**: Bridge networking uses NAT, which breaks multicast group membership. mDNS is link-local only (RFC 6762 §11).

---

## Production Dockerfile

**File**: `docs/deployment/docker-example/Dockerfile`

```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first (Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (CGO_ENABLED=0 for static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -o /beacon-service \
    -ldflags="-w -s" \
    ./main.go

# Runtime stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /beacon-service .

# Expose application port (NOT mDNS port 5353)
# Port 5353 is handled by network_mode: host
EXPOSE 8080

# Run as non-root user (best practice)
RUN adduser -D -s /bin/sh mdns
USER mdns

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Start service
CMD ["./beacon-service"]
```

---

## docker-compose.yml

**File**: `docs/deployment/docker-example/docker-compose.yml`

```yaml
version: '3.8'

services:
  beacon-responder:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: beacon-responder
    network_mode: "host"  # Required for multicast
    restart: unless-stopped
    environment:
      - LOG_LEVEL=info
      - SERVICE_NAME=My Docker Service
      - SERVICE_PORT=8080
    # Health check
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s
    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 128M
        reservations:
          cpus: '0.25'
          memory: 64M
```

---

## Example Application

**File**: `docs/deployment/docker-example/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	// Structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Read environment variables
	serviceName := getEnv("SERVICE_NAME", "Docker Service")
	servicePort := getEnv("SERVICE_PORT", "8080")

	// Create responder
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r, err := responder.New(ctx)
	if err != nil {
		logger.Error("failed to create responder", "error", err)
		os.Exit(1)
	}
	defer r.Close()

	// Register service
	svc := &responder.Service{
		InstanceName: serviceName,
		ServiceType:  "_http._tcp.local",
		Port:         parsePort(servicePort),
		TXTRecords: map[string]string{
			"version": "1.0",
			"env":     "docker",
		},
	}

	if err := r.Register(svc); err != nil {
		logger.Error("failed to register service", "error", err)
		os.Exit(1)
	}

	logger.Info("service registered",
		"instance", svc.InstanceName,
		"service", svc.ServiceType,
		"port", svc.Port,
	)

	// Start HTTP server
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from Beacon Docker service!\n")
	})

	go func() {
		addr := ":" + servicePort
		logger.Info("starting HTTP server", "addr", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	logger.Info("shutting down gracefully")
	r.Close() // Sends goodbye packets (TTL=0)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parsePort(portStr string) int {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	if port < 1 || port > 65535 {
		return 8080
	}
	return port
}
```

---

## Testing

### 1. Build and Run

```bash
cd docs/deployment/docker-example
docker-compose up --build
```

Expected output:
```json
{"time":"2026-01-06T...", "level":"INFO", "msg":"service registered", "instance":"My Docker Service", "service":"_http._tcp.local", "port":8080}
{"time":"2026-01-06T...", "level":"INFO", "msg":"starting HTTP server", "addr":":8080"}
```

### 2. Verify mDNS Discovery (from host)

**macOS**:
```bash
dns-sd -B _http._tcp local
# Should show: My Docker Service._http._tcp.local

dns-sd -L "My Docker Service" _http._tcp
# Should show: Port 8080, IP address
```

**Linux**:
```bash
avahi-browse -t _http._tcp
# Should show: My Docker Service._http._tcp

avahi-resolve-service "My Docker Service" _http._tcp
# Should show: Port 8080, IP address
```

### 3. Test HTTP Endpoint

```bash
curl http://localhost:8080/
# Expected: Hello from Beacon Docker service!

curl http://localhost:8080/health
# Expected: OK
```

### 4. Verify Goodbye Packets on Shutdown

```bash
# In one terminal, monitor mDNS traffic
sudo tcpdump -i any udp port 5353 -v

# In another terminal, stop container
docker-compose down

# tcpdump should show TTL=0 records (goodbye packets)
```

---

## Troubleshooting

### Problem: Service not discoverable from host

**Symptom**: `dns-sd` or `avahi-browse` doesn't show service

**Diagnosis**:
```bash
# Check container is using host network
docker inspect beacon-responder | grep NetworkMode
# Should show: "host"

# Check container logs
docker logs beacon-responder
# Look for "service registered" message

# Check multicast group membership
netstat -g | grep 224.0.0.251
# Should show interface joined to mDNS group
```

**Solution**:
1. Verify `network_mode: "host"` in docker-compose.yml
2. Check firewall allows UDP port 5353
3. Ensure no other mDNS responder conflicts (e.g., Avahi)

### Problem: Container fails to start

**Symptom**: Container exits immediately

**Diagnosis**:
```bash
docker logs beacon-responder
```

**Common Causes**:
- Port 8080 already in use (with host networking)
- Insufficient permissions for network operations
- Application error (check logs)

**Solution**:
```bash
# Change SERVICE_PORT environment variable
docker-compose up -e SERVICE_PORT=8081

# Or modify docker-compose.yml
environment:
  - SERVICE_PORT=8081
```

### Problem: Health check failing

**Symptom**: Container marked unhealthy

**Diagnosis**:
```bash
docker ps  # Check HEALTH status
docker inspect beacon-responder | grep -A 10 Health
```

**Solution**:
- Verify HTTP server started (check logs)
- Increase `start-period` in health check config
- Test health endpoint manually: `curl http://localhost:8080/health`

---

## Security Best Practices

### 1. Run as Non-Root User

```dockerfile
RUN adduser -D -s /bin/sh mdns
USER mdns
```

### 2. Read-Only Root Filesystem (Optional)

```yaml
services:
  beacon-responder:
    read_only: true
    tmpfs:
      - /tmp
```

### 3. Drop Unnecessary Capabilities

```yaml
services:
  beacon-responder:
    cap_drop:
      - ALL
    cap_add:
      - NET_RAW  # Required for raw sockets (mDNS)
```

### 4. Resource Limits

Always set CPU and memory limits to prevent resource exhaustion:

```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 128M
```

---

## Production Considerations

### High Availability

**Don't**: Run multiple responders for the same service on one host
```yaml
# ❌ This will cause name conflicts
services:
  beacon-1:
    # ...
  beacon-2:
    # ... same service name
```

**Do**: Use SO_REUSEPORT (Beacon handles this automatically) or unique instance names
```yaml
# ✅ Different instance names
services:
  beacon-primary:
    environment:
      - SERVICE_NAME=My Service (Primary)
```

### Logging

Use structured JSON logging for production:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
```

Aggregate logs with Docker logging drivers:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### Monitoring

Add Prometheus metrics:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

---

## Next Steps

- [Monitoring Guide](./monitoring.md) - Set up structured logging and metrics
- [Production Checklist](./production-checklist.md) - Pre-deployment validation
- [Troubleshooting Guide](./troubleshooting.md) - Common issues and solutions

## References

- [Docker Networking](https://docs.docker.com/network/) - Docker networking modes
- [RFC 6762 §11](https://www.rfc-editor.org/rfc/rfc6762.html#section-11) - Link-Local Multicast Addresses
- [Beacon README](../../README.md) - Quick start and examples
