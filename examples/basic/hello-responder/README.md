# Hello Responder

**Category**: Basic
**Estimated Time**: 5 minutes
**Prerequisites**: Go 1.21+, basic understanding of network services

## What This Example Demonstrates

This example shows the absolute minimum code needed to register a service and make it discoverable on the local network using mDNS.

**Key Concepts**:
- Responder creation and lifecycle
- Service structure (Instance, Service, Domain, Port)
- Graceful shutdown pattern

## Why This Matters

mDNS (Multicast DNS) enables zero-configuration service discovery on local networks. This is perfect for IoT devices, local development tools, and microservices that need to discover each other without centralized configuration. Instead of hardcoding IP addresses, clients can find your service by name.

## How to Run

### Quick Start

```bash
# Clone the repository
git clone https://github.com/joshuafuller/beacon.git
cd beacon/examples/basic/hello-responder

# Run the example
make run
```

### Step-by-Step

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build the example**:
   ```bash
   make build
   ```

3. **Run the example**:
   ```bash
   ./bin/hello-responder
   # OR
   make run
   ```

4. **Verify service is visible**:
   ```bash
   # macOS:
   dns-sd -B _http._tcp

   # Linux:
   avahi-browse -t _http._tcp

   # Expected: "Hello World._http._tcp.local" appears in list
   ```

## Expected Output

```
Service registered: Hello World._http._tcp.local
Press Ctrl+C to exit
```

**What's Happening**:
1. Beacon creates a responder that listens on the mDNS multicast address (224.0.0.251:5353)
2. The service undergoes RFC 6762 §8.1 probing to detect name conflicts
3. After successful probing, RFC 6762 §8.3 announcing broadcasts the service to the network
4. Other devices can now discover "Hello World._http._tcp.local" at port 8080

## Code Walkthrough

### Key Parts

**1. Responder Creation** (`main.go` lines 17-23):
```go
// Create responder
r, err := responder.New(ctx)
if err != nil {
	log.Fatalf("Failed to create responder: %v", err)
}
defer r.Close()
```
Creates the responder with default configuration. The `defer r.Close()` ensures goodbye packets are sent when the program exits.

**2. Service Definition** (`main.go` lines 27-32):
```go
// Define service
svc := &responder.Service{
	InstanceName: "Hello World",
	ServiceType:  "_http._tcp.local",
	Port:         8080,
}
```
Defines the service metadata:
- **InstanceName**: Human-readable name (e.g., "Hello World")
- **ServiceType**: Service type including domain following RFC 6763 (e.g., "_http._tcp.local" for HTTP)
- **Port**: TCP/UDP port where your service listens

**3. Service Registration** (`main.go` lines 34-37):
```go
// Register service
if err := r.Register(svc); err != nil {
	log.Fatalf("Failed to register service: %v", err)
}
```
Registers the service, triggering probing and announcing. This is a non-blocking call - the service becomes visible within ~750ms (RFC 6762 §8.1 probing takes 250ms × 3 probes).

## Troubleshooting

### Problem: "Failed to create responder: permission denied"
**Symptom**: Error on startup about binding to multicast address
**Solution**: On Linux, you may need to run with elevated privileges or configure firewall rules to allow multicast traffic on 224.0.0.251:5353

### Problem: Service not visible in dns-sd/avahi-browse
**Symptom**: Command runs but "Hello World" doesn't appear
**Solution**:
1. Check firewall - ensure UDP port 5353 is open
2. Verify multicast routing - run `ip mroute show` (Linux) to check multicast routes
3. Wait 1-2 seconds - probing/announcing takes ~750ms
4. Try on same subnet - mDNS doesn't cross routers by default

### Problem: "Failed to register service: service already exists"
**Symptom**: Error during registration
**Solution**: Another service with the name "Hello World._http._tcp.local" already exists on your network. Change the `InstanceName` field to a unique name.

## Next Steps

- [Error Handling](../error-handling/) - Learn how to handle validation and network errors
- [Graceful Shutdown](../graceful-shutdown/) - Implement proper cleanup with goodbye packets
- [Multi-Service](../multi-service/) - Register multiple services simultaneously
