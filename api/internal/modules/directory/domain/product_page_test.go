package domain

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestNewProductPageQuery_WithoutPosition_HasNoOrigin(t *testing.T) {
	query, err := NewProductPageQuery(ProductPageForm{ConsignorSlug: "roti-mak-ijah", SKU: "ROT-0001"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if query.Origin.IsKnown || query.SKU != "ROT-0001" {
		t.Fatalf("query = %+v", query)
	}
}

func TestNewProductPageQuery_FRPUB05_VisitorPositionBecomesTheOrigin(t *testing.T) {
	query, err := NewProductPageQuery(ProductPageForm{
		ConsignorSlug: "roti-mak-ijah", SKU: "ROT-0001", Latitude: "-7.05", Longitude: "110.43",
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if !query.Origin.IsKnown || query.Origin.Point.Latitude != -7.05 || query.Origin.Point.Longitude != 110.43 {
		t.Fatalf("origin = %+v", query.Origin)
	}
}

func TestNewProductPageQuery_PositionOutOfRange_IsAValidationError(t *testing.T) {
	_, err := NewProductPageQuery(ProductPageForm{
		ConsignorSlug: "roti-mak-ijah", SKU: "ROT-0001", Latitude: "91", Longitude: "110",
	})

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestNewProductPageQuery_MalformedAddress_IsNotFound(t *testing.T) {
	for _, form := range []ProductPageForm{
		{ConsignorSlug: "Roti", SKU: "ROT-0001"},
		{ConsignorSlug: "roti", SKU: "rot-0001"},
		{ConsignorSlug: "roti", SKU: ""},
	} {
		if _, err := NewProductPageQuery(form); apperror.From(err).Code() != apperror.CodeNotFound {
			t.Errorf("form %+v: err = %v", form, err)
		}
	}
}
