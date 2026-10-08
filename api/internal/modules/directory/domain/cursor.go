package domain

import (
	"encoding/base64"
	"encoding/json"
	"slices"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const cursorField = "cursor"

type Position struct {
	Text     string  `json:"t"`
	Number   float64 `json:"n"`
	ID       string  `json:"i"`
	RadiusKm int     `json:"r"`
}

func (p Position) IsStart() bool {
	return p.ID == ""
}

func CursorOf(position Position) string {
	encoded, err := json.Marshal(position)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func ParseCursor(raw string) (Position, error) {
	if raw == "" {
		return Position{}, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Position{}, invalidCursor()
	}
	var position Position
	if err := json.Unmarshal(decoded, &position); err != nil {
		return Position{}, invalidCursor()
	}
	isKnownRadius := position.RadiusKm == EveryDistance || slices.Contains(RadiusSteps(), position.RadiusKm)
	if !idFormat.MatchString(position.ID) || !isKnownRadius {
		return Position{}, invalidCursor()
	}
	return position, nil
}

func invalidCursor() error {
	return apperror.Validation(apperror.InvalidField(cursorField, apperror.FieldUnknownChoice))
}
