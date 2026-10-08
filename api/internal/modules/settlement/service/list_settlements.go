package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type ListRequest struct {
	Filter domain.ListFilter
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.ListItem
	NextCursor string
	HasMore    bool
}

type ListSettlements struct {
	settlements SettlementReader
	clock       Clock
}

func NewListSettlements(deps Deps) *ListSettlements {
	return &ListSettlements{settlements: deps.Settlements, clock: deps.Clock}
}

func (uc *ListSettlements) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return Page{}, fmt.Errorf("check settlement reader: %w", err)
	}
	now := uc.clock.Now()
	query, err := domain.NewListQuery(request.Filter, now)
	if err != nil {
		return Page{}, fmt.Errorf("read settlement filter: %w", err)
	}
	if query.After, err = pagination.ParseCursor(request.Cursor); err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	query.OrganizationID = signedIn.Membership.OrganizationID
	query.Limit = limit + 1
	found, err := uc.settlements.ListSettlements(ctx, query)
	if err != nil {
		return Page{}, fmt.Errorf("list settlements: %w", err)
	}
	for index, item := range found {
		found[index].Standing = domain.StandingOf(item.Status, item.DueDate, now)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items: found[:limit], HasMore: true,
		NextCursor: pagination.CursorOf(pagination.Position{At: last.IssuedAt, ID: last.ID}),
	}, nil
}
