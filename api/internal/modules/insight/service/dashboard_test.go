package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/service"
)

type fakeDashboards struct {
	isQuiet          bool
	snapshotQueries  []domain.SnapshotQuery
	dashboardQueries []domain.DashboardQuery
	fail             error
}

func (f *fakeDashboards) ReadSnapshot(_ context.Context, query domain.SnapshotQuery) (domain.Snapshot, error) {
	f.snapshotQueries = append(f.snapshotQueries, query)
	if f.isQuiet {
		return domain.Snapshot{}, f.fail
	}
	overdue := domain.OpenSettlement{ID: "settlement-1", DueOn: query.Now.AddDate(0, 0, -2), Outstanding: 31500}
	settlements := []domain.OpenSettlement{overdue, overdue, overdue, overdue, overdue, overdue}
	return domain.Snapshot{Settlements: settlements}, f.fail
}

func (f *fakeDashboards) ReadDashboard(_ context.Context, query domain.DashboardQuery) (domain.Dashboard, error) {
	f.dashboardQueries = append(f.dashboardQueries, query)
	return domain.Dashboard{Figures: domain.Figures{Overdue: domain.Tally{Count: 1, Amount: 31500}}}, f.fail
}

func (f *fakeDashboards) ListOpenProposalCounterparts(context.Context, string) ([]string, error) {
	return []string{"org-proposed"}, nil
}

type fakeCandidates struct {
	fail error
}

func (f fakeCandidates) FindCandidates(context.Context, actor.Actor) ([]domain.Candidate, error) {
	return []domain.Candidate{
		{OrganizationID: "org-proposed", Name: "Warung Diajak", Score: 90},
		{OrganizationID: "org-new", Name: "Warung Baru", Score: 80},
	}, f.fail
}

func consignorStaff() actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: "consignor-1", OrganizationType: actor.Consignor, Role: actor.Staff,
	}}
}

func TestListInsights_RejectsVisitorWithoutOrganization(t *testing.T) {
	dashboards := &fakeDashboards{}

	_, err := service.NewListInsights(dashboards, fakeCandidates{}, fixedClock{}).
		Execute(t.Context(), actor.Actor{UserID: "u"})

	if apperror.From(err).Code() != apperror.CodeForbidden || len(dashboards.snapshotQueries) != 0 {
		t.Fatalf("err = %v, want FORBIDDEN without reading", err)
	}
}

func TestListInsights_FailureIsWrapped(t *testing.T) {
	_, err := service.NewListInsights(&fakeDashboards{fail: errDown}, fakeCandidates{}, fixedClock{}).
		Execute(t.Context(), storeStaff())

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v, want wrapped store failure", err)
	}
}

func TestListInsights_DSH_TopFiveFromTheActorSnapshotForItsSide(t *testing.T) {
	dashboards := &fakeDashboards{}

	got, err := service.NewListInsights(dashboards, fakeCandidates{}, fixedClock{}).
		Execute(t.Context(), consignorStaff())

	if err != nil || len(got) != 5 {
		t.Fatalf("insights = %d, err = %v, want 5", len(got), err)
	}
	if got[0].Kind != domain.KindSettlementOverdue || got[0].Action != domain.ActionRemindPayment {
		t.Fatalf("first = %+v, want an overdue settlement reminder for the consignor", got[0])
	}
	if query := dashboards.snapshotQueries[0]; query.OrganizationID != "consignor-1" ||
		!query.Now.Equal(fixedClock{}.Now()) {
		t.Fatalf("query = %+v", query)
	}
}

func TestGetDashboard_RejectsVisitorWithoutOrganization(t *testing.T) {
	dashboards := &fakeDashboards{}

	_, err := service.NewGetDashboard(dashboards, fixedClock{}).Execute(t.Context(), actor.Actor{UserID: "u"})

	if apperror.From(err).Code() != apperror.CodeForbidden || len(dashboards.dashboardQueries) != 0 {
		t.Fatalf("err = %v, want FORBIDDEN without reading", err)
	}
}

func TestGetDashboard_FailureIsWrapped(t *testing.T) {
	_, err := service.NewGetDashboard(&fakeDashboards{fail: errDown}, fixedClock{}).Execute(t.Context(), storeStaff())

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v, want wrapped store failure", err)
	}
}

func TestGetDashboard_DSH_ReadsTheActorDashboardOnTheServerDate(t *testing.T) {
	dashboards := &fakeDashboards{}

	got, err := service.NewGetDashboard(dashboards, fixedClock{}).Execute(t.Context(), storeStaff())

	if err != nil || got.Figures.Overdue != (domain.Tally{Count: 1, Amount: money.Rupiah(31500)}) {
		t.Fatalf("dashboard = %+v, err = %v", got, err)
	}
	query := dashboards.dashboardQueries[0]
	if query.OrganizationID != storeID || query.Today != "2026-11-11" ||
		!query.PendingSince.Equal(fixedClock{}.Now().Add(-48*time.Hour)) {
		t.Fatalf("query = %+v", query)
	}
}

func TestListInsights_INS10_SuggestsMatchesWithoutAnOpenProposal(t *testing.T) {
	dashboards := &fakeDashboards{isQuiet: true}

	got, err := service.NewListInsights(dashboards, fakeCandidates{}, fixedClock{}).
		Execute(t.Context(), storeStaff())

	if err != nil || len(got) != 1 || got[0].Kind != domain.KindMatchingPartner ||
		got[0].Subject.OrganizationID != "org-new" {
		t.Fatalf("insights = %+v, err = %v, want only org-new", got, err)
	}
}

func TestListInsights_CandidateFailureIsWrapped(t *testing.T) {
	_, err := service.NewListInsights(&fakeDashboards{}, fakeCandidates{fail: errDown}, fixedClock{}).
		Execute(t.Context(), storeStaff())

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v, want wrapped candidate failure", err)
	}
}
