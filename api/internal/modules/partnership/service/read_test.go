package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
)

const (
	storeID       = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	partnershipID = "0192f0c4-0000-7000-8000-000000000501"
)

type fakeReader struct {
	asked []domain.ListQuery
	count int
}

func (f *fakeReader) ListPartnerships(_ context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	f.asked = append(f.asked, query)
	found := []domain.Summary{}
	for index := range min(f.count, query.Limit) {
		found = append(found, domain.Summary{
			ID:        fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index),
			StartedAt: time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC),
		})
	}
	return found, nil
}

func (f *fakeReader) GetPartnership(_ context.Context, key domain.Key) (domain.Partnership, error) {
	if key.OrganizationID != storeID || key.PartnershipID != partnershipID {
		return domain.Partnership{}, apperror.NotFound()
	}
	return domain.Partnership{ID: partnershipID, Status: domain.Active}, nil
}

func (f *fakeReader) ListCatalog(context.Context, domain.Key) ([]domain.CatalogItem, error) {
	return []domain.CatalogItem{{ProductID: "p1", IsAgreed: true, IsActive: true}}, nil
}

func staff(organizationID string) actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: organizationID, OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func TestListPartnerships_FRPRT01_PagesThroughTheOrganizationsPartners(t *testing.T) {
	reader := &fakeReader{count: 3}

	page, err := service.NewListPartnerships(reader).Execute(t.Context(), staff(storeID),
		service.ListRequest{Status: "ENDED", Limit: "2"})

	if err != nil || len(page.Items) != 2 || !page.HasMore || reader.asked[0].Status != domain.Ended ||
		reader.asked[0].OrganizationID != storeID {
		t.Fatalf("page = %+v, asked = %+v, err = %v", page, reader.asked, err)
	}
}

func TestListPartnerships_NewcomerIsForbidden(t *testing.T) {
	_, err := service.NewListPartnerships(&fakeReader{}).Execute(t.Context(), actor.Actor{UserID: "u"},
		service.ListRequest{})

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestListCatalog_FRPRD06_StoreSeesThePartnerCatalogWithTerms(t *testing.T) {
	items, err := service.NewListCatalog(&fakeReader{}).Execute(t.Context(), staff(storeID), partnershipID)

	if err != nil || len(items) != 1 || !items[0].IsAgreed {
		t.Fatalf("items = %+v, err = %v", items, err)
	}
}

func TestListCatalog_BRAUTHZ02_OtherOrganizationGetsNotFound(t *testing.T) {
	_, err := service.NewListCatalog(&fakeReader{}).Execute(t.Context(), staff("elsewhere"), partnershipID)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}
