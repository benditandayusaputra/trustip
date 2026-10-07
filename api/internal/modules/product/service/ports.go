package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
)

type ProductWriter interface {
	NextSKUNumber(ctx context.Context, series domain.SKUSeries) (int, error)
	SaveProduct(ctx context.Context, change domain.Change) error
	SetArchived(ctx context.Context, change domain.StatusChange) error
}

type ProductReader interface {
	GetProduct(ctx context.Context, key domain.ProductKey) (domain.Product, error)
	ListProducts(ctx context.Context, query domain.PageQuery) ([]domain.Product, error)
}

type ProductStore interface {
	ProductWriter
	ProductReader
}

type AuditTrail interface {
	Record(ctx context.Context, entry audit.Entry) error
}

type Transactor interface {
	WithinTx(ctx context.Context, work func(context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}
