package responder

import (
	"context"
	"fmt"
	"net"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/records"
	"github.com/joshuafuller/beacon/internal/state"
	"github.com/joshuafuller/beacon/internal/transport"
)

// maxRenameAttempts is the maximum number of times to rename a service on conflict.
//
// RFC 6762 §9: No explicit limit specified, but we use 10 as a reasonable maximum
// to prevent infinite loops and resource exhaustion.
//
// FR-032: System MUST handle registration failures gracefully
const maxRenameAttempts = 10

// Register registers a service with probing and announcing per RFC 6762 §8.
//
// IMPORTANT: Register blocks for approximately 1.75 seconds while performing
// the required probing (3 probes × 250ms) and announcing (2 announcements × 1s)
// phases per RFC 6762 §8. Use a goroutine if non-blocking behavior is needed.
//
// Process:
//  1. Validate service parameters
//  2. Probe for name conflicts (RFC 6762 §8.1, ~750ms)
//  3. Announce the service (RFC 6762 §8.3, ~1s)
//  4. Add to registry on success
//
// If a naming conflict is detected during probing, the service is automatically
// renamed per RFC 6762 §9 (e.g., "My Service" → "My Service-2") and probing
// restarts, up to 10 attempts.
//
// Returns:
//   - error: validation error, conflict error, max attempts error, or context error
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
		serviceInfo := buildServiceInfo(service.InstanceName, service.ServiceType,
			service.Hostname, service.Port, ipv4, service.TXTRecords)
		recordSet := records.BuildRecordSet(serviceInfo)

		// US2 GREEN: Store record set for contract test validation
		r.lastAnnouncedRecords = recordSet

		serviceName := service.InstanceName + "." + service.ServiceType
		finalState, err := r.runProbeAndAnnounce(recordSet, serviceName)
		if err != nil {
			return err
		}

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
		// US5: toInternalService carries TXT records for UpdateService support
		if err := r.registry.Register(toInternalService(service)); err != nil {
			return fmt.Errorf("failed to add to registry: %w", err)
		}

		return nil // Successfully registered
	}

	// Should never reach here (loop returns on success or max attempts)
	return fmt.Errorf("unexpected: register loop completed without result")
}

// runProbeAndAnnounce builds a state machine for recordSet/serviceName, wires
// the transport and any test hooks, runs the probing + announcing sequence
// (RFC 6762 §8), and returns the resulting state.
func (r *Responder) runProbeAndAnnounce(recordSet []*message.ResourceRecord, serviceName string) (state.State, error) {
	machine := state.NewMachine()

	// Wire transport so probes and announcements are sent in every configured
	// .local zone (RFC 6762 §20).
	registrationTransport, err := r.registrationTransport(recordSet)
	if err != nil {
		var zero state.State
		return zero, fmt.Errorf("build registration transport: %w", err)
	}
	machine.SetTransport(registrationTransport)

	// Apply test hooks (if any)
	if r.injectConflict {
		machine.SetInjectConflict(true)
	}

	// US2 GREEN: Store machine for message capture (contract test support)
	r.lastMachine = machine

	// US2 GREEN: Apply callbacks to new machine (if any)
	if r.onProbeCallback != nil {
		if prober := machine.GetProber(); prober != nil {
			prober.SetOnSendQuery(r.onProbeCallback)
		}
	}
	if r.onAnnounceCallback != nil {
		if announcer := machine.GetAnnouncer(); announcer != nil {
			announcer.SetOnSendAnnouncement(r.onAnnounceCallback)
		}
	}

	// Provide resource records to announcer for DNS message serialization
	if announcer := machine.GetAnnouncer(); announcer != nil {
		announcer.SetRecords(recordSet)
	}

	// Run state machine (probing + announcing)
	if err := machine.Run(r.ctx, serviceName); err != nil {
		var zero state.State
		return zero, fmt.Errorf("state machine failed: %w", err)
	}

	return machine.GetState(), nil
}

// registrationMulticastTransport mirrors state-machine lifecycle sends into
// the independent IPv4 and IPv6 .local zones described by RFC 6762 §20. It
// borrows the responder's transports and therefore never closes them.
type registrationMulticastTransport struct {
	ipv4              transport.Transport
	ipv6              transport.Transport
	ipv6Announcements map[int][]byte
}

type interfaceScopedIPv6Transport interface {
	transport.Transport
	JoinedInterfaceIndexes() []int
	SendOnInterface(context.Context, []byte, net.Addr, int) error
}

func (t *registrationMulticastTransport) Send(ctx context.Context, packet []byte, destination net.Addr) error {
	err := t.ipv4.Send(ctx, packet, destination)
	if t.ipv6 != nil {
		if scoped, ok := t.ipv6.(interfaceScopedIPv6Transport); ok && isDNSResponse(packet) && len(t.ipv6Announcements) > 0 {
			for _, ifIndex := range scoped.JoinedInterfaceIndexes() {
				projected, found := t.ipv6Announcements[ifIndex]
				if !found {
					if err == nil {
						err = fmt.Errorf("missing IPv6 announcement projection for interface %d", ifIndex)
					}
					continue
				}
				if ipv6Err := scoped.SendOnInterface(ctx, projected, protocol.MulticastGroupIPv6(), ifIndex); ipv6Err != nil && err == nil {
					err = ipv6Err
				}
			}
		} else if ipv6Err := t.ipv6.Send(ctx, packet, protocol.MulticastGroupIPv6()); err == nil {
			err = ipv6Err
		}
	}
	return err
}

func isDNSResponse(packet []byte) bool {
	return len(packet) >= 3 && packet[2]&0x80 != 0
}

func (t *registrationMulticastTransport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
	return t.ipv4.Receive(ctx)
}

func (*registrationMulticastTransport) Close() error { return nil }

func (r *Responder) registrationTransport(recordSet []*message.ResourceRecord) (transport.Transport, error) {
	if r.transport6 == nil {
		return r.transport, nil
	}
	registrationTransport := &registrationMulticastTransport{ipv4: r.transport, ipv6: r.transport6}
	if scoped, ok := r.transport6.(interfaceScopedIPv6Transport); ok {
		projections, err := buildIPv6AnnouncementProjections(recordSet, scoped.JoinedInterfaceIndexes(), lookupLifecycleAddresses)
		if err != nil {
			return nil, err
		}
		registrationTransport.ipv6Announcements = projections
	}
	return registrationTransport, nil
}

func projectLifecycleRecordsForInterface(recordSet []*message.ResourceRecord, ipv4 []byte, ipv6Addresses []net.IP) []*message.ResourceRecord {
	projected := make([]*message.ResourceRecord, 0, len(recordSet)+len(ipv6Addresses))
	var hostname string
	addressClass := protocol.ClassIN
	addressTTL := uint32(protocol.TTLHostname)
	addressCacheFlush := true
	for _, record := range recordSet {
		if record.Type == protocol.RecordTypeA || record.Type == protocol.RecordTypeAAAA {
			if hostname == "" {
				hostname = record.Name
				addressClass = record.Class
				addressTTL = record.TTL
				addressCacheFlush = record.CacheFlush
			}
			continue
		}
		clone := *record
		clone.Data = append([]byte(nil), record.Data...)
		projected = append(projected, &clone)
	}
	if hostname == "" {
		return projected
	}
	if len(ipv4) == net.IPv4len {
		projected = append(projected, &message.ResourceRecord{
			Name:       hostname,
			Type:       protocol.RecordTypeA,
			Class:      addressClass,
			TTL:        addressTTL,
			Data:       append([]byte(nil), ipv4...),
			CacheFlush: addressCacheFlush,
		})
	}
	for _, ip := range ipv6Addresses {
		if ip16 := ip.To16(); ip16 != nil && ip.To4() == nil {
			projected = append(projected, &message.ResourceRecord{
				Name:       hostname,
				Type:       protocol.RecordTypeAAAA,
				Class:      addressClass,
				TTL:        addressTTL,
				Data:       append([]byte(nil), ip16...),
				CacheFlush: addressCacheFlush,
			})
		}
	}
	return projected
}

type lifecycleAddressLookup func(int) ([]byte, []net.IP, error)

func buildIPv6AnnouncementProjections(recordSet []*message.ResourceRecord, interfaceIndexes []int, lookup lifecycleAddressLookup) (map[int][]byte, error) {
	projections := make(map[int][]byte, len(interfaceIndexes))
	for _, ifIndex := range interfaceIndexes {
		ipv4, ipv6Addresses, err := lookup(ifIndex)
		if err != nil {
			return nil, fmt.Errorf("lookup addresses for interface %d: %w", ifIndex, err)
		}
		recordsForInterface := projectLifecycleRecordsForInterface(recordSet, ipv4, ipv6Addresses)
		packet, err := message.BuildResponse(recordsForInterface)
		if err != nil {
			return nil, fmt.Errorf("build announcement for interface %d: %w", ifIndex, err)
		}
		projections[ifIndex] = packet
	}
	return projections, nil
}

func lookupLifecycleAddresses(ifIndex int) ([]byte, []net.IP, error) {
	ipv6Addresses, err := getIPv6ForInterface(ifIndex)
	if err != nil {
		return nil, nil, err
	}
	ipv4, err := getIPv4ForInterface(ifIndex)
	if err != nil {
		// IPv4 is optional when the interface is IPv6-only.
		return nil, ipv6Addresses, nil
	}
	return ipv4, ipv6Addresses, nil
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

	// Get local IPv4 address for goodbye records
	ipv4, err := getLocalIPv4()
	if err != nil {
		// If we can't get IP, still remove from registry but skip goodbye
		_ = r.registry.Remove(svc.InstanceName) // nosemgrep: beacon-error-swallowing
		return fmt.Errorf("failed to get local IP for goodbye: %w", err)
	}

	// Build goodbye records with TTL=0 (RFC 6762 §10.1)
	serviceInfo := buildServiceInfo(svc.InstanceName, svc.ServiceType,
		r.hostname, svc.Port, ipv4, svc.TXTRecords)
	goodbyeRecords := records.BuildGoodbyeRecords(serviceInfo)

	// Build and send goodbye packet
	goodbyePacket, err := message.BuildResponse(goodbyeRecords)
	if err != nil {
		// If we can't build packet, still remove from registry
		_ = r.registry.Remove(svc.InstanceName) // nosemgrep: beacon-error-swallowing
		return fmt.Errorf("failed to build goodbye packet: %w", err)
	}

	// RFC 6762 §10.1: Goodbye is best-effort (SHOULD, not MUST).
	// Ignore send errors; we still remove from the registry below.
	_ = r.sendLifecyclePacket(goodbyePacket, goodbyeRecords, lookupLifecycleAddresses) // nosemgrep: beacon-error-swallowing

	// Remove from registry using instance name
	if err := r.registry.Remove(svc.InstanceName); err != nil {
		return fmt.Errorf("service %q not registered", serviceID)
	}

	return nil
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
		return fromInternalService(svc), true
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
			return fromInternalService(svc), true
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

	// Announce updated records per RFC 6762 §8.4.
	// The registry is already updated above; the multicast announcement below is
	// best-effort (RFC 6762 §8.4 is a SHOULD), so failures to obtain an address,
	// build, or send the packet do not roll back the update.
	ipv4, err := getLocalIPv4()
	if err != nil {
		return nil // Registry updated; cannot announce without an IP (best-effort).
	}

	serviceInfo := buildServiceInfo(svc.InstanceName, svc.ServiceType,
		r.hostname, svc.Port, ipv4, txtRecords)
	announcedRecords := records.BuildRecordSet(serviceInfo)

	// Convert to message.ResourceRecord for BuildResponse
	msgRecords := make([]*message.ResourceRecord, len(announcedRecords))
	for i, rr := range announcedRecords {
		msgRecords[i] = &message.ResourceRecord{
			Name:       rr.Name,
			Type:       rr.Type,
			Class:      rr.Class,
			TTL:        rr.TTL,
			Data:       rr.Data,
			CacheFlush: rr.CacheFlush,
		}
	}

	_ = r.sendLifecycleRecords(msgRecords, lookupLifecycleAddresses) // nosemgrep: beacon-error-swallowing

	return nil
}

func (r *Responder) sendLifecycleRecords(recordSet []*message.ResourceRecord, lookup lifecycleAddressLookup) error {
	packet, err := message.BuildResponse(recordSet)
	if err != nil {
		return err
	}
	return r.sendLifecyclePacket(packet, recordSet, lookup)
}

func (r *Responder) sendLifecyclePacket(packet []byte, recordSet []*message.ResourceRecord, lookup lifecycleAddressLookup) error {
	lifecycleTransport := &registrationMulticastTransport{ipv4: r.transport, ipv6: r.transport6}
	if scoped, ok := r.transport6.(interfaceScopedIPv6Transport); ok {
		projections, err := buildIPv6AnnouncementProjections(recordSet, scoped.JoinedInterfaceIndexes(), lookup)
		if err != nil {
			return err
		}
		lifecycleTransport.ipv6Announcements = projections
	}
	return lifecycleTransport.Send(r.ctx, packet, protocol.MulticastGroupIPv4())
}
