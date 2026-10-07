package revocation

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type List struct {
	mu        sync.Mutex
	clock     Clock
	until     map[string]time.Time
	lastSweep time.Time
}

func New(clock Clock) *List {
	return &List{clock: clock, until: map[string]time.Time{}, lastSweep: clock.Now()}
}

func (l *List) Revoke(familyID string, until time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if until.After(l.until[familyID]) {
		l.until[familyID] = until
	}
}

func (l *List) IsRevoked(familyID string) bool {
	now := l.clock.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepExpired(now)
	return now.Before(l.until[familyID])
}

func (l *List) sweepExpired(now time.Time) {
	if now.Sub(l.lastSweep) < time.Minute {
		return
	}
	for familyID, until := range l.until {
		if !now.Before(until) {
			delete(l.until, familyID)
		}
	}
	l.lastSweep = now
}

func (l *List) tracked() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.until)
}
