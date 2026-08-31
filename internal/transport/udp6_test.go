package transport

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/ipv6"

	"github.com/joshuafuller/beacon/internal/protocol"
)

type ipv6PacketConnStub struct {
	deadline time.Time
}

func (s *ipv6PacketConnStub) ReadFrom([]byte) (int, net.Addr, error) {
	return 0, nil, errors.New("unused")
}
func (s *ipv6PacketConnStub) WriteTo(p []byte, _ net.Addr) (int, error) { return len(p), nil }
func (s *ipv6PacketConnStub) Close() error                              { return nil }
func (s *ipv6PacketConnStub) LocalAddr() net.Addr                       { return &net.UDPAddr{} }
func (s *ipv6PacketConnStub) SetDeadline(deadline time.Time) error      { s.deadline = deadline; return nil }
func (s *ipv6PacketConnStub) SetReadDeadline(deadline time.Time) error {
	s.deadline = deadline
	return nil
}
func (s *ipv6PacketConnStub) SetWriteDeadline(time.Time) error { return nil }

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

func TestUDPv6TransportReceiveReturnsPacketSourceAndInterface(t *testing.T) {
	src := &net.UDPAddr{IP: net.ParseIP("fe80::2"), Port: protocol.Port, Zone: "eth0"}
	conn := &ipv6PacketConnStub{}
	tr := &UDPv6Transport{
		conn: conn,
		readFrom: func(buffer []byte) (int, *ipv6.ControlMessage, net.Addr, error) {
			copy(buffer, []byte{1, 2, 3})
			return 3, &ipv6.ControlMessage{IfIndex: 7}, src, nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	packet, gotSrc, ifIndex, err := tr.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive() failed: %v", err)
	}
	if !reflect.DeepEqual(packet, []byte{1, 2, 3}) || gotSrc != src || ifIndex != 7 {
		t.Fatalf("Receive() = (%v, %v, %d), want ([1 2 3], %v, 7)", packet, gotSrc, ifIndex, src)
	}
	if conn.deadline.IsZero() {
		t.Fatal("Receive() did not propagate the context deadline")
	}
}

func TestUDPv6TransportReceiveHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tr := &UDPv6Transport{}
	if _, _, _, err := tr.Receive(ctx); err == nil {
		t.Fatal("Receive() succeeded with a canceled context")
	}
}
