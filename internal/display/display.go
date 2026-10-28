// Package display defines the LED output port used by the application.
package display

import "time"

// Panel is an LED output adapter. Pin numbers are GPIO pin numbers.
type Panel interface {
	Set(pin uint8, on bool) error
	Pulse(pin uint8, duration time.Duration) error
	Close() error
}

// Noop is a Panel that discards all events. It is used when no display is
// configured, so callers never have to check for nil.
type Noop struct{}

// Set implements Panel.
func (Noop) Set(uint8, bool) error { return nil }

// Pulse implements Panel.
func (Noop) Pulse(uint8, time.Duration) error { return nil }

// Close implements Panel.
func (Noop) Close() error { return nil }
