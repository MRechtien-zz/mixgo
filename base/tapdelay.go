package base

const (
	FX_SEND_1 = "FX_SEND_1"
	FX_SEND_2 = "FX_SEND_2"
	FX_SEND_3 = "FX_SEND_3"
	FX_SEND_4 = "FX_SEND_4"
)

type TapDelay interface {
	Trigger()
}
