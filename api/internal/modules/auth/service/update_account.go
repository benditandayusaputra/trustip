package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type AccountProfileStore interface {
	GetAccount(ctx context.Context, userID string) (domain.Account, error)
	UpdateProfile(ctx context.Context, change domain.AccountChange) error
}

type UpdateAccountDeps struct {
	Accounts AccountProfileStore
	Audit    AuditTrail
	Tx       Transactor
	Clock    Clock
}

type UpdateAccount struct {
	deps UpdateAccountDeps
}

func NewUpdateAccount(deps UpdateAccountDeps) *UpdateAccount {
	return &UpdateAccount{deps: deps}
}

func (uc *UpdateAccount) Execute(
	ctx context.Context, signedIn actor.Actor, form domain.ProfileForm,
) (domain.Account, error) {
	profile, err := domain.NewProfileChange(form)
	if err != nil {
		return domain.Account{}, fmt.Errorf("check account profile: %w", err)
	}
	change := domain.AccountChange{
		UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID,
		Profile: profile, At: uc.deps.Clock.Now(),
	}
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.deps.Accounts.UpdateProfile(ctx, change); err != nil {
			return explainProfileFailure(err)
		}
		return uc.deps.Audit.Record(ctx, accountEntry(change, audit.AccountUpdated))
	})
	if err != nil {
		return domain.Account{}, fmt.Errorf("update account %s: %w", signedIn.UserID, err)
	}
	account, err := uc.deps.Accounts.GetAccount(ctx, signedIn.UserID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("read updated account %s: %w", signedIn.UserID, err)
	}
	account.Membership = signedIn.Membership
	return account, nil
}

func explainProfileFailure(err error) error {
	if errors.Is(err, domain.ErrUnknownAvatar) {
		return apperror.Validation(apperror.InvalidField("avatar_file_id", apperror.FieldUnknownChoice))
	}
	return fmt.Errorf("save account profile: %w", err)
}

func accountEntry(change domain.AccountChange, action audit.Action) audit.Entry {
	return audit.Entry{
		OrganizationID: change.OrganizationID, ActorID: change.UserID, Action: action,
		EntityType: audit.UserEntity, EntityID: change.UserID, At: change.At,
	}
}
