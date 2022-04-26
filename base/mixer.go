package base

type Mixer interface {
	NewMuteGroup(muteChannel byte) *MuteGroup
	NewTapDelay(fxChannel byte) *TapDelay
}
