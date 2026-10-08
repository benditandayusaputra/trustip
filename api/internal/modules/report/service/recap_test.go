package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/service"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 19, 18, 0, 0, 0, time.UTC)
}

type fakeRecaps struct {
	daily       []domain.DailyQuery
	settlements []domain.SettlementQuery
}

func (f *fakeRecaps) DailyRows(_ context.Context, query domain.DailyQuery) ([]domain.RecapRow, error) {
	f.daily = append(f.daily, query)
	return []domain.RecapRow{{Name: "A", Paid: 100}, {Name: "B", Paid: 50}}, nil
}

func (f *fakeRecaps) SettlementRows(_ context.Context, query domain.SettlementQuery) ([]domain.SettlementRow, error) {
	f.settlements = append(f.settlements, query)
	return []domain.SettlementRow{{Issued: domain.Tally{Count: 2, Amount: 900}}}, nil
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return appErr.Code()
	}
	return ""
}

func TestGetDailyRecap_FRRPT01_StoreOwnerGetsTodaysRecapPerConsignor(t *testing.T) {
	recaps := &fakeRecaps{}

	recap, err := service.NewGetDailyRecap(recaps, fixedClock{}).Execute(t.Context(), storeMember(actor.Owner), "")

	if err != nil || recaps.daily[0].Day != "2026-11-20" || recaps.daily[0].OrganizationID != storeID ||
		len(recap.Rows) != 2 || recap.Total.Paid != 150 {
		t.Fatalf("recap = %+v, err = %v", recap, err)
	}
}

func TestGetDailyRecap_FRRPT01_IsAStoreReport(t *testing.T) {
	consignor := storeMember(actor.Owner)
	consignor.Membership.OrganizationType = actor.Consignor

	_, err := service.NewGetDailyRecap(&fakeRecaps{}, fixedClock{}).Execute(t.Context(), consignor, "")

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestGetSettlementReport_FRRPT06_ComparesDueDatesWithTodayInJakarta(t *testing.T) {
	recaps := &fakeRecaps{}

	report, err := service.NewGetSettlementReport(recaps, fixedClock{}).Execute(t.Context(), storeMember(actor.Owner),
		domain.Filter{From: "2026-11-01", To: "2026-11-30"})

	if err != nil || recaps.settlements[0].Today != "2026-11-20" || report.Total.Issued.Count != 2 {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
}
