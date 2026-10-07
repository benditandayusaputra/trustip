package main

import (
	"fmt"
	"log/slog"
	"time"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	orgrepository "github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/captcha"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/revocation"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const (
	signupsPerHourPerIP = 3
	loginsPerMinute     = 5
	loginsPerHour       = 20
	refreshesPerMinute  = 30
)

func passwordHashing() security.Argon2Params {
	return security.MinimumArgon2Params()
}

type authComponents struct {
	captcha      service.CaptchaVerifier
	passwords    service.PasswordVerifier
	users        service.UserCreator
	accounts     service.AccountReader
	credentials  service.CredentialStore
	sessions     service.SessionRotator
	userSessions service.UserSessionRevoker
	tokens       service.TokenIssuer
	memberships  service.MembershipFinder
	audit        service.AuditTrail
	revoked      *revocation.List
	accessKeys   security.SecretRing
	tx           service.Transactor
	clock        service.Clock
	ids          service.IDGenerator
	cookies      authhandler.CookiePolicy
}

func buildAuthComponents(cfg config.Config, pool *database.Pool, clock service.Clock) (authComponents, error) {
	hasher, err := security.NewPasswordHasher(cfg.PasswordPeppers, passwordHashing())
	if err != nil {
		return authComponents{}, fmt.Errorf("build password hasher: %w", err)
	}
	users := repository.NewUsers(pool)
	sessions := repository.NewSessions(pool)
	return authComponents{
		captcha:      captchaVerifier{turnstile: newTurnstile(cfg.Turnstile)},
		passwords:    passwordHasher{hasher: hasher},
		users:        users,
		accounts:     users,
		credentials:  users,
		sessions:     sessions,
		userSessions: sessions,
		tokens:       newTokenIssuer(cfg.JWTKeys),
		memberships:  orgrepository.NewMemberships(pool),
		audit:        newAuditTrail(auditrepository.NewRecords(pool)),
		accessKeys:   cfg.JWTKeys,
		revoked:      revocation.New(clock),
		tx:           pool,
		clock:        clock,
		ids:          idgen.UUIDv7{},
		cookies:      cookiePolicy(cfg.Environment),
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
		Authenticate: httpserver.Authenticate(newAccessVerifier(components), components.cookies.AccessName()),
		Account:      service.NewGetAccount(components.accounts),
		Register:     newRegisterUser(components),
		Login:        newLoginUser(components),
		Refresh:      newRefreshSession(components),
		LogOut: authhandler.SessionEnders{
			ThisDevice: service.NewLogOut(logoutDeps(components)),
			AllDevices: service.NewLogOutEverywhere(logoutDeps(components)),
		},
		SignupLimiter: ratelimit.New(ratelimit.Rule{Limit: signupsPerHourPerIP, Window: time.Hour}, limiterClock),
		LoginLimiters: authhandler.LoginLimiters{
			PerIPMinute:    ratelimit.New(perMinute, limiterClock),
			PerIPHour:      ratelimit.New(perHour, limiterClock),
			PerEmailMinute: ratelimit.New(perMinute, limiterClock),
			PerEmailHour:   ratelimit.New(perHour, limiterClock),
		},
		RefreshLimiter:  ratelimit.New(ratelimit.Rule{Limit: refreshesPerMinute, Window: time.Minute}, limiterClock),
		RefreshLimitKey: byRefreshFamily(components.cookies),
		Cookies:         components.cookies,
		SecurityLog:     securityLog,
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
		Memberships: components.memberships,
		Audit:       components.audit,
		Tx:          components.tx,
		Clock:       components.clock,
		IDs:         components.ids,
	})
}
