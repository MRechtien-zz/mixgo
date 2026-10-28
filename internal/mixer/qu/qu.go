// Package qu implements the Allen & Heath QU mixer protocol.
package qu

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/mrechtien/mixgo/internal/midi/nrpn"
	"github.com/mrechtien/mixgo/internal/mixer"
	"github.com/mrechtien/mixgo/internal/transport"
)

const (
	Name = "qu"

	minDelayMillis = 5
	maxDelayMillis = 1360

	muteGroupsStart = 0x50
	muteOn          = 0x40
	muteOff         = 0x10
)

// Mixer controls an Allen & Heath QU console over MIDI/TCP.
type Mixer struct {
	sender transport.Sender
	log    *slog.Logger
}

// New creates a QU mixer for addr (host:port).
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
	return &Mixer{sender: sender, log: log.With("component", "mixer/qu")}
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

// SetMuted sets a mute group. Mute channels are not supported by QU.
func (m *Mixer) SetMuted(ctx context.Context, target mixer.Target, muted bool) error {
	switch target.Kind {
	case mixer.KindMuteGroup:
		if target.Number > 0x2F {
			return fmt.Errorf("mute group %d out of range for QU", target.Number)
		}
		m.log.Debug("set mute group", "group", target.Number, "muted", muted)
		return m.sender.Send(ctx, muteGroupMessage(target.Number, muted))
	case mixer.KindMuteChannel:
		return mixer.ErrUnsupported
	default:
		return mixer.ErrUnsupported
	}
}

// SetTapTempo converts tempo to the QU delay parameters and sends them. A zero
// tempo means that no average is available yet and is ignored.
func (m *Mixer) SetTapTempo(ctx context.Context, _ uint8, tempo time.Duration) error {
	if tempo <= 0 {
		return nil
	}
	coarse, fine := delayValues(tempo)
	m.log.Debug("set tap tempo", "tempo", tempo, "coarse", coarse, "fine", fine)
	message := append(
		nrpn.Message(0, 0x49, 0x05, fine, 0x05),
		nrpn.Message(0, 0x48, 0x05, coarse, 0x05)...,
	)
	return m.sender.Send(ctx, message)
}

// Close closes the transport.
func (m *Mixer) Close() error { return m.sender.Close() }

func muteGroupMessage(group uint8, muted bool) []byte {
	value := byte(muteOff)
	if muted {
		value = muteOn
	}
	channel := byte(muteGroupsStart) + group
	return []byte{0x90, channel, 0x7F, 0x90, channel, value}
}

func delayValues(tempo time.Duration) (coarse, fine uint8) {
	millis := int(tempo.Milliseconds())
	if millis <= minDelayMillis {
		// Limit to the default minimum of about 69 ms.
		return 0x3C, 0x00
	}
	if millis >= maxDelayMillis {
		return 0x7F, 0x7F
	}

	// According to the Allen & Heath MIDI specification (30 June 2014).
	value := math.Round(16383 * (math.Log10(float64(millis)) - math.Log10(5)) / 2.4346)
	coarse = uint8(math.Round(math.Floor(value / 128)))
	fine = uint8(math.Round(math.Mod(value, 128)))
	return coarse, fine
}
