package transport

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"golang.org/x/net/ipv6"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// UDPv6Transport implements the IPv6 side of dual-stack mDNS per RFC 6762 §20.
//
// It binds to [::]:5353, joins the mDNS IPv6 multicast group FF02::FB on all
// multicast-capable interfaces, and extracts the receiving interface index from
// IPv6 control messages for RFC 6762 §6.2 interface scoping.
type UDPv6Transport struct {
	conn             net.PacketConn
	ipv6Conn         *ipv6.PacketConn
	joinedInterfaces []int
	writeTo          func([]byte, *ipv6.ControlMessage, net.Addr) (int, error)
	readFrom         func([]byte) (int, *ipv6.ControlMessage, net.Addr, error)
}

// NewUDPv6Transport creates a UDP IPv6 multicast transport per RFC 6762 §§20, 22.
//
// It binds to [::]:5353, applies SO_REUSEADDR/SO_REUSEPORT platform options so
// multiple mDNS daemons can coexist, and joins FF02::FB on every multicast-
// capable interface that is currently up.
//
// Returns NetworkError if no interface can join the multicast group (e.g., the
// host has no IPv6-capable interfaces).
func NewUDPv6Transport() (*UDPv6Transport, error) {
	lc := net.ListenConfig{
		Control: platformControl,
	}

	addr := net.JoinHostPort("::", strconv.Itoa(protocol.Port))
	conn, err := lc.ListenPacket(context.Background(), "udp6", addr)
	if err != nil {
		return nil, &errors.NetworkError{
			Operation: "create IPv6 socket",
			Err:       err,
			Details:   fmt.Sprintf("failed to bind udp6 %s", addr),
		}
	}

	if udpConn, ok := conn.(*net.UDPConn); ok {
		if bufferErr := udpConn.SetReadBuffer(65536); bufferErr != nil {
			_ = conn.Close()
			return nil, &errors.NetworkError{
				Operation: "configure IPv6 socket",
				Err:       bufferErr,
				Details:   "failed to set read buffer size",
			}
		}
	}

	ipv6Conn := ipv6.NewPacketConn(conn)

	// Enable interface index in incoming control messages (RFC 6762 §6.2).
	// Non-fatal on platforms that don't support it; interfaceIndex will be 0.
	_ = ipv6Conn.SetControlMessage(ipv6.FlagInterface, true)

	// Join FF02::FB on every multicast-capable, up interface.
	group := &net.UDPAddr{IP: net.ParseIP(protocol.MulticastAddrIPv6)}
	ifaces, err := net.Interfaces()
	if err != nil {
		_ = conn.Close()
		return nil, &errors.NetworkError{
			Operation: "enumerate interfaces",
			Err:       err,
			Details:   "failed to list network interfaces for IPv6 multicast join",
		}
	}

	var lastJoinErr error
	joined := 0
	joinedInterfaces := make([]int, 0, len(ifaces))
	for i := range ifaces {
		iface := &ifaces[i]
		if iface.Flags&net.FlagMulticast == 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addresses, addrErr := iface.Addrs()
		if addrErr != nil {
			lastJoinErr = addrErr
			continue
		}
		if !hasUsableIPv6Address(addresses) {
			continue
		}
		if err := ipv6Conn.JoinGroup(iface, group); err != nil {
			lastJoinErr = err
			continue
		}
		joined++
		joinedInterfaces = append(joinedInterfaces, iface.Index)
	}

	if joined == 0 {
		_ = conn.Close()
		err := lastJoinErr
		if err == nil {
			err = fmt.Errorf("no multicast-capable interfaces found")
		}
		return nil, &errors.NetworkError{
			Operation: "join IPv6 multicast group",
			Err:       err,
			Details:   fmt.Sprintf("failed to join %s on any interface", protocol.MulticastAddrIPv6),
		}
	}

	return &UDPv6Transport{
		conn:             conn,
		ipv6Conn:         ipv6Conn,
		joinedInterfaces: joinedInterfaces,
		writeTo:          ipv6Conn.WriteTo,
		readFrom:         ipv6Conn.ReadFrom,
	}, nil
}

func hasUsableIPv6Address(addresses []net.Addr) bool {
	for _, address := range addresses {
		var ip net.IP
		switch value := address.(type) {
		case *net.IPNet:
			ip = value.IP
		case *net.IPAddr:
			ip = value.IP
		}
		if ip != nil && ip.To4() == nil && ip.To16() != nil && !ip.IsUnspecified() {
			return true
		}
	}
	return false
}

// Send transmits a packet to dest over IPv6.
//
// RFC 6762 §22: Deployed IPv6 mDNS uses FF02::FB:5353.
func (t *UDPv6Transport) Send(ctx context.Context, packet []byte, dest net.Addr) error {
	select {
	case <-ctx.Done():
		return &errors.NetworkError{
			Operation: "send IPv6 query",
			Err:       ctx.Err(),
			Details:   "context canceled before send",
		}
	default:
	}

	udpDest, isUDP := dest.(*net.UDPAddr)
	if isUDP && udpDest.IP.IsMulticast() && udpDest.IP.To4() == nil && udpDest.Zone == "" {
		return t.sendMulticastOnJoinedInterfaces(packet, dest)
	}

	n, err := t.conn.WriteTo(packet, dest)
	if err != nil {
		return &errors.NetworkError{
			Operation: "send IPv6 query",
			Err:       err,
			Details:   fmt.Sprintf("failed to send %d bytes to %s", len(packet), dest),
		}
	}
	if n != len(packet) {
		return &errors.NetworkError{
			Operation: "send IPv6 query",
			Err:       fmt.Errorf("partial write: %d/%d bytes", n, len(packet)),
			Details:   "incomplete transmission",
		}
	}
	return nil
}

func (t *UDPv6Transport) sendMulticastOnJoinedInterfaces(packet []byte, dest net.Addr) error {
	var lastErr error
	sent := 0
	for _, ifIndex := range t.joinedInterfaces {
		n, err := t.writeTo(packet, &ipv6.ControlMessage{IfIndex: ifIndex}, dest)
		if err != nil {
			lastErr = err
			continue
		}
		if n != len(packet) {
			lastErr = fmt.Errorf("partial write on interface %d: %d/%d bytes", ifIndex, n, len(packet))
			continue
		}
		sent++
	}
	if sent == len(t.joinedInterfaces) && sent > 0 {
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("transport has no joined IPv6 multicast interfaces")
	}
	return &errors.NetworkError{
		Operation: "send IPv6 multicast query",
		Err:       lastErr,
		Details:   fmt.Sprintf("failed to send %d bytes to %s", len(packet), dest),
	}
}

// Receive waits for an incoming IPv6 packet, respecting context cancellation.
//
// Returns the interface index from the IPv6 control message when available
// (RFC 6762 §6.2); zero indicates the interface is unknown.
func (t *UDPv6Transport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
	select {
	case <-ctx.Done():
		return nil, nil, 0, &errors.NetworkError{
			Operation: "receive IPv6 response",
			Err:       ctx.Err(),
			Details:   "context canceled before receive",
		}
	default:
	}

	if deadline, ok := ctx.Deadline(); ok {
		if err := t.conn.SetReadDeadline(deadline); err != nil {
			return nil, nil, 0, &errors.NetworkError{
				Operation: "set IPv6 read deadline",
				Err:       err,
				Details:   fmt.Sprintf("failed to set deadline %v", deadline),
			}
		}
	}

	bufPtr := GetBuffer()
	defer PutBuffer(bufPtr)
	buffer := *bufPtr

	n, cm, srcAddr, err := t.readFrom(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil, 0, &errors.NetworkError{
				Operation: "receive IPv6 response",
				Err:       err,
				Details:   "timeout",
			}
		}
		return nil, nil, 0, &errors.NetworkError{
			Operation: "receive IPv6 response",
			Err:       err,
			Details:   "failed to read from IPv6 socket",
		}
	}

	interfaceIndex := 0
	if cm != nil {
		interfaceIndex = cm.IfIndex
	}

	result := make([]byte, n)
	copy(result, buffer[:n])
	return result, srcAddr, interfaceIndex, nil
}

// Close releases the IPv6 socket.
func (t *UDPv6Transport) Close() error {
	if t.conn == nil {
		return nil
	}
	if err := t.conn.Close(); err != nil {
		return &errors.NetworkError{
			Operation: "close IPv6 socket",
			Err:       err,
			Details:   "failed to close UDP IPv6 connection",
		}
	}
	return nil
}

// Compile-time verification that UDPv6Transport implements Transport.
var _ Transport = (*UDPv6Transport)(nil)
