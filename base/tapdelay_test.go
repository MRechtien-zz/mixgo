package base

import (
	"math"
	"testing"
	"time"
)

func TestCalculateTapTempo(t *testing.T) {

	tapDelay := &BaseTapDelay{
		LastTriggered: 0,
		Tapping:       []int64{},
	}

	// 1. a series of somewhat in-time / constant tapping
	//    => result should just be average tap time from 2nd tap on
	tapTempo := triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 0, tapTempo)

	time.Sleep(100 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 100, tapTempo)

	time.Sleep(110 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 105, tapTempo)

	time.Sleep(90 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 100, tapTempo)

	// 2. tapping a different tempo
	//    => should disregard current average and data and start with 0
	//       as first tap of new tempo doesn't indicate anything yet
	time.Sleep(200 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 0, tapTempo)

	time.Sleep(220 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 220, tapTempo)

	time.Sleep(180 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 200, tapTempo)

	// 3. tapping slower than max-delay time
	//    => should disregard current average and data and start with 0
	time.Sleep(330 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 0, tapTempo)

	time.Sleep(3300 * time.Millisecond)
	tapTempo = triggerTapTempo(t, tapDelay)
	verifyTapTempo(t, 0, tapTempo)
}

func verifyTapTempo(t *testing.T, expected int, actual int) {
	ratio := (math.Min(float64(expected), float64(actual)) / math.Max(float64(expected), float64(actual)))
	// give 'some room' incase of heavy cpu load
	if ratio < 0.95 {
		t.Errorf("Expected TapTempo == %d, got %d", expected, actual)
	}
}

func triggerTapTempo(t *testing.T, tapDelay *BaseTapDelay) int {
	// max delay time is 300
	return CalculateTapTempo(tapDelay, 300)
}

func TestCalculateAverageDelay(t *testing.T) {
	average := calculateAverageDelay([]int64{0})
	if average != 0 {
		t.Error("Expected average == 0, got", average)
	}

	average = calculateAverageDelay([]int64{100, 90, 110})
	if average != 100 {
		t.Error("Expected average == 0, got", average)
	}

	average = calculateAverageDelay([]int64{17, 333, 100, 150})
	if average != 150 {
		t.Error("Expected average == 150, got", average)
	}
}
