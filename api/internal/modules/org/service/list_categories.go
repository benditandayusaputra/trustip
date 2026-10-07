package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type ListCategories struct {
	categories CategoryReader
}

func NewListCategories(categories CategoryReader) *ListCategories {
	return &ListCategories{categories: categories}
}

func (uc *ListCategories) Execute(ctx context.Context) ([]domain.Category, error) {
	categories, err := uc.categories.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}
