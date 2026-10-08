package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type PartnershipEnder interface {
	LockForEnding(ctx context.Context, key domain.EndKey) (domain.EndState, error)
	SaveEnding(ctx context.Context, ending domain.Ending) error
	SetShownOnPublic(ctx context.Context, change domain.ListingChange) error
}

type OutstandingSettler interface {
	SettleOutstanding(ctx context.Context, partnershipID string, at time.Time) error
}

type Sealer interface {
	Seal(ctx context.Context, request domain.SealRequest) error
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

type EndDeps struct {
	Partnerships PartnershipEnder
	Reader       PartnershipReader
	Settler      OutstandingSettler
	Sealer       Sealer
	Announcer    Announcer
	Audit        AuditTrail
	Tx           Transactor
	Clock        Clock
}

func endKeyOf(key domain.Key, now time.Time) domain.EndKey {
	return domain.EndKey{Key: key, ConfirmCutoff: now.Add(-domain.PendingWindow)}
}
