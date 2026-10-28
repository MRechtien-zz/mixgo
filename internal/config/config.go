// Package config loads and validates MixGo configuration files.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Action describes the mixer feature a binding controls.
type Action string

const (
	ActionMuteGroup   Action = "mute_group"
	ActionMuteChannel Action = "mute_channel"
	ActionTapDelay    Action = "tap_delay"
)

// NormalizeAction converts both the current snake_case names and the legacy
// CamelCase names to an Action.
func NormalizeAction(s string) Action {
	normalized := strings.ToLower(strings.TrimSpace(s))
	normalized = strings.NewReplacer("_", "", "-", "", " ", "").Replace(normalized)
	switch normalized {
	case "mutegroup":
		return ActionMuteGroup
	case "mutechannel":
		return ActionMuteChannel
	case "tapdelay":
		return ActionTapDelay
	default:
		return Action(strings.TrimSpace(s))
	}
}

// Valid reports whether a is a known action.
func (a Action) Valid() bool {
	switch a {
	case ActionMuteGroup, ActionMuteChannel, ActionTapDelay:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (a Action) String() string { return string(a) }

// UnmarshalYAML accepts "mute_group", "MuteGroup" and similar spellings.
func (a *Action) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*a = NormalizeAction(raw)
	if !a.Valid() {
		return fmt.Errorf("unknown action %q", raw)
	}
	return nil
}

// Mode describes how a value change is interpreted.
type Mode string

const (
	// ModeAbsolute sets the mixer state from the on_value comparison. This is
	// appropriate for latching controllers.
	ModeAbsolute Mode = "absolute"
	// ModeToggle flips the state on every press. The matching release event is
	// ignored.
	ModeToggle Mode = "toggle"
	// ModeMomentary keeps the state active while the control is pressed.
	ModeMomentary Mode = "momentary"
)

// Valid reports whether m is a known mode.
func (m Mode) Valid() bool {
	switch m {
	case ModeAbsolute, ModeToggle, ModeMomentary:
		return true
	default:
		return false
	}
}

// Input configures the control input source.
type Input struct {
	Type    string `yaml:"type"`    // gpio or portmidi
	Device  string `yaml:"device"`  // MIDI device name (portmidi only)
	Channel uint8  `yaml:"channel"` // zero based MIDI channel, 0-15
}

// Output configures the mixer connection.
type Output struct {
	Type    string `yaml:"type"`    // qu, cq or xr
	Address string `yaml:"address"` // host:port
}

// Display configures the LED display. An empty type disables it.
type Display struct {
	Type string `yaml:"type"` // gpio or empty
}

// Binding maps an input control to a mixer target.
type Binding struct {
	Action  Action `yaml:"action"`
	Control uint8  `yaml:"control"`
	Target  uint8  `yaml:"target"`
	OnValue uint8  `yaml:"on_value"`
	LED     *uint8 `yaml:"led,omitempty"`
	Mode    Mode   `yaml:"mode,omitempty"`
}

// Config is a runtime configuration.
type Config struct {
	Input    Input     `yaml:"input"`
	Output   Output    `yaml:"output"`
	Display  Display   `yaml:"display,omitempty"`
	Bindings []Binding `yaml:"bindings"`
}

// Bootstrap is the configuration used before the runtime profile is known.
type Bootstrap struct {
	Input    Input     `yaml:"input"`
	Display  Display   `yaml:"display,omitempty"`
	Profiles []Profile `yaml:"profiles"`
	Default  string    `yaml:"default,omitempty"`
	Timeout  Duration  `yaml:"timeout,omitempty"`
}

// Profile selects a runtime configuration when a control is pressed.
type Profile struct {
	Control uint8  `yaml:"control"`
	OnValue uint8  `yaml:"on_value,omitempty"`
	LED     *uint8 `yaml:"led,omitempty"`
	Config  string `yaml:"config"`
}

// Duration is a time.Duration that accepts strings such as "30s" in YAML.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	*d = Duration(parsed)
	return nil
}

// Or returns the duration, or def when it is not positive.
func (d Duration) Or(def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return time.Duration(d)
}

func decode(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(out)
}

// Load reads and validates a runtime configuration.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := decode(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return cfg, nil
}

// LoadBootstrap reads and validates a bootstrap configuration. Profile config
// paths are resolved relative to the bootstrap file.
func LoadBootstrap(path string) (Bootstrap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Bootstrap{}, fmt.Errorf("read bootstrap config %s: %w", path, err)
	}
	var b Bootstrap
	if err := decode(data, &b); err != nil {
		return Bootstrap{}, fmt.Errorf("parse bootstrap config %s: %w", path, err)
	}
	base := filepath.Dir(path)
	for i := range b.Profiles {
		b.Profiles[i].Config = resolvePath(base, b.Profiles[i].Config)
	}
	if b.Default != "" {
		b.Default = resolvePath(base, b.Default)
	}
	b.normalize()
	if err := b.Validate(); err != nil {
		return Bootstrap{}, fmt.Errorf("invalid bootstrap config %s: %w", path, err)
	}
	return b, nil
}

func resolvePath(base, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}

func (c *Config) normalize() {
	c.Input.Type = normalizeName(c.Input.Type)
	c.Output.Type = normalizeName(c.Output.Type)
	c.Display.Type = normalizeName(c.Display.Type)
	for i := range c.Bindings {
		b := &c.Bindings[i]
		b.Action = NormalizeAction(string(b.Action))
		if b.Mode == "" {
			b.Mode = ModeAbsolute
		}
		b.Mode = Mode(normalizeName(string(b.Mode)))
	}
}

func (b *Bootstrap) normalize() {
	b.Input.Type = normalizeName(b.Input.Type)
	b.Display.Type = normalizeName(b.Display.Type)
	for i := range b.Profiles {
		if b.Profiles[i].OnValue == 0 {
			b.Profiles[i].OnValue = 1
		}
	}
}

func normalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Validate checks the runtime configuration for consistency.
func (c Config) Validate() error {
	var errs []error
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	switch c.Input.Type {
	case "gpio", "portmidi":
	default:
		add("input.type must be \"gpio\" or \"portmidi\", got %q", c.Input.Type)
	}
	if c.Input.Channel > 15 {
		add("input.channel must be between 0 and 15, got %d", c.Input.Channel)
	}
	if c.Input.Type == "portmidi" && c.Input.Device == "" {
		add("input.device is required for portmidi input")
	}

	switch c.Output.Type {
	case "qu", "cq", "xr":
	default:
		add("output.type must be \"qu\", \"cq\" or \"xr\", got %q", c.Output.Type)
	}
	if err := validateAddress(c.Output.Address); err != nil {
		add("output.address: %v", err)
	}

	switch c.Display.Type {
	case "", "gpio":
	default:
		add("display.type must be \"gpio\" or empty, got %q", c.Display.Type)
	}

	if len(c.Bindings) == 0 {
		add("at least one binding is required")
	}

	seen := map[uint8]int{}
	ledsUsed := false
	for i, b := range c.Bindings {
		prefix := fmt.Sprintf("bindings[%d]", i)
		if !b.Action.Valid() {
			add("%s.action: unknown action %q", prefix, b.Action)
		}
		if b.Control > 127 {
			add("%s.control must be between 0 and 127, got %d", prefix, b.Control)
		}
		if b.Target > 127 {
			add("%s.target must be between 0 and 127, got %d", prefix, b.Target)
		}
		if b.OnValue > 127 {
			add("%s.on_value must be between 0 and 127, got %d", prefix, b.OnValue)
		}
		if !b.Mode.Valid() {
			add("%s.mode must be \"absolute\", \"toggle\" or \"momentary\", got %q", prefix, b.Mode)
		}
		if b.Mode == ModeToggle && b.Action == ActionTapDelay {
			add("%s.mode: toggle mode is not applicable to tap_delay", prefix)
		}
		if b.Mode == ModeToggle && b.OnValue == 0 {
			add("%s.on_value must be greater than zero in toggle mode", prefix)
		}
		if b.Mode == ModeMomentary && b.OnValue == 0 {
			add("%s.on_value must be greater than zero in momentary mode", prefix)
		}
		if b.Action == ActionTapDelay && b.OnValue == 0 {
			add("%s.on_value must be greater than zero for tap_delay", prefix)
		}
		if b.LED != nil {
			ledsUsed = true
			if *b.LED > 127 {
				add("%s.led must be between 0 and 127, got %d", prefix, *b.LED)
			}
		}
		if prev, ok := seen[b.Control]; ok {
			add("%s.control %d is already used by bindings[%d]", prefix, b.Control, prev)
		} else {
			seen[b.Control] = i
		}
	}
	if ledsUsed && c.Display.Type == "" {
		add("display.type is required when a binding uses an led")
	}

	return errors.Join(errs...)
}

// Validate checks the bootstrap configuration for consistency.
func (b Bootstrap) Validate() error {
	var errs []error
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	switch b.Input.Type {
	case "gpio", "portmidi":
	default:
		add("input.type must be \"gpio\" or \"portmidi\", got %q", b.Input.Type)
	}
	if b.Input.Channel > 15 {
		add("input.channel must be between 0 and 15, got %d", b.Input.Channel)
	}
	if b.Input.Type == "portmidi" && b.Input.Device == "" {
		add("input.device is required for portmidi input")
	}
	switch b.Display.Type {
	case "", "gpio":
	default:
		add("display.type must be \"gpio\" or empty, got %q", b.Display.Type)
	}
	if len(b.Profiles) == 0 {
		add("at least one profile is required")
	}

	seen := map[uint8]int{}
	ledsUsed := false
	for i, p := range b.Profiles {
		prefix := fmt.Sprintf("profiles[%d]", i)
		if p.Control > 127 {
			add("%s.control must be between 0 and 127, got %d", prefix, p.Control)
		}
		if p.OnValue > 127 {
			add("%s.on_value must be between 0 and 127, got %d", prefix, p.OnValue)
		}
		if p.Config == "" {
			add("%s.config is required", prefix)
		}
		if p.LED != nil {
			ledsUsed = true
			if *p.LED > 127 {
				add("%s.led must be between 0 and 127, got %d", prefix, *p.LED)
			}
		}
		if prev, ok := seen[p.Control]; ok {
			add("%s.control %d is already used by profiles[%d]", prefix, p.Control, prev)
		} else {
			seen[p.Control] = i
		}
	}
	if ledsUsed && b.Display.Type == "" {
		add("display.type is required when a profile uses an led")
	}

	return errors.Join(errs...)
}

func validateAddress(address string) error {
	host, portRaw, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("must be host:port: %w", err)
	}
	if host == "" {
		return errors.New("host must not be empty")
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q", portRaw)
	}
	return nil
}
