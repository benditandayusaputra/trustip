package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
)

type AgreementWriter interface {
	CreatePartnership(ctx context.Context, partnership domain.Partnership) error
	CreateAgreement(ctx context.Context, agreement domain.Agreement) error
}

type Sealer interface {
	Seal(ctx context.Context, request domain.SealRequest) error
}

type IDGenerator interface {
	NewID() (string, error)
}

type FounderDeps struct {
	Agreements AgreementWriter
	Sealer     Sealer
	IDs        IDGenerator
}

type Founder struct {
	deps FounderDeps
}

func NewFounder(deps FounderDeps) *Founder {
	return &Founder{deps: deps}
}

func (f *Founder) Found(ctx context.Context, input domain.FoundingInput) (domain.Agreement, error) {
	partnershipID, partnershipErr := f.deps.IDs.NewID()
	agreementID, agreementErr := f.deps.IDs.NewID()
	if err := errors.Join(partnershipErr, agreementErr); err != nil {
		return domain.Agreement{}, fmt.Errorf("generate founding ids: %w", err)
	}
	founding := domain.Found(input, domain.FoundingIDs{PartnershipID: partnershipID, AgreementID: agreementID})
	if err := f.deps.Agreements.CreatePartnership(ctx, founding.Partnership); err != nil {
		return domain.Agreement{}, fmt.Errorf("create partnership: %w", err)
	}
	if err := f.deps.Agreements.CreateAgreement(ctx, founding.Agreement); err != nil {
		return domain.Agreement{}, fmt.Errorf("create agreement: %w", err)
	}
	if err := f.deps.Sealer.Seal(ctx, founding.SealRequest(input)); err != nil {
		return domain.Agreement{}, fmt.Errorf("seal agreement %s: %w", agreementID, err)
	}
	return founding.Agreement, nil
}
