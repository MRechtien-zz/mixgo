// Package mixer defines the mixer port used by the application.
package mixer

import (
	"context"
	"errors"
	"time"
)

// Kind identifies a controllable mixer feature.
type Kind uint8

const (
	KindMuteGroup Kind = iota
	KindMuteChannel
	KindTapDelay
)

// String implements fmt.Stringer.
func (k Kind) String() string {
	switch k {
	case KindMuteGroup:
		return "mute_group"
	case KindMuteChannel:
		return "mute_channel"
	case KindTapDelay:
		return "tap_delay"
	default:
		return "unknown"
	}
}

// Target selects a concrete mixer element.
type Target struct {
	Kind   Kind
	Number uint8
}

// Capabilities describes which features a mixer implementation supports.
type Capabilities struct {
	MuteGroup   bool
	MuteChannel bool
	TapDelay    bool
}

// Supports reports whether the capability set contains k.
func (c Capabilities) Supports(k Kind) bool {
	switch k {
	case KindMuteGroup:
		return c.MuteGroup
	case KindMuteChannel:
		return c.MuteChannel
	case KindTapDelay:
		return c.TapDelay
	default:
		return false
	}
}

// TapDelayLimits describes the tap tempo range of a mixer.
type TapDelayLimits struct {
	Min     time.Duration
	Max     time.Duration
	MaxTaps int
}

// ErrUnsupported is returned when an operation is not supported by a mixer.
var ErrUnsupported = errors.New("operation not supported by mixer")

// Mixer is the mixer port. Implementations encode protocol messages and send
// them through a transport.
type Mixer interface {
	Name() string
	Capabilities() Capabilities
	TapDelayLimits() TapDelayLimits
	// SetMuted sets the mute state of a mute group or a mute channel.
	SetMuted(ctx context.Context, target Target, muted bool) error
	// SetTapTempo sends a tap event. A zero tempo means that the tap detector
	// does not have enough data yet; mixers that need a concrete tempo ignore it.
	SetTapTempo(ctx context.Context, fx uint8, tempo time.Duration) error
	Close() error
}
