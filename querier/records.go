// Package querier provides a high-level API for querying mDNS (.local) services.
package querier

import (
	"net"
	"strings"

	"github.com/joshuafuller/beacon/internal/protocol"
)

// RecordType represents a DNS record type for querying.
//
// Supported types in M1 (Basic mDNS Querier) per FR-002:
//   - RecordTypeA: IPv4 address records
//   - RecordTypePTR: Pointer records (service discovery)
//   - RecordTypeSRV: Service records (hostname and port)
//   - RecordTypeTXT: Text records (service metadata)
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

// Response represents the aggregated results from an mDNS query.
//
// Response contains all records received within the timeout window per FR-008.
// Empty Records slice indicates timeout with no responses (not an error).
type Response struct {
	// Records contains all discovered resource records.
	//
	// Records may include:
	//   - Answer records (direct answers to the query)
	//   - Additional records (supplementary information, e.g., A records for SRV targets)
	//
	// Per FR-010, Authority records are ignored in M1.
	Records []ResourceRecord

	// Additionals contains records from the response's Additional section.
	//
	// DNS-SD responders bundle SRV/TXT/A records alongside a PTR answer to let a
	// single query resolve a service instance without follow-up round-trips
	// (RFC 6763 §12). These are kept separate from Records (which holds only
	// answer-section records of the queried type) so existing callers are
	// unaffected; DiscoverServices consumes them to avoid extra queries.
	Additionals []ResourceRecord
}

// ResourceRecord represents a single DNS resource record from an mDNS response.
//
// ResourceRecord provides access to both raw DNS fields and type-specific
// parsed data through helper methods (AsA, AsPTR, AsSRV, AsTXT).
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

// ParseTXT parses TXT record strings into key-value pairs per RFC 6763 §6.
//
// TXT records contain "key=value" pairs. Keys without "=" are treated as
// boolean flags with empty string values. Empty strings are skipped.
//
// Example:
//
//	txt := record.AsTXT() // ["version=1.0", "path=/api", "debug"]
//	kv := querier.ParseTXT(txt)
//	// kv = {"version": "1.0", "path": "/api", "debug": ""}
func ParseTXT(txt []string) map[string]string {
	result := make(map[string]string, len(txt))
	for _, entry := range txt {
		if entry == "" {
			continue
		}
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			result[entry[:idx]] = entry[idx+1:]
		} else {
			// Boolean flag (key with no value) per RFC 6763 §6.4
			result[entry] = ""
		}
	}
	return result
}

// ServiceInstance represents a fully resolved mDNS service discovered via DNS-SD.
//
// This is returned by [Querier.DiscoverServices] after performing the full
// PTR → SRV → TXT → A query sequence.
type ServiceInstance struct {
	// InstanceName is the human-readable service name (e.g., "My Printer").
	InstanceName string

	// ServiceType is the DNS-SD service type (e.g., "_http._tcp.local").
	ServiceType string

	// Hostname is the target host from the SRV record (e.g., "printer.local").
	Hostname string

	// Port is the service port from the SRV record.
	Port uint16

	// AddrIPv4 is the IPv4 address from the A record, or nil if unresolved.
	AddrIPv4 net.IP

	// TXT contains parsed key-value metadata from the TXT record.
	TXT map[string]string
}
