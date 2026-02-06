package responder_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

// ExampleResponder_Register demonstrates basic service registration.
// This example shows how to create a responder and register a simple HTTP service.
func ExampleResponder_Register() {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create responder
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Define service
	svc := &responder.Service{
		InstanceName: "My Web Server",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords: map[string]string{
			"path":    "/",
			"version": "1.0",
		},
	}

	// Register service (probing and announcing happens automatically)
	if err := r.Register(svc); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	// Output:
	// Service registered: My Web Server._http._tcp.local
}

// ExampleResponder_Unregister demonstrates how to cleanly unregister a service.
// This sends goodbye packets per RFC 6762 §10.1.
func ExampleResponder_Unregister() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := responder.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Register a service first
	svc := &responder.Service{
		InstanceName: "Temporary Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	if err := r.Register(svc); err != nil {
		log.Fatal(err)
	}

	// Unregister the service (sends goodbye packets with TTL=0)
	serviceID := "Temporary Service._http._tcp.local"
	if err := r.Unregister(serviceID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Service unregistered successfully")
	// Output:
	// Service unregistered successfully
}

// ExampleResponder_UpdateService demonstrates dynamic TXT record updates.
// This allows updating service metadata without re-registration.
func ExampleResponder_UpdateService() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := responder.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Register initial service
	svc := &responder.Service{
		InstanceName: "My API",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords: map[string]string{
			"version": "1.0",
			"status":  "starting",
		},
	}

	if err := r.Register(svc); err != nil {
		log.Fatal(err)
	}

	// Update TXT records (e.g., after startup completes)
	newTXT := map[string]string{
		"version": "1.0",
		"status":  "ready",
		"load":    "low",
	}

	serviceID := "My API._http._tcp.local"
	if err := r.UpdateService(serviceID, newTXT); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Service TXT records updated")
	// Output:
	// Service TXT records updated
}

// ExampleService_Validate demonstrates service validation before registration.
// Validation checks service type format, port range, and instance name constraints.
func ExampleService_Validate() {
	// Valid service
	validService := &responder.Service{
		InstanceName: "My Printer",
		ServiceType:  "_ipp._tcp.local",
		Port:         631,
		TXTRecords: map[string]string{
			"ty":       "HP LaserJet",
			"priority": "high",
		},
	}

	if err := validService.Validate(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Valid service")

	// Invalid service (bad port)
	invalidService := &responder.Service{
		InstanceName: "Bad Service",
		ServiceType:  "_http._tcp.local",
		Port:         0, // Invalid: port must be 1-65535
	}

	if err := invalidService.Validate(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	// Output:
	// Valid service
	// Validation error: port must be in range 1-65535 (got 0)
}

// ExampleNew demonstrates creating a responder with default options.
func ExampleNew() {
	ctx := context.Background()

	// Create responder with default settings
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	fmt.Println("Responder created successfully")
	// Output:
	// Responder created successfully
}

// ExampleNew_withOptions demonstrates creating a responder with custom options.
func ExampleNew_withOptions() {
	ctx := context.Background()

	// Create responder with custom hostname
	r, err := responder.New(ctx, responder.WithHostname("myserver.local"))
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	fmt.Println("Responder created with custom hostname")
	// Output:
	// Responder created with custom hostname
}
