package responder

import "testing"

func TestIPv6ResponseAddressesRejectUnknownInterface(t *testing.T) {
	if addresses, err := getIPv6ResponseAddresses(0); err == nil {
		t.Fatalf("getIPv6ResponseAddresses(0) = %v, nil; want an error rather than cross-interface addresses", addresses)
	}
}
