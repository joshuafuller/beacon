// Package responder implements mDNS service registration and response per RFC 6762.
//
// # Overview
//
// The responder package provides a production-ready mDNS (Multicast DNS) responder
// that registers services on the local network, responds to queries from other devices,
// and handles naming conflicts automatically per RFC 6762 and RFC 6763.
//
// # Quick Start
//
// Register a service on the local network:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//
//	    "github.com/joshuafuller/beacon/responder"
//	)
//
//	func main() {
//	    ctx, cancel := context.WithCancel(context.Background())
//	    defer cancel()
//
//	    r, err := responder.New(ctx)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer r.Close()
//
//	    svc := &responder.Service{
//	        InstanceName: "My Web Server",
//	        ServiceType:  "_http._tcp.local",
//	        Port:         8080,
//	        TXTRecords:   map[string]string{"path": "/", "version": "1.0"},
//	    }
//
//	    if err := r.Register(svc); err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Service is now discoverable via mDNS.
//	    // Block until context is cancelled.
//	    <-ctx.Done()
//	}
//
// # Registration Lifecycle
//
// Register() performs the full RFC 6762 §8 registration sequence:
//
//  1. Probing: Sends 3 probe queries at 250ms intervals to check for name conflicts (~750ms)
//  2. Announcing: Sends 2 announcement packets at 1s intervals to claim the name (~1s)
//  3. Established: Service is registered and the responder answers queries
//
// IMPORTANT: Register() blocks for approximately 1.75 seconds during probing and
// announcing. Use a goroutine if non-blocking behavior is needed:
//
//	go func() {
//	    if err := r.Register(svc); err != nil {
//	        log.Printf("registration failed: %v", err)
//	    }
//	}()
//
// # Conflict Resolution
//
// If another device on the network already claims the same name, the responder
// automatically renames the service per RFC 6762 §9:
//
//   - "My Service" → "My Service-2" → "My Service-3" (up to 10 attempts)
//
// # Interface-Specific Addressing
//
// On multi-interface hosts (e.g., WiFi + Ethernet), the responder detects which
// interface received each query and responds with the correct IP address per
// RFC 6762 §15:
//
//   - Query on WiFi → responds with WiFi IP (e.g., 10.0.0.50)
//   - Query on Ethernet → responds with Ethernet IP (e.g., 192.168.1.100)
//
// # Configuration
//
// Use functional options to customize the responder:
//
//	r, err := responder.New(ctx,
//	    responder.WithHostname("myhost.local"),
//	)
//
// # Resource Management
//
// Always call Close() to send goodbye packets and release resources:
//
//	r, err := responder.New(ctx)
//	if err != nil {
//	    return err
//	}
//	defer r.Close() // Sends TTL=0 goodbye packets per RFC 6762 §10.1
//
// # Thread Safety
//
// All public methods are goroutine-safe. Multiple services can be registered
// concurrently on the same Responder instance.
//
// # RFC Compliance
//
// This implementation follows:
//   - RFC 6762: Multicast DNS (probing, announcing, conflict resolution, goodbye packets)
//   - RFC 6763: DNS-Based Service Discovery (service types, TXT records)
//   - RFC 2782: DNS SRV Records (service location)
package responder
