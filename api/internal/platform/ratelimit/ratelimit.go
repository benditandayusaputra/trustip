package ratelimit

import (
	"math"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Rule struct {
	Limit  int
	Window time.Duration
}

type Decision struct {
	Allowed           bool
	Limit             int
	Remaining         int
	ResetSeconds      int
	RetryAfterSeconds int
}

type bucket struct {
	tokens    float64
	updatedAt time.Time
}

type Limiter struct {
	mu        sync.Mutex
	rule      Rule
	clock     Clock
	buckets   map[string]*bucket
	lastSweep time.Time
}

func New(rule Rule, clock Clock) *Limiter {
	return &Limiter{rule: rule, clock: clock, buckets: map[string]*bucket{}, lastSweep: clock.Now()}
}

func (l *Limiter) Allow(key string) Decision {
	now := l.clock.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweepIdle(now)
	current := l.refill(key, now)
	if current.tokens < 1 {
		denied := l.snapshot(current.tokens)
		denied.RetryAfterSeconds = l.secondsToEarn(1 - current.tokens)
		return denied
	}
	current.tokens--
	allowed := l.snapshot(current.tokens)
	allowed.Allowed = true
	return allowed
}

func (l *Limiter) refill(key string, now time.Time) *bucket {
	current, ok := l.buckets[key]
	if !ok {
		current = &bucket{tokens: float64(l.rule.Limit), updatedAt: now}
		l.buckets[key] = current
	}
	elapsed := now.Sub(current.updatedAt).Seconds()
	current.tokens = math.Min(float64(l.rule.Limit), current.tokens+elapsed*l.tokensPerSecond())
	current.updatedAt = now
	return current
}

func (l *Limiter) snapshot(tokens float64) Decision {
	return Decision{
		Limit:        l.rule.Limit,
		Remaining:    int(math.Floor(tokens)),
		ResetSeconds: l.secondsToEarn(float64(l.rule.Limit) - tokens),
	}
}

func (l *Limiter) secondsToEarn(tokens float64) int {
	return int(math.Ceil(tokens / l.tokensPerSecond()))
}

func (l *Limiter) tokensPerSecond() float64 {
	return float64(l.rule.Limit) / l.rule.Window.Seconds()
}

func (l *Limiter) sweepIdle(now time.Time) {
	if now.Sub(l.lastSweep) < l.rule.Window {
		return
	}
	for key, idle := range l.buckets {
		if now.Sub(idle.updatedAt) >= l.rule.Window {
			delete(l.buckets, key)
		}
	}
	l.lastSweep = now
}

func (l *Limiter) trackedKeys() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
