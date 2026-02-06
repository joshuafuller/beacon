# Monitoring Guide

**Category**: Deployment
**Estimated Time**: 20 minutes
**Prerequisites**: Production Beacon service deployed

## Overview

This guide demonstrates how to monitor Beacon-based mDNS services using structured logging, metrics collection, health checks, and alerting strategies.

---

## Structured Logging with slog

### Basic Setup

```go
package main

import (
	"log/slog"
	"os"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	// Create JSON logger for production
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Set as default logger
	slog.SetDefault(logger)

	// Use throughout application
	logger.Info("service starting",
		"version", "1.0.0",
		"environment", "production",
	)

	// ... responder setup ...
}
```

### Log Schema

**Recommended fields for all log entries**:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `time` | RFC3339 | Event timestamp | `2026-01-06T12:34:56Z` |
| `level` | string | Log level | `INFO`, `WARN`, `ERROR` |
| `msg` | string | Human-readable message | `service registered` |
| `instance` | string | Service instance name | `My API Server` |
| `service` | string | Service type | `_http._tcp.local` |
| `port` | int | Service port | `8080` |
| `error` | string | Error message (if applicable) | `port in use` |

### Key Events to Log

#### 1. Service Registration

```go
logger.Info("service registered",
	"instance", svc.InstanceName,
	"service", svc.ServiceType,
	"port", svc.Port,
	"hostname", svc.Hostname,
	"txt_count", len(svc.TXTRecords),
)
```

Expected output:
```json
{
  "time": "2026-01-06T12:34:56Z",
  "level": "INFO",
  "msg": "service registered",
  "instance": "My API Server",
  "service": "_http._tcp.local",
  "port": 8080,
  "hostname": "server.local",
  "txt_count": 2
}
```

#### 2. Service Updates

```go
logger.Info("service updated",
	"service_id", serviceID,
	"txt_records_updated", len(newTXT),
	"old_txt_count", len(oldTXT),
)
```

#### 3. Errors

```go
if err := r.Register(svc); err != nil {
	logger.Error("service registration failed",
		"instance", svc.InstanceName,
		"error", err.Error(),
		"port", svc.Port,
	)
	return err
}
```

Expected output:
```json
{
  "time": "2026-01-06T12:35:00Z",
  "level": "ERROR",
  "msg": "service registration failed",
  "instance": "My API Server",
  "error": "port must be in range 1-65535 (got 0)",
  "port": 0
}
```

#### 4. Network Events

```go
logger.Warn("multicast join failed",
	"interface", "eth0",
	"multicast_addr", "224.0.0.251",
	"error", err.Error(),
)
```

#### 5. Shutdown

```go
logger.Info("shutting down",
	"active_services", len(activeServices),
	"uptime_seconds", time.Since(startTime).Seconds(),
)

r.Close()

logger.Info("shutdown complete",
	"goodbye_packets_sent", goodbyeCount,
)
```

---

## Health Checks

### HTTP Health Endpoint

```go
import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

type HealthStatus struct {
	Status           string `json:"status"`
	ActiveServices   int    `json:"active_services"`
	UptimeSeconds    int    `json:"uptime_seconds"`
	LastProbeSuccess bool   `json:"last_probe_success"`
}

var (
	startTime = time.Now()
	serviceCount int32
	lastProbeOK int32
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:           "healthy",
		ActiveServices:   int(atomic.LoadInt32(&serviceCount)),
		UptimeSeconds:    int(time.Since(startTime).Seconds()),
		LastProbeSuccess: atomic.LoadInt32(&lastProbeOK) == 1,
	}

	// Mark unhealthy if no services active
	if status.ActiveServices == 0 {
		status.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readinessHandler) // Kubernetes readiness
	http.HandleFunc("/live", livenessHandler)   // Kubernetes liveness

	go http.ListenAndServe(":8080", nil)
}
```

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
  timeoutSeconds: 3
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 2
```

---

## Metrics Collection

### Prometheus Metrics (Optional)

```go
import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	servicesRegistered = promauto.NewCounter(prometheus.CounterOpts{
		Name: "beacon_services_registered_total",
		Help: "Total number of services registered",
	})

	servicesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "beacon_services_active",
		Help: "Current number of active services",
	})

	registrationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "beacon_registration_errors_total",
		Help: "Total number of registration errors",
	})

	queryLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "beacon_query_duration_seconds",
		Help:    "Query response latency",
		Buckets: prometheus.DefBuckets,
	})
)

func main() {
	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9090", nil)

	// Increment metrics
	servicesRegistered.Inc()
	servicesActive.Set(float64(len(activeServices)))

	// Record latency
	timer := prometheus.NewTimer(queryLatency)
	defer timer.ObserveDuration()
}
```

### Key Metrics to Track

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| `beacon_services_registered_total` | Counter | Total services registered | N/A |
| `beacon_services_active` | Gauge | Current active services | `< 1` (no services) |
| `beacon_registration_errors_total` | Counter | Registration failures | `> 0` (any errors) |
| `beacon_query_duration_seconds` | Histogram | Query latency | `p99 > 1s` |
| `beacon_network_errors_total` | Counter | Network errors | `> 10/min` |
| `beacon_goodbye_packets_sent_total` | Counter | Shutdown cleanness | N/A |

---

## Log Aggregation

### ELK Stack (Elasticsearch, Logstash, Kibana)

**Docker Compose**:
```yaml
services:
  beacon-service:
    # ...
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service,environment"

  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
```

**Logstash Config**:
```
input {
  file {
    path => "/var/lib/docker/containers/*/*.log"
    codec => json
  }
}

filter {
  if [msg] {
    mutate {
      add_field => { "event_type" => "%{msg}" }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "beacon-logs-%{+YYYY.MM.dd}"
  }
}
```

### Loki + Grafana

**Promtail Config**:
```yaml
scrape_configs:
  - job_name: beacon
    static_configs:
      - targets:
          - localhost
        labels:
          job: beacon-service
          __path__: /var/log/beacon/*.log
```

**Grafana Dashboard Query**:
```logql
{job="beacon-service"} |= "service registered"
```

---

## Alerting

### Alert Conditions

#### Critical Alerts

1. **No Active Services**
   ```yaml
   alert: NoActiveServices
   expr: beacon_services_active == 0
   for: 5m
   labels:
     severity: critical
   annotations:
     summary: "Beacon service has no active registrations"
   ```

2. **High Error Rate**
   ```yaml
   alert: HighErrorRate
   expr: rate(beacon_registration_errors_total[5m]) > 0.1
   for: 2m
   labels:
     severity: critical
   annotations:
     summary: "Beacon registration error rate > 10%"
   ```

3. **Service Down**
   ```yaml
   alert: BeaconServiceDown
   expr: up{job="beacon"} == 0
   for: 1m
   labels:
     severity: critical
   annotations:
     summary: "Beacon service is down"
   ```

#### Warning Alerts

1. **High Query Latency**
   ```yaml
   alert: HighQueryLatency
   expr: histogram_quantile(0.99, beacon_query_duration_seconds) > 1
   for: 10m
   labels:
     severity: warning
   annotations:
     summary: "p99 query latency > 1s"
   ```

2. **Network Errors**
   ```yaml
   alert: NetworkErrors
   expr: rate(beacon_network_errors_total[5m]) > 1
   for: 5m
   labels:
     severity: warning
   annotations:
     summary: "Elevated network error rate"
   ```

---

## Dashboard Examples

### Grafana Dashboard (JSON)

```json
{
  "dashboard": {
    "title": "Beacon mDNS Monitoring",
    "panels": [
      {
        "title": "Active Services",
        "targets": [{
          "expr": "beacon_services_active"
        }],
        "type": "stat"
      },
      {
        "title": "Registration Rate",
        "targets": [{
          "expr": "rate(beacon_services_registered_total[5m])"
        }],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(beacon_registration_errors_total[5m])"
        }],
        "type": "graph"
      },
      {
        "title": "Query Latency (p99)",
        "targets": [{
          "expr": "histogram_quantile(0.99, beacon_query_duration_seconds)"
        }],
        "type": "graph"
      }
    ]
  }
}
```

---

## Best Practices

### 1. Log Sampling for High-Volume Events

```go
var logCounter atomic.Int64

func logWithSampling(logger *slog.Logger, msg string, attrs ...any) {
	count := logCounter.Add(1)
	// Log every 100th event
	if count%100 == 0 {
		logger.Info(msg, append(attrs, "sampled_count", count)...)
	}
}
```

### 2. Context Propagation

```go
type contextKey string

const requestIDKey = contextKey("request_id")

func logWithContext(ctx context.Context, msg string, attrs ...any) {
	if reqID := ctx.Value(requestIDKey); reqID != nil {
		attrs = append(attrs, "request_id", reqID)
	}
	logger.Info(msg, attrs...)
}
```

### 3. Error Correlation

```go
import "github.com/google/uuid"

func registerWithTracking(svc *responder.Service) error {
	traceID := uuid.New().String()

	logger := logger.With("trace_id", traceID)
	logger.Info("registration starting", "instance", svc.InstanceName)

	err := r.Register(svc)
	if err != nil {
		logger.Error("registration failed", "error", err)
		return fmt.Errorf("[%s] %w", traceID, err)
	}

	logger.Info("registration complete")
	return nil
}
```

---

## Troubleshooting Monitoring Issues

### Logs Not Appearing

**Check**:
- Logger is configured correctly (`slog.SetDefault()`)
- Output is JSON (some log aggregators require JSON)
- Permissions on log files
- Log driver configuration (Docker)

### Metrics Not Updating

**Check**:
- Prometheus scrape configuration
- Firewall allows port 9090
- `/metrics` endpoint accessible: `curl http://localhost:9090/metrics`

### Alerts Not Firing

**Check**:
- AlertManager configuration
- Alert rules syntax
- Notification channels (email, Slack, PagerDuty)
- Alert rule evaluation interval

---

## Next Steps

- [Troubleshooting Guide](./troubleshooting.md) - Diagnose and fix common issues
- [Production Checklist](./production-checklist.md) - Pre-deployment validation
- [Docker Deployment](./docker.md) - Container deployment guide

## References

- [slog Package](https://pkg.go.dev/log/slog) - Go structured logging
- [Prometheus Best Practices](https://prometheus.io/docs/practices/) - Metrics guidelines
- [RFC 6762](https://www.rfc-editor.org/rfc/rfc6762.html) - Multicast DNS specification
