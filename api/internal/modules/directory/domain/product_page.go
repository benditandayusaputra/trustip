package domain

import (
	"regexp"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const PublicStoreLimit = 20

var skuFormat = regexp.MustCompile(`^[A-Z0-9]+(-[A-Z0-9]+)*$`)

type ProductPageForm struct {
	ConsignorSlug string
	SKU           string
	Latitude      string
	Longitude     string
}

type ProductPageQuery struct {
	ConsignorSlug string
	SKU           string
	Origin        Place
}

func NewProductPageQuery(form ProductPageForm) (ProductPageQuery, error) {
	if !IsSlug(form.ConsignorSlug) || !skuFormat.MatchString(form.SKU) {
		return ProductPageQuery{}, apperror.NotFound()
	}
	query := ProductPageQuery{ConsignorSlug: form.ConsignorSlug, SKU: form.SKU}
	if form.Latitude == "" && form.Longitude == "" {
		return query, nil
	}
	point, ok := readPoint(form.Latitude, form.Longitude)
	if !ok {
		return ProductPageQuery{}, apperror.Validation(apperror.InvalidField(originField, apperror.FieldOutOfRange))
	}
	query.Origin = Place{Point: point, IsKnown: true}
	return query, nil
}

type ProductDetail struct {
	ID                   string
	SKU                  string
	Name                 string
	Unit                 string
	Description          string
	Composition          string
	NetContent           string
	CategoryName         string
	SuggestedRetailPrice int64
	PhotoStorageKey      string
}

type ProductMaker struct {
	ProfileSlug string
	Name        string
	District    string
	CityName    string
	Legality    DeclaredLegality
}

type AvailableStore struct {
	ProfileSlug string
	Name        string
	Address     string
	District    string
	CityName    string
	Location    Place
	Distance    Distance
}

type ProductPage struct {
	Product ProductDetail
	Maker   ProductMaker
	Stores  []AvailableStore
}
