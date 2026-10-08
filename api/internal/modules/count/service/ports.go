package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type CountWriter interface {
	LockSheet(ctx context.Context, key domain.SheetKey) (domain.Sheet, error)
	NextNumber(ctx context.Context, key domain.SequenceKey) (int, error)
	CreateCount(ctx context.Context, count domain.Count) error
}

type CountReader interface {
	GetSheet(ctx context.Context, key domain.SheetKey) (domain.Sheet, error)
	FindByKey(ctx context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error)
	GetCount(ctx context.Context, key domain.CountKey) (domain.View, error)
}

type CountStore interface {
	CountWriter
	CountReader
}

type Sealer interface {
	Seal(ctx context.Context, request domain.SealRequest) error
}

type SealChecker interface {
	Check(ctx context.Context, check domain.SealCheck) (domain.SealVerdict, error)
}

type Announcer interface {
	Announce(ctx context.Context, notice domain.Notice) error
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
	Counts    CountStore
	Sealer    Sealer
	Seals     SealChecker
	Announcer Announcer
	Audit     AuditTrail
	Tx        Transactor
	Clock     Clock
	IDs       IDGenerator
}
