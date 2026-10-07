package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type SettingsReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, organizationID string) (domain.Settings, error)
}

type ProfileEditor interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.UpdateProfileCommand) (domain.Settings, error)
}

type profileUpdateRequest struct {
	profileRequest
	LogoFileID *string `json:"logo_file_id"`
}

type imageRef struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func readSettings(reader SettingsReader, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		settings, err := reader.Execute(c.Context(), signedIn, organizationID)
		if err != nil {
			return fmt.Errorf("read settings: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func updateProfile(editor ProfileEditor, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		var request profileUpdateRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read profile update: %w", err)
		}
		profile, err := request.toForm()
		if err != nil {
			return err
		}
		form := domain.ProfileUpdateForm{ProfileForm: profile, LogoFileID: valueOrEmpty(request.LogoFileID)}
		settings, err := editor.Execute(c.Context(), signedIn,
			service.UpdateProfileCommand{OrganizationID: organizationID, Form: form})
		if err != nil {
			return fmt.Errorf("update profile: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func actorAndOrganization(c fiber.Ctx) (actor.Actor, string, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, "", apperror.New(apperror.CodeUnauthenticated)
	}
	organizationID, err := httpserver.IDParam(c, "id")
	if err != nil {
		return actor.Actor{}, "", fmt.Errorf("read organization id: %w", err)
	}
	return signedIn, organizationID, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func toSettingsResponse(settings domain.Settings, filesBaseURL string) organizationResponse {
	response := toOrganizationResponse(domain.Organization{
		ID:         settings.ID,
		Onboarding: domain.Onboarding{Type: settings.Type, Profile: settings.Profile},
		Slug:       settings.Slug,
		CityName:   settings.CityName,
	})
	response.Listing = toListingResponse(settings.Listing)
	if settings.Logo.ID != "" {
		response.Logo = &imageRef{ID: settings.Logo.ID, URL: filesBaseURL + "/" + settings.Logo.StorageKey}
	}
	if settings.Type == actor.Store {
		response.StoreProfile = toStoreProfileResponse(settings.Store, filesBaseURL)
		return response
	}
	response.ConsignorProfile = toConsignorProfileResponse(settings.Consignor)
	response.Payment = toPaymentResponse(settings.Payment)
	return response
}
