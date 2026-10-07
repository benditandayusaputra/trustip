package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
)

type AgreementReader interface {
	GetAgreement(ctx context.Context, key domain.VersionKey) (domain.View, error)
	ListVersions(ctx context.Context, key domain.PartnershipKey) ([]domain.Agreement, error)
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

type VersionRequest struct {
	PartnershipID string
	Version       int
}

func (uc *GetAgreement) Execute(
	ctx context.Context, signedIn actor.Actor, request VersionRequest,
) (domain.SealedView, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.SealedView{}, fmt.Errorf("check agreement reader: %w", err)
	}
	view, err := uc.agreements.GetAgreement(ctx, domain.VersionKey{
		OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: request.PartnershipID,
		Version: request.Version,
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("get agreement of %s: %w", request.PartnershipID, err)
	}
	verdict, err := uc.seals.Check(ctx, view.SealCheck())
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("check agreement seal: %w", err)
	}
	return domain.SealedView{View: view, Seal: verdict}, nil
}

type ListVersions struct {
	agreements AgreementReader
}

func NewListVersions(agreements AgreementReader) *ListVersions {
	return &ListVersions{agreements: agreements}
}

func (uc *ListVersions) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) ([]domain.Version, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check agreement reader: %w", err)
	}
	agreements, err := uc.agreements.ListVersions(ctx, domain.PartnershipKey{
		OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return nil, fmt.Errorf("list versions of %s: %w", partnershipID, err)
	}
	return domain.History(agreements), nil
}
