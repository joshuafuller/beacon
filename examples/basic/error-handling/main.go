// Package main demonstrates error handling patterns in Beacon.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	fmt.Println("=== Beacon Error Handling Examples ===\n")

	// Scenario 1: Validation Error - Empty Instance Name
	scenario1_ValidationError()

	// Scenario 2: Validation Error - Invalid Port
	scenario2_InvalidPort()

	// Scenario 3: Context Cancellation
	scenario3_ContextCancellation()

	// Scenario 4: Production Pattern - Proper Error Handling
	scenario4_ProductionPattern()

	fmt.Println("\n=== All scenarios complete ===")
}

// Scenario 1: Validation Error - Empty Instance Name
func scenario1_ValidationError() {
	fmt.Println("Scenario 1: Validation Error (Empty Instance Name)")
	fmt.Println("---")

	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to create responder: %v\n", err)
		return
	}
	defer r.Close()

	// Invalid service: empty instance name
	svc := &responder.Service{
		InstanceName: "", // INVALID: empty string
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register will fail validation
	err = r.Register(svc)
	if err != nil {
		fmt.Printf("✓ Caught validation error: %v\n", err)
		fmt.Printf("  Fix: Set InstanceName to a non-empty value (1-63 characters)\n\n")
	} else {
		fmt.Println("✗ Expected error but registration succeeded\n")
	}
}

// Scenario 2: Validation Error - Invalid Port
func scenario2_InvalidPort() {
	fmt.Println("Scenario 2: Validation Error (Invalid Port)")
	fmt.Println("---")

	ctx := context.Background()
	r, err := responder.New(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to create responder: %v\n", err)
		return
	}
	defer r.Close()

	// Invalid service: port out of range
	svc := &responder.Service{
		InstanceName: "Test Service",
		ServiceType:  "_http._tcp.local",
		Port:         0, // INVALID: port must be 1-65535
	}

	err = r.Register(svc)
	if err != nil {
		fmt.Printf("✓ Caught validation error: %v\n", err)
		fmt.Printf("  Fix: Set Port to a value between 1 and 65535\n\n")
	} else {
		fmt.Println("✗ Expected error but registration succeeded\n")
	}
}

// Scenario 3: Context Cancellation
func scenario3_ContextCancellation() {
	fmt.Println("Scenario 3: Context Cancellation")
	fmt.Println("---")

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Try to create responder with cancelled context
	_, err := responder.New(ctx)
	if err != nil {
		fmt.Printf("✓ Caught context cancellation: %v\n", err)
		fmt.Printf("  This is expected when context is cancelled\n\n")
	} else {
		fmt.Println("✗ Expected context error but operation succeeded\n")
	}
}

// Scenario 4: Production Pattern - Proper Error Handling
func scenario4_ProductionPattern() {
	fmt.Println("Scenario 4: Production Pattern (Error Handling Best Practices)")
	fmt.Println("---")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create responder with error handling
	r, err := responder.New(ctx)
	if err != nil {
		// In production, use structured logging
		log.Printf("ERROR: Failed to create responder: %v", err)
		fmt.Printf("  Production action: Log error, return 500, trigger alert\n\n")
		return
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("ERROR: Failed to close responder: %v", err)
		}
	}()

	// Define valid service
	svc := &responder.Service{
		InstanceName: "Production Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Validate before registering (optional defensive check)
	if err := svc.Validate(); err != nil {
		log.Printf("ERROR: Service validation failed: %v", err)
		fmt.Printf("  Production action: Return 400 Bad Request to client\n\n")
		return
	}

	// Register with error handling
	if err := r.Register(svc); err != nil {
		log.Printf("ERROR: Failed to register service: %v", err)
		fmt.Printf("  Production action: Log error, retry with backoff, or fail gracefully\n\n")
		return
	}

	fmt.Printf("✓ Service registered successfully: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Printf("  Production pattern:\n")
	fmt.Printf("    1. Context with timeout prevents hanging\n")
	fmt.Printf("    2. Validate before register (fail fast)\n")
	fmt.Printf("    3. Structured logging for observability\n")
	fmt.Printf("    4. Defer cleanup to ensure resources are released\n\n")
}
