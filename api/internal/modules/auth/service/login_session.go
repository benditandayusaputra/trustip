package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type successfulLogin struct {
	credentials domain.Credentials
	check       domain.PasswordCheck
	password    string
	now         time.Time
}

func (uc *LoginUser) startSession(ctx context.Context, login successfulLogin) (LoginResult, error) {
	userID := login.credentials.UserID
	identity, err := uc.newSessionIdentity(userID)
	if err != nil {
		return LoginResult{}, err
	}
	refresh, err := uc.deps.Tokens.NewRefreshToken(identity.FamilyID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("create refresh token: %w", err)
	}
	session := domain.NewSession(identity, refresh.Hash, login.now)
	if err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		return uc.saveSuccessfulLogin(ctx, login, session)
	}); err != nil {
		return LoginResult{}, fmt.Errorf("start session for %s: %w", userID, err)
	}
	claims, err := uc.claimsFor(ctx, identity, login.now)
	if err != nil {
		return LoginResult{}, err
	}
	access, err := uc.deps.Tokens.SignAccessToken(claims)
	if err != nil {
		return LoginResult{}, fmt.Errorf("sign access token: %w", err)
	}
	return LoginResult{
		Account: domain.Account{ID: userID, Email: login.credentials.Email, FullName: login.credentials.FullName},
		Tokens: domain.SessionTokens{
			Access: access, AccessExpires: claims.ExpiresAt,
			Refresh: refresh.Plain, RefreshExpires: session.ExpiresAt,
		},
	}, nil
}

func (uc *LoginUser) claimsFor(
	ctx context.Context, identity domain.SessionIdentity, now time.Time,
) (domain.AccessClaims, error) {
	claims := domain.NewAccessClaims(identity.UserID, identity.FamilyID, now)
	membership, _, err := uc.deps.Memberships.FindMembership(ctx, identity.UserID)
	if err != nil {
		return domain.AccessClaims{}, fmt.Errorf("find membership for access token: %w", err)
	}
	claims.Membership = membership
	return claims, nil
}

func (uc *LoginUser) newSessionIdentity(userID string) (domain.SessionIdentity, error) {
	sessionID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.SessionIdentity{}, fmt.Errorf("generate session id: %w", err)
	}
	familyID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.SessionIdentity{}, fmt.Errorf("generate session family id: %w", err)
	}
	return domain.SessionIdentity{ID: sessionID, UserID: userID, FamilyID: familyID}, nil
}

func (uc *LoginUser) saveSuccessfulLogin(ctx context.Context, login successfulLogin, session domain.Session) error {
	userID := login.credentials.UserID
	cleared := login.credentials.Attempts.AfterSuccess()
	if err := uc.deps.Credentials.SaveLoginAttempts(ctx, userID, cleared); err != nil {
		return fmt.Errorf("clear login attempts: %w", err)
	}
	if login.check.NeedsRehash {
		if err := uc.rehash(ctx, userID, login.password); err != nil {
			return err
		}
	}
	if err := uc.deps.Sessions.CreateSession(ctx, session); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (uc *LoginUser) rehash(ctx context.Context, userID, password string) error {
	stored, err := uc.deps.Passwords.Hash(password)
	if err != nil {
		return fmt.Errorf("rehash password: %w", err)
	}
	if err := uc.deps.Credentials.SavePassword(ctx, userID, stored); err != nil {
		return fmt.Errorf("save rehashed password: %w", err)
	}
	return nil
}
