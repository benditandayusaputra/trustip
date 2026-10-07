package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type PartnershipReader interface {
	ListPartnerships(ctx context.Context, query domain.ListQuery) ([]domain.Summary, error)
	GetPartnership(ctx context.Context, key domain.Key) (domain.Partnership, error)
	ListCatalog(ctx context.Context, key domain.Key) ([]domain.CatalogItem, error)
}

func keyOf(signedIn actor.Actor, partnershipID string) domain.Key {
	return domain.Key{OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID}
}

type ListRequest struct {
	Status string
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Summary
	NextCursor string
	HasMore    bool
}

type ListPartnerships struct {
	reader PartnershipReader
}

func NewListPartnerships(reader PartnershipReader) *ListPartnerships {
	return &ListPartnerships{reader: reader}
}

func (uc *ListPartnerships) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page{}, fmt.Errorf("check partner reader: %w", err)
	}
	status, err := domain.ParseStatus(request.Status)
	if err != nil {
		return Page{}, fmt.Errorf("read status: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.reader.ListPartnerships(ctx, domain.ListQuery{
		OrganizationID: signedIn.Membership.OrganizationID, Status: status, After: after, Limit: limit + 1,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list partnerships: %w", err)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items:      found[:limit],
		NextCursor: pagination.CursorOf(pagination.Position{At: last.StartedAt, ID: last.ID}),
		HasMore:    true,
	}, nil
}

type GetPartnership struct {
	reader PartnershipReader
}

func NewGetPartnership(reader PartnershipReader) *GetPartnership {
	return &GetPartnership{reader: reader}
}

func (uc *GetPartnership) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) (domain.Partnership, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.Partnership{}, fmt.Errorf("check partner reader: %w", err)
	}
	partnership, err := uc.reader.GetPartnership(ctx, keyOf(signedIn, partnershipID))
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("get partnership %s: %w", partnershipID, err)
	}
	return partnership, nil
}

type ListCatalog struct {
	reader PartnershipReader
}

func NewListCatalog(reader PartnershipReader) *ListCatalog {
	return &ListCatalog{reader: reader}
}

func (uc *ListCatalog) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) ([]domain.CatalogItem, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check catalog reader: %w", err)
	}
	key := keyOf(signedIn, partnershipID)
	if _, err := uc.reader.GetPartnership(ctx, key); err != nil {
		return nil, fmt.Errorf("find partnership %s: %w", partnershipID, err)
	}
	items, err := uc.reader.ListCatalog(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("list catalog of %s: %w", partnershipID, err)
	}
	return items, nil
}
