package responder

import (
	"net"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/responder"
	"github.com/joshuafuller/beacon/internal/transport"
)

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
func (r *Responder) runQueryHandler(tr transport.Transport, ipv6Network bool) {
	defer r.queryHandlerWg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.queryHandlerDone:
			return
		default:
			// Receive query with timeout
			// 007-interface-specific-addressing T027: Extract interfaceIndex for RFC 6762 §6.2 compliance
			// Task 2: Capture source address for subnet validation (RFC 6762 §6.4)
			packet, srcAddr, interfaceIndex, err := tr.Receive(r.ctx)
			if err != nil {
				// Context canceled or transport closed
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
			// Task 2: Pass source address for subnet validation
			_ = r.handleQueryOnTransport(tr, ipv6Network, packet, srcAddr, interfaceIndex)
		}
	}
}

// validateSourceAddressOnNetwork validates that a unicast query source is on
// the same subnet or IPv6 on-link prefix as the receiving interface.
//
// RFC 6762 §11 defines the IPv4 subnet-mask and IPv6 on-link-prefix checks.
//
// Parameters:
//   - srcAddr: Source address of the query
//   - interfaceIndex: OS interface index that received the query
//
// Returns:
//   - bool: true if source is on same subnet, false otherwise
//
// Task 2: Source address validation
func validateSourceAddressOnNetwork(srcAddr net.Addr, interfaceIndex int, ipv6Network bool) bool {
	// If interface index is unknown (0), skip validation (graceful degradation)
	if interfaceIndex == 0 {
		return true
	}

	// Extract IP from source address
	udpAddr, ok := srcAddr.(*net.UDPAddr)
	if !ok {
		return false
	}
	// Get interface by index
	iface, err := net.InterfaceByIndex(interfaceIndex)
	if err != nil {
		return false
	}

	// Get interface addresses
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}

	// Check if source IP is on the same IPv4 subnet or IPv6 on-link prefix as
	// any address configured on the receiving interface (RFC 6762 §11).
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipv6Network {
			if udpAddr.IP.To4() == nil && ipnet.IP.To4() == nil && ipnet.Contains(udpAddr.IP) {
				return true
			}
			continue
		}
		srcIP := udpAddr.IP.To4()
		if srcIP != nil && ipnet.IP.To4() != nil && ipnet.Contains(srcIP) {
			return true
		}
	}

	// Source IP not on same subnet
	return false
}

// handleQuery processes a single mDNS query and sends response.
//
// RFC 6762 §6: "When a Multicast DNS responder receives a query, it must determine
// whether the query is requesting information for which this responder is authoritative."
//
// RFC 6762 §11: For unicast traffic, source addresses are checked against the
// receiving interface's IPv4 subnets or IPv6 on-link prefixes.
//
// RFC 6762 §6.2: Responses MUST include only addresses valid on the receiving interface,
// and MUST NOT include addresses from other interfaces.
//
// Process:
//  1. Parse query message
//  2. Validate source address (RFC 6762 §6.4)
//  3. Extract questions
//  4. Check if we have matching registered services
//  5. Build response using ResponseBuilder with interface-specific IP (T029)
//  6. Apply QU bit logic (unicast vs multicast)
//  7. Apply rate limiting (RFC 6762 §6.2)
//  8. Send response
//
// Parameters:
//   - packet: DNS query in wire format
//   - srcAddr: Source address of the query
//   - interfaceIndex: OS interface index that received the query (0 = unknown)
//
// Returns:
//   - error: parse error or send error (logged, not propagated)
//
// T079: Implement handleQuery()
// T029: Added interfaceIndex parameter for interface-specific addressing
// Task 2: Added srcAddr parameter for source address validation
func (r *Responder) handleQuery(tr transport.Transport, packet []byte, srcAddr net.Addr, interfaceIndex int) error {
	_, ipv6Network := tr.(*transport.UDPv6Transport)
	return r.handleQueryOnTransport(tr, ipv6Network, packet, srcAddr, interfaceIndex)
}

func (r *Responder) handleQueryOnTransport(tr transport.Transport, ipv6Network bool, packet []byte, srcAddr net.Addr, interfaceIndex int) error {
	// Task 2: RFC 6762 §11 - Validate source address is on-link
	if !validateSourceAddressOnNetwork(srcAddr, interfaceIndex, ipv6Network) {
		// Source not on-link - ignore the query.
		return nil
	}

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
		// RFC 6763 §9: Service Type Enumeration
		// A PTR query for "_services._dns-sd._udp.local" returns all unique service types.
		if question.QTYPE == uint16(protocol.RecordTypePTR) && question.QNAME == serviceEnumerationName {
			r.handleServiceEnumerationQuery(tr, ipv6Network, question, srcAddr)
			continue
		}

		matchedService := r.findMatchingService(question)
		if matchedService == nil {
			continue
		}

		r.respondToQuery(tr, ipv6Network, matchedService, msg, question, srcAddr, interfaceIndex)
	}

	return nil
}

// serviceEnumerationName is the DNS-SD meta-query name per RFC 6763 §9.
const serviceEnumerationName = "_services._dns-sd._udp.local"

// handleServiceEnumerationQuery responds to a DNS-SD Service Type Enumeration
// query (RFC 6763 §9) with a PTR record for each unique registered service type.
func (r *Responder) handleServiceEnumerationQuery(tr transport.Transport, ipv6Network bool, question message.Question, srcAddr net.Addr) {
	serviceTypes := r.registry.ListServiceTypes()
	if len(serviceTypes) == 0 {
		return // No services registered, no response needed
	}

	// Build PTR records: one for each unique service type
	ptrRecords := make([]*message.ResourceRecord, 0, len(serviceTypes))
	for _, svcType := range serviceTypes {
		// RDATA for PTR record is the encoded service type name
		encodedTarget, encErr := message.EncodeName(svcType)
		if encErr != nil {
			continue // Skip types that cannot be encoded
		}
		ptrRecords = append(ptrRecords, &message.ResourceRecord{
			Name:       serviceEnumerationName,
			Type:       protocol.RecordTypePTR,
			Class:      protocol.ClassIN,
			TTL:        protocol.TTLHostname, // 4500s per RFC 6762 §10
			Data:       encodedTarget,
			CacheFlush: false, // PTR is a shared record
		})
	}

	if len(ptrRecords) == 0 {
		return
	}

	// Build DNS response message
	responseMsg, buildErr := message.BuildResponse(ptrRecords)
	if buildErr != nil {
		return
	}

	// Determine destination (unicast vs multicast based on QU bit)
	quBit := (question.QCLASS & 0x8000) != 0
	var dest net.Addr
	if quBit {
		dest = srcAddr
	} else if ipv6Network {
		dest = protocol.MulticastGroupIPv6()
	}
	// nil dest = multicast to 224.0.0.251:5353

	_ = tr.Send(r.ctx, responseMsg, dest)
}

// findMatchingService finds a registered service matching question's QTYPE/QNAME.
func (r *Responder) findMatchingService(question message.Question) *responder.Service {
	services := r.registry.List()

	for _, instanceName := range services {
		service, found := r.registry.Get(instanceName)
		if !found {
			continue
		}

		switch question.QTYPE {
		case uint16(protocol.RecordTypePTR):
			// PTR: match by service type (e.g., "_http._tcp.local")
			if service.ServiceType == question.QNAME {
				return service
			}
		case uint16(protocol.RecordTypeSRV), uint16(protocol.RecordTypeTXT):
			// SRV/TXT: match by full instance name (e.g., "My Printer._http._tcp.local")
			fullName := service.InstanceName + "." + service.ServiceType
			if fullName == question.QNAME {
				return service
			}
		case uint16(protocol.RecordTypeA), uint16(protocol.RecordTypeAAAA):
			// A: match by hostname (e.g., "myhost.local")
			if r.hostname == question.QNAME {
				return service
			}
		}
	}

	return nil
}

// respondToQuery builds and sends a response for a matched service using the
// query's receiving interface (RFC 6762 §6.2), applying per-source rate limiting
// (RFC 6762 §6.2) and unicast/multicast destination selection based on the QU
// bit (RFC 6762 §5.4).
//
// RFC 6762 §6.2 "Responding to Address Queries":
// "When a Multicast DNS responder sends a Multicast DNS response message
// containing its own address records in response to a query received on
// a particular interface, it MUST include only addresses that are valid
// on that interface, and MUST NOT include addresses configured on other
// interfaces."
func (r *Responder) respondToQuery(tr transport.Transport, ipv6Network bool, matchedService *responder.Service, msg *message.DNSMessage, question message.Question, srcAddr net.Addr, interfaceIndex int) {
	ipv4, ipv6Addresses, ipErr := getResponseAddresses(ipv6Network, interfaceIndex)

	if ipErr != nil {
		// T031: If interface-specific IP lookup fails, skip response for this query
		// This is correct behavior per RFC 6762 §6.2: Better to not respond than
		// to respond with an incorrect (wrong interface) IP address
		// TODO T032: Add error logging when F-6 is implemented
		// Common failure causes: interface went down, no IPv4 configured, invalid index
		return
	}

	serviceWithIP := &responder.ServiceWithIP{
		InstanceName:  matchedService.InstanceName,
		ServiceType:   matchedService.ServiceType,
		Domain:        "local",
		Port:          matchedService.Port,
		IPv4Address:   ipv4,
		IPv6Addresses: ipv6Addresses,
		TXTRecords:    matchedService.TXT, // internal.Service uses TXT field
		Hostname:      r.hostname,
	}

	// Build response (T076)
	response, err := r.responseBuilder.BuildResponse(serviceWithIP, msg)
	if err != nil {
		return
	}

	// Per-source-IP rate limiting (FR-026, RFC 6762 §6.2)
	if r.rateLimiter != nil && srcAddr != nil {
		srcIP := srcAddr.String()
		if udpAddr, ok := srcAddr.(*net.UDPAddr); ok {
			srcIP = udpAddr.IP.String()
		}
		if !r.rateLimiter.Allow(srcIP) {
			return // Rate-limited, skip response
		}
	}

	// RFC 6762 §5.4: Check QU bit (bit 15 of QCLASS) to determine unicast vs multicast
	// Task 4: QU bit handling
	quBit := (question.QCLASS & 0x8000) != 0

	var dest net.Addr
	if quBit {
		// RFC 6762 §5.4: QU bit set → send unicast response to querier
		dest = srcAddr
	} else if ipv6Network {
		dest = protocol.MulticastGroupIPv6()
	} else {
		// RFC 6762 §5.4: QU bit clear → send multicast response
		dest = nil // nil = multicast to 224.0.0.251:5353
	}

	// Send response
	responsePacket := buildResponsePacket(response)
	_ = tr.Send(r.ctx, responsePacket, dest)
}

func ipBytes(ips []net.IP) [][]byte {
	result := make([][]byte, 0, len(ips))
	for _, ip := range ips {
		if ip16 := ip.To16(); ip16 != nil && ip.To4() == nil {
			result = append(result, append([]byte(nil), ip16...))
		}
	}
	return result
}

// parseMessage is a wrapper around message.ParseMessage for easier imports.
func parseMessage(packet []byte) (*message.DNSMessage, error) {
	return message.ParseMessage(packet)
}

// buildResponsePacket serializes a DNSMessage to wire format using message.SerializeMessage.
//
// RFC 1035 §4.1: Converts the complete DNSMessage struct (header, questions,
// answers, authority, additional sections) into wire-format bytes.
func buildResponsePacket(msg *message.DNSMessage) []byte {
	data, err := message.SerializeMessage(msg)
	if err != nil {
		// Serialization failed - return minimal valid DNS response header
		// so the responder doesn't crash on unexpected serialization errors
		return []byte{
			0x00, 0x00, // ID
			0x84, 0x00, // Flags (QR=1, AA=1)
			0x00, 0x00, // QDCount
			0x00, 0x00, // ANCount
			0x00, 0x00, // NSCount
			0x00, 0x00, // ARCount
		}
	}
	return data
}
