package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/display"
	"github.com/mrechtien/mixgo/internal/input"
)

// selectProfile waits for a profile selection control and returns the path of
// the matching runtime configuration. On timeout it falls back to the
// configured default profile.
func selectProfile(ctx context.Context, boot config.Bootstrap, log *slog.Logger) (string, error) {
	src, err := newInput(boot.Input, controlsForProfiles(boot.Profiles), log)
	if err != nil {
		return "", fmt.Errorf("bootstrap input: %w", err)
	}
	defer src.Close()

	panel, err := newPanel(boot.Display, log)
	if err != nil {
		return "", fmt.Errorf("bootstrap display: %w", err)
	}
	defer panel.Close()

	return selectProfileWith(ctx, boot, src, panel, log)
}

func selectProfileWith(ctx context.Context, boot config.Bootstrap, src input.Source, panel display.Panel, log *slog.Logger) (string, error) {
	if log == nil {
		log = slog.Default()
	}
	profiles := make(map[input.Key]config.Profile, len(boot.Profiles))
	for _, profile := range boot.Profiles {
		profiles[input.Key{Channel: boot.Input.Channel, Control: profile.Control}] = profile
	}

	timeout := boot.Timeout.Or(30 * time.Second)
	selectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var selected string
	var selectedMu sync.Mutex
	runErr := src.Run(selectCtx, func(ev input.Event) {
		selectedMu.Lock()
		if selected != "" {
			selectedMu.Unlock()
			return
		}
		profile, ok := profiles[input.Key{Channel: ev.Channel, Control: ev.Control}]
		if !ok || ev.Value != profile.OnValue {
			selectedMu.Unlock()
			return
		}
		selected = profile.Config
		selectedMu.Unlock()

		if profile.LED != nil {
			_ = panel.Pulse(*profile.LED, 200*time.Millisecond)
		}
		log.Info("profile selected", "config", profile.Config)
		cancel()
	})

	selectedMu.Lock()
	result := selected
	selectedMu.Unlock()

	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if result != "" {
		return result, nil
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) && !errors.Is(runErr, context.DeadlineExceeded) {
		return "", fmt.Errorf("bootstrap input: %w", runErr)
	}
	if boot.Default != "" {
		log.Warn("no profile selected, using default", "config", boot.Default, "timeout", timeout)
		return boot.Default, nil
	}
	return "", fmt.Errorf("no profile selected within %s", timeout)
}
