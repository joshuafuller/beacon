// Package main demonstrates integrating mDNS service discovery with an HTTP server.
// This example shows how to make a web application discoverable on the local network
// without manual DNS configuration.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

const (
	httpPort     = 8080
	shutdownTime = 30 * time.Second
)

func main() {
	// Create context for responder
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create mDNS responder
	resp, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer resp.Close()

	// Define HTTP service with TXT metadata
	svc := &responder.Service{
		InstanceName: "Web Demo",
		ServiceType:  "_http._tcp.local",
		Port:         httpPort,
		TXTRecords: map[string]string{
			"path":    "/",
			"version": "1.0",
		},
	}

	// Register mDNS service
	if err := resp.Register(svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("mDNS service registered: %s.%s\n", svc.InstanceName, svc.ServiceType)
	fmt.Printf("TXT records: path=/, version=1.0\n")

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(handleRequest),
	}

	// Start HTTP server in background goroutine
	go func() {
		fmt.Printf("HTTP server listening on :%d\n", httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal (Ctrl+C)
	fmt.Println("Press Ctrl+C to stop")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")

	// Graceful HTTP server shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTime)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// mDNS responder closes via defer
	fmt.Println("Shutdown complete")
}

// handleRequest serves HTTP requests with a simple message.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintln(w, "Hello from mDNS-discoverable server!")
	fmt.Fprintf(w, "Discovered via: %s.%s\n", "Web Demo", "_http._tcp.local")
	fmt.Fprintf(w, "Time: %s\n", time.Now().Format(time.RFC3339))
}
