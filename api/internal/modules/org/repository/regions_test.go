package repository_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	provincesInIndonesia = 38
	citiesInCentralJava  = 35
)

func TestListProvinces_EveryProvinceOnceInCodeOrder(t *testing.T) {
	regions := repository.NewRegions(dbtest.MigratedPool(t))

	provinces, err := regions.ListProvinces(t.Context())
	if err != nil {
		t.Fatalf("list provinces: %v", err)
	}

	isSorted := slices.IsSortedFunc(provinces, func(a, b domain.Province) int { return cmp.Compare(a.Code, b.Code) })
	aceh := domain.Province{Code: "11", Name: "Aceh"}
	if len(provinces) != provincesInIndonesia || !isSorted || provinces[0] != aceh {
		t.Fatalf("provinces = %d, first = %+v", len(provinces), provinces[0])
	}
}

func TestListCities_OneProvinceInNameOrder(t *testing.T) {
	regions := repository.NewRegions(dbtest.MigratedPool(t))

	cities, err := regions.ListCities(t.Context(), "33")
	if err != nil {
		t.Fatalf("list cities: %v", err)
	}

	semarang := domain.City{Code: semarangCode, Name: "Kota Semarang", ProvinceName: "Jawa Tengah"}
	isSorted := slices.IsSortedFunc(cities, func(a, b domain.City) int { return cmp.Compare(a.Name, b.Name) })
	if len(cities) != citiesInCentralJava || !isSorted || !slices.Contains(cities, semarang) {
		t.Fatalf("cities = %d %+v", len(cities), cities)
	}
}
