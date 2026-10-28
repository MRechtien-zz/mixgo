package transport

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/hypebeast/go-osc/osc"
)

// OSC sends OSC messages over UDP.
type OSC struct {
	client *osc.Client

	mu     sync.Mutex
	closed bool
}

// NewOSC creates an OSC sender for addr (host:port).
func NewOSC(addr string) (*OSC, error) {
	host, portRaw, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid OSC address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid OSC port %q", portRaw)
	}
	return &OSC{client: osc.NewClient(host, port)}, nil
}

// Send sends an OSC message. UDP sends are not cancellable, so ctx is only
// checked before sending.
func (o *OSC) Send(ctx context.Context, msg *osc.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return ErrClosed
	}
	if err := o.client.Send(msg); err != nil {
		return fmt.Errorf("send OSC message: %w", err)
	}
	return nil
}

// Close marks the sender as closed. UDP has no connection to close.
func (o *OSC) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.closed = true
	return nil
}
