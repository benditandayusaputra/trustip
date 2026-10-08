package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type UnansweredConfirmer interface {
	ConfirmUnanswered(ctx context.Context, deadline domain.Deadline) (int, error)
}

type DeemConfirmed struct {
	drops UnansweredConfirmer
}

func NewDeemConfirmed(drops UnansweredConfirmer) *DeemConfirmed {
	return &DeemConfirmed{drops: drops}
}

func (uc *DeemConfirmed) Execute(ctx context.Context, now time.Time) (int, error) {
	confirmed, err := uc.drops.ConfirmUnanswered(ctx, domain.DeadlineAt(now))
	if err != nil {
		return 0, fmt.Errorf("deem unanswered drops confirmed: %w", err)
	}
	return confirmed, nil
}
