# Error Handling

**Category**: Basic
**Estimated Time**: 10 minutes
**Prerequisites**: Go 1.21+, understanding of Go error handling

## What This Example Demonstrates

This example shows all common error scenarios in Beacon and proper error handling patterns for production code.

**Key Concepts**:
- Validation errors (empty instance name, invalid port)
- Context cancellation handling
- Production-ready error handling patterns
- Defensive programming with Validate()

## Why This Matters

Robust error handling is critical for production services. This example shows how to catch and handle errors properly, log them for debugging, and implement graceful degradation. Understanding these patterns helps you build reliable mDNS services that fail predictably and provide actionable error messages.

## How to Run

### Quick Start

```bash
# Clone the repository
git clone https://github.com/joshuafuller/beacon.git
cd beacon/examples/basic/error-handling

# Run the example
make run
```

### Step-by-Step

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Run the example**:
   ```bash
   make run
   ```

## Expected Output

```
=== Beacon Error Handling Examples ===

Scenario 1: Validation Error (Empty Instance Name)
---
✓ Caught validation error: instance name cannot be empty
  Fix: Set InstanceName to a non-empty value (1-63 characters)

Scenario 2: Validation Error (Invalid Port)
---
✓ Caught validation error: port must be in range 1-65535 (got 0)
  Fix: Set Port to a value between 1 and 65535

Scenario 3: Context Cancellation
---
✓ Caught context cancellation: context canceled
  This is expected when context is cancelled

Scenario 4: Production Pattern (Error Handling Best Practices)
---
✓ Service registered successfully: Production Service._http._tcp.local
  Production pattern:
    1. Context with timeout prevents hanging
    2. Validate before register (fail fast)
    3. Structured logging for observability
    4. Defer cleanup to ensure resources are released

=== All scenarios complete ===
```

**What's Happening**:
1. **Scenario 1**: Catches validation error from empty instance name
2. **Scenario 2**: Catches validation error from invalid port (0)
3. **Scenario 3**: Demonstrates context cancellation error handling
4. **Scenario 4**: Shows production-ready pattern with timeout, validation, and proper cleanup

## Code Walkthrough

### Scenario 1: Validation Error (Empty Instance Name)

**Key Code** (`main.go` lines 32-56):
```go
svc := &responder.Service{
	InstanceName: "", // INVALID: empty string
	ServiceType:  "_http._tcp.local",
	Port:         8080,
}

err = r.Register(svc)
if err != nil {
	fmt.Printf("✓ Caught validation error: %v\n", err)
}
```

**What it demonstrates**:
- `Register()` automatically validates the service
- Empty instance name violates RFC 1035 §2.3.4 (labels must be 1-63 octets)
- Error message is actionable: tells you exactly what's wrong

### Scenario 2: Invalid Port

**Key Code** (`main.go` lines 75-79):
```go
svc := &responder.Service{
	InstanceName: "Test Service",
	ServiceType:  "_http._tcp.local",
	Port:         0, // INVALID: port must be 1-65535
}
```

**What it demonstrates**:
- Port validation catches out-of-range values
- RFC-compliant: ports must be 1-65535
- Fails fast before network operations begin

### Scenario 3: Context Cancellation

**Key Code** (`main.go` lines 94-101):
```go
ctx, cancel := context.WithCancel(context.Background())
cancel() // Cancel immediately

_, err := responder.New(ctx)
if err != nil {
	fmt.Printf("✓ Caught context cancellation: %v\n", err)
}
```

**What it demonstrates**:
- Beacon respects context cancellation
- Useful for graceful shutdown or timeout scenarios
- Standard Go pattern for cancellation

### Scenario 4: Production Pattern

**Key Code** (`main.go` lines 109-152):
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

r, err := responder.New(ctx)
if err != nil {
	log.Printf("ERROR: Failed to create responder: %v", err)
	return
}
defer func() {
	if err := r.Close(); err != nil {
		log.Printf("ERROR: Failed to close responder: %v", err)
	}
}()

// Validate before registering (optional defensive check)
if err := svc.Validate(); err != nil {
	log.Printf("ERROR: Service validation failed: %v", err)
	return
}
```

**What it demonstrates**:
- **Timeout context**: Prevents operations from hanging indefinitely
- **Defensive validation**: Call `Validate()` explicitly for early failure
- **Structured logging**: Use `log.Printf` with ERROR prefix for grep-ability
- **Defer cleanup**: Ensures `Close()` is called even if panics occur
- **Error checking**: Check error from `Close()` - important for goodbye packets

## Error Types in Beacon

### Validation Errors

**Trigger**: Invalid service parameters
**Examples**:
- Empty instance name
- Instance name > 63 characters
- Port < 1 or > 65535
- Invalid service type format

**How to handle**:
```go
if err := r.Register(svc); err != nil {
	// Check if it's a validation error by examining message
	log.Printf("Validation failed: %v", err)
	// Return 400 Bad Request to client
	return
}
```

### Context Errors

**Trigger**: Context cancelled or deadline exceeded
**Examples**:
- `context.Canceled` - user cancelled operation
- `context.DeadlineExceeded` - timeout reached

**How to handle**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := r.Register(svc); err != nil {
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Registration timed out after 30s")
	} else if ctx.Err() == context.Canceled {
		log.Printf("Registration cancelled by user")
	}
	return
}
```

## Production Best Practices

### 1. Always Use Timeouts

```go
// ✓ GOOD: Timeout prevents hanging forever
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// ✗ BAD: No timeout, operation could hang indefinitely
ctx := context.Background()
```

### 2. Validate Before Register

```go
// ✓ GOOD: Validate early for fast failure
if err := svc.Validate(); err != nil {
	return fmt.Errorf("invalid service: %w", err)
}
if err := r.Register(svc); err != nil {
	return fmt.Errorf("registration failed: %w", err)
}

// ✗ BAD: Rely only on Register's internal validation
if err := r.Register(svc); err != nil {
	// Error comes later, harder to debug
}
```

### 3. Always Defer Close()

```go
// ✓ GOOD: Defer ensures goodbye packets are sent
r, err := responder.New(ctx)
if err != nil {
	return err
}
defer r.Close()

// ✗ BAD: Manual close might be skipped if error occurs
r, err := responder.New(ctx)
// ... do work ...
r.Close() // What if work panics? Close never called!
```

### 4. Use Structured Logging

```go
// ✓ GOOD: Structured, searchable logs
log.Printf("ERROR: Failed to register service %s: %v", svc.InstanceName, err)

// ✗ BAD: Unstructured, hard to search
fmt.Println("Error:", err)
```

## Troubleshooting

### Problem: "instance name cannot be empty"
**Symptom**: Registration fails with validation error
**Solution**: Set `InstanceName` to a non-empty value (1-63 characters). Example: `InstanceName: "My Service"`

### Problem: "port must be in range 1-65535"
**Symptom**: Registration fails with port validation error
**Solution**: Set `Port` to a valid TCP/UDP port number. Example: `Port: 8080`

### Problem: "context canceled" or "context deadline exceeded"
**Symptom**: Operations fail with context errors
**Solution**:
- Check if context timeout is too short (increase if needed)
- Verify operation isn't being cancelled prematurely
- For deadline exceeded, consider if network latency is high

## Next Steps

- [Graceful Shutdown](../graceful-shutdown/) - Implement proper cleanup with goodbye packets
- [Multi-Service](../multi-service/) - Register multiple services and handle errors for each
- [Hello Responder](../hello-responder/) - Start with the basics if you haven't already
