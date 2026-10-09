package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	countrepository "github.com/benditandayusaputra/trustip/api/internal/modules/count/repository"
	countservice "github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	droprepository "github.com/benditandayusaputra/trustip/api/internal/modules/drop/repository"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	needrepository "github.com/benditandayusaputra/trustip/api/internal/modules/need/repository"
	needservice "github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
	proposalrepository "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/repository"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	settlementservice "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	trustrepository "github.com/benditandayusaputra/trustip/api/internal/modules/trust/repository"
	trustservice "github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/scheduler"
)

const (
	dailyMinute      = 5
	reputationMinute = 15
	settlementMinute = 10
	reminderHour     = 6
)

type jobRunners struct {
	deemDrops   *dropservice.DeemConfirmed
	deemCounts  *countservice.DeemConfirmed
	expire      *proposalservice.ExpireProposals
	settlements *settlementservice.IssueDueSettlements
	needs       *needservice.ExpireNeeds
	reminders   *needservice.RemindExpiringNeeds
	reputation  *trustservice.RefreshReputation
}

func buildJobRunners(base foundation, trade tradeComponents) jobRunners {
	pool := base.pool
	reminderBase := foundation{pool: pool, clock: clock.System{}, feed: base.feed}
	return jobRunners{
		deemDrops:   dropservice.NewDeemConfirmed(droprepository.NewDrops(pool)),
		deemCounts:  countservice.NewDeemConfirmed(countrepository.NewCounts(pool)),
		expire:      proposalservice.NewExpireProposals(proposalrepository.NewProposals(pool)),
		settlements: settlementservice.NewIssueDueSettlements(trade.settlements),
		needs:       needservice.NewExpireNeeds(needrepository.NewNeeds(pool)),
		reminders:   needservice.NewRemindExpiringNeeds(buildNeedDeps(reminderBase)),
		reputation:  trustservice.NewRefreshReputation(trustrepository.NewReputation(pool)),
	}
}

func scheduledJobs(runners jobRunners) []scheduler.Job {
	return []scheduler.Job{
		{Name: "deem-confirmed-and-expire", Minute: dailyMinute, Days: scheduler.EveryDay, Work: runners.daily},
		{
			Name: "weekly-settlements", Minute: settlementMinute, Days: scheduler.Mondays,
			Work: runners.issue(terms.ScheduleWeekly),
		},
		{
			Name: "monthly-settlements", Minute: settlementMinute, Days: scheduler.FirstOfMonth,
			Work: runners.issue(terms.ScheduleMonthly),
		},
		{Name: "need-reminders", Hour: reminderHour, Days: scheduler.EveryDay, Work: runners.remind},
		{Name: "reputation-summary", Minute: reputationMinute, Days: scheduler.EveryDay, Work: runners.summarize},
	}
}

func (r jobRunners) remind(ctx context.Context, now time.Time) error {
	if _, err := r.reminders.Execute(ctx, now); err != nil {
		return fmt.Errorf("remind expiring needs: %w", err)
	}
	return nil
}

func (r jobRunners) daily(ctx context.Context, now time.Time) error {
	_, dropErr := r.deemDrops.Execute(ctx, now)
	_, countErr := r.deemCounts.Execute(ctx, now)
	_, proposalErr := r.expire.Execute(ctx, now)
	_, needErr := r.needs.Execute(ctx, now)
	return errors.Join(dropErr, countErr, proposalErr, needErr)
}

func (r jobRunners) summarize(ctx context.Context, now time.Time) error {
	if _, err := r.reputation.Execute(ctx, now); err != nil {
		return fmt.Errorf("summarize reputation: %w", err)
	}
	return nil
}

func (r jobRunners) issue(schedule terms.Schedule) scheduler.Work {
	return func(ctx context.Context, now time.Time) error {
		if _, err := r.settlements.Execute(ctx, schedule, now); err != nil {
			return fmt.Errorf("issue %s settlements: %w", schedule, err)
		}
		return nil
	}
}

type schedulerParts struct {
	runners jobRunners
	clock   serverClock
	waits   scheduler.Waits
	log     *slog.Logger
}

func newScheduler(parts schedulerParts) *scheduler.Scheduler {
	return scheduler.New(scheduler.Options{
		Jobs: scheduledJobs(parts.runners), Clock: parts.clock, Waits: parts.waits, Location: period.Jakarta(),
		Log: parts.log,
	})
}
