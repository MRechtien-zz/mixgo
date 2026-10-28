package tempo

import (
	"testing"
	"time"
)

func TestDetectorTap(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	d := New(300*time.Millisecond, 3)

	tests := []struct {
		name   string
		offset time.Duration
		want   time.Duration
		wantOK bool
	}{
		{name: "first tap", offset: 0, want: 0, wantOK: false},
		{name: "second tap", offset: 100 * time.Millisecond, want: 100 * time.Millisecond, wantOK: true},
		{name: "third tap averages", offset: 210 * time.Millisecond, want: 105 * time.Millisecond, wantOK: true},
		{name: "fourth tap averages", offset: 300 * time.Millisecond, want: 100 * time.Millisecond, wantOK: true},
		{name: "tempo change resets", offset: 500 * time.Millisecond, want: 0, wantOK: false},
		{name: "new tempo first interval", offset: 720 * time.Millisecond, want: 220 * time.Millisecond, wantOK: true},
		{name: "new tempo average", offset: 900 * time.Millisecond, want: 200 * time.Millisecond, wantOK: true},
		{name: "gap resets", offset: 4200 * time.Millisecond, want: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := d.Tap(base.Add(tt.offset))
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("Tap(%s) = (%s, %v), want (%s, %v)", tt.offset, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestDetectorLimitsTapCount(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	d := New(10*time.Second, 1)

	d.Tap(base)
	if tempo, ok := d.Tap(base.Add(100 * time.Millisecond)); !ok || tempo != 100*time.Millisecond {
		t.Fatalf("unexpected first interval: %s, %v", tempo, ok)
	}
	if tempo, ok := d.Tap(base.Add(200 * time.Millisecond)); !ok || tempo != 100*time.Millisecond {
		t.Fatalf("unexpected second interval: %s, %v", tempo, ok)
	}
	if len(d.taps) != 2 {
		t.Fatalf("stored taps = %d, want 2 (maxTaps+1)", len(d.taps))
	}
}

func TestDetectorReset(t *testing.T) {
	d := New(time.Second, 3)
	base := time.Now()
	d.Tap(base)
	d.Tap(base.Add(100 * time.Millisecond))
	d.Reset()
	if len(d.taps) != 0 || !d.last.IsZero() {
		t.Fatal("Reset did not clear the detector")
	}
}

func TestAverage(t *testing.T) {
	if got := average(nil); got != 0 {
		t.Fatalf("average(nil) = %s, want 0", got)
	}
	if got := average([]time.Duration{100, 90, 110}); got != 100 {
		t.Fatalf("average = %s, want 100ms", got)
	}
}
