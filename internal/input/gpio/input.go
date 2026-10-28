// Package gpio implements a GPIO button input source using periph.io.
package gpio

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"

	"github.com/mrechtien/mixgo/internal/input"
)

// Options configures the GPIO input source.
type Options struct {
	Channel  uint8
	Controls []uint8
}

// Source listens to GPIO pins and emits press/release events.
type Source struct {
	opts Options
	log  *slog.Logger
}

var initHost sync.Once
var initErr error

// New creates a GPIO input source.
func New(opts Options, log *slog.Logger) (*Source, error) {
	if len(opts.Controls) == 0 {
		return nil, fmt.Errorf("gpio input requires at least one control")
	}
	if log == nil {
		log = slog.Default()
	}
	return &Source{opts: opts, log: log.With("component", "input/gpio")}, nil
}

// Run listens to all configured pins until ctx is cancelled.
func (s *Source) Run(ctx context.Context, emit func(input.Event)) error {
	initHost.Do(func() { _, initErr = host.Init() })
	if initErr != nil {
		return fmt.Errorf("initialise gpio host: %w", initErr)
	}

	controls := unique(s.opts.Controls)
	pins := make([]gpio.PinIO, 0, len(controls))
	for _, control := range controls {
		pin := gpioreg.ByName(strconv.Itoa(int(control)))
		if pin == nil {
			haltAll(pins)
			return fmt.Errorf("gpio pin %d not found", control)
		}
		if err := pin.In(gpio.PullDown, gpio.BothEdges); err != nil {
			haltAll(pins)
			return fmt.Errorf("configure gpio pin %d: %w", control, err)
		}
		pins = append(pins, pin)
		s.log.Debug("listening to gpio pin", "pin", pin, "control", control)
	}
	defer haltAll(pins)

	var wg sync.WaitGroup
	for i, control := range controls {
		wg.Add(1)
		go func(pin gpio.PinIO, control uint8) {
			defer wg.Done()
			s.listen(ctx, pin, control, emit)
		}(pins[i], control)
	}
	wg.Wait()
	return nil
}

func (s *Source) listen(ctx context.Context, pin gpio.PinIO, control uint8, emit func(input.Event)) {
	last := gpio.Level(false) // pull-down
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// A finite timeout makes it possible to notice a cancelled context.
		if !pin.WaitForEdge(100 * time.Millisecond) {
			continue
		}
		state := pin.Read()
		if state == last {
			continue
		}
		// Simple debounce for mechanical contacts.
		time.Sleep(5 * time.Millisecond)
		stable := pin.Read()
		if stable != state {
			last = stable
			continue
		}
		last = state

		value := uint8(0)
		if state {
			value = 1
		}
		emit(input.Event{
			Key:   input.Key{Channel: s.opts.Channel, Control: control},
			Value: value,
			At:    time.Now(),
		})
	}
}

// Close is a no-op. Pins are halted when Run returns.
func (s *Source) Close() error { return nil }

func unique(values []uint8) []uint8 {
	set := map[uint8]struct{}{}
	for _, v := range values {
		set[v] = struct{}{}
	}
	out := make([]uint8, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func haltAll(pins []gpio.PinIO) {
	for _, pin := range pins {
		if pin != nil {
			pin.Halt()
		}
	}
}
