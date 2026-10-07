package domain

import (
	"crypto/sha256"
	"errors"
	"path"
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	maxFileMegabytes = 1
	MaxFileBytes     = maxFileMegabytes << 20
)

var (
	ErrUnsupportedImage = errors.New("upload is not a supported image")
	ErrImageTooLarge    = errors.New("image dimensions exceed the limit")
	ErrFileMissing      = errors.New("stored file is missing")
)

type Purpose string

const (
	Logo       Purpose = "LOGO"
	StorePhoto Purpose = "STORE_PHOTO"
	PaymentQR  Purpose = "PAYMENT_QR"
)

type Visibility string

const (
	Public  Visibility = "PUBLIC"
	Private Visibility = "PRIVATE"
)

type purposeRule struct {
	visibility Visibility
	allowedFor []actor.OrganizationType
}

func purposeRules() map[Purpose]purposeRule {
	return map[Purpose]purposeRule{
		Logo:       {visibility: Public, allowedFor: []actor.OrganizationType{actor.Store, actor.Consignor}},
		StorePhoto: {visibility: Public, allowedFor: []actor.OrganizationType{actor.Store}},
		PaymentQR:  {visibility: Private, allowedFor: []actor.OrganizationType{actor.Consignor}},
	}
}

func ParsePurpose(raw string, organizationType actor.OrganizationType) (Purpose, error) {
	purpose := Purpose(raw)
	rule, known := purposeRules()[purpose]
	for _, allowed := range rule.allowedFor {
		if known && allowed == organizationType {
			return purpose, nil
		}
	}
	return "", apperror.Validation(apperror.InvalidField("purpose", apperror.FieldUnknownChoice))
}

func (p Purpose) Visibility() Visibility {
	if rule, known := purposeRules()[p]; known {
		return rule.visibility
	}
	return Private
}

func CheckUploadSize(size int) error {
	if size < 1 {
		return apperror.Validation(apperror.InvalidField("file", apperror.FieldRequired))
	}
	if size > MaxFileBytes {
		return apperror.Validation(apperror.InvalidField("file", apperror.FieldFileTooLarge,
			apperror.Param{Name: "max", Value: strconv.Itoa(maxFileMegabytes)}))
	}
	return nil
}

type Image struct {
	Data        []byte
	ContentType string
	Extension   string
	Width       int
	Height      int
}

type Upload struct {
	ID             string
	OrganizationID string
	CreatedBy      string
	Purpose        Purpose
}

type Stored struct {
	Image Image
	Key   string
}

type File struct {
	ID             string
	OrganizationID string
	Purpose        Purpose
	Visibility     Visibility
	ContentType    string
	SizeBytes      int
	StorageKey     string
	SHA256         []byte
	Width          int
	Height         int
	CreatedBy      string
	CreatedAt      time.Time
}

func NewFile(upload Upload, stored Stored, now time.Time) File {
	digest := sha256.Sum256(stored.Image.Data)
	return File{
		ID:             upload.ID,
		OrganizationID: upload.OrganizationID,
		Purpose:        upload.Purpose,
		Visibility:     upload.Purpose.Visibility(),
		ContentType:    stored.Image.ContentType,
		SizeBytes:      len(stored.Image.Data),
		StorageKey:     stored.Key,
		SHA256:         digest[:],
		Width:          stored.Image.Width,
		Height:         stored.Image.Height,
		CreatedBy:      upload.CreatedBy,
		CreatedAt:      now,
	}
}

func ContentTypeOfKey(key string) string {
	contentTypes := map[string]string{".jpg": "image/jpeg", ".png": "image/png"}
	return contentTypes[path.Ext(key)]
}
