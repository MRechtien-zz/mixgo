package xr

import (
	"fmt"
	"time"

	"github.com/hypebeast/go-osc/osc"
	"github.com/mrechtien/mixgo/base"
)

const (
	MIN_DELAY_MILLIS = 50
	MAX_DELAY_MILLIS = 3000
)

var tapDelayMapping = map[string]string{
	"FX_SEND_1": "1",
	"FX_SEND_2": "2",
	"FX_SEND_3": "3",
	"FX_SEND_4": "4",
}

type XRTapDelay struct {
	base.TapDelay
	lastTriggered int64
	tapping       []int64
	fxChannel     string
	output        chan osc.Message
}

// channel is the mixer channel (FX) to trigger the tap delay on
func NewTapDelay(fxChannel string, output chan osc.Message) *XRTapDelay {
	tapDelay := XRTapDelay{
		lastTriggered: 0,
		tapping:       []int64{},
		fxChannel:     tapDelayMapping[fxChannel],
		output:        output,
	}
	return &tapDelay
}

/**
 * Takes input number and value to send value on LR Mix to mixer
 * @param channel Input number 1 - 48 (e.g. Ip1)
 * @param value ValueLevel Class with Level from -inf db to +10db
 */
func (tapDelay *XRTapDelay) Trigger() {
	tempo := tryComputeTapTempo(tapDelay)
	if tempo > 0 {
		percentage := normalizeTempo(tempo)
		message := generateDelayMessage(tapDelay, percentage)
		tapDelay.output <- message
	}
}

func tryComputeTapTempo(tapDelay *XRTapDelay) float32 {
	now := time.Now().UnixMilli()
	if tapDelay.lastTriggered > 0 && tapDelay.lastTriggered < now-MAX_DELAY_MILLIS {
		// reset if last trigger is longer than MAX_DELAY_TIME ago
		tapDelay.lastTriggered = 0
		tapDelay.tapping = []int64{}
	} else if tapDelay.lastTriggered > 0 {
		// calculate diff to last trigger
		diff := now - tapDelay.lastTriggered
		tapDelay.tapping = append(tapDelay.tapping, diff)
	}
	tapDelay.lastTriggered = now
	// calculate average delay
	var sum int
	for i := 0; i < len(tapDelay.tapping); i++ {
		sum += int(tapDelay.tapping[i])
	}
	if sum > 0 {
		return float32(sum / len(tapDelay.tapping))
	}
	return 0
}

func normalizeTempo(tempo float32) float32 {
	if tempo > MAX_DELAY_MILLIS {
		return 1
	} else if tempo < MIN_DELAY_MILLIS {
		return MIN_DELAY_MILLIS / MAX_DELAY_MILLIS
	}
	return tempo / MAX_DELAY_MILLIS
}

func generateDelayMessage(tapDelay *XRTapDelay, tempoPercentage float32) osc.Message {
	return *osc.NewMessage(fmt.Sprintf("/fx/%s/par/01", tapDelay.fxChannel), tempoPercentage)
}
