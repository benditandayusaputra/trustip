package scheduler

import (
	"context"
	"log/slog"
	"time"
)

const daysToSearch = 62

type Work func(ctx context.Context, now time.Time) error

type DayRule func(day time.Time) bool

func EveryDay(time.Time) bool {
	return true
}

func Mondays(day time.Time) bool {
	return day.Weekday() == time.Monday
}

func FirstOfMonth(day time.Time) bool {
	return day.Day() == 1
}

type Job struct {
	Name   string
	Hour   int
	Minute int
	Days   DayRule
	Work   Work
}

type Clock interface {
	Now() time.Time
}

type Waits interface {
	After(wait time.Duration) <-chan time.Time
}

type Options struct {
	Jobs     []Job
	Clock    Clock
	Waits    Waits
	Location *time.Location
	Log      *slog.Logger
}

type Scheduler struct {
	options Options
}

func New(options Options) *Scheduler {
	return &Scheduler{options: options}
}

func NextRun(job Job, after time.Time, location *time.Location) time.Time {
	local := after.In(location)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	for range daysToSearch {
		candidate := time.Date(day.Year(), day.Month(), day.Day(), job.Hour, job.Minute, 0, 0, location)
		if job.Days(day) && candidate.After(after) {
			return candidate
		}
		day = day.AddDate(0, 0, 1)
	}
	return time.Time{}
}

func (s *Scheduler) Run(ctx context.Context) {
	after := s.options.Clock.Now()
	for {
		due, runAt := s.nextDue(after)
		if len(due) == 0 {
			<-ctx.Done()
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-s.options.Waits.After(runAt.Sub(s.options.Clock.Now())):
		}
		for _, job := range due {
			s.runJob(ctx, job, runAt)
		}
		after = runAt
	}
}

func (s *Scheduler) nextDue(after time.Time) ([]Job, time.Time) {
	var due []Job
	var earliest time.Time
	for _, job := range s.options.Jobs {
		next := NextRun(job, after, s.options.Location)
		switch {
		case earliest.IsZero() || next.Before(earliest):
			due, earliest = []Job{job}, next
		case next.Equal(earliest):
			due = append(due, job)
		}
	}
	return due, earliest
}

func (s *Scheduler) runJob(ctx context.Context, job Job, runAt time.Time) {
	started := s.options.Clock.Now()
	err := job.Work(ctx, runAt)
	duration := s.options.Clock.Now().Sub(started).Milliseconds()
	if err != nil {
		s.options.Log.Error("scheduled job failed", "job", job.Name, "at", runAt, "duration_ms", duration, "error", err)
		return
	}
	s.options.Log.Info("scheduled job finished", "job", job.Name, "at", runAt, "duration_ms", duration)
}

type SystemWaits struct{}

func (SystemWaits) After(wait time.Duration) <-chan time.Time {
	return time.After(wait)
}
