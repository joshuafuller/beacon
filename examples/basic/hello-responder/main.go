// Package main demonstrates minimal service registration with Beacon.
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
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create responder
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer r.Close()

	// Define service
	svc := &responder.Service{
		InstanceName: "Hello World",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	// Register service
	if err := r.Register(svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("Service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
