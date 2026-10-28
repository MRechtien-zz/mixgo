# MixGo

MixGo maps MIDI control change messages or GPIO buttons to a digital mixing
console. It is designed to run on a Raspberry Pi and control mute groups, mute
channels and tap delay from a foot controller.

Supported consoles:

| Mixer | Protocol | Mute groups | Mute channels | Tap delay |
|-------|----------|-------------|---------------|-----------|
| Allen & Heath QU | MIDI over TCP | yes | no | yes |
| Allen & Heath CQ | MIDI over TCP | yes | yes | yes |
| Behringer X Air (XR) | OSC/UDP | yes | no | yes |

Supported inputs:

| Input | Build requirement |
|-------|-------------------|
| GPIO buttons (periph.io) | none |
| MIDI (PortMidi) | `-tags portmidi` and `libportmidi` |

## Architecture

```
cmd/mixgo            command line entry point, flags and signal handling
internal/app         composition root, profile selection, dispatcher, lifecycle
internal/config      typed configuration loading and validation
internal/input       input port and GPIO/PortMidi adapters
internal/mixer       mixer port and QU/CQ/XR protocol adapters
internal/transport   shared TCP and OSC transports
internal/display     LED port and GPIO adapter
internal/tempo       tap tempo detector
internal/midi/nrpn   MIDI NRPN message builder
```

The application layer owns the mapping table, mute state and LED state. Input
adapters emit raw press/release events, mixer adapters only encode and send
protocol messages, and the display adapter only drives LEDs. Optional
components use a no-op implementation instead of nil checks.

## Build

Requirements: Go 1.24 or newer.

```bash
# Full build with PortMidi input (default; needs libportmidi headers):
make build

# Pure Go build (GPIO input only, no PortMidi/cgo):
make build-pure
```

The binary is written to `bin/mixgo`.

Install PortMidi on Debian/Ubuntu based systems:

```bash
sudo apt install libportmidi-dev libportmidi0
```

On macOS with Homebrew, point cgo at the Homebrew installation (see `.env`):

```bash
export CGO_LDFLAGS="-l portmidi -L/opt/homebrew/lib"
export CGO_CFLAGS="-I/opt/homebrew/include"
make build
```

## Run

Run a specific profile:

```bash
./bin/mixgo --config config/config-cq-gpio.yml
```

Run without a configuration file to use the bootstrap profile selector:

```bash
./bin/mixgo --bootstrap config/bootstrap.yml
```

A positional configuration argument is also accepted:

```bash
./bin/mixgo config/config-qu.yml
```

Useful flags:

```
--config, -c   runtime configuration (skips bootstrap)
--bootstrap    bootstrap configuration (default config/bootstrap.yml)
--log-level    debug, info, warn or error (default info)
--version      print the build version
```

`SIGINT` and `SIGTERM` cancel the context, stop all input listeners and close
mixer connections.

## Configuration

Runtime configuration (`config/config-cq-gpio.yml`):

```yaml
input:
  type: gpio          # gpio or portmidi
  channel: 0          # zero based MIDI channel, 0-15

display:
  type: gpio          # gpio, or omit to disable LEDs

output:
  type: cq            # qu, cq or xr
  address: 192.168.178.55:51325

bindings:
  - action: tap_delay # mute_group, mute_channel or tap_delay
    control: 17       # MIDI CC number or GPIO pin number
    target: 2         # mute group / channel / FX slot number
    on_value: 1       # value that counts as "pressed"
    led: 26           # optional GPIO pin
```

Binding fields:

- `action`: `mute_group`, `mute_channel` or `tap_delay`. The legacy spellings
  `MuteGroup`, `MuteChannel` and `TapDelay` are also accepted.
- `control`: MIDI CC number or GPIO pin. Must be unique within a profile.
- `target`: mixer element number. Its meaning depends on the mixer.
- `on_value`: raw control value that means "pressed". For a latching MIDI
  controller use `127`; for a momentary button use `1`.
- `mode`: `absolute` (default), `toggle` or `momentary`.
  - `absolute`: mixer state is `value == on_value`; use for latching controllers.
  - `toggle`: every press flips the state; release events are ignored.
  - `momentary`: state is active while the control is pressed.
- `led`: optional GPIO pin. `display.type` must be set when an LED is used.

MIDI channels are zero based internally: channel 1 on the device is configured
as `channel: 0`.

Tap delay reacts to press events only. `on_value` must be greater than zero.

Bootstrap configuration (`config/bootstrap.yml`):

```yaml
input:
  type: gpio
  channel: 0
display:
  type: gpio
timeout: 30s
default: config-cq-gpio.yml
profiles:
  - control: 17
    on_value: 1
    led: 26
    config: config-cq-gpio.yml
```

Profile paths are resolved relative to the bootstrap file. If no profile is
selected within `timeout`, `default` is used.

## Raspberry Pi setup

The hardware layout is:

```
MIDI foot controller => USB-MIDI interface => Raspberry Pi => WiFi => console
```

Install a recent Go runtime:

```bash
export GOLANG="$(curl -s https://go.dev/dl/ | awk -F[\>\<] '/linux-armv6l/ && !/beta/ {print $5;exit}')"
wget https://golang.org/dl/$GOLANG
sudo tar -C /usr/local -xzf $GOLANG
rm $GOLANG
```

Update `.profile`/`.bashrc`:

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH:/usr/local/go/bin
```

The sample service assumes the application is installed in `/opt/mixgo` and was
built with `make build`. Install `resources/mixgo.service` as a
user-level systemd unit:

```bash
mkdir -p ~/.config/systemd/user
cp resources/mixgo.service ~/.config/systemd/user/
sudo loginctl enable-linger "$USER"
systemctl --user enable --now mixgo.service
```

`script/start-mixgo.sh` selects a configuration based on the address of
`wlan0` and passes it via `--config`, which bypasses the bootstrap selector.
Remove the argument from the script if you prefer bootstrap selection.

A `wpa_supplicant` template for multiple WiFi networks is available in
`resources/wpa_supplicant.conf`.

## Testing

```bash
make test   # go test -race ./...
make lint   # go vet ./...
```

The tests do not need hardware. PortMidi-specific code is excluded from the
default build and from the default test run.

## Notes

- On macOS, building the PortMidi bindings may fail with
  `ld: library not found for -lporttime` or `'portmidi.h' file not found`.
  Install PortMidi and set `CGO_LDFLAGS`/`CGO_CFLAGS` as shown above.
- Configuration files from older MixGo versions used PascalCase keys
  (`Input`, `Output`, `Mappings`, ...). They must be migrated to the schema
  shown above; startup validation reports the offending field.
