package pagination

import (
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	DefaultLimit = 20
	MaxLimit     = 50
	separator    = "|"
	cursorField  = "cursor"
)

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Position struct {
	At time.Time
	ID string
}

func (p Position) IsStart() bool {
	return p.ID == ""
}

func ParseLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return DefaultLimit
	}
	return min(limit, MaxLimit)
}

func CursorOf(position Position) string {
	plain := position.At.UTC().Format(time.RFC3339Nano) + separator + position.ID
	return base64.RawURLEncoding.EncodeToString([]byte(plain))
}

func ParseCursor(raw string) (Position, error) {
	if raw == "" {
		return Position{}, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Position{}, invalidCursor()
	}
	at, id, found := strings.Cut(string(decoded), separator)
	parsed, timeErr := time.Parse(time.RFC3339Nano, at)
	if !found || timeErr != nil || !idFormat.MatchString(id) {
		return Position{}, invalidCursor()
	}
	return Position{At: parsed, ID: id}, nil
}

func invalidCursor() error {
	return apperror.Validation(apperror.InvalidField(cursorField, apperror.FieldUnknownChoice))
}
