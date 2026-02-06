// Package main demonstrates dynamic TXT record updates at runtime.
// This example shows how to publish changing service metadata like health status,
// load levels, or feature flags without re-registering the service.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

const (
	serviceName   = "Load Monitor"
	serviceType   = "_http._tcp.local"
	port          = 8080
	updateInterval = 5 * time.Second
	maxUpdates    = 6 // Run for 30 seconds total
)

func main() {
	fmt.Println("=== Dynamic Service Updates Example ===")
	fmt.Println("This example demonstrates RFC 6762 §10.3 dynamic TXT record updates.\n")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create mDNS responder
	resp, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer resp.Close()

	// Register service with initial TXT metadata
	initialLoad := 0
	svc := &responder.Service{
		InstanceName: serviceName,
		ServiceType:  serviceType,
		Port:         port,
		TXTRecords: map[string]string{
			"status":   "healthy",
			"load":     fmt.Sprintf("%d", initialLoad),
			"features": "v1,v2",
		},
	}

	if err := resp.Register(svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("✓ Service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Printf("  Initial TXT: status=healthy, load=%d, features=v1,v2\n\n", initialLoad)
	fmt.Printf("Updating TXT records every %v...\n", updateInterval)
	fmt.Println("Watch updates with: dns-sd -L \"Load Monitor\" _http._tcp\n")

	// Create ticker for periodic updates
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Create channel for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	updateCount := 0

	for {
		select {
		case <-ticker.C:
			updateCount++

			// Simulate load change (random 0-100)
			currentLoad := rand.Intn(101)

			// Update TXT record
			newTXT := map[string]string{
				"status":   "healthy",
				"load":     fmt.Sprintf("%d", currentLoad),
				"features": "v1,v2",
			}

			// Call UpdateService (RFC 6762 §10.3 - triggers reannouncement)
			if err := resp.UpdateService(serviceName, newTXT); err != nil {
				log.Printf("ERROR: Failed to update service: %v", err)
			} else {
				elapsed := updateCount * int(updateInterval.Seconds())
				fmt.Printf("[%2ds] ✓ Updated service: status=healthy, load=%d\n", elapsed, currentLoad)
			}

			// Stop after maxUpdates
			if updateCount >= maxUpdates {
				fmt.Println("\nReached maximum updates, shutting down...")
				cancel()
				return
			}

		case sig := <-sigChan:
			fmt.Printf("\n✓ Received signal: %v\n", sig)
			fmt.Println("Shutting down gracefully...")
			cancel()
			time.Sleep(250 * time.Millisecond) // Allow goodbye packets
			fmt.Println("✓ Shutdown complete")
			return
		}
	}
}
