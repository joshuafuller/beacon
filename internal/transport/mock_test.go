package transport_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/transport"
)

// TDD - RED Phase: Tests for MockTransport
// These tests are written FIRST, before implementation exists
// Expected: COMPILATION ERRORS (MockTransport doesn't exist yet)

// T012: Contract test - MockTransport implements Transport interface
// NOTE: This test will FAIL to compile until MockTransport is defined in T025
func TestMockTransport_ImplementsTransportInterface(_ *testing.T) {
	// This will fail to compile until MockTransport exists
	var _ transport.Transport = (*transport.MockTransport)(nil)
}

// T017: Unit test - MockTransport.Send() records calls for verification
// NOTE: This test will FAIL to compile until MockTransport exists (T025)
func TestMockTransport_Send_RecordsCalls(t *testing.T) {
	mock := transport.NewMockTransport()
	defer func() { _ = mock.Close() }()

	ctx := context.Background()
	packet1 := []byte{0x01, 0x02}
	packet2 := []byte{0x03, 0x04}
	addr1 := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	addr2 := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 252), Port: 5353}

	// Send two packets
	err := mock.Send(ctx, packet1, addr1)
	if err != nil {
		t.Fatalf("Send(packet1) failed: %v", err)
	}

	err = mock.Send(ctx, packet2, addr2)
	if err != nil {
		t.Fatalf("Send(packet2) failed: %v", err)
	}

	// Verify calls were recorded
	calls := mock.SendCalls()
	if len(calls) != 2 {
		t.Fatalf("Expected 2 Send() calls, got %d", len(calls))
	}

	// Verify first call
	if string(calls[0].Packet) != string(packet1) {
		t.Errorf("First call packet mismatch: got %v, want %v", calls[0].Packet, packet1)
	}
	if calls[0].Dest.String() != addr1.String() {
		t.Errorf("First call addr mismatch: got %v, want %v", calls[0].Dest, addr1)
	}

	// Verify second call
	if string(calls[1].Packet) != string(packet2) {
		t.Errorf("Second call packet mismatch: got %v, want %v", calls[1].Packet, packet2)
	}
	if calls[1].Dest.String() != addr2.String() {
		t.Errorf("Second call addr mismatch: got %v, want %v", calls[1].Dest, addr2)
	}
}

func TestMockTransport_Receive_QueuedResponse(t *testing.T) {
	mock := transport.NewMockTransport()
	defer func() { _ = mock.Close() }()

	packet := []byte{0x01, 0x02}
	addr := &net.UDPAddr{IP: net.IPv4(192, 0, 2, 1), Port: 5353}
	mock.QueueReceive(packet, addr, 7)
	packet[0] = 0xff

	gotPacket, gotAddr, gotIfIndex, err := mock.Receive(context.Background())
	if err != nil {
		t.Fatalf("Receive() failed: %v", err)
	}
	if gotPacket[0] != 0x01 || gotPacket[1] != 0x02 {
		t.Fatalf("Receive() packet = %v, want [1 2]", gotPacket)
	}
	if gotAddr.String() != addr.String() || gotIfIndex != 7 {
		t.Fatalf("Receive() address/index = %v/%d, want %v/7", gotAddr, gotIfIndex, addr)
	}
}

func TestMockTransport_Receive_RespectsContextCancellation(t *testing.T) {
	mock := transport.NewMockTransport()
	defer func() { _ = mock.Close() }()
	mock.EnableBlockingReceive()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, _, err := mock.Receive(ctx)
	if err == nil {
		t.Fatal("Receive() returned nil error for canceled context")
	}
}

func TestMockTransport_Receive_BlockingSpuriousWakeContinuesWaiting(t *testing.T) {
	mock := transport.NewMockTransport()
	defer func() { _ = mock.Close() }()

	mock.QueueReceive([]byte{0x01}, nil, 0)
	if _, _, _, err := mock.Receive(context.Background()); err != nil {
		t.Fatalf("Receive() failed while consuming queued response: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, _, _, err := mock.Receive(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Receive() error = %v, want context deadline after spurious wake", err)
	}
}
