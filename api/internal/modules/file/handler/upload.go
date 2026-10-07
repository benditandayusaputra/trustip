package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	fileField    = "file"
	purposeField = "purpose"
)

type fileResponse struct {
	ID          string `json:"id"`
	Purpose     string `json:"purpose"`
	Visibility  string `json:"visibility"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	URL         string `json:"url"`
}

func uploadFile(deps Deps) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		command, err := readUpload(c)
		if err != nil {
			return err
		}
		file, err := deps.Upload.Execute(c.Context(), signedIn, command)
		if err != nil {
			logRejectedUpload(c, deps, err)
			return fmt.Errorf("upload image: %w", err)
		}
		c.Location(httpserver.APIPrefix + filesPath + "/" + file.ID)
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Data: toFileResponse(file, deps.PublicBaseURL),
		})
	}
}

func readUpload(c fiber.Ctx) (service.UploadCommand, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return service.UploadCommand{}, apperror.New(apperror.CodeUnsupportedMediaType).WithCause(err)
	}
	if err := rejectUnknownParts(form); err != nil {
		return service.UploadCommand{}, err
	}
	files := form.File[fileField]
	if len(files) != 1 {
		return service.UploadCommand{}, apperror.Validation(apperror.InvalidField(fileField, apperror.FieldRequired))
	}
	data, err := readLimited(files[0])
	if err != nil {
		return service.UploadCommand{}, err
	}
	return service.UploadCommand{Purpose: firstValue(form, purposeField), Data: data}, nil
}

func rejectUnknownParts(form *multipart.Form) error {
	for name := range form.Value {
		if name != purposeField {
			return apperror.New(apperror.CodeUnknownField)
		}
	}
	for name := range form.File {
		if name != fileField {
			return apperror.New(apperror.CodeUnknownField)
		}
	}
	return nil
}

func readLimited(header *multipart.FileHeader) ([]byte, error) {
	opened, err := header.Open()
	if err != nil {
		return nil, apperror.New(apperror.CodeMalformedRequest).WithCause(err)
	}
	data, readErr := io.ReadAll(io.LimitReader(opened, domain.MaxFileBytes+1))
	if err := errors.Join(readErr, opened.Close()); err != nil {
		return nil, apperror.New(apperror.CodeMalformedRequest).WithCause(err)
	}
	return data, nil
}

func firstValue(form *multipart.Form, name string) string {
	if values := form.Value[name]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func logRejectedUpload(c fiber.Ctx, deps Deps, err error) {
	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		return
	}
	deps.SecurityLog.Warn("upload rejected",
		"event", "upload_rejected", "request_id", httpserver.RequestIDFrom(c), "ip", httpserver.ClientIP(c))
}

func toFileResponse(file domain.File, publicBaseURL string) fileResponse {
	url := httpserver.APIPrefix + filesPath + "/" + file.ID
	if file.Visibility == domain.Public {
		url = publicBaseURL + "/" + file.StorageKey
	}
	return fileResponse{
		ID: file.ID, Purpose: string(file.Purpose), Visibility: string(file.Visibility),
		ContentType: file.ContentType, SizeBytes: file.SizeBytes, Width: file.Width, Height: file.Height, URL: url,
	}
}
