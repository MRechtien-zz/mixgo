package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeAction(t *testing.T) {
	tests := map[string]Action{
		"mute_group":   ActionMuteGroup,
		"MuteGroup":    ActionMuteGroup,
		"mute-channel": ActionMuteChannel,
		"MUTE_CHANNEL": ActionMuteChannel,
		"TapDelay":     ActionTapDelay,
		"tap_delay":    ActionTapDelay,
	}
	for input, want := range tests {
		if got := NormalizeAction(input); got != want {
			t.Errorf("NormalizeAction(%q) = %q, want %q", input, got, want)
		}
	}
	if NormalizeAction("nope").Valid() {
		t.Error("unknown action must not be valid")
	}
}

func TestLoadValidConfig(t *testing.T) {
	path := writeConfig(t, `
input:
  type: gpio
  channel: 0
display:
  type: gpio
output:
  type: cq
  address: 192.168.1.10:51325
bindings:
  - action: mute_group
    control: 5
    target: 0
    on_value: 1
    mode: toggle
    led: 26
  - action: tap_delay
    control: 17
    target: 2
    on_value: 1
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Bindings) != 2 {
		t.Fatalf("bindings = %d, want 2", len(cfg.Bindings))
	}
	if cfg.Bindings[0].Mode != ModeToggle {
		t.Fatalf("mode = %q, want toggle", cfg.Bindings[0].Mode)
	}
	if cfg.Bindings[0].LED == nil || *cfg.Bindings[0].LED != 26 {
		t.Fatalf("led = %v, want 26", cfg.Bindings[0].LED)
	}
}

func TestLoadDefaultsModeToAbsolute(t *testing.T) {
	path := writeConfig(t, `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: mute_group, control: 1, target: 1, on_value: 1}
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Bindings[0].Mode != ModeAbsolute {
		t.Fatalf("mode = %q, want absolute", cfg.Bindings[0].Mode)
	}
}

func TestLoadValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "duplicate control",
			yaml: `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: mute_group, control: 1, target: 1, on_value: 1}
  - {action: mute_channel, control: 1, target: 2, on_value: 1}
`,
			wantErr: "already used",
		},
		{
			name: "invalid address",
			yaml: `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1}
bindings:
  - {action: mute_group, control: 1, target: 1, on_value: 1}
`,
			wantErr: "output.address",
		},
		{
			name: "tap delay without on value",
			yaml: `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: tap_delay, control: 1, target: 1}
`,
			wantErr: "tap_delay",
		},
		{
			name: "led without display",
			yaml: `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: mute_group, control: 1, target: 1, on_value: 1, led: 26}
`,
			wantErr: "display.type",
		},
		{
			name: "portmidi without device",
			yaml: `
input: {type: portmidi}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: mute_group, control: 1, target: 1, on_value: 1}
`,
			wantErr: "input.device",
		},
		{
			name: "invalid action",
			yaml: `
input: {type: gpio}
output: {type: cq, address: 127.0.0.1:51325}
bindings:
  - {action: nope, control: 1, target: 1, on_value: 1}
`,
			wantErr: "unknown action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.yaml))
			if err == nil {
				t.Fatal("Load() succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := writeConfig(t, `
Input:
  Type: gpio
Output:
  Name: cq
  Ip: 127.0.0.1
  Port: 51325
Mappings: []
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "field Input not found") {
		t.Fatalf("error = %v, want unknown field error for legacy schema", err)
	}
}

func TestLoadBootstrapResolvesRelativePaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bootstrap.yml")
	data := `
input: {type: gpio}
display: {type: gpio}
timeout: 5s
default: runtime.yml
profiles:
  - control: 17
    led: 26
    config: runtime.yml
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	boot, err := LoadBootstrap(path)
	if err != nil {
		t.Fatalf("LoadBootstrap() error = %v", err)
	}
	want := filepath.Join(dir, "runtime.yml")
	if boot.Default != want {
		t.Errorf("default = %q, want %q", boot.Default, want)
	}
	if boot.Profiles[0].Config != want {
		t.Errorf("profile config = %q, want %q", boot.Profiles[0].Config, want)
	}
	if boot.Profiles[0].OnValue != 1 {
		t.Errorf("profile on_value = %d, want 1", boot.Profiles[0].OnValue)
	}
	if got := boot.Timeout.Or(time.Second); got != 5*time.Second {
		t.Errorf("timeout = %s, want 5s", got)
	}
}

func TestRepositoryConfigsLoad(t *testing.T) {
	dir := filepath.Join("..", "..", "config")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" || entry.Name() == "bootstrap.yml" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			if _, err := Load(filepath.Join(dir, entry.Name())); err != nil {
				t.Fatalf("Load() error = %v", err)
			}
		})
	}
	if _, err := LoadBootstrap(filepath.Join(dir, "bootstrap.yml")); err != nil {
		t.Fatalf("LoadBootstrap() error = %v", err)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
