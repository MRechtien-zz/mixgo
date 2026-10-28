// Package transport contains shared network transports for mixer adapters.
package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const writeTimeout = 2 * time.Second

// ErrClosed is returned when a closed sender is used.
var ErrClosed = errors.New("transport is closed")

// Sender sends raw protocol payloads to a mixer.
type Sender interface {
	Send(ctx context.Context, payload []byte) error
	Close() error
}

// TCP keeps a single persistent connection and reconnects on demand.
type TCP struct {
	addr   string
	dialer net.Dialer

	mu     sync.Mutex
	conn   net.Conn
	closed bool
}

// NewTCP creates a TCP sender for addr (host:port).
func NewTCP(addr string) (*TCP, error) {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return nil, fmt.Errorf("invalid TCP address %q: %w", addr, err)
	}
	return &TCP{
		addr:   addr,
		dialer: net.Dialer{Timeout: 5 * time.Second},
	}, nil
}

// Send writes payload, reconnecting once if the existing connection broke.
func (t *TCP) Send(ctx context.Context, payload []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}
	err := t.sendLocked(ctx, payload)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return err
	}

	// The console may have closed the connection; reconnect and retry once.
	t.closeConnLocked()
	return t.sendLocked(ctx, payload)
}

func (t *TCP) sendLocked(ctx context.Context, payload []byte) error {
	if t.conn == nil {
		conn, err := t.dialer.DialContext(ctx, "tcp", t.addr)
		if err != nil {
			return fmt.Errorf("dial %s: %w", t.addr, err)
		}
		t.conn = conn
	}

	deadline := time.Now().Add(writeTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = t.conn.SetWriteDeadline(deadline)
	if _, err := t.conn.Write(payload); err != nil {
		t.closeConnLocked()
		return fmt.Errorf("write %s: %w", t.addr, err)
	}
	return nil
}

// Close closes the underlying connection. It is safe to call multiple times.
func (t *TCP) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	return t.closeConnLocked()
}

func (t *TCP) closeConnLocked() error {
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	return err
}
