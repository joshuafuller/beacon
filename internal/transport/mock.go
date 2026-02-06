package transport

import (
	"context"
	"net"
	"sync"
)

// MockTransport is a test double for Transport interface.
//
// This mock records all Send() calls for verification in tests,
// enabling unit testing of querier without real network sockets.
//
// T025: For testing, make T012 and T017 pass
type MockTransport struct {
	mu              sync.Mutex
	sendCalls       []SendCall
	closed          bool
	receiveQueue    []mockReceiveResponse // Queued responses for Receive()
	receiveNotifyCh chan struct{}          // Signals when a new response is queued
	blockOnReceive  bool                  // When true, Receive blocks until data or ctx cancel
}

// mockReceiveResponse holds a prepared response for Receive().
type mockReceiveResponse struct {
	Packet []byte
	Addr   net.Addr
	IfIdx  int
}

// SendCall records a single Send() invocation.
type SendCall struct {
	Packet []byte
	Dest   net.Addr
}

// NewMockTransport creates a new mock transport for testing.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		sendCalls:       make([]SendCall, 0),
		receiveNotifyCh: make(chan struct{}, 64),
	}
}

// Send records the call for verification.
//
// T017: MockTransport.Send() records calls for verification
func (m *MockTransport) Send(_ context.Context, packet []byte, dest net.Addr) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.sendCalls = append(m.sendCalls, SendCall{
		Packet: append([]byte(nil), packet...), // Copy to avoid aliasing
		Dest:   dest,
	})

	return nil
}

// Receive returns the next queued response. Behavior depends on blocking mode:
//   - Non-blocking (default): Returns immediately with nil data if queue is empty.
//     This preserves backward compatibility for existing tests.
//   - Blocking (enabled via EnableBlockingReceive or QueueReceive): Blocks until a
//     response is available or the context is cancelled.
//
// 007-interface-specific-addressing: Updated to return interfaceIndex (T012-T013)
// Extended to support queued responses for prober conflict detection testing.
func (m *MockTransport) Receive(ctx context.Context) ([]byte, net.Addr, int, error) {
	// Fast path: check if a response is already queued
	m.mu.Lock()
	if len(m.receiveQueue) > 0 {
		resp := m.receiveQueue[0]
		m.receiveQueue = m.receiveQueue[1:]
		m.mu.Unlock()
		return resp.Packet, resp.Addr, resp.IfIdx, nil
	}
	blocking := m.blockOnReceive
	m.mu.Unlock()

	// Non-blocking mode (backward compatible): return immediately
	if !blocking {
		return nil, nil, 0, nil
	}

	// Blocking mode: wait for a response or context cancellation
	select {
	case <-ctx.Done():
		return nil, nil, 0, ctx.Err()
	case <-m.receiveNotifyCh:
		m.mu.Lock()
		if len(m.receiveQueue) > 0 {
			resp := m.receiveQueue[0]
			m.receiveQueue = m.receiveQueue[1:]
			m.mu.Unlock()
			return resp.Packet, resp.Addr, resp.IfIdx, nil
		}
		m.mu.Unlock()
		// Spurious wake; treat as timeout
		return nil, nil, 0, ctx.Err()
	}
}

// Close marks the transport as closed.
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// EnableBlockingReceive switches Receive() into blocking mode where it waits
// for a queued response or context cancellation instead of returning immediately.
func (m *MockTransport) EnableBlockingReceive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blockOnReceive = true
}

// QueueReceive adds a prepared response to the receive queue and enables
// blocking mode so that Receive() will wait for responses instead of
// returning nil immediately.
func (m *MockTransport) QueueReceive(packet []byte, addr net.Addr, ifIdx int) {
	m.mu.Lock()
	m.blockOnReceive = true
	m.receiveQueue = append(m.receiveQueue, mockReceiveResponse{
		Packet: append([]byte(nil), packet...), // Copy to avoid aliasing
		Addr:   addr,
		IfIdx:  ifIdx,
	})
	m.mu.Unlock()
	// Notify any blocked Receive() call
	select {
	case m.receiveNotifyCh <- struct{}{}:
	default:
	}
}

// SendCalls returns all recorded Send() calls.
//
// This allows tests to verify:
// - Number of Send() calls
// - Packet contents
// - Destination addresses
//
// T017: Verification helper for tests
func (m *MockTransport) SendCalls() []SendCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to avoid race conditions
	calls := make([]SendCall, len(m.sendCalls))
	copy(calls, m.sendCalls)
	return calls
}
