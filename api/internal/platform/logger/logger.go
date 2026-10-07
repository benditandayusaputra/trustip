package logger

import (
	"io"
	"log/slog"
	"regexp"
	"strings"
)

type Stream string

const (
	Access   Stream = "access"
	App      Stream = "app"
	Security Stream = "security"
	Client   Stream = "client"
)

const redacted = "[REDACTED]"

var (
	connectionString = regexp.MustCompile(`(?i)postgres(ql)?://\S+`)
	emailAddress     = regexp.MustCompile(`[A-Za-z0-9._%+-]+@([A-Za-z0-9.-]+\.[A-Za-z]{2,})`)
)

func New(out io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redact,
	}))
}

func ForStream(base *slog.Logger, stream Stream) *slog.Logger {
	return base.With(slog.String("log_type", string(stream)))
}

func redact(_ []string, attr slog.Attr) slog.Attr {
	if isSensitiveKey(attr.Key) {
		return slog.String(attr.Key, redacted)
	}
	value := attr.Value.Resolve()
	if value.Kind() == slog.KindString {
		return slog.String(attr.Key, scrub(value.String()))
	}
	if err, ok := value.Any().(error); ok {
		return slog.String(attr.Key, scrub(err.Error()))
	}
	return attr
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	for _, marker := range sensitiveMarkers() {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func sensitiveMarkers() []string {
	return []string{
		"password", "token", "cookie", "authorization", "signature", "secret", "pepper",
		"email", "phone", "whatsapp", "address", "database_url", "dsn", "body",
	}
}

func scrub(text string) string {
	withoutConnections := connectionString.ReplaceAllString(text, redacted)
	return emailAddress.ReplaceAllString(withoutConnections, "***@$1")
}
