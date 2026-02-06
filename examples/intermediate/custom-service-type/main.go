// Package main demonstrates custom mDNS service type registration.
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer resp.Close()

	// Define custom service type: _myapp._tcp
	svc := &responder.Service{
		InstanceName: "My App Instance",
		ServiceType:  "_myapp._tcp.local",
		Port:         9000,
		TXTRecords: map[string]string{
			"api_version": "2.0",
			"protocol":    "custom",
			"features":    "feature1,feature2",
			"endpoint":    "/api/v2",
		},
	}

	if err := resp.Register(svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("Custom service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Printf("TXT records: api_version=2.0, protocol=custom, features=feature1,feature2\n")
	fmt.Println("Discover with: dns-sd -B _myapp._tcp")
	fmt.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
