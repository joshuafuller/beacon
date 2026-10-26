package responder

import (
	"net"
	"strconv"
	"strings"

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

	// Answer every question with every matching service in one response
	// (RFC 6762 §6; RFC 6763 §4.1: a browse lists all instances of a type).
	var matched []matchedQuestion
	for _, question := range msg.Questions {
		// RFC 6763 §9: Service Type Enumeration
		// A PTR query for "_services._dns-sd._udp.local" returns all unique service types.
		if question.QTYPE == uint16(protocol.RecordTypePTR) && namesEqual(question.QNAME, serviceEnumerationName) {
			r.handleServiceEnumerationQuery(tr, ipv6Network, msg, question, srcAddr)
			continue
		}

		services := r.findMatchingServices(question)
		if len(services) > 0 {
			matched = append(matched, matchedQuestion{question: question, services: services})
		}
	}
	if len(matched) == 0 {
		return nil
	}

	r.respondToQuery(tr, ipv6Network, msg, matched, srcAddr, interfaceIndex)
	return nil
}

// matchedQuestion pairs a query question with the registered services that
// can answer it.
type matchedQuestion struct {
	question message.Question
	services []*responder.Service
}

// namesEqual compares DNS names case-insensitively (RFC 1035 §2.3.3,
// RFC 6762 §16), ignoring a trailing root dot.
func namesEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSuffix(a, "."), strings.TrimSuffix(b, "."))
}

// isLegacyQuery reports whether srcAddr is a legacy (one-shot) querier:
// RFC 6762 §6.7 — a query whose source UDP port is not 5353.
func isLegacyQuery(srcAddr net.Addr) bool {
	udpAddr, ok := srcAddr.(*net.UDPAddr)
	return ok && udpAddr.Port != protocol.Port
}

// serviceEnumerationName is the DNS-SD meta-query name per RFC 6763 §9.
const serviceEnumerationName = "_services._dns-sd._udp.local"

// handleServiceEnumerationQuery responds to a DNS-SD Service Type Enumeration
// query (RFC 6763 §9) with a PTR record for each unique registered service type.
func (r *Responder) handleServiceEnumerationQuery(tr transport.Transport, ipv6Network bool, query *message.DNSMessage, question message.Question, srcAddr net.Addr) {
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

	// RFC 6762 §6.7: a legacy querier gets a conventional unicast reply that
	// repeats its ID and questions, without cache-flush bits, TTL <= 10 s.
	if isLegacyQuery(srcAddr) {
		parsed, err := message.ParseMessage(responseMsg)
		if err != nil {
			return
		}
		makeLegacyUnicastResponse(parsed, query)
		if packet, err := message.SerializeMessage(parsed); err == nil {
			_ = tr.Send(r.ctx, packet, srcAddr)
		}
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

// findMatchingServices returns every registered service that answers
// question's QTYPE/QNAME. Names compare case-insensitively (RFC 6762 §16).
func (r *Responder) findMatchingServices(question message.Question) []*responder.Service {
	var matches []*responder.Service

	for _, instanceName := range r.registry.List() {
		service, found := r.registry.Get(instanceName)
		if !found {
			continue
		}

		switch question.QTYPE {
		case uint16(protocol.RecordTypePTR):
			// PTR: match by service type (e.g., "_http._tcp.local")
			if namesEqual(service.ServiceType, question.QNAME) {
				matches = append(matches, service)
			}
		case uint16(protocol.RecordTypeSRV), uint16(protocol.RecordTypeTXT):
			// SRV/TXT: match by full instance name (e.g., "My Printer._http._tcp.local")
			if namesEqual(service.InstanceName+"."+service.ServiceType, question.QNAME) {
				matches = append(matches, service)
			}
		case uint16(protocol.RecordTypeA), uint16(protocol.RecordTypeAAAA):
			// A/AAAA: the host records are the same for every service, so one
			// match is enough (e.g., "myhost.local").
			if namesEqual(r.hostname, question.QNAME) {
				return []*responder.Service{service}
			}
		}
	}

	return matches
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
//
// All matched questions and services are merged into one response. A legacy
// query (RFC 6762 §6.7) gets a conventional unicast DNS response; otherwise
// the response is multicast unless every matched question has the QU bit.
func (r *Responder) respondToQuery(tr transport.Transport, ipv6Network bool, msg *message.DNSMessage, matched []matchedQuestion, srcAddr net.Addr, interfaceIndex int) {
	ipv4, ipv6Addresses, ipErr := getResponseAddresses(ipv6Network, interfaceIndex)

	if ipErr != nil {
		// T031: If interface-specific IP lookup fails, skip response for this query
		// This is correct behavior per RFC 6762 §6.2: Better to not respond than
		// to respond with an incorrect (wrong interface) IP address
		// TODO T032: Add error logging when F-6 is implemented
		// Common failure causes: interface went down, no IPv4 configured, invalid index
		return
	}

	response := r.responseBuilder.NewResponseSkeleton()
	allQU := true
	for _, m := range matched {
		// RFC 6762 §5.4: QU bit (bit 15 of QCLASS) asks for a unicast response
		if m.question.QCLASS&0x8000 == 0 {
			allQU = false
		}
		for _, svc := range m.services {
			serviceWithIP := &responder.ServiceWithIP{
				InstanceName:  svc.InstanceName,
				ServiceType:   svc.ServiceType,
				Domain:        "local",
				Port:          svc.Port,
				IPv4Address:   ipv4,
				IPv6Addresses: ipv6Addresses,
				TXTRecords:    svc.TXT, // internal.Service uses TXT field
				Hostname:      r.hostname,
			}
			part, err := r.responseBuilder.BuildResponseForQuestion(serviceWithIP, msg, m.question)
			if err != nil {
				continue
			}
			mergeResponse(response, part)
		}
	}

	// Never put an empty response on the wire (all answers known or suppressed).
	if len(response.Answers) == 0 {
		return
	}

	legacy := isLegacyQuery(srcAddr)
	if legacy {
		makeLegacyUnicastResponse(response, msg)
	}
	if err := r.responseBuilder.FinalizeResponse(response); err != nil {
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

	var dest net.Addr
	if legacy || allQU {
		// RFC 6762 §6.7: legacy query → unicast to the source address and port.
		// RFC 6762 §5.4: QU bit set → send unicast response to querier.
		dest = srcAddr
	} else if ipv6Network {
		dest = protocol.MulticastGroupIPv6()
	} else {
		// RFC 6762 §5.4: QU bit clear → send multicast response
		dest = nil // nil = multicast to 224.0.0.251:5353
	}

	// Send response. Legacy unicast is one conventional DNS reply (TC keeps its
	// unicast meaning there). Multicast replies MUST have TC clear (§18.5) and
	// are split so each packet fits the MTU (§17, §7.2).
	if legacy {
		if packet, err := message.SerializeMessage(response); err == nil {
			_ = tr.Send(r.ctx, packet, dest)
		}
		return
	}
	response.Header.Flags &^= 0x0200
	for _, part := range splitResponse(response, maxMulticastPayload) {
		if packet, err := message.SerializeMessage(part); err == nil {
			_ = tr.Send(r.ctx, packet, dest)
		}
	}
}

// maxMulticastPayload keeps an mDNS packet inside a 1500-byte Ethernet MTU
// with IPv6 (40) and UDP (8) headers and some margin. RFC 6762 §17: a packet
// larger than the MTU "MUST NOT contain more than one resource record".
const maxMulticastPayload = 1440

// recordWireSize is the uncompressed wire size of a resource record.
func recordWireSize(a message.Answer) int {
	name, err := message.EncodeDomainName(a.NAME)
	if err != nil {
		return 1 + len(a.NAME) + 1 + 10 + len(a.RDATA) // upper bound
	}
	return len(name) + 10 + len(a.RDATA)
}

// splitResponse splits msg into packets whose payload is at most limit bytes.
// Answers are kept whole and in order; additionals fill the remaining space
// and are dropped when they do not fit (they are optional, RFC 6762 §6). A
// single answer larger than limit is sent alone.
func splitResponse(msg *message.DNSMessage, limit int) []*message.DNSMessage {
	const headerSize = 12
	newPart := func() *message.DNSMessage {
		return &message.DNSMessage{Header: message.DNSHeader{ID: msg.Header.ID, Flags: msg.Header.Flags}}
	}
	parts := []*message.DNSMessage{newPart()}
	size := headerSize
	for _, a := range msg.Answers {
		n := recordWireSize(a)
		cur := parts[len(parts)-1]
		if len(cur.Answers) > 0 && size+n > limit {
			cur = newPart()
			parts = append(parts, cur)
			size = headerSize
		}
		cur.Answers = append(cur.Answers, a)
		size += n
	}
	cur := parts[len(parts)-1]
	for _, a := range msg.Additionals {
		if n := recordWireSize(a); size+n <= limit {
			cur.Additionals = append(cur.Additionals, a)
			size += n
		}
	}
	for _, p := range parts {
		// #nosec G115 -- each part holds a subset of msg's uint16-bounded sections.
		p.Header.ANCount, p.Header.ARCount = uint16(len(p.Answers)), uint16(len(p.Additionals)) //nolint:gosec
	}
	return parts
}

// mergeResponse appends part's records to response, skipping records already
// present. A record that is an answer is not repeated as an additional.
func mergeResponse(response, part *message.DNSMessage) {
	key := func(a message.Answer) string {
		return strings.ToLower(a.NAME) + "|" + strconv.Itoa(int(a.TYPE)) + "|" + string(a.RDATA)
	}
	seen := make(map[string]bool, len(response.Answers)+len(response.Additionals))
	for _, a := range response.Answers {
		seen[key(a)] = true
	}
	for _, a := range part.Answers {
		if !seen[key(a)] {
			seen[key(a)] = true
			response.Answers = append(response.Answers, a)
		}
	}
	for _, a := range response.Additionals {
		seen[key(a)] = true
	}
	for _, a := range part.Additionals {
		if !seen[key(a)] {
			seen[key(a)] = true
			response.Additionals = append(response.Additionals, a)
		}
	}
	// A later question may have promoted an earlier additional to an answer.
	inAnswers := make(map[string]bool, len(response.Answers))
	for _, a := range response.Answers {
		inAnswers[key(a)] = true
	}
	kept := response.Additionals[:0]
	for _, a := range response.Additionals {
		if !inAnswers[key(a)] {
			kept = append(kept, a)
		}
	}
	response.Additionals = kept
}

// legacyUnicastMaxTTL is the TTL cap for legacy unicast responses
// (RFC 6762 §6.7: "SHOULD NOT be greater than ten seconds").
const legacyUnicastMaxTTL = 10

// makeLegacyUnicastResponse turns response into a conventional unicast DNS
// response to query per RFC 6762 §6.7: it repeats the query ID and questions,
// clears the cache-flush bit, and caps TTLs at ten seconds.
func makeLegacyUnicastResponse(response, query *message.DNSMessage) {
	response.Header.ID = query.Header.ID // RFC 6762 §18.1
	response.Questions = append([]message.Question(nil), query.Questions...)
	// #nosec G115 -- the query was parsed from a uint16 QDCOUNT.
	response.Header.QDCount = uint16(len(response.Questions)) //nolint:gosec
	for _, section := range [][]message.Answer{response.Answers, response.Additionals} {
		for i := range section {
			section[i].CLASS &^= 0x8000 // cache-flush bit MUST NOT be set
			if section[i].TTL > legacyUnicastMaxTTL {
				section[i].TTL = legacyUnicastMaxTTL
			}
		}
	}
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
