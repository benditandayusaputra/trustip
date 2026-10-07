package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

type ListRequest struct {
	Box    string
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Summary
	NextCursor string
	HasMore    bool
}

type ListProposals struct {
	lister ProposalLister
}

func NewListProposals(lister ProposalLister) *ListProposals {
	return &ListProposals{lister: lister}
}

func (uc *ListProposals) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page{}, fmt.Errorf("check proposal reader: %w", err)
	}
	box, err := domain.ParseBox(request.Box)
	if err != nil {
		return Page{}, fmt.Errorf("read box: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.lister.ListProposals(ctx, domain.ListQuery{
		OrganizationID: signedIn.Membership.OrganizationID, Box: box, After: after, Limit: limit + 1,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list proposals: %w", err)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items:      found[:limit],
		NextCursor: pagination.CursorOf(pagination.Position{At: last.CreatedAt, ID: last.ID}),
		HasMore:    true,
	}, nil
}
