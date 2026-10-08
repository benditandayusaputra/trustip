package service

import (
	"context"
	"fmt"
	"time"
)

type ProposalExpirer interface {
	ExpireOpen(ctx context.Context, now time.Time) (int, error)
}

type ExpireProposals struct {
	proposals ProposalExpirer
}

func NewExpireProposals(proposals ProposalExpirer) *ExpireProposals {
	return &ExpireProposals{proposals: proposals}
}

func (uc *ExpireProposals) Execute(ctx context.Context, now time.Time) (int, error) {
	expired, err := uc.proposals.ExpireOpen(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("mark expired proposals: %w", err)
	}
	return expired, nil
}
