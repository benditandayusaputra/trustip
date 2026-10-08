package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type StoreDirectory interface {
	GetViewer(ctx context.Context, key domain.ViewerKey) (domain.Viewer, error)
	CountStoresWithin(ctx context.Context, query domain.StoreQuery) (map[int]int, error)
	ListStores(ctx context.Context, query domain.StoreQuery) ([]domain.StoreCard, error)
}

type Clock interface {
	Now() time.Time
}
