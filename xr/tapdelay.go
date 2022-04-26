package xr

import (
	"fmt"

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
	base.BaseTapDelay
	fxChannel string
	output    chan osc.Message
}

// channel is the mixer channel (FX) to trigger the tap delay on
func NewTapDelay(fxChannel string, output chan osc.Message) *XRTapDelay {
	tapDelay := XRTapDelay{
		BaseTapDelay: base.BaseTapDelay{
			LastTriggered: 0,
			Tapping:       []int64{},
		},
		fxChannel: tapDelayMapping[fxChannel],
		output:    output,
	}
	return &tapDelay
}

/**
 * Takes input number and value to send value on LR Mix to mixer
 * @param channel Input number 1 - 48 (e.g. Ip1)
 * @param value ValueLevel Class with Level from -inf db to +10db
 */
func (tapDelay *XRTapDelay) Trigger() {
	tempo := base.CalculateTapTempo(&tapDelay.BaseTapDelay, MAX_DELAY_MILLIS)
	if tempo > 0 {
		percentage := normalizeTempo(tempo)
		message := generateDelayMessage(tapDelay, percentage)
		tapDelay.output <- message
	}
}

func normalizeTempo(tempo int) float32 {
	if tempo > MAX_DELAY_MILLIS {
		return 1
	} else if tempo < MIN_DELAY_MILLIS {
		return MIN_DELAY_MILLIS / MAX_DELAY_MILLIS
	}
	return float32(tempo) / float32(MAX_DELAY_MILLIS)
}

func generateDelayMessage(tapDelay *XRTapDelay, tempoPercentage float32) osc.Message {
	return *osc.NewMessage(fmt.Sprintf("/fx/%s/par/01", tapDelay.fxChannel), tempoPercentage)
}
