package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type UnansweredConfirmer interface {
	ConfirmUnanswered(ctx context.Context, deadline domain.Deadline) (int, error)
}

type DeemConfirmed struct {
	counts UnansweredConfirmer
}

func NewDeemConfirmed(counts UnansweredConfirmer) *DeemConfirmed {
	return &DeemConfirmed{counts: counts}
}

func (uc *DeemConfirmed) Execute(ctx context.Context, now time.Time) (int, error) {
	confirmed, err := uc.counts.ConfirmUnanswered(ctx, domain.DeadlineAt(now))
	if err != nil {
		return 0, fmt.Errorf("deem unanswered counts confirmed: %w", err)
	}
	return confirmed, nil
}
