package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
)

type AgreementReader interface {
	GetActiveAgreement(ctx context.Context, key domain.PartnershipKey) (domain.View, error)
}

type SealChecker interface {
	Check(ctx context.Context, check domain.SealCheck) (domain.SealVerdict, error)
}

type GetAgreement struct {
	agreements AgreementReader
	seals      SealChecker
}

func NewGetAgreement(agreements AgreementReader, seals SealChecker) *GetAgreement {
	return &GetAgreement{agreements: agreements, seals: seals}
}

func (uc *GetAgreement) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) (domain.SealedView, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.SealedView{}, fmt.Errorf("check agreement reader: %w", err)
	}
	view, err := uc.agreements.GetActiveAgreement(ctx, domain.PartnershipKey{
		OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("get agreement of %s: %w", partnershipID, err)
	}
	verdict, err := uc.seals.Check(ctx, view.SealCheck())
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("check agreement seal: %w", err)
	}
	return domain.SealedView{View: view, Seal: verdict}, nil
}
