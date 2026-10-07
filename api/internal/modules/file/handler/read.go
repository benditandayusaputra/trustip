package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

func readFile(reader Reader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		fileID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read file id: %w", err)
		}
		content, err := reader.Execute(c.Context(), signedIn, fileID)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		return httpserver.RespondFile(c, httpserver.FileReply{ContentType: content.ContentType, Data: content.Data})
	}
}

func readPublicFile(reader PublicReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		content, err := reader.Execute(c.Context(), c.Params("key"))
		if err != nil {
			return fmt.Errorf("read public file: %w", err)
		}
		return httpserver.RespondFile(c, httpserver.FileReply{
			ContentType: content.ContentType, Data: content.Data, IsPublic: true,
		})
	}
}
