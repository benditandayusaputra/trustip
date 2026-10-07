package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type CredentialStore interface {
	FindCredentials(ctx context.Context, email string) (domain.Credentials, bool, error)
	LockLoginAttempts(ctx context.Context, userID string) (domain.LoginAttempts, error)
	SaveLoginAttempts(ctx context.Context, userID string, attempts domain.LoginAttempts) error
	SavePassword(ctx context.Context, userID string, password domain.StoredPassword) error
}

type PasswordVerifier interface {
	Hash(password string) (domain.StoredPassword, error)
	Verify(stored domain.StoredPassword, password string) (domain.PasswordCheck, error)
	VerifyAbsent(password string) error
}

type SessionCreator interface {
	CreateSession(ctx context.Context, session domain.Session) error
}

type TokenIssuer interface {
	NewRefreshToken(familyID string) (domain.RefreshToken, error)
	SuccessorOf(previous string) domain.RefreshToken
	HashRefreshToken(plain string) []byte
	SignAccessToken(claims domain.AccessClaims) (string, error)
}

type SessionRotator interface {
	LockSession(ctx context.Context, tokenHash []byte) (domain.StoredSession, bool, error)
	MarkRotated(ctx context.Context, sessionID string, at time.Time) error
	CreateSession(ctx context.Context, session domain.Session) error
	RevokeFamily(ctx context.Context, familyID string, at time.Time) error
}

type UserSessionRevoker interface {
	RevokeUser(ctx context.Context, userID string, at time.Time) ([]string, error)
}

type AccessRevoker interface {
	Revoke(familyID string, until time.Time)
}

type Transactor interface {
	WithinTx(ctx context.Context, work func(context.Context) error) error
}
