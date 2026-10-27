package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
)

const (
	currentPasswordField = "current_password"
	newPasswordField     = "new_password"
)

var ErrUnknownAvatar = errors.New("avatar is not an avatar file of this user")

type ProfileForm struct {
	FullName     string
	WhatsApp     string
	AvatarFileID string
}

type ProfileChange struct {
	FullName     string
	WhatsApp     contact.WhatsAppNumber
	AvatarFileID string
}

type AccountChange struct {
	UserID         string
	OrganizationID string
	Profile        ProfileChange
	At             time.Time
}

func NewProfileChange(form ProfileForm) (ProfileChange, error) {
	var problems []apperror.FieldError
	if problem, found := findFullNameProblem(form.FullName); found {
		problems = append(problems, problem)
	}
	whatsApp, err := contact.ParseWhatsAppNumber(form.WhatsApp)
	if err != nil {
		problems = append(problems, apperror.InvalidField(whatsAppField, apperror.FieldInvalidWhatsApp))
	}
	if len(problems) > 0 {
		return ProfileChange{}, apperror.Validation(problems...)
	}
	return ProfileChange{FullName: form.FullName, WhatsApp: whatsApp, AvatarFileID: form.AvatarFileID}, nil
}

type PasswordChangeForm struct {
	Current string
	New     string
}

type PasswordOwner struct {
	FullName string
	Email    string
	Password StoredPassword
}

func CheckNewPassword(form PasswordChangeForm, owner PasswordOwner, policy PasswordPolicy) error {
	if form.Current == "" {
		return apperror.Validation(apperror.InvalidField(currentPasswordField, apperror.FieldRequired))
	}
	if form.New == form.Current {
		return apperror.Validation(apperror.InvalidField(newPasswordField, apperror.FieldSameAsCurrent))
	}
	candidate := PasswordCandidate{Password: form.New, FullName: owner.FullName, Email: owner.Email}
	if problem, found := policy.FindProblem(candidate); found {
		problem.Field = newPasswordField
		return apperror.Validation(problem)
	}
	return nil
}

func WrongCurrentPassword() error {
	return apperror.Validation(apperror.InvalidField(currentPasswordField, apperror.FieldWrongCurrent))
}
