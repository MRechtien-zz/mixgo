// Package xr implements the Behringer X Air mixer protocol over OSC.
package xr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hypebeast/go-osc/osc"

	"github.com/mrechtien/mixgo/internal/mixer"
	"github.com/mrechtien/mixgo/internal/transport"
)

const (
	Name = "xr"

	minDelayMillis = 50
	maxDelayMillis = 3000
)

// Mixer controls a Behringer X Air console over OSC.
type Mixer struct {
	sender oscSender
	log    *slog.Logger
}

type oscSender interface {
	Send(ctx context.Context, msg *osc.Message) error
	Close() error
}

// New creates an XR mixer for addr (host:port).
func New(addr string, log *slog.Logger) (*Mixer, error) {
	sender, err := transport.NewOSC(addr)
	if err != nil {
		return nil, err
	}
	return newWithSender(sender, log), nil
}

func newWithSender(sender oscSender, log *slog.Logger) *Mixer {
	if log == nil {
		log = slog.Default()
	}
	return &Mixer{sender: sender, log: log.With("component", "mixer/xr")}
}

// Name returns the mixer type.
func (m *Mixer) Name() string { return Name }

// Capabilities reports supported features.
func (m *Mixer) Capabilities() mixer.Capabilities {
	return mixer.Capabilities{MuteGroup: true, TapDelay: true}
}

// TapDelayLimits returns the supported tap tempo range.
func (m *Mixer) TapDelayLimits() mixer.TapDelayLimits {
	return mixer.TapDelayLimits{
		Min:     minDelayMillis * time.Millisecond,
		Max:     maxDelayMillis * time.Millisecond,
		MaxTaps: 3,
	}
}

// SetMuted sets a mute group. Mute channels are not supported by XR.
func (m *Mixer) SetMuted(ctx context.Context, target mixer.Target, muted bool) error {
	if target.Kind != mixer.KindMuteGroup {
		return mixer.ErrUnsupported
	}
	value := int32(0)
	if muted {
		value = 1
	}
	m.log.Debug("set mute group", "group", target.Number, "muted", muted)
	msg := osc.NewMessage(fmt.Sprintf("/config/mute/%d", target.Number), value)
	return m.sender.Send(ctx, msg)
}

// SetTapTempo sends the normalized delay time to an FX slot.
func (m *Mixer) SetTapTempo(ctx context.Context, fx uint8, tempo time.Duration) error {
	if tempo <= 0 {
		return nil
	}
	normalized := normalizeTempo(tempo)
	m.log.Debug("set tap tempo", "fx", fx, "tempo", tempo, "normalized", normalized)
	msg := osc.NewMessage(fmt.Sprintf("/fx/%d/par/01", fx), normalized)
	return m.sender.Send(ctx, msg)
}

// Close closes the transport.
func (m *Mixer) Close() error { return m.sender.Close() }

func normalizeTempo(tempo time.Duration) float32 {
	millis := tempo.Milliseconds()
	switch {
	case millis > maxDelayMillis:
		return 1
	case millis < minDelayMillis:
		return float32(minDelayMillis) / float32(maxDelayMillis)
	default:
		return float32(millis) / float32(maxDelayMillis)
	}
}
