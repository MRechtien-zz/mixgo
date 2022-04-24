package qu

import (
	"math"
	"time"
)

const (
	FX_SEND_1 = 0x00
	FX_SEND_2 = 0x01
	FX_SEND_3 = 0x02
	FX_SEND_4 = 0x03

	MIN_DELAY_MILLIS = 5
	MAX_DELAY_MILLIS = 1360

	// byte indices
	PLACE_MSB    = 2
	PLACE_LSB    = 5
	PLACE_OPTION = 7
	PLACE_VC     = 8
	PLACE_VF     = 11
)

type TapDelay struct {
	lastTriggered int64
	tapping       []int64
	midiChannel   byte
	fxChannel     byte
	output        chan []byte
}

// channel is the mixer channel (FX) to trigger the tap delay on
func NewTapDelay(midiChannel byte, fxChannel byte, output chan []byte) TapDelay {
	tapDelay := TapDelay{
		lastTriggered: 0,
		tapping:       []int64{},
		midiChannel:   midiChannel,
		fxChannel:     fxChannel,
		output:        output,
	}
	return tapDelay
}

/**
 * Takes input number and value to send value on LR Mix to mixer
 * @param channel Input number 1 - 48 (e.g. Ip1)
 * @param value ValueLevel Class with Level from -inf db to +10db
 */
func (tapDelay *TapDelay) Trigger() {
	tempo := tryComputeTapTempo(tapDelay)
	if tempo > 0 {
		course, fine := computeDelayValues(tempo)
		message := generateDelayMessage(tapDelay, 2, course, fine)
		tapDelay.output <- message
	}
}

func tryComputeTapTempo(tapDelay *TapDelay) int {
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
	var sum int64
	for i := 0; i < len(tapDelay.tapping); i++ {
		sum += tapDelay.tapping[i]
	}
	if sum > 0 {
		return int(sum) / len(tapDelay.tapping)
	}
	return 0
}

func generateDelayMessage(tapDelay *TapDelay, channel byte, coarseValue byte, fineValue byte) []byte {
	// Fine and course value resolution time value = 00 to 7F
	// Last byte - left tap: 0x05, right tap: 0x07
	fineData := toSendValue(channel, 0x49, fineValue, 0x05)
	coarseData := toSendValue(channel, 0x48, coarseValue, 0x05)

	setMidiChannel(tapDelay.midiChannel, fineData)
	setMidiChannel(tapDelay.midiChannel, coarseData)

	return append(fineData, coarseData...)
}

func computeDelayValues(delayMillis int) (byte, byte) {

	// Returns a tuple with MIDI parameter values representing the given delay (seconds as float).
	// Returns (0x00,0x00) if delay time is below minimum time.
	// (0x7F, 0x7F) if it is above maximum delay time.
	if delayMillis <= MIN_DELAY_MILLIS {
		// limit to default to ~69ms min (slapback kinda delay)
		return 0x3C, 0x00
	}
	if delayMillis >= MAX_DELAY_MILLIS {
		// limit to maximum delay time
		return 0x7F, 0x7F
	}

	// the next three lines are according to the specs from A & H, 30 June 2014, 15: 19
	value := math.Round(16383 * (math.Log10(float64(delayMillis)) - math.Log10(5)) / 2.4346)
	course := math.Floor(value / 128)
	fine := math.Mod(value, 128)

	return byte(math.Round(course)), byte(math.Round(fine))
}

func toSendValue(msb byte, lsb byte, vc byte, vf byte) []byte {
	message := []byte{0xB0, 0x63, 0x00, 0xB0, 0x62, 0x00, 0xB0, 0x06, 0x00, 0xB0, 0x26, 0x00}
	message[PLACE_MSB] = msb
	message[PLACE_LSB] = lsb
	message[PLACE_VC] = vc
	message[PLACE_VF] = vf
	return message
}

func setMidiChannel(channel byte, message []byte) {
	if len(message) != 9 && len(message) != 12 {
		panic("MIDI message length must be 9 or 12 bytes")
	}

	message[0] = 0xB0 + channel
	message[3] = 0xB0 + channel
	message[6] = 0xB0 + channel
	if len(message) > 9 {
		message[9] = 0xB0 + channel
	}
}
