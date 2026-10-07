package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
)

type ImageRecoder interface {
	Recode(data []byte) (domain.Image, error)
}

type FileStore interface {
	Put(ctx context.Context, visibility domain.Visibility, data []byte, extension string) (string, error)
	Read(visibility domain.Visibility, key string) ([]byte, error)
}

type FileRecorder interface {
	CreateFile(ctx context.Context, file domain.File) error
	GetFile(ctx context.Context, organizationID, fileID string) (domain.File, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

type Content struct {
	ContentType string
	Data        []byte
}
