package responder

import (
	"net"
	"testing"

	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
)

func ipv6ServiceForTest() *ServiceWithIP {
	return &ServiceWithIP{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Domain:       "local",
		Port:         8080,
		IPv6Addresses: [][]byte{
			[]byte(net.ParseIP("fe80::1234").To16()),
			[]byte(net.ParseIP("2001:db8::1234").To16()),
		},
		Hostname: "example.local",
	}
}

func TestBuildResponseAAAAQueryReturnsAAAAAnswer(t *testing.T) {
	query := &message.DNSMessage{Questions: []message.Question{{
		QNAME: "example.local", QTYPE: uint16(protocol.RecordTypeAAAA), QCLASS: uint16(protocol.ClassIN),
	}}}
	response, err := NewResponseBuilder().BuildResponse(ipv6ServiceForTest(), query)
	if err != nil {
		t.Fatalf("BuildResponse() failed: %v", err)
	}
	if len(response.Answers) != 2 {
		t.Fatalf("AAAA answer count = %d, want both interface addresses", len(response.Answers))
	}
	for _, answer := range response.Answers {
		if answer.TYPE != uint16(protocol.RecordTypeAAAA) {
			t.Fatalf("answer type = %d, want AAAA", answer.TYPE)
		}
	}
}

func TestBuildResponsePTRIncludesAAAAAdditional(t *testing.T) {
	query := &message.DNSMessage{Questions: []message.Question{{
		QNAME: "_http._tcp.local", QTYPE: uint16(protocol.RecordTypePTR), QCLASS: uint16(protocol.ClassIN),
	}}}
	response, err := NewResponseBuilder().BuildResponse(ipv6ServiceForTest(), query)
	if err != nil {
		t.Fatalf("BuildResponse() failed: %v", err)
	}
	for _, additional := range response.Additionals {
		if additional.TYPE == uint16(protocol.RecordTypeAAAA) {
			return
		}
	}
	t.Fatal("PTR response omitted its AAAA additional record")
}
