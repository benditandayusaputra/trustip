package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type LedgerStore interface {
	GetLedger(ctx context.Context, partnershipID string) (domain.Ledger, error)
	NextNumber(ctx context.Context, key domain.SequenceKey) (int, error)
	CreateSettlement(ctx context.Context, settlement domain.Settlement) error
}

type SettlementReader interface {
	GetSettlement(ctx context.Context, key domain.SettlementKey) (domain.View, error)
	ListSettlements(ctx context.Context, query domain.ListQuery) ([]domain.ListItem, error)
	GetRunning(ctx context.Context, key domain.RunningKey) (domain.Running, error)
}

type DueLister interface {
	ListDue(ctx context.Context, key domain.DueKey) ([]string, error)
	LockLedger(ctx context.Context, partnershipID string) (domain.Ledger, error)
}

type PaymentWriter interface {
	FindPaymentByKey(ctx context.Context, key domain.PaymentKeyRef) (domain.StoredPayment, bool, error)
	LockPayable(ctx context.Context, key domain.PayableKey) (domain.Payable, error)
	CreatePayment(ctx context.Context, payment domain.Payment) error
}

type PaymentReviewer interface {
	LockPayment(ctx context.Context, key domain.PaymentKey) (domain.Payment, error)
	SaveReview(ctx context.Context, review domain.Review) error
	MarkPaid(ctx context.Context, settlementID string, paidAt time.Time) error
}

type PaymentStore interface {
	PaymentWriter
	PaymentReviewer
}

type SettlementStore interface {
	LedgerStore
	SettlementReader
	DueLister
	PaymentStore
}

type CountApprover interface {
	ApproveSources(ctx context.Context, approval domain.Approval) error
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
	Settlements SettlementStore
	Sealer      Sealer
	Seals       SealChecker
	Announcer   Announcer
	Audit       AuditTrail
	Approver    CountApprover
	Tx          Transactor
	Clock       Clock
	IDs         IDGenerator
}
