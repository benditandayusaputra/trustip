package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type AccountGetter interface {
	Execute(ctx context.Context, signedIn actor.Actor) (domain.Account, error)
}

func currentAccount(accounts AccountGetter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		account, err := accounts.Execute(c.Context(), signedIn)
		if err != nil {
			return fmt.Errorf("read current account: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSignedInAccount(account)})
	}
}
