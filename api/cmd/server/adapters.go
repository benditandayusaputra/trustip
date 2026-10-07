package main

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/captcha"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type captchaVerifier struct {
	turnstile *captcha.Turnstile
}

func (v captchaVerifier) Verify(ctx context.Context, challenge service.CaptchaChallenge) error {
	if err := v.turnstile.Verify(ctx, captcha.Challenge(challenge)); err != nil {
		return fmt.Errorf("verify captcha: %w", err)
	}
	return nil
}

type passwordHasher struct {
	hasher *security.PasswordHasher
}

func (h passwordHasher) Hash(password string) (domain.StoredPassword, error) {
	stored, err := h.hasher.Hash(password)
	if err != nil {
		return domain.StoredPassword{}, fmt.Errorf("hash password: %w", err)
	}
	return domain.StoredPassword(stored), nil
}

func (h passwordHasher) Verify(stored domain.StoredPassword, password string) (domain.PasswordCheck, error) {
	check, err := h.hasher.Verify(security.StoredPassword(stored), password)
	if err != nil {
		return domain.PasswordCheck{}, fmt.Errorf("verify password: %w", err)
	}
	return domain.PasswordCheck(check), nil
}

func (h passwordHasher) VerifyAbsent(password string) error {
	if err := h.hasher.VerifyAbsent(password); err != nil {
		return fmt.Errorf("spend comparison hash: %w", err)
	}
	return nil
}

type tokenIssuer struct {
	signer  *security.TokenSigner
	refresh *security.RefreshTokens
}

func (i tokenIssuer) NewRefreshToken(familyID string) (domain.RefreshToken, error) {
	token, err := i.refresh.New(familyID)
	if err != nil {
		return domain.RefreshToken{}, fmt.Errorf("new refresh token: %w", err)
	}
	return domain.RefreshToken(token), nil
}

func (i tokenIssuer) SuccessorOf(previous string) domain.RefreshToken {
	return domain.RefreshToken(i.refresh.Successor(previous))
}

func (i tokenIssuer) HashRefreshToken(plain string) []byte {
	return security.HashRefreshToken(plain)
}

func (i tokenIssuer) SignAccessToken(claims domain.AccessClaims) (string, error) {
	token, err := i.signer.Sign(security.AccessClaims{
		Subject:          claims.Subject,
		SessionID:        claims.SessionFamily,
		Organization:     claims.Membership.OrganizationID,
		OrganizationType: string(claims.Membership.OrganizationType),
		Role:             string(claims.Membership.Role),
		IssuedAt:         claims.IssuedAt.Unix(),
		ExpiresAt:        claims.ExpiresAt.Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return token, nil
}
