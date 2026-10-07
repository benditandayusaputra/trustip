package httpserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const unknownFieldPrefix = "json: unknown field "

func DecodeJSON(c fiber.Ctx, target any) error {
	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || mediaType != fiber.MIMEApplicationJSON {
		return apperror.New(apperror.CodeUnsupportedMediaType)
	}
	body := c.Body()
	if !utf8.Valid(body) {
		return apperror.New(apperror.CodeMalformedRequest)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return decodeFailure(err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return apperror.New(apperror.CodeMalformedRequest)
	}
	return nil
}

func decodeFailure(err error) error {
	if strings.HasPrefix(err.Error(), unknownFieldPrefix) {
		return apperror.New(apperror.CodeUnknownField).WithCause(err)
	}
	return apperror.New(apperror.CodeMalformedRequest).WithCause(fmt.Errorf("decode json body: %w", err))
}
