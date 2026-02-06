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
