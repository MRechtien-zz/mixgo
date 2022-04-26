package base

import (
	"time"
)

const (
	FX_SEND_1 = "FX_SEND_1"
	FX_SEND_2 = "FX_SEND_2"
	FX_SEND_3 = "FX_SEND_3"
	FX_SEND_4 = "FX_SEND_4"
)

type TapDelay interface {
	Trigger()
}

type BaseTapDelay struct {
	LastTriggered int64
	Tapping       []int64
}

func CalculateTapTempo(tapDelay *BaseTapDelay, maxDelayTime int) int {
	now := time.Now().UnixMilli()
	if tapDelay.LastTriggered > 0 && tapDelay.LastTriggered < now-int64(maxDelayTime) {
		// reset if last trigger is longer than MAX_DELAY_TIME ago
		tapDelay.LastTriggered = 0
		tapDelay.Tapping = []int64{}
	} else if tapDelay.LastTriggered > 0 {
		// calculate diff to last trigger
		diff := now - tapDelay.LastTriggered
		tapDelay.Tapping = append(tapDelay.Tapping, diff)
	}
	tapDelay.LastTriggered = now
	// calculate average delay
	var sum int64
	for i := 0; i < len(tapDelay.Tapping); i++ {
		sum += tapDelay.Tapping[i]
	}
	if sum > 0 {
		return int(sum) / len(tapDelay.Tapping)
	}
	return 0
}
