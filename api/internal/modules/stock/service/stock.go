package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
)

type StockReader interface {
	ListPartnershipStock(ctx context.Context, key domain.PartnershipKey) ([]domain.Product, error)
	ListOrganizationStock(ctx context.Context, organizationID string) ([]domain.Product, error)
	GetCardProduct(ctx context.Context, key domain.CardKey) (domain.CardProduct, error)
	ListCard(ctx context.Context, query domain.CardQuery) ([]domain.CardEntry, error)
}

type Clock interface {
	Now() time.Time
}

func viewsOn(products []domain.Product, now time.Time) []domain.ProductView {
	views := make([]domain.ProductView, 0, len(products))
	for _, product := range products {
		views = append(views, product.ViewOn(now))
	}
	return views
}

type GetPartnershipStock struct {
	stocks StockReader
	clock  Clock
}

func NewGetPartnershipStock(stocks StockReader, clock Clock) *GetPartnershipStock {
	return &GetPartnershipStock{stocks: stocks, clock: clock}
}

func (uc *GetPartnershipStock) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) ([]domain.ProductView, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check stock reader: %w", err)
	}
	products, err := uc.stocks.ListPartnershipStock(ctx, domain.PartnershipKey{
		OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return nil, fmt.Errorf("list stock of %s: %w", partnershipID, err)
	}
	return viewsOn(products, uc.clock.Now()), nil
}

type ListStock struct {
	stocks StockReader
	clock  Clock
}

func NewListStock(stocks StockReader, clock Clock) *ListStock {
	return &ListStock{stocks: stocks, clock: clock}
}

func (uc *ListStock) Execute(ctx context.Context, signedIn actor.Actor) ([]domain.ProductView, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check stock reader: %w", err)
	}
	products, err := uc.stocks.ListOrganizationStock(ctx, signedIn.Membership.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("list stock: %w", err)
	}
	return viewsOn(products, uc.clock.Now()), nil
}

type CardRequest struct {
	PartnershipID string
	ProductID     string
	Cursor        string
	Limit         string
}

type CardPage struct {
	Product    domain.CardProduct
	Entries    []domain.CardEntry
	NextCursor string
	HasMore    bool
}

type GetStockCard struct {
	stocks StockReader
}

func NewGetStockCard(stocks StockReader) *GetStockCard {
	return &GetStockCard{stocks: stocks}
}

func (uc *GetStockCard) Execute(ctx context.Context, signedIn actor.Actor, request CardRequest) (CardPage, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return CardPage{}, fmt.Errorf("check stock reader: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return CardPage{}, fmt.Errorf("read cursor: %w", err)
	}
	key := domain.CardKey{
		PartnershipKey: domain.PartnershipKey{
			OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: request.PartnershipID,
		},
		ProductID: request.ProductID,
	}
	product, err := uc.stocks.GetCardProduct(ctx, key)
	if err != nil {
		return CardPage{}, fmt.Errorf("get card product: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	entries, err := uc.stocks.ListCard(ctx, domain.CardQuery{CardKey: key, After: after, Limit: limit + 1})
	if err != nil {
		return CardPage{}, fmt.Errorf("list stock card: %w", err)
	}
	return pageOf(product, entries, limit), nil
}

func pageOf(product domain.CardProduct, entries []domain.CardEntry, limit int) CardPage {
	if len(entries) <= limit {
		return CardPage{Product: product, Entries: entries}
	}
	last := entries[limit-1]
	return CardPage{
		Product: product, Entries: entries[:limit], HasMore: true,
		NextCursor: pagination.CursorOf(pagination.Position{At: last.OccurredAt, ID: last.ID}),
	}
}
