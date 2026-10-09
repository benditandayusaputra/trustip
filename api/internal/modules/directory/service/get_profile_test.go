package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/service"
)

type profileShelf struct {
	asked []string
}

func (s *profileShelf) GetStoreProfile(_ context.Context, slug string) (domain.StoreProfile, error) {
	s.asked = append(s.asked, slug)
	if slug != "warung-bu-tini" {
		return domain.StoreProfile{}, apperror.NotFound()
	}
	return domain.StoreProfile{Slug: slug, Name: "Warung Bu Tini"}, nil
}

func (s *profileShelf) GetConsignorProfile(_ context.Context, slug string) (domain.ConsignorProfile, error) {
	s.asked = append(s.asked, slug)
	if slug != "roti-mak-ijah" {
		return domain.ConsignorProfile{}, apperror.NotFound()
	}
	return domain.ConsignorProfile{Slug: slug, Name: "Roti Mak Ijah"}, nil
}

func isNotFound(err error) bool {
	return err != nil && apperror.From(err).Code() == apperror.CodeNotFound
}

func TestGetStoreProfile_MalformedSlug_IsNotFoundWithoutLookup(t *testing.T) {
	shelf := &profileShelf{}

	_, err := service.NewGetStoreProfile(shelf).Execute(t.Context(), "Warung/../x")

	if !isNotFound(err) || len(shelf.asked) != 0 {
		t.Fatalf("err = %v, asked = %v", err, shelf.asked)
	}
}

func TestGetStoreProfile_BRPUB02_HiddenStore_IsNotFound(t *testing.T) {
	_, err := service.NewGetStoreProfile(&profileShelf{}).Execute(t.Context(), "warung-tersembunyi")

	if !isNotFound(err) {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestGetStoreProfile_ListedStore_ReturnsProfile(t *testing.T) {
	profile, err := service.NewGetStoreProfile(&profileShelf{}).Execute(t.Context(), "warung-bu-tini")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if profile.Name != "Warung Bu Tini" {
		t.Fatalf("profile = %+v", profile)
	}
}

func TestGetConsignorProfile_MalformedSlug_IsNotFoundWithoutLookup(t *testing.T) {
	shelf := &profileShelf{}

	_, err := service.NewGetConsignorProfile(shelf).Execute(t.Context(), "ROTI")

	if !isNotFound(err) || len(shelf.asked) != 0 {
		t.Fatalf("err = %v, asked = %v", err, shelf.asked)
	}
}

func TestGetConsignorProfile_BRPUB02_HiddenConsignor_IsNotFound(t *testing.T) {
	_, err := service.NewGetConsignorProfile(&profileShelf{}).Execute(t.Context(), "roti-lain")

	if !isNotFound(err) {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestGetConsignorProfile_ListedConsignor_ReturnsProfile(t *testing.T) {
	profile, err := service.NewGetConsignorProfile(&profileShelf{}).Execute(t.Context(), "roti-mak-ijah")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if profile.Name != "Roti Mak Ijah" {
		t.Fatalf("profile = %+v", profile)
	}
}

func (s *profileShelf) GetProductPage(_ context.Context, query domain.ProductPageQuery) (domain.ProductPage, error) {
	s.asked = append(s.asked, query.SKU)
	return domain.ProductPage{Product: domain.ProductDetail{SKU: query.SKU}}, nil
}

func TestGetProductPage_MalformedSKU_IsNotFoundWithoutLookup(t *testing.T) {
	shelf := &profileShelf{}

	_, err := service.NewGetProductPage(shelf).Execute(t.Context(),
		domain.ProductPageForm{ConsignorSlug: "roti-mak-ijah", SKU: "rot 1"})

	if !isNotFound(err) || len(shelf.asked) != 0 {
		t.Fatalf("err = %v, asked = %v", err, shelf.asked)
	}
}

func TestGetProductPage_FRPUB04_KnownProduct_ReturnsThePage(t *testing.T) {
	page, err := service.NewGetProductPage(&profileShelf{}).Execute(t.Context(),
		domain.ProductPageForm{ConsignorSlug: "roti-mak-ijah", SKU: "ROT-0001"})
	if err != nil {
		t.Fatalf("get page: %v", err)
	}

	if page.Product.SKU != "ROT-0001" {
		t.Fatalf("page = %+v", page)
	}
}
