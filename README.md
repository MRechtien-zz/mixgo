## MixGo PoC

This project handles MIDI CC messages from a given MIDI interface to control (digital) audio mixing consoles.

Main goal is to control the mute groups and tap delay from a Raspberry PI via MIDI foot controller.

Current implementation supports `Allen & Heath QU 24` and `Behringer X Air 18`. While `Allen & Heath` uses *NRPN* MIDI messages the `Behringer` requires *OSC* messages for which [go-osc](https://github.com/hypebeast/go-osc) is used.

### Raspberry Pi setup

The hardware layout of my setup looks like:

*_MIDI FootController => USB-Midi Interface => Raspberry PI => WIFI => Digital Mixing Console_*

#### Install PortMidi

```bash
sudo apt install libportmidi-dev libportmidi0
```
The Raspberry PI and portmidi should work with most class compliant USB-MIDI interfaces.

#### Install (recent) GO runtime

```bash
export GOLANG="$(curl -s https://go.dev/dl/ | awk -F[\>\<] '/linux-armv6l/ && !/beta/ {print $5;exit}')"
wget https://golang.org/dl/$GOLANG
sudo tar -C /usr/local -xzf $GOLANG
rm $GOLANG
unset GOLANG
```

Update .profile/.bashrc

```bash
GOPATH=$HOME/go
PATH=$PATH:$GOPATH:/usr/local/go/bin
```

#### Setting up mixgo service

The contained configuration files assume the app is installed in `/opt/mixgo` and has been build running `go build`.

`resource/mixgo.service` can be installed as (user-level) `systemd` service - see file for more instructions. 
This _unit_ definition references the start-up `script/start-mixgo.sh`. 
The startup-script greps the current IP/network of interface "wlan0" in order to determine the matching configuration for that setup and launches the `mixgo` executable.

A _wpa_supplicant_ configuration template can be found in `resources/wpa_supplicant.conf` to indicate howto setup multiple WIFI configurations. 


### Hints

-   On Mac OS X building the gomidi (V2) modules might fail due to `ld: library not found for -lporttime`. See <https://gitlab.com/gomidi/midi/-/issues/33>
