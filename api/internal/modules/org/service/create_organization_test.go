package service_test

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

const founderID = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"

type fakeFounder struct {
	taken      []domain.Slug
	failures   []error
	foundings  []domain.Founding
	slugLookup []domain.Slug
}

func (f *fakeFounder) TakenSlugs(_ context.Context, base domain.Slug) ([]domain.Slug, error) {
	f.slugLookup = append(f.slugLookup, base)
	return f.taken, nil
}

func (f *fakeFounder) CreateOrganization(_ context.Context, founding domain.Founding) (string, error) {
	if len(f.failures) > 0 {
		failure := f.failures[0]
		f.failures = f.failures[1:]
		return "", failure
	}
	f.foundings = append(f.foundings, founding)
	return "Kota Semarang", nil
}

type fakeAudit struct{ entries []audit.Entry }

func (f *fakeAudit) Record(_ context.Context, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeAudit) actions() []audit.Action {
	actions := []audit.Action{}
	for _, entry := range f.entries {
		actions = append(actions, entry.Action)
	}
	return actions
}

type countingTx struct{ count int }

func (tx *countingTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	tx.count++
	return work(ctx)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC) }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	attempt := strconv.Itoa((s.next + 1) / 2)
	if s.next%2 == 1 {
		return "org-" + attempt, nil
	}
	return "membership-" + attempt, nil
}

type createFixture struct {
	audit   *fakeAudit
	founder *fakeFounder
	tx      *countingTx
	useCase *service.CreateOrganization
}

func newCreateFixture() createFixture {
	fixture := createFixture{founder: &fakeFounder{}, tx: &countingTx{}, audit: &fakeAudit{}}
	fixture.useCase = service.NewCreateOrganization(service.CreateDeps{
		Founder: fixture.founder, Audit: fixture.audit, Tx: fixture.tx, Clock: fixedClock{}, IDs: &sequenceIDs{},
	})
	return fixture
}

func newcomer() actor.Actor {
	return actor.Actor{UserID: founderID, SessionFamily: "family-1"}
}

func storeForm() domain.OnboardingForm {
	return domain.OnboardingForm{Type: "STORE", ProfileForm: domain.ProfileForm{
		Name: "Warung Bu Tini", WhatsApp: "081234567890", CityCode: "33.74",
	}}
}

func TestCreateOrganization_MemberIsRefusedBeforeAnythingIsWritten(t *testing.T) {
	fixture := newCreateFixture()
	signedIn := newcomer()
	signedIn.Membership = actor.Membership{OrganizationID: "org-0", OrganizationType: actor.Store, Role: actor.Owner}

	_, err := fixture.useCase.Execute(t.Context(), signedIn, storeForm())

	if apperror.From(err).Code() != apperror.CodeOrganizationExists || fixture.tx.count != 0 {
		t.Fatalf("err = %v, tx = %d", err, fixture.tx.count)
	}
}

func TestCreateOrganization_InvalidFormIsRejectedWithoutWriting(t *testing.T) {
	fixture := newCreateFixture()
	form := storeForm()
	form.Name = "W"

	_, err := fixture.useCase.Execute(t.Context(), newcomer(), form)

	if apperror.From(err).Code() != apperror.CodeValidationFailed || fixture.tx.count != 0 {
		t.Fatalf("err = %v, tx = %d", err, fixture.tx.count)
	}
}

func TestCreateOrganization_UnknownCityIsAFieldError(t *testing.T) {
	fixture := newCreateFixture()
	fixture.founder.failures = []error{domain.ErrUnknownCity}

	_, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm())

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "city_code" || fields[0].Code != "UNKNOWN_CHOICE" {
		t.Fatalf("err = %v, fields = %+v", err, fields)
	}
}

func TestCreateOrganization_ConcurrentOnboardingOfTheSameUserIsAConflict(t *testing.T) {
	fixture := newCreateFixture()
	fixture.founder.failures = []error{domain.ErrAlreadyMember}

	_, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm())

	if apperror.From(err).Code() != apperror.CodeOrganizationExists {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateOrganization_BRORG02_SlugRaceIsRetriedInAFreshTransaction(t *testing.T) {
	fixture := newCreateFixture()
	fixture.founder.failures = []error{domain.ErrSlugTaken}

	organization, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if fixture.tx.count != 2 || len(fixture.founder.slugLookup) != 2 || organization.ID != "org-2" {
		t.Fatalf("tx = %d, lookups = %v, organization = %+v",
			fixture.tx.count, fixture.founder.slugLookup, organization)
	}
}

func TestCreateOrganization_BRORG02_SlugRaceThatNeverEndsGivesUp(t *testing.T) {
	fixture := newCreateFixture()
	fixture.founder.failures = []error{domain.ErrSlugTaken, domain.ErrSlugTaken, domain.ErrSlugTaken}

	_, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm())

	if !errors.Is(err, domain.ErrSlugTaken) || fixture.tx.count != 3 {
		t.Fatalf("err = %v, tx = %d", err, fixture.tx.count)
	}
}

func TestCreateOrganization_BRAUTH02_FoundsTheOrganizationWithTheUserAsOwner(t *testing.T) {
	fixture := newCreateFixture()
	fixture.founder.taken = []domain.Slug{"warung-bu-tini"}

	organization, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	founding := fixture.founder.foundings[0]
	if founding.Owner != (domain.Membership{ID: "membership-1", UserID: founderID, Role: actor.Owner}) {
		t.Fatalf("owner = %+v", founding.Owner)
	}
	if organization.Slug != "warung-bu-tini-2" || organization.CityName != "Kota Semarang" ||
		!slices.Equal(fixture.founder.slugLookup, []domain.Slug{"warung-bu-tini"}) || fixture.tx.count != 1 {
		t.Fatalf("organization = %+v", organization)
	}
}

func TestCreateOrganization_NFR14_FoundingIsAuditedWithTheOrganizationType(t *testing.T) {
	fixture := newCreateFixture()

	if _, err := fixture.useCase.Execute(t.Context(), newcomer(), storeForm()); err != nil {
		t.Fatalf("execute: %v", err)
	}

	entries := fixture.audit.entries
	if len(entries) != 1 || entries[0].Action != audit.OrganizationCreated || entries[0].OrganizationID != "org-1" ||
		entries[0].ActorID != founderID || entries[0].Metadata["type"] != "STORE" {
		t.Fatalf("entries = %+v", entries)
	}
}
