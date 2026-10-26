package querier

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/joshuafuller/beacon/internal/transport"
)

// RFC 6763 §4.1.1: instance MUST NOT contain ASCII control characters.
func TestReview_QuerySRVControlChars(t *testing.T) {
	mt := transport.NewMockTransport()
	q := &Querier{transport: mt, responseChan: make(chan []byte, 1), defaultTimeout: 10 * time.Millisecond}
	_, err := q.Query(context.Background(), "evil\x00\x1b[2J._http._tcp.local", RecordTypeSRV)
	calls := mt.SendCalls()
	t.Logf("err=%v sends=%d", err, len(calls))
	if err == nil && len(calls) == 1 && bytes.Contains(calls[0].Packet, []byte{0x00, 0x1b}) {
		t.Errorf("Query put NUL/ESC control bytes on the wire: % x", calls[0].Packet[12:32])
	}
}

// Query after Close must return, not spin on the closed response channel.
func TestReview_QueryAfterCloseReturns(t *testing.T) {
	mt := transport.NewMockTransport()
	mt.EnableBlockingReceive()
	q, err := New(WithTransport(mt))
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, _ = q.Query(ctx, "printer.local", RecordTypeA)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Query after Close did not return within 1s")
	}
}
