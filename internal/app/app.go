// Package app wires the input, mixer and display adapters together and runs
// the MixGo event loop.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mrechtien/mixgo/internal/config"
	"github.com/mrechtien/mixgo/internal/input"
)

// Options configures the application.
type Options struct {
	// ConfigPath is a runtime configuration file. When set, bootstrap
	// selection is skipped.
	ConfigPath string
	// BootstrapPath is the profile selection configuration.
	BootstrapPath string
	// Logger is the root logger. slog.Default is used when nil.
	Logger *slog.Logger
}

// Run loads the configuration and runs the event loop until ctx is cancelled.
func Run(ctx context.Context, opts Options) error {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}

	cfg, err := loadConfig(ctx, opts, log)
	if err != nil {
		return err
	}

	mix, err := newMixer(cfg.Output, log)
	if err != nil {
		return fmt.Errorf("create mixer: %w", err)
	}
	defer func() {
		if err := mix.Close(); err != nil {
			log.Warn("failed to close mixer", "error", err)
		}
	}()

	if err := validateCapabilities(cfg.Bindings, mix); err != nil {
		return err
	}

	panel, err := newPanel(cfg.Display, log)
	if err != nil {
		return fmt.Errorf("create display: %w", err)
	}
	defer func() {
		if err := panel.Close(); err != nil {
			log.Warn("failed to close display", "error", err)
		}
	}()

	src, err := newInput(cfg.Input, controlsFor(cfg.Bindings), log)
	if err != nil {
		return fmt.Errorf("create input: %w", err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Warn("failed to close input", "error", err)
		}
	}()

	dispatcher, err := NewDispatcher(cfg, mix, panel, log)
	if err != nil {
		return fmt.Errorf("create dispatcher: %w", err)
	}
	defer dispatcher.Close()

	log.Info("mixgo started",
		"mixer", mix.Name(),
		"address", cfg.Output.Address,
		"input", cfg.Input.Type,
		"bindings", len(cfg.Bindings),
	)

	err = src.Run(ctx, func(ev input.Event) {
		dispatcher.Handle(ctx, ev)
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("input: %w", err)
	}

	log.Info("mixgo stopped")
	return nil
}

func loadConfig(ctx context.Context, opts Options, log *slog.Logger) (config.Config, error) {
	if opts.ConfigPath != "" {
		return config.Load(opts.ConfigPath)
	}

	bootstrapPath := opts.BootstrapPath
	if bootstrapPath == "" {
		bootstrapPath = "config/bootstrap.yml"
	}
	boot, err := config.LoadBootstrap(bootstrapPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("load bootstrap configuration (use --config to bypass bootstrap): %w", err)
	}

	path, err := selectProfile(ctx, boot, log)
	if err != nil {
		return config.Config{}, err
	}
	return config.Load(path)
}
