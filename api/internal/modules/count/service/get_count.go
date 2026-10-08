package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type GetCount struct {
	counts CountReader
	seals  SealChecker
	clock  Clock
}

func NewGetCount(deps Deps) *GetCount {
	return &GetCount{counts: deps.Counts, seals: deps.Seals, clock: deps.Clock}
}

func (uc *GetCount) Execute(ctx context.Context, signedIn actor.Actor, countID string) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return domain.SealedView{}, fmt.Errorf("check count reader: %w", err)
	}
	return uc.read(ctx, signedIn, countID)
}

func (uc *GetCount) read(ctx context.Context, signedIn actor.Actor, countID string) (domain.SealedView, error) {
	view, err := uc.counts.GetCount(ctx, domain.CountKey{
		OrganizationID: signedIn.Membership.OrganizationID, CountID: countID,
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("get count %s: %w", countID, err)
	}
	verdict, err := uc.seals.Check(ctx, view.SealCheck())
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("check seal of count %s: %w", countID, err)
	}
	return domain.SealedView{View: view.Effective(uc.clock.Now()), Seal: verdict}, nil
}

type GetSheet struct {
	counts CountReader
}

func NewGetSheet(deps Deps) *GetSheet {
	return &GetSheet{counts: deps.Counts}
}

func (uc *GetSheet) Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.Sheet, error) {
	if err := policy.Require(signedIn, policy.RecordCounts); err != nil {
		return domain.Sheet{}, fmt.Errorf("check count recorder: %w", err)
	}
	sheet, err := uc.counts.GetSheet(ctx, domain.SheetKey{
		StoreID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return domain.Sheet{}, fmt.Errorf("get count sheet of %s: %w", partnershipID, err)
	}
	return sheet, nil
}
