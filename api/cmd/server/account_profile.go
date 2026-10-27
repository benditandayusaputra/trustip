package main

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const (
	passwordChangesPerMinute = 5
	passwordChangesPerHour   = 20
)

func newUpdateAccount(components authComponents) *service.UpdateAccount {
	return service.NewUpdateAccount(service.UpdateAccountDeps{
		Accounts: components.profiles,
		Audit:    components.audit,
		Tx:       components.tx,
		Clock:    components.clock,
	})
}

func newChangePassword(components authComponents) *service.ChangePassword {
	return service.NewChangePassword(service.ChangePasswordDeps{
		Owners:    components.owners,
		Sessions:  components.others,
		Passwords: components.passwords,
		Policy:    domain.NewPasswordPolicy(),
		Access:    components.revoked,
		Audit:     components.audit,
		Tx:        components.tx,
		Clock:     components.clock,
	})
}

func passwordLimiters(clock ratelimit.Clock) authhandler.PasswordLimiters {
	return authhandler.PasswordLimiters{
		PerUserMinute: ratelimit.New(ratelimit.Rule{Limit: passwordChangesPerMinute, Window: time.Minute}, clock),
		PerUserHour:   ratelimit.New(ratelimit.Rule{Limit: passwordChangesPerHour, Window: time.Hour}, clock),
	}
}
