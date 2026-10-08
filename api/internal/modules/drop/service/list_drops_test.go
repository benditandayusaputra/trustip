package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
)

type fakeLister struct {
	queries []domain.PageQuery
	total   int
}

func (f *fakeLister) ListDrops(_ context.Context, query domain.PageQuery) ([]domain.Summary, error) {
	f.queries = append(f.queries, query)
	summaries := []domain.Summary{}
	start := time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
	for index := range min(f.total, query.Limit) {
		summaries = append(summaries, domain.Summary{
			ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index), OccurredAt: start.Add(-time.Duration(index)),
		})
	}
	return summaries, nil
}

func TestListDrops_FRDRP08_PagesWithACursorInTheActorsScope(t *testing.T) {
	lister := &fakeLister{total: 3}
	list := service.NewListDrops(lister, movingClock{at: new(time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC))})

	page, err := list.Execute(t.Context(), consignorOwner(), service.ListRequest{
		Filter: domain.ListFilter{Status: "RECORDED"}, Limit: "2",
	})

	if err != nil || len(page.Items) != 2 || !page.HasMore || page.NextCursor == "" {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
	if lister.queries[0].OrganizationID != consignorID || lister.queries[0].Limit != 3 ||
		lister.queries[0].Status != domain.Recorded {
		t.Fatalf("query = %+v", lister.queries[0])
	}
}

func TestListDrops_UnknownStatusIsAValidationError(t *testing.T) {
	list := service.NewListDrops(&fakeLister{}, movingClock{at: new(time.Time)})

	_, err := list.Execute(t.Context(), cashier(), service.ListRequest{Filter: domain.ListFilter{Status: "X"}})

	if codeOf(err) != apperror.CodeValidationFailed {
		t.Fatalf("err = %v, want VALIDATION_FAILED", err)
	}
}

func TestListDrops_RSEC36_RequiresAMembership(t *testing.T) {
	list := service.NewListDrops(&fakeLister{}, movingClock{at: new(time.Time)})

	_, err := list.Execute(t.Context(), actor.Actor{UserID: cashierID}, service.ListRequest{})

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}
