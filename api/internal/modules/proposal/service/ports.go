package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

type ProposalWriter interface {
	NextNumber(ctx context.Context, key domain.SequenceKey) (int, error)
	CreateInvitation(ctx context.Context, invitation domain.Invitation) error
	ClaimInvitation(ctx context.Context, claim domain.Claim) error
}

type ProposalReader interface {
	GetProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error)
	GetInvitation(ctx context.Context, code string) (domain.InvitationView, error)
	HasActivePartnership(ctx context.Context, pair domain.Pair) (bool, error)
}

type ProposalStore interface {
	ProposalWriter
	ProposalReader
}

type CodeGenerator interface {
	NewCode() (string, error)
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
