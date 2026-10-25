package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	filedomain "github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	filerepository "github.com/benditandayusaputra/trustip/api/internal/modules/file/repository"
	fileservice "github.com/benditandayusaputra/trustip/api/internal/modules/file/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/storage"
)

//go:embed demo_photos
var demoPhotoFiles embed.FS

const (
	organizationPhotos = "organizations/"
	productPhotos      = "products/"
)

type demoPhotos struct {
	upload *fileservice.UploadImage
	disk   *storage.Disk
}

func newDemoPhotos(pool *database.Pool, manual *clock.Manual, fileDir string) demoPhotos {
	disk := storage.NewDisk(fileDir)
	return demoPhotos{
		upload: fileservice.NewUploadImage(fileservice.UploadDeps{
			Images: imageRecoder{}, Store: diskStore{disk: disk}, Files: filerepository.NewFiles(pool),
			Clock: manual, IDs: idgen.UUIDv7{},
		}),
		disk: disk,
	}
}

func (p demoPhotos) put(
	ctx context.Context, owner actor.Actor, purpose filedomain.Purpose, name string,
) (string, error) {
	data, err := demoPhotoFiles.ReadFile("demo_photos/" + name + ".jpg")
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read demo photo %s: %w", name, err)
	}
	file, err := p.upload.Execute(ctx, owner, fileservice.UploadCommand{Purpose: string(purpose), Data: data})
	if err != nil {
		return "", fmt.Errorf("upload demo photo %s: %w", name, err)
	}
	return file.ID, nil
}

func (p demoPhotos) forget(stale []sqlc.ListDemoFilesRow) error {
	for _, file := range stale {
		err := p.disk.Delete(areaOf(filedomain.Visibility(file.Visibility)), file.StorageKey)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("delete demo file %s: %w", file.StorageKey, err)
		}
	}
	return nil
}
