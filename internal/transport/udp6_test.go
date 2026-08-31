package transport

import (
	"context"
	"net"
	"reflect"
	"testing"

	"golang.org/x/net/ipv6"

	"github.com/joshuafuller/beacon/internal/protocol"
)

func TestUDPv6TransportSendMulticastUsesEveryJoinedInterface(t *testing.T) {
	var got []int
	tr := &UDPv6Transport{
		joinedInterfaces: []int{2, 7},
		writeTo: func(packet []byte, cm *ipv6.ControlMessage, dest net.Addr) (int, error) {
			got = append(got, cm.IfIndex)
			return len(packet), nil
		},
	}

	if err := tr.Send(context.Background(), []byte{1, 2, 3}, protocol.MulticastGroupIPv6()); err != nil {
		t.Fatalf("Send() failed: %v", err)
	}
	if want := []int{2, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("multicast interface indexes = %v, want %v", got, want)
	}
}

func TestUDPv6TransportSendMulticastFailsWithoutJoinedInterface(t *testing.T) {
	tr := &UDPv6Transport{}
	if err := tr.Send(context.Background(), []byte{1}, protocol.MulticastGroupIPv6()); err == nil {
		t.Fatal("Send() succeeded without a joined IPv6 multicast interface")
	}
}
