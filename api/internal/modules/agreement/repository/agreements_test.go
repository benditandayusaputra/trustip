package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	consignorID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a71"
	storeID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a73"
)

func TestCreatePartnership_BRPRT01_OneActivePartnershipPerPair(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	startedAt := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	for id, kind := range map[string]string{consignorID: "CONSIGNOR", storeID: "STORE"} {
		if _, err := queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
			ID: id, Type: kind, Name: kind, Slug: "usaha-" + id[len(id)-2:], Whatsapp: "6281234567890",
			CityCode: "33.74", CreatedAt: startedAt,
		}); err != nil {
			t.Fatalf("organization: %v", err)
		}
	}
	agreements := repository.NewAgreements(pool)
	first := domain.Partnership{
		ID: "0192f0c4-0000-7000-8000-000000000501", ConsignorID: consignorID, StoreID: storeID,
		StartedAt: startedAt,
	}
	second := first
	second.ID = "0192f0c4-0000-7000-8000-000000000502"

	firstErr := agreements.CreatePartnership(t.Context(), first)
	secondErr := agreements.CreatePartnership(t.Context(), second)

	if firstErr != nil || !errors.Is(secondErr, domain.ErrAlreadyPartners) {
		t.Fatalf("first = %v, second = %v", firstErr, secondErr)
	}
}
