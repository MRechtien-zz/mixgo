## MixGo PoC

This is a simple (wip) project to handle MIDI CC messages to control an Allen & Heath QU 24 audio mixer.

Main goal is to control the mute groups and tap delay from a Raspberry PI via MIDI foot controller.

### Raspberry Pi setup

#### Install PortMidi

```bash
sudo apt install libportmidi-dev libportmidi0
```

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
PATH=$PATH:/usr/local/go/bin
GOPATH=$HOME/golang
```


### Hints

- On Mac OS X building the gomidi (V2) modules might fail due to `ld: library not found for -lporttime`. See https://gitlab.com/gomidi/midi/-/issues/33
