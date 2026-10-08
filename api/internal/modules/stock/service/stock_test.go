package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/service"
)

const (
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
)

var errDown = errors.New("store down")

type fakeStocks struct {
	keys    []domain.PartnershipKey
	entries int
	fail    error
}

func (f *fakeStocks) ListPartnershipStock(_ context.Context, key domain.PartnershipKey) ([]domain.Product, error) {
	f.keys = append(f.keys, key)
	return []domain.Product{{ID: "bread", ShelfLife: stock.ShelfLifeOf(3), Batches: []domain.Batch{
		{ID: "b1", Expiry: stock.ExpiresOn(time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)), Balance: 4},
	}}}, f.fail
}

func (f *fakeStocks) ListOrganizationStock(context.Context, string) ([]domain.Product, error) {
	return []domain.Product{}, f.fail
}

func (f *fakeStocks) GetCardProduct(_ context.Context, key domain.CardKey) (domain.CardProduct, error) {
	return domain.CardProduct{ID: key.ProductID, Name: "Roti Manis"}, f.fail
}

func (f *fakeStocks) ListCard(_ context.Context, query domain.CardQuery) ([]domain.CardEntry, error) {
	entries := []domain.CardEntry{}
	for index := range min(f.entries, query.Limit) {
		entries = append(entries, domain.CardEntry{ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index)})
	}
	return entries, nil
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
}

func storeOwner() actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Owner,
	}}
}

func TestGetPartnershipStock_FRSTK05_FlagsBatchesOnTheServerDate(t *testing.T) {
	stocks := &fakeStocks{}

	views, err := service.NewGetPartnershipStock(stocks, fixedClock{}).Execute(t.Context(), storeOwner(), partnershipID)

	if err != nil || len(views) != 1 || views[0].Flag != stock.Expired || views[0].Balance != 4 {
		t.Fatalf("views = %+v, err = %v", views, err)
	}
	if stocks.keys[0] != (domain.PartnershipKey{OrganizationID: storeID, PartnershipID: partnershipID}) {
		t.Fatalf("key = %+v", stocks.keys[0])
	}
}

func TestGetPartnershipStock_FailureIsWrapped(t *testing.T) {
	_, err := service.NewGetPartnershipStock(&fakeStocks{fail: errDown}, fixedClock{}).
		Execute(t.Context(), storeOwner(), partnershipID)

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v", err)
	}
}

func TestStock_RSEC36_RequiresAMembership(t *testing.T) {
	stranger := actor.Actor{UserID: "u"}

	_, partnershipErr := service.NewGetPartnershipStock(&fakeStocks{}, fixedClock{}).Execute(t.Context(), stranger, "x")
	_, listErr := service.NewListStock(&fakeStocks{}, fixedClock{}).Execute(t.Context(), stranger)
	_, cardErr := service.NewGetStockCard(&fakeStocks{}).Execute(t.Context(), stranger, service.CardRequest{})

	for _, err := range []error{partnershipErr, listErr, cardErr} {
		if appErr, ok := errors.AsType[*apperror.Error](err); !ok || appErr.Code() != apperror.CodeForbidden {
			t.Fatalf("err = %v, want FORBIDDEN", err)
		}
	}
}

func TestGetStockCard_FRSTK04_PagesTheMovements(t *testing.T) {
	card := service.NewGetStockCard(&fakeStocks{entries: 3})

	page, err := card.Execute(t.Context(), storeOwner(), service.CardRequest{
		PartnershipID: partnershipID, ProductID: "bread", Limit: "2",
	})

	if err != nil || len(page.Entries) != 2 || !page.HasMore || page.NextCursor == "" ||
		page.Product.Name != "Roti Manis" {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}

func TestListStock_FRSTK02_ReturnsAnEmptyListWithoutStock(t *testing.T) {
	views, err := service.NewListStock(&fakeStocks{}, fixedClock{}).Execute(t.Context(), storeOwner())

	if err != nil || views == nil || len(views) != 0 {
		t.Fatalf("views = %+v, err = %v", views, err)
	}
}
