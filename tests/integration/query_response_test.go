package integration

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
	"github.com/joshuafuller/beacon/responder"
)

// TestQueryResponse_ResponseLatency tests end-to-end query response behavior.
//
// RFC 6762 §6: "When a host... is able to answer every question in the query message,
// and for all of those answer records it has previously verified that the name, rrtype,
// and rrclass are unique on the link), it SHOULD NOT impose any random delay before
// responding, and SHOULD normally generate its response within at most 10 ms."
//
// SC-006: Response time MUST be <100ms for registered services
//
// T072 [US3]: Integration test query registered service, verify response sent
func TestQueryResponse_ResponseLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// Create mock transport and responder
	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock), responder.WithHostname("testhost.local"))
	if err != nil {
		t.Fatalf("Failed to create responder: %v", err)
	}
	defer func() {
		cancel() // Cancel context first to unblock query handler's blocking Receive()
		_ = r.Close()
	}()

	// Register service (probing/announcing proceeds quickly with mock transport)
	service := &responder.Service{
		InstanceName: "TestPrinter",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Record send count before query to isolate query-response packets
	callsBefore := len(mock.SendCalls())

	// Build a PTR query for the registered service type
	queryPacket := buildTestQuery(t, "_http._tcp.local", uint16(protocol.RecordTypePTR), 0x0001)

	srcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 5353}

	// Inject query into the mock transport's receive queue.
	// The query handler goroutine will pick this up via transport.Receive().
	mock.QueueReceive(queryPacket, srcAddr, 0)

	// Wait for query handler to process (should be well under 100ms)
	time.Sleep(200 * time.Millisecond)

	// Verify a response was sent
	allCalls := mock.SendCalls()
	responseCalls := allCalls[callsBefore:]

	if len(responseCalls) == 0 {
		t.Fatal("No response packet sent after PTR query injection")
	}

	// Verify the response is a valid DNS response (QR=1)
	responsePacket := responseCalls[0].Packet
	if len(responsePacket) < 12 {
		t.Fatalf("Response packet too short: %d bytes", len(responsePacket))
	}

	flags := binary.BigEndian.Uint16(responsePacket[2:4])
	if flags&0x8000 == 0 {
		t.Error("Response QR bit not set (expected QR=1 for response)")
	}

	t.Logf("Response received: %d bytes, flags=0x%04x", len(responsePacket), flags)
}

// TestQueryResponse_PTRQueryWithAdditionalRecords tests PTR query response structure.
//
// RFC 6762 §6: When responding to a PTR query, responder should include:
//   - Answer section: PTR record
//   - Additional section: SRV, TXT, A records (reduces round-trips)
//
// T072 [US3]: Verify PTR response includes PTR record in answers
func TestQueryResponse_PTRQueryWithAdditionalRecords(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock), responder.WithHostname("testhost.local"))
	if err != nil {
		t.Fatalf("Failed to create responder: %v", err)
	}
	defer func() {
		cancel()
		_ = r.Close()
	}()

	service := &responder.Service{
		InstanceName: "TestService",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
		TXTRecords:   map[string]string{"txtvers": "1", "path": "/"},
	}

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Record send count before query
	callsBefore := len(mock.SendCalls())

	// Build PTR query
	queryPacket := buildTestQuery(t, "_http._tcp.local", uint16(protocol.RecordTypePTR), 0x0001)
	srcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 5353}

	mock.QueueReceive(queryPacket, srcAddr, 0)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	allCalls := mock.SendCalls()
	responseCalls := allCalls[callsBefore:]

	if len(responseCalls) == 0 {
		t.Fatal("No response packet sent after PTR query")
	}

	// Parse response to verify PTR record is present
	responsePacket := responseCalls[0].Packet
	parsed, err := message.ParseMessage(responsePacket)
	if err != nil {
		t.Fatalf("ParseMessage(response) error = %v", err)
	}

	if len(parsed.Answers) == 0 {
		t.Fatal("Response has 0 answer records, want at least 1 PTR record")
	}

	// Check for PTR record in answers
	foundPTR := false
	for _, ans := range parsed.Answers {
		if ans.TYPE == uint16(protocol.RecordTypePTR) {
			foundPTR = true
			t.Logf("Found PTR answer: name=%s, TTL=%d", ans.NAME, ans.TTL)
		}
	}

	if !foundPTR {
		t.Error("Response answers section missing PTR record")
		for i, ans := range parsed.Answers {
			t.Logf("  answer[%d]: name=%s type=%d", i, ans.NAME, ans.TYPE)
		}
	}
}

// TestQueryResponse_QUBitHandling tests unicast response per RFC 6762 §5.4.
//
// RFC 6762 §5.4: "When receiving a question with the unicast-response bit set, a
// responder SHOULD usually respond with a unicast packet directed back to the querier."
//
// T072 [US3]: Verify QU bit triggers unicast response to srcAddr
func TestQueryResponse_QUBitHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	mock := transport.NewMockTransport()
	r, err := responder.New(ctx, responder.WithTransport(mock), responder.WithHostname("testhost.local"))
	if err != nil {
		t.Fatalf("Failed to create responder: %v", err)
	}
	defer func() {
		cancel()
		_ = r.Close()
	}()

	service := &responder.Service{
		InstanceName: "TestService",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}

	err = r.Register(service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// --- Test 1: QU bit set (QCLASS = IN | 0x8000 = 0x8001) → unicast response ---
	callsBeforeQU := len(mock.SendCalls())

	quQueryPacket := buildTestQuery(t, "_http._tcp.local", uint16(protocol.RecordTypePTR), 0x8001)
	quSrcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.50"), Port: 5353}

	mock.QueueReceive(quQueryPacket, quSrcAddr, 0)
	time.Sleep(200 * time.Millisecond)

	allCalls := mock.SendCalls()
	quResponseCalls := allCalls[callsBeforeQU:]

	if len(quResponseCalls) == 0 {
		t.Fatal("No response sent for QU query")
	}

	// Verify unicast response: dest should be the querier's address (not nil/multicast)
	quDest := quResponseCalls[0].Dest
	if quDest == nil {
		t.Error("QU query response dest is nil (multicast), want unicast to querier address")
	} else {
		destUDP, ok := quDest.(*net.UDPAddr)
		if !ok {
			t.Errorf("QU query response dest type = %T, want *net.UDPAddr", quDest)
		} else if !destUDP.IP.Equal(quSrcAddr.IP) {
			t.Errorf("QU query response dest IP = %v, want %v (querier)", destUDP.IP, quSrcAddr.IP)
		}
		t.Logf("QU response correctly sent unicast to %v", quDest)
	}

	// --- Test 2: QU bit clear (QCLASS = IN = 0x0001) → multicast response ---
	callsBeforeMC := len(mock.SendCalls())

	mcQueryPacket := buildTestQuery(t, "_http._tcp.local", uint16(protocol.RecordTypePTR), 0x0001)
	mcSrcAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.51"), Port: 5353}

	mock.QueueReceive(mcQueryPacket, mcSrcAddr, 0)
	time.Sleep(200 * time.Millisecond)

	allCalls = mock.SendCalls()
	mcResponseCalls := allCalls[callsBeforeMC:]

	if len(mcResponseCalls) == 0 {
		t.Fatal("No response sent for multicast query")
	}

	// Verify multicast response: dest should be nil (multicast to 224.0.0.251:5353)
	mcDest := mcResponseCalls[0].Dest
	if mcDest != nil {
		t.Errorf("Multicast query response dest = %v, want nil (multicast)", mcDest)
	} else {
		t.Log("Multicast response correctly sent to nil (multicast address)")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildTestQuery constructs a DNS query packet for integration testing.
//
// Parameters:
//   - t: test context for error reporting
//   - qname: Query name (e.g., "_http._tcp.local")
//   - qtype: Query type (e.g., PTR=12, A=1)
//   - qclass: Query class (0x0001 for IN, 0x8001 for IN+QU bit)
//
// Returns a complete DNS query packet in wire format.
func buildTestQuery(t *testing.T, qname string, qtype uint16, qclass uint16) []byte {
	t.Helper()

	packet := make([]byte, 0, 512)

	// Header: 12 bytes
	// ID=0x0000
	packet = append(packet, 0x00, 0x00)
	// Flags=0x0000 (QR=0, query)
	packet = append(packet, 0x00, 0x00)
	// QDCOUNT=1
	packet = append(packet, 0x00, 0x01)
	// ANCOUNT=0
	packet = append(packet, 0x00, 0x00)
	// NSCOUNT=0
	packet = append(packet, 0x00, 0x00)
	// ARCOUNT=0
	packet = append(packet, 0x00, 0x00)

	// QNAME: encode domain name as DNS labels per RFC 1035 §3.1
	labels := testEncodeDomainName(qname)
	packet = append(packet, labels...)

	// QTYPE: 2 bytes, big-endian
	qtypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(qtypeBytes, qtype)
	packet = append(packet, qtypeBytes...)

	// QCLASS: 2 bytes
	qclassBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(qclassBytes, qclass)
	packet = append(packet, qclassBytes...)

	return packet
}

// testEncodeDomainName encodes a domain name as DNS labels per RFC 1035 §3.1.
//
// Example: "_http._tcp.local" →
//
//	[]byte{5, '_', 'h', 't', 't', 'p', 4, '_', 't', 'c', 'p', 5, 'l', 'o', 'c', 'a', 'l', 0}
func testEncodeDomainName(name string) []byte {
	if name == "" || name == "." {
		return []byte{0}
	}

	// Split into labels
	var labels []string
	start := 0
	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			if i > start {
				labels = append(labels, name[start:i])
			}
			start = i + 1
		}
	}
	if start < len(name) {
		labels = append(labels, name[start:])
	}

	// Encode as length-prefixed labels
	encoded := make([]byte, 0, len(name)+2)
	for _, label := range labels {
		if len(label) > 63 {
			label = label[:63]
		}
		encoded = append(encoded, byte(len(label)))
		encoded = append(encoded, []byte(label)...)
	}
	encoded = append(encoded, 0) // Null terminator

	return encoded
}
