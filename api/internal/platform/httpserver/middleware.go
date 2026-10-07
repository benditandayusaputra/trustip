package httpserver

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	jsonBodyLimit  = 64 * 1024
	requestTimeout = 10 * time.Second
)

func AccessLog(accessLog *slog.Logger, clock Clock, handleError fiber.ErrorHandler) fiber.Handler {
	return func(c fiber.Ctx) error {
		started := clock.Now()
		if err := c.Next(); err != nil {
			if writeErr := handleError(c, err); writeErr != nil {
				return writeErr
			}
		}
		accessLog.Info("request",
			"request_id", RequestIDFrom(c),
			"method", c.Method(),
			"route", routePattern(c),
			"status", c.Response().StatusCode(),
			"duration_ms", clock.Now().Sub(started).Milliseconds(),
			"bytes", len(c.Response().Body()),
			"ip", ClientIP(c),
		)
		return nil
	}
}

func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderStrictTransportSecurity, "max-age=63072000; includeSubDomains")
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderReferrerPolicy, "strict-origin-when-cross-origin")
		c.Set(fiber.HeaderPermissionsPolicy, "geolocation=(self), camera=(), microphone=(), payment=()")
		c.Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set(fiber.HeaderCrossOriginResourcePolicy, "same-origin")
		return c.Next()
	}
}

func LimitJSONBody() fiber.Handler {
	return func(c fiber.Ctx) error {
		isUpload := strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEMultipartForm)
		if !isUpload && len(c.Body()) > jsonBodyLimit {
			return apperror.New(apperror.CodePayloadTooLarge)
		}
		return c.Next()
	}
}

func Deadline(timeout time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c, timeout)
		defer cancel()
		c.SetContext(ctx)
		return c.Next()
	}
}

type clientIPContextKey struct{}

func CarryClientIP() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.SetContext(context.WithValue(c.Context(), clientIPContextKey{}, ClientIP(c)))
		return c.Next()
	}
}

func ClientIPFrom(ctx context.Context) string {
	clientIP, _ := ctx.Value(clientIPContextKey{}).(string)
	return clientIP
}
