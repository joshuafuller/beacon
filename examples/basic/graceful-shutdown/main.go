// Package main demonstrates graceful shutdown with goodbye packets.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	fmt.Println("=== Graceful Shutdown Example ===")
	fmt.Println("This example demonstrates proper service lifecycle management.")
	fmt.Println("Press Ctrl+C to trigger graceful shutdown with goodbye packets.\n")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create responder
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}

	// Define service
	svc := &responder.Service{
		InstanceName: "Graceful Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register service
	if err := r.Register(svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("✓ Service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Println("  Service is now visible on the network")
	fmt.Println("  Waiting for shutdown signal...\n")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("\n✓ Received signal: %v\n", sig)
	fmt.Println("  Starting graceful shutdown...")

	// Cancel context (stops responder operations)
	cancel()

	// Close responder (sends goodbye packets per RFC 6762 §10.1)
	fmt.Println("  Sending goodbye packets (TTL=0)...")
	if err := r.Close(); err != nil {
		log.Printf("ERROR during close: %v", err)
	}

	// Give time for goodbye packets to be sent
	// RFC 6762 §10.1: Goodbye packets should be sent before process exits
	time.Sleep(250 * time.Millisecond)

	fmt.Println("  ✓ Goodbye packets sent")
	fmt.Println("  ✓ Service removed from network")
	fmt.Println("\n=== Shutdown complete ===")
}
