package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
)

type RecapReader interface {
	DailyRows(ctx context.Context, query domain.DailyQuery) ([]domain.RecapRow, error)
}

type Clock interface {
	Now() time.Time
}

type GetDailyRecap struct {
	recaps RecapReader
	clock  Clock
}

func NewGetDailyRecap(recaps RecapReader, clock Clock) *GetDailyRecap {
	return &GetDailyRecap{recaps: recaps, clock: clock}
}

func (uc *GetDailyRecap) Execute(ctx context.Context, signedIn actor.Actor, day string) (domain.DailyRecap, error) {
	if err := policy.Require(signedIn, policy.ViewReports); err != nil {
		return domain.DailyRecap{}, fmt.Errorf("check recap reader: %w", err)
	}
	if signedIn.Membership.OrganizationType != actor.Store {
		return domain.DailyRecap{}, apperror.New(apperror.CodeForbidden)
	}
	query, err := domain.NewDailyQuery(day, uc.clock.Now())
	if err != nil {
		return domain.DailyRecap{}, fmt.Errorf("read recap date: %w", err)
	}
	query.OrganizationID = signedIn.Membership.OrganizationID
	rows, err := uc.recaps.DailyRows(ctx, query)
	if err != nil {
		return domain.DailyRecap{}, fmt.Errorf("read recap rows: %w", err)
	}
	return domain.DailyRecap{Query: query, Rows: rows, Total: domain.SummarizeRecap(rows)}, nil
}
