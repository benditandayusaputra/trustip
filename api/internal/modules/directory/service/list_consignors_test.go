package service

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

const (
	storeID   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	productID = "0192f0c4-0000-7000-8000-000000000101"
)

type fakeConsignors struct {
	viewer          domain.Viewer
	consignorCounts int
	productCounts   int
	queries         []domain.ConsignorQuery
}

func (f *fakeConsignors) GetViewer(_ context.Context, key domain.ViewerKey) (domain.Viewer, error) {
	viewer := f.viewer
	viewer.OrganizationID = key.OrganizationID
	return viewer, nil
}

func (f *fakeConsignors) CountConsignorsWithin(_ context.Context, _ domain.ConsignorQuery) (map[int]int, error) {
	f.consignorCounts++
	return map[int]int{10: 5}, nil
}

func (f *fakeConsignors) CountProductsWithin(_ context.Context, _ domain.ConsignorQuery) (map[int]int, error) {
	f.productCounts++
	return map[int]int{10: 1, 25: 7}, nil
}

func (f *fakeConsignors) ListConsignors(
	_ context.Context, query domain.ConsignorQuery,
) ([]domain.ConsignorCard, error) {
	f.queries = append(f.queries, query)
	return []domain.ConsignorCard{{ID: consignorID, Distance: domain.KnownDistance(2000)}}, nil
}

func (f *fakeConsignors) ListProducts(_ context.Context, query domain.ConsignorQuery) ([]domain.ProductCard, error) {
	f.queries = append(f.queries, query)
	return []domain.ProductCard{{ID: productID, Category: domain.Category{ID: snacksID}}}, nil
}

func storeOwner() actor.Actor {
	return actor.Actor{UserID: "u2", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Owner,
	}}
}

func locatedStore() domain.Viewer {
	return domain.Viewer{
		Type: actor.Store, Location: domain.Point{Latitude: -7, Longitude: 110.4}, HasLocation: true,
		Store: domain.StoreFacts{StoreType: "CANTEEN", AcceptedCategoryIDs: []string{snacksID}},
	}
}

func TestListConsignors_FRDIR02_StoreSeesScoredConsignors(t *testing.T) {
	directory := &fakeConsignors{viewer: locatedStore()}

	page, err := NewListConsignors(directory, directory, fixedClock{}).Execute(context.Background(), storeOwner(),
		ListRequest[domain.ConsignorSearchForm]{})

	if err != nil || len(page.Items) != 1 || !page.Items[0].HasMatch || directory.consignorCounts != 1 ||
		directory.queries[0].ViewerLocation.Latitude != -7 {
		t.Fatalf("page = %+v, queries = %+v, err = %v", page, directory.queries, err)
	}
}

func TestListProducts_FRDIR02_BRDIR03_ProductsTabCountsSeekingProducts(t *testing.T) {
	directory := &fakeConsignors{viewer: locatedStore()}

	page, err := NewListProducts(directory, directory, fixedClock{}).Execute(context.Background(), storeOwner(),
		ListRequest[domain.ConsignorSearchForm]{})

	if err != nil || directory.productCounts != 1 || page.Context.EffectiveRadiusKm != 25 ||
		page.Items[0].Match.Score < 20 {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}

func TestListConsignors_FRDIR07_DeliveryFilterWithoutStoreLocationIsRejected(t *testing.T) {
	directory := &fakeConsignors{viewer: domain.Viewer{Type: actor.Store}}

	_, err := NewListConsignors(directory, directory, fixedClock{}).Execute(context.Background(), storeOwner(),
		ListRequest[domain.ConsignorSearchForm]{Search: domain.ConsignorSearchForm{DeliversToViewer: true}})

	if apperror.From(err).Code() != apperror.CodeValidationFailed || len(directory.queries) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestListProducts_BRAUTHZ01_NeedsAMembership(t *testing.T) {
	directory := &fakeConsignors{}

	_, err := NewListProducts(directory, directory, fixedClock{}).Execute(context.Background(), actor.Actor{},
		ListRequest[domain.ConsignorSearchForm]{})

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}
