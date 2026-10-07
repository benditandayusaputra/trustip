package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
)

type ReadFile struct {
	files FileRecorder
	store FileStore
}

func NewReadFile(files FileRecorder, store FileStore) *ReadFile {
	return &ReadFile{files: files, store: store}
}

func (uc *ReadFile) Execute(ctx context.Context, signedIn actor.Actor, fileID string) (Content, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Content{}, fmt.Errorf("check file reader: %w", err)
	}
	file, err := uc.files.GetFile(ctx, signedIn.Membership.OrganizationID, fileID)
	if err != nil {
		return Content{}, fmt.Errorf("get file %s: %w", fileID, err)
	}
	data, err := uc.store.Read(file.Visibility, file.StorageKey)
	if err != nil {
		return Content{}, explainMissing(err)
	}
	return Content{ContentType: file.ContentType, Data: data}, nil
}

func explainMissing(err error) error {
	if errors.Is(err, domain.ErrFileMissing) {
		return apperror.NotFound().WithCause(err)
	}
	return fmt.Errorf("read stored file: %w", err)
}
