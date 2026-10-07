package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type LogoutDeps struct {
	Sessions SessionRotator
	Users    UserSessionRevoker
	Tokens   TokenIssuer
	Access   AccessRevoker
	Tx       Transactor
	Clock    Clock
}

type LogOut struct {
	deps LogoutDeps
}

func NewLogOut(deps LogoutDeps) *LogOut {
	return &LogOut{deps: deps}
}

func (uc *LogOut) Execute(ctx context.Context, presented string) error {
	if presented == "" {
		return nil
	}
	now := uc.deps.Clock.Now()
	var families []string
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		session, found, err := uc.deps.Sessions.LockSession(ctx, uc.deps.Tokens.HashRefreshToken(presented))
		if err != nil || !found {
			return wrapLockFailure(err)
		}
		families = []string{session.FamilyID}
		if err := uc.deps.Sessions.RevokeFamily(ctx, session.FamilyID, now); err != nil {
			return fmt.Errorf("revoke family %s: %w", session.FamilyID, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("log out: %w", err)
	}
	revokeAccess(uc.deps.Access, families, now)
	return nil
}

type LogOutEverywhere struct {
	deps LogoutDeps
}

func NewLogOutEverywhere(deps LogoutDeps) *LogOutEverywhere {
	return &LogOutEverywhere{deps: deps}
}

func (uc *LogOutEverywhere) Execute(ctx context.Context, presented string) error {
	now := uc.deps.Clock.Now()
	var families []string
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		session, found, err := uc.deps.Sessions.LockSession(ctx, uc.deps.Tokens.HashRefreshToken(presented))
		if err != nil {
			return wrapLockFailure(err)
		}
		if !found || !isUsable(session, now) {
			return apperror.New(apperror.CodeSessionExpired)
		}
		families, err = uc.deps.Users.RevokeUser(ctx, session.UserID, now)
		if err != nil {
			return fmt.Errorf("revoke sessions of %s: %w", session.UserID, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("log out everywhere: %w", err)
	}
	revokeAccess(uc.deps.Access, families, now)
	return nil
}

func isUsable(session domain.StoredSession, now time.Time) bool {
	decision := domain.DecideRefresh(session, now)
	return decision == domain.RefreshRotate || decision == domain.RefreshReplay
}

func revokeAccess(access AccessRevoker, families []string, now time.Time) {
	for _, familyID := range families {
		access.Revoke(familyID, now.Add(domain.AccessTokenLifetime))
	}
}

func wrapLockFailure(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("lock session: %w", err)
}
