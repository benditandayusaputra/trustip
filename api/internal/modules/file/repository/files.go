package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Files struct {
	pool *database.Pool
}

func NewFiles(pool *database.Pool) *Files {
	return &Files{pool: pool}
}

func (r *Files) CreateFile(ctx context.Context, file domain.File) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).CreateFile(writeCtx, sqlc.CreateFileParams{
		ID:             file.ID,
		OrganizationID: file.OrganizationID,
		Purpose:        string(file.Purpose),
		Visibility:     string(file.Visibility),
		ContentType:    file.ContentType,
		SizeBytes:      int64(file.SizeBytes),
		StorageKey:     file.StorageKey,
		Sha256:         file.SHA256,
		Width:          int64(file.Width),
		Height:         int64(file.Height),
		CreatedBy:      file.CreatedBy,
		CreatedAt:      file.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("insert file %s: %w", file.ID, err)
	}
	return nil
}

func (r *Files) GetFile(ctx context.Context, organizationID, fileID string) (domain.File, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetOrganizationFile(readCtx, sqlc.GetOrganizationFileParams{
		ID: fileID, OrganizationID: organizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.File{}, apperror.NotFound()
	}
	if err != nil {
		return domain.File{}, fmt.Errorf("get file %s: %w", fileID, err)
	}
	return domain.File{
		ID: row.ID, OrganizationID: row.OrganizationID, Purpose: domain.Purpose(row.Purpose),
		Visibility: domain.Visibility(row.Visibility), ContentType: row.ContentType, SizeBytes: int(row.SizeBytes),
		StorageKey: row.StorageKey, SHA256: row.Sha256, Width: int(row.Width), Height: int(row.Height),
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt,
	}, nil
}
