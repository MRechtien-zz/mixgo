package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mrechtien/mixgo/internal/display"
)

// blinker repeatedly pulses an LED at a given interval. A new Start replaces
// the previously running blink interval.
type blinker struct {
	panel display.Panel
	pin   uint8
	log   *slog.Logger

	mu     sync.Mutex
	cancel context.CancelFunc
}

func newBlinker(panel display.Panel, pin uint8, log *slog.Logger) *blinker {
	return &blinker{panel: panel, pin: pin, log: log}
}

func (b *blinker) Start(interval time.Duration) {
	if interval <= 0 {
		return
	}

	b.mu.Lock()
	if b.cancel != nil {
		b.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	b.mu.Unlock()

	pulse := interval / 4
	if pulse < 20*time.Millisecond {
		pulse = 20 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := b.panel.Pulse(b.pin, pulse); err != nil {
					b.log.Warn("failed to pulse led", "pin", b.pin, "error", err)
				}
			}
		}
	}()
}

func (b *blinker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
}
