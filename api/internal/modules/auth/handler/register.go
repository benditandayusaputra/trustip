package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type registerRequest struct {
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	WhatsApp        string `json:"whatsapp"`
	PrivacyAccepted bool   `json:"privacy_accepted"`
	CaptchaToken    string `json:"captcha_token"`
}

type accountResponse struct {
	ID         string              `json:"id"`
	FullName   string              `json:"full_name"`
	Email      string              `json:"email"`
	Membership *membershipResponse `json:"membership"`
}

type membershipResponse struct {
	OrganizationID   string `json:"organization_id"`
	OrganizationType string `json:"organization_type"`
	Role             string `json:"role"`
}

func register(registrar Registrar) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request registerRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read registration: %w", err)
		}
		command, err := request.toCommand(httpserver.ClientIP(c))
		if err != nil {
			return err
		}
		user, err := registrar.Execute(c.Context(), command)
		if err != nil {
			return fmt.Errorf("register user: %w", err)
		}
		c.Location(accountPath)
		return httpserver.Respond(c, httpserver.Reply{Status: fiber.StatusCreated, Data: toAccountResponse(user)})
	}
}

func (r registerRequest) toCommand(clientIP string) (service.RegisterCommand, error) {
	fullName := httpserver.LineInput("full_name", r.FullName)
	email := httpserver.LineInput("email", r.Email)
	whatsApp := httpserver.LineInput("whatsapp", r.WhatsApp)
	if err := httpserver.RejectUncleanInputs(fullName, email, whatsApp); err != nil {
		return service.RegisterCommand{}, fmt.Errorf("check registration text: %w", err)
	}
	return service.RegisterCommand{
		Form: domain.RegistrationForm{
			FullName:        fullName.Value,
			Email:           strings.ToLower(email.Value),
			Password:        httpserver.NormalizePassword(r.Password),
			WhatsApp:        whatsApp.Value,
			PrivacyAccepted: r.PrivacyAccepted,
		},
		CaptchaToken: r.CaptchaToken,
		ClientIP:     clientIP,
	}, nil
}

func toAccountResponse(user domain.User) accountResponse {
	return accountResponse{ID: user.ID, FullName: user.FullName, Email: user.Email}
}
