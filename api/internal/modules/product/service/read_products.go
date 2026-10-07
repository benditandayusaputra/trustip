package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type GetProduct struct {
	reader ProductReader
}

func NewGetProduct(reader ProductReader) *GetProduct {
	return &GetProduct{reader: reader}
}

func (uc *GetProduct) Execute(ctx context.Context, signedIn actor.Actor, productID string) (domain.Product, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return domain.Product{}, fmt.Errorf("check product reader: %w", err)
	}
	product, err := uc.reader.GetProduct(ctx, keyOf(signedIn, productID))
	if err != nil {
		return domain.Product{}, fmt.Errorf("get product %s: %w", productID, err)
	}
	return product, nil
}

type ListRequest struct {
	Search domain.SearchForm
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Product
	NextCursor string
	HasMore    bool
}

type ListProducts struct {
	reader ProductReader
}

func NewListProducts(reader ProductReader) *ListProducts {
	return &ListProducts{reader: reader}
}

func (uc *ListProducts) Execute(ctx context.Context, signedIn actor.Actor, request ListRequest) (Page, error) {
	if err := policy.Require(signedIn, policy.ManageProducts); err != nil {
		return Page{}, fmt.Errorf("check product reader: %w", err)
	}
	search, err := domain.NewSearch(request.Search)
	if err != nil {
		return Page{}, fmt.Errorf("read search: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.reader.ListProducts(ctx, domain.PageQuery{
		ConsignorID: signedIn.Membership.OrganizationID, Search: search, After: after, Limit: limit + 1,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list products: %w", err)
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
