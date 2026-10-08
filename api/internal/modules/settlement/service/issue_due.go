package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type IssueDueSettlements struct {
	deps   Deps
	issuer *Issuer
}

func NewIssueDueSettlements(deps Deps) *IssueDueSettlements {
	return &IssueDueSettlements{deps: deps, issuer: NewIssuer(deps)}
}

func (uc *IssueDueSettlements) Execute(ctx context.Context, schedule terms.Schedule, now time.Time) (int, error) {
	current, err := domain.PeriodContaining(schedule, now)
	if err != nil {
		return 0, fmt.Errorf("find the current %s period: %w", schedule, err)
	}
	partnershipIDs, err := uc.deps.Settlements.ListDue(ctx, domain.DueKey{
		Schedule: string(schedule), Before: current.Start,
	})
	if err != nil {
		return 0, fmt.Errorf("list partnerships due for %s settlements: %w", schedule, err)
	}
	issued := 0
	var failures []error
	for _, partnershipID := range partnershipIDs {
		count, err := uc.issueFor(ctx, partnershipID, schedule, now)
		issued += count
		if err != nil {
			failures = append(failures, err)
		}
	}
	return issued, errors.Join(failures...)
}

func (uc *IssueDueSettlements) issueFor(
	ctx context.Context, partnershipID string, schedule terms.Schedule, now time.Time,
) (int, error) {
	issued := 0
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		ledger, err := uc.deps.Settlements.LockLedger(ctx, partnershipID)
		if err != nil {
			return fmt.Errorf("lock ledger: %w", err)
		}
		if ledger.Schedule != schedule {
			return nil
		}
		drafts, err := ledger.DraftClosedPeriods(now)
		if err != nil {
			return fmt.Errorf("draft closed periods: %w", err)
		}
		for _, draft := range drafts {
			if _, err := uc.issuer.issue(ctx, ledger, draft); err != nil {
				return err
			}
			issued++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("issue %s settlements for %s: %w", schedule, partnershipID, err)
	}
	return issued, nil
}
