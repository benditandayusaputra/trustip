package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type RemindExpiringNeeds struct {
	deps Deps
}

func NewRemindExpiringNeeds(deps Deps) *RemindExpiringNeeds {
	return &RemindExpiringNeeds{deps: deps}
}

func (uc *RemindExpiringNeeds) Execute(ctx context.Context, now time.Time) (int, error) {
	reminded := 0
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		expiring, err := uc.deps.Interest.ClaimExpiring(ctx, domain.ReminderWindowAt(now))
		if err != nil {
			return fmt.Errorf("claim expiring needs: %w", err)
		}
		for _, need := range expiring {
			if err := uc.deps.Announcer.Announce(ctx, need.Notice()); err != nil {
				return fmt.Errorf("remind need %s: %w", need.NeedID, err)
			}
		}
		reminded = len(expiring)
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("remind expiring needs: %w", err)
	}
	return reminded, nil
}
