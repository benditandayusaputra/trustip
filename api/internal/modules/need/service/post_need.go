package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type PostNeed struct {
	deps Deps
}

func NewPostNeed(deps Deps) *PostNeed {
	return &PostNeed{deps: deps}
}

func (uc *PostNeed) Execute(ctx context.Context, signedIn actor.Actor, form domain.Form) (domain.Need, error) {
	if err := policy.Require(signedIn, policy.ManageNeeds); err != nil {
		return domain.Need{}, fmt.Errorf("check need poster: %w", err)
	}
	draft, err := domain.NewDraft(form)
	if err != nil {
		return domain.Need{}, fmt.Errorf("validate need: %w", err)
	}
	needID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.Need{}, fmt.Errorf("generate need id: %w", err)
	}
	posting := draft.Post(domain.Author{
		NeedID: needID, StoreID: signedIn.Membership.OrganizationID, UserID: signedIn.UserID,
	}, uc.deps.Clock.Now())
	var need domain.Need
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := explainPostFailure(uc.deps.Writer.CreateNeed(ctx, posting)); err != nil {
			return err
		}
		need, err = finishNeedEvent(ctx, uc.deps, needEvent{
			signedIn: signedIn, needID: needID, action: audit.NeedPosted, at: posting.At,
		})
		return err
	})
	if err != nil {
		return domain.Need{}, fmt.Errorf("post need: %w", err)
	}
	return need, nil
}

func explainPostFailure(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrUnknownCategory):
		return apperror.Validation(apperror.InvalidField("category_id", apperror.FieldUnknownChoice))
	case errors.Is(err, domain.ErrOpenLimit):
		return apperror.New(apperror.CodeNeedLimitReached)
	}
	return fmt.Errorf("create need: %w", err)
}
