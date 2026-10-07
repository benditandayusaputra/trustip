package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

const ownOrganizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"

type fakeProfiles struct {
	failure error
	updated []domain.ProfileUpdate
	at      []time.Time
}

func (f *fakeProfiles) GetSettings(_ context.Context, organizationID string) (domain.Settings, error) {
	return domain.Settings{ID: organizationID, Type: actor.Store, Slug: "warung-bu-tini"}, nil
}

func (f *fakeProfiles) UpdateProfile(_ context.Context, change domain.ProfileChange) error {
	if f.failure != nil {
		return f.failure
	}
	f.updated = append(f.updated, change.Update)
	f.at = append(f.at, change.At)
	return nil
}

func storeMember(role actor.Role) actor.Actor {
	return actor.Actor{UserID: founderID, Membership: actor.Membership{
		OrganizationID: ownOrganizationID, OrganizationType: actor.Store, Role: role,
	}}
}

func profileChange(organizationID string) service.UpdateProfileCommand {
	return service.UpdateProfileCommand{OrganizationID: organizationID, Form: domain.ProfileUpdateForm{
		ProfileForm: domain.ProfileForm{Name: "Warung Bu Tini", WhatsApp: "081234567890", CityCode: "33.74"},
	}}
}

func newUpdateProfile(profiles *fakeProfiles) *service.UpdateProfile {
	return service.NewUpdateProfile(profiles, &countingTx{}, fixedClock{})
}

func TestUpdateProfile_StaffIsForbidden(t *testing.T) {
	profiles := &fakeProfiles{}

	staff := storeMember(actor.Staff)

	_, err := newUpdateProfile(profiles).Execute(t.Context(), staff, profileChange(ownOrganizationID))

	if apperror.From(err).Code() != apperror.CodeForbidden || len(profiles.updated) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateProfile_BRAUTHZ02_AnotherOrganizationIsNotFound(t *testing.T) {
	profiles := &fakeProfiles{}
	other := "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a99"

	_, err := newUpdateProfile(profiles).Execute(t.Context(), storeMember(actor.Owner), profileChange(other))

	if apperror.From(err).Code() != apperror.CodeNotFound || len(profiles.updated) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateProfile_UnknownCityOrLogoIsAFieldError(t *testing.T) {
	failures := map[error]string{domain.ErrUnknownCity: "city_code", domain.ErrUnknownLogo: "logo_file_id"}
	for failure, field := range failures {
		profiles := &fakeProfiles{failure: failure}
		owner := storeMember(actor.Owner)

		_, err := newUpdateProfile(profiles).Execute(t.Context(), owner, profileChange(ownOrganizationID))

		fields := apperror.From(err).Fields()
		if len(fields) != 1 || fields[0].Field != field {
			t.Errorf("%v: fields = %+v", failure, fields)
		}
	}
}

func TestUpdateProfile_FRORG01_OwnerSavesAndGetsTheFreshSettings(t *testing.T) {
	profiles := &fakeProfiles{}

	owner := storeMember(actor.Owner)

	settings, err := newUpdateProfile(profiles).Execute(t.Context(), owner, profileChange(ownOrganizationID))

	if err != nil || len(profiles.updated) != 1 || settings.ID != ownOrganizationID ||
		profiles.updated[0].WhatsApp.String() != "6281234567890" || !profiles.at[0].Equal(fixedClock{}.Now()) {
		t.Fatalf("settings = %+v, err = %v", settings, err)
	}
}

func TestGetSettings_BRAUTHZ02_AnotherOrganizationIsNotFound(t *testing.T) {
	_, err := service.NewGetSettings(&fakeProfiles{}).Execute(t.Context(), storeMember(actor.Staff), founderID)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestGetSettings_StaffReadsTheirOwnOrganization(t *testing.T) {
	settings, err := service.NewGetSettings(&fakeProfiles{}).
		Execute(t.Context(), storeMember(actor.Staff), ownOrganizationID)

	if err != nil || settings.Slug != "warung-bu-tini" {
		t.Fatalf("settings = %+v, err = %v", settings, err)
	}
}
