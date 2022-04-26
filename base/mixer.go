package base

type Mixer interface {
	NewMuteGroup(muteChannel string) *MuteGroup
	NewTapDelay(fxChannel string) *TapDelay
}
