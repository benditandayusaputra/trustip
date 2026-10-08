package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type DropLister interface {
	ListDrops(ctx context.Context, query domain.PageQuery) ([]domain.Summary, error)
}

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

type ListDrops struct {
	drops DropLister
	clock Clock
}

func NewListDrops(drops DropLister, clock Clock) *ListDrops {
	return &ListDrops{drops: drops, clock: clock}
}

func (uc *ListDrops) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return Page{}, fmt.Errorf("check drop reader: %w", err)
	}
	query, err := domain.NewListQuery(request.Filter, uc.clock.Now())
	if err != nil {
		return Page{}, fmt.Errorf("read drop filter: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	query.OrganizationID = signedIn.Membership.OrganizationID
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.drops.ListDrops(ctx, domain.PageQuery{ListQuery: query, After: after, Limit: limit + 1})
	if err != nil {
		return Page{}, fmt.Errorf("list drops: %w", err)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items:      found[:limit],
		NextCursor: pagination.CursorOf(pagination.Position{At: last.OccurredAt, ID: last.ID}),
		HasMore:    true,
	}, nil
}
