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
	deadline        time.Time
	writeN          int
	writeErr        error
	closeErr        error
	readDeadlineErr error
}

func (s *ipv6PacketConnStub) ReadFrom([]byte) (int, net.Addr, error) {
	return 0, nil, errors.New("unused")
}
func (s *ipv6PacketConnStub) WriteTo(p []byte, _ net.Addr) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	if s.writeN != 0 {
		return s.writeN, nil
	}
	return len(p), nil
}
func (s *ipv6PacketConnStub) Close() error                         { return s.closeErr }
func (s *ipv6PacketConnStub) LocalAddr() net.Addr                  { return &net.UDPAddr{} }
func (s *ipv6PacketConnStub) SetDeadline(deadline time.Time) error { s.deadline = deadline; return nil }
func (s *ipv6PacketConnStub) SetReadDeadline(deadline time.Time) error {
	s.deadline = deadline
	return s.readDeadlineErr
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

func TestUDPv6TransportJoinedInterfaceIndexesReturnsCopy(t *testing.T) {
	tr := &UDPv6Transport{joinedInterfaces: []int{2, 7}}

	got := tr.JoinedInterfaceIndexes()
	if want := []int{2, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("JoinedInterfaceIndexes() = %v, want %v", got, want)
	}

	got[0] = 99
	if want := []int{2, 7}; !reflect.DeepEqual(tr.joinedInterfaces, want) {
		t.Fatalf("mutating returned indexes changed transport state to %v", tr.joinedInterfaces)
	}
}

// RFC 6762 §20: A host participates in the IPv6 .local zone only through
// interfaces on which it has IPv6 connectivity.
func TestHasUsableIPv6Address(t *testing.T) {
	tests := []struct {
		name      string
		addresses []net.Addr
		want      bool
	}{
		{name: "none"},
		{name: "IPv4 only", addresses: []net.Addr{&net.IPNet{IP: net.ParseIP("192.0.2.1")}}},
		{name: "IPv6 link-local", addresses: []net.Addr{&net.IPNet{IP: net.ParseIP("fe80::1")}}, want: true},
		{name: "IPv6 routable", addresses: []net.Addr{&net.IPNet{IP: net.ParseIP("2001:db8::1")}}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasUsableIPv6Address(test.addresses); got != test.want {
				t.Fatalf("hasUsableIPv6Address() = %v, want %v", got, test.want)
			}
		})
	}
}

// RFC 6762 §§20, 22: Construction either joins the IPv6 mDNS group on at
// least one IPv6-capable interface or reports that this host cannot participate.
func TestNewUDPv6TransportMatchesHostCapability(t *testing.T) {
	tr, err := NewUDPv6Transport()
	if err != nil {
		for _, iface := range mustNetworkInterfaces(t) {
			addresses, addrErr := iface.Addrs()
			if addrErr == nil && iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagMulticast != 0 && hasUsableIPv6Address(addresses) {
				t.Fatalf("NewUDPv6Transport() failed with an IPv6-capable multicast interface %s: %v", iface.Name, err)
			}
		}
		return
	}
	if len(tr.joinedInterfaces) == 0 {
		t.Fatal("NewUDPv6Transport() succeeded without a joined IPv6 interface")
	}
	if closeErr := tr.Close(); closeErr != nil {
		t.Fatalf("Close() failed: %v", closeErr)
	}
}

func mustNetworkInterfaces(t *testing.T) []net.Interface {
	t.Helper()
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("net.Interfaces(): %v", err)
	}
	return interfaces
}

func TestUDPv6TransportSendMulticastFailsWithoutJoinedInterface(t *testing.T) {
	tr := &UDPv6Transport{}
	if err := tr.Send(context.Background(), []byte{1}, protocol.MulticastGroupIPv6()); err == nil {
		t.Fatal("Send() succeeded without a joined IPv6 multicast interface")
	}
}

func TestUDPv6TransportSendMulticastReportsPartialInterfaceFailure(t *testing.T) {
	var attempted []int
	tr := &UDPv6Transport{
		joinedInterfaces: []int{2, 7},
		writeTo: func(packet []byte, cm *ipv6.ControlMessage, _ net.Addr) (int, error) {
			attempted = append(attempted, cm.IfIndex)
			if cm.IfIndex == 7 {
				return 0, errors.New("interface unavailable")
			}
			return len(packet), nil
		},
	}

	err := tr.Send(context.Background(), []byte{1, 2, 3}, protocol.MulticastGroupIPv6())
	if err == nil {
		t.Fatal("Send() succeeded despite failing on one joined IPv6 interface")
	}
	if want := []int{2, 7}; !reflect.DeepEqual(attempted, want) {
		t.Fatalf("attempted interface indexes = %v, want %v", attempted, want)
	}
}

func TestUDPv6TransportSendOnInterfaceRejectsUnjoinedInterface(t *testing.T) {
	tr := &UDPv6Transport{
		joinedInterfaces: []int{2},
		writeTo: func(packet []byte, _ *ipv6.ControlMessage, _ net.Addr) (int, error) {
			return len(packet), nil
		},
	}
	if err := tr.SendOnInterface(context.Background(), []byte{1}, protocol.MulticastGroupIPv6(), 7); err == nil {
		t.Fatal("SendOnInterface() accepted an interface that was not joined")
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

func TestUDPv6TransportSendUnicastAndErrorPaths(t *testing.T) {
	destination := &net.UDPAddr{IP: net.ParseIP("2001:db8::2"), Port: protocol.Port}
	tests := []struct {
		name     string
		conn     *ipv6PacketConnStub
		canceled bool
		wantErr  bool
	}{
		{name: "success", conn: &ipv6PacketConnStub{}},
		{name: "write error", conn: &ipv6PacketConnStub{writeErr: errors.New("write failed")}, wantErr: true},
		{name: "partial write", conn: &ipv6PacketConnStub{writeN: 1}, wantErr: true},
		{name: "canceled context", conn: &ipv6PacketConnStub{}, canceled: true, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if test.canceled {
				cancel()
			}
			tr := &UDPv6Transport{conn: test.conn}
			err := tr.Send(ctx, []byte{1, 2, 3}, destination)
			if (err != nil) != test.wantErr {
				t.Fatalf("Send() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

type timeoutStubError struct{}

func (timeoutStubError) Error() string   { return "timeout" }
func (timeoutStubError) Timeout() bool   { return true }
func (timeoutStubError) Temporary() bool { return true }

func TestUDPv6TransportReceiveErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		conn    *ipv6PacketConnStub
		readErr error
		wantErr bool
	}{
		{name: "deadline error", conn: &ipv6PacketConnStub{readDeadlineErr: errors.New("deadline failed")}, wantErr: true},
		{name: "timeout", conn: &ipv6PacketConnStub{}, readErr: timeoutStubError{}, wantErr: true},
		{name: "read error", conn: &ipv6PacketConnStub{}, readErr: errors.New("read failed"), wantErr: true},
		{name: "nil control message", conn: &ipv6PacketConnStub{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := &UDPv6Transport{
				conn: test.conn,
				readFrom: func(buffer []byte) (int, *ipv6.ControlMessage, net.Addr, error) {
					if test.readErr != nil {
						return 0, nil, nil, test.readErr
					}
					buffer[0] = 1
					return 1, nil, &net.UDPAddr{}, nil
				},
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, _, index, err := tr.Receive(ctx)
			if (err != nil) != test.wantErr {
				t.Fatalf("Receive() error = %v, wantErr %v", err, test.wantErr)
			}
			if err == nil && index != 0 {
				t.Fatalf("Receive() interface index = %d, want 0 without a control message", index)
			}
		})
	}
}

func TestUDPv6TransportClose(t *testing.T) {
	if err := (&UDPv6Transport{}).Close(); err != nil {
		t.Fatalf("Close() with nil connection: %v", err)
	}
	if err := (&UDPv6Transport{conn: &ipv6PacketConnStub{}}).Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
	if err := (&UDPv6Transport{conn: &ipv6PacketConnStub{closeErr: errors.New("close failed")}}).Close(); err == nil {
		t.Fatal("Close() succeeded despite connection close error")
	}
}
