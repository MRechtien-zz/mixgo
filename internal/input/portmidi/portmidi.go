// Package portmidi implements a MIDI input source using PortMidi.
//
// The adapter is only compiled when the "portmidi" build tag is set and cgo is
// available:
//
//	go build -tags portmidi ./cmd/mixgo
package portmidi

import "errors"

// Options configures the PortMidi input source.
type Options struct {
	Device  string
	Channel uint8
}

// ErrNotBuilt is returned when MixGo was built without PortMidi support.
var ErrNotBuilt = errors.New("portmidi support is not compiled in (build with -tags portmidi)")
