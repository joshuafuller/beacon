// Package main demonstrates discovering mDNS services on the local network.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshuafuller/beacon/querier"
)

func main() {
	fmt.Println("=== Service Browser Example ===")
	fmt.Println("Discovering mDNS services on the local network...\n")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create querier
	q, err := querier.New()
	if err != nil {
		log.Fatalf("Failed to create querier: %v", err)
	}
	defer q.Close()

	// Query for HTTP services (PTR records point to service instances)
	fmt.Println("Searching for HTTP services (_http._tcp.local)...")
	resp, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	if len(resp.Records) == 0 {
		fmt.Println("  No HTTP services found")
	} else {
		fmt.Printf("  Found %d HTTP service(s):\n", len(resp.Records))
		for i, r := range resp.Records {
			ptr := r.AsPTR()
			fmt.Printf("    %d. %s (TTL: %ds)\n", i+1, ptr, r.TTL)
		}
	}

	fmt.Println("\nSearching for SSH services (_ssh._tcp.local)...")
	resp, err = q.Query(ctx, "_ssh._tcp.local", querier.RecordTypePTR)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	if len(resp.Records) == 0 {
		fmt.Println("  No SSH services found")
	} else {
		fmt.Printf("  Found %d SSH service(s):\n", len(resp.Records))
		for i, r := range resp.Records {
			ptr := r.AsPTR()
			fmt.Printf("    %d. %s (TTL: %ds)\n", i+1, ptr, r.TTL)
		}
	}

	fmt.Println("\n=== Discovery complete ===")
}
