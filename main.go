package main

import (
	"fmt"

	"github.com/mrechtien/mixgo/base"
	"github.com/mrechtien/mixgo/xr"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
)

func printMidiDevices() {
	// allows you to get the ports when using "real" drivers like rtmididrv or portmididrv
	fmt.Printf("MIDI IN Ports\n")
	for i, port := range midi.InPorts() {
		fmt.Printf("no: %v %q\n", i, port)
	}
	fmt.Printf("\n\nMIDI OUT Ports\n")
	for i, port := range midi.OutPorts() {
		fmt.Printf("no: %v %q\n", i, port)
	}
	fmt.Printf("\n\n")
}

func main() {
	defer midi.CloseDriver()

	//if len(os.Args) == 2 && os.Args[1] == "list" {
	go printMidiDevices()
	//return
	//}

	in := midi.FindInPort("MIDIMATE II Port 1")
	if in < 0 {
		fmt.Println("can't find given MIDI input device")
		return
	}

	//mixer := qu.NewMixer()
	mixer := xr.NewMixer()

	muteGroup3 := mixer.NewMuteGroup(base.MUTE_GROUP_3)
	muteGroup4 := mixer.NewMuteGroup(base.MUTE_GROUP_4)
	tapDelay3 := mixer.NewTapDelay(base.FX_SEND_2)

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, cc, val uint8
		switch {
		case msg.GetControlChange(&ch, &cc, &val):
			fmt.Printf("Received MIDI %s on channel %v with value %v\n", midi.ControlChange(ch, cc, val), ch, val)
			switch {
			case cc == 0x02:
				muteGroup3.Toggle(val == 127)
			case cc == 0x03:
				muteGroup4.Toggle(val == 127)
			case cc == 0x04:
				tapDelay3.Trigger()
			default:
			}
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &val):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, val)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s on channel %v\n", midi.Note(key), ch)
		default:
			// ignore
		}
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Scanln()
	fmt.Println("Exitting!")

	stop()
}
