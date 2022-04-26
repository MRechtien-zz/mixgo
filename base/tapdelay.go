package base

import (
	"math"
	"time"
)

const (
	TAP_DELAY = "TapDelay"
)

type TapDelay interface {
	Trigger()
}

type BaseTapDelay struct {
	LastTriggered int64
	Tapping       []int64
}

// TODO broken
func CalculateTapTempo(tapDelay *BaseTapDelay, maxDelayTime int) int {
	now := time.Now().UnixMilli()
	if len(tapDelay.Tapping) > 3 {
		// limit length to prevent slow time approximation with long delays
		tapDelay.Tapping = tapDelay.Tapping[:3]
	}
	if tapDelay.LastTriggered > 0 && tapDelay.LastTriggered < now-int64(maxDelayTime) {
		// reset if last trigger is longer than MAX_DELAY_TIME ago
		tapDelay.LastTriggered = 0
		tapDelay.Tapping = []int64{}
	} else if tapDelay.LastTriggered > 0 {
		// calculate diff to last trigger
		diff := now - tapDelay.LastTriggered
		average := calculateAverageDelay(tapDelay.Tapping)
		if average > 0 && math.Abs(float64(diff-average)) > float64(average/4) {
			// tap is off more then X % from average: reset
			tapDelay.Tapping = []int64{}
		} else {
			// prepend new time as we might truncate (above)
			tapDelay.Tapping = append([]int64{diff}, tapDelay.Tapping...)
		}
	}
	tapDelay.LastTriggered = now
	return int(calculateAverageDelay(tapDelay.Tapping))
}

func calculateAverageDelay(tapping []int64) int64 {
	var sum int64
	for i := 0; i < len(tapping); i++ {
		sum += tapping[i]
	}
	if sum > 0 {
		return sum / int64(len(tapping))
	}
	return 0
}
