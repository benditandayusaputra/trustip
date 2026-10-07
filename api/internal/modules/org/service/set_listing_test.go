package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

type fakeListings struct {
	facts   domain.ListingFacts
	changes []domain.ListingChange
}

func (f *fakeListings) GetListingFacts(context.Context, string) (domain.ListingFacts, error) {
	return f.facts, nil
}

func (f *fakeListings) SetListed(_ context.Context, change domain.ListingChange) error {
	f.changes = append(f.changes, change)
	return nil
}

func newSetListing(listings *fakeListings) *service.SetListing {
	return service.NewSetListing(service.ListingDeps{
		Profiles: &fakeProfiles{}, Listings: listings, Tx: &countingTx{}, Clock: fixedClock{},
	})
}

func listingCommand(isListed bool) service.SetListingCommand {
	return service.SetListingCommand{OrganizationID: ownOrganizationID, IsListed: isListed}
}

func TestSetListing_FRORG06_StaffIsForbidden(t *testing.T) {
	_, err := newSetListing(&fakeListings{}).Execute(t.Context(), storeMember(actor.Staff), listingCommand(true))

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestSetListing_BRORG05_IncompleteProfileIsRefusedWithWhatIsMissing(t *testing.T) {
	listings := &fakeListings{facts: domain.ListingFacts{Type: actor.Store, HasName: true, HasCity: true}}

	_, err := newSetListing(listings).Execute(t.Context(), storeMember(actor.Owner), listingCommand(true))

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeProfileIncomplete || len(listings.changes) != 0 ||
		appErr.Message() != "Lengkapi profil dulu: lokasi, jenis toko, kategori yang diterima." {
		t.Fatalf("err = %v, message = %q", err, appErr.Message())
	}
}

func TestSetListing_BRORG05_CompleteStoreIsListed(t *testing.T) {
	listings := &fakeListings{facts: domain.ListingFacts{
		Type: actor.Store, HasName: true, HasCity: true, HasLocation: true, HasStoreType: true, AcceptedCategories: 2,
	}}

	_, err := newSetListing(listings).Execute(t.Context(), storeMember(actor.Owner), listingCommand(true))

	if err != nil || len(listings.changes) != 1 || !listings.changes[0].IsListed {
		t.Fatalf("err = %v, changes = %+v", err, listings.changes)
	}
}

func TestSetListing_HidingNeverNeedsACompleteProfile(t *testing.T) {
	listings := &fakeListings{}

	_, err := newSetListing(listings).Execute(t.Context(), storeMember(actor.Owner), listingCommand(false))

	if err != nil || len(listings.changes) != 1 || listings.changes[0].IsListed {
		t.Fatalf("err = %v, changes = %+v", err, listings.changes)
	}
}
