// Package integration provides end-to-end IPv6 integration tests per RFC 6762 §§20, 22.
package integration

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/protocol"
	"github.com/joshuafuller/beacon/internal/transport"
	"github.com/joshuafuller/beacon/querier"
)

func TestIPv6MulticastTransportRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping IPv6 network integration test in short mode")
	}

	tr, err := transport.NewUDPv6Transport()
	if err != nil {
		t.Skipf("host has no usable IPv6 multicast transport: %v", err)
	}
	defer func() { _ = tr.Close() }()

	want := []byte("beacon-ipv6-integration")
	if err := tr.Send(context.Background(), want, protocol.MulticastGroupIPv6()); err != nil {
		t.Fatalf("IPv6 multicast send failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		got, _, ifIndex, err := tr.Receive(ctx)
		if err != nil {
			t.Fatalf("did not receive the IPv6 multicast packet: %v", err)
		}
		if bytes.Equal(got, want) {
			if ifIndex == 0 {
				t.Fatal("IPv6 multicast packet arrived without a receiving interface index")
			}
			return
		}
	}
}

// TestQuerier_IPv6_Transport validates that a dual-stack querier can be created
// and issues AAAA queries on the IPv6 multicast transport (FF02::FB:5353).
//
// RFC 6762 §22: Deployed IPv6 mDNS uses FF02::FB on port 5353.
//
// This test skips if IPv6 is unavailable on the host (graceful degradation).
// It accepts empty results — the network may have no IPv6 mDNS responders.
func TestQuerier_IPv6_Transport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping IPv6 integration test in short mode")
	}

	q, err := querier.New(querier.WithIPv6())
	if err != nil {
		t.Fatalf("querier.New(WithIPv6()) failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := q.Query(ctx, "_services._dns-sd._udp.local", querier.RecordTypeAAAA)
	if err != nil {
		t.Logf("AAAA query returned error (acceptable if no IPv6 mDNS on network): %v", err)
		return
	}

	if response == nil || len(response.Records) == 0 {
		t.Log("AAAA query returned no records (no IPv6 mDNS responders on network)")
		return
	}

	t.Logf("AAAA query SUCCESS: received %d records", len(response.Records))
	for i, rec := range response.Records {
		if rec.Type == querier.RecordTypeAAAA {
			ip := rec.AsAAAA()
			if ip == nil {
				t.Errorf("Record[%d]: AsAAAA() returned nil for AAAA record", i)
				continue
			}
			t.Logf("  Record[%d]: %s → %s (TTL=%d)", i, rec.Name, ip, rec.TTL)
		}
	}
}

// TestQuerier_IPv6_DualStack validates that a dual-stack querier initializes correctly
// and can send queries on both IPv4 and IPv6 transports simultaneously.
//
// RFC 6762 §20: Dual-stack hosts should register and look up names using both IPv4 and IPv6.
//
// Test strategy: construct the querier (validates transport creation), issue a
// standard PTR query, then an AAAA query. Both must complete without panicking.
// Results are logged rather than asserted — the test environment may have no
// mDNS traffic at all.
func TestQuerier_IPv6_DualStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping IPv6 dual-stack integration test in short mode")
	}

	q, err := querier.New(querier.WithIPv6())
	if err != nil {
		t.Fatalf("querier.New(WithIPv6()) failed: %v", err)
	}
	defer func() { _ = q.Close() }()

	// PTR query on IPv4 transport
	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel4()
	resp4, err4 := q.Query(ctx4, "_services._dns-sd._udp.local", querier.RecordTypePTR)
	if err4 != nil {
		t.Logf("IPv4 PTR query: %v (acceptable timeout/no traffic)", err4)
	} else {
		t.Logf("IPv4 PTR query: %d records received", len(resp4.Records))
	}

	// AAAA query on IPv6 transport
	ctx6, cancel6 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel6()
	resp6, err6 := q.Query(ctx6, "_services._dns-sd._udp.local", querier.RecordTypeAAAA)
	if err6 != nil {
		t.Logf("IPv6 AAAA query: %v (acceptable if no IPv6 mDNS on network)", err6)
	} else {
		t.Logf("IPv6 AAAA query: %d records received", len(resp6.Records))
	}
}
