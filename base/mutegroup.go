package base

const (
	MUTE_ON  = true
	MUTE_OFF = false

	MUTE_GROUP = "MuteGroup"
)

type MuteGroup interface {
	Toggle(onOff bool)
}
