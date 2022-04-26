package qu

const (
	MUTE_ON     = 0x40
	MUTE_OFF    = 0x10
	MUTE_GROUPS = 0x50 // start channel
)

type QuMuteGroup struct {
	midiChannel byte
	muteChannel byte
	output      chan []byte
}

func NewMuteGroup(midiChannel byte, muteChannel byte, output chan []byte) *QuMuteGroup {
	muteGroup := QuMuteGroup{
		midiChannel: midiChannel,
		muteChannel: MUTE_GROUPS + muteChannel,
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
