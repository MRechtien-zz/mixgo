// Package tempo contains a tap tempo detector with an injectable clock.
package tempo

import "time"

const (
	defaultMaxGap  = 3 * time.Second
	defaultMaxTaps = 3
)

// Detector calculates an average tap interval from a sequence of taps.
type Detector struct {
	maxGap  time.Duration
	maxTaps int
	taps    []time.Duration
	last    time.Time
}

// New creates a detector. maxGap is the longest accepted interval between two
// taps; longer pauses reset the detector. maxTaps limits the number of stored
// intervals to prevent slow adaptation.
func New(maxGap time.Duration, maxTaps int) *Detector {
	if maxGap <= 0 {
		maxGap = defaultMaxGap
	}
	if maxTaps <= 0 {
		maxTaps = defaultMaxTaps
	}
	return &Detector{maxGap: maxGap, maxTaps: maxTaps}
}

// Tap records a tap at the given time and returns the averaged interval. The
// boolean is false until at least one interval has been measured.
func (d *Detector) Tap(at time.Time) (time.Duration, bool) {
	if d.last.IsZero() || at.Sub(d.last) > d.maxGap {
		d.last = at
		d.taps = d.taps[:0]
		return 0, false
	}

	diff := at.Sub(d.last)
	if len(d.taps) > d.maxTaps {
		d.taps = d.taps[:d.maxTaps]
	}

	if average := average(d.taps); average > 0 && abs(diff-average) > average/4 {
		// The new tap is more than 25% away from the previous average: start
		// over with the next interval instead of mixing two tempos.
		d.taps = d.taps[:0]
	} else {
		d.taps = append([]time.Duration{diff}, d.taps...)
	}
	d.last = at

	average := average(d.taps)
	return average, average > 0
}

// Reset clears the detector.
func (d *Detector) Reset() {
	d.taps = d.taps[:0]
	d.last = time.Time{}
}

func average(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	var sum time.Duration
	for _, value := range values {
		sum += value
	}
	return sum / time.Duration(len(values))
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
