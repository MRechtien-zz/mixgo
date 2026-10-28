package cq

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

func TestSetMuted(t *testing.T) {
	sender := &fakeSender{}
	m := newWithSender(sender, nil)

	tests := []struct {
		target mixer.Target
		muted  bool
		want   []byte
	}{
		{mixer.Target{Kind: mixer.KindMuteGroup, Number: 2}, true, nrpn.Message(0, 0x04, 2, 0, 1)},
		{mixer.Target{Kind: mixer.KindMuteChannel, Number: 81}, false, nrpn.Message(0, 0x00, 81, 0, 0)},
	}
	for _, tt := range tests {
		if err := m.SetMuted(context.Background(), tt.target, tt.muted); err != nil {
			t.Fatalf("SetMuted() error = %v", err)
		}
		if got := sender.messages[len(sender.messages)-1]; !bytes.Equal(got, tt.want) {
			t.Errorf("message = % X, want % X", got, tt.want)
		}
	}
}

func TestSetMutedUnsupportedKind(t *testing.T) {
	m := newWithSender(&fakeSender{}, nil)
	err := m.SetMuted(context.Background(), mixer.Target{Kind: mixer.KindTapDelay}, true)
	if !errors.Is(err, mixer.ErrUnsupported) {
		t.Fatalf("error = %v, want ErrUnsupported", err)
	}
}

func TestSetTapTempoAlwaysSendsSoftKey(t *testing.T) {
	sender := &fakeSender{}
	m := newWithSender(sender, nil)

	if err := m.SetTapTempo(context.Background(), 2, 0); err != nil {
		t.Fatalf("SetTapTempo() error = %v", err)
	}
	want := []byte{0x90, 0x32, 0x7F}
	if len(sender.messages) != 1 || !bytes.Equal(sender.messages[0], want) {
		t.Fatalf("message = % X, want % X", sender.messages, want)
	}

	_ = m.SetTapTempo(context.Background(), 2, time.Second)
	if len(sender.messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(sender.messages))
	}
}
