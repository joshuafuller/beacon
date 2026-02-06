package querier_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joshuafuller/beacon/querier"
)

// ExampleQuerier_Query demonstrates basic mDNS query for PTR records.
// This is the most common use case for service discovery.
func ExampleQuerier_Query() {
	// Create querier
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	// Query for HTTP services on local network
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
	if err != nil {
		log.Fatal(err)
	}

	// Print discovered services
	for _, record := range response.Records {
		if target := record.AsPTR(); target != "" {
			fmt.Printf("Found service: %s\n", target)
		}
	}

	// Output examples:
	// Found service: My Web Server._http._tcp.local
	// Found service: API Server._http._tcp.local
}

// ExampleQuerier_Query_aRecord demonstrates querying for IPv4 addresses.
// This resolves a hostname to its IP address.
func ExampleQuerier_Query_aRecord() {
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	// Query for A record (IPv4 address)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := q.Query(ctx, "printer.local", querier.RecordTypeA)
	if err != nil {
		log.Fatal(err)
	}

	// Print IP addresses
	for _, record := range response.Records {
		if ip := record.AsA(); ip != nil {
			fmt.Printf("Printer IP: %s\n", ip)
		}
	}

	// Output example:
	// Printer IP: 192.168.1.100
}

// ExampleQuerier_Query_srvRecord demonstrates querying for SRV records.
// SRV records provide hostname and port for a service instance.
func ExampleQuerier_Query_srvRecord() {
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	// Query for SRV record
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := q.Query(ctx, "webserver._http._tcp.local", querier.RecordTypeSRV)
	if err != nil {
		log.Fatal(err)
	}

	// Print service details
	for _, record := range response.Records {
		if srv := record.AsSRV(); srv != nil {
			fmt.Printf("Host: %s, Port: %d\n", srv.Target, srv.Port)
		}
	}

	// Output example:
	// Host: server.local, Port: 8080
}

// ExampleQuerier_Query_txtRecord demonstrates querying for TXT records.
// TXT records contain service metadata as key-value pairs.
func ExampleQuerier_Query_txtRecord() {
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	// Query for TXT record
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := q.Query(ctx, "api._http._tcp.local", querier.RecordTypeTXT)
	if err != nil {
		log.Fatal(err)
	}

	// Print service metadata
	for _, record := range response.Records {
		if attrs := record.AsTXT(); attrs != nil {
			for _, attr := range attrs {
				fmt.Printf("Attribute: %s\n", attr)
			}
		}
	}

	// Output examples:
	// Attribute: version=1.0
	// Attribute: status=ready
}

// ExampleNew demonstrates creating a querier with default settings.
func ExampleNew() {
	// Create querier with default configuration
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	fmt.Println("Querier created successfully")
	// Output:
	// Querier created successfully
}

// ExampleNew_withTimeout demonstrates creating a querier with custom timeout.
func ExampleNew_withTimeout() {
	// Create querier with 5-second default timeout
	q, err := querier.New(querier.WithTimeout(5 * time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	fmt.Println("Querier created with custom timeout")
	// Output:
	// Querier created with custom timeout
}

// Example demonstrates complete service discovery workflow.
// This shows discovering services, resolving details, and connecting.
func Example() {
	// Step 1: Create querier
	q, err := querier.New()
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Step 2: Discover services (PTR query)
	ptrResp, err := q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
	if err != nil {
		log.Fatal(err)
	}

	if len(ptrResp.Records) == 0 {
		fmt.Println("No services found")
		return
	}

	// Step 3: Get first service name
	serviceName := ptrResp.Records[0].AsPTR()

	// Step 4: Query SRV to get hostname and port
	srvResp, err := q.Query(ctx, serviceName, querier.RecordTypeSRV)
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range srvResp.Records {
		if srv := record.AsSRV(); srv != nil {
			// Step 5: Query A record to get IP
			aResp, err := q.Query(ctx, srv.Target, querier.RecordTypeA)
			if err != nil {
				continue
			}

			for _, aRecord := range aResp.Records {
				if ip := aRecord.AsA(); ip != nil {
					fmt.Printf("Connect to: http://%s:%d\n", ip, srv.Port)
					return
				}
			}
		}
	}

	// Output example:
	// Connect to: http://192.168.1.100:8080
}
