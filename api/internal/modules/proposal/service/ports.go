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
	SaveDecision(ctx context.Context, decision domain.Decision) error
}

type ProposalReader interface {
	GetProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error)
	GetInvitation(ctx context.Context, code string) (domain.InvitationView, error)
	HasActivePartnership(ctx context.Context, pair domain.Pair) (bool, error)
	LockProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error)
}

type ProposalLister interface {
	ListProposals(ctx context.Context, query domain.ListQuery) ([]domain.Summary, error)
}

type ProposalStore interface {
	ProposalWriter
	ProposalReader
	ProposalLister
}

type AgreementFounder interface {
	Found(ctx context.Context, founding domain.Founding) (string, error)
}

type Announcer interface {
	Announce(ctx context.Context, notice domain.Notice) error
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
