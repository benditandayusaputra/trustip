package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

const SignupCaptchaAction = "signup"

type RegisterDeps struct {
	Captcha CaptchaVerifier
	Hasher  PasswordHasher
	Users   UserCreator
	Policy  domain.PasswordPolicy
	Clock   Clock
	IDs     IDGenerator
}

type RegisterCommand struct {
	Form         domain.RegistrationForm
	CaptchaToken string
	ClientIP     string
}

type RegisterUser struct {
	deps RegisterDeps
}

func NewRegisterUser(deps RegisterDeps) *RegisterUser {
	return &RegisterUser{deps: deps}
}

func (uc *RegisterUser) Execute(ctx context.Context, cmd RegisterCommand) (domain.User, error) {
	challenge := CaptchaChallenge{Token: cmd.CaptchaToken, Action: SignupCaptchaAction, RemoteIP: cmd.ClientIP}
	if err := uc.deps.Captcha.Verify(ctx, challenge); err != nil {
		return domain.User{}, fmt.Errorf("verify signup captcha: %w", err)
	}
	registration, err := domain.NewRegistration(cmd.Form, uc.deps.Policy)
	if err != nil {
		return domain.User{}, fmt.Errorf("validate registration: %w", err)
	}
	user, err := uc.completeRegistration(registration)
	if err != nil {
		return domain.User{}, err
	}
	if err := uc.deps.Users.CreateUser(ctx, user); err != nil {
		return domain.User{}, explainCreateFailure(err)
	}
	return user, nil
}

func (uc *RegisterUser) completeRegistration(registration domain.Registration) (domain.User, error) {
	password, err := uc.deps.Hasher.Hash(registration.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	id, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.User{}, fmt.Errorf("generate user id: %w", err)
	}
	return registration.Complete(id, password, uc.deps.Clock.Now()), nil
}

func explainCreateFailure(err error) error {
	if errors.Is(err, domain.ErrEmailTaken) {
		return apperror.Validation(apperror.InvalidField("email", apperror.FieldEmailTaken))
	}
	return fmt.Errorf("create user: %w", err)
}
