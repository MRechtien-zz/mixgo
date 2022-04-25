package base

type Mixer interface {
	NewMuteGroup(group string) *MuteGroup
	NewTapDelay(fx string) *TapDelay
}
