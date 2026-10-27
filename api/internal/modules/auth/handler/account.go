package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const privateFilesPath = httpserver.APIPrefix + "/files/"

type AccountGetter interface {
	Execute(ctx context.Context, signedIn actor.Actor) (domain.Account, error)
}

type AccountUpdater interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.ProfileForm) (domain.Account, error)
}

type PasswordChanger interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.PasswordChangeForm) error
}

type accountProfileResponse struct {
	ID         string              `json:"id"`
	FullName   string              `json:"full_name"`
	Email      string              `json:"email"`
	WhatsApp   string              `json:"whatsapp"`
	Avatar     *imageResponse      `json:"avatar"`
	Membership *membershipResponse `json:"membership"`
}

type imageResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type accountUpdateRequest struct {
	FullName     string  `json:"full_name"`
	WhatsApp     string  `json:"whatsapp"`
	AvatarFileID *string `json:"avatar_file_id"`
}

type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func currentAccount(accounts AccountGetter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		account, err := accounts.Execute(c.Context(), signedIn)
		if err != nil {
			return fmt.Errorf("read current account: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toAccountProfile(account)})
	}
}

func updateAccount(accounts AccountUpdater) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		var request accountUpdateRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read account update: %w", err)
		}
		form, err := request.toForm()
		if err != nil {
			return err
		}
		account, err := accounts.Execute(c.Context(), signedIn, form)
		if err != nil {
			return fmt.Errorf("update account: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toAccountProfile(account)})
	}
}

func changePassword(passwords PasswordChanger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		var request passwordChangeRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read password change: %w", err)
		}
		form := domain.PasswordChangeForm{
			Current: httpserver.NormalizePassword(request.CurrentPassword),
			New:     httpserver.NormalizePassword(request.NewPassword),
		}
		if err := passwords.Execute(c.Context(), signedIn, form); err != nil {
			return fmt.Errorf("change password: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{})
	}
}

func (r accountUpdateRequest) toForm() (domain.ProfileForm, error) {
	fullName := httpserver.LineInput("full_name", r.FullName)
	whatsApp := httpserver.LineInput("whatsapp", r.WhatsApp)
	if err := httpserver.RejectUncleanInputs(fullName, whatsApp); err != nil {
		return domain.ProfileForm{}, fmt.Errorf("check account text: %w", err)
	}
	form := domain.ProfileForm{FullName: fullName.Value, WhatsApp: whatsApp.Value}
	if r.AvatarFileID != nil {
		form.AvatarFileID = *r.AvatarFileID
	}
	return form, nil
}

func toAccountProfile(account domain.Account) accountProfileResponse {
	signedIn := toSignedInAccount(account)
	response := accountProfileResponse{
		ID: account.ID, FullName: account.FullName, Email: account.Email, WhatsApp: account.WhatsApp,
		Membership: signedIn.Membership,
	}
	if account.AvatarFileID != "" {
		response.Avatar = &imageResponse{ID: account.AvatarFileID, URL: privateFilesPath + account.AvatarFileID}
	}
	return response
}
