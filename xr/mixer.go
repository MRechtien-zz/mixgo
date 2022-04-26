package xr

import (
	"fmt"

	"github.com/hypebeast/go-osc/osc"
	"github.com/mrechtien/mixgo/base"
)

type XRMixer struct {
	output chan osc.Message
}

func NewMixer() base.Mixer {
	xrMixer := XRMixer{
		output: make(chan osc.Message),
	}
	go sendToMixer(xrMixer.output)
	var mixer base.Mixer = &xrMixer
	return mixer
}

func sendToMixer(output chan osc.Message) {
	client := osc.NewClient("192.168.178.126", int(10024))
	for message := range output {
		fmt.Printf("Sending message to mixer: %v\n", message)
		if err := client.Send(&message); err != nil {
			fmt.Println(err)
		}
	}
}

func (mixer *XRMixer) NewMuteGroup(muteChannel string) *base.MuteGroup {
	var muteGroup base.MuteGroup = NewMuteGroup(muteChannel, mixer.output)
	return &muteGroup
}

func (mixer *XRMixer) NewTapDelay(fxChannel string) *base.TapDelay {
	var tapDelay base.TapDelay = NewTapDelay(fxChannel, mixer.output)
	return &tapDelay
}
