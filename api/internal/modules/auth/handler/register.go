package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
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
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
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

type normalizedField struct {
	name  string
	value string
	ok    bool
}

func (r registerRequest) toCommand(clientIP string) (service.RegisterCommand, error) {
	fullName := normalizeField("full_name", r.FullName)
	email := normalizeField("email", r.Email)
	whatsApp := normalizeField("whatsapp", r.WhatsApp)
	if err := rejectForbiddenCharacters(fullName, email, whatsApp); err != nil {
		return service.RegisterCommand{}, err
	}
	return service.RegisterCommand{
		Form: domain.RegistrationForm{
			FullName:        fullName.value,
			Email:           strings.ToLower(email.value),
			Password:        httpserver.NormalizePassword(r.Password),
			WhatsApp:        whatsApp.value,
			PrivacyAccepted: r.PrivacyAccepted,
		},
		CaptchaToken: r.CaptchaToken,
		ClientIP:     clientIP,
	}, nil
}

func normalizeField(name, raw string) normalizedField {
	value, ok := httpserver.NormalizeLine(raw)
	return normalizedField{name: name, value: value, ok: ok}
}

func rejectForbiddenCharacters(fields ...normalizedField) error {
	var problems []apperror.FieldError
	for _, field := range fields {
		if !field.ok {
			problems = append(problems, apperror.InvalidField(field.name, apperror.FieldInvalidCharacters))
		}
	}
	if len(problems) > 0 {
		return apperror.Validation(problems...)
	}
	return nil
}

func toAccountResponse(user domain.User) accountResponse {
	return accountResponse{ID: user.ID, FullName: user.FullName, Email: user.Email}
}
