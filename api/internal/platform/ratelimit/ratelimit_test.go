package ratelimit

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

func newTestLimiter(limit int) (*Limiter, *clock.Manual) {
	manual := clock.NewManual(time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	return New(Rule{Limit: limit, Window: time.Minute}, manual), manual
}

func TestAllow_UpToLimit_ThenDenies(t *testing.T) {
	limiter, _ := newTestLimiter(5)

	for i := range 5 {
		if decision := limiter.Allow("ip:1"); !decision.Allowed || decision.Remaining != 4-i {
			t.Fatalf("request %d: %+v", i+1, decision)
		}
	}

	if decision := limiter.Allow("ip:1"); decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("sixth request: %+v", decision)
	}
}

func TestAllow_Denied_RetryAfterIsTimeToNextToken(t *testing.T) {
	limiter, _ := newTestLimiter(5)
	for range 5 {
		limiter.Allow("ip:1")
	}

	decision := limiter.Allow("ip:1")

	if decision.RetryAfterSeconds != 12 || decision.ResetSeconds != 60 {
		t.Fatalf("decision = %+v, want retry 12s and reset 60s", decision)
	}
}

func TestAllow_AfterRefillInterval_AllowsAgain(t *testing.T) {
	limiter, manual := newTestLimiter(5)
	for range 5 {
		limiter.Allow("ip:1")
	}

	manual.Advance(12 * time.Second)

	if decision := limiter.Allow("ip:1"); !decision.Allowed {
		t.Fatalf("decision after refill = %+v", decision)
	}
}

func TestAllow_JustBeforeRefill_StillDenies(t *testing.T) {
	limiter, manual := newTestLimiter(5)
	for range 5 {
		limiter.Allow("ip:1")
	}

	manual.Advance(11 * time.Second)

	if decision := limiter.Allow("ip:1"); decision.Allowed {
		t.Fatalf("decision before refill = %+v", decision)
	}
}

func TestAllow_KeysAreIndependent(t *testing.T) {
	limiter, _ := newTestLimiter(1)
	limiter.Allow("ip:1")

	if decision := limiter.Allow("ip:2"); !decision.Allowed {
		t.Fatalf("other key was limited: %+v", decision)
	}
}

func TestAllow_IdleKeys_AreSweptAfterWindow(t *testing.T) {
	limiter, manual := newTestLimiter(5)
	limiter.Allow("ip:1")
	limiter.Allow("ip:2")

	manual.Advance(2 * time.Minute)
	limiter.Allow("ip:3")

	if size := limiter.trackedKeys(); size != 1 {
		t.Fatalf("tracked keys = %d, want only the fresh key", size)
	}
}
