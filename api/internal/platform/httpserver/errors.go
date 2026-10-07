package httpserver

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const defaultRetryAfterSeconds = 5

type fieldErrorBody struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Errors  []fieldErrorBody `json:"errors,omitempty"`
	Meta    meta             `json:"meta"`
}

func NewErrorHandler(appLog, securityLog *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		appErr := toAppError(err).WithParam("request_id", RequestIDFrom(c))
		if appErr.Kind() == apperror.KindInternal {
			logUnexpected(appLog, c, appErr)
		}
		if event, ok := securityEvents()[appErr.Code()]; ok {
			securityLog.Warn(event, "event", event, "request_id", RequestIDFrom(c), "ip", ClientIP(c), "reason", appErr)
		}
		return writeError(c, appErr)
	}
}

func securityEvents() map[apperror.Code]string {
	return map[apperror.Code]string{
		apperror.CodeCaptchaFailed:  "captcha_failed",
		apperror.CodeLoginFailed:    "login_failed",
		apperror.CodeAccountLocked:  "account_locked",
		apperror.CodeSessionRevoked: "refresh_token_reused",
	}
}

func toAppError(err error) *apperror.Error {
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		return fromFiberStatus(fiberErr.Code)
	}
	return apperror.From(err)
}

func fromFiberStatus(status int) *apperror.Error {
	switch status {
	case fiber.StatusNotFound, fiber.StatusMethodNotAllowed:
		return apperror.NotFound()
	case fiber.StatusRequestEntityTooLarge:
		return apperror.New(apperror.CodePayloadTooLarge)
	case fiber.StatusUnsupportedMediaType:
		return apperror.New(apperror.CodeUnsupportedMediaType)
	case fiber.StatusBadRequest:
		return apperror.New(apperror.CodeMalformedRequest)
	case fiber.StatusRequestTimeout, fiber.StatusServiceUnavailable:
		return apperror.Unavailable(defaultRetryAfterSeconds)
	}
	return apperror.Internal(fiber.NewError(status))
}

func logUnexpected(appLog *slog.Logger, c fiber.Ctx, appErr *apperror.Error) {
	attrs := []any{"request_id", RequestIDFrom(c), "route", routePattern(c), "error", appErr}
	if recovered, ok := errors.AsType[*panicError](appErr); ok {
		attrs = append(attrs, "stack", string(recovered.stack))
	}
	appLog.Error("unexpected error", attrs...)
}

func writeError(c fiber.Ctx, appErr *apperror.Error) error {
	retryAfter := appErr.RetryAfter()
	if appErr.Kind() == apperror.KindUnavailable && retryAfter == 0 {
		retryAfter = defaultRetryAfterSeconds
	}
	if retryAfter > 0 {
		c.Set(fiber.HeaderRetryAfter, strconv.Itoa(retryAfter))
	}
	return writeJSON(c, appErr.Status(), errorEnvelope{
		Code:    string(appErr.Code()),
		Message: appErr.Message(),
		Errors:  fieldErrors(appErr),
		Meta:    meta{RequestID: RequestIDFrom(c)},
	})
}

func fieldErrors(appErr *apperror.Error) []fieldErrorBody {
	if appErr.Code() != apperror.CodeValidationFailed {
		return nil
	}
	bodies := make([]fieldErrorBody, 0, len(appErr.Fields()))
	for _, field := range appErr.Fields() {
		bodies = append(bodies, fieldErrorBody{Field: field.Field, Code: field.Code, Message: field.Message})
	}
	return bodies
}

func routePattern(c fiber.Ctx) string {
	if route := c.Route(); route != nil && route.Path != "" {
		return route.Path
	}
	return "unmatched"
}
