package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	fmt.Println("=== Multi-Interface mDNS Demonstration ===")
	fmt.Println("This example demonstrates RFC 6762 §15 interface-specific IP addressing")
	fmt.Println()

	// Step 1: Show available network interfaces
	fmt.Println("📡 Available Network Interfaces:")
	interfaces, err := getActiveInterfaces()
	if err != nil {
		fmt.Printf("❌ Error getting interfaces: %v\n", err)
		os.Exit(1)
	}

	if len(interfaces) < 2 {
		fmt.Printf("⚠️  Only %d interface(s) found. This demo works best with 2+ interfaces.\n", len(interfaces))
		fmt.Println("   (e.g., WiFi + Ethernet, or Ethernet + Docker)")
		fmt.Println()
	}

	for _, iface := range interfaces {
		fmt.Printf("  • %s (index %d)\n", iface.Name, iface.Index)
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip interfaces where we can't get addresses
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				fmt.Printf("    → IPv4: %s\n", ipnet.IP)
			}
		}
	}
	fmt.Println()

	// Step 2: Create and register a test service
	fmt.Println("🚀 Starting mDNS Responder...")

	service := &responder.Service{
		InstanceName: "Multi-Interface Demo",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords: map[string]string{
			"version": "1.0",
			"demo":    "interface-specific-addressing",
		},
	}

	ctx := context.Background()
	resp, err := responder.New(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to create responder: %v\n", err)
		os.Exit(1)
	}
	defer resp.Close()

	err = resp.Register(service)
	if err != nil {
		fmt.Printf("❌ Failed to register service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Service registered: %s.%s\n", service.InstanceName, service.ServiceType)
	fmt.Println()

	// Step 3: Explain what's happening
	fmt.Println("📋 What's Happening (RFC 6762 §15 Compliance):")
	fmt.Println()
	fmt.Println("The responder is now listening on ALL interfaces (0.0.0.0:5353).")
	fmt.Println("When a query arrives on a specific interface, the responder will:")
	fmt.Println()
	fmt.Println("  1. Extract the interface index from the IP_PKTINFO control message")
	fmt.Println("  2. Resolve the IPv4 address for THAT specific interface")
	fmt.Println("  3. Respond with ONLY that interface's IP address")
	fmt.Println()
	fmt.Println("This ensures clients can actually REACH the advertised IP address!")
	fmt.Println()

	// Step 4: Show expected behavior per interface
	fmt.Println("🔍 Expected Behavior:")
	fmt.Println()
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip interfaces where we can't get addresses
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				fmt.Printf("  Query on %s (index %d):\n", iface.Name, iface.Index)
				fmt.Printf("    → Response will advertise: %s\n", ipnet.IP)
				fmt.Printf("    → Clients on %s network can connect to %s:8080 ✅\n", iface.Name, ipnet.IP)
				fmt.Println()
			}
		}
	}

	// Step 5: Provide testing instructions
	fmt.Println("🧪 How to Test:")
	fmt.Println()
	fmt.Println("1. Keep this program running")
	fmt.Println("2. On the SAME machine, open another terminal and run:")
	fmt.Println()
	fmt.Println("   # Query for the service")
	fmt.Println("   avahi-browse -r _http._tcp")
	fmt.Println()
	fmt.Println("   # Or use dig to query the A record")
	fmt.Println("   dig @224.0.0.251 -p 5353 \"Multi-Interface Demo._http._tcp.local\" A")
	fmt.Println()
	fmt.Println("3. On a DIFFERENT machine on the same network:")
	fmt.Println()
	fmt.Println("   avahi-browse -r _http._tcp")
	fmt.Println()
	fmt.Println("4. Observe that each interface advertises its OWN IP address!")
	fmt.Println()

	// Step 6: Wait for Ctrl+C
	fmt.Println("✋ Press Ctrl+C to stop the responder...")
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("👋 Shutting down...")

	// Unregister service (sends goodbye)
	err = resp.Unregister(service.InstanceName)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to unregister service: %v\n", err)
	} else {
		fmt.Println("✅ Service unregistered (goodbye sent)")
	}
}

// getActiveInterfaces returns all active non-loopback interfaces with IPv4 addresses
func getActiveInterfaces() ([]net.Interface, error) {
	allIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var active []net.Interface
	for _, iface := range allIfaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Check if it has an IPv4 address
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}

		if hasIPv4 {
			active = append(active, iface)
		}
	}

	return active, nil
}
