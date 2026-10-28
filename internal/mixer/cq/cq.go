// Package cq implements the Allen & Heath CQ mixer protocol.
package cq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mrechtien/mixgo/internal/midi/nrpn"
	"github.com/mrechtien/mixgo/internal/mixer"
	"github.com/mrechtien/mixgo/internal/transport"
)

const Name = "cq"

const (
	muteGroupMSB   = 0x04
	muteChannelMSB = 0x00

	maxDelayMillis = 2400
)

// Mixer controls an Allen & Heath CQ console over MIDI/TCP.
type Mixer struct {
	sender transport.Sender
	log    *slog.Logger
}

// New creates a CQ mixer for addr (host:port).
func New(addr string, log *slog.Logger) (*Mixer, error) {
	sender, err := transport.NewTCP(addr)
	if err != nil {
		return nil, err
	}
	return newWithSender(sender, log), nil
}

func newWithSender(sender transport.Sender, log *slog.Logger) *Mixer {
	if log == nil {
		log = slog.Default()
	}
	return &Mixer{sender: sender, log: log.With("component", "mixer/cq")}
}

// Name returns the mixer type.
func (m *Mixer) Name() string { return Name }

// Capabilities reports supported features.
func (m *Mixer) Capabilities() mixer.Capabilities {
	return mixer.Capabilities{MuteGroup: true, MuteChannel: true, TapDelay: true}
}

// TapDelayLimits returns the supported tap tempo range.
func (m *Mixer) TapDelayLimits() mixer.TapDelayLimits {
	return mixer.TapDelayLimits{
		Min:     0,
		Max:     maxDelayMillis * time.Millisecond,
		MaxTaps: 1,
	}
}

// SetMuted sets a mute group or mute channel.
func (m *Mixer) SetMuted(ctx context.Context, target mixer.Target, muted bool) error {
	if target.Number > 127 {
		return fmt.Errorf("target %d out of range for CQ", target.Number)
	}
	value := uint8(0)
	if muted {
		value = 1
	}

	var msb uint8
	switch target.Kind {
	case mixer.KindMuteGroup:
		msb = muteGroupMSB
	case mixer.KindMuteChannel:
		msb = muteChannelMSB
	default:
		return mixer.ErrUnsupported
	}
	m.log.Debug("set mute", "kind", target.Kind, "target", target.Number, "muted", muted)
	return m.sender.Send(ctx, nrpn.Message(0, msb, target.Number, 0, value))
}

// SetTapTempo sends the soft key event that drives the CQ tap tempo. The CQ
// calculates the tempo itself, so tempo is not used here.
func (m *Mixer) SetTapTempo(ctx context.Context, _ uint8, _ time.Duration) error {
	m.log.Debug("send tap tempo soft key")
	return m.sender.Send(ctx, softKeyDownMessage())
}

// Close closes the transport.
func (m *Mixer) Close() error { return m.sender.Close() }

func softKeyDownMessage() []byte {
	// A soft key release is not required.
	return []byte{0x90, 0x32, 0x7F}
}
