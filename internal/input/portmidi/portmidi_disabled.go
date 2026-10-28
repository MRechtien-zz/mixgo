//go:build !cgo || !portmidi

package portmidi

import (
	"log/slog"

	"github.com/mrechtien/mixgo/internal/input"
)

// New returns ErrNotBuilt in builds without PortMidi support.
func New(_ Options, _ *slog.Logger) (input.Source, error) {
	return nil, ErrNotBuilt
}
