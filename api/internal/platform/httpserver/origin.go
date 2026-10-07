package httpserver

import (
	"mime"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	HeaderClient = "X-Trustip-Client"
	webClient    = "web"
)

func RequireSameOrigin(appOrigin string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if isReadOnly(c.Method()) {
			return c.Next()
		}
		if c.Get(fiber.HeaderOrigin) != appOrigin || c.Get(HeaderClient) != webClient {
			return apperror.New(apperror.CodeCrossSiteRejected)
		}
		if len(c.Body()) > 0 && !isWritableBody(c.Get(fiber.HeaderContentType)) {
			return apperror.New(apperror.CodeUnsupportedMediaType)
		}
		return c.Next()
	}
}

func isReadOnly(method string) bool {
	return method == fiber.MethodGet || method == fiber.MethodHead
}

func isWritableBody(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && (mediaType == fiber.MIMEApplicationJSON || mediaType == fiber.MIMEMultipartForm)
}
