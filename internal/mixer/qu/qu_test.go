package qu

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrechtien/mixgo/internal/midi/nrpn"
	"github.com/mrechtien/mixgo/internal/mixer"
)

type fakeSender struct {
	messages [][]byte
	err      error
}

func (f *fakeSender) Send(_ context.Context, payload []byte) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, append([]byte(nil), payload...))
	return nil
}

func (f *fakeSender) Close() error { return nil }

func TestSetMutedGroup(t *testing.T) {
	sender := &fakeSender{}
	m := newWithSender(sender, nil)

	if err := m.SetMuted(context.Background(), mixer.Target{Kind: mixer.KindMuteGroup, Number: 2}, true); err != nil {
		t.Fatalf("SetMuted() error = %v", err)
	}
	want := []byte{0x90, 0x52, 0x7F, 0x90, 0x52, 0x40}
	if len(sender.messages) != 1 || !bytes.Equal(sender.messages[0], want) {
		t.Fatalf("message = % X, want % X", sender.messages, want)
	}

	if err := m.SetMuted(context.Background(), mixer.Target{Kind: mixer.KindMuteGroup, Number: 2}, false); err != nil {
		t.Fatalf("SetMuted() error = %v", err)
	}
	want = []byte{0x90, 0x52, 0x7F, 0x90, 0x52, 0x10}
	if !bytes.Equal(sender.messages[1], want) {
		t.Fatalf("message = % X, want % X", sender.messages[1], want)
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
		t.Fatalf("zero tempo must not send a message, got % X", sender.messages)
	}

	if err := m.SetTapTempo(context.Background(), 2, 500*time.Millisecond); err != nil {
		t.Fatalf("SetTapTempo() error = %v", err)
	}
	want := append(
		nrpn.Message(0, 0x49, 0x05, 0x12, 0x05),
		nrpn.Message(0, 0x48, 0x05, 0x69, 0x05)...,
	)
	if len(sender.messages) != 1 || !bytes.Equal(sender.messages[0], want) {
		t.Fatalf("message = % X, want % X", sender.messages, want)
	}
}

func TestDelayValues(t *testing.T) {
	tests := []struct {
		tempo        time.Duration
		coarse, fine uint8
	}{
		{1 * time.Millisecond, 0x3C, 0x00},
		{500 * time.Millisecond, 0x69, 0x12},
		{5 * time.Second, 0x7F, 0x7F},
	}
	for _, tt := range tests {
		coarse, fine := delayValues(tt.tempo)
		if coarse != tt.coarse || fine != tt.fine {
			t.Errorf("delayValues(%s) = (%#x, %#x), want (%#x, %#x)", tt.tempo, coarse, fine, tt.coarse, tt.fine)
		}
	}
}
