package responder

import (
	"net"
	"testing"
)

// #56: an address bound to the loopback interface (WSL puts 10.255.255.254
// on lo) is not reachable by other hosts and must not be announced.
func TestGetLocalIPv4_SkipsLoopbackInterface(t *testing.T) {
	_, loNet, _ := net.ParseCIDR("10.255.255.254/32")
	_, ethNet, _ := net.ParseCIDR("172.31.243.86/20")
	ethNet.IP = net.IPv4(172, 31, 243, 86)
	_, downNet, _ := net.ParseCIDR("192.168.50.2/24")
	downNet.IP = net.IPv4(192, 168, 50, 2)
	cands := []ifaceAddrs{
		{flags: net.FlagUp | net.FlagLoopback, addrs: []net.Addr{loNet}},
		{flags: net.FlagMulticast, addrs: []net.Addr{downNet}}, // down
		{flags: net.FlagUp | net.FlagMulticast, addrs: []net.Addr{ethNet}},
	}
	got, err := pickIPv4(cands)
	if err != nil {
		t.Fatal(err)
	}
	if !net.IP(got).Equal(net.IPv4(172, 31, 243, 86)) {
		t.Errorf("picked %v, want 172.31.243.86 (eth0)", net.IP(got))
	}
}
