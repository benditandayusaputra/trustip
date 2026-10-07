package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type ListProvinces struct {
	regions RegionReader
}

func NewListProvinces(regions RegionReader) *ListProvinces {
	return &ListProvinces{regions: regions}
}

func (uc *ListProvinces) Execute(ctx context.Context) ([]domain.Province, error) {
	provinces, err := uc.regions.ListProvinces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list provinces: %w", err)
	}
	return provinces, nil
}
