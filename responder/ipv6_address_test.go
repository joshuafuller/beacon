package responder

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	internalresponder "github.com/joshuafuller/beacon/internal/responder"
)

func TestIPv6ResponseAddressesRejectUnknownInterface(t *testing.T) {
	if addresses, err := getIPv6ResponseAddresses(0); err == nil {
		t.Fatalf("getIPv6ResponseAddresses(0) = %v, nil; want an error rather than cross-interface addresses", addresses)
	}
}

func TestIPBytesKeepsOnlyIPv6AndCopiesInput(t *testing.T) {
	ipv6Address := net.ParseIP("2001:db8::1")
	got := ipBytes([]net.IP{net.ParseIP("192.0.2.1"), nil, ipv6Address})
	if len(got) != 1 || !reflect.DeepEqual(got[0], []byte(ipv6Address.To16())) {
		t.Fatalf("ipBytes() = %v, want one normalized IPv6 address", got)
	}
	ipv6Address[0] ^= 0xff
	if reflect.DeepEqual(got[0], []byte(ipv6Address.To16())) {
		t.Fatal("ipBytes() result aliases its input")
	}
}

func TestIPv6ResponseAddressesRejectNegativeInterface(t *testing.T) {
	if addresses, err := getIPv6ResponseAddresses(-1); err == nil {
		t.Fatalf("getIPv6ResponseAddresses(-1) = %v, nil; want an error", addresses)
	}
}

// RFC 6762 §6.2: A response containing an A or AAAA record SHOULD include
// records of the other address family for the same name when they are available.
func TestIPv6PTRResponseIncludesBothAddressFamilies(t *testing.T) {
	interfaceIndex := dualStackInterfaceIndex(t)
	var sentPacket []byte
	tr := &MockTransport{sendFunc: func(_ context.Context, packet []byte, _ net.Addr) error {
		sentPacket = append([]byte(nil), packet...)
		return nil
	}}
	r := &Responder{
		ctx:             context.Background(),
		hostname:        "host.local",
		responseBuilder: internalresponder.NewResponseBuilder(),
	}
	service := &internalresponder.Service{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}
	question := message.Question{
		QNAME:  service.ServiceType,
		QTYPE:  uint16(protocol.RecordTypePTR),
		QCLASS: uint16(protocol.ClassIN),
	}

	r.respondToQuery(tr, true, service, &message.DNSMessage{Questions: []message.Question{question}}, question, nil, interfaceIndex)
	if sentPacket == nil {
		t.Fatal("respondToQuery() did not send an IPv6 response")
	}
	response, err := parseMessage(sentPacket)
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	var hasA, hasAAAA bool
	for _, record := range response.Additionals {
		hasA = hasA || record.TYPE == uint16(protocol.RecordTypeA)
		hasAAAA = hasAAAA || record.TYPE == uint16(protocol.RecordTypeAAAA)
	}
	if !hasA || !hasAAAA {
		t.Fatalf("response address additionals: A=%v AAAA=%v, want both per RFC 6762 §6.2", hasA, hasAAAA)
	}
}

// RFC 6762 §11: An IPv6 unicast source is on-link when it matches an on-link
// prefix configured on the receiving interface; it need not be link-local.
func TestValidateIPv6SourceAcceptsAddressOnInterfacePrefix(t *testing.T) {
	loopback, err := net.InterfaceByName("lo")
	if err != nil {
		for _, candidate := range mustInterfaces(t) {
			if candidate.Flags&net.FlagLoopback != 0 {
				loopback = &candidate
				break
			}
		}
	}
	if loopback == nil {
		t.Skip("host has no loopback interface")
	}
	if !validateSourceAddressOnNetwork(&net.UDPAddr{IP: net.ParseIP("::1"), Port: protocol.Port}, loopback.Index, true) {
		t.Fatal("validateSourceAddressOnNetwork() rejected ::1 on its configured interface prefix")
	}
}

func dualStackInterfaceIndex(t *testing.T) int {
	t.Helper()
	for _, iface := range mustInterfaces(t) {
		addresses, addrErr := iface.Addrs()
		if addrErr != nil {
			continue
		}
		var hasIPv4, hasIPv6 bool
		for _, address := range addresses {
			ipNet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.To4() != nil {
				hasIPv4 = true
			} else if ipNet.IP.To16() != nil {
				hasIPv6 = true
			}
		}
		if hasIPv4 && hasIPv6 {
			return iface.Index
		}
	}
	t.Skip("host has no dual-stack interface")
	return 0
}

func mustInterfaces(t *testing.T) []net.Interface {
	t.Helper()
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("list interfaces: %v", err)
	}
	return interfaces
}
