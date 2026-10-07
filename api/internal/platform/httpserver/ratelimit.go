package httpserver

import (
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const (
	HeaderRateLimitLimit     = "RateLimit-Limit"
	HeaderRateLimitRemaining = "RateLimit-Remaining"
	HeaderRateLimitReset     = "RateLimit-Reset"
)

type RateLimiter interface {
	Allow(key string) ratelimit.Decision
}

type RateLimitKey func(c fiber.Ctx) string

func RateLimit(limiter RateLimiter, keyOf RateLimitKey, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		decision := limiter.Allow(keyOf(c))
		c.Set(HeaderRateLimitLimit, strconv.Itoa(decision.Limit))
		c.Set(HeaderRateLimitRemaining, strconv.Itoa(decision.Remaining))
		c.Set(HeaderRateLimitReset, strconv.Itoa(decision.ResetSeconds))
		if !decision.Allowed {
			securityLog.Warn("rate limited",
				"event", "rate_limited", "request_id", RequestIDFrom(c), "ip", ClientIP(c), "route", c.Path())
			return apperror.RateLimited(decision.RetryAfterSeconds)
		}
		return c.Next()
	}
}

func ByClientIP(c fiber.Ctx) string {
	return "ip:" + ClientIP(c)
}

func BySignedInUser(c fiber.Ctx) string {
	if signedIn, ok := ActorFrom(c); ok {
		return "user:" + signedIn.UserID
	}
	return ByClientIP(c)
}
