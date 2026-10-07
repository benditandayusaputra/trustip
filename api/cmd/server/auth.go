package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/captcha"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const signupsPerHourPerIP = 3

func passwordHashing() security.Argon2Params {
	return security.MinimumArgon2Params()
}

type authComponents struct {
	captcha service.CaptchaVerifier
	hasher  service.PasswordHasher
	users   service.UserCreator
	clock   service.Clock
	ids     service.IDGenerator
}

func buildAuthComponents(cfg config.Config, pool *database.Pool, clock service.Clock) (authComponents, error) {
	hasher, err := security.NewPasswordHasher(cfg.PasswordPeppers, passwordHashing())
	if err != nil {
		return authComponents{}, fmt.Errorf("build password hasher: %w", err)
	}
	turnstile := captcha.NewTurnstile(captcha.Options{
		Secret:           cfg.Turnstile.SecretKey,
		ExpectedHostname: cfg.Turnstile.ExpectedHostname,
		Endpoint:         captcha.SiteVerifyURL,
		Timeout:          captcha.DefaultTimeout,
	})
	return authComponents{
		captcha: captchaVerifier{turnstile: turnstile},
		hasher:  passwordHasher{hasher: hasher},
		users:   repository.NewUsers(pool),
		clock:   clock,
		ids:     idgen.UUIDv7{},
	}, nil
}

func authDeps(components authComponents, securityLog *slog.Logger, limiterClock ratelimit.Clock) authhandler.Deps {
	return authhandler.Deps{
		Register: service.NewRegisterUser(service.RegisterDeps{
			Captcha: components.captcha,
			Hasher:  components.hasher,
			Users:   components.users,
			Policy:  domain.NewPasswordPolicy(),
			Clock:   components.clock,
			IDs:     components.ids,
		}),
		SignupLimiter: ratelimit.New(ratelimit.Rule{Limit: signupsPerHourPerIP, Window: time.Hour}, limiterClock),
		SecurityLog:   securityLog,
	}
}

type captchaVerifier struct {
	turnstile *captcha.Turnstile
}

func (v captchaVerifier) Verify(ctx context.Context, challenge service.CaptchaChallenge) error {
	if err := v.turnstile.Verify(ctx, captcha.Challenge(challenge)); err != nil {
		return fmt.Errorf("verify captcha: %w", err)
	}
	return nil
}

type passwordHasher struct {
	hasher *security.PasswordHasher
}

func (h passwordHasher) Hash(password string) (domain.StoredPassword, error) {
	stored, err := h.hasher.Hash(password)
	if err != nil {
		return domain.StoredPassword{}, fmt.Errorf("hash password: %w", err)
	}
	return domain.StoredPassword(stored), nil
}
