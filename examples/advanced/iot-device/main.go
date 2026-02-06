// Package main demonstrates Beacon running on an IoT device (Raspberry Pi).
//
// This example shows:
//   - Hardware capability detection (GPIO, camera, sensors)
//   - Multi-service registration based on available hardware
//   - Graceful shutdown with signal handling
//   - Resource-efficient operation suitable for constrained devices
//
// Usage:
//
//	sudo go run main.go
//
// The example registers multiple services depending on detected hardware:
//   - _device-info._tcp: Device metadata (always)
//   - _gpio._tcp: GPIO controller (if GPIO available)
//   - _http._tcp: Camera stream (if camera available)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joshuafuller/beacon/responder"
)

// HardwareCapabilities represents detected hardware on the device.
type HardwareCapabilities struct {
	HasGPIO   bool
	HasCamera bool
	ModelName string
	SerialNum string
}

func main() {
	log.Println("IoT Device mDNS Registration Example")
	log.Println("=====================================")

	// Detect hardware capabilities
	log.Println("Detecting hardware capabilities...")
	caps := detectHardware()

	log.Printf("Hardware: %s", caps.ModelName)
	if caps.HasGPIO {
		log.Println("  - GPIO: available")
	} else {
		log.Println("  - GPIO: not detected")
	}
	if caps.HasCamera {
		log.Println("  - Camera: available")
	} else {
		log.Println("  - Camera: not detected")
	}
	log.Printf("  - Serial: %s", caps.SerialNum)

	// Build service list based on hardware
	services := buildServices(caps)
	log.Printf("Prepared %d services for registration", len(services))

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start responder
	log.Println("Starting mDNS responder...")
	r, err := responder.New(context.Background())
	if err != nil {
		log.Fatalf("Failed to start responder: %v", err)
	}

	// Register services
	for _, svc := range services {
		if err := r.Register(&svc); err != nil {
			log.Fatalf("Failed to register service %s: %v", svc.InstanceName, err)
		}
	}

	log.Printf("✓ Registered %d services successfully", len(services))
	log.Println("Services are now discoverable via mDNS")
	log.Println("Press Ctrl+C to shutdown gracefully...")

	// Block until signal received
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)
	log.Println("Initiating graceful shutdown...")

	// Graceful shutdown
	if err := r.Close(); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("✓ Shutdown complete")
}

// detectHardware detects available hardware capabilities on the device.
func detectHardware() HardwareCapabilities {
	caps := HardwareCapabilities{}

	// Check for GPIO access (Raspberry Pi specific)
	// On Raspberry Pi, /dev/gpiomem provides GPIO access
	if _, err := os.Stat("/dev/gpiomem"); err == nil {
		caps.HasGPIO = true
	}

	// Alternative: Check for /sys/class/gpio (more generic)
	if !caps.HasGPIO {
		if _, err := os.Stat("/sys/class/gpio"); err == nil {
			caps.HasGPIO = true
		}
	}

	// Check for camera device (Video4Linux)
	// Raspberry Pi Camera appears as /dev/video0
	if _, err := os.Stat("/dev/video0"); err == nil {
		caps.HasCamera = true
	}

	// Read Raspberry Pi model from device tree
	// This file contains the full model name (e.g., "Raspberry Pi 3 Model B Plus Rev 1.3")
	if data, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		caps.ModelName = strings.TrimSpace(string(data))
		// Remove null terminator if present
		caps.ModelName = strings.TrimRight(caps.ModelName, "\x00")
	} else {
		// Fallback to hostname if model not available
		hostname, _ := os.Hostname()
		caps.ModelName = hostname
	}

	// Read serial number from cpuinfo
	caps.SerialNum = getSerialNumber()

	return caps
}

// getSerialNumber extracts the device serial number from /proc/cpuinfo.
func getSerialNumber() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}

	// Parse cpuinfo for Serial line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Serial") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				serial := strings.TrimSpace(parts[1])
				// Remove leading zeros for cleaner display
				serial = strings.TrimLeft(serial, "0")
				if serial == "" {
					return "000000000000"
				}
				return serial
			}
		}
	}

	return "unknown"
}

// buildServices creates service definitions based on detected hardware.
func buildServices(caps HardwareCapabilities) []responder.Service {
	services := []responder.Service{}

	// Always register device info service
	// This provides basic device metadata for discovery
	deviceName := fmt.Sprintf("Pi-%s", caps.SerialNum[:min(4, len(caps.SerialNum))])
	services = append(services, responder.Service{
		InstanceName: deviceName,
		ServiceType:  "_device-info._tcp.local.",
		Port:         8080, // Device info HTTP endpoint
		TXTRecords: map[string]string{
			"model":  caps.ModelName,
			"serial": caps.SerialNum,
			"os":     "Linux",
			"type":   "iot",
		},
	})

	// Register GPIO service if available
	if caps.HasGPIO {
		services = append(services, responder.Service{
			InstanceName: "Pi GPIO Controller",
			ServiceType:  "_gpio._tcp.local.",
			Port:         8081, // GPIO REST API endpoint
			TXTRecords: map[string]string{
				"pins":     "40",   // Raspberry Pi has 40 GPIO pins
				"protocol": "REST", // REST API for GPIO control
				"path":     "/gpio",
				"version":  "1.0",
			},
		})
	}

	// Register camera service if available
	if caps.HasCamera {
		services = append(services, responder.Service{
			InstanceName: "Pi Camera Stream",
			ServiceType:  "_http._tcp.local.",
			Port:         8082, // Camera MJPEG stream endpoint
			TXTRecords: map[string]string{
				"path":       "/stream.mjpg",
				"resolution": "1920x1080",
				"format":     "mjpeg",
			},
		})
	}

	return services
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
