package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

type fakeStoreProfiles struct {
	failure    error
	changes    []domain.StoreProfileChange
	consignors []domain.ConsignorProfileChange
	payments   []domain.PaymentInfoChange
}

func (f *fakeStoreProfiles) UpdateStoreProfile(_ context.Context, change domain.StoreProfileChange) error {
	if f.failure != nil {
		return f.failure
	}
	f.changes = append(f.changes, change)
	return nil
}

func (f *fakeStoreProfiles) UpdateConsignorProfile(_ context.Context, change domain.ConsignorProfileChange) error {
	if f.failure != nil {
		return f.failure
	}
	f.consignors = append(f.consignors, change)
	return nil
}

func (f *fakeStoreProfiles) UpdatePaymentInfo(_ context.Context, change domain.PaymentInfoChange) error {
	if f.failure != nil {
		return f.failure
	}
	f.payments = append(f.payments, change)
	return nil
}

func storeProfileCommand(organizationID string) service.UpdateStoreProfileCommand {
	return service.UpdateStoreProfileCommand{OrganizationID: organizationID, Form: domain.StoreProfileForm{
		StoreType: "CAFE", Acceptance: domain.Acceptance{AcceptsFixedDeposit: true},
	}}
}

func newUpdateStoreProfile(stores *fakeStoreProfiles) *service.UpdateSection[domain.StoreProfileForm] {
	return service.NewUpdateStoreProfile(sectionDeps(stores))
}

func sectionDeps(sections *fakeStoreProfiles) service.SectionDeps {
	return service.SectionDeps{Profiles: &fakeProfiles{}, Sections: sections, Tx: &countingTx{}, Clock: fixedClock{}}
}

func TestUpdateStoreProfile_StaffIsForbidden(t *testing.T) {
	stores := &fakeStoreProfiles{}

	_, err := newUpdateStoreProfile(stores).Execute(t.Context(), storeMember(actor.Staff),
		storeProfileCommand(ownOrganizationID))

	if apperror.From(err).Code() != apperror.CodeForbidden || len(stores.changes) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateStoreProfile_ConsignorHasNoStoreProfile(t *testing.T) {
	consignor := storeMember(actor.Owner)
	consignor.Membership.OrganizationType = actor.Consignor

	_, err := newUpdateStoreProfile(&fakeStoreProfiles{}).Execute(t.Context(), consignor,
		storeProfileCommand(ownOrganizationID))

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateStoreProfile_BRAUTHZ02_AnotherOrganizationIsNotFound(t *testing.T) {
	_, err := newUpdateStoreProfile(&fakeStoreProfiles{}).Execute(t.Context(), storeMember(actor.Owner),
		storeProfileCommand(founderID))

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateStoreProfile_UnknownCategoryOrPhotoIsAFieldError(t *testing.T) {
	failures := map[error]string{
		domain.ErrUnknownCategory: "accepted_category_ids",
		domain.ErrUnknownPhoto:    "photo_file_ids",
	}
	for failure, field := range failures {
		_, err := newUpdateStoreProfile(&fakeStoreProfiles{failure: failure}).Execute(t.Context(),
			storeMember(actor.Owner), storeProfileCommand(ownOrganizationID))

		fields := apperror.From(err).Fields()
		if len(fields) != 1 || fields[0].Field != field {
			t.Errorf("%v: fields = %+v", failure, fields)
		}
	}
}

func TestUpdateStoreProfile_FRORG04_OwnerSavesTheStoreProfile(t *testing.T) {
	stores := &fakeStoreProfiles{}

	settings, err := newUpdateStoreProfile(stores).Execute(t.Context(), storeMember(actor.Owner),
		storeProfileCommand(ownOrganizationID))

	if err != nil || len(stores.changes) != 1 || stores.changes[0].Profile.StoreType != "CAFE" ||
		settings.ID != ownOrganizationID {
		t.Fatalf("settings = %+v, err = %v", settings, err)
	}
}

type fakeCategories struct{}

func (fakeCategories) ListCategories(context.Context) ([]domain.Category, error) {
	return []domain.Category{{ID: "c1", Name: "Minuman", Slug: "minuman"}}, nil
}

func TestListCategories_ReturnsTheStoredList(t *testing.T) {
	categories, err := service.NewListCategories(fakeCategories{}).Execute(t.Context())

	if err != nil || len(categories) != 1 {
		t.Fatalf("categories = %+v, err = %v", categories, err)
	}
}
