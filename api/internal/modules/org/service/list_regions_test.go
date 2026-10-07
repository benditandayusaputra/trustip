package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

type fakeRegions struct{ asked []string }

func (f *fakeRegions) ListProvinces(context.Context) ([]domain.Province, error) {
	return []domain.Province{{Code: "33", Name: "Jawa Tengah"}}, nil
}

func (f *fakeRegions) ListCities(_ context.Context, provinceCode string) ([]domain.City, error) {
	f.asked = append(f.asked, provinceCode)
	return []domain.City{{Code: "33.74", Name: "Kota Semarang", ProvinceName: "Jawa Tengah"}}, nil
}

func TestListCities_MalformedProvinceCodeIsAFieldErrorWithoutQuerying(t *testing.T) {
	for _, code := range []string{"", "3", "333", "ab", "33.74"} {
		regions := &fakeRegions{}

		_, err := service.NewListCities(regions).Execute(t.Context(), code)

		fields := apperror.From(err).Fields()
		if len(fields) != 1 || fields[0].Field != "province_code" || len(regions.asked) != 0 {
			t.Errorf("%q: err = %v, asked = %v", code, err, regions.asked)
		}
	}
}

func TestListCities_WellFormedProvinceCodeIsLookedUp(t *testing.T) {
	regions := &fakeRegions{}

	cities, err := service.NewListCities(regions).Execute(t.Context(), "33")

	if err != nil || len(cities) != 1 || regions.asked[0] != "33" {
		t.Fatalf("cities = %+v, err = %v", cities, err)
	}
}

func TestListProvinces_ReturnsTheStoredList(t *testing.T) {
	provinces, err := service.NewListProvinces(&fakeRegions{}).Execute(t.Context())

	if err != nil || len(provinces) != 1 || provinces[0].Code != "33" {
		t.Fatalf("provinces = %+v, err = %v", provinces, err)
	}
}
