package qu

import (
	"fmt"
	"net"

	"github.com/mrechtien/mixgo/base"
)

type QuMixer struct {
	base.Mixer
	output chan []byte
}

func NewMixer() QuMixer {
	mixer := QuMixer{
		output: make(chan []byte),
	}
	go sendToMixer(mixer.output)
	return mixer
}

func sendToMixer(output chan []byte) {
	conn, err := net.Dial("tcp", "192.168.0.150:51325")
	if err != nil {
		fmt.Println("could not connect to TCP server: ", err)
	}
	for message := range output {
		fmt.Printf("Sending message to mixer: %v\n", message)
		conn.Write(message)
	}
	defer conn.Close()
}

func (mixer *QuMixer) NewMuteGroup(muteGroup string) *QuMuteGroup {
	return NewMuteGroup(0x00, muteGroup, mixer.output)
}

func (mixer *QuMixer) NewTapDelay(fxChannel string) *QuTapDelay {
	return NewTapDelay(0x00, fxChannel, mixer.output)
}
