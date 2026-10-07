package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type CaptchaChallenge struct {
	Token    string
	Action   string
	RemoteIP string
}

type CaptchaVerifier interface {
	Verify(ctx context.Context, challenge CaptchaChallenge) error
}

type PasswordHasher interface {
	Hash(password string) (domain.StoredPassword, error)
}

type UserCreator interface {
	CreateUser(ctx context.Context, user domain.User) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}
