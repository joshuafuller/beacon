// Package main demonstrates structured logging integration with Beacon.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	// Configure structured logging
	logLevel := getLogLevel(os.Getenv("LOG_LEVEL"))
	logFormat := os.Getenv("LOG_FORMAT")

	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("Starting service", "log_level", logLevel, "format", logFormat)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := responder.New(ctx)
	if err != nil {
		logger.Error("Failed to create responder", "error", err)
		log.Fatal(err)
	}
	defer resp.Close()

	svc := &responder.Service{
		InstanceName: "Logged Service",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords: map[string]string{
			"version": "1.0",
			"status":  "ready",
		},
	}

	logger.Debug("Registering service", "instance", svc.InstanceName, "type", svc.ServiceType, "port", svc.Port)

	if err := resp.Register(svc); err != nil {
		logger.Error("Service registration failed", "error", err, "service", svc.InstanceName)
		log.Fatal(err)
	}

	logger.Info("Service registered successfully",
		"service", svc.InstanceName,
		"type", svc.ServiceType,
		"port", svc.Port,
		"txt_records", svc.TXTRecords)

	fmt.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received")

	// Simulate graceful shutdown logging
	shutdownStart := time.Now()
	logger.Debug("Beginning graceful shutdown")

	time.Sleep(100 * time.Millisecond) // Simulate cleanup

	logger.Info("Shutdown complete", "duration_ms", time.Since(shutdownStart).Milliseconds())
}

func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
