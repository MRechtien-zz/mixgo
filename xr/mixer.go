package xr

import (
	"fmt"

	"github.com/hypebeast/go-osc/osc"
	"github.com/mrechtien/mixgo/base"
)

type XRMixer struct {
	base.Mixer
	output chan osc.Message
}

func NewMixer() XRMixer {
	mixer := XRMixer{
		output: make(chan osc.Message),
	}
	go sendToMixer(mixer.output)
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

func (mixer *XRMixer) NewMuteGroup(muteGroup string) *XRMuteGroup {
	return NewMuteGroup(muteGroup, mixer.output)
}

func (mixer *XRMixer) NewTapDelay(fxChannel string) *XRTapDelay {
	return NewTapDelay(fxChannel, mixer.output)
}
