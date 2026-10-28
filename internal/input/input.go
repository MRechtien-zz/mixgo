// Package input defines the control input port used by the application.
package input

import (
	"context"
	"time"
)

// Key identifies a control on a channel. Channels are zero based.
type Key struct {
	Channel uint8
	Control uint8
}

// Event is a raw control change emitted by an input source.
type Event struct {
	Key
	Value uint8
	At    time.Time
}

// Source is a control input adapter. Run blocks until ctx is cancelled or the
// source fails. Implementations must return nil after a context cancellation.
type Source interface {
	Run(ctx context.Context, emit func(Event)) error
	Close() error
}
