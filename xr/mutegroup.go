package xr

import (
	"fmt"

	"github.com/hypebeast/go-osc/osc"
	"github.com/mrechtien/mixgo/base"
)

const (
	MUTE_ON  = int32(1)
	MUTE_OFF = int32(0)
)

var muteGroupMapping = map[string]string{
	"MUTE_GROUP_1": "0",
	"MUTE_GROUP_2": "1",
	"MUTE_GROUP_3": "2",
	"MUTE_GROUP_4": "3",
}

type XRMuteGroup struct {
	base.MuteGroup
	muteChannel string
	output      chan osc.Message
}

func NewMuteGroup(muteChannel string, output chan osc.Message) *XRMuteGroup {
	muteGroup := XRMuteGroup{
		muteChannel: muteGroupMapping[muteChannel],
		output:      output,
	}
	return &muteGroup
}

func (muteGroup *XRMuteGroup) Toggle(onOff bool) {
	value := MUTE_OFF
	if onOff {
		value = MUTE_ON
	}

	message := osc.NewMessage(fmt.Sprintf("/config/mute/%s", muteGroup.muteChannel), value)
	muteGroup.output <- *message
}
