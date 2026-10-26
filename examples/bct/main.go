// Package main is the device-under-test program for Apple's Bonjour
// Conformance Test (BCT).
//
// BCT runs on a Mac and tests a separate device on the same network: it
// probes, queries, injects name conflicts and checks the replies against
// RFC 6762 / RFC 6763. This program runs on that device and keeps one service
// registered with Beacon until it is stopped. See README.md for the runbook.
//
// Usage:
//
//	go run ./examples/bct -host beacon-bct -name "Beacon BCT" -type _http._tcp.local -port 8080
//
// Probes, announcements and a rename forced during registration (RFC 6762
// §9) are printed, so the log can be matched against the BCT results file.
package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	host := flag.String("host", "beacon-bct", "host name label (\".local\" is added)")
	name := flag.String("name", "Beacon BCT", "service instance name")
	svcType := flag.String("type", "_http._tcp.local", "service type")
	port := flag.Uint("port", 8080, "service port")
	ipv6 := flag.Bool("ipv6", false, "also answer on IPv6 (ff02::fb)")
	flag.Parse()

	if *port == 0 || *port > 65535 {
		log.Fatalf("invalid -port %d", *port)
	}

	// The responder gets its own context: if it shared the signal context,
	// Ctrl-C would cancel it before Close sends the goodbye packets.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := []responder.Option{responder.WithHostname(*host + ".local")}
	if *ipv6 {
		opts = append(opts, responder.WithIPv6())
	}
	r, err := responder.New(context.Background(), opts...)
	if err != nil {
		log.Fatalf("start responder: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close: %v", err)
		}
	}()

	r.OnProbe(func() { log.Printf("probe sent") })
	r.OnAnnounce(func() { log.Printf("announcement sent") })

	svc := &responder.Service{
		InstanceName: *name,
		ServiceType:  *svcType,
		Port:         uint16(*port), // #nosec G115 -- range checked above
		TXTRecords:   map[string]string{"txtvers": "1"},
	}
	if err := r.Register(svc); err != nil {
		log.Fatalf("register %q: %v", *name, err)
	}
	if svc.InstanceName != *name {
		log.Printf("renamed after conflict: %q -> %q", *name, svc.InstanceName)
	}
	log.Printf("registered %q (%s) on %s.local port %d; type \"rename <new name>\" or Ctrl-C to stop",
		svc.InstanceName, svc.ServiceType, *host, *port)

	// BCT II.16 (manual name change): the operator renames the service to a
	// name BCT chooses. Unregister sends the goodbye (PTR TTL=0) for the old
	// name, then Register probes and announces the new one (RFC 6763 App. C).
	lines := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()
	for {
		select {
		case <-sigCtx.Done():
			log.Printf("stopping: sending goodbye packets")
			return
		case line := <-lines:
			newName, ok := strings.CutPrefix(strings.TrimSpace(line), "rename ")
			if !ok || newName == "" {
				log.Printf("unknown command %q (want: rename <new name>)", line)
				continue
			}
			if err := r.Unregister(svc.InstanceName); err != nil {
				log.Printf("unregister %q: %v", svc.InstanceName, err)
			}
			svc = &responder.Service{InstanceName: newName, ServiceType: svc.ServiceType, Port: svc.Port, TXTRecords: svc.TXTRecords}
			if err := r.Register(svc); err != nil {
				log.Printf("register %q: %v", newName, err)
				continue
			}
			log.Printf("renamed by operator: now %q", svc.InstanceName)
		}
	}
}
