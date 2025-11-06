// Package contracts defines the API contracts for interface-specific addressing.
//
// This is a specification file - NOT compiled code. It documents the expected
// interface changes and serves as a contract for implementation.
package contracts

import (
	"context"
	"net"
)

// Transport defines the network I/O abstraction for mDNS.
//
// CHANGE: Added interfaceIndex return value to Receive()
// RFC 6762 §15: Interface index enables building responses with addresses
// valid on the receiving interface.
//
// Version: 2.0 (007-interface-specific-addressing)
// Previous: 1.0 (003-m1-refactoring)
type Transport interface {
	// Send transmits a packet to the specified destination address.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - packet: DNS message bytes to send
	//   - dest: Destination address (nil = multicast to 224.0.0.251:5353)
	//
	// Returns:
	//   - error: NetworkError on transmission failure
	//
	// Unchanged from v1.0
	Send(ctx context.Context, packet []byte, dest net.Addr) error

	// Receive waits for an incoming packet, returning packet data, source address,
	// and the network interface index that received the packet.
	//
	// BREAKING CHANGE: Added interfaceIndex return value
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//
	// Returns:
	//   - packet: Received DNS message bytes
	//   - src: Source address of the querier
	//   - interfaceIndex: OS interface index (from IP_PKTINFO/IP_RECVIF control messages)
	//                     Zero (0) indicates interface unknown (fallback to getLocalIPv4)
	//   - error: NetworkError on timeout or receive failure
	//
	// RFC 6762 §15 Compliance:
	// The interfaceIndex enables the responder to look up the IPv4 address
	// assigned to the receiving interface and include ONLY that address in
	// the response A record.
	//
	// Platform Support:
	//   - Linux: IP_PKTINFO socket option provides ifIndex
	//   - macOS/BSD: IP_RECVIF socket option provides ifIndex
	//   - Windows: LPFN_WSARECVMSG provides ifIndex
	//
	// Implementation Note:
	// Uses golang.org/x/net/ipv4.PacketConn.ReadFrom() to access control messages.
	//
	// Changed in v2.0: Added interfaceIndex return value
	Receive(ctx context.Context) (packet []byte, src net.Addr, interfaceIndex int, err error)

	// Close releases network resources.
	//
	// Returns:
	//   - error: NetworkError if close fails
	//
	// Unchanged from v1.0
	Close() error
}

// Migration Guide for Transport Implementers
//
// 1. UDPv4Transport (internal/transport/udp.go):
//    - Add ipv4.PacketConn wrapper field
//    - Call SetControlMessage(ipv4.FlagInterface, true) in NewUDPv4Transport
//    - Use ipv4Conn.ReadFrom() instead of conn.ReadFrom()
//    - Extract cm.IfIndex and return as interfaceIndex
//
// 2. MockTransport (internal/transport/mock_transport.go):
//    - Add InterfaceIndex field to ReceiveResponse struct
//    - Return mock.ReceiveResponses[i].InterfaceIndex
//    - Update all test fixtures to include InterfaceIndex
//
// 3. Test Callsites:
//    - Update all t.transport.Receive() calls to handle 4 return values
//    - packet, src, ifIndex, err := transport.Receive(ctx)
//    - Use ifIndex in test assertions
//
// Example Migration:
//
// Before (v1.0):
//   packet, src, err := transport.Receive(ctx)
//
// After (v2.0):
//   packet, src, ifIndex, err := transport.Receive(ctx)
//   if ifIndex > 0 {
//       // Use interface-specific logic
//   }
