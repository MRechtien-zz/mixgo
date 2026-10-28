// Package gpio implements an LED display panel using periph.io.
package gpio

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"

	"github.com/mrechtien/mixgo/internal/display"
)

var initHost sync.Once
var initErr error

// Panel controls LEDs connected to GPIO pins.
type Panel struct {
	log *slog.Logger

	mu     sync.Mutex
	pins   map[uint8]gpio.PinIO
	timers map[uint8]*time.Timer
	closed bool
}

// New creates a GPIO LED panel.
func New(log *slog.Logger) (*Panel, error) {
	if log == nil {
		log = slog.Default()
	}
	initHost.Do(func() { _, initErr = host.Init() })
	if initErr != nil {
		return nil, fmt.Errorf("initialise gpio host: %w", initErr)
	}
	return &Panel{
		log:    log.With("component", "display/gpio"),
		pins:   map[uint8]gpio.PinIO{},
		timers: map[uint8]*time.Timer{},
	}, nil
}

// Set switches an LED on or off.
func (p *Panel) Set(pin uint8, on bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	if timer, ok := p.timers[pin]; ok {
		timer.Stop()
		delete(p.timers, pin)
	}
	return p.setLocked(pin, on)
}

// Pulse switches an LED on and off again after duration.
func (p *Panel) Pulse(pin uint8, duration time.Duration) error {
	if duration <= 0 {
		duration = 100 * time.Millisecond
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	if err := p.setLocked(pin, true); err != nil {
		return err
	}
	if timer, ok := p.timers[pin]; ok {
		timer.Stop()
	}
	p.timers[pin] = time.AfterFunc(duration, func() {
		_ = p.Set(pin, false)
	})
	return nil
}

// Close stops pending pulses and switches all LEDs off.
func (p *Panel) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	for _, timer := range p.timers {
		timer.Stop()
	}
	for _, gpioPin := range p.pins {
		_ = gpioPin.Out(gpio.Level(false))
	}
	p.timers = map[uint8]*time.Timer{}
	return nil
}

func (p *Panel) setLocked(pin uint8, on bool) error {
	gpioPin, err := p.pinLocked(pin)
	if err != nil {
		return err
	}
	if err := gpioPin.Out(gpio.Level(on)); err != nil {
		return fmt.Errorf("set gpio pin %d: %w", pin, err)
	}
	return nil
}

func (p *Panel) pinLocked(pin uint8) (gpio.PinIO, error) {
	if gpioPin, ok := p.pins[pin]; ok {
		return gpioPin, nil
	}
	gpioPin := gpioreg.ByName(strconv.Itoa(int(pin)))
	if gpioPin == nil {
		return nil, fmt.Errorf("gpio pin %d not found", pin)
	}
	p.pins[pin] = gpioPin
	p.log.Debug("using gpio pin for LED", "pin", gpioPin)
	return gpioPin, nil
}

var _ display.Panel = (*Panel)(nil)
