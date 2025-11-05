package transport

import (
	"context"
	"net"
)

// UDPv6Transport is a stub implementation to validate Transport interface extensibility.
//
// This stub demonstrates that the Transport interface successfully enables IPv6 support
// without requiring changes to the querier package (FR-001 validation).
//
// T042: Validates Transport interface extensibility for M1.1 IPv6 support
//
// Full implementation will be added in M1.1 (F-9, F-10).
type UDPv6Transport struct {
	//lint:ignore U1000 M2: Will be used in full IPv6 implementation
	conn net.PacketConn
}

// NewUDPv6Transport creates a UDP IPv6 multicast transport (stub).
//
// M1.1 TODO: Implement full IPv6 multicast support per F-9 REQ-F9-1
func NewUDPv6Transport() (*UDPv6Transport, error) {
	// Stub: Full implementation in M1.1
	return nil, nil
}

// Send transmits a packet over IPv6 (stub).
func (t *UDPv6Transport) Send(_ context.Context, _ []byte, _ net.Addr) error {
	// Stub: Full implementation in M1.1
	return nil
}

// Receive waits for an incoming IPv6 packet (stub).
func (t *UDPv6Transport) Receive(_ context.Context) ([]byte, net.Addr, error) {
	// Stub: Full implementation in M1.1
	return nil, nil, nil
}

// Close releases IPv6 resources (stub).
func (t *UDPv6Transport) Close() error {
	// Stub: Full implementation in M1.1
	return nil
}

// Compile-time verification that UDPv6Transport implements Transport interface
var _ Transport = (*UDPv6Transport)(nil)
