package clock

import (
	"sync"
	"time"
)

type System struct{}

func (System) Now() time.Time {
	return time.Now().UTC()
}

type Manual struct {
	mu  sync.Mutex
	now time.Time
}

func NewManual(start time.Time) *Manual {
	return &Manual{now: start.UTC()}
}

func (m *Manual) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

func (m *Manual) Advance(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = m.now.Add(duration)
}
