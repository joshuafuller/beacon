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
	mu        sync.Mutex
	sendCalls []SendCall
	closed    bool
}

// SendCall records a single Send() invocation.
type SendCall struct {
	Packet []byte
	Dest   net.Addr
}

// NewMockTransport creates a new mock transport for testing.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		sendCalls: make([]SendCall, 0),
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

// Receive is not implemented in mock (querier doesn't use it in current tests).
//
// 007-interface-specific-addressing: Updated to return interfaceIndex (T012-T013)
// Future: Can be extended to return pre-configured responses.
func (m *MockTransport) Receive(_ context.Context) ([]byte, net.Addr, int, error) {
	// Not needed for current tests
	// Return 0 for interfaceIndex (unknown/mock interface)
	return nil, nil, 0, nil
}

// Close marks the transport as closed.
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
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
