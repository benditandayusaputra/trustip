package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type consignorBrowse struct {
	browse
	search domain.ConsignorSearch
}

func (b consignorBrowse) query(listing domain.Listing) domain.ConsignorQuery {
	return domain.ConsignorQuery{Listing: listing, Filters: b.search.Filters, ViewerLocation: b.viewer.Location}
}

func startConsignorBrowsing(
	ctx context.Context, viewers ViewerReader, signedIn actor.Actor, request ListRequest[domain.ConsignorSearchForm],
) (consignorBrowse, error) {
	search, err := domain.NewConsignorSearch(request.Search)
	if err != nil {
		return consignorBrowse{}, fmt.Errorf("read consignor search: %w", err)
	}
	browsing, err := startBrowsing(ctx, viewers, signedIn, ListRequest[domain.SearchForm]{
		Search: request.Search.Search, Cursor: request.Cursor, Limit: request.Limit,
	})
	if err != nil {
		return consignorBrowse{}, err
	}
	if err := search.CheckViewer(browsing.viewer); err != nil {
		return consignorBrowse{}, fmt.Errorf("check delivery filter: %w", err)
	}
	return consignorBrowse{browse: browsing, search: search}, nil
}

type consignorSource[T domain.Listed] func(browsing consignorBrowse, now time.Time) source[T]

type consignorAsk struct {
	signedIn actor.Actor
	request  ListRequest[domain.ConsignorSearchForm]
}

func browseConsignors[T domain.Listed](
	ctx context.Context, deps consignorDeps, ask consignorAsk, sourceOf consignorSource[T],
) (Page[T], error) {
	browsing, err := startConsignorBrowsing(ctx, deps.viewers, ask.signedIn, ask.request)
	if err != nil {
		return Page[T]{}, err
	}
	return browseFrom(ctx, browsing.browse, browsing.search.Search, sourceOf(browsing, deps.clock.Now()))
}

type consignorDeps struct {
	viewers    ViewerReader
	consignors ConsignorDirectory
	clock      Clock
}

type ListConsignors struct {
	deps consignorDeps
}

func NewListConsignors(viewers ViewerReader, consignors ConsignorDirectory, clock Clock) *ListConsignors {
	return &ListConsignors{deps: consignorDeps{viewers: viewers, consignors: consignors, clock: clock}}
}

func (uc *ListConsignors) Execute(
	ctx context.Context, signedIn actor.Actor, request ListRequest[domain.ConsignorSearchForm],
) (Page[domain.ConsignorCard], error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page[domain.ConsignorCard]{}, fmt.Errorf("check directory reader: %w", err)
	}
	page, err := browseConsignors(ctx, uc.deps, consignorAsk{signedIn: signedIn, request: request}, uc.sourceOf)
	if err != nil {
		return Page[domain.ConsignorCard]{}, fmt.Errorf("list consignors: %w", err)
	}
	return page, nil
}

func (uc *ListConsignors) sourceOf(browsing consignorBrowse, now time.Time) source[domain.ConsignorCard] {
	return source[domain.ConsignorCard]{
		count: func(ctx context.Context, listing domain.Listing) (map[int]int, error) {
			return uc.deps.consignors.CountConsignorsWithin(ctx, browsing.query(listing))
		},
		list: func(ctx context.Context, listing domain.Listing) ([]domain.ConsignorCard, error) {
			return uc.deps.consignors.ListConsignors(ctx, browsing.query(listing))
		},
		score: func(card domain.ConsignorCard) domain.ConsignorCard {
			return browsing.viewer.MatchConsignor(card, now)
		},
	}
}

type ListProducts struct {
	deps consignorDeps
}

func NewListProducts(viewers ViewerReader, consignors ConsignorDirectory, clock Clock) *ListProducts {
	return &ListProducts{deps: consignorDeps{viewers: viewers, consignors: consignors, clock: clock}}
}

func (uc *ListProducts) Execute(
	ctx context.Context, signedIn actor.Actor, request ListRequest[domain.ConsignorSearchForm],
) (Page[domain.ProductCard], error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page[domain.ProductCard]{}, fmt.Errorf("check directory reader: %w", err)
	}
	page, err := browseConsignors(ctx, uc.deps, consignorAsk{signedIn: signedIn, request: request}, uc.sourceOf)
	if err != nil {
		return Page[domain.ProductCard]{}, fmt.Errorf("list products: %w", err)
	}
	return page, nil
}

func (uc *ListProducts) sourceOf(browsing consignorBrowse, now time.Time) source[domain.ProductCard] {
	return source[domain.ProductCard]{
		count: func(ctx context.Context, listing domain.Listing) (map[int]int, error) {
			return uc.deps.consignors.CountProductsWithin(ctx, browsing.query(listing))
		},
		list: func(ctx context.Context, listing domain.Listing) ([]domain.ProductCard, error) {
			return uc.deps.consignors.ListProducts(ctx, browsing.query(listing))
		},
		score: func(card domain.ProductCard) domain.ProductCard {
			return browsing.viewer.MatchProduct(card, now)
		},
	}
}
