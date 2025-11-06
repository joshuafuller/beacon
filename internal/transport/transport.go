// Package transport provides network transport abstractions for mDNS communication.
//
// This package decouples the querier from specific network implementations,
// enabling IPv4, IPv6, and mock transports (FR-001, US1).
//
// Design Pattern: From specs/003-m1-refactoring/research.md Topic 1
package transport

import (
	"context"
	"net"
)

// Transport abstracts network operations for sending and receiving mDNS packets.
//
// This interface enables:
// - IPv4/IPv6 transport implementations (US1: FR-001)
// - Mock transports for testing (improves testability)
// - Context-aware operations (M1.1 F-9 REQ-F9-7 alignment)
//
// Implementations:
// - UDPv4Transport: Production IPv4 multicast transport
// - MockTransport: Test double for unit testing
//
// T019: Minimal interface to make T010 pass (TDD GREEN)
type Transport interface {
	// Send transmits a packet to the specified destination address.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline propagation
	//   - packet: DNS message in wire format
	//   - dest: Destination address (e.g., mDNS multicast 224.0.0.251:5353)
	//
	// Returns:
	//   - error: NetworkError on transmission failure
	Send(ctx context.Context, packet []byte, dest net.Addr) error

	// Receive waits for an incoming packet, respecting context cancellation/deadline.
	//
	// 007-interface-specific-addressing: Added interfaceIndex return value for RFC 6762 §15 compliance.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline propagation
	//
	// Returns:
	//   - packet: DNS response message in wire format
	//   - srcAddr: Source address of the response
	//   - interfaceIndex: OS interface index that received the packet (from IP_PKTINFO/IP_RECVIF)
	//                     Zero (0) indicates interface unknown (graceful degradation)
	//   - error: NetworkError on timeout or receive failure
	//
	// RFC 6762 §15: Interface index enables building responses with addresses valid on
	// the receiving interface only (MUST include interface IP, MUST NOT include other IPs).
	//
	// Context handling (F-9 REQ-F9-7):
	//   - ctx.Done(): Return immediately on cancellation
	//   - ctx.Deadline(): Propagate deadline to socket SetReadDeadline
	Receive(ctx context.Context) (packet []byte, srcAddr net.Addr, interfaceIndex int, err error)

	// Close releases network resources.
	//
	// Returns:
	//   - error: NetworkError on close failure (FR-004: must propagate errors, not swallow)
	Close() error
}
