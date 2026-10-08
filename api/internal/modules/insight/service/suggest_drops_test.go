package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/service"
)

const (
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
)

var errDown = errors.New("store down")

type fakeHistories struct {
	queries []domain.HistoryQuery
	fail    error
}

func (f *fakeHistories) ListDropHistories(
	_ context.Context, query domain.HistoryQuery,
) ([]domain.ProductHistory, error) {
	f.queries = append(f.queries, query)
	steady := []domain.DaySales{{Sold: 17}, {Sold: 17}, {Sold: 18}, {Sold: 16}}
	return []domain.ProductHistory{
		{ProductID: "roti", History: domain.DropHistory{Rhythm: terms.RhythmDaily, SameWeekdaySales: steady}},
	}, f.fail
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 11, 2, 0, 0, 0, time.UTC)
}

func storeStaff() actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func TestSuggestDrops_RejectsVisitorWithoutOrganization(t *testing.T) {
	histories := &fakeHistories{}

	visitor := actor.Actor{UserID: "u"}

	_, err := service.NewSuggestDrops(histories, fixedClock{}).Execute(t.Context(), visitor, partnershipID)

	if apperror.From(err).Code() != apperror.CodeForbidden || len(histories.queries) != 0 {
		t.Fatalf("err = %v, queries = %d, want FORBIDDEN without reading", err, len(histories.queries))
	}
}

func TestSuggestDrops_FailureIsWrapped(t *testing.T) {
	_, err := service.NewSuggestDrops(&fakeHistories{fail: errDown}, fixedClock{}).
		Execute(t.Context(), storeStaff(), partnershipID)

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v, want wrapped store failure", err)
	}
}

func TestSuggestDrops_INS09_SuggestsFromTheActorScopedHistory(t *testing.T) {
	histories := &fakeHistories{}

	got, err := service.NewSuggestDrops(histories, fixedClock{}).Execute(t.Context(), storeStaff(), partnershipID)

	want := domain.DropSuggestion{ProductID: "roti", Qty: 19, HasEnoughData: true}
	if err != nil || len(got) != 1 || got[0] != want {
		t.Fatalf("suggestions = %+v, err = %v, want roti 19", got, err)
	}
	query := histories.queries[0]
	if query.OrganizationID != storeID || query.PartnershipID != partnershipID || query.ISOWeekday != 3 {
		t.Fatalf("query = %+v, want store scope on Wednesday", query)
	}
}
