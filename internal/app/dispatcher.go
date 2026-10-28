package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/display"
	"github.com/mrechtien/mixgo/internal/input"
	"github.com/mrechtien/mixgo/internal/mixer"
	"github.com/mrechtien/mixgo/internal/tempo"
)

// Dispatcher maps input events to mixer commands and LED updates.
type Dispatcher struct {
	log      *slog.Logger
	mixer    mixer.Mixer
	panel    display.Panel
	bindings map[input.Key]config.Binding

	detectors map[input.Key]*tempo.Detector
	blinkers  map[input.Key]*blinker
	muted     map[mixer.Target]bool

	mu sync.Mutex
}

// NewDispatcher creates a dispatcher for the given bindings.
func NewDispatcher(cfg config.Config, m mixer.Mixer, panel display.Panel, log *slog.Logger) (*Dispatcher, error) {
	if log == nil {
		log = slog.Default()
	}
	d := &Dispatcher{
		log:       log.With("component", "dispatcher"),
		mixer:     m,
		panel:     panel,
		bindings:  make(map[input.Key]config.Binding, len(cfg.Bindings)),
		detectors: make(map[input.Key]*tempo.Detector),
		blinkers:  make(map[input.Key]*blinker),
		muted:     make(map[mixer.Target]bool),
	}

	limits := m.TapDelayLimits()
	for _, binding := range cfg.Bindings {
		key := input.Key{Channel: cfg.Input.Channel, Control: binding.Control}
		if _, exists := d.bindings[key]; exists {
			return nil, fmt.Errorf("duplicate binding for control %d", binding.Control)
		}
		d.bindings[key] = binding

		if binding.Action != config.ActionTapDelay {
			continue
		}
		d.detectors[key] = tempo.New(limits.Max, limits.MaxTaps)
		if binding.LED != nil {
			d.blinkers[key] = newBlinker(panel, *binding.LED, d.log)
		}
	}
	return d, nil
}

// Handle processes a single input event. Events are processed serially to keep
// mute state and tap tempo consistent.
func (d *Dispatcher) Handle(ctx context.Context, ev input.Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := input.Key{Channel: ev.Channel, Control: ev.Control}
	binding, ok := d.bindings[key]
	if !ok {
		d.log.Debug("unmapped input", "channel", ev.Channel, "control", ev.Control, "value", ev.Value)
		return
	}

	switch binding.Action {
	case config.ActionMuteGroup, config.ActionMuteChannel:
		d.handleMute(ctx, binding, ev)
	case config.ActionTapDelay:
		d.handleTap(ctx, binding, key, ev)
	default:
		d.log.Warn("unsupported action", "action", binding.Action)
	}
}

func (d *Dispatcher) handleMute(ctx context.Context, binding config.Binding, ev input.Event) {
	target := mixer.Target{Number: binding.Target}
	if binding.Action == config.ActionMuteGroup {
		target.Kind = mixer.KindMuteGroup
	} else {
		target.Kind = mixer.KindMuteChannel
	}

	pressed := ev.Value == binding.OnValue
	var desired bool
	switch binding.Mode {
	case config.ModeToggle:
		if !pressed {
			return
		}
		desired = !d.muted[target]
	case config.ModeMomentary:
		desired = pressed
	default:
		desired = pressed
	}

	if err := d.mixer.SetMuted(ctx, target, desired); err != nil {
		d.log.Warn("failed to set mute", "target", target.Number, "kind", target.Kind, "muted", desired, "error", err)
		return
	}
	d.muted[target] = desired

	if binding.LED != nil {
		if err := d.panel.Set(*binding.LED, desired); err != nil {
			d.log.Warn("failed to set led", "pin", *binding.LED, "error", err)
		}
	}
}

func (d *Dispatcher) handleTap(ctx context.Context, binding config.Binding, key input.Key, ev input.Event) {
	if ev.Value != binding.OnValue {
		return
	}

	if binding.LED != nil {
		if err := d.panel.Pulse(*binding.LED, 100*time.Millisecond); err != nil {
			d.log.Warn("failed to pulse led", "pin", *binding.LED, "error", err)
		}
	}

	detector := d.detectors[key]
	if detector == nil {
		return
	}
	interval, ready := detector.Tap(ev.At)

	// Always forward the tap: CQ derives the tempo from repeated soft key
	// events and ignores the tempo argument.
	if err := d.mixer.SetTapTempo(ctx, binding.Target, interval); err != nil {
		d.log.Warn("failed to set tap tempo", "fx", binding.Target, "error", err)
		return
	}

	if ready && interval > 0 {
		if blinker := d.blinkers[key]; blinker != nil {
			blinker.Start(interval)
		}
	}
}

// Close stops all LED blinkers.
func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, blinker := range d.blinkers {
		blinker.Stop()
	}
}
