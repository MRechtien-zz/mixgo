package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrechtien/mixgo/base"
	"github.com/mrechtien/mixgo/config"
	"github.com/mrechtien/mixgo/input"
	"github.com/mrechtien/mixgo/qu"
	"github.com/mrechtien/mixgo/xr"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
)

func midiToKey(ch uint8, status uint8) string {
	return fmt.Sprintf("%02X%02X", ch, status)
}

func main() {

	var cfg config.Config
	if len(os.Args) == 2 {
		configPath := os.Args[1]
		cfg = config.ReadConfig(configPath)
	}

	// setup mixer
	var mixer base.Mixer
	switch cfg.Output.Name {
	case qu.MIXER_NAME:
		mixer = qu.NewMixer(cfg.Output.Ip, cfg.Output.Port)
	case xr.MIXER_NAME:
		mixer = xr.NewMixer(cfg.Output.Ip, cfg.Output.Port)
	}

	// create callbacks for trigger mapping
	callbacks := map[string]interface{}{}
	for _, mapping := range cfg.Mappings {
		key := midiToKey(cfg.Input.Channel, mapping.CC)
		switch mapping.Name {
		case base.MUTE_GROUP:
			muteGroup := mixer.NewMuteGroup(mapping.Target)
			callbacks[key] = func(ch uint8, status uint8, val uint8) {
				(*muteGroup).Toggle(val == mapping.ValueOn)
			}
		case base.TAP_DELAY:
			tapDelay := mixer.NewTapDelay(mapping.Target)
			callbacks[key] = func(ch uint8, status uint8, val uint8) {
				(*tapDelay).Trigger()
			}
		default:
			log.Fatalln("Invalid mapping name in config: ", mapping.Name)
		}
	}

	// setup midi & input handling / callback
	stop := input.SetupAndHandleMidi(&cfg, func(ch, status, val byte) {
		key := midiToKey(ch, status)
		callback := callbacks[key]
		if callback == nil {
			log.Printf("Unmapped MIDI control change: %+v\n", midi.ControlChange(ch, status, val))
			return
		}
		callback.(func(ch, status, val byte))(ch, status, val)
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	log.Println("MixGo is up and running!")

	signal := <-signalChan
	log.Printf("Exitting on signal: %d\n", signal)
	stop()
	midi.CloseDriver()
	log.Println("Done.")
}
