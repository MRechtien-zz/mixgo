package main

import (
	"fmt"
	"net"
	"time"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv"
)

const (
	MUTE_ON  = 0x40
	MUTE_OFF = 0x10

	MUTE_GROUP_1 = 0x50
	MUTE_GROUP_2 = 0x51
	MUTE_GROUP_3 = 0x52
	MUTE_GROUP_4 = 0x53
)

func toMute(muteChannel byte, onOff bool) []byte {
	msg := []byte{0x90, 0x00, 0x7F, 0x90, 0x00, 0x40}
	msg[1] = muteChannel
	msg[4] = muteChannel
	if onOff {
		msg[5] = MUTE_ON
	} else {
		msg[5] = MUTE_OFF
	}
	return msg
}

func sendMuteToMixer(muteChannel byte, onOff bool) {
	conn, err := net.Dial("tcp", "192.168.0.150:51325")
	if err != nil {
		fmt.Println("could not connect to TCP server: ", err)
	}
	msg := toMute(muteChannel, onOff)
	fmt.Println("sending mute to mixer: ", msg)
	conn.Write(msg)
	defer conn.Close()
}

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

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, cc, val uint8
		switch {
		case msg.GetControlChange(&ch, &cc, &val):
			fmt.Printf("got cc %s on channel %v with value %v\n", midi.ControlChange(ch, cc, val), ch, val)
			switch {
			case cc == 0x02:
				sendMuteToMixer(MUTE_GROUP_3, val == 127)
			case cc == 0x03:
				sendMuteToMixer(MUTE_GROUP_4, val == 127)
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

	time.Sleep(time.Second * 10)

	stop()
}
