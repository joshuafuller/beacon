// Package security provides security features including rate limiting
// and source IP validation for mDNS multicast traffic.
package security

import (
	"net"
)

// SourceFilter validates source IPs before parsing packets.
// Per RFC 6762 §2, mDNS is link-local scope - source IPs must be
// link-local (169.254.0.0/16) or same subnet as receiving interface.
type SourceFilter struct {
	iface      net.Interface // Receiving interface
	ifaceAddrs []net.IPNet   // Cached interface addresses (avoids syscall per packet)
}

// NewSourceFilter creates a new source filter for the given interface.
// It caches the interface addresses to avoid syscalls in the hot path (per-packet validation).
//
// Task T073: Initialize SourceFilter with cached interface addresses
func NewSourceFilter(iface net.Interface) (*SourceFilter, error) {
	// Get interface addresses
	addrs, err := iface.Addrs()
	if err != nil {
		// If we can't get addresses, create filter without cached addresses
		// IsValid() will fall back to link-local check only
		return &SourceFilter{
			iface:      iface,
			ifaceAddrs: []net.IPNet{},
		}, nil
	}

	// Extract IPNet addresses and cache them
	var ipnets []net.IPNet
	for _, addr := range addrs {
		// addr is *net.IPNet or *net.IPAddr
		if ipnet, ok := addr.(*net.IPNet); ok {
			ipnets = append(ipnets, *ipnet)
		}
	}

	return &SourceFilter{
		iface:      iface,
		ifaceAddrs: ipnets,
	}, nil
}

// IsValid checks if the source IP is valid for mDNS (link-local or same subnet).
// Returns false for non-link-local IPs outside the receiving interface's subnet.
//
// Per RFC 6762 §2, mDNS is link-local scope. Valid source IPs are:
// 1. IPv4 link-local (169.254.0.0/16) - RFC 3927
// 2. Same subnet as the receiving interface
//
// Task T074: Implement link-local + same subnet check per FR-023
func (sf *SourceFilter) IsValid(srcIP net.IP) bool {
	// Convert to IPv4 if possible
	ip4 := srcIP.To4()
	if ip4 == nil {
		// IPv6 support deferred to M2
		// For now, reject IPv6 packets
		return false
	}

	// Check 1: IPv4 link-local (169.254.0.0/16) - RFC 3927
	// Link-local addresses are ALWAYS valid per RFC 6762 §2
	if ip4[0] == 169 && ip4[1] == 254 {
		return true // RFC 3927 link-local address
	}

	// Check 2: Same subnet as interface
	// Packets from the same subnet as the receiving interface are valid
	for _, ipnet := range sf.ifaceAddrs {
		if ipnet.Contains(srcIP) {
			return true // Same subnet as interface
		}
	}

	// Not link-local and not same subnet - reject
	return false
}

// IsPrivate returns true if the IP is in a private address range
// (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16).
func IsPrivate(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false // Not IPv4
	}

	// Check private ranges:
	// 10.0.0.0/8
	if ip4[0] == 10 {
		return true
	}

	// 172.16.0.0/12 (172.16.0.0 - 172.31.255.255)
	if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
		return true
	}

	// 192.168.0.0/16
	if ip4[0] == 192 && ip4[1] == 168 {
		return true
	}

	return false
}
