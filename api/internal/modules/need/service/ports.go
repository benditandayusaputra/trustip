package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type NeedWriter interface {
	CreateNeed(ctx context.Context, posting domain.Posting) error
	LockNeed(ctx context.Context, key domain.NeedKey) (domain.Need, error)
	SaveChange(ctx context.Context, change domain.Change) error
	ExpireOpen(ctx context.Context, now time.Time) (int, error)
}

type NeedReader interface {
	GetNeed(ctx context.Context, key domain.NeedKey) (domain.Need, error)
	ListNeeds(ctx context.Context, query domain.ListQuery) ([]domain.Need, error)
	ListResponses(ctx context.Context, key domain.NeedKey) ([]domain.Response, error)
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

type Deps struct {
	Writer NeedWriter
	Reader NeedReader
	Audit  AuditTrail
	Tx     Transactor
	Clock  Clock
	IDs    IDGenerator
}
