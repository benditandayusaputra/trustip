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

type StoreProfileEditor interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.UpdateStoreProfileCommand) (domain.Settings, error)
}

type CategoryLister interface {
	Execute(ctx context.Context) ([]domain.Category, error)
}

type storeProfileFields struct {
	StoreType                  *string  `json:"store_type"`
	AcceptedCategoryIDs        []string `json:"accepted_category_ids"`
	AcceptsFixedDeposit        bool     `json:"accepts_fixed_deposit"`
	AcceptsCommission          bool     `json:"accepts_commission"`
	CommissionMinBPS           *int     `json:"commission_min_bps"`
	CommissionMaxBPS           *int     `json:"commission_max_bps"`
	RequiresDistributionPermit bool     `json:"requires_distribution_permit"`
	RequiresHalal              bool     `json:"requires_halal"`
	RequiresSample             bool     `json:"requires_sample"`
	OtherRequirements          string   `json:"other_requirements"`
	OpeningHours               string   `json:"opening_hours"`
}

type storeProfileRequest struct {
	storeProfileFields
	PhotoFileIDs []string `json:"photo_file_ids"`
}

type storeProfileResponse struct {
	storeProfileFields
	Photos []imageRef `json:"photos"`
}

type categoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func updateStoreProfile(editor StoreProfileEditor, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		var request storeProfileRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read store profile: %w", err)
		}
		form, err := request.toForm()
		if err != nil {
			return err
		}
		settings, err := editor.Execute(c.Context(), signedIn,
			service.UpdateStoreProfileCommand{OrganizationID: organizationID, Form: form})
		if err != nil {
			return fmt.Errorf("update store profile: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func (r storeProfileRequest) toForm() (domain.StoreProfileForm, error) {
	inputs := []httpserver.TextInput{
		httpserver.LineInput("store_type", valueOrEmpty(r.StoreType)),
		httpserver.MultilineInput("other_requirements", r.OtherRequirements),
		httpserver.LineInput("opening_hours", r.OpeningHours),
	}
	if err := httpserver.RejectUncleanInputs(inputs...); err != nil {
		return domain.StoreProfileForm{}, fmt.Errorf("check store profile text: %w", err)
	}
	return domain.StoreProfileForm{
		StoreType:   inputs[0].Value,
		CategoryIDs: orEmpty(r.AcceptedCategoryIDs),
		Acceptance: domain.Acceptance{
			AcceptsFixedDeposit: r.AcceptsFixedDeposit, AcceptsCommission: r.AcceptsCommission,
		},
		Commission: commissionOf(r.CommissionMinBPS, r.CommissionMaxBPS),
		Conditions: domain.StoreConditions{
			RequiresDistributionPermit: r.RequiresDistributionPermit, RequiresHalal: r.RequiresHalal,
			RequiresSample: r.RequiresSample, OtherRequirements: inputs[1].Value,
		},
		OpeningHours: inputs[2].Value,
		PhotoFileIDs: orEmpty(r.PhotoFileIDs),
	}, nil
}

func commissionOf(minimum, maximum *int) domain.CommissionRange {
	if minimum == nil || maximum == nil {
		return domain.CommissionRange{}
	}
	return domain.CommissionRange{Given: true, MinBPS: *minimum, MaxBPS: *maximum}
}

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func toStoreProfileResponse(store domain.StoreSettings, filesBaseURL string) *storeProfileResponse {
	profile := store.Profile
	response := &storeProfileResponse{Photos: make([]imageRef, 0, len(store.Photos))}
	response.storeProfileFields = storeProfileFields{
		AcceptedCategoryIDs: orEmpty(profile.CategoryIDs), AcceptsFixedDeposit: profile.AcceptsFixedDeposit,
		AcceptsCommission: profile.AcceptsCommission, RequiresHalal: profile.Conditions.RequiresHalal,
		RequiresDistributionPermit: profile.Conditions.RequiresDistributionPermit,
		RequiresSample:             profile.Conditions.RequiresSample,
		OtherRequirements:          profile.Conditions.OtherRequirements, OpeningHours: profile.OpeningHours,
	}
	if profile.StoreType != "" {
		response.StoreType = &profile.StoreType
	}
	if profile.Commission.Given {
		response.CommissionMinBPS = &profile.Commission.MinBPS
		response.CommissionMaxBPS = &profile.Commission.MaxBPS
	}
	for _, photo := range store.Photos {
		response.Photos = append(response.Photos, imageRef{ID: photo.ID, URL: filesBaseURL + "/" + photo.StorageKey})
	}
	return response
}

func listCategories(categories CategoryLister) fiber.Handler {
	return func(c fiber.Ctx) error {
		found, err := categories.Execute(c.Context())
		if err != nil {
			return fmt.Errorf("list categories: %w", err)
		}
		items := make([]categoryResponse, 0, len(found))
		for _, category := range found {
			items = append(items, categoryResponse{ID: category.ID, Name: category.Name, Slug: category.Slug})
		}
		return httpserver.RespondList(c, httpserver.Page[categoryResponse]{Items: items})
	}
}
