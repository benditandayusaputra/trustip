package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Users) GetAccount(ctx context.Context, userID string) (domain.Account, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetAccount(readCtx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, apperror.NotFound()
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account %s: %w", userID, err)
	}
	return domain.Account{
		ID: row.ID, Email: row.Email, FullName: row.FullName, WhatsApp: row.Whatsapp, AvatarFileID: row.AvatarFileID,
	}, nil
}

func (r *Users) UpdateProfile(ctx context.Context, change domain.AccountChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := checkOwnAvatar(writeCtx, queries, change); err != nil {
		return err
	}
	updated, err := queries.UpdateAccountProfile(writeCtx, sqlc.UpdateAccountProfileParams{
		ID:           change.UserID,
		FullName:     change.Profile.FullName,
		Whatsapp:     change.Profile.WhatsApp.String(),
		AvatarFileID: change.Profile.AvatarFileID,
		UpdatedAt:    change.At,
	})
	if err != nil {
		return fmt.Errorf("update account %s: %w", change.UserID, err)
	}
	if updated == 0 {
		return apperror.NotFound()
	}
	return nil
}

func checkOwnAvatar(ctx context.Context, queries *sqlc.Queries, change domain.AccountChange) error {
	fileID := change.Profile.AvatarFileID
	if fileID == "" {
		return nil
	}
	if change.OrganizationID == "" {
		return domain.ErrUnknownAvatar
	}
	found, err := queries.CountOwnAvatarFiles(ctx, sqlc.CountOwnAvatarFilesParams{
		ID: fileID, OrganizationID: change.OrganizationID, CreatedBy: change.UserID,
	})
	if err != nil {
		return fmt.Errorf("check avatar %s: %w", fileID, err)
	}
	if found == 0 {
		return domain.ErrUnknownAvatar
	}
	return nil
}

func (r *Users) LockPasswordOwner(ctx context.Context, userID string) (domain.PasswordOwner, error) {
	row, err := sqlc.New(r.pool.Querier(ctx)).LockPasswordOwner(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PasswordOwner{}, apperror.NotFound()
	}
	if err != nil {
		return domain.PasswordOwner{}, fmt.Errorf("lock password of %s: %w", userID, err)
	}
	return domain.PasswordOwner{
		FullName: row.FullName,
		Email:    row.Email,
		Password: domain.StoredPassword{Hash: row.PasswordHash, PepperVersion: row.PepperVersion},
	}, nil
}
