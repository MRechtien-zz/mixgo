package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/display"
	"github.com/mrechtien/mixgo/internal/input"
)

type fakeSource struct {
	run func(ctx context.Context, emit func(input.Event)) error
}

func (f *fakeSource) Run(ctx context.Context, emit func(input.Event)) error {
	return f.run(ctx, emit)
}
func (f *fakeSource) Close() error { return nil }

func TestSelectProfileWithMatch(t *testing.T) {
	led := uint8(26)
	boot := config.Bootstrap{
		Input:   config.Input{Type: "gpio", Channel: 0},
		Timeout: config.Duration(time.Second),
		Profiles: []config.Profile{{
			Control: 17,
			OnValue: 1,
			LED:     &led,
			Config:  "runtime.yml",
		}},
	}
	src := &fakeSource{run: func(_ context.Context, emit func(input.Event)) error {
		emit(input.Event{Key: input.Key{Control: 17}, Value: 0})
		emit(input.Event{Key: input.Key{Control: 17}, Value: 1})
		return nil
	}}
	panel := &fakePanel{}

	got, err := selectProfileWith(context.Background(), boot, src, panel, testLogger())
	if err != nil {
		t.Fatalf("selectProfileWith() error = %v", err)
	}
	if got != "runtime.yml" {
		t.Fatalf("config = %q, want runtime.yml", got)
	}
	if panel.pulses != 1 {
		t.Fatalf("pulses = %d, want 1", panel.pulses)
	}
}

func TestSelectProfileWithDefaultFallback(t *testing.T) {
	boot := config.Bootstrap{
		Input:    config.Input{Type: "gpio", Channel: 0},
		Timeout:  config.Duration(10 * time.Millisecond),
		Default:  "default.yml",
		Profiles: []config.Profile{{Control: 17, OnValue: 1, Config: "runtime.yml"}},
	}
	src := &fakeSource{run: func(ctx context.Context, _ func(input.Event)) error {
		<-ctx.Done()
		return nil
	}}

	got, err := selectProfileWith(context.Background(), boot, src, display.Noop{}, testLogger())
	if err != nil {
		t.Fatalf("selectProfileWith() error = %v", err)
	}
	if got != "default.yml" {
		t.Fatalf("config = %q, want default.yml", got)
	}
}

func TestSelectProfileWithTimeout(t *testing.T) {
	boot := config.Bootstrap{
		Input:    config.Input{Type: "gpio", Channel: 0},
		Timeout:  config.Duration(10 * time.Millisecond),
		Profiles: []config.Profile{{Control: 17, OnValue: 1, Config: "runtime.yml"}},
	}
	src := &fakeSource{run: func(ctx context.Context, _ func(input.Event)) error {
		<-ctx.Done()
		return nil
	}}

	_, err := selectProfileWith(context.Background(), boot, src, display.Noop{}, testLogger())
	if err == nil || !strings.Contains(err.Error(), "no profile selected") {
		t.Fatalf("error = %v, want no profile selected", err)
	}
}

func TestSelectProfileWithParentCancellation(t *testing.T) {
	boot := config.Bootstrap{
		Input:    config.Input{Type: "gpio", Channel: 0},
		Timeout:  config.Duration(time.Second),
		Default:  "default.yml",
		Profiles: []config.Profile{{Control: 17, OnValue: 1, Config: "runtime.yml"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := &fakeSource{run: func(ctx context.Context, _ func(input.Event)) error {
		<-ctx.Done()
		return nil
	}}

	_, err := selectProfileWith(ctx, boot, src, display.Noop{}, testLogger())
	if err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
