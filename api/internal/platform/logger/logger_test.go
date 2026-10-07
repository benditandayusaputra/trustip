package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

func decodeLine(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatalf("decode %q: %v", buffer.String(), err)
	}
	return entry
}

func TestNew_SensitiveFields_AreRedacted(t *testing.T) {
	sensitive := []string{
		"password", "new_password", "access_token", "refresh_token", "cookie", "Authorization",
		"x-edge-signature", "jwt_secret", "pepper", "email", "phone", "whatsapp_number",
		"address", "database_url", "body",
	}
	for _, key := range sensitive {
		t.Run(key, func(t *testing.T) {
			var buffer bytes.Buffer

			logger.New(&buffer, slog.LevelInfo).Info("event", key, "rahasia-123")

			if strings.Contains(buffer.String(), "rahasia-123") {
				t.Fatalf("%s leaked: %s", key, buffer.String())
			}
		})
	}
}

func TestNew_SensitiveFieldInsideGroup_IsRedacted(t *testing.T) {
	var buffer bytes.Buffer

	logger.New(&buffer, slog.LevelInfo).Info("event", slog.Group("request", slog.String("cookie", "sid=abc")))

	if strings.Contains(buffer.String(), "sid=abc") {
		t.Fatalf("grouped cookie leaked: %s", buffer.String())
	}
}

func TestNew_ConnectionStringInError_IsRedacted(t *testing.T) {
	var buffer bytes.Buffer
	err := errors.New("dial postgres://owner:pw@ep-x.neon.tech/db failed")

	logger.New(&buffer, slog.LevelInfo).Error("database", "error", err)

	if strings.Contains(buffer.String(), "owner:pw") || strings.Contains(buffer.String(), "neon.tech") {
		t.Fatalf("connection string leaked: %s", buffer.String())
	}
}

func TestNew_EmailInsideMessageValue_IsMasked(t *testing.T) {
	var buffer bytes.Buffer

	logger.New(&buffer, slog.LevelInfo).Info("event", "reason", "login failed for sri.warung@gmail.com")

	if strings.Contains(buffer.String(), "sri.warung@gmail.com") {
		t.Fatalf("email leaked: %s", buffer.String())
	}
}

func TestNew_HarmlessFields_AreKept(t *testing.T) {
	var buffer bytes.Buffer

	logger.New(&buffer, slog.LevelInfo).Info("request", "route", "/api/v1/drops/:id", "status", 201)

	entry := decodeLine(t, &buffer)
	if entry["route"] != "/api/v1/drops/:id" || entry["status"] != float64(201) {
		t.Fatalf("entry = %v", entry)
	}
}

func TestNew_BelowLevel_IsDropped(t *testing.T) {
	var buffer bytes.Buffer

	logger.New(&buffer, slog.LevelInfo).Debug("noise")

	if buffer.Len() != 0 {
		t.Fatalf("debug line written at info level: %s", buffer.String())
	}
}

func TestForStream_AddsLogType(t *testing.T) {
	var buffer bytes.Buffer
	base := logger.New(&buffer, slog.LevelInfo)

	logger.ForStream(base, logger.Security).Warn("edge signature invalid")

	if entry := decodeLine(t, &buffer); entry["log_type"] != "security" {
		t.Fatalf("log_type = %v", entry["log_type"])
	}
}
