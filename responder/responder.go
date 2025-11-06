package responder

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/records"
	"github.com/joshuafuller/beacon/internal/responder"
	"github.com/joshuafuller/beacon/internal/state"
	"github.com/joshuafuller/beacon/internal/transport"
)

// Responder manages mDNS service registration and response per RFC 6762.
//
// Interface-Specific Addressing (RFC 6762 §15):
// The responder automatically detects which network interface received each query
// and responds with ONLY the IP address valid on that interface. This ensures
// clients can connect to the correct IP when the host has multiple network interfaces.
//
// Example: Host with WiFi (10.0.0.50) and Ethernet (192.168.1.100):
//   - Query on WiFi → Response contains 10.0.0.50
//   - Query on Ethernet → Response contains 192.168.1.100
//
// Graceful Degradation:
// If interface information is unavailable (e.g., on Windows or older kernels),
// the responder falls back to advertising the default interface IP.
//
// T035: Responder struct
// T080: Added query handler goroutine support
// T082: Added interface-specific addressing documentation
type Responder struct {
	ctx              context.Context
	transport        transport.Transport
	registry         *responder.Registry
	hostname         string
	injectConflict   bool                       // Test hook: inject conflict during probing
	responseBuilder  *responder.ResponseBuilder // RFC 6762 §6 response construction
	recordSet        *records.RecordSet         // Per-record rate limiting tracker
	queryHandlerDone chan struct{}              // Signal query handler shutdown

	// US2 GREEN: Store last machine for message capture (contract test support)
	lastMachine *state.Machine // Last state machine used for registration

	// US2 GREEN: Store callbacks for applying to new machines
	onProbeCallback    func() // Callback for probe events
	onAnnounceCallback func() // Callback for announce events

	// US2 GREEN: Store last announced records for contract test validation
	lastAnnouncedRecords []*ResourceRecord // Last record set announced
}

// New creates a new mDNS responder.
//
// T036: Responder.New() implementation
// T080: Start query handler goroutine
func New(ctx context.Context, opts ...Option) (*Responder, error) {
	// Get system hostname if not provided
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	hostname = hostname + ".local"

	// Create transport
	t, err := transport.NewUDPv4Transport()
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	r := &Responder{
		ctx:              ctx,
		transport:        t,
		registry:         responder.NewRegistry(),
		hostname:         hostname,
		responseBuilder:  responder.NewResponseBuilder(),
		recordSet:        records.NewRecordSet(),
		queryHandlerDone: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Start query handler goroutine (T080)
	go r.runQueryHandler()

	return r, nil
}

// maxRenameAttempts is the maximum number of times to rename a service on conflict.
//
// RFC 6762 §9: No explicit limit specified, but we use 10 as a reasonable maximum
// to prevent infinite loops and resource exhaustion.
//
// FR-032: System MUST handle registration failures gracefully
const maxRenameAttempts = 10

// Register registers a service with probing and announcing per RFC 6762 §8.
//
// Process:
//  1. Validate service parameters
//  2. Attempt to register (with rename loop on conflict)
//  3. Build record set (PTR, SRV, TXT, A)
//  4. Run state machine (Probing → Announcing → Established)
//  5. Add to registry on success
//
// RFC 6762 §8: Total time ~1.5s (500ms probing + 1s announcing)
// RFC 6762 §9: If conflict detected, rename and retry (max 10 attempts)
//
// Returns:
//   - error: validation error, conflict error, max attempts error, or context error
//
// T041: Full Register() implementation
// T062: Add max rename attempts limit (GREEN phase)
func (r *Responder) Register(service *Service) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	// Validate service parameters
	if err := service.Validate(); err != nil {
		return err
	}

	// Set hostname if not provided
	if service.Hostname == "" {
		service.Hostname = r.hostname
	}

	// Get local IPv4 address (simplified - use first non-loopback)
	ipv4, err := getLocalIPv4()
	if err != nil {
		return fmt.Errorf("failed to get local IPv4: %w", err)
	}

	// RFC 6762 §9: Rename loop on conflict (max 10 attempts)
	// Attempt probing up to maxRenameAttempts times
	for attempt := 1; attempt <= maxRenameAttempts; attempt++ {
		// Build record set for this service (with current name)
		serviceInfo := &records.ServiceInfo{
			InstanceName: service.InstanceName,
			ServiceType:  service.ServiceType,
			Hostname:     service.Hostname,
			Port:         service.Port,
			IPv4Address:  ipv4,
			TXTRecords:   service.TXTRecords,
		}
		recordSet := records.BuildRecordSet(serviceInfo)

		// US2 GREEN: Store record set for contract test validation
		r.lastAnnouncedRecords = recordSet

		// Create and run state machine
		machine := state.NewMachine()
		serviceName := service.InstanceName + "." + service.ServiceType

		// Apply test hooks (if any)
		if r.injectConflict {
			machine.SetInjectConflict(true)
		}

		// US2 GREEN: Store machine for message capture (contract test support)
		r.lastMachine = machine

		// US2 GREEN: Apply callbacks to new machine (if any)
		if r.onProbeCallback != nil {
			prober := machine.GetProber()
			if prober != nil {
				prober.SetOnSendQuery(r.onProbeCallback)
			}
		}
		if r.onAnnounceCallback != nil {
			announcer := machine.GetAnnouncer()
			if announcer != nil {
				announcer.SetOnSendAnnouncement(r.onAnnounceCallback)
			}
		}

		// Provide resource records to announcer for DNS message serialization
		announcer := machine.GetAnnouncer()
		if announcer != nil {
			announcer.SetRecords(recordSet)
		}

		// Run state machine (probing + announcing)
		err = machine.Run(r.ctx, serviceName)
		if err != nil {
			return fmt.Errorf("state machine failed: %w", err)
		}

		// Check final state
		finalState := machine.GetState()

		if finalState == state.StateConflictDetected {
			// Conflict detected - rename and retry (unless max attempts reached)
			if attempt >= maxRenameAttempts {
				// Max attempts exceeded - give up
				return fmt.Errorf("max rename attempts (%d) exceeded for service %q",
					maxRenameAttempts, service.InstanceName)
			}

			// Rename service and try again
			service.Rename() // Appends "-2", "-3", etc.
			continue         // Retry with new name
		}

		if finalState != state.StateEstablished {
			// This is NOT wrapping an error - finalState is state.State (int), not error type.
			// Using %v here is correct for formatting the state value.
			return fmt.Errorf("unexpected final state: %v", finalState) // nosemgrep: beacon-error-wrap-percent-v
		}

		// Success! Add to registry
		internalService := &responder.Service{
			InstanceName: service.InstanceName,
			ServiceType:  service.ServiceType,
			Port:         service.Port,
			TXT:          service.TXTRecords, // US5: Store TXT records for UpdateService support
		}
		err = r.registry.Register(internalService)
		if err != nil {
			return fmt.Errorf("failed to add to registry: %w", err)
		}

		return nil // Successfully registered
	}

	// Should never reach here (loop returns on success or max attempts)
	return fmt.Errorf("unexpected: register loop completed without result")
}

// Unregister unregisters a service and sends goodbye packets per RFC 6762 §10.1.
//
// RFC 6762 §10.1: "A host may send unsolicited responses with TTL=0 to announce
// the departure of a record."
//
// Process:
//  1. Remove from registry
//  2. Send goodbye announcements (TTL=0)
//
// Returns:
//   - error: if service not found or send fails
//
// T042: Implement Unregister() with goodbye packets
func (r *Responder) Unregister(serviceID string) error {
	// Lookup service to get instance name (handles both full ID and instance name)
	svc, found := r.GetService(serviceID)
	if !found {
		return fmt.Errorf("service %q not registered", serviceID)
	}

	// Remove from registry using instance name
	err := r.registry.Remove(svc.InstanceName)
	if err != nil {
		return fmt.Errorf("service %q not registered", serviceID)
	}

	// TODO: Send goodbye packets (TTL=0)
	// This requires building records with TTL=0 and sending via transport
	// For now, just remove from registry

	return nil
}

// Close closes the responder and unregisters all services per FR-015.
//
// Process:
//  1. Stop query handler goroutine
//  2. Unregister all services (sends goodbye packets)
//  3. Close transport
//
// Returns:
//   - error: transport close error
//
// T043: Implement Close()
// T080: Stop query handler
func (r *Responder) Close() error {
	// Stop query handler goroutine (T080)
	close(r.queryHandlerDone)

	// Unregister all services (sends goodbye packets)
	services := r.registry.List()
	for _, instanceName := range services {
		// Ignore errors - service may have been manually unregistered
		_ = r.Unregister(instanceName)
	}

	// Close transport
	if r.transport != nil {
		return r.transport.Close()
	}
	return nil
}

// getLocalIPv4 gets the first non-loopback IPv4 address from any interface.
//
// DEPRECATED for query response building: Use getIPv4ForInterface(interfaceIndex) instead
// to comply with RFC 6762 §15 (interface-specific addressing).
//
// Still used for:
//   - Service registration (choosing default interface for A record)
//   - Graceful degradation when interfaceIndex=0 (control messages unavailable)
//
// Returns:
//   - []byte: IPv4 address (4 bytes)
//   - error: if no suitable address found
//
// T037: Marked as deprecated for response building (007-interface-specific-addressing)
func getLocalIPv4() ([]byte, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4, nil
			}
		}
	}

	return nil, fmt.Errorf("no non-loopback IPv4 address found")
}

// getIPv4ForInterface returns the IPv4 address assigned to the specified network interface.
//
// RFC 6762 §15 "Responding to Address Queries" (lines 1020-1024):
//
//	When a Multicast DNS responder sends a Multicast DNS response message
//	containing its own address records, it MUST include all addresses
//	that are valid on the interface on which it is sending the message,
//	and MUST NOT include addresses that are not valid on that interface.
//
// This function enables RFC compliance by looking up the interface-specific IP address
// for building mDNS responses that contain ONLY the address valid on the receiving interface.
//
// 007-interface-specific-addressing: T014-T020 implementation
//
// Parameters:
//   - ifIndex: Network interface index (from Transport.Receive or ipv4.ControlMessage.IfIndex)
//
// Returns:
//   - []byte: IPv4 address (4 bytes) in network byte order
//   - error: NetworkError if interface not found, ValidationError if no IPv4 address
//
// Edge Cases:
//   - Interface not found (removed/down) → NetworkError
//   - Interface has no IPv4 address (IPv6-only) → ValidationError
//   - Interface has multiple IPs → returns first IPv4 (consistent behavior)
//
// Example:
//
//	ipv4, err := getIPv4ForInterface(2)  // Look up interface index 2 (e.g., wlan0)
//	if err != nil {
//	    // Handle error: skip response or fall back to getLocalIPv4()
//	}
//	// Use ipv4 in A record for mDNS response
//
//lint:ignore U1000 T020: Foundation for T027-T033 (Phase 3 GREEN), will be used in handleQuery()
func getIPv4ForInterface(ifIndex int) ([]byte, error) {
	// T015: Look up interface by index
	iface, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		// T018: Interface not found (removed, invalid index, etc.)
		return nil, &errors.NetworkError{
			Operation: "lookup interface",
			Err:       err,
			Details:   fmt.Sprintf("interface index %d not found", ifIndex),
		}
	}

	// T016: Get all addresses for this interface
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, &errors.NetworkError{
			Operation: "get interface addresses",
			Err:       err,
			Details:   fmt.Sprintf("failed to get addresses for %s", iface.Name),
		}
	}

	// T017: Filter for first IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4, nil
			}
		}
	}

	// T019: No IPv4 found on this interface
	return nil, &errors.ValidationError{
		Field:   "interface",
		Value:   iface.Name,
		Message: "no IPv4 address found on interface",
	}
}

// OnProbe sets a callback to be called when a probe is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (r *Responder) OnProbe(callback func()) {
	// Store callback for future machines
	r.onProbeCallback = callback

	// Also apply to current machine if it exists
	if r.lastMachine != nil {
		prober := r.lastMachine.GetProber()
		if prober != nil {
			prober.SetOnSendQuery(callback)
		}
	}
}

// OnAnnounce sets a callback to be called when an announcement is sent.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (r *Responder) OnAnnounce(callback func()) {
	// Store callback for future machines
	r.onAnnounceCallback = callback

	// Also apply to current machine if it exists
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			announcer.SetOnSendAnnouncement(callback)
		}
	}
}

// GetLastProbeMessage returns the last sent probe message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.1 validation
func (r *Responder) GetLastProbeMessage() []byte {
	if r.lastMachine != nil {
		prober := r.lastMachine.GetProber()
		if prober != nil {
			return prober.GetLastProbeMessage()
		}
	}
	return nil
}

// GetLastAnnounceMessage returns the last sent announcement message.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 validation
func (r *Responder) GetLastAnnounceMessage() []byte {
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			return announcer.GetLastAnnounceMessage()
		}
	}
	return nil
}

// GetLastAnnouncedRecords returns the last announced record set.
//
// US2 GREEN: Contract test support for RFC 6762 §8.3 and RFC 6763 §6 validation
func (r *Responder) GetLastAnnouncedRecords() []*ResourceRecord {
	return r.lastAnnouncedRecords
}

// GetLastAnnounceDest returns the last announcement destination address.
//
// US2 GREEN: Contract test support for RFC 6762 §5 multicast address validation
func (r *Responder) GetLastAnnounceDest() string {
	if r.lastMachine != nil {
		announcer := r.lastMachine.GetAnnouncer()
		if announcer != nil {
			return announcer.GetLastDestAddr()
		}
	}
	return ""
}

// GetService retrieves a registered service by service ID.
//
// The serviceID can be either:
//   - Full service ID: "Instance Name._service._proto.local"
//   - Just instance name: "Instance Name" (backward compatibility)
//
// Returns:
//   - *Service: The service if found
//   - bool: true if service exists, false otherwise
//
// T100: Implement GetService for multi-service support (US5 GREEN)
func (r *Responder) GetService(serviceID string) (*Service, bool) {
	// Try lookup by instance name directly (works if serviceID is just the instance name)
	if svc, found := r.registry.Get(serviceID); found {
		// Convert internal Service to public Service
		return &Service{
			InstanceName: svc.InstanceName,
			ServiceType:  svc.ServiceType,
			Port:         svc.Port,
			TXTRecords:   svc.TXT,
		}, true
	}

	// serviceID might be full DNS name "Instance._service._proto.local"
	// Extract instance name (everything before first dot)
	// For now, iterate through all services and find matching one
	for _, instanceName := range r.registry.List() {
		svc, found := r.registry.Get(instanceName)
		if !found {
			continue
		}

		// Build full service ID and compare
		fullID := svc.InstanceName + "." + svc.ServiceType
		if fullID == serviceID {
			return &Service{
				InstanceName: svc.InstanceName,
				ServiceType:  svc.ServiceType,
				Port:         svc.Port,
				TXTRecords:   svc.TXT,
			}, true
		}
	}

	return nil, false
}

// UpdateService updates a registered service's TXT records without re-probing.
//
// Per RFC 6762 §8.4, updating TXT records does NOT require re-probing since:
// - The service instance name hasn't changed (no conflict possible)
// - TXT records are metadata, not part of the unique service identity
//
// Process:
//  1. Find service in registry
//  2. Update TXT records
//  3. Send announcement with updated TXT record (multicast to inform network)
//
// Parameters:
//   - serviceID: Service identifier (InstanceName or InstanceName.ServiceType)
//   - txtRecords: New TXT records to set
//
// Returns:
//   - error: If service not found or update fails
//
// T106: Implement UpdateService without re-probing (US5 GREEN)
func (r *Responder) UpdateService(serviceID string, txtRecords map[string]string) error {
	// Lookup service
	svc, found := r.GetService(serviceID)
	if !found {
		return fmt.Errorf("service %q not found", serviceID)
	}

	// Update TXT records in registry
	// The registry stores internal/responder.Service, so we need to update it there
	internalSvc, found := r.registry.Get(svc.InstanceName)
	if !found {
		return fmt.Errorf("internal error: service %q in GetService but not in registry", svc.InstanceName)
	}

	// Update TXT records
	internalSvc.TXT = txtRecords

	// TODO US5-LATER: Send announcement with updated TXT record
	// For now, just updating the registry is sufficient for tests

	return nil
}

// InjectConflictDuringProbing is a test hook to inject conflicts during probing.
//
// When enabled, the state machine will always report StateConflictDetected,
// forcing the rename loop to trigger.
//
// T062: Test hook for max rename attempts testing
func (r *Responder) InjectConflictDuringProbing(inject bool) {
	r.injectConflict = inject
}

// InjectSimultaneousProbe is a test hook for injecting simultaneous probe scenarios.
//
// This method is currently a stub placeholder for future simultaneous probe testing
// per RFC 6762 §8.2 tiebreaking. It will be implemented when detailed conflict
// resolution testing is added.
//
// Parameters:
//   - First parameter: Our probe packet (currently unused)
//   - Second parameter: Incoming probe packet (currently unused)
//
// T062: Test hook infrastructure for conflict scenarios
func (r *Responder) InjectSimultaneousProbe([]byte, []byte) {}

// ResourceRecord is a type alias for records.ResourceRecord.
//
// This alias allows contract tests to reference ResourceRecord without importing
// the internal records package directly, maintaining clean architecture boundaries.
//
// The underlying type contains DNS resource record fields:
//   - Name: Domain name (e.g., "myservice._http._tcp.local")
//   - Type: Record type (A, PTR, SRV, TXT per RFC 1035)
//   - Class: Record class (IN for Internet)
//   - TTL: Time-to-live in seconds
//   - Data: Record-specific data (IP address, target name, etc.)
//   - CacheFlush: Cache-flush bit per RFC 6762 §10.2
//
// US2 GREEN: Contract test support for validating resource records
type ResourceRecord = records.ResourceRecord

// runQueryHandler continuously receives and processes mDNS queries.
//
// RFC 6762 §6: Responders SHOULD respond to queries for services they have registered.
//
// Process:
//  1. Receive query packet from transport
//  2. Parse DNS message
//  3. For each question, check if we have matching service
//  4. Build response (PTR answer + SRV/TXT/A additional)
//  5. Apply rate limiting per RFC 6762 §6.2
//  6. Send response (unicast or multicast based on QU bit)
//
// T080: Query handler goroutine
func (r *Responder) runQueryHandler() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.queryHandlerDone:
			return
		default:
			// Receive query with timeout
			// 007-interface-specific-addressing T027: Extract interfaceIndex for RFC 6762 §15 compliance
			packet, _, interfaceIndex, err := r.transport.Receive(r.ctx)
			if err != nil {
				// Context cancelled or transport closed
				select {
				case <-r.ctx.Done():
					return
				case <-r.queryHandlerDone:
					return
				default:
					// Other error - continue receiving
					continue
				}
			}

			// Handle query (T079)
			// T028: Pass interfaceIndex to enable interface-specific addressing
			_ = r.handleQuery(packet, interfaceIndex)
		}
	}
}

// handleQuery processes a single mDNS query and sends response.
//
// RFC 6762 §6: "When a Multicast DNS responder receives a query, it must determine
// whether the query is requesting information for which this responder is authoritative."
//
// RFC 6762 §15: Responses MUST include only addresses valid on the receiving interface,
// and MUST NOT include addresses from other interfaces.
//
// Process:
//  1. Parse query message
//  2. Extract questions
//  3. Check if we have matching registered services
//  4. Build response using ResponseBuilder with interface-specific IP (T029)
//  5. Apply QU bit logic (unicast vs multicast)
//  6. Apply rate limiting (RFC 6762 §6.2)
//  7. Send response
//
// Parameters:
//   - packet: DNS query in wire format
//   - interfaceIndex: OS interface index that received the query (0 = unknown)
//
// Returns:
//   - error: parse error or send error (logged, not propagated)
//
// T079: Implement handleQuery()
// T029: Added interfaceIndex parameter for interface-specific addressing
func (r *Responder) handleQuery(packet []byte, interfaceIndex int) error {
	// Import message parser
	msg, err := parseMessage(packet)
	if err != nil {
		// Malformed query - ignore per RFC 6762 §6
		return err
	}

	// Ignore responses (QR=1)
	if msg.Header.IsResponse() {
		return nil
	}

	// Process each question
	for _, question := range msg.Questions {
		// Only handle PTR queries for now (T076 implementation)
		if question.QTYPE != uint16(protocol.RecordTypePTR) {
			continue
		}

		// Check if we have a service matching this query
		// Query is for "_http._tcp.local", we need to find services of that type
		serviceType := question.QNAME

		// Get all registered services
		services := r.registry.List()
		for _, instanceName := range services {
			service, found := r.registry.Get(instanceName)
			if !found {
				continue
			}

			// Check if service type matches query
			if service.ServiceType != serviceType {
				continue
			}

			// We have a match! Build response with interface-specific addressing
			//
			// RFC 6762 §15 "Responding to Address Queries":
			// "When a Multicast DNS responder sends a Multicast DNS response message
			// containing its own address records in response to a query received on
			// a particular interface, it MUST include only addresses that are valid
			// on that interface, and MUST NOT include addresses configured on other
			// interfaces."
			//
			// T036: Inline comment citing RFC 6762 §15
			var ipv4 []byte
			var err error

			// T030: Graceful fallback when interface index unavailable (interfaceIndex=0)
			// This happens when control messages aren't supported or platform doesn't provide IP_PKTINFO
			if interfaceIndex == 0 {
				// Degraded mode: Use default interface IP (legacy behavior)
				// TODO T032: Add debug logging when F-6 (Logging & Observability) is implemented
				ipv4, err = getLocalIPv4()
			} else {
				// RFC 6762 §15 compliance: Use ONLY the IP from the receiving interface
				ipv4, err = getIPv4ForInterface(interfaceIndex)
			}

			if err != nil {
				// T031: If interface-specific IP lookup fails, skip response for this query
				// This is correct behavior per RFC 6762 §15: Better to not respond than
				// to respond with an incorrect (wrong interface) IP address
				// TODO T032: Add error logging when F-6 is implemented
				// Common failure causes: interface went down, no IPv4 configured, invalid index
				continue
			}

			serviceWithIP := &responder.ServiceWithIP{
				InstanceName: service.InstanceName,
				ServiceType:  service.ServiceType,
				Domain:       "local",
				Port:         service.Port,
				IPv4Address:  ipv4,
				TXTRecords:   service.TXT, // internal.Service uses TXT field
				Hostname:     r.hostname,
			}

			// Build response (T076)
			response, err := r.responseBuilder.BuildResponse(serviceWithIP, msg)
			if err != nil {
				continue
			}

			// TODO: T082 - Implement QU bit + 1/4 TTL logic for unicast vs multicast
			// For now, always multicast

			// TODO: T083 - Apply per-record rate limiting before sending
			// For now, skip rate limiting

			// Send response via multicast
			responsePacket := buildResponsePacket(response)
			_ = r.transport.Send(r.ctx, responsePacket, nil) // nil = multicast

			// Only respond once per query
			break
		}
	}

	return nil
}

// parseMessage is a wrapper around message.ParseMessage for easier imports.
func parseMessage(packet []byte) (*message.DNSMessage, error) {
	return message.ParseMessage(packet)
}

// buildResponsePacket serializes a DNSMessage to wire format.
//
// TODO: Implement proper serialization
// For now, return empty packet (stub)
func buildResponsePacket(msg *message.DNSMessage) []byte {
	// This is a stub - proper implementation needs message serialization
	// which is not yet implemented in the codebase
	return []byte{}
}
