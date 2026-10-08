package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ListRequest[F any] struct {
	Search F
	Cursor string
	Limit  string
}

type Page[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
	Context    domain.SearchContext
}

type counter func(ctx context.Context, listing domain.Listing) (map[int]int, error)

type lister[T domain.Listed] func(ctx context.Context, listing domain.Listing) ([]T, error)

type source[T domain.Listed] struct {
	count counter
	list  lister[T]
	score func(T) T
}

type browse struct {
	viewer domain.Viewer
	after  domain.Position
	limit  int
}

func startBrowsing(
	ctx context.Context, viewers ViewerReader, signedIn actor.Actor, request ListRequest[domain.SearchForm],
) (browse, error) {
	after, err := domain.ParseCursor(request.Cursor)
	if err != nil {
		return browse{}, fmt.Errorf("read cursor: %w", err)
	}
	viewer, err := viewers.GetViewer(ctx, domain.ViewerKey{
		OrganizationID: signedIn.Membership.OrganizationID, CityCode: request.Search.Origin.CityCode,
	})
	if err != nil {
		return browse{}, fmt.Errorf("read directory viewer: %w", err)
	}
	return browse{viewer: viewer, after: after, limit: pagination.ParseLimit(request.Limit)}, nil
}

func (b browse) resolve(ctx context.Context, search domain.Search, count counter) (domain.Listing, error) {
	origin, err := search.Origin.ResolveFor(b.viewer)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("resolve origin: %w", err)
	}
	listing := domain.Listing{
		ViewerID: b.viewer.OrganizationID, Origin: origin, RadiusKm: domain.EveryDistance,
		Sort: search.SortFor(origin), After: b.after,
	}
	if !origin.IsKnown {
		return listing, nil
	}
	listing.RadiusKm, err = effectiveRadius(ctx, count, listing, search.RadiusKm)
	return listing, err
}

func effectiveRadius(ctx context.Context, count counter, listing domain.Listing, selectedKm int) (int, error) {
	if !listing.After.IsStart() {
		return listing.After.RadiusKm, nil
	}
	if selectedKm == domain.EveryDistance {
		return domain.EveryDistance, nil
	}
	counts, err := count(ctx, listing)
	if err != nil {
		return 0, fmt.Errorf("count results by radius: %w", err)
	}
	return domain.EffectiveRadius(selectedKm, counts), nil
}

func browseFrom[T domain.Listed](
	ctx context.Context, browsing browse, search domain.Search, from source[T],
) (Page[T], error) {
	listing, err := browsing.resolve(ctx, search, from.count)
	if err != nil {
		return Page[T]{}, err
	}
	page, err := fetchPage(ctx, from, listing, browsing.limit)
	if err != nil {
		return Page[T]{}, err
	}
	page.Context = domain.SearchContext{
		Origin: listing.Origin, Sort: listing.Sort, SelectedRadiusKm: search.RadiusKm,
		EffectiveRadiusKm: listing.RadiusKm, CategoryIDs: search.CategoryIDs,
	}
	return page, nil
}

func fetchPage[T domain.Listed](
	ctx context.Context, from source[T], listing domain.Listing, limit int,
) (Page[T], error) {
	if listing.Sort == domain.BestMatch {
		return fetchBestMatches(ctx, from, listing, limit)
	}
	listing.Limit = limit + 1
	cards, err := from.list(ctx, listing)
	if err != nil {
		return Page[T]{}, fmt.Errorf("list directory: %w", err)
	}
	cards = scoreAll(cards, from.score)
	if len(cards) <= limit {
		return Page[T]{Items: cards}, nil
	}
	last := cards[limit-1].SortPosition()
	last.RadiusKm = listing.RadiusKm
	return Page[T]{Items: cards[:limit], NextCursor: domain.CursorOf(last), HasMore: true}, nil
}

func fetchBestMatches[T domain.Listed](
	ctx context.Context, from source[T], listing domain.Listing, limit int,
) (Page[T], error) {
	candidates := listing
	candidates.Sort, candidates.After, candidates.Limit = domain.Nearest, domain.Position{}, domain.MatchCandidateLimit
	if !listing.Origin.IsKnown {
		candidates.Sort = domain.ByCity
	}
	cards, err := from.list(ctx, candidates)
	if err != nil {
		return Page[T]{}, fmt.Errorf("list match candidates: %w", err)
	}
	items, hasMore := domain.PageAfter(domain.RankByMatch(scoreAll(cards, from.score)), listing.After.ID, limit)
	if !hasMore {
		return Page[T]{Items: items}, nil
	}
	last := domain.Position{ID: items[len(items)-1].SortPosition().ID, RadiusKm: listing.RadiusKm}
	return Page[T]{Items: items, NextCursor: domain.CursorOf(last), HasMore: true}, nil
}

func scoreAll[T any](cards []T, score func(T) T) []T {
	scored := make([]T, 0, len(cards))
	for _, card := range cards {
		scored = append(scored, score(card))
	}
	return scored
}
