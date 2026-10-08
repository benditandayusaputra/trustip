package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type ListRequest struct {
	Filter domain.ListFilter
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Summary
	NextCursor string
	HasMore    bool
}

type ListCounts struct {
	counts CountLister
	clock  Clock
}

func NewListCounts(deps Deps) *ListCounts {
	return &ListCounts{counts: deps.Counts, clock: deps.Clock}
}

func (uc *ListCounts) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return Page{}, fmt.Errorf("check count reader: %w", err)
	}
	query, err := domain.NewListQuery(request.Filter)
	if err != nil {
		return Page{}, fmt.Errorf("read count filter: %w", err)
	}
	if query.After, err = pagination.ParseCursor(request.Cursor); err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	query.OrganizationID = signedIn.Membership.OrganizationID
	query.ConfirmCutoff = uc.clock.Now().Add(-domain.ConfirmWindow)
	query.Limit = limit + 1
	found, err := uc.counts.ListCounts(ctx, query)
	if err != nil {
		return Page{}, fmt.Errorf("list counts: %w", err)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items: found[:limit], HasMore: true,
		NextCursor: pagination.CursorOf(pagination.Position{At: last.CountedAt, ID: last.ID}),
	}, nil
}
