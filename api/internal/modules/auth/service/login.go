package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

const LoginCaptchaAction = "login"

type LoginDeps struct {
	Captcha     CaptchaVerifier
	Passwords   PasswordVerifier
	Credentials CredentialStore
	Sessions    SessionCreator
	Tokens      TokenIssuer
	Memberships MembershipFinder
	Tx          Transactor
	Clock       Clock
	IDs         IDGenerator
}

type LoginCommand struct {
	Email        string
	Password     string
	CaptchaToken string
	ClientIP     string
}

type LoginResult struct {
	Account domain.Account
	Tokens  domain.SessionTokens
}

type LoginUser struct {
	deps LoginDeps
}

func NewLoginUser(deps LoginDeps) *LoginUser {
	return &LoginUser{deps: deps}
}

func (uc *LoginUser) Execute(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
	challenge := CaptchaChallenge{Token: cmd.CaptchaToken, Action: LoginCaptchaAction, RemoteIP: cmd.ClientIP}
	if err := uc.deps.Captcha.Verify(ctx, challenge); err != nil {
		return LoginResult{}, fmt.Errorf("verify login captcha: %w", err)
	}
	credentials, found, err := uc.deps.Credentials.FindCredentials(ctx, cmd.Email)
	if err != nil {
		return LoginResult{}, fmt.Errorf("find credentials: %w", err)
	}
	now := uc.deps.Clock.Now()
	if !found || credentials.Attempts.IsLockedAt(now) {
		return LoginResult{}, uc.refuseWithoutChecking(credentials, cmd.Password, now)
	}
	check, err := uc.deps.Passwords.Verify(credentials.Password, cmd.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify password for %s: %w", credentials.UserID, err)
	}
	if !check.Matches {
		return LoginResult{}, uc.recordFailure(ctx, credentials.UserID, now)
	}
	login := successfulLogin{credentials: credentials, check: check, password: cmd.Password, now: now}
	return uc.startSession(ctx, login)
}

func (uc *LoginUser) refuseWithoutChecking(credentials domain.Credentials, password string, now time.Time) error {
	if err := uc.deps.Passwords.VerifyAbsent(password); err != nil {
		return fmt.Errorf("spend comparison hash: %w", err)
	}
	if credentials.Attempts.IsLockedAt(now) {
		return apperror.AccountLocked(credentials.Attempts.SecondsUntilUnlock(now))
	}
	return apperror.New(apperror.CodeLoginFailed)
}

func (uc *LoginUser) recordFailure(ctx context.Context, userID string, now time.Time) error {
	var attempts domain.LoginAttempts
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		current, err := uc.deps.Credentials.LockLoginAttempts(ctx, userID)
		if err != nil {
			return fmt.Errorf("lock login attempts: %w", err)
		}
		attempts = current.AfterFailure(now)
		return uc.deps.Credentials.SaveLoginAttempts(ctx, userID, attempts)
	})
	if err != nil {
		return fmt.Errorf("record failed login for %s: %w", userID, err)
	}
	if attempts.IsLockedAt(now) {
		return apperror.AccountLocked(attempts.SecondsUntilUnlock(now))
	}
	return apperror.New(apperror.CodeLoginFailed)
}
