// Package main demonstrates multi-interface mDNS bridging for IoT gateways.
// This example shows how to forward mDNS queries between network interfaces
// to enable service discovery across isolated subnets.
//
// NOTE: This is an educational example demonstrating the concepts.
// Production use requires additional error handling and optimization.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// Interface names (adjust for your system)
	interfaceA = "wlan0" // WiFi
	interfaceB = "eth0"  // Ethernet
)

func main() {
	fmt.Println("=== mDNS Multi-Interface Bridge ===")
	fmt.Println("This example demonstrates RFC 6762 §15 multi-interface operations.\n")

	// Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		// Use defaults if config file doesn't exist
		log.Printf("WARNING: Could not load config.yaml, using defaults: %v", err)
		config = &BridgeConfig{
			Interfaces: []string{interfaceA, interfaceB},
			AllowedServices: []string{
				"_http._tcp",
				"_homekit._tcp",
			},
			ExcludeSubnets: []string{
				"172.17.0.0/16", // Docker
				"10.8.0.0/24",   // VPN
			},
		}
	}

	// Validate configuration
	if len(config.Interfaces) < 2 {
		log.Fatalf("ERROR: Bridge requires at least 2 interfaces (got %d)", len(config.Interfaces))
	}

	fmt.Printf("Bridge configuration:\n")
	fmt.Printf("  Interfaces: %v\n", config.Interfaces)
	fmt.Printf("  Allowed services: %v\n", config.AllowedServices)
	fmt.Printf("  Excluded subnets: %v\n\n", config.ExcludeSubnets)

	// Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create bridge
	bridge, err := NewBridge(config)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Start bridge
	if err := bridge.Start(ctx); err != nil {
		log.Fatalf("Failed to start bridge: %v", err)
	}
	defer bridge.Stop()

	fmt.Printf("✓ Bridge started: %s ↔ %s\n", config.Interfaces[0], config.Interfaces[1])
	fmt.Printf("  Forwarding queries for: %v\n", config.AllowedServices)
	fmt.Println("\nPress Ctrl+C to stop\n")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt
	sig := <-sigChan
	fmt.Printf("\n✓ Received signal: %v\n", sig)
	fmt.Println("Shutting down bridge...")

	// Stop bridge
	cancel()
	time.Sleep(250 * time.Millisecond) // Allow cleanup

	fmt.Println("✓ Bridge stopped")
}
