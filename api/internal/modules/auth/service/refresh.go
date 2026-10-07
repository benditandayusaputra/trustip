package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type RefreshDeps struct {
	Sessions SessionRotator
	Tokens   TokenIssuer
	Access   AccessRevoker
	Tx       Transactor
	Clock    Clock
	IDs      IDGenerator
}

type RefreshSession struct {
	deps RefreshDeps
}

func NewRefreshSession(deps RefreshDeps) *RefreshSession {
	return &RefreshSession{deps: deps}
}

type refreshOutcome struct {
	tokens        domain.SessionTokens
	refusal       error
	revokedFamily string
}

func (uc *RefreshSession) Execute(ctx context.Context, presented string) (domain.SessionTokens, error) {
	if presented == "" {
		return domain.SessionTokens{}, apperror.New(apperror.CodeSessionExpired)
	}
	now := uc.deps.Clock.Now()
	var outcome refreshOutcome
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		decided, err := uc.decide(ctx, presented, now)
		outcome = decided
		return err
	})
	if err != nil {
		return domain.SessionTokens{}, fmt.Errorf("refresh session: %w", err)
	}
	if outcome.revokedFamily != "" {
		uc.deps.Access.Revoke(outcome.revokedFamily, now.Add(domain.AccessTokenLifetime))
	}
	if outcome.refusal != nil {
		return domain.SessionTokens{}, outcome.refusal
	}
	return outcome.tokens, nil
}

func (uc *RefreshSession) decide(ctx context.Context, presented string, now time.Time) (refreshOutcome, error) {
	session, found, err := uc.deps.Sessions.LockSession(ctx, uc.deps.Tokens.HashRefreshToken(presented))
	if err != nil {
		return refreshOutcome{}, fmt.Errorf("lock session: %w", err)
	}
	if !found {
		return refreshOutcome{refusal: apperror.New(apperror.CodeSessionExpired)}, nil
	}
	switch domain.DecideRefresh(session, now) {
	case domain.RefreshRotate:
		return uc.rotate(ctx, session, presented, now)
	case domain.RefreshReplay:
		return uc.issue(session, uc.deps.Tokens.SuccessorOf(presented), session.RotatedAt)
	case domain.RefreshReuseDetected:
		return uc.revoke(ctx, session, now)
	case domain.RefreshRejected:
	}
	return refreshOutcome{refusal: apperror.New(apperror.CodeSessionExpired)}, nil
}

func (uc *RefreshSession) rotate(
	ctx context.Context, session domain.StoredSession, presented string, now time.Time,
) (refreshOutcome, error) {
	successorID, err := uc.deps.IDs.NewID()
	if err != nil {
		return refreshOutcome{}, fmt.Errorf("generate session id: %w", err)
	}
	successor := uc.deps.Tokens.SuccessorOf(presented)
	if err := uc.deps.Sessions.MarkRotated(ctx, session.ID, now); err != nil {
		return refreshOutcome{}, fmt.Errorf("mark rotated: %w", err)
	}
	if err := uc.deps.Sessions.CreateSession(ctx, session.Rotate(successorID, successor.Hash, now)); err != nil {
		return refreshOutcome{}, fmt.Errorf("create successor session: %w", err)
	}
	return uc.issue(session, successor, now)
}

func (uc *RefreshSession) issue(
	session domain.StoredSession, refresh domain.RefreshToken, issuedAt time.Time,
) (refreshOutcome, error) {
	claims := domain.NewAccessClaims(session.UserID, session.FamilyID, issuedAt)
	access, err := uc.deps.Tokens.SignAccessToken(claims)
	if err != nil {
		return refreshOutcome{}, fmt.Errorf("sign access token: %w", err)
	}
	return refreshOutcome{tokens: domain.SessionTokens{
		Access: access, AccessExpires: claims.ExpiresAt,
		Refresh: refresh.Plain, RefreshExpires: issuedAt.Add(domain.RefreshTokenLifetime),
	}}, nil
}

func (uc *RefreshSession) revoke(
	ctx context.Context, session domain.StoredSession, now time.Time,
) (refreshOutcome, error) {
	if err := uc.deps.Sessions.RevokeFamily(ctx, session.FamilyID, now); err != nil {
		return refreshOutcome{}, fmt.Errorf("revoke reused family: %w", err)
	}
	refusal := apperror.New(apperror.CodeSessionRevoked).WithCause(domain.ErrRefreshReused)
	return refreshOutcome{refusal: refusal, revokedFamily: session.FamilyID}, nil
}
