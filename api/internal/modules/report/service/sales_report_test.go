package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/service"
)

const storeID = "0192f0c4-0000-7000-8000-000000000401"

type fakeSales struct {
	asked []domain.Query
	rows  []domain.Row
}

func (f *fakeSales) SalesRows(_ context.Context, query domain.Query) ([]domain.Row, error) {
	f.asked = append(f.asked, query)
	return f.rows, nil
}

func storeMember(role actor.Role) actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: role,
	}}
}

func TestGetSalesReport_BRRPT02_ScopesTheReportToTheOwnersOrganization(t *testing.T) {
	sales := &fakeSales{rows: []domain.Row{{Key: "a", Figures: domain.Figures{SoldQty: 18}}}}

	report, err := service.NewGetSalesReport(sales).Execute(t.Context(), storeMember(actor.Owner),
		domain.Filter{From: "2026-11-01", To: "2026-11-30"})

	if err != nil || sales.asked[0].OrganizationID != storeID || report.Summary.SoldQty != 18 ||
		report.Query.GroupBy != domain.ByPartner {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
}

func TestGetSalesReport_FRRPT05_DayGroupingHasEveryDay(t *testing.T) {
	sales := &fakeSales{}

	report, err := service.NewGetSalesReport(sales).Execute(t.Context(), storeMember(actor.Owner),
		domain.Filter{From: "2026-11-01", To: "2026-11-07", GroupBy: "DAY"})

	if err != nil || len(report.Rows) != 7 {
		t.Fatalf("rows = %+v, err = %v", report.Rows, err)
	}
}

func TestGetSalesReport_BRRPT01_StaffCannotReadReports(t *testing.T) {
	_, err := service.NewGetSalesReport(&fakeSales{}).Execute(t.Context(), storeMember(actor.Staff),
		domain.Filter{From: "2026-11-01", To: "2026-11-30"})

	if appErr, ok := errors.AsType[*apperror.Error](err); !ok || appErr.Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}
