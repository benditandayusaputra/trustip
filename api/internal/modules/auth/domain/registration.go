package domain

import (
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
)

const (
	MinFullNameLength = 2
	MaxFullNameLength = 80
	MaxEmailLength    = 254
	fullNameField     = "full_name"
	emailField        = "email"
	whatsAppField     = "whatsapp"
	privacyField      = "privacy_accepted"
)

type RegistrationForm struct {
	FullName        string
	Email           string
	Password        string
	WhatsApp        string
	PrivacyAccepted bool
}

type Registration struct {
	FullName string
	Email    string
	WhatsApp contact.WhatsAppNumber
	Password string
}

func NewRegistration(form RegistrationForm, policy PasswordPolicy) (Registration, error) {
	email := strings.ToLower(strings.TrimSpace(form.Email))
	whatsApp, whatsAppErr := contact.ParseWhatsAppNumber(form.WhatsApp)
	var problems []apperror.FieldError
	if problem, found := findFullNameProblem(form.FullName); found {
		problems = append(problems, problem)
	}
	if !isValidEmail(email) {
		problems = append(problems, apperror.InvalidField(emailField, apperror.FieldInvalidEmail))
	}
	candidate := PasswordCandidate{Password: form.Password, FullName: form.FullName, Email: email}
	if problem, found := policy.FindProblem(candidate); found {
		problems = append(problems, problem)
	}
	if whatsAppErr != nil {
		problems = append(problems, apperror.InvalidField(whatsAppField, apperror.FieldInvalidWhatsApp))
	}
	if !form.PrivacyAccepted {
		problems = append(problems, apperror.InvalidField(privacyField, apperror.FieldConsentRequired))
	}
	if len(problems) > 0 {
		return Registration{}, apperror.Validation(problems...)
	}
	return Registration{FullName: form.FullName, Email: email, WhatsApp: whatsApp, Password: form.Password}, nil
}

func (r Registration) Complete(id string, password StoredPassword, now time.Time) User {
	return User{
		ID:                id,
		Email:             r.Email,
		FullName:          r.FullName,
		WhatsApp:          r.WhatsApp,
		Password:          password,
		PrivacyAcceptedAt: now,
	}
}

func findFullNameProblem(fullName string) (apperror.FieldError, bool) {
	length := utf8.RuneCountInString(fullName)
	if length < MinFullNameLength {
		return apperror.InvalidField(fullNameField, apperror.FieldTooShort, lengthParam("min", MinFullNameLength)), true
	}
	if length > MaxFullNameLength {
		return apperror.InvalidField(fullNameField, apperror.FieldTooLong, lengthParam("max", MaxFullNameLength)), true
	}
	return apperror.FieldError{}, false
}

func lengthParam(name string, limit int) apperror.Param {
	return apperror.Param{Name: name, Value: strconv.Itoa(limit)}
}

func isValidEmail(email string) bool {
	if len(email) > MaxEmailLength {
		return false
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	_, host, _ := strings.Cut(email, "@")
	return strings.Contains(host, ".")
}
