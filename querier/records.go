// Package querier provides a high-level API for querying mDNS (.local) services.
package querier

import (
	"net"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// RecordType represents a DNS record type for querying per RFC 1035.
//
// RFC 1035 §3.2.2: TYPE Values
// RFC 6762 §5: mDNS Query Types
//
// RecordType specifies which kind of resource records to query from the network.
// Each type serves a specific purpose in DNS-SD service discovery:
//
//  - A records: Resolve hostnames to IPv4 addresses
//  - PTR records: Enumerate service instances of a given type
//  - SRV records: Get service location (hostname and port)
//  - TXT records: Retrieve service metadata (key=value pairs)
//
// Supported types in M1 (Basic mDNS Querier) per FR-002:
//   - RecordTypeA: IPv4 address records (type 1)
//   - RecordTypePTR: Pointer records (type 12) for service discovery
//   - RecordTypeSRV: Service records (type 33) for hostname and port
//   - RecordTypeTXT: Text records (type 16) for service metadata
//
// Functional Requirements:
//   - FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
//
// Example:
//
//	// Query for IPv4 address
//	response, _ := q.Query(ctx, "printer.local", querier.RecordTypeA)
//
//	// Discover HTTP services
//	response, _ = q.Query(ctx, "_http._tcp.local", querier.RecordTypePTR)
type RecordType uint16

const (
	// RecordTypeA queries for IPv4 address records (type 1).
	//
	// Example: Query("printer.local", RecordTypeA) → 192.168.1.100
	RecordTypeA RecordType = RecordType(protocol.RecordTypeA)

	// RecordTypePTR queries for pointer records (type 12).
	//
	// Used for service discovery.
	// Example: Query("_http._tcp.local", RecordTypePTR) → "webserver._http._tcp.local"
	RecordTypePTR RecordType = RecordType(protocol.RecordTypePTR)

	// RecordTypeTXT queries for text records (type 16).
	//
	// Used for service metadata (key=value pairs).
	// Example: Query("webserver._http._tcp.local", RecordTypeTXT) → ["version=1.0", "path=/"]
	RecordTypeTXT RecordType = RecordType(protocol.RecordTypeTXT)

	// RecordTypeSRV queries for service records (type 33).
	//
	// Used to get service hostname and port.
	// Example: Query("webserver._http._tcp.local", RecordTypeSRV) → {Priority:0, Weight:0, Port:8080, Target:"server.local"}
	RecordTypeSRV RecordType = RecordType(protocol.RecordTypeSRV)
)

// String returns a human-readable name for the record type.
func (r RecordType) String() string {
	return protocol.RecordType(r).String()
}

// Response represents the aggregated results from an mDNS query per RFC 6762 §6.
//
// RFC 6762 §6: Responding
// RFC 6762 §7: Traffic Reduction (response aggregation)
//
// Response contains all unique resource records received within the timeout window.
// Multiple responders may send identical records; these are deduplicated per FR-007.
//
// An empty Records slice indicates no devices responded within the timeout. This is
// NOT an error condition - it simply means no services/devices were discovered.
//
// Functional Requirements:
//   - FR-007: System MUST deduplicate identical responses from multiple responders
//   - FR-008: System MUST aggregate responses received within timeout window
//   - FR-010: System MUST filter answer section records, ignoring authority/additional
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	response, err := q.Query(ctx, "printer.local", querier.RecordTypeA)
//	if err != nil {
//	    return err
//	}
//
//	if len(response.Records) == 0 {
//	    fmt.Println("No devices found (timeout - not an error)")
//	} else {
//	    for _, record := range response.Records {
//	        if ip := record.AsA(); ip != nil {
//	            fmt.Printf("Found device at %s\n", ip)
//	        }
//	    }
//	}
type Response struct {
	// Records contains all discovered resource records.
	//
	// Records may include:
	//   - Answer records (direct answers to the query)
	//   - Additional records (supplementary information, e.g., A records for SRV targets)
	//
	// Per FR-010, Authority records are ignored in M1.
	Records []ResourceRecord
}

// ResourceRecord represents a single DNS resource record from an mDNS response.
//
// RFC 1035 §3.2.1: Resource Record Format
// RFC 6762 §5: mDNS Resource Record Format
//
// ResourceRecord provides access to both raw DNS fields (Name, Type, Class, TTL)
// and type-specific parsed data through helper methods (AsA, AsPTR, AsSRV, AsTXT).
//
// The Data field contains parsed, type-specific information:
//   - A record: net.IP (IPv4 address)
//   - PTR record: string (target domain name)
//   - SRV record: SRVData struct (priority, weight, port, target)
//   - TXT record: []string (text strings)
//
// Wire format (RFC 1035 §3.2.1):
//
//	0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     NAME                      |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     TYPE                      |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     CLASS                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      TTL                      |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                   RDLENGTH                    |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                     RDATA                     |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//
// Functional Requirements:
//   - FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
//   - FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
//
// Example:
//
//	for _, record := range response.Records {
//	    switch record.Type {
//	    case querier.RecordTypeA:
//	        if ip := record.AsA(); ip != nil {
//	            fmt.Printf("IPv4: %s → %s\n", record.Name, ip)
//	        }
//	    case querier.RecordTypePTR:
//	        if target := record.AsPTR(); target != "" {
//	            fmt.Printf("Service: %s → %s\n", record.Name, target)
//	        }
//	    }
//	}
type ResourceRecord struct {
	// Data contains the type-specific parsed data:
	//   - A record: net.IP (IPv4 address)
	//   - PTR record: string (target domain name)
	//   - SRV record: SRVData struct
	//   - TXT record: []string (text strings)
	//
	// Use AsA(), AsPTR(), AsSRV(), or AsTXT() for type-safe access.
	Data interface{}

	// Name is the domain name for this record (e.g., "printer.local").
	Name string

	// TTL is the time-to-live in seconds.
	//
	// Per RFC 6762, TTL=0 may indicate cache flush.
	TTL uint32

	// Type is the DNS record type (A, PTR, SRV, TXT).
	Type RecordType

	// Class is the DNS class (typically IN=1 for Internet).
	Class uint16
}

// SRVData represents parsed SRV record data per RFC 2782.
//
// SRV records specify the location (hostname and port) of a service.
type SRVData struct {
	// Target is the domain name of the host providing the service.
	// Target may require additional A/AAAA query to resolve to IP.
	Target string

	// Priority is the priority of this target host.
	// Lower values indicate higher priority.
	Priority uint16

	// Weight is used for load balancing among targets with same priority.
	// Higher values indicate higher weight.
	Weight uint16

	// Port is the TCP or UDP port where the service is available.
	Port uint16
}

// AsA returns the IPv4 address for an A record, or nil if not an A record.
//
// Example:
//
//	for _, record := range response.Records {
//	    if ip := record.AsA(); ip != nil {
//	        fmt.Printf("Found IP: %s\n", ip)
//	    }
//	}
func (r *ResourceRecord) AsA() net.IP {
	if r.Type != RecordTypeA {
		return nil
	}

	ip, ok := r.Data.(net.IP)
	if !ok {
		return nil
	}

	return ip
}

// AsPTR returns the target name for a PTR record, or empty string if not a PTR record.
//
// Example:
//
//	for _, record := range response.Records {
//	    if target := record.AsPTR(); target != "" {
//	        fmt.Printf("Found service: %s\n", target)
//	    }
//	}
func (r *ResourceRecord) AsPTR() string {
	if r.Type != RecordTypePTR {
		return ""
	}

	target, ok := r.Data.(string)
	if !ok {
		return ""
	}

	return target
}

// AsSRV returns the SRV data for an SRV record, or nil if not an SRV record.
//
// Example:
//
//	for _, record := range response.Records {
//	    if srv := record.AsSRV(); srv != nil {
//	        fmt.Printf("Service at %s:%d\n", srv.Target, srv.Port)
//	    }
//	}
func (r *ResourceRecord) AsSRV() *SRVData {
	if r.Type != RecordTypeSRV {
		return nil
	}

	srv, ok := r.Data.(SRVData)
	if !ok {
		return nil
	}

	return &srv
}

// AsTXT returns the text strings for a TXT record, or nil if not a TXT record.
//
// Example:
//
//	for _, record := range response.Records {
//	    if txt := record.AsTXT(); txt != nil {
//	        for _, kv := range txt {
//	            fmt.Printf("Metadata: %s\n", kv)
//	        }
//	    }
//	}
func (r *ResourceRecord) AsTXT() []string {
	if r.Type != RecordTypeTXT {
		return nil
	}

	txt, ok := r.Data.([]string)
	if !ok {
		return nil
	}

	return txt
}
