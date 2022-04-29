package base

type Mixer interface {
	NewMuteGroup(muteChannel byte) *MuteGroup
	NewTapDelay(fxChannel byte) *TapDelay
}

var mixerRegistry = map[string]interface{}{}

func AddMixer(name string, creator func(ip string, port uint) *Mixer) {
	mixerRegistry[name] = creator
}

func CreateMixer(name string, ip string, port uint) *Mixer {
	return mixerRegistry[name].(func(ip string, port uint) *Mixer)(ip, port)
}
