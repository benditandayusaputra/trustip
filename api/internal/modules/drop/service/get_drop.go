package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type GetDrop struct {
	drops DropReader
	seals SealChecker
}

func NewGetDrop(deps Deps) *GetDrop {
	return &GetDrop{drops: deps.Drops, seals: deps.Seals}
}

func (uc *GetDrop) Execute(ctx context.Context, signedIn actor.Actor, dropID string) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return domain.SealedView{}, fmt.Errorf("check drop reader: %w", err)
	}
	return uc.read(ctx, signedIn, dropID)
}

func (uc *GetDrop) read(ctx context.Context, signedIn actor.Actor, dropID string) (domain.SealedView, error) {
	view, err := uc.drops.GetDrop(ctx, domain.DropKey{
		OrganizationID: signedIn.Membership.OrganizationID, DropID: dropID,
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("get drop %s: %w", dropID, err)
	}
	verdict, err := uc.seals.Check(ctx, view.SealCheck())
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("check seal of drop %s: %w", dropID, err)
	}
	return domain.SealedView{View: view, Seal: verdict}, nil
}
