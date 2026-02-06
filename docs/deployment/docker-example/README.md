# Docker Deployment Example

This example demonstrates deploying a Beacon-based mDNS service in Docker.

## Quick Start

```bash
# Build and run
docker-compose up --build

# Test from host
dns-sd -B _http._tcp local          # macOS
avahi-browse -t _http._tcp          # Linux

# Access HTTP endpoint
curl http://localhost:8080
```

## Key Points

1. **network_mode: "host"** is required for multicast to work
2. Health check verifies service is running
3. Graceful shutdown sends goodbye packets (TTL=0)
4. Structured JSON logging for production

## Files

- `Dockerfile` - Multi-stage build for minimal image
- `docker-compose.yml` - Production configuration with host networking
- `main.go` - Example Beacon responder with HTTP server
- `go.mod` - Go module configuration

## See Also

- [Docker Deployment Guide](../docker.md) - Complete guide with troubleshooting
- [Production Checklist](../production-checklist.md) - Pre-deployment validation
