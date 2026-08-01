package querier

import (
	"context"
	goerrors "errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/security"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Querier provides high-level mDNS query functionality.
//
// Querier manages a UDP multicast socket and background receiver goroutine
// to handle mDNS queries per FR-005, FR-006.
//
// Example:
//
//	q, err := querier.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer q.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	response, err := q.Query(ctx, "printer.local", querier.RecordTypeA)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, record := range response.Records {
//	    if ip := record.AsA(); ip != nil {
//	        fmt.Printf("Found printer at %s\n", ip)
//	    }
//	}
//
// NOTE: Fields are ordered for memory alignment (fieldalignment optimization).
// Larger types (interfaces, slices, sync types) come first, then smaller types.
// This reduces struct size from 144 → 120 bytes (16.7% memory savings).
// Related fields are still documented together via comments.
type Querier struct {
	// transport is the IPv4 network transport (always present)
	// T031: Migrated from socket net.PacketConn to transport.Transport interface
	transport transport.Transport

	// transport6 is the IPv6 network transport (nil unless WithIPv6() is used)
	// RFC 6762 §11: IPv6 mDNS uses FF02::FB:5353
	transport6 transport.Transport

	// ctx is the lifecycle context for the Querier
	ctx context.Context

	// wg tracks background goroutines (receiver)
	// Placed early due to 16-byte alignment requirement of sync.WaitGroup
	wg sync.WaitGroup

	// explicitInterfaces is the user-provided explicit list of interfaces (if set)
	// Takes priority over interfaceFilter if non-nil
	explicitInterfaces []net.Interface

	// defaultTimeout is the default timeout for queries (default: 1 second per SC-002)
	defaultTimeout time.Duration

	// rateLimitCooldown is the duration to drop packets after threshold exceeded (default: 60s)
	// Per FR-028: Configurable via WithRateLimitCooldown()
	rateLimitCooldown time.Duration

	// cancel cancels the lifecycle context
	cancel context.CancelFunc

	// responseChan receives incoming mDNS responses from both IPv4 and IPv6 receiver goroutines
	responseChan chan []byte

	// interfaceFilter is a custom interface selection function (if set)
	// Used only if explicitInterfaces is nil
	interfaceFilter func(net.Interface) bool

	// rateLimiter is the rate limiter instance (created in New() if enabled)
	rateLimiter *security.RateLimiter

	// rateLimitThreshold is the max queries/second per source IP (default: 100)
	// Per FR-027: Configurable via WithRateLimitThreshold()
	rateLimitThreshold int

	// mu protects concurrent access to Query operations
	mu sync.Mutex

	// rateLimitEnabled indicates whether rate limiting is enabled (default: true)
	// Per FR-033: Configurable via WithRateLimit()
	rateLimitEnabled bool

	// ipv6Enabled indicates whether dual-stack IPv6 operation is requested
	// Set by WithIPv6(); actual IPv6 availability depends on transport6 != nil
	ipv6Enabled bool
}

// New creates a new Querier with optional configuration.
//
// New initializes the UDP multicast socket and starts a background receiver
// goroutine per FR-005, FR-006.
//
// FR-004: System MUST use mDNS port 5353 and multicast address 224.0.0.251
// FR-005: System MUST send queries to multicast group
// FR-006: System MUST receive responses with configurable timeout
//
// Parameters:
//   - opts: Optional functional options (e.g., WithTimeout)
//
// Returns:
//   - *Querier: Configured querier instance
//   - error: NetworkError if socket creation fails
//
// Example:
//
//	q, err := querier.New(querier.WithTimeout(2 * time.Second))
func New(opts ...Option) (*Querier, error) {
	// T032: Create UDP multicast transport (migrated from network.CreateSocket)
	tr, err := transport.NewUDPv4Transport()
	if err != nil {
		return nil, err // Already wrapped as NetworkError
	}

	// Create lifecycle context
	ctx, cancel := context.WithCancel(context.Background())

	// Create querier with defaults
	q := &Querier{
		transport:          tr,
		defaultTimeout:     1 * time.Second,        // SC-002: discover devices within 1 second
		responseChan:       make(chan []byte, 100), // Buffer for incoming responses
		ctx:                ctx,
		cancel:             cancel,
		rateLimitEnabled:   true,             // FR-033: Default enabled
		rateLimitThreshold: 100,              // FR-027: Default 100 qps
		rateLimitCooldown:  60 * time.Second, // FR-028: Default 60s
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(q); err != nil {
			cancel()       // Clean up context before returning error
			_ = tr.Close() // Ignore error, already returning primary error
			return nil, err
		}
	}

	// Initialize rate limiter if enabled (after options applied)
	if q.rateLimitEnabled {
		q.rateLimiter = security.NewRateLimiter(
			q.rateLimitThreshold,
			q.rateLimitCooldown,
			10000, // Max 10,000 source IPs tracked
		)

		// Start periodic cleanup goroutine (every 5 minutes per FR-031)
		q.wg.Add(1)
		go q.cleanupLoop()
	}

	// Start background IPv4 receiver goroutine per FR-006
	q.wg.Add(1)
	go q.runReceiveLoop(q.transport)

	// If IPv6 was requested, attempt to create the IPv6 transport.
	// Failure is non-fatal: the querier continues in IPv4-only mode.
	if q.ipv6Enabled {
		tr6, err6 := transport.NewUDPv6Transport()
		if err6 == nil {
			q.transport6 = tr6
			q.wg.Add(1)
			go q.runReceiveLoop(q.transport6)
		}
		// If err6 != nil, IPv6 is silently unavailable (no IPv6 interface).
	}

	return q, nil
}

// Query sends an mDNS query and returns all responses received within the timeout.
//
// Query validates inputs, builds the query message, sends it to the multicast group,
// and aggregates responses per FR-001 through FR-012.
//
// FR-001: System MUST construct valid mDNS query messages per RFC 6762
// FR-002: System MUST support querying for A, PTR, SRV, and TXT record types
// FR-003: System MUST validate queried names follow DNS naming rules
// FR-007: System MUST deduplicate identical responses from multiple responders
// FR-008: System MUST aggregate responses received within timeout window
// FR-009: System MUST parse mDNS response messages per RFC 6762 wire format
// FR-010: System MUST filter answer section records, ignoring authority/additional
// FR-011: System MUST validate response message format and discard malformed packets
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4
//
// Parameters:
//   - ctx: Context for timeout/cancellation (use context.WithTimeout for custom timeout)
//   - name: DNS name to query (e.g., "printer.local")
//   - recordType: Type of record to query (RecordTypeA, RecordTypePTR, etc.)
//
// Returns:
//   - *Response: Aggregated response with all discovered records
//   - error: ValidationError for invalid inputs, context.Canceled/context.DeadlineExceeded, or other errors
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
//	for _, record := range response.Records {
//	    fmt.Printf("Found: %s → %v\n", record.Name, record.Data)
//	}
func (q *Querier) Query(ctx context.Context, name string, recordType RecordType) (*Response, error) {
	// Protect concurrent query operations
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check context cancellation upfront
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Honor the configured default timeout when the caller's context carries no
	// deadline, so queries are always bounded. Without this, Query on a
	// deadline-less context (e.g. context.Background()) blocks forever in
	// collectResponses, and WithTimeout silently does nothing (issue #5).
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && q.defaultTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, q.defaultTimeout)
		defer cancel()
	}

	// FR-003: Validate name
	err := protocol.ValidateName(name)
	if err != nil {
		return nil, err // Already wrapped as ValidationError
	}

	// FR-002: Validate record type
	err = protocol.ValidateRecordType(uint16(recordType))
	if err != nil {
		return nil, err // Already wrapped as ValidationError
	}

	// FR-001: Build query message
	queryMsg, err := message.BuildQuery(name, uint16(recordType))
	if err != nil {
		return nil, err
	}

	// FR-005: Send query to the IPv4 mDNS multicast group (224.0.0.251:5353).
	err = q.transport.Send(ctx, queryMsg, protocol.MulticastGroupIPv4())
	if err != nil {
		return nil, err // Already wrapped as NetworkError
	}

	// RFC 6762 §11: Also send on IPv6 (FF02::FB:5353) when dual-stack is active.
	// Non-fatal: IPv4 query already sent; IPv6 is best-effort.
	if q.transport6 != nil {
		_ = q.transport6.Send(ctx, queryMsg, protocol.MulticastGroupIPv6())
	}

	// FR-008: Aggregate responses received within timeout window
	return q.collectResponses(ctx, name, recordType)
}

// DiscoverServices performs a full DNS-SD discovery for the given service type.
//
// This is a convenience method that chains multiple queries to return fully
// resolved service instances:
//  1. PTR query to find service instances (browse phase)
//  2. SRV query per instance for hostname and port
//  3. TXT query per instance for metadata
//  4. A query per hostname for IPv4 address
//
// The context deadline is split across all phases. For best results, use a
// timeout of at least 2-3 seconds to allow time for both browsing and resolution.
//
// For fine-grained control over individual queries, use [Querier.Query] directly.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	defer cancel()
//
//	services, err := q.DiscoverServices(ctx, "_http._tcp.local")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, svc := range services {
//	    fmt.Printf("%s at %s:%d (%s)\n",
//	        svc.InstanceName, svc.AddrIPv4, svc.Port, svc.TXT["path"])
//	}
func (q *Querier) DiscoverServices(ctx context.Context, serviceType string) ([]ServiceInstance, error) {
	// Phase 1: Browse for instances via PTR query.
	// Allocate ~40% of the remaining time for browsing, rest for resolving details.
	browseTimeout := 1 * time.Second // default if no deadline
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		browseTimeout = remaining * 2 / 5
		if browseTimeout < 200*time.Millisecond {
			browseTimeout = 200 * time.Millisecond
		}
	}

	browseCtx, browseCancel := context.WithTimeout(ctx, browseTimeout)
	ptrResp, err := q.Query(browseCtx, serviceType, RecordTypePTR)
	browseCancel()
	if err != nil {
		return nil, fmt.Errorf("browse %s: %w", serviceType, err)
	}

	// Phase 2: Resolve each discovered instance.
	var services []ServiceInstance
	resolveTimeout := 500 * time.Millisecond

	for _, record := range ptrResp.Records {
		target := record.AsPTR()
		if target == "" {
			continue
		}

		svc := ServiceInstance{ServiceType: serviceType}

		// Extract instance name: "My Printer._http._tcp.local" → "My Printer"
		if strings.HasSuffix(target, "."+serviceType) {
			svc.InstanceName = strings.TrimSuffix(target, "."+serviceType)
		} else {
			svc.InstanceName = target
		}

		// RFC 6763 §12: prefer SRV/TXT/A bundled in the browse response's
		// additional section; fall back to explicit queries only for what is
		// missing (issue #4 — saves up to 3 round-trips per instance).
		if rr := findInAdditionals(ptrResp.Additionals, target, RecordTypeSRV); rr != nil {
			if srv := rr.AsSRV(); srv != nil {
				svc.Hostname = srv.Target
				svc.Port = srv.Port
			}
		}
		if rr := findInAdditionals(ptrResp.Additionals, target, RecordTypeTXT); rr != nil {
			if txt := rr.AsTXT(); txt != nil {
				svc.TXT = ParseTXT(txt)
			}
		}
		if svc.Hostname != "" {
			if rr := findInAdditionals(ptrResp.Additionals, svc.Hostname, RecordTypeA); rr != nil {
				if ip := rr.AsA(); ip != nil {
					svc.AddrIPv4 = ip
				}
			}
		}

		// Fallback: SRV query for hostname + port if not bundled as an additional.
		if svc.Hostname == "" {
			srvCtx, srvCancel := context.WithTimeout(ctx, resolveTimeout)
			srvResp, srvErr := q.Query(srvCtx, target, RecordTypeSRV)
			srvCancel()
			if srvErr == nil {
				for _, r := range srvResp.Records {
					if srv := r.AsSRV(); srv != nil {
						svc.Hostname = srv.Target
						svc.Port = srv.Port
						break
					}
				}
			}
		}

		// Fallback: TXT query for metadata if not bundled as an additional.
		if svc.TXT == nil {
			txtCtx, txtCancel := context.WithTimeout(ctx, resolveTimeout)
			txtResp, txtErr := q.Query(txtCtx, target, RecordTypeTXT)
			txtCancel()
			if txtErr == nil {
				for _, r := range txtResp.Records {
					if txt := r.AsTXT(); txt != nil {
						svc.TXT = ParseTXT(txt)
						break
					}
				}
			}
		}

		// Fallback: A query for IPv4 if we have a hostname but no address yet.
		if svc.Hostname != "" && svc.AddrIPv4 == nil {
			aCtx, aCancel := context.WithTimeout(ctx, resolveTimeout)
			aResp, aErr := q.Query(aCtx, svc.Hostname, RecordTypeA)
			aCancel()
			if aErr == nil {
				for _, r := range aResp.Records {
					if ip := r.AsA(); ip != nil {
						svc.AddrIPv4 = ip
						break
					}
				}
			}
		}

		// AAAA query for IPv6 address when dual-stack is active.
		if q.transport6 != nil && svc.Hostname != "" && svc.AddrIPv6 == nil {
			aaaaCtx, aaaaCancel := context.WithTimeout(ctx, resolveTimeout)
			aaaaResp, aaaaErr := q.Query(aaaaCtx, svc.Hostname, RecordTypeAAAA)
			aaaaCancel()
			if aaaaErr == nil {
				for _, r := range aaaaResp.Records {
					if ip := r.AsAAAA(); ip != nil {
						svc.AddrIPv6 = ip
						break
					}
				}
			}
		}

		services = append(services, svc)
	}

	return services, nil
}

// toRecordData normalizes parsed RDATA into the querier's public types.
// message.ParseRDATA returns the internal message.SRVData for SRV records;
// convert it to the public SRVData so ResourceRecord.AsSRV() works (the named
// types are distinct, so without this the type assertion in AsSRV always fails
// and SRV resolution silently yields nil). A/PTR/TXT already use shared types.
func toRecordData(data interface{}) interface{} {
	if srv, ok := data.(message.SRVData); ok {
		return SRVData{
			Priority: srv.Priority,
			Weight:   srv.Weight,
			Port:     srv.Port,
			Target:   srv.Target,
		}
	}
	return data
}

// findInAdditionals returns the first additional-section record matching the
// given name and type, or nil. Used to resolve a service instance from records
// bundled in a browse response (RFC 6763 §12) before falling back to queries.
func findInAdditionals(additionals []ResourceRecord, name string, t RecordType) *ResourceRecord {
	for i := range additionals {
		if additionals[i].Type == t && additionals[i].Name == name {
			return &additionals[i]
		}
	}
	return nil
}

// collectResponses aggregates mDNS responses within the timeout window.
//
// FR-007: Deduplicate identical responses
// FR-008: Aggregate responses received within timeout window
// FR-009: Parse mDNS response messages
// FR-010: Filter answer section records
// FR-011: Validate and discard malformed packets
// FR-016: Continue collecting after discarding malformed packets
func (q *Querier) collectResponses(ctx context.Context, _ string, queryType RecordType) (*Response, error) {
	response := &Response{
		Records: make([]ResourceRecord, 0),
	}

	// Deduplication map per FR-007
	seen := make(map[string]bool)

	// Collect responses until timeout or cancellation
	for {
		select {
		case <-ctx.Done():
			// Timeout is NOT an error per FR-008 - return what we collected
			return response, nil

		case responseMsg := <-q.responseChan:
			// FR-009: Parse response message
			parsedMsg, err := message.ParseMessage(responseMsg)
			if err != nil {
				// FR-011, FR-016: Log and continue on malformed packets
				// In M1, we silently continue (production might log)
				continue
			}

			// FR-021, FR-022: Validate response flags
			err = protocol.ValidateResponse(parsedMsg.Header.Flags)
			if err != nil {
				// Invalid response (QR=0 or RCODE≠0) - discard per FR-011
				continue
			}

			// FR-010: Process only Answer section (ignore Authority, Additional)
			for _, answer := range parsedMsg.Answers {
				// Filter by query type (optional - could also return all types)
				if RecordType(answer.TYPE) != queryType {
					// Skip records of different type
					// (Production might include related records)
					continue
				}

				// Parse type-specific RDATA against the full message so compressed
				// PTR/SRV target names (used by Avahi/Bonjour) resolve.
				data, err := message.ParseRDATAInMessage(answer.TYPE, responseMsg, answer.RDATAOffset, int(answer.RDLENGTH))
				if err != nil {
					// Malformed RDATA - skip this record per FR-011
					continue
				}

				// FR-007: Deduplicate identical records
				// Key: name + type + data representation
				dedupeKey := fmt.Sprintf("%s|%d|%v", answer.NAME, answer.TYPE, data)
				if seen[dedupeKey] {
					continue // Duplicate - skip
				}
				seen[dedupeKey] = true

				// Convert to public ResourceRecord
				record := ResourceRecord{
					Name:  answer.NAME,
					Type:  RecordType(answer.TYPE),
					Class: answer.CLASS,
					TTL:   answer.TTL,
					Data:  toRecordData(data),
				}

				response.Records = append(response.Records, record)
			}

			// Retain Additional-section records (RFC 6763 §12). DNS-SD responders
			// bundle SRV/TXT/A here so one PTR query resolves a whole instance;
			// DiscoverServices consumes them to skip follow-up queries (issue #4).
			// Parsing against the full message resolves compressed SRV/PTR target
			// names, so bundled additionals from Avahi/Bonjour resolve too.
			for _, add := range parsedMsg.Additionals {
				data, err := message.ParseRDATAInMessage(add.TYPE, responseMsg, add.RDATAOffset, int(add.RDLENGTH))
				if err != nil {
					continue
				}
				dedupeKey := fmt.Sprintf("add|%s|%d|%v", add.NAME, add.TYPE, data)
				if seen[dedupeKey] {
					continue
				}
				seen[dedupeKey] = true

				response.Additionals = append(response.Additionals, ResourceRecord{
					Name:  add.NAME,
					Type:  RecordType(add.TYPE),
					Class: add.CLASS,
					TTL:   add.TTL,
					Data:  toRecordData(data),
				})
			}
		}
	}
}

// runReceiveLoop is the shared receive goroutine for both IPv4 and IPv6 transports.
//
// The transport type itself determines which source-IP validation rules apply:
//   - *transport.UDPv6Transport → accept fe80::/10 IPv6 link-local addresses
//   - anything else (IPv4)      → accept 169.254.x.x and RFC-1918 private ranges
//
// FR-006: System MUST receive responses with configurable timeout
// FR-017: System MUST close socket after query completion
//
// nolint:gocyclo // Complexity justified: network packet handling with rate limiting, context management, source IP validation, and error recovery
func (q *Querier) runReceiveLoop(tr transport.Transport) {
	defer q.wg.Done()

	_, isIPv6Transport := tr.(*transport.UDPv6Transport)

	for {
		select {
		case <-q.ctx.Done():
			return

		default:
			ctx, cancel := context.WithTimeout(q.ctx, 100*time.Millisecond)
			responseMsg, srcAddr, _, err := tr.Receive(ctx)
			cancel()

			if err != nil {
				var netErr *errors.NetworkError
				if goerrors.As(err, &netErr) {
					continue // timeout is expected
				}
				continue
			}

			// T077: Packet size validation per RFC 6762 §17 (FR-034)
			const maxMDNSPacketSize = 9000
			if len(responseMsg) > maxMDNSPacketSize {
				continue
			}

			// Extract source IP for validation and rate limiting
			var srcIP net.IP
			var srcIPStr string
			if udpAddr, ok := srcAddr.(*net.UDPAddr); ok {
				srcIP = udpAddr.IP
				srcIPStr = udpAddr.IP.String()
			}

			// RFC 6762 §2: mDNS is link-local scope — validate source IP accordingly.
			if srcIP != nil {
				if !isAcceptableSourceIP(srcIP, isIPv6Transport) {
					continue
				}
			}

			// FR-029: Rate limiting — drop packets from flooding sources.
			if q.rateLimitEnabled && q.rateLimiter != nil && srcIPStr != "" {
				if !q.rateLimiter.Allow(srcIPStr) {
					continue
				}
			}

			select {
			case q.responseChan <- responseMsg:
			default:
				// Channel full — drop packet.
			}
		}
	}
}

// isAcceptableSourceIP returns true when the source IP is within the link-local
// scope expected for mDNS traffic per RFC 6762 §2.
//
// For IPv4: accept 169.254.0.0/16 (link-local) and RFC-1918 private ranges.
// For IPv6: accept fe80::/10 (link-local); reject all other addresses.
func isAcceptableSourceIP(ip net.IP, isIPv6Transport bool) bool {
	if isIPv6Transport {
		// fe80::/10 is the only link-local scope for IPv6 per RFC 4291 §2.5.6.
		if len(ip) == 16 && ip[0] == 0xfe && ip[1]&0xc0 == 0x80 {
			return true
		}
		// IPv4-mapped IPv6 addresses — apply IPv4 rules.
		if ip4 := ip.To4(); ip4 != nil {
			return isAcceptableSourceIP(ip4, false)
		}
		return false
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	// 169.254.0.0/16 — RFC 3927 IPv4 link-local
	if ip4[0] == 169 && ip4[1] == 254 {
		return true
	}
	return security.IsPrivate(ip)
}

// cleanupLoop periodically cleans up stale rate limiter entries.
// Per FR-031: Cleanup runs every 5 minutes to prevent memory growth.
func (q *Querier) cleanupLoop() {
	defer q.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			// Querier closed - exit loop
			return

		case <-ticker.C:
			// Periodic cleanup
			if q.rateLimiter != nil {
				q.rateLimiter.Cleanup()
			}
		}
	}
}

// Close gracefully shuts down the Querier and releases resources.
//
// Close cancels the lifecycle context, waits for background goroutines to exit,
// and closes the UDP socket per FR-017, FR-018.
//
// FR-017: System MUST close socket after query completion
// FR-018: System MUST support graceful shutdown via context cancellation
//
// Example:
//
//	q, err := querier.New()
//	if err != nil {
//	    return err
//	}
//	defer q.Close() // Always close to release resources
func (q *Querier) Close() error {
	// Cancel lifecycle context (stops receiver goroutine)
	q.cancel()

	// Wait for receiver goroutine to exit
	q.wg.Wait()

	// Close IPv4 transport per FR-017
	// T035: Migrated from network.CloseSocket to transport.Close()
	// FR-004 FIX: Now properly propagates errors (CloseSocket was swallowing them)
	err := q.transport.Close()

	// Close IPv6 transport if active; preserve the first error.
	if q.transport6 != nil {
		if err6 := q.transport6.Close(); err6 != nil && err == nil {
			err = err6
		}
	}

	// Only close the response channel when all transports closed successfully.
	// If a transport returned an error the caller may retry Close(); closing
	// the channel on a partial-failure would panic on the retry.
	if err != nil {
		return err
	}

	close(q.responseChan)
	return nil
}
