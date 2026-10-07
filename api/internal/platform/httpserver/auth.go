package httpserver

import (
	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type AccessVerifier interface {
	VerifyAccess(token string) (actor.Actor, error)
}

type actorKey struct{}

func Authenticate(verifier AccessVerifier, cookieName string) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(cookieName)
		if token == "" {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		signedIn, err := verifier.VerifyAccess(token)
		if err != nil {
			return apperror.New(apperror.CodeSessionExpired).WithCause(err)
		}
		c.Locals(actorKey{}, signedIn)
		return c.Next()
	}
}

func ActorFrom(c fiber.Ctx) (actor.Actor, bool) {
	signedIn, ok := c.Locals(actorKey{}).(actor.Actor)
	return signedIn, ok
}
