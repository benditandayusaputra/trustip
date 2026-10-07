package repository_test

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID      = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	ownerOrgID  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	strangerID  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	fileID      = "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a7e"
	storedHash  = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
	storageKey  = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1.jpg"
	semarangKey = "33.74"
)

func createdAt() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func poolWithOrganizations(t *testing.T) *database.Pool {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	err := queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "sari@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Sari Wulandari", Whatsapp: "6281234567890", PrivacyAcceptedAt: createdAt(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	for id, slug := range map[string]string{ownerOrgID: "warung-bu-tini", strangerID: "toko-lain"} {
		_, err := queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
			ID: id, Type: "STORE", Name: slug, Slug: slug, Whatsapp: "6281234567890",
			CityCode: semarangKey, CreatedAt: createdAt(),
		})
		if err != nil {
			t.Fatalf("create organization: %v", err)
		}
	}
	return pool
}

func logo() domain.File {
	digest := sha256.Sum256([]byte("jpeg"))
	return domain.File{
		ID: fileID, OrganizationID: ownerOrgID, Purpose: domain.Logo, Visibility: domain.Public,
		ContentType: "image/jpeg", SizeBytes: 4, StorageKey: storageKey, SHA256: digest[:],
		Width: 640, Height: 480, CreatedBy: userID, CreatedAt: createdAt(),
	}
}

func TestCreateFile_StoresEveryFieldForItsOrganization(t *testing.T) {
	files := repository.NewFiles(poolWithOrganizations(t))
	if err := files.CreateFile(t.Context(), logo()); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := files.GetFile(t.Context(), ownerOrgID, fileID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	want := logo()
	if got.StorageKey != want.StorageKey || !bytes.Equal(got.SHA256, want.SHA256) || got.Width != 640 ||
		got.Visibility != domain.Public || !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf("got = %+v", got)
	}
}

func TestGetFile_RSEC38_FileOfAnotherOrganizationIsNotFound(t *testing.T) {
	files := repository.NewFiles(poolWithOrganizations(t))
	if err := files.CreateFile(t.Context(), logo()); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err := files.GetFile(t.Context(), strangerID, fileID)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateFile_NFR09_NonImageContentTypeIsRejectedByTheDatabase(t *testing.T) {
	files := repository.NewFiles(poolWithOrganizations(t))
	notAnImage := logo()
	notAnImage.ContentType = "image/gif"

	if err := files.CreateFile(t.Context(), notAnImage); err == nil {
		t.Fatal("a non-image content type must violate files_content_type_image")
	}
}
