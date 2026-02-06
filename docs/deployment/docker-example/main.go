package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	// Structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Read environment variables
	serviceName := getEnv("SERVICE_NAME", "Docker Service")
	servicePort := getEnv("SERVICE_PORT", "8080")

	// Create responder
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r, err := responder.New(ctx)
	if err != nil {
		logger.Error("failed to create responder", "error", err)
		os.Exit(1)
	}
	defer r.Close()

	// Register service
	svc := &responder.Service{
		InstanceName: serviceName,
		ServiceType:  "_http._tcp.local",
		Port:         parsePort(servicePort),
		TXTRecords: map[string]string{
			"version": "1.0",
			"env":     "docker",
		},
	}

	if err := r.Register(svc); err != nil {
		logger.Error("failed to register service", "error", err)
		os.Exit(1)
	}

	logger.Info("service registered",
		"instance", svc.InstanceName,
		"service", svc.ServiceType,
		"port", svc.Port,
	)

	// Start HTTP server
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from Beacon Docker service!\n")
	})

	go func() {
		addr := ":" + servicePort
		logger.Info("starting HTTP server", "addr", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	logger.Info("shutting down gracefully")
	r.Close() // Sends goodbye packets (TTL=0)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parsePort(portStr string) int {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	if port < 1 || port > 65535 {
		return 8080
	}
	return port
}
