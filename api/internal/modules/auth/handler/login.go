package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type loginRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
}

func login(authenticator Authenticator, cookies CookiePolicy) fiber.Handler {
	return func(c fiber.Ctx) error {
		var request loginRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read login: %w", err)
		}
		email := normalizeField("email", request.Email)
		if err := rejectForbiddenCharacters(email); err != nil {
			return err
		}
		result, err := authenticator.Execute(c.Context(), service.LoginCommand{
			Email:        strings.ToLower(email.value),
			Password:     httpserver.NormalizePassword(request.Password),
			CaptchaToken: request.CaptchaToken,
			ClientIP:     httpserver.ClientIP(c),
		})
		if err != nil {
			return fmt.Errorf("log in: %w", err)
		}
		cookies.setSession(c, result.Tokens)
		return httpserver.Respond(c, httpserver.Reply{Data: toSignedInAccount(result.Account)})
	}
}

func byLoginEmail(c fiber.Ctx) string {
	var probe struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(c.Body(), &probe); err != nil {
		return "email:unreadable"
	}
	return "email:" + strings.ToLower(strings.TrimSpace(probe.Email))
}

func toSignedInAccount(account domain.Account) accountResponse {
	return accountResponse{ID: account.ID, FullName: account.FullName, Email: account.Email}
}
