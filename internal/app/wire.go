package app

import (
	"fmt"
	"log/slog"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/display"
	displaygpio "github.com/mrechtien/mixgo/internal/display/gpio"
	"github.com/mrechtien/mixgo/internal/input"
	inputgpio "github.com/mrechtien/mixgo/internal/input/gpio"
	inputportmidi "github.com/mrechtien/mixgo/internal/input/portmidi"
	"github.com/mrechtien/mixgo/internal/mixer"
	"github.com/mrechtien/mixgo/internal/mixer/cq"
	"github.com/mrechtien/mixgo/internal/mixer/qu"
	"github.com/mrechtien/mixgo/internal/mixer/xr"
)

func newInput(cfg config.Input, controls []uint8, log *slog.Logger) (input.Source, error) {
	switch cfg.Type {
	case "gpio":
		return inputgpio.New(inputgpio.Options{
			Channel:  cfg.Channel,
			Controls: controls,
		}, log)
	case "portmidi":
		return inputportmidi.New(inputportmidi.Options{
			Device:  cfg.Device,
			Channel: cfg.Channel,
		}, log)
	default:
		return nil, fmt.Errorf("unsupported input type %q", cfg.Type)
	}
}

func newMixer(cfg config.Output, log *slog.Logger) (mixer.Mixer, error) {
	switch cfg.Type {
	case "qu":
		return qu.New(cfg.Address, log)
	case "cq":
		return cq.New(cfg.Address, log)
	case "xr":
		return xr.New(cfg.Address, log)
	default:
		return nil, fmt.Errorf("unsupported mixer type %q", cfg.Type)
	}
}

func newPanel(cfg config.Display, log *slog.Logger) (display.Panel, error) {
	switch cfg.Type {
	case "":
		return display.Noop{}, nil
	case "gpio":
		return displaygpio.New(log)
	default:
		return nil, fmt.Errorf("unsupported display type %q", cfg.Type)
	}
}

func kindFor(action config.Action) (mixer.Kind, bool) {
	switch action {
	case config.ActionMuteGroup:
		return mixer.KindMuteGroup, true
	case config.ActionMuteChannel:
		return mixer.KindMuteChannel, true
	case config.ActionTapDelay:
		return mixer.KindTapDelay, true
	default:
		return 0, false
	}
}

func validateCapabilities(bindings []config.Binding, m mixer.Mixer) error {
	caps := m.Capabilities()
	for i, binding := range bindings {
		kind, ok := kindFor(binding.Action)
		if !ok {
			continue // reported by config validation
		}
		if !caps.Supports(kind) {
			return fmt.Errorf("bindings[%d]: mixer %q does not support %s", i, m.Name(), kind)
		}
	}
	return nil
}

func controlsFor(bindings []config.Binding) []uint8 {
	seen := map[uint8]struct{}{}
	controls := make([]uint8, 0, len(bindings))
	for _, binding := range bindings {
		if _, ok := seen[binding.Control]; ok {
			continue
		}
		seen[binding.Control] = struct{}{}
		controls = append(controls, binding.Control)
	}
	return controls
}

func controlsForProfiles(profiles []config.Profile) []uint8 {
	seen := map[uint8]struct{}{}
	controls := make([]uint8, 0, len(profiles))
	for _, profile := range profiles {
		if _, ok := seen[profile.Control]; ok {
			continue
		}
		seen[profile.Control] = struct{}{}
		controls = append(controls, profile.Control)
	}
	return controls
}
