package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type onboardingRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WhatsApp    string `json:"whatsapp"`
	CityCode    string `json:"city_code"`
	District    string `json:"district"`
	Address     string `json:"address"`
}

type organizationResponse struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	WhatsApp    string `json:"whatsapp"`
	CityCode    string `json:"city_code"`
	CityName    string `json:"city_name"`
	District    string `json:"district"`
	Address     string `json:"address"`
}

func createOrganization(creator Creator) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		var request onboardingRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read onboarding: %w", err)
		}
		form, err := request.toForm()
		if err != nil {
			return err
		}
		organization, err := creator.Execute(c.Context(), signedIn, form)
		if err != nil {
			return fmt.Errorf("create organization: %w", err)
		}
		c.Location(httpserver.APIPrefix + organizationsPath + "/" + organization.ID)
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Data: toOrganizationResponse(organization),
		})
	}
}

func (r onboardingRequest) toForm() (domain.OnboardingForm, error) {
	inputs := []httpserver.TextInput{
		httpserver.LineInput("type", r.Type),
		httpserver.LineInput("name", r.Name),
		httpserver.MultilineInput("description", r.Description),
		httpserver.LineInput("whatsapp", r.WhatsApp),
		httpserver.LineInput("city_code", r.CityCode),
		httpserver.LineInput("district", r.District),
		httpserver.LineInput("address", r.Address),
	}
	if err := httpserver.RejectUncleanInputs(inputs...); err != nil {
		return domain.OnboardingForm{}, fmt.Errorf("check onboarding text: %w", err)
	}
	return domain.OnboardingForm{
		Type: inputs[0].Value, Name: inputs[1].Value, Description: inputs[2].Value, WhatsApp: inputs[3].Value,
		CityCode: inputs[4].Value, District: inputs[5].Value, Address: inputs[6].Value,
	}, nil
}

func toOrganizationResponse(organization domain.Organization) organizationResponse {
	return organizationResponse{
		ID:          organization.ID,
		Type:        string(organization.Type),
		Name:        organization.Name,
		Slug:        string(organization.Slug),
		Description: organization.Description,
		WhatsApp:    organization.WhatsApp.String(),
		CityCode:    organization.CityCode,
		CityName:    organization.CityName,
		District:    organization.District,
		Address:     organization.Address,
	}
}
