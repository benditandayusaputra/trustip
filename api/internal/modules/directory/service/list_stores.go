package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ListStores struct {
	viewers ViewerReader
	stores  StoreDirectory
	clock   Clock
}

func NewListStores(viewers ViewerReader, stores StoreDirectory, clock Clock) *ListStores {
	return &ListStores{viewers: viewers, stores: stores, clock: clock}
}

func (uc *ListStores) Execute(
	ctx context.Context, signedIn actor.Actor, request ListRequest[domain.StoreSearchForm],
) (Page[domain.StoreCard], error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page[domain.StoreCard]{}, fmt.Errorf("check directory reader: %w", err)
	}
	search, err := domain.NewStoreSearch(request.Search)
	if err != nil {
		return Page[domain.StoreCard]{}, fmt.Errorf("read store search: %w", err)
	}
	browsing, err := startBrowsing(ctx, uc.viewers, signedIn, ListRequest[domain.SearchForm]{
		Search: request.Search.Search, Cursor: request.Cursor, Limit: request.Limit,
	})
	if err != nil {
		return Page[domain.StoreCard]{}, err
	}
	search = withDefaultCategories(search, browsing.viewer)
	filters := search.Filters
	now := uc.clock.Now()
	page, err := browseFrom(ctx, browsing, search.Search, source[domain.StoreCard]{
		count: func(ctx context.Context, listing domain.Listing) (map[int]int, error) {
			return uc.stores.CountStoresWithin(ctx, domain.StoreQuery{Listing: listing, Filters: filters})
		},
		list: func(ctx context.Context, listing domain.Listing) ([]domain.StoreCard, error) {
			return uc.stores.ListStores(ctx, domain.StoreQuery{Listing: listing, Filters: filters})
		},
		score: func(card domain.StoreCard) domain.StoreCard { return browsing.viewer.MatchStore(card, now) },
	})
	if err != nil {
		return Page[domain.StoreCard]{}, fmt.Errorf("list stores: %w", err)
	}
	return page, nil
}

func withDefaultCategories(search domain.StoreSearch, viewer domain.Viewer) domain.StoreSearch {
	if search.UsesOwnCategories {
		search.CategoryIDs = viewer.DefaultCategoryIDs()
		search.Filters.CategoryIDs = search.CategoryIDs
	}
	return search
}
