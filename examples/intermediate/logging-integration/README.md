# Example 10: Logging Integration

**Difficulty**: Intermediate
**Target Audience**: Production deployments
**Estimated Time**: 15 minutes

## Overview

Demonstrates integrating Beacon with Go's `log/slog` for production observability with structured JSON logging.

## Running

```bash
cd examples/intermediate/logging-integration
LOG_LEVEL=debug LOG_FORMAT=json make run
```

## Environment Variables
- `LOG_LEVEL`: debug, info, warn, error (default: info)
- `LOG_FORMAT`: json, text (default: json)

## Production Patterns
- Structured logging for log aggregation (ELK, Splunk)
- Log level configuration
- Service lifecycle event tracking
