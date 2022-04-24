package qu

const (
	MUTE_ON  = 0x40
	MUTE_OFF = 0x10

	MUTE_GROUP_1 = 0x50
	MUTE_GROUP_2 = 0x51
	MUTE_GROUP_3 = 0x52
	MUTE_GROUP_4 = 0x53
)

type MuteGroup struct {
	midiChannel byte
	muteChannel byte
	output      chan []byte
}

func NewMuteGroup(midiChannel byte, muteChannel byte, output chan []byte) MuteGroup {
	muteGroup := MuteGroup{
		midiChannel: midiChannel,
		muteChannel: muteChannel,
		output:      output,
	}
	return muteGroup
}

func (muteGroup *MuteGroup) Toggle(onOff bool) {
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
