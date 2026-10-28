// Command mixgo controls a digital mixing console from MIDI or GPIO controls.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mrechtien/mixgo/internal/app"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("mixgo failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath    string
		bootstrapPath string
		logLevel      string
		showVersion   bool
	)

	flag.StringVar(&configPath, "config", "", "runtime configuration file (skips bootstrap selection)")
	flag.StringVar(&configPath, "c", "", "shorthand for -config")
	flag.StringVar(&bootstrapPath, "bootstrap", "config/bootstrap.yml", "bootstrap configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "log level: debug, info, warn or error")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return nil
	}
	if flag.NArg() > 1 {
		return fmt.Errorf("expected at most one configuration file argument, got %d", flag.NArg())
	}
	if configPath == "" && flag.NArg() == 1 {
		configPath = flag.Arg(0)
	}

	level, err := parseLogLevel(logLevel)
	if err != nil {
		return err
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = app.Run(ctx, app.Options{
		ConfigPath:    configPath,
		BootstrapPath: bootstrapPath,
		Logger:        slog.Default(),
	})
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q", raw)
	}
}
