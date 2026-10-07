package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

var provinceCodeFormat = regexp.MustCompile(`^[0-9]{2}$`)

type ListCities struct {
	regions RegionReader
}

func NewListCities(regions RegionReader) *ListCities {
	return &ListCities{regions: regions}
}

func (uc *ListCities) Execute(ctx context.Context, provinceCode string) ([]domain.City, error) {
	if !provinceCodeFormat.MatchString(provinceCode) {
		return nil, apperror.Validation(apperror.InvalidField("province_code", apperror.FieldUnknownChoice))
	}
	cities, err := uc.regions.ListCities(ctx, provinceCode)
	if err != nil {
		return nil, fmt.Errorf("list cities in %s: %w", provinceCode, err)
	}
	return cities, nil
}
