package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

const (
	consignorID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	snacksID    = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"
)

var errDatabaseDown = errors.New("database down")

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)
}

type fakeDirectory struct {
	viewer    domain.Viewer
	viewerErr error
	counts    map[int]int
	cards     []domain.StoreCard
	counted   int
	queries   []domain.StoreQuery
}

func (f *fakeDirectory) GetViewer(_ context.Context, key domain.ViewerKey) (domain.Viewer, error) {
	viewer := f.viewer
	viewer.OrganizationID = key.OrganizationID
	viewer.HasCity = key.CityCode == "33.74"
	return viewer, f.viewerErr
}

func (f *fakeDirectory) CountStoresWithin(_ context.Context, _ domain.StoreQuery) (map[int]int, error) {
	f.counted++
	return f.counts, nil
}

func (f *fakeDirectory) ListStores(_ context.Context, query domain.StoreQuery) ([]domain.StoreCard, error) {
	f.queries = append(f.queries, query)
	return f.cards[:min(len(f.cards), query.Limit)], nil
}

func consignorOwner() actor.Actor {
	return actor.Actor{UserID: "u1", Membership: actor.Membership{
		OrganizationID: consignorID, OrganizationType: actor.Consignor, Role: actor.Owner,
	}}
}

func locatedConsignor() domain.Viewer {
	return domain.Viewer{
		Type: actor.Consignor, Location: domain.Point{Latitude: -7, Longitude: 110.4}, HasLocation: true,
		Consignor: domain.ConsignorFacts{Categories: []domain.Category{{ID: snacksID, Name: "Kue kering dan camilan"}}},
	}
}

func cardsNumbered(count int) []domain.StoreCard {
	cards := make([]domain.StoreCard, 0, count)
	for index := range count {
		id := fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index)
		cards = append(cards, domain.StoreCard{
			ID: id, Distance: domain.KnownDistance(index * 1000),
			Position: domain.Position{Number: float64(index), ID: id},
		})
	}
	return cards
}

func TestListStores_BRAUTHZ01_NeedsAMembership(t *testing.T) {
	empty := &fakeDirectory{}
	_, err := NewListStores(empty, empty, fixedClock{}).Execute(context.Background(), actor.Actor{UserID: "u"},
		ListRequest[domain.StoreSearchForm]{})

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestListStores_InvalidSearchIsRejectedBeforeReading(t *testing.T) {
	directory := &fakeDirectory{}

	_, err := NewListStores(directory, directory, fixedClock{}).Execute(context.Background(), consignorOwner(),
		ListRequest[domain.StoreSearchForm]{Search: domain.StoreSearchForm{Search: domain.SearchForm{Radius: "3"}}})

	if apperror.From(err).Code() != apperror.CodeValidationFailed || len(directory.queries) != 0 {
		t.Fatalf("err = %v, queries = %d", err, len(directory.queries))
	}
}

func TestListStores_ViewerFailureIsReturned(t *testing.T) {
	directory := &fakeDirectory{viewerErr: errDatabaseDown}

	_, err := listAsConsignor(directory)

	if !errors.Is(err, errDatabaseDown) {
		t.Fatalf("err = %v", err)
	}
}

func TestListStores_BRDIR03_FewNearbyStoresExpandTheRadius(t *testing.T) {
	directory := &fakeDirectory{viewer: locatedConsignor(), counts: map[int]int{10: 4, 25: 6}, cards: cardsNumbered(3)}

	page, err := listAsConsignor(directory)

	if err != nil || !page.Context.IsExpanded() || page.Context.EffectiveRadiusKm != 25 ||
		directory.queries[0].RadiusKm != 25 || directory.queries[0].Sort != domain.Nearest {
		t.Fatalf("page = %+v, queries = %+v, err = %v", page.Context, directory.queries, err)
	}
}

func TestListStores_BRDIR06_DefaultCategoriesComeFromTheConsignorsProducts(t *testing.T) {
	directory := &fakeDirectory{viewer: locatedConsignor(), counts: map[int]int{10: 5}}

	page, _ := listAsConsignor(directory)

	if len(page.Context.CategoryIDs) != 1 || directory.queries[0].Filters.CategoryIDs[0] != snacksID {
		t.Fatalf("context = %+v", page.Context)
	}
}

func TestListStores_BRDIR05_TwentyPerPageWithACursorThatKeepsTheRadius(t *testing.T) {
	directory := &fakeDirectory{viewer: locatedConsignor(), counts: map[int]int{10: 30}, cards: cardsNumbered(30)}
	lister := NewListStores(directory, directory, fixedClock{})

	first, _ := lister.Execute(context.Background(), consignorOwner(), ListRequest[domain.StoreSearchForm]{})
	next := ListRequest[domain.StoreSearchForm]{Cursor: first.NextCursor}
	_, err := lister.Execute(context.Background(), consignorOwner(), next)

	if err != nil || len(first.Items) != 20 || !first.HasMore || directory.counted != 1 ||
		directory.queries[1].After.ID != first.Items[19].ID || directory.queries[1].RadiusKm != 10 {
		t.Fatalf("first = %d %v, counted = %d, second query = %+v, err = %v",
			len(first.Items), first.HasMore, directory.counted, directory.queries[1], err)
	}
}

func TestListStores_BRDIR04_WithoutALocationStoresAreSortedByCity(t *testing.T) {
	directory := &fakeDirectory{viewer: domain.Viewer{Type: actor.Consignor}}

	page, _ := listAsConsignor(directory)

	if page.Context.Origin.IsKnown || directory.queries[0].Sort != domain.ByCity || directory.counted != 0 ||
		directory.queries[0].RadiusKm != domain.EveryDistance {
		t.Fatalf("context = %+v, query = %+v", page.Context, directory.queries[0])
	}
}

func TestListStores_FRDIR04_CityOriginMustBeAKnownCity(t *testing.T) {
	empty := &fakeDirectory{}
	_, err := NewListStores(empty, empty, fixedClock{}).Execute(context.Background(), consignorOwner(),
		ListRequest[domain.StoreSearchForm]{Search: domain.StoreSearchForm{Search: domain.SearchForm{
			Origin: domain.OriginForm{Kind: "CITY", CityCode: "99.99"},
		}}})

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestListStores_FRDIR08_BestMatchRanksTheNearestCandidatesByScore(t *testing.T) {
	cards := cardsNumbered(3)
	cards[2].AcceptedCategories = []domain.Category{{ID: snacksID}}
	directory := &fakeDirectory{viewer: locatedConsignor(), counts: map[int]int{10: 5}, cards: cards}

	page, err := NewListStores(directory, directory, fixedClock{}).Execute(context.Background(), consignorOwner(),
		bestMatchRequest())

	if err != nil || page.Items[0].ID != cards[2].ID || !page.Items[0].HasMatch ||
		directory.queries[0].Limit != domain.MatchCandidateLimit || directory.queries[0].Sort != domain.Nearest {
		t.Fatalf("items = %+v, query = %+v, err = %v", page.Items, directory.queries[0], err)
	}
}

func TestListStores_FRDIR08_BestMatchPagesThroughTheRanking(t *testing.T) {
	directory := &fakeDirectory{viewer: locatedConsignor(), counts: map[int]int{10: 25}, cards: cardsNumbered(25)}
	lister := NewListStores(directory, directory, fixedClock{})
	request := bestMatchRequest()

	first, _ := lister.Execute(context.Background(), consignorOwner(), request)
	request.Cursor = first.NextCursor
	second, err := lister.Execute(context.Background(), consignorOwner(), request)

	if err != nil || len(first.Items) != 20 || len(second.Items) != 5 || second.HasMore ||
		second.Items[0].ID == first.Items[0].ID {
		t.Fatalf("first = %d, second = %d %v, err = %v", len(first.Items), len(second.Items), second.HasMore, err)
	}
}

func listAsConsignor(directory *fakeDirectory) (Page[domain.StoreCard], error) {
	lister := NewListStores(directory, directory, fixedClock{})
	return lister.Execute(context.Background(), consignorOwner(), ListRequest[domain.StoreSearchForm]{})
}

func bestMatchRequest() ListRequest[domain.StoreSearchForm] {
	return ListRequest[domain.StoreSearchForm]{
		Search: domain.StoreSearchForm{Search: domain.SearchForm{Sort: "BEST_MATCH"}},
	}
}
