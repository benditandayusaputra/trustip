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

type UploadDeps struct {
	Images ImageRecoder
	Store  FileStore
	Files  FileRecorder
	Clock  Clock
	IDs    IDGenerator
}

type UploadCommand struct {
	Purpose string
	Data    []byte
}

type UploadImage struct {
	deps UploadDeps
}

func NewUploadImage(deps UploadDeps) *UploadImage {
	return &UploadImage{deps: deps}
}

func (uc *UploadImage) Execute(ctx context.Context, signedIn actor.Actor, cmd UploadCommand) (domain.File, error) {
	if err := policy.Require(signedIn, domain.RequiredAction(cmd.Purpose)); err != nil {
		return domain.File{}, fmt.Errorf("check upload permission: %w", err)
	}
	purpose, err := domain.ParsePurpose(cmd.Purpose, signedIn.Membership.OrganizationType)
	if err != nil {
		return domain.File{}, fmt.Errorf("check upload purpose: %w", err)
	}
	image, err := uc.recode(cmd.Data)
	if err != nil {
		return domain.File{}, err
	}
	id, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.File{}, fmt.Errorf("generate file id: %w", err)
	}
	key, err := uc.deps.Store.Put(ctx, purpose.Visibility(), image.Data, image.Extension)
	if err != nil {
		return domain.File{}, fmt.Errorf("store file %s: %w", id, err)
	}
	upload := domain.Upload{
		ID: id, OrganizationID: signedIn.Membership.OrganizationID, CreatedBy: signedIn.UserID, Purpose: purpose,
	}
	file := domain.NewFile(upload, domain.Stored{Image: image, Key: key}, uc.deps.Clock.Now())
	if err := uc.deps.Files.CreateFile(ctx, file); err != nil {
		return domain.File{}, fmt.Errorf("record file %s: %w", id, err)
	}
	return file, nil
}

func (uc *UploadImage) recode(data []byte) (domain.Image, error) {
	if err := domain.CheckUploadSize(len(data)); err != nil {
		return domain.Image{}, fmt.Errorf("check upload size: %w", err)
	}
	image, err := uc.deps.Images.Recode(data)
	switch {
	case errors.Is(err, domain.ErrUnsupportedImage):
		return domain.Image{}, apperror.Validation(apperror.InvalidField("file", apperror.FieldInvalidImage))
	case errors.Is(err, domain.ErrImageTooLarge):
		return domain.Image{}, apperror.Validation(apperror.InvalidField("file", apperror.FieldImageTooLarge))
	case err != nil:
		return domain.Image{}, fmt.Errorf("recode image: %w", err)
	}
	if err := domain.CheckUploadSize(len(image.Data)); err != nil {
		return domain.Image{}, fmt.Errorf("check recoded size: %w", err)
	}
	return image, nil
}
