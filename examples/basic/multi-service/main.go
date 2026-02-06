// Package main demonstrates registering multiple services simultaneously.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	fmt.Println("=== Multi-Service Registration Example ===\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create single responder for all services
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer r.Close()

	// Define multiple services
	services := []*responder.Service{
		{
			InstanceName: "Web Server",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		},
		{
			InstanceName: "API Server",
			ServiceType:  "_http._tcp.local",
			Port:         8081,
		},
		{
			InstanceName: "SSH Server",
			ServiceType:  "_ssh._tcp.local",
			Port:         22,
		},
	}

	// Register all services
	fmt.Println("Registering services...")
	for i, svc := range services {
		if err := r.Register(svc); err != nil {
			log.Fatalf("Failed to register service %d: %v", i+1, err)
		}
		fmt.Printf("  ✓ %s.%s (port %d)\n", svc.InstanceName, svc.ServiceType, svc.Port)
	}

	fmt.Printf("\n✓ All %d services registered successfully\n", len(services))
	fmt.Println("Press Ctrl+C to exit\n")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down all services...")
}
