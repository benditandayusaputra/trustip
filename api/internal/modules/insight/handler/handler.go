package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const dropSuggestionsPath = "/partnerships/:id/drop-suggestions"

type DropSuggester interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) ([]domain.DropSuggestion, error)
}

type Deps struct {
	Authenticate fiber.Handler
	DropSuggest  DropSuggester
}

func Mount(api fiber.Router, deps Deps) {
	api.Get(dropSuggestionsPath, deps.Authenticate, dropSuggestions(deps.DropSuggest))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func dropSuggestions(suggester DropSuggester) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		suggestions, err := suggester.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("suggest drops: %w", err)
		}
		return httpserver.RespondList(c, httpserver.Page[dropSuggestionResponse]{Items: toDropSuggestions(suggestions)})
	}
}
