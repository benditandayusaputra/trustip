package httpserver

import (
	"fmt"
	"regexp"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const HeaderRequestID = "X-Request-Id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

type requestIDKey struct{}

type IDGenerator interface {
	NewID() (string, error)
}

func RequestID(generator IDGenerator) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(HeaderRequestID)
		if !validRequestID.MatchString(id) {
			generated, err := generator.NewID()
			if err != nil {
				return apperror.Internal(fmt.Errorf("generate request id: %w", err))
			}
			id = generated
		}
		c.Locals(requestIDKey{}, id)
		c.Set(HeaderRequestID, id)
		return c.Next()
	}
}

func RequestIDFrom(c fiber.Ctx) string {
	id, _ := c.Locals(requestIDKey{}).(string)
	return id
}
