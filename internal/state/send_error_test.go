package state

import (
	"context"
	"errors"
	"net"
	"testing"
)

type failingSendTransport struct{ err error }

func (t *failingSendTransport) Send(context.Context, []byte, net.Addr) error {
	return t.err
}

func (*failingSendTransport) Receive(context.Context) ([]byte, net.Addr, int, error) {
	return nil, nil, 0, errors.New("unused")
}

func (*failingSendTransport) Close() error { return nil }

// RFC 6762 §8.1: Registration cannot complete probing when a mandatory probe
// query could not be transmitted.
func TestProberPropagatesTransportSendError(t *testing.T) {
	prober := NewProber()
	prober.SetTransport(&failingSendTransport{err: errors.New("send failed")})
	result := prober.Probe(context.Background(), testServiceName)
	if result.Error == nil {
		t.Fatal("Probe() succeeded after its transport rejected the probe")
	}
}

// RFC 6762 §8.3: Registration cannot become established when a mandatory
// unsolicited announcement could not be transmitted.
func TestAnnouncerPropagatesTransportSendError(t *testing.T) {
	announcer := NewAnnouncer()
	announcer.SetTransport(&failingSendTransport{err: errors.New("send failed")})
	if err := announcer.Announce(context.Background(), testServiceName, nil); err == nil {
		t.Fatal("Announce() succeeded after its transport rejected the announcement")
	}
}
