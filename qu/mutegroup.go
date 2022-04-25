package qu

import "github.com/mrechtien/mixgo/base"

const (
	MUTE_ON  = 0x40
	MUTE_OFF = 0x10
)

var muteGroupMapping = map[string]byte{
	"MUTE_ON":  0x40,
	"MUTE_OFF": 0x10,

	"MUTE_GROUP_1": 0x50,
	"MUTE_GROUP_2": 0x51,
	"MUTE_GROUP_3": 0x52,
	"MUTE_GROUP_4": 0x53,
}

type QuMuteGroup struct {
	base.MuteGroup
	midiChannel byte
	muteChannel byte
	output      chan []byte
}

func NewMuteGroup(midiChannel byte, muteChannel string, output chan []byte) *QuMuteGroup {
	muteGroup := QuMuteGroup{
		midiChannel: midiChannel,
		muteChannel: muteGroupMapping[muteChannel],
		output:      output,
	}
	return &muteGroup
}

func (muteGroup *QuMuteGroup) Toggle(onOff bool) {
	message := toMute(muteGroup.muteChannel, onOff)
	muteGroup.output <- message
}

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
