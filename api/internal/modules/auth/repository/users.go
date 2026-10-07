package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const emailUniqueConstraint = "users_email_key"

type Users struct {
	pool *database.Pool
}

func NewUsers(pool *database.Pool) *Users {
	return &Users{pool: pool}
}

func (r *Users) CreateUser(ctx context.Context, user domain.User) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).CreateUser(writeCtx, sqlc.CreateUserParams{
		ID:                user.ID,
		Email:             user.Email,
		PasswordHash:      user.Password.Hash,
		PepperVersion:     user.Password.PepperVersion,
		FullName:          user.FullName,
		Whatsapp:          user.WhatsApp.String(),
		PrivacyAcceptedAt: user.PrivacyAcceptedAt,
	})
	if database.IsUniqueViolation(err, emailUniqueConstraint) {
		return domain.ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("insert user %s: %w", user.ID, err)
	}
	return nil
}
