package base

const (
	MUTE_ON  = true
	MUTE_OFF = false

	MUTE_GROUP_1 = "MUTE_GROUP_1"
	MUTE_GROUP_2 = "MUTE_GROUP_2"
	MUTE_GROUP_3 = "MUTE_GROUP_3"
	MUTE_GROUP_4 = "MUTE_GROUP_4"
)

type MuteGroup interface {
	Toggle(onOff bool)
}
