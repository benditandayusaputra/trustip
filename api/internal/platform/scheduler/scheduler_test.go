package scheduler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/scheduler"
)

func jakarta() *time.Location {
	return time.FixedZone("WIB", 7*60*60)
}

func at(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, jakarta())
}

func TestNextRun_SpecTwentyOneFive_ComputesTheNextWallClockTimeInJakarta(t *testing.T) {
	daily := scheduler.Job{Name: "daily", Hour: 0, Minute: 5, Days: scheduler.EveryDay}
	weekly := scheduler.Job{Name: "weekly", Hour: 0, Minute: 10, Days: scheduler.Mondays}
	monthly := scheduler.Job{Name: "monthly", Hour: 0, Minute: 10, Days: scheduler.FirstOfMonth}
	friday := at(2026, 11, 13).Add(9 * time.Hour)
	cases := []struct {
		name  string
		job   scheduler.Job
		after time.Time
		want  time.Time
	}{
		{"daily later today is tomorrow", daily, friday, at(2026, 11, 14).Add(5 * time.Minute)},
		{
			"daily exactly at the run is the next day", daily, at(2026, 11, 14).Add(5 * time.Minute),
			at(2026, 11, 15).Add(5 * time.Minute),
		},
		{
			"daily just before the run is today", daily, at(2026, 11, 14).Add(4 * time.Minute),
			at(2026, 11, 14).Add(5 * time.Minute),
		},
		{"weekly on friday is next monday", weekly, friday, at(2026, 11, 16).Add(10 * time.Minute)},
		{"monthly mid month is the first", monthly, friday, at(2026, 12, 1).Add(10 * time.Minute)},
		{
			"monthly from UTC on the eve", monthly, time.Date(2026, 11, 30, 17, 0, 0, 0, time.UTC),
			at(2026, 12, 1).Add(10 * time.Minute),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scheduler.NextRun(tc.job, tc.after, jakarta()); !got.Equal(tc.want) {
				t.Fatalf("next run = %s, want %s", got, tc.want)
			}
		})
	}
}

type steppingClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *steppingClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *steppingClock) jump(to time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = to
}

type waits struct {
	clock  *steppingClock
	asked  chan time.Duration
	allows chan struct{}
}

func (w *waits) After(wait time.Duration) <-chan time.Time {
	fired := make(chan time.Time, 1)
	w.asked <- wait
	go func() {
		if _, ok := <-w.allows; ok {
			woke := w.clock.Now().Add(wait)
			w.clock.jump(woke)
			fired <- woke
		}
	}()
	return fired
}

func TestRun_RCC72_RunsEveryJobDueAtTheSameMinuteAndStopsWithItsContext(t *testing.T) {
	clock := &steppingClock{now: at(2026, 11, 29).Add(23 * time.Hour)}
	pending := &waits{clock: clock, asked: make(chan time.Duration, 4), allows: make(chan struct{})}
	var guard sync.Mutex
	var ran []string
	record := func(name string) scheduler.Work {
		return func(_ context.Context, now time.Time) error {
			guard.Lock()
			defer guard.Unlock()
			ran = append(ran, name+"@"+now.In(jakarta()).Format("2006-01-02 15:04"))
			return errors.New("logged and the loop goes on")
		}
	}
	runner := scheduler.New(scheduler.Options{
		Jobs: []scheduler.Job{
			{Name: "weekly", Hour: 0, Minute: 10, Days: scheduler.Mondays, Work: record("weekly")},
			{Name: "monthly", Hour: 0, Minute: 10, Days: scheduler.FirstOfMonth, Work: record("monthly")},
		},
		Clock: clock, Waits: pending, Location: jakarta(), Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan struct{})

	go func() {
		runner.Run(ctx)
		close(stopped)
	}()
	if wait := <-pending.asked; wait != 70*time.Minute {
		t.Fatalf("first wait = %s, want 70m until monday 30 november 00.10", wait)
	}
	pending.allows <- struct{}{}
	<-pending.asked
	cancel()

	<-stopped
	close(pending.allows)
	guard.Lock()
	defer guard.Unlock()
	if len(ran) != 1 || ran[0] != "weekly@2026-11-30 00:10" {
		t.Fatalf("ran = %v", ran)
	}
}

func TestRun_FirstOfMonthOnAMondayRunsBothJobs(t *testing.T) {
	clock := &steppingClock{now: at(2027, 2, 28).Add(23 * time.Hour)}
	pending := &waits{clock: clock, asked: make(chan time.Duration, 4), allows: make(chan struct{})}
	var guard sync.Mutex
	ran := map[string]bool{}
	record := func(name string) scheduler.Work {
		return func(context.Context, time.Time) error {
			guard.Lock()
			defer guard.Unlock()
			ran[name] = true
			return nil
		}
	}
	runner := scheduler.New(scheduler.Options{
		Jobs: []scheduler.Job{
			{Name: "weekly", Hour: 0, Minute: 10, Days: scheduler.Mondays, Work: record("weekly")},
			{Name: "monthly", Hour: 0, Minute: 10, Days: scheduler.FirstOfMonth, Work: record("monthly")},
		},
		Clock: clock, Waits: pending, Location: jakarta(), Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan struct{})

	go func() {
		runner.Run(ctx)
		close(stopped)
	}()
	<-pending.asked
	pending.allows <- struct{}{}
	<-pending.asked
	cancel()
	<-stopped
	close(pending.allows)

	guard.Lock()
	defer guard.Unlock()
	if !ran["weekly"] || !ran["monthly"] {
		t.Fatalf("ran = %v, want both jobs on monday 1 march 2027", ran)
	}
}
