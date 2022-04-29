package xr

import (
	"log"

	"github.com/hypebeast/go-osc/osc"
	"github.com/mrechtien/mixgo/base"
)

const (
	MIXER_NAME = "xr"
)

type XRMixer struct {
	output chan osc.Message
}

func init() {
	base.AddMixer(MIXER_NAME, func(ip string, port uint) *base.Mixer {
		return NewMixer(ip, port)
	})
}

func NewMixer(ip string, port uint) *base.Mixer {
	xrMixer := XRMixer{
		output: make(chan osc.Message),
	}
	go sendToMixer(ip, port, xrMixer.output)
	var mixer base.Mixer = &xrMixer
	return &mixer
}

func sendToMixer(ip string, port uint, output chan osc.Message) {
	client := osc.NewClient(ip, int(port))
	for message := range output {
		log.Printf("Sending message to mixer: %v\n", message)
		if err := client.Send(&message); err != nil {
			log.Println("Error while sending message: ", err)
		}
	}
}

func (mixer *XRMixer) NewMuteGroup(muteChannel byte) *base.MuteGroup {
	var muteGroup base.MuteGroup = NewMuteGroup(muteChannel, mixer.output)
	return &muteGroup
}

func (mixer *XRMixer) NewTapDelay(fxChannel byte) *base.TapDelay {
	var tapDelay base.TapDelay = NewTapDelay(fxChannel, mixer.output)
	return &tapDelay
}
