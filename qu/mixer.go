package qu

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/mrechtien/mixgo/base"
)

const (
	MIXER_NAME = "qu"
)

func init() {
	base.AddMixer(MIXER_NAME, func(ip string, port uint) *base.Mixer {
		return NewMixer(ip, port)
	})
}

type QuMixer struct {
	output chan []byte
}

func NewMixer(ip string, port uint) *base.Mixer {
	quMixer := QuMixer{
		output: make(chan []byte),
	}
	go sendToMixer(ip, port, quMixer.output)
	var mixer base.Mixer = &quMixer
	return &mixer
}

func sendToMixer(ip string, port uint, output chan []byte) {
	for message := range output {
		dialer := net.Dialer{Timeout: (time.Second * 5)}
		connection, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			log.Printf("Could not connect to TCP server: %s", err)
		} else {
			log.Printf("Sending message to mixer: %v\n", message)
			connection.Write(message)
		}
		connection.Close()
	}
}

func (mixer *QuMixer) NewMuteGroup(muteChannel byte) *base.MuteGroup {
	var muteGroup base.MuteGroup = NewMuteGroup(0x00, muteChannel, mixer.output)
	return &muteGroup
}

func (mixer *QuMixer) NewTapDelay(fxChannel byte) *base.TapDelay {
	var tapDelay base.TapDelay = NewTapDelay(0x00, fxChannel, mixer.output)
	return &tapDelay
}
