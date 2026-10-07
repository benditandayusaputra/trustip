package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type ListingEditor interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.SetListingCommand) (domain.Settings, error)
}

type listingRequest struct {
	IsListed bool `json:"is_listed"`
}

type listingItem struct {
	Requirement string `json:"requirement"`
	IsRequired  bool   `json:"is_required"`
	IsDone      bool   `json:"is_done"`
}

type listingResponse struct {
	IsListed bool          `json:"is_listed"`
	Percent  int           `json:"percent"`
	Items    []listingItem `json:"items"`
}

func setListing(editor ListingEditor, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		var request listingRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read listing: %w", err)
		}
		settings, err := editor.Execute(c.Context(), signedIn,
			service.SetListingCommand{OrganizationID: organizationID, IsListed: request.IsListed})
		if err != nil {
			return fmt.Errorf("set listing: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func toListingResponse(listing domain.Listing) *listingResponse {
	items := make([]listingItem, 0, len(listing.Completeness.Items))
	for _, item := range listing.Completeness.Items {
		items = append(items, listingItem{
			Requirement: string(item.Requirement), IsRequired: item.IsRequired, IsDone: item.IsDone,
		})
	}
	return &listingResponse{IsListed: listing.IsListed, Percent: listing.Completeness.Percent(), Items: items}
}
