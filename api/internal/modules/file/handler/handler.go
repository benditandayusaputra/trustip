package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	filesPath      = "/files"
	fileByIDPath   = filesPath + "/:id"
	publicFilePath = httpserver.PublicFilesPrefix + ":key"
)

type Uploader interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.UploadCommand) (domain.File, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, fileID string) (service.Content, error)
}

type PublicReader interface {
	Execute(ctx context.Context, key string) (service.Content, error)
}

type Deps struct {
	Authenticate  fiber.Handler
	Upload        Uploader
	Read          Reader
	ReadPublic    PublicReader
	UploadLimiter httpserver.RateLimiter
	SecurityLog   *slog.Logger
	PublicBaseURL string
}

func Mount(api fiber.Router, deps Deps) {
	uploadLimit := httpserver.RateLimit(deps.UploadLimiter, httpserver.ByOrganization, deps.SecurityLog)
	api.Post(filesPath, deps.Authenticate, uploadLimit, uploadFile(deps))
	api.Get(fileByIDPath, deps.Authenticate, readFile(deps.Read))
}

func MountPublic(root fiber.Router, deps Deps) {
	root.Get(publicFilePath, readPublicFile(deps.ReadPublic))
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{{Method: fiber.MethodGet, Path: publicFilePath}}
}
