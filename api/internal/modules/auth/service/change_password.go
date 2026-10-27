package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type PasswordOwnerStore interface {
	LockPasswordOwner(ctx context.Context, userID string) (domain.PasswordOwner, error)
	SavePassword(ctx context.Context, userID string, password domain.StoredPassword) error
}

type OtherSessionRevoker interface {
	RevokeOtherSessions(ctx context.Context, userID, keepFamily string, at time.Time) ([]string, error)
}

type ChangePasswordDeps struct {
	Owners    PasswordOwnerStore
	Sessions  OtherSessionRevoker
	Passwords PasswordVerifier
	Policy    domain.PasswordPolicy
	Access    AccessRevoker
	Audit     AuditTrail
	Tx        Transactor
	Clock     Clock
}

type ChangePassword struct {
	deps ChangePasswordDeps
}

func NewChangePassword(deps ChangePasswordDeps) *ChangePassword {
	return &ChangePassword{deps: deps}
}

func (uc *ChangePassword) Execute(ctx context.Context, signedIn actor.Actor, form domain.PasswordChangeForm) error {
	change := domain.AccountChange{
		UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID, At: uc.deps.Clock.Now(),
	}
	var families []string
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.replace(ctx, change.UserID, form); err != nil {
			return err
		}
		revoked, err := uc.deps.Sessions.RevokeOtherSessions(ctx, change.UserID, signedIn.SessionFamily, change.At)
		if err != nil {
			return fmt.Errorf("revoke other sessions of %s: %w", change.UserID, err)
		}
		families = revoked
		return uc.deps.Audit.Record(ctx, accountEntry(change, audit.PasswordChanged))
	})
	if err != nil {
		return fmt.Errorf("change password of %s: %w", signedIn.UserID, err)
	}
	revokeAccess(uc.deps.Access, families, change.At)
	return nil
}

func (uc *ChangePassword) replace(ctx context.Context, userID string, form domain.PasswordChangeForm) error {
	owner, err := uc.deps.Owners.LockPasswordOwner(ctx, userID)
	if err != nil {
		return fmt.Errorf("lock password of %s: %w", userID, err)
	}
	if err := domain.CheckNewPassword(form, owner, uc.deps.Policy); err != nil {
		return fmt.Errorf("check new password: %w", err)
	}
	check, err := uc.deps.Passwords.Verify(owner.Password, form.Current)
	if err != nil {
		return fmt.Errorf("verify current password: %w", err)
	}
	if !check.Matches {
		return fmt.Errorf("compare current password: %w", domain.WrongCurrentPassword())
	}
	stored, err := uc.deps.Passwords.Hash(form.New)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	if err := uc.deps.Owners.SavePassword(ctx, userID, stored); err != nil {
		return fmt.Errorf("save new password of %s: %w", userID, err)
	}
	return nil
}
