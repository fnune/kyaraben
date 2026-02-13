package packages

import (
	"testing"
	"time"
)

func TestSpeedTrackerBytesPerSecond(t *testing.T) {
	tracker := NewSpeedTracker(500 * time.Millisecond)
	tracker.Record(1024)
	time.Sleep(120 * time.Millisecond)
	tracker.Record(1024)

	speed := tracker.BytesPerSecond()
	if speed <= 0 {
		t.Fatalf("expected positive speed, got %d", speed)
	}
}

func TestSpeedTrackerWindowTrim(t *testing.T) {
	tracker := NewSpeedTracker(50 * time.Millisecond)
	tracker.Record(1024)
	time.Sleep(80 * time.Millisecond)
	tracker.Record(1024)

	speed := tracker.BytesPerSecond()
	if speed != 0 {
		t.Fatalf("expected speed to reset after window trim, got %d", speed)
	}
}
