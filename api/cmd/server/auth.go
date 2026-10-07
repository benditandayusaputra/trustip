package main

import (
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

const (
	signupsPerHourPerIP = 3
	loginsPerMinute     = 5
	loginsPerHour       = 20
)

func passwordHashing() security.Argon2Params {
	return security.MinimumArgon2Params()
}

type authComponents struct {
	captcha     service.CaptchaVerifier
	passwords   service.PasswordVerifier
	users       service.UserCreator
	credentials service.CredentialStore
	sessions    service.SessionCreator
	tokens      service.TokenIssuer
	tx          service.Transactor
	clock       service.Clock
	ids         service.IDGenerator
	cookies     authhandler.CookiePolicy
}

func buildAuthComponents(cfg config.Config, pool *database.Pool, clock service.Clock) (authComponents, error) {
	hasher, err := security.NewPasswordHasher(cfg.PasswordPeppers, passwordHashing())
	if err != nil {
		return authComponents{}, fmt.Errorf("build password hasher: %w", err)
	}
	users := repository.NewUsers(pool)
	return authComponents{
		captcha:     captchaVerifier{turnstile: newTurnstile(cfg.Turnstile)},
		passwords:   passwordHasher{hasher: hasher},
		users:       users,
		credentials: users,
		sessions:    repository.NewSessions(pool),
		tokens:      tokenIssuer{signer: security.NewTokenSigner(cfg.JWTKeys)},
		tx:          pool,
		clock:       clock,
		ids:         idgen.UUIDv7{},
		cookies:     cookiePolicy(cfg.Environment),
	}, nil
}

func newTurnstile(cfg config.Turnstile) *captcha.Turnstile {
	return captcha.NewTurnstile(captcha.Options{
		Secret:           cfg.SecretKey,
		ExpectedHostname: cfg.ExpectedHostname,
		Endpoint:         captcha.SiteVerifyURL,
		Timeout:          captcha.DefaultTimeout,
	})
}

func cookiePolicy(environment config.Environment) authhandler.CookiePolicy {
	if environment == config.Production {
		return authhandler.ProductionCookies()
	}
	return authhandler.LocalCookies()
}

func authDeps(components authComponents, securityLog *slog.Logger, limiterClock ratelimit.Clock) authhandler.Deps {
	perMinute := ratelimit.Rule{Limit: loginsPerMinute, Window: time.Minute}
	perHour := ratelimit.Rule{Limit: loginsPerHour, Window: time.Hour}
	return authhandler.Deps{
		Register:      newRegisterUser(components),
		Login:         newLoginUser(components),
		SignupLimiter: ratelimit.New(ratelimit.Rule{Limit: signupsPerHourPerIP, Window: time.Hour}, limiterClock),
		LoginLimiters: authhandler.LoginLimiters{
			PerIPMinute:    ratelimit.New(perMinute, limiterClock),
			PerIPHour:      ratelimit.New(perHour, limiterClock),
			PerEmailMinute: ratelimit.New(perMinute, limiterClock),
			PerEmailHour:   ratelimit.New(perHour, limiterClock),
		},
		Cookies:     components.cookies,
		SecurityLog: securityLog,
	}
}

func newRegisterUser(components authComponents) *service.RegisterUser {
	return service.NewRegisterUser(service.RegisterDeps{
		Captcha: components.captcha,
		Hasher:  components.passwords,
		Users:   components.users,
		Policy:  domain.NewPasswordPolicy(),
		Clock:   components.clock,
		IDs:     components.ids,
	})
}

func newLoginUser(components authComponents) *service.LoginUser {
	return service.NewLoginUser(service.LoginDeps{
		Captcha:     components.captcha,
		Passwords:   components.passwords,
		Credentials: components.credentials,
		Sessions:    components.sessions,
		Tokens:      components.tokens,
		Tx:          components.tx,
		Clock:       components.clock,
		IDs:         components.ids,
	})
}
