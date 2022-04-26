package input

import (
	"log"

	"github.com/mrechtien/mixgo/config"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
)

func printMidiDevices() {
	// allows you to get the ports when using "real" drivers like rtmididrv or portmididrv
	log.Printf("MIDI IN Ports\n")
	for i, port := range midi.InPorts() {
		log.Printf("no: %v %q\n", i, port)
	}
	log.Printf("\n\nMIDI OUT Ports\n")
	for i, port := range midi.OutPorts() {
		log.Printf("no: %v %q\n", i, port)
	}
	log.Printf("\n\n")
}

func SetupAndHandleMidi(config *config.Config, callback func(byte, byte, byte)) func() {

	go printMidiDevices()

	in := midi.FindInPort(config.Input.Name)
	if in < 0 {
		log.Fatalln("can't find given MIDI input device")
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, status, val byte
		switch {
		case msg.GetControlChange(&ch, &status, &val):
			log.Printf("Received MIDI control change: channel %02X, status %02X, value %02X\n", ch, status, val)
			callback(ch, status, val)
		/*
			case msg.GetSysEx(&bt):
				log.Printf("got sysex: %X\n", bt)
			case msg.GetNoteStart(&ch, &status, &val):
				log.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(status), ch, val)
			case msg.GetNoteEnd(&ch, &status):
				log.Printf("ending note %s on channel %v\n", midi.Note(status), ch)
		*/
		default:
			msg.GetSysEx(&bt)
			log.Printf("Unmapped MIDI event:  %+v\n", bt)
		}
	}, midi.UseSysEx())

	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return nil
	}

	return stop
}
