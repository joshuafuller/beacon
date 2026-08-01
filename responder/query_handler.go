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
// runQueryHandler continuously receives and processes mDNS queries on the given transport.
//
// IPv6 mode is detected by type-asserting tr to *transport.UDPv6Transport, following
// the same pattern as querier.runReceiveLoop.
//
// T080: Query handler goroutine
func (r *Responder) runQueryHandler(tr transport.Transport) {
	defer r.queryHandlerWg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.queryHandlerDone:
			return
		default:
			// 007-interface-specific-addressing T027: Extract interfaceIndex for RFC 6762 §15 compliance
			packet, srcAddr, interfaceIndex, err := tr.Receive(r.ctx)
			if err != nil {
				select {
				case <-r.ctx.Done():
					return
				case <-r.queryHandlerDone:
					return
				default:
					continue
				}
			}

			// T028: Pass interfaceIndex and transport for interface-specific addressing
			_ = r.handleQuery(tr, packet, srcAddr, interfaceIndex)
		}
	}
}

// validateSourceAddress validates that the query source is on the same subnet as the interface.
//
// RFC 6762 §6.4: Responders MUST only reply to queries from the same subnet.
// For IPv6, link-local addresses (fe80::/10) are always accepted per RFC 4291 §2.5.6.
func validateSourceAddress(srcAddr net.Addr, interfaceIndex int, isIPv6 bool) bool {
	if interfaceIndex == 0 {
		return true
	}

	udpAddr, ok := srcAddr.(*net.UDPAddr)
	if !ok {
		return false
	}

	if isIPv6 {
		ip := udpAddr.IP
		// Accept link-local IPv6 (fe80::/10) per RFC 4291 §2.5.6
		return len(ip) == 16 && ip[0] == 0xfe && ip[1]&0xc0 == 0x80
	}

	srcIP := udpAddr.IP.To4()
	if srcIP == nil {
		return false
	}

	iface, err := net.InterfaceByIndex(interfaceIndex)
	if err != nil {
		return false
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.Contains(srcIP) {
			return true
		}
	}
	return false
}

// handleQuery processes a single mDNS query and sends response.
//
// RFC 6762 §6: "When a Multicast DNS responder receives a query, it must determine
// whether the query is requesting information for which this responder is authoritative."
//
// RFC 6762 §6.4: "When a Multicast DNS responder receives a query, it MUST only respond
// if the source address of the query is on the same subnet as the interface on which
// the query was received."
//
// RFC 6762 §15: Responses MUST include only addresses valid on the receiving interface,
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
// RFC 6762 §11: isIPv6 selects FF02::FB multicast destination and AAAA address resolution.
func (r *Responder) handleQuery(tr transport.Transport, packet []byte, srcAddr net.Addr, interfaceIndex int) error {
	_, isIPv6 := tr.(*transport.UDPv6Transport)

	// Task 2: RFC 6762 §6.4 - Validate source address is on same subnet
	if !validateSourceAddress(srcAddr, interfaceIndex, isIPv6) {
		// Source not on same subnet - ignore query per RFC 6762 §6.4
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

	// DNS-SD meta-query name per RFC 6763 §9
	const serviceEnumerationName = "_services._dns-sd._udp.local"

	// Process each question
	for _, question := range msg.Questions {
		// RFC 6763 §9: Service Type Enumeration
		// A PTR query for "_services._dns-sd._udp.local" returns all unique service types.
		if question.QTYPE == uint16(protocol.RecordTypePTR) && question.QNAME == serviceEnumerationName {
			serviceTypes := r.registry.ListServiceTypes()
			if len(serviceTypes) == 0 {
				continue // No services registered, no response needed
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
				continue
			}

			// Build DNS response message
			responseMsg, buildErr := message.BuildResponse(ptrRecords)
			if buildErr != nil {
				continue
			}

			quBit := (question.QCLASS & 0x8000) != 0
			var dest net.Addr
			if quBit {
				dest = srcAddr
			} else if isIPv6 {
				dest = protocol.MulticastGroupIPv6()
			}
			// else nil = multicast to 224.0.0.251:5353

			_ = tr.Send(r.ctx, responseMsg, dest)
			continue
		}

		// Get all registered services
		services := r.registry.List()

		var matchedService *responder.Service
		for _, instanceName := range services {
			service, found := r.registry.Get(instanceName)
			if !found {
				continue
			}

			switch question.QTYPE {
			case uint16(protocol.RecordTypePTR):
				// PTR: match by service type (e.g., "_http._tcp.local")
				if service.ServiceType == question.QNAME {
					matchedService = service
				}
			case uint16(protocol.RecordTypeSRV), uint16(protocol.RecordTypeTXT):
				// SRV/TXT: match by full instance name (e.g., "My Printer._http._tcp.local")
				fullName := service.InstanceName + "." + service.ServiceType
				if fullName == question.QNAME {
					matchedService = service
				}
			case uint16(protocol.RecordTypeA):
				// A: match by hostname (e.g., "myhost.local")
				if r.hostname == question.QNAME {
					matchedService = service
				}
			case uint16(protocol.RecordTypeAAAA):
				// AAAA: match by hostname per RFC 3596
				if r.hostname == question.QNAME {
					matchedService = service
				}
			}

			if matchedService != nil {
				break
			}
		}

		if matchedService == nil {
			continue
		}

		// AAAA query on IPv6 transport: build AAAA record directly (RFC 3596, RFC 6762 §11)
		if isIPv6 && question.QTYPE == uint16(protocol.RecordTypeAAAA) {
			var ipv6Addr net.IP
			if interfaceIndex == 0 {
				ipv6Addr, _ = getLocalIPv6()
			} else {
				ipv6Addr, _ = getIPv6ForInterface(interfaceIndex)
			}
			if ipv6Addr == nil {
				continue
			}
			aaaaRec := &message.ResourceRecord{
				Name:       r.hostname,
				Type:       protocol.RecordTypeAAAA,
				Class:      protocol.ClassIN,
				TTL:        protocol.TTLHostname,
				Data:       []byte(ipv6Addr),
				CacheFlush: true,
			}
			aaaaMsg, buildErr := message.BuildResponse([]*message.ResourceRecord{aaaaRec})
			if buildErr != nil {
				continue
			}
			if r.rateLimiter != nil && srcAddr != nil {
				srcIP := srcAddr.String()
				if udpAddr, ok := srcAddr.(*net.UDPAddr); ok {
					srcIP = udpAddr.IP.String()
				}
				if !r.rateLimiter.Allow(srcIP) {
					continue
				}
			}
			quBit := (question.QCLASS & 0x8000) != 0
			var dest net.Addr
			if quBit {
				dest = srcAddr
			} else {
				dest = protocol.MulticastGroupIPv6()
			}
			_ = tr.Send(r.ctx, aaaaMsg, dest)
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
		var ipErr error

		// T030: Graceful fallback when interface index unavailable (interfaceIndex=0)
		// This happens when control messages aren't supported or platform doesn't provide IP_PKTINFO
		if interfaceIndex == 0 {
			// Degraded mode: Use default interface IP (legacy behavior)
			// TODO T032: Add debug logging when F-6 (Logging & Observability) is implemented
			ipv4, ipErr = getLocalIPv4()
		} else {
			// RFC 6762 §15 compliance: Use ONLY the IP from the receiving interface
			ipv4, ipErr = getIPv4ForInterface(interfaceIndex)
		}

		if ipErr != nil {
			// T031: If interface-specific IP lookup fails, skip response for this query
			// This is correct behavior per RFC 6762 §15: Better to not respond than
			// to respond with an incorrect (wrong interface) IP address
			// TODO T032: Add error logging when F-6 is implemented
			// Common failure causes: interface went down, no IPv4 configured, invalid index
			continue
		}

		serviceWithIP := &responder.ServiceWithIP{
			InstanceName: matchedService.InstanceName,
			ServiceType:  matchedService.ServiceType,
			Domain:       "local",
			Port:         matchedService.Port,
			IPv4Address:  ipv4,
			TXTRecords:   matchedService.TXT, // internal.Service uses TXT field
			Hostname:     r.hostname,
		}

		// Build response (T076)
		response, err := r.responseBuilder.BuildResponse(serviceWithIP, msg)
		if err != nil {
			continue
		}

		// Per-source-IP rate limiting (FR-026, RFC 6762 §6.2)
		if r.rateLimiter != nil && srcAddr != nil {
			srcIP := srcAddr.String()
			if udpAddr, ok := srcAddr.(*net.UDPAddr); ok {
				srcIP = udpAddr.IP.String()
			}
			if !r.rateLimiter.Allow(srcIP) {
				continue // Rate-limited, skip response
			}
		}

		// RFC 6762 §5.4: QU bit (bit 15 of QCLASS) selects unicast vs multicast destination
		quBit := (question.QCLASS & 0x8000) != 0

		var dest net.Addr
		if quBit {
			dest = srcAddr
		} else if isIPv6 {
			dest = protocol.MulticastGroupIPv6()
		}
		// else nil = multicast to 224.0.0.251:5353

		responsePacket := buildResponsePacket(response)
		_ = tr.Send(r.ctx, responsePacket, dest)
	}

	return nil
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
