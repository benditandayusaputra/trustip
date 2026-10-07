package repository_test

import (
	"crypto/sha256"
	"errors"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
)

const logoFileID = "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a90"

func (w world) file(id, purpose string) {
	w.t.Helper()
	digest := sha256.Sum256([]byte(id))
	err := w.queries.CreateFile(w.t.Context(), sqlc.CreateFileParams{
		ID: id, OrganizationID: organizationID, Purpose: purpose, Visibility: "PUBLIC", ContentType: "image/png",
		SizeBytes: 10, StorageKey: strings.ReplaceAll(id, "-", "") + ".png", Sha256: digest[:], Width: 10, Height: 10,
		CreatedBy: userID, CreatedAt: createdAt(),
	})
	if err != nil {
		w.t.Fatalf("create file: %v", err)
	}
}

func foundedConsignor(t *testing.T) world {
	t.Helper()
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	if _, err := repository.NewFounders(given.pool).CreateOrganization(t.Context(),
		founding(t, firstFounder(), "kripik")); err != nil {
		t.Fatalf("found: %v", err)
	}
	return given
}

func renamed(t *testing.T, logo string) domain.ProfileChange {
	t.Helper()
	whatsApp, err := contact.ParseWhatsAppNumber("0811 2233 4455")
	if err != nil {
		t.Fatalf("whatsapp: %v", err)
	}
	return domain.ProfileChange{OrganizationID: organizationID, At: createdAt(), Update: domain.ProfileUpdate{
		Profile: domain.Profile{
			Name: "Kripik Mbok Darmi Asli", WhatsApp: whatsApp, CityCode: "33.22", District: "Ungaran Barat",
		},
		LogoFileID: logo,
	}}
}

func TestUpdateProfile_FRORG01_SavesTheProfileAndItsLogo(t *testing.T) {
	given := foundedConsignor(t)
	given.file(logoFileID, "LOGO")
	profiles := repository.NewProfiles(given.pool)

	if err := profiles.UpdateProfile(t.Context(), renamed(t, logoFileID)); err != nil {
		t.Fatalf("update: %v", err)
	}

	settings, err := profiles.GetSettings(t.Context(), organizationID)
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if settings.Profile.Name != "Kripik Mbok Darmi Asli" || settings.CityName != "Kabupaten Semarang" ||
		settings.Logo.ID != logoFileID || settings.Slug != "kripik" || settings.Profile.HasLocation {
		t.Fatalf("settings = %+v", settings)
	}
}

func TestUpdateProfile_LogoMustBeALogoOfTheSameOrganization(t *testing.T) {
	given := foundedConsignor(t)
	given.file(logoFileID, "PAYMENT_QR")

	err := repository.NewProfiles(given.pool).UpdateProfile(t.Context(), renamed(t, logoFileID))

	if !errors.Is(err, domain.ErrUnknownLogo) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateProfile_UnknownCityChangesNothing(t *testing.T) {
	given := foundedConsignor(t)
	change := renamed(t, "")
	change.Update.CityCode = "99.99"
	profiles := repository.NewProfiles(given.pool)

	err := profiles.UpdateProfile(t.Context(), change)

	settings, readErr := profiles.GetSettings(t.Context(), organizationID)
	if !errors.Is(err, domain.ErrUnknownCity) || readErr != nil || settings.Profile.Name != "Kripik Mbok Darmi" {
		t.Fatalf("err = %v, settings = %+v", err, settings)
	}
}
