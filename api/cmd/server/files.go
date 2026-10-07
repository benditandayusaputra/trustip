package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	filedomain "github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	filehandler "github.com/benditandayusaputra/trustip/api/internal/modules/file/handler"
	filerepository "github.com/benditandayusaputra/trustip/api/internal/modules/file/repository"
	fileservice "github.com/benditandayusaputra/trustip/api/internal/modules/file/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/imaging"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/storage"
)

const uploadsPerHourPerOrganization = 20

type fileComponents struct {
	store         fileservice.FileStore
	files         fileservice.FileRecorder
	clock         fileservice.Clock
	publicBaseURL string
}

func buildFileComponents(pool *database.Pool, cfg config.Config, clock fileservice.Clock) fileComponents {
	return fileComponents{
		store:         diskStore{disk: storage.NewDisk(cfg.FileStorageDir)},
		files:         filerepository.NewFiles(pool),
		clock:         clock,
		publicBaseURL: cfg.PublicFilesBaseURL,
	}
}

type fileRoutes struct {
	authenticate fiber.Handler
	limiterClock ratelimit.Clock
	securityLog  *slog.Logger
}

func fileDeps(components fileComponents, routes fileRoutes) filehandler.Deps {
	return filehandler.Deps{
		Authenticate: routes.authenticate,
		Upload: fileservice.NewUploadImage(fileservice.UploadDeps{
			Images: imageRecoder{}, Store: components.store, Files: components.files,
			Clock: components.clock, IDs: idgen.UUIDv7{},
		}),
		Read:       fileservice.NewReadFile(components.files, components.store),
		ReadPublic: fileservice.NewReadPublicFile(components.store),
		UploadLimiter: ratelimit.New(
			ratelimit.Rule{Limit: uploadsPerHourPerOrganization, Window: time.Hour}, routes.limiterClock,
		),
		SecurityLog:   routes.securityLog,
		PublicBaseURL: components.publicBaseURL,
	}
}

type imageRecoder struct{}

func (imageRecoder) Recode(data []byte) (filedomain.Image, error) {
	recoded, err := imaging.Recode(data)
	switch {
	case errors.Is(err, imaging.ErrUnsupportedImage):
		return filedomain.Image{}, filedomain.ErrUnsupportedImage
	case errors.Is(err, imaging.ErrImageTooLarge):
		return filedomain.Image{}, filedomain.ErrImageTooLarge
	case err != nil:
		return filedomain.Image{}, fmt.Errorf("recode image: %w", err)
	}
	return filedomain.Image{
		Data: recoded.Data, ContentType: recoded.ContentType, Extension: recoded.Extension,
		Width: recoded.Width, Height: recoded.Height,
	}, nil
}

type diskStore struct {
	disk *storage.Disk
}

func areaOf(visibility filedomain.Visibility) storage.Area {
	if visibility == filedomain.Public {
		return storage.Public
	}
	return storage.Private
}

func (s diskStore) Put(
	ctx context.Context, visibility filedomain.Visibility, data []byte, extension string,
) (string, error) {
	key, err := s.disk.Put(ctx, areaOf(visibility), data, extension)
	if err != nil {
		return "", fmt.Errorf("put file: %w", err)
	}
	return key, nil
}

func (s diskStore) Read(visibility filedomain.Visibility, key string) ([]byte, error) {
	data, err := s.disk.Read(areaOf(visibility), key)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, filedomain.ErrFileMissing
	}
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}
