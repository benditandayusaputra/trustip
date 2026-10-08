package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
)

const targetID = "0192f0c4-0000-7000-8000-000000000101"

type fakeReports struct {
	filings []domain.Filing
	fail    error
}

func (f *fakeReports) SaveReport(_ context.Context, filing domain.Filing) (domain.Filed, error) {
	f.filings = append(f.filings, filing)
	return domain.Filed{ID: filing.ID, Status: "OPEN"}, f.fail
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
}

type fixedIDs struct{}

func (fixedIDs) NewID() (string, error) {
	return "0192f0c4-0000-7000-8000-00000000aaaa", nil
}

func storeStaff() actor.Actor {
	return actor.Actor{UserID: "user-1", Membership: actor.Membership{
		OrganizationID: "store-1", OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func newReportListing(reports *fakeReports) *service.ReportListing {
	return service.NewReportListing(service.ReportDeps{Reports: reports, Clock: fixedClock{}, IDs: fixedIDs{}})
}

func spamForm() domain.ReportForm {
	return domain.ReportForm{TargetType: "PRODUCT", TargetID: targetID, Reason: "SPAM"}
}

func TestReportListing_RejectsVisitorWithoutOrganization(t *testing.T) {
	reports := &fakeReports{}

	_, err := newReportListing(reports).Execute(t.Context(), actor.Actor{UserID: "u"}, spamForm())

	if apperror.From(err).Code() != apperror.CodeForbidden || len(reports.filings) != 0 {
		t.Fatalf("err = %v, want FORBIDDEN without saving", err)
	}
}

func TestReportListing_FRTRS04_InvalidFormIsNotSaved(t *testing.T) {
	reports := &fakeReports{}

	_, err := newReportListing(reports).Execute(t.Context(), storeStaff(), domain.ReportForm{})

	if apperror.From(err).Code() != apperror.CodeValidationFailed || len(reports.filings) != 0 {
		t.Fatalf("err = %v, want VALIDATION_FAILED without saving", err)
	}
}

func TestReportListing_BRAUTHZ02_MissingOrOwnTargetIsNotFound(t *testing.T) {
	_, err := newReportListing(&fakeReports{fail: domain.ErrTargetNotFound}).
		Execute(t.Context(), storeStaff(), spamForm())

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestReportListing_FRTRS04_StaffFilesAReportForTheTeam(t *testing.T) {
	reports := &fakeReports{}

	got, err := newReportListing(reports).Execute(t.Context(), storeStaff(), spamForm())

	filing := reports.filings[0]
	if err != nil || got.Status != "OPEN" || filing.ReporterUserID != "user-1" || filing.ReporterOrgID != "store-1" ||
		!filing.ReportedAt.Equal(fixedClock{}.Now()) || filing.Draft.Reason != domain.ReasonSpam {
		t.Fatalf("filed = %+v, filing = %+v, err = %v", got, filing, err)
	}
}
