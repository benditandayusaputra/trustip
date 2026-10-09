package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ProductPages interface {
	GetProductPage(ctx context.Context, query domain.ProductPageQuery) (domain.ProductPage, error)
}

type GetProductPage struct {
	pages ProductPages
}

func NewGetProductPage(pages ProductPages) *GetProductPage {
	return &GetProductPage{pages: pages}
}

func (uc *GetProductPage) Execute(ctx context.Context, form domain.ProductPageForm) (domain.ProductPage, error) {
	query, err := domain.NewProductPageQuery(form)
	if err != nil {
		return domain.ProductPage{}, fmt.Errorf("read product page address: %w", err)
	}
	page, err := uc.pages.GetProductPage(ctx, query)
	if err != nil {
		return domain.ProductPage{}, fmt.Errorf("get product page %s/%s: %w", form.ConsignorSlug, form.SKU, err)
	}
	return page, nil
}
