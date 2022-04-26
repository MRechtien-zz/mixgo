package qu

import (
	"fmt"
	"net"
	"time"

	"github.com/mrechtien/mixgo/base"
)

type QuMixer struct {
	output chan []byte
}

func NewMixer() base.Mixer {
	quMixer := QuMixer{
		output: make(chan []byte),
	}
	go sendToMixer(quMixer.output)
	var mixer base.Mixer = &quMixer
	return mixer
}

func sendToMixer(output chan []byte) {
	dialer := net.Dialer{Timeout: (time.Second * 5)}
	conn, err := dialer.Dial("tcp", "192.168.0.150:51325")
	if err != nil {
		fmt.Println("could not connect to TCP server: ", err)
	}
	for message := range output {
		fmt.Printf("Sending message to mixer: %v\n", message)
		conn.Write(message)
	}
	defer conn.Close()
}

func (mixer *QuMixer) NewMuteGroup(muteChannel string) *base.MuteGroup {
	var muteGroup base.MuteGroup = NewMuteGroup(0x00, muteChannel, mixer.output)
	return &muteGroup
}

func (mixer *QuMixer) NewTapDelay(fxChannel string) *base.TapDelay {
	var tapDelay base.TapDelay = NewTapDelay(0x00, fxChannel, mixer.output)
	return &tapDelay
}
