package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type DropWriter interface {
	NextNumber(ctx context.Context, key domain.SequenceKey) (int, error)
	CreateDrop(ctx context.Context, drop domain.Drop) error
}

type DropReader interface {
	GetSetting(ctx context.Context, key domain.SettingKey) (domain.Setting, error)
	FindByKey(ctx context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error)
	GetDrop(ctx context.Context, key domain.DropKey) (domain.View, error)
}

type DropStore interface {
	DropWriter
	DropReader
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
	Drops     DropStore
	Sealer    Sealer
	Seals     SealChecker
	Announcer Announcer
	Audit     AuditTrail
	Tx        Transactor
	Clock     Clock
	IDs       IDGenerator
}
