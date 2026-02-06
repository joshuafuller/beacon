package contract

import (
	"context"
	"encoding/binary"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
	"github.com/joshuafuller/beacon/responder"
)

// serviceEnumerationName is the DNS-SD meta-query name per RFC 6763 §9.
const serviceEnumerationName = "_services._dns-sd._udp.local"

// buildPTRQuery builds a DNS query packet for a PTR record of the given name.
func buildPTRQuery(name string) ([]byte, error) {
	encodedName, err := message.EncodeName(name)
	if err != nil {
		return nil, err
	}

	// DNS header (12 bytes): QR=0, QDCOUNT=1
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[4:6], 1) // QDCOUNT = 1

	// Question section: QNAME + QTYPE(PTR=12) + QCLASS(IN=1)
	question := make([]byte, len(encodedName)+4)
	copy(question, encodedName)
	binary.BigEndian.PutUint16(question[len(encodedName):], uint16(protocol.RecordTypePTR))
	binary.BigEndian.PutUint16(question[len(encodedName)+2:], uint16(protocol.ClassIN))

	return append(header, question...), nil
}

// extractPTRTargetsFromPacket parses a DNS response packet and returns the
// PTR record target names from the answer section.
func extractPTRTargetsFromPacket(packet []byte) ([]string, error) {
	msg, err := message.ParseMessage(packet)
	if err != nil {
		return nil, err
	}

	var targets []string
	for _, answer := range msg.Answers {
		if answer.TYPE == uint16(protocol.RecordTypePTR) {
			// Parse the encoded DNS name from RDATA
			name, _, parseErr := message.ParseName(answer.RDATA, 0)
			if parseErr != nil {
				continue
			}
			targets = append(targets, name)
		}
	}
	return targets, nil
}

// TestRFC6763_ServiceEnumeration_MetaQuery tests RFC 6763 §9 service type enumeration.
//
// RFC 6763 §9: "A DNS query for PTR records with the name '_services._dns-sd._udp.<Domain>'
// yields a set of PTR records, where the rdata of each PTR record is the two-label <Service>
// name, plus the same domain, e.g., '_http._tcp.<Domain>'."
//
// FR-027: System MUST respond to "_services._dns-sd._udp.local" PTR queries with a list
// of all registered service types
//
// T103: Contract test for service enumeration
func TestRFC6763_ServiceEnumeration_MetaQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() {
		cancel()              // Cancel context first to unblock Receive()
		_ = r.Close()
	}()

	// Register 3 services with DIFFERENT service types (bypass probing)
	services := []*responder.Service{
		{
			InstanceName: "Web Server",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		},
		{
			InstanceName: "SSH Server",
			ServiceType:  "_ssh._tcp.local",
			Port:         22,
		},
		{
			InstanceName: "FTP Server",
			ServiceType:  "_ftp._tcp.local",
			Port:         21,
		},
	}

	for _, svc := range services {
		if err := r.RegisterServiceWithoutProbing(svc); err != nil {
			t.Fatalf("RegisterServiceWithoutProbing(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Build and queue a PTR query for the DNS-SD meta-query name
	queryPacket, err := buildPTRQuery(serviceEnumerationName)
	if err != nil {
		t.Fatalf("buildPTRQuery() error = %v", err)
	}

	srcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 5353}
	mock.QueueReceive(queryPacket, srcAddr, 0)

	// Wait for the query handler goroutine to process the packet and send a response
	var sendCalls []transport.SendCall
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sendCalls = mock.SendCalls()
		if len(sendCalls) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(sendCalls) == 0 {
		t.Fatal("expected at least 1 send call for meta-query response, got 0")
	}

	// Extract PTR targets from the last sent packet (the meta-query response)
	responsePacket := sendCalls[len(sendCalls)-1].Packet
	targets, err := extractPTRTargetsFromPacket(responsePacket)
	if err != nil {
		t.Fatalf("failed to parse response packet: %v", err)
	}

	// RFC 6763 §9: Should return exactly 3 PTR records (one per unique service type)
	if len(targets) != 3 {
		t.Fatalf("expected 3 PTR targets, got %d: %v", len(targets), targets)
	}

	// Sort for deterministic comparison
	sort.Strings(targets)
	expected := []string{"_ftp._tcp.local", "_http._tcp.local", "_ssh._tcp.local"}
	sort.Strings(expected)

	for i, want := range expected {
		if targets[i] != want {
			t.Errorf("PTR target[%d] = %q, want %q", i, targets[i], want)
		}
	}
}

// TestRFC6763_ServiceEnumeration_DuplicateTypes tests that duplicate service types
// only appear once in enumeration response.
//
// RFC 6763 §9: Service type enumeration lists unique service types, not instances.
// If 3 services all use "_http._tcp.local", enumeration should list "_http._tcp.local" once.
//
// T103: Edge case - duplicate service types
func TestRFC6763_ServiceEnumeration_DuplicateTypes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() {
		cancel()
		_ = r.Close()
	}()

	// Register 3 services with SAME service type (bypass probing)
	services := []*responder.Service{
		{
			InstanceName: "Web Server 1",
			ServiceType:  "_http._tcp.local",
			Port:         8080,
		},
		{
			InstanceName: "Web Server 2",
			ServiceType:  "_http._tcp.local",
			Port:         8081,
		},
		{
			InstanceName: "Web Server 3",
			ServiceType:  "_http._tcp.local",
			Port:         8082,
		},
	}

	for _, svc := range services {
		if err := r.RegisterServiceWithoutProbing(svc); err != nil {
			t.Fatalf("RegisterServiceWithoutProbing(%q) error = %v", svc.InstanceName, err)
		}
	}

	// Build and queue a PTR query for the DNS-SD meta-query name
	queryPacket, err := buildPTRQuery(serviceEnumerationName)
	if err != nil {
		t.Fatalf("buildPTRQuery() error = %v", err)
	}

	srcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 5353}
	mock.QueueReceive(queryPacket, srcAddr, 0)

	// Wait for the query handler goroutine to process the packet and send a response
	var sendCalls []transport.SendCall
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sendCalls = mock.SendCalls()
		if len(sendCalls) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(sendCalls) == 0 {
		t.Fatal("expected at least 1 send call for meta-query response, got 0")
	}

	// Extract PTR targets from the last sent packet
	responsePacket := sendCalls[len(sendCalls)-1].Packet
	targets, err := extractPTRTargetsFromPacket(responsePacket)
	if err != nil {
		t.Fatalf("failed to parse response packet: %v", err)
	}

	// RFC 6763 §9: Should return exactly 1 PTR record for "_http._tcp.local"
	// NOT 3 PTR records (one per instance)
	if len(targets) != 1 {
		t.Fatalf("expected 1 unique PTR target for duplicate service types, got %d: %v", len(targets), targets)
	}

	if targets[0] != "_http._tcp.local" {
		t.Errorf("PTR target = %q, want %q", targets[0], "_http._tcp.local")
	}
}

// TestRFC6763_ServiceEnumeration_EmptyRegistry tests enumeration when no services registered.
//
// RFC 6763 §9: If no services are registered, enumeration query should return empty response
// (no PTR records sent).
//
// T103: Edge case - empty registry
func TestRFC6763_ServiceEnumeration_EmptyRegistry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock))
	if err != nil {
		t.Fatalf("responder.New() error = %v, want nil", err)
	}
	defer func() {
		cancel()
		_ = r.Close()
	}()

	// No services registered - empty registry

	// Build and queue a PTR query for the DNS-SD meta-query name
	queryPacket, err := buildPTRQuery(serviceEnumerationName)
	if err != nil {
		t.Fatalf("buildPTRQuery() error = %v", err)
	}

	srcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 5353}
	mock.QueueReceive(queryPacket, srcAddr, 0)

	// Wait briefly for the query handler to process (should NOT send a response)
	time.Sleep(200 * time.Millisecond)

	sendCalls := mock.SendCalls()
	if len(sendCalls) != 0 {
		t.Errorf("expected 0 send calls for empty registry meta-query, got %d", len(sendCalls))
	}
}
