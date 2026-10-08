package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ViewerReader interface {
	GetViewer(ctx context.Context, key domain.ViewerKey) (domain.Viewer, error)
}

type StoreDirectory interface {
	CountStoresWithin(ctx context.Context, query domain.StoreQuery) (map[int]int, error)
	ListStores(ctx context.Context, query domain.StoreQuery) ([]domain.StoreCard, error)
}

type ConsignorDirectory interface {
	CountConsignorsWithin(ctx context.Context, query domain.ConsignorQuery) (map[int]int, error)
	CountProductsWithin(ctx context.Context, query domain.ConsignorQuery) (map[int]int, error)
	ListConsignors(ctx context.Context, query domain.ConsignorQuery) ([]domain.ConsignorCard, error)
	ListProducts(ctx context.Context, query domain.ConsignorQuery) ([]domain.ProductCard, error)
}

type NeedBoard interface {
	CountNeedsWithin(ctx context.Context, query domain.NeedQuery) (map[int]int, error)
	ListNeeds(ctx context.Context, query domain.NeedQuery) ([]domain.NeedCard, error)
}

type Clock interface {
	Now() time.Time
}
