package packages

import (
	"sync"
	"time"
)

type SpeedTracker struct {
	samples    []speedSample
	windowSize time.Duration
	mu         sync.Mutex
}

type speedSample struct {
	bytes     int64
	timestamp time.Time
}

func NewSpeedTracker(windowSize time.Duration) *SpeedTracker {
	return &SpeedTracker{
		samples:    make([]speedSample, 0),
		windowSize: windowSize,
	}
}

func (s *SpeedTracker) Record(bytes int64) {
	if bytes <= 0 {
		return
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = append(s.samples, speedSample{bytes: bytes, timestamp: now})
	s.trim(now)
}

func (s *SpeedTracker) BytesPerSecond() int64 {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trim(now)
	if len(s.samples) < 2 {
		return 0
	}
	var total int64
	for _, sample := range s.samples {
		total += sample.bytes
	}
	duration := s.samples[len(s.samples)-1].timestamp.Sub(s.samples[0].timestamp)
	if duration <= 0 {
		return 0
	}
	return int64(float64(total) / duration.Seconds())
}

func (s *SpeedTracker) trim(now time.Time) {
	if s.windowSize <= 0 {
		return
	}
	cutoff := now.Add(-s.windowSize)
	var idx int
	for idx < len(s.samples) && s.samples[idx].timestamp.Before(cutoff) {
		idx++
	}
	if idx > 0 {
		s.samples = s.samples[idx:]
	}
}
