package responder

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"

	internalerrors "github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/message"
	"github.com/joshuafuller/beacon/internal/protocol"
	internalresponder "github.com/joshuafuller/beacon/internal/responder"
)

// RFC 6762 §20: A dual-stack host should register its names in both the IPv4
// and IPv6 .local zones. RFC 6762 §8.4 applies that rule to record updates.
func TestUpdateServiceAnnouncesOnBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	ipv4Transport := &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
		ipv4Sends++
		return nil
	}}
	ipv6Transport := &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
		ipv6Sends++
		return nil
	}}
	r := &Responder{
		ctx:        context.Background(),
		transport:  ipv4Transport,
		transport6: ipv6Transport,
		registry:   internalresponder.NewRegistry(),
		hostname:   "host.local",
	}
	if err := r.registry.Register(&internalresponder.Service{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}); err != nil {
		t.Fatalf("register fixture: %v", err)
	}

	if err := r.UpdateService("Example", map[string]string{"version": "2"}); err != nil {
		t.Fatalf("UpdateService() failed: %v", err)
	}
	if ipv4Sends != 1 || ipv6Sends != 1 {
		t.Fatalf("update sends: IPv4=%d IPv6=%d, want one announcement in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends)
	}
}

// RFC 6762 §§10.1, 20: A dual-stack responder sends goodbye records into
// both independent .local zones when a service is unregistered.
func TestUnregisterSendsGoodbyeOnBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	r := &Responder{
		ctx: context.Background(),
		transport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv4Sends++
			return nil
		}},
		transport6: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv6Sends++
			return nil
		}},
		registry: internalresponder.NewRegistry(),
		hostname: "host.local",
	}
	if err := r.registry.Register(&internalresponder.Service{
		InstanceName: "Example",
		ServiceType:  "_http._tcp.local",
		Port:         8080,
	}); err != nil {
		t.Fatalf("register fixture: %v", err)
	}

	if err := r.Unregister("Example"); err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}
	if ipv4Sends != 1 || ipv6Sends != 1 {
		t.Fatalf("goodbye sends: IPv4=%d IPv6=%d, want one goodbye in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends)
	}
}

// RFC 6762 §§8.1, 8.3, 20: A dual-stack registration probes and announces
// in both independent .local zones before becoming established.
func TestProbeAndAnnounceUsesBothIPVersions(t *testing.T) {
	var ipv4Sends, ipv6Sends int
	r := &Responder{
		ctx: context.Background(),
		transport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv4Sends++
			return nil
		}},
		transport6: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			ipv6Sends++
			return nil
		}},
	}

	if _, err := r.runProbeAndAnnounce(nil, "Example._http._tcp.local"); err != nil {
		t.Fatalf("runProbeAndAnnounce() failed: %v", err)
	}
	const lifecyclePackets = 5 // Three probes plus two announcements.
	if ipv4Sends != lifecyclePackets || ipv6Sends != lifecyclePackets {
		t.Fatalf("lifecycle sends: IPv4=%d IPv6=%d, want %d in each RFC 6762 §20 zone", ipv4Sends, ipv6Sends, lifecyclePackets)
	}
}

// RFC 6762 §6.2: Address records in a response are scoped to the sending
// interface, and every valid address on that interface is included.
func TestProjectLifecycleRecordsForInterface(t *testing.T) {
	records := []*message.ResourceRecord{
		{Name: "_http._tcp.local", Type: protocol.RecordTypePTR, Data: []byte{1}},
		{Name: "host.local", Type: protocol.RecordTypeA, Data: []byte{192, 0, 2, 99}},
		{Name: "host.local", Type: protocol.RecordTypeAAAA, Data: net.ParseIP("2001:db8:ffff::1").To16()},
	}
	wantIPv4 := []byte{192, 0, 2, 7}
	wantIPv6 := []net.IP{net.ParseIP("fe80::7"), net.ParseIP("2001:db8:7::1")}

	projected := projectLifecycleRecordsForInterface(records, wantIPv4, wantIPv6)
	var ptrCount, aCount, aaaaCount int
	for _, record := range projected {
		switch record.Type {
		case protocol.RecordTypePTR:
			ptrCount++
		case protocol.RecordTypeA:
			aCount++
			if !bytes.Equal(record.Data, wantIPv4) {
				t.Fatalf("projected A = %v, want %v", record.Data, wantIPv4)
			}
		case protocol.RecordTypeAAAA:
			if aaaaCount >= len(wantIPv6) || !bytes.Equal(record.Data, wantIPv6[aaaaCount].To16()) {
				t.Fatalf("projected AAAA %d = %v", aaaaCount, record.Data)
			}
			aaaaCount++
		}
	}
	if ptrCount != 1 || aCount != 1 || aaaaCount != 2 {
		t.Fatalf("projected counts: PTR=%d A=%d AAAA=%d, want 1/1/2", ptrCount, aCount, aaaaCount)
	}
}

type scopedIPv6LifecycleTransport struct {
	*MockTransport
	joined        []int
	interfaceSend map[int][]byte
}

func (t *scopedIPv6LifecycleTransport) JoinedInterfaceIndexes() []int {
	return append([]int(nil), t.joined...)
}

func (t *scopedIPv6LifecycleTransport) SendOnInterface(_ context.Context, packet []byte, _ net.Addr, ifIndex int) error {
	if t.interfaceSend == nil {
		t.interfaceSend = make(map[int][]byte)
	}
	t.interfaceSend[ifIndex] = append([]byte(nil), packet...)
	return nil
}

func TestRegistrationAnnouncementUsesInterfaceScopedIPv6Packets(t *testing.T) {
	var genericIPv6Sends int
	scoped := &scopedIPv6LifecycleTransport{
		MockTransport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			genericIPv6Sends++
			return nil
		}},
		joined: []int{2, 7},
	}
	tr := &registrationMulticastTransport{
		ipv4: &MockTransport{},
		ipv6: scoped,
		ipv6Announcements: map[int][]byte{
			2: {0, 0, 0x80, 2},
			7: {0, 0, 0x80, 7},
		},
	}

	if err := tr.Send(context.Background(), []byte{0, 0, 0x80}, &net.UDPAddr{}); err != nil {
		t.Fatalf("Send() failed: %v", err)
	}
	if genericIPv6Sends != 0 {
		t.Fatalf("generic IPv6 sends = %d, want interface-scoped sends", genericIPv6Sends)
	}
	for _, ifIndex := range []int{2, 7} {
		packet := scoped.interfaceSend[ifIndex]
		if len(packet) != 4 || packet[3] != byte(ifIndex) {
			t.Fatalf("interface %d packet = %v, want its scoped projection", ifIndex, packet)
		}
	}
}

func TestBuildIPv6AnnouncementProjections(t *testing.T) {
	records := []*message.ResourceRecord{
		{Name: "_http._tcp.local", Type: protocol.RecordTypePTR, Class: protocol.ClassIN, TTL: 120, Data: []byte{1}},
		{Name: "host.local", Type: protocol.RecordTypeA, Class: protocol.ClassIN, TTL: protocol.TTLHostname, Data: []byte{192, 0, 2, 99}},
	}
	lookup := func(ifIndex int) ([]byte, []net.IP, error) {
		return []byte{192, 0, 2, byte(ifIndex)}, []net.IP{net.ParseIP("2001:db8::" + string(rune('0'+ifIndex)))}, nil
	}

	projections, err := buildIPv6AnnouncementProjections(records, []int{2, 7}, lookup)
	if err != nil {
		t.Fatalf("buildIPv6AnnouncementProjections() failed: %v", err)
	}
	for _, ifIndex := range []int{2, 7} {
		response, parseErr := parseMessage(projections[ifIndex])
		if parseErr != nil {
			t.Fatalf("parse interface %d projection: %v", ifIndex, parseErr)
		}
		var gotA, gotAAAA []byte
		for _, answer := range response.Answers {
			switch answer.TYPE {
			case uint16(protocol.RecordTypeA):
				gotA = answer.RDATA
			case uint16(protocol.RecordTypeAAAA):
				gotAAAA = answer.RDATA
			}
		}
		wantA := []byte{192, 0, 2, byte(ifIndex)}
		if !bytes.Equal(gotA, wantA) {
			t.Fatalf("interface %d A record = %v, want %v", ifIndex, gotA, wantA)
		}
		wantAAAA := net.ParseIP("2001:db8::" + string(rune('0'+ifIndex))).To16()
		if !bytes.Equal(gotAAAA, wantAAAA) {
			t.Fatalf("interface %d AAAA record = %v, want %v", ifIndex, gotAAAA, wantAAAA)
		}
	}
}

func TestLookupLifecycleAddresses_AllowsIPv6OnlyInterface(t *testing.T) {
	ipv4, ipv6, err := lookupLifecycleAddressesWith(
		7,
		func(int) ([]net.IP, error) { return []net.IP{net.ParseIP("fe80::7")}, nil },
		func(int) ([]byte, error) {
			return nil, &internalerrors.ValidationError{Field: "interface", Message: "no IPv4 address"}
		},
	)
	if err != nil {
		t.Fatalf("lookupLifecycleAddressesWith() error = %v, want nil for IPv6-only interface", err)
	}
	if ipv4 != nil || len(ipv6) != 1 {
		t.Fatalf("lookupLifecycleAddressesWith() = (%v, %v), want (nil, one IPv6 address)", ipv4, ipv6)
	}
}

func TestLookupLifecycleAddresses_PropagatesNetworkError(t *testing.T) {
	wantErr := &internalerrors.NetworkError{Operation: "get interface addresses"}
	_, _, err := lookupLifecycleAddressesWith(
		7,
		func(int) ([]net.IP, error) { return []net.IP{net.ParseIP("fe80::7")}, nil },
		func(int) ([]byte, error) { return nil, wantErr },
	)
	var gotErr *internalerrors.NetworkError
	if !errors.As(err, &gotErr) {
		t.Fatalf("lookupLifecycleAddressesWith() error = %v, want NetworkError", err)
	}
}

// RFC 6762 §10.1: Every address record in a projected goodbye packet retains
// TTL zero so peers on each interface remove it immediately.
func TestProjectLifecycleRecordsPreservesGoodbyeTTL(t *testing.T) {
	records := []*message.ResourceRecord{
		{Name: "host.local", Type: protocol.RecordTypeA, Class: protocol.ClassIN, TTL: 0, Data: []byte{192, 0, 2, 1}, CacheFlush: true},
	}
	projected := projectLifecycleRecordsForInterface(records, []byte{192, 0, 2, 2}, []net.IP{net.ParseIP("fe80::2")})
	for _, record := range projected {
		if (record.Type == protocol.RecordTypeA || record.Type == protocol.RecordTypeAAAA) && record.TTL != 0 {
			t.Fatalf("projected goodbye %v TTL = %d, want 0 per RFC 6762 §10.1", record.Type, record.TTL)
		}
	}
}

func TestSendLifecycleRecordsUsesScopedIPv6Projections(t *testing.T) {
	var genericIPv6Sends int
	scoped := &scopedIPv6LifecycleTransport{
		MockTransport: &MockTransport{sendFunc: func(context.Context, []byte, net.Addr) error {
			genericIPv6Sends++
			return nil
		}},
		joined: []int{2, 7},
	}
	r := &Responder{
		ctx:        context.Background(),
		transport:  &MockTransport{},
		transport6: scoped,
	}
	records := []*message.ResourceRecord{
		{Name: "host.local", Type: protocol.RecordTypeA, Class: protocol.ClassIN, TTL: protocol.TTLHostname, Data: []byte{192, 0, 2, 99}, CacheFlush: true},
	}
	lookup := func(ifIndex int) ([]byte, []net.IP, error) {
		return []byte{192, 0, 2, byte(ifIndex)}, []net.IP{net.ParseIP("fe80::" + string(rune('0'+ifIndex)))}, nil
	}

	if err := r.sendLifecycleRecords(records, lookup); err != nil {
		t.Fatalf("sendLifecycleRecords() failed: %v", err)
	}
	if genericIPv6Sends != 0 || len(scoped.interfaceSend) != 2 {
		t.Fatalf("IPv6 sends: generic=%d scoped=%d, want 0/2", genericIPv6Sends, len(scoped.interfaceSend))
	}
}
