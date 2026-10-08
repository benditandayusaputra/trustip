package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type StoreListRequest struct {
	Search domain.StoreSearchForm
	Cursor string
	Limit  string
}

type StorePage struct {
	Items      []domain.StoreCard
	NextCursor string
	HasMore    bool
	Context    domain.SearchContext
}

type ListStores struct {
	directory StoreDirectory
	clock     Clock
}

func NewListStores(directory StoreDirectory, clock Clock) *ListStores {
	return &ListStores{directory: directory, clock: clock}
}

func (uc *ListStores) Execute(ctx context.Context, signedIn actor.Actor, request StoreListRequest) (StorePage, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return StorePage{}, fmt.Errorf("check directory reader: %w", err)
	}
	search, err := domain.NewStoreSearch(request.Search)
	if err != nil {
		return StorePage{}, fmt.Errorf("read store search: %w", err)
	}
	after, err := domain.ParseCursor(request.Cursor)
	if err != nil {
		return StorePage{}, fmt.Errorf("read cursor: %w", err)
	}
	viewer, err := uc.directory.GetViewer(ctx, domain.ViewerKey{
		OrganizationID: signedIn.Membership.OrganizationID, CityCode: search.Origin.CityCode,
	})
	if err != nil {
		return StorePage{}, fmt.Errorf("read directory viewer: %w", err)
	}
	query, searchContext, err := uc.prepare(ctx, search, viewer, after)
	if err != nil {
		return StorePage{}, err
	}
	page, err := uc.fetch(ctx, viewer, query, pagination.ParseLimit(request.Limit))
	if err != nil {
		return StorePage{}, err
	}
	page.Context = searchContext
	return page, nil
}

func (uc *ListStores) prepare(
	ctx context.Context, search domain.StoreSearch, viewer domain.Viewer, after domain.Position,
) (domain.StoreQuery, domain.SearchContext, error) {
	origin, err := search.Origin.ResolveFor(viewer)
	if err != nil {
		return domain.StoreQuery{}, domain.SearchContext{}, fmt.Errorf("resolve origin: %w", err)
	}
	filters := search.Filters
	if search.UsesOwnCategories {
		filters.CategoryIDs = viewer.DefaultCategoryIDs()
	}
	query := domain.StoreQuery{
		ViewerID: viewer.OrganizationID, Origin: origin, RadiusKm: domain.EveryDistance, Filters: filters,
		Sort: search.SortFor(origin), After: after,
	}
	if origin.IsKnown {
		query.RadiusKm, err = uc.effectiveRadius(ctx, query, search.RadiusKm)
	}
	return query, domain.SearchContext{
		Origin: origin, Sort: query.Sort, SelectedRadiusKm: search.RadiusKm, EffectiveRadiusKm: query.RadiusKm,
		CategoryIDs: filters.CategoryIDs,
	}, err
}

func (uc *ListStores) effectiveRadius(ctx context.Context, query domain.StoreQuery, selectedKm int) (int, error) {
	if !query.After.IsStart() {
		return query.After.RadiusKm, nil
	}
	if selectedKm == domain.EveryDistance {
		return domain.EveryDistance, nil
	}
	counts, err := uc.directory.CountStoresWithin(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("count stores by radius: %w", err)
	}
	return domain.EffectiveRadius(selectedKm, counts), nil
}

func (uc *ListStores) fetch(
	ctx context.Context, viewer domain.Viewer, query domain.StoreQuery, limit int,
) (StorePage, error) {
	if query.Sort == domain.BestMatch {
		return uc.fetchBestMatches(ctx, viewer, query, limit)
	}
	query.Limit = limit + 1
	cards, err := uc.directory.ListStores(ctx, query)
	if err != nil {
		return StorePage{}, fmt.Errorf("list stores: %w", err)
	}
	cards = uc.scoreAll(viewer, cards)
	if len(cards) <= limit {
		return StorePage{Items: cards}, nil
	}
	last := cards[limit-1].Position
	last.RadiusKm = query.RadiusKm
	return StorePage{Items: cards[:limit], NextCursor: domain.CursorOf(last), HasMore: true}, nil
}

func (uc *ListStores) fetchBestMatches(
	ctx context.Context, viewer domain.Viewer, query domain.StoreQuery, limit int,
) (StorePage, error) {
	candidates := query
	candidates.Sort, candidates.After, candidates.Limit = domain.Nearest, domain.Position{}, domain.MatchCandidateLimit
	if !query.Origin.IsKnown {
		candidates.Sort = domain.ByCity
	}
	cards, err := uc.directory.ListStores(ctx, candidates)
	if err != nil {
		return StorePage{}, fmt.Errorf("list match candidates: %w", err)
	}
	items, hasMore := domain.PageAfter(domain.RankByMatch(uc.scoreAll(viewer, cards)), query.After.ID, limit)
	if !hasMore {
		return StorePage{Items: items}, nil
	}
	last := domain.Position{ID: items[len(items)-1].ID, RadiusKm: query.RadiusKm}
	return StorePage{Items: items, NextCursor: domain.CursorOf(last), HasMore: true}, nil
}

func (uc *ListStores) scoreAll(viewer domain.Viewer, cards []domain.StoreCard) []domain.StoreCard {
	now := uc.clock.Now()
	scored := make([]domain.StoreCard, 0, len(cards))
	for _, card := range cards {
		scored = append(scored, viewer.MatchStore(card, now))
	}
	return scored
}
