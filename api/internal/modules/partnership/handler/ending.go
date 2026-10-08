package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	blockersPath = partnershipPath + "/blockers"
	endPath      = partnershipPath + "/end"
	listingPath  = partnershipPath + "/public-listing"
)

type BlockerReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) ([]domain.Blocker, error)
}

type Ender interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.EndCommand) (domain.Partnership, error)
}

type ListingSetter interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.ListingCommand) (domain.Partnership, error)
}

type EndingDeps struct {
	Blockers     BlockerReader
	End          Ender
	Listing      ListingSetter
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func mountEnding(api fiber.Router, deps Deps, files fileLinks) {
	ending := deps.Ending
	writeLimit := httpserver.RateLimit(ending.WriteLimiter, httpserver.BySignedInUser, ending.SecurityLog)
	api.Get(blockersPath, deps.Authenticate, blockers(ending.Blockers))
	api.Post(endPath, deps.Authenticate, writeLimit, end(ending.End, files))
	api.Patch(listingPath, deps.Authenticate, writeLimit, listing(ending.Listing, files))
}

type blockerResponse struct {
	Kind           string `json:"kind"`
	DocumentID     string `json:"document_id"`
	DocumentNumber string `json:"document_number"`
	Qty            int    `json:"qty"`
	Amount         int64  `json:"amount"`
	Description    string `json:"description"`
}

func blockers(reader BlockerReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, partnershipID, err := actorAndID(c)
		if err != nil {
			return err
		}
		found, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read blockers: %w", err)
		}
		items := make([]blockerResponse, 0, len(found))
		for _, blocker := range found {
			items = append(items, blockerResponse{
				Kind: string(blocker.Kind), DocumentID: blocker.DocumentID, DocumentNumber: blocker.DocumentNumber,
				Qty: blocker.Qty, Amount: blocker.Amount, Description: blocker.Describe(),
			})
		}
		return httpserver.RespondList(c, httpserver.Page[blockerResponse]{Items: items})
	}
}

type endRequest struct {
	Reason string `json:"reason"`
}

func end(ender Ender, files fileLinks) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, partnershipID, err := actorAndID(c)
		if err != nil {
			return err
		}
		var request endRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read ending: %w", err)
		}
		reason := httpserver.MultilineInput("reason", request.Reason)
		if err := httpserver.RejectUncleanInputs(reason); err != nil {
			return fmt.Errorf("check end reason: %w", err)
		}
		ended, err := ender.Execute(c.Context(), signedIn, service.EndCommand{
			PartnershipID: partnershipID, Reason: reason.Value,
		})
		if err != nil {
			return fmt.Errorf("end partnership: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Message: "Kemitraan berakhir.", Data: files.partnership(ended)})
	}
}

type listingRequest struct {
	ShowOnPublic bool `json:"show_on_public"`
}

func listing(setter ListingSetter, files fileLinks) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, partnershipID, err := actorAndID(c)
		if err != nil {
			return err
		}
		var request listingRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read public listing: %w", err)
		}
		changed, err := setter.Execute(c.Context(), signedIn, service.ListingCommand{
			PartnershipID: partnershipID, ShowOnPublic: request.ShowOnPublic,
		})
		if err != nil {
			return fmt.Errorf("set public listing: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: files.partnership(changed)})
	}
}

func actorAndID(c fiber.Ctx) (actor.Actor, string, error) {
	signedIn, err := signedInActor(c)
	if err != nil {
		return actor.Actor{}, "", err
	}
	partnershipID, err := httpserver.IDParam(c, "id")
	if err != nil {
		return actor.Actor{}, "", fmt.Errorf("read partnership id: %w", err)
	}
	return signedIn, partnershipID, nil
}
