package service

import (
	"context"

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
	NewRefreshToken() (domain.RefreshToken, error)
	SignAccessToken(claims domain.AccessClaims) (string, error)
}

type Transactor interface {
	WithinTx(ctx context.Context, work func(context.Context) error) error
}
