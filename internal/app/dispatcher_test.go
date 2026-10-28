package app

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/input"
	"github.com/mrechtien/mixgo/internal/mixer"
)

type muteCall struct {
	target mixer.Target
	muted  bool
}

type tapCall struct {
	fx       uint8
	interval time.Duration
}

type fakeMixer struct {
	caps   mixer.Capabilities
	limits mixer.TapDelayLimits
	mutes  []muteCall
	taps   []tapCall
	err    error
}

func (f *fakeMixer) Name() string                         { return "fake" }
func (f *fakeMixer) Capabilities() mixer.Capabilities     { return f.caps }
func (f *fakeMixer) TapDelayLimits() mixer.TapDelayLimits { return f.limits }
func (f *fakeMixer) Close() error                         { return nil }

func (f *fakeMixer) SetMuted(_ context.Context, target mixer.Target, muted bool) error {
	if f.err != nil {
		return f.err
	}
	f.mutes = append(f.mutes, muteCall{target: target, muted: muted})
	return nil
}

func (f *fakeMixer) SetTapTempo(_ context.Context, fx uint8, interval time.Duration) error {
	if f.err != nil {
		return f.err
	}
	f.taps = append(f.taps, tapCall{fx: fx, interval: interval})
	return nil
}

type fakePanel struct {
	sets   map[uint8]bool
	pulses int
}

func (p *fakePanel) Set(pin uint8, on bool) error {
	if p.sets == nil {
		p.sets = map[uint8]bool{}
	}
	p.sets[pin] = on
	return nil
}
func (p *fakePanel) Pulse(uint8, time.Duration) error { p.pulses++; return nil }
func (p *fakePanel) Close() error                     { return nil }

func TestDispatcherToggle(t *testing.T) {
	led := uint8(26)
	cfg := config.Config{
		Input: config.Input{Type: "gpio", Channel: 0},
		Bindings: []config.Binding{{
			Action:  config.ActionMuteGroup,
			Control: 5,
			Target:  0,
			OnValue: 1,
			Mode:    config.ModeToggle,
			LED:     &led,
		}},
	}
	mix := &fakeMixer{caps: mixer.Capabilities{MuteGroup: true}}
	panel := &fakePanel{}
	d, err := NewDispatcher(cfg, mix, panel, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	press := input.Event{Key: input.Key{Channel: 0, Control: 5}, Value: 1, At: time.Now()}
	release := input.Event{Key: input.Key{Channel: 0, Control: 5}, Value: 0, At: time.Now()}

	d.Handle(context.Background(), press)
	if len(mix.mutes) != 1 || !mix.mutes[0].muted {
		t.Fatalf("after press mutes = %+v, want one muted call", mix.mutes)
	}
	if panel.sets[led] != true {
		t.Fatalf("led = %v, want true", panel.sets[led])
	}

	d.Handle(context.Background(), release)
	if len(mix.mutes) != 1 {
		t.Fatalf("release must be ignored in toggle mode, mutes = %+v", mix.mutes)
	}

	d.Handle(context.Background(), press)
	if len(mix.mutes) != 2 || mix.mutes[1].muted {
		t.Fatalf("after second press mutes = %+v, want second call unmuted", mix.mutes)
	}
	if panel.sets[led] != false {
		t.Fatalf("led = %v, want false", panel.sets[led])
	}
}

func TestDispatcherAbsolute(t *testing.T) {
	cfg := config.Config{
		Input: config.Input{Type: "portmidi", Channel: 2},
		Bindings: []config.Binding{{
			Action:  config.ActionMuteChannel,
			Control: 7,
			Target:  81,
			OnValue: 127,
			Mode:    config.ModeAbsolute,
		}},
	}
	mix := &fakeMixer{caps: mixer.Capabilities{MuteChannel: true}}
	d, err := NewDispatcher(cfg, mix, &fakePanel{}, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	d.Handle(context.Background(), input.Event{Key: input.Key{Channel: 2, Control: 7}, Value: 127})
	d.Handle(context.Background(), input.Event{Key: input.Key{Channel: 2, Control: 7}, Value: 0})
	if len(mix.mutes) != 2 {
		t.Fatalf("mutes = %+v, want 2 calls", mix.mutes)
	}
	if !mix.mutes[0].muted || mix.mutes[1].muted {
		t.Fatalf("mutes = %+v, want [true, false]", mix.mutes)
	}
	if mix.mutes[0].target.Kind != mixer.KindMuteChannel || mix.mutes[0].target.Number != 81 {
		t.Fatalf("target = %+v", mix.mutes[0].target)
	}
}

func TestDispatcherTapDelayUsesPressOnly(t *testing.T) {
	cfg := config.Config{
		Input: config.Input{Type: "gpio", Channel: 0},
		Bindings: []config.Binding{{
			Action:  config.ActionTapDelay,
			Control: 17,
			Target:  2,
			OnValue: 1,
		}},
	}
	mix := &fakeMixer{
		caps:   mixer.Capabilities{TapDelay: true},
		limits: mixer.TapDelayLimits{Max: 3 * time.Second, MaxTaps: 3},
	}
	d, err := NewDispatcher(cfg, mix, &fakePanel{}, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	d.Handle(context.Background(), input.Event{Key: input.Key{Control: 17}, Value: 1, At: base})
	d.Handle(context.Background(), input.Event{Key: input.Key{Control: 17}, Value: 0, At: base.Add(50 * time.Millisecond)})
	d.Handle(context.Background(), input.Event{Key: input.Key{Control: 17}, Value: 1, At: base.Add(100 * time.Millisecond)})

	if len(mix.taps) != 2 {
		t.Fatalf("taps = %+v, want 2 calls", mix.taps)
	}
	if mix.taps[0].interval != 0 {
		t.Errorf("first tap interval = %s, want 0", mix.taps[0].interval)
	}
	if mix.taps[1].interval != 100*time.Millisecond {
		t.Errorf("second tap interval = %s, want 100ms", mix.taps[1].interval)
	}
	if mix.taps[1].fx != 2 {
		t.Errorf("fx = %d, want 2", mix.taps[1].fx)
	}
}

func TestDispatcherUnmappedEvent(t *testing.T) {
	cfg := config.Config{
		Input: config.Input{Type: "gpio", Channel: 0},
		Bindings: []config.Binding{{
			Action:  config.ActionMuteGroup,
			Control: 1,
			Target:  1,
			OnValue: 1,
		}},
	}
	mix := &fakeMixer{caps: mixer.Capabilities{MuteGroup: true}}
	d, err := NewDispatcher(cfg, mix, &fakePanel{}, testLogger())
	if err != nil {
		t.Fatal(err)
	}
	d.Handle(context.Background(), input.Event{Key: input.Key{Control: 99}, Value: 1})
	if len(mix.mutes) != 0 {
		t.Fatalf("unmapped event caused mixer calls: %+v", mix.mutes)
	}
}

func TestValidateCapabilities(t *testing.T) {
	bindings := []config.Binding{
		{Action: config.ActionMuteGroup},
		{Action: config.ActionTapDelay},
	}
	mix := &fakeMixer{caps: mixer.Capabilities{MuteGroup: true}}
	err := validateCapabilities(bindings, mix)
	if err == nil || !strings.Contains(err.Error(), "tap_delay") {
		t.Fatalf("error = %v, want unsupported tap_delay", err)
	}
}

func TestControlsForDeduplicates(t *testing.T) {
	controls := controlsFor([]config.Binding{
		{Control: 17},
		{Control: 17},
		{Control: 5},
	})
	if len(controls) != 2 || controls[0] != 17 || controls[1] != 5 {
		t.Fatalf("controls = %v, want [17 5]", controls)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
