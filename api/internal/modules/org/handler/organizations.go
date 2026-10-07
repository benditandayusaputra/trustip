package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type onboardingRequest struct {
	Type string `json:"type"`
	profileRequest
}

type profileRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WhatsApp    string `json:"whatsapp"`
	CityCode    string `json:"city_code"`
	District    string `json:"district"`
	Address     string `json:"address"`
	Location    *point `json:"location"`
}

type point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type organizationResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	Description  string                `json:"description"`
	WhatsApp     string                `json:"whatsapp"`
	CityCode     string                `json:"city_code"`
	CityName     string                `json:"city_name"`
	District     string                `json:"district"`
	Address      string                `json:"address"`
	Location     *point                `json:"location"`
	Logo         *imageRef             `json:"logo"`
	StoreProfile *storeProfileResponse `json:"store_profile"`
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
	organizationType := httpserver.LineInput("type", r.Type)
	if err := httpserver.RejectUncleanInputs(organizationType); err != nil {
		return domain.OnboardingForm{}, fmt.Errorf("check organization type: %w", err)
	}
	profile, err := r.profileRequest.toForm()
	if err != nil {
		return domain.OnboardingForm{}, err
	}
	return domain.OnboardingForm{Type: organizationType.Value, ProfileForm: profile}, nil
}

func (r profileRequest) toForm() (domain.ProfileForm, error) {
	inputs := []httpserver.TextInput{
		httpserver.LineInput("name", r.Name),
		httpserver.MultilineInput("description", r.Description),
		httpserver.LineInput("whatsapp", r.WhatsApp),
		httpserver.LineInput("city_code", r.CityCode),
		httpserver.LineInput("district", r.District),
		httpserver.LineInput("address", r.Address),
	}
	if err := httpserver.RejectUncleanInputs(inputs...); err != nil {
		return domain.ProfileForm{}, fmt.Errorf("check profile text: %w", err)
	}
	return domain.ProfileForm{
		Name: inputs[0].Value, Description: inputs[1].Value, WhatsApp: inputs[2].Value,
		CityCode: inputs[3].Value, District: inputs[4].Value, Address: inputs[5].Value,
		Location: r.Location.toInput(),
	}, nil
}

func (p *point) toInput() domain.LocationInput {
	if p == nil {
		return domain.LocationInput{}
	}
	return domain.LocationInput{Given: true, Latitude: p.Latitude, Longitude: p.Longitude}
}

func toOrganizationResponse(organization domain.Organization) organizationResponse {
	response := organizationResponse{
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
	if organization.HasLocation {
		exact := organization.Location.Exact
		response.Location = &point{Latitude: exact.Latitude, Longitude: exact.Longitude}
	}
	return response
}
