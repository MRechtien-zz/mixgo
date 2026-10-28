package xr

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hypebeast/go-osc/osc"

	"github.com/mrechtien/mixgo/internal/mixer"
)

type fakeSender struct {
	messages []*osc.Message
	err      error
}

func (f *fakeSender) Send(_ context.Context, msg *osc.Message) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, msg)
	return nil
}

func (f *fakeSender) Close() error { return nil }

func TestSetMuted(t *testing.T) {
	sender := &fakeSender{}
	m := newWithSender(sender, nil)

	if err := m.SetMuted(context.Background(), mixer.Target{Kind: mixer.KindMuteGroup, Number: 2}, true); err != nil {
		t.Fatalf("SetMuted() error = %v", err)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(sender.messages))
	}
	msg := sender.messages[0]
	if msg.Address != "/config/mute/2" {
		t.Errorf("address = %q, want /config/mute/2", msg.Address)
	}
	if len(msg.Arguments) != 1 || msg.Arguments[0] != int32(1) {
		t.Errorf("arguments = %#v, want [int32(1)]", msg.Arguments)
	}
}

func TestSetMutedChannelUnsupported(t *testing.T) {
	m := newWithSender(&fakeSender{}, nil)
	err := m.SetMuted(context.Background(), mixer.Target{Kind: mixer.KindMuteChannel, Number: 1}, true)
	if !errors.Is(err, mixer.ErrUnsupported) {
		t.Fatalf("error = %v, want ErrUnsupported", err)
	}
}

func TestSetTapTempo(t *testing.T) {
	sender := &fakeSender{}
	m := newWithSender(sender, nil)

	if err := m.SetTapTempo(context.Background(), 2, 0); err != nil {
		t.Fatalf("SetTapTempo(0) error = %v", err)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("zero tempo must not send a message, got %v", sender.messages)
	}

	if err := m.SetTapTempo(context.Background(), 2, 1500*time.Millisecond); err != nil {
		t.Fatalf("SetTapTempo() error = %v", err)
	}
	msg := sender.messages[0]
	if msg.Address != "/fx/2/par/01" {
		t.Errorf("address = %q, want /fx/2/par/01", msg.Address)
	}
	if got := msg.Arguments[0].(float32); got != 0.5 {
		t.Errorf("tempo = %v, want 0.5", got)
	}
}

func TestNormalizeTempo(t *testing.T) {
	tests := []struct {
		tempo time.Duration
		want  float32
	}{
		{10 * time.Millisecond, float32(minDelayMillis) / float32(maxDelayMillis)},
		{1500 * time.Millisecond, 0.5},
		{5 * time.Second, 1},
	}
	for _, tt := range tests {
		if got := normalizeTempo(tt.tempo); got != tt.want {
			t.Errorf("normalizeTempo(%s) = %v, want %v", tt.tempo, got, tt.want)
		}
	}
}
