//go:build cgo && portmidi

package portmidi

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/rakyll/portmidi"

	"github.com/mrechtien/mixgo/internal/input"
)

const readBufferSize = 64

var midiMu sync.Mutex

// Source reads control change events from a PortMidi input device.
type Source struct {
	stream  *portmidi.Stream
	channel uint8
	once    sync.Once
}

// New initialises PortMidi, discovers the configured device and opens it.
func New(opts Options, log *slog.Logger) (input.Source, error) {
	if log == nil {
		log = slog.Default()
	}
	log = log.With("component", "input/portmidi")

	midiMu.Lock()
	if err := portmidi.Initialize(); err != nil {
		midiMu.Unlock()
		return nil, fmt.Errorf("initialise portmidi: %w", err)
	}

	devices, err := discoverInputs(log)
	if err != nil {
		portmidi.Terminate()
		midiMu.Unlock()
		return nil, err
	}
	id, ok := devices[opts.Device]
	if !ok {
		portmidi.Terminate()
		midiMu.Unlock()
		return nil, fmt.Errorf("MIDI input %q not found (available: %v)", opts.Device, sortedKeys(devices))
	}

	stream, err := portmidi.NewInputStream(id, 1024)
	if err != nil {
		portmidi.Terminate()
		midiMu.Unlock()
		return nil, fmt.Errorf("open MIDI input %q: %w", opts.Device, err)
	}
	midiMu.Unlock()

	log.Info("opened MIDI input", "device", opts.Device, "channel", opts.Channel)
	return &Source{stream: stream, channel: opts.Channel}, nil
}

// Run polls the MIDI stream until ctx is cancelled.
func (s *Source) Run(ctx context.Context, emit func(input.Event)) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		events, err := s.stream.Read(readBufferSize)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("read MIDI input: %w", err)
		}
		for _, evt := range events {
			status := uint8(evt.Status)
			// Control change is 0xB0..0xBF. Status contains the zero based channel.
			if status&0xF0 != 0xB0 {
				continue
			}
			channel := status & 0x0F
			if channel != s.channel {
				continue
			}
			emit(input.Event{
				Key:   input.Key{Channel: channel, Control: uint8(evt.Data1)},
				Value: uint8(evt.Data2),
				At:    time.Now(),
			})
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Close closes the stream and terminates PortMidi.
func (s *Source) Close() error {
	var err error
	s.once.Do(func() {
		midiMu.Lock()
		defer midiMu.Unlock()
		if s.stream != nil {
			err = s.stream.Close()
		}
		if termErr := portmidi.Terminate(); err == nil {
			err = termErr
		}
	})
	return err
}

func discoverInputs(log *slog.Logger) (map[string]portmidi.DeviceID, error) {
	devices := map[string]portmidi.DeviceID{}
	count := portmidi.CountDevices()
	for i := 0; i < count; i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info == nil {
			continue
		}
		switch {
		case info.IsInputAvailable:
			log.Info("found MIDI input", "name", info.Name)
			devices[info.Name] = portmidi.DeviceID(i)
		case info.IsOutputAvailable:
			log.Debug("found MIDI output", "name", info.Name)
		}
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no MIDI input devices found")
	}
	return devices, nil
}

func sortedKeys(m map[string]portmidi.DeviceID) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
