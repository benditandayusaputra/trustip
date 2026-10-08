package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	locatedStore  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a80"
	nearConsignor = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a81"
	farConsignor  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a82"
	kmInLatitude  = 1 / 111.195
)

type located struct {
	id   string
	kind string
	km   float64
}

func locatedOrganization(t *testing.T, queries *sqlc.Queries, seed located) {
	t.Helper()
	id, kind := seed.id, seed.kind
	latitude := -6.99 + seed.km*kmInLatitude
	_, err := queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
		ID: id, Type: kind, Name: "Usaha " + id[len(id)-2:], Slug: "usaha-" + id[len(id)-2:],
		Whatsapp: "6281234567890", CityCode: "33.74", HasLocation: true, Latitude: latitude, Longitude: 110.42,
		PublicLatitude: latitude, PublicLongitude: 110.42, CreatedAt: base(),
	})
	must(t, err)
}

func TestListInterested_FRNED07_ConsignorsOfTheCategoryWithinTenKilometres(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	locatedOrganization(t, queries, located{locatedStore, "STORE", 0})
	locatedOrganization(t, queries, located{nearConsignor, "CONSIGNOR", 2})
	locatedOrganization(t, queries, located{farConsignor, "CONSIGNOR", 15})
	for index, consignor := range []string{nearConsignor, farConsignor} {
		_, err := queries.SaveProduct(t.Context(), sqlc.SaveProductParams{
			ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index+101), ConsignorOrgID: consignor,
			Sku: "KUE-0001", Name: "Kue kacang", CategoryID: snacksID, Unit: "pcs", SuggestedRetailPrice: 5000,
			ChangedAt: base(),
		})
		must(t, err)
	}

	interested, err := repository.NewNeeds(pool).ListInterested(t.Context(), domain.Need{
		ID: "n1", StoreID: locatedStore, Draft: domain.Draft{CategoryID: snacksID},
	})

	if err != nil || len(interested) != 1 || interested[0].OrganizationID != nearConsignor ||
		interested[0].DistanceMeters < 1990 || interested[0].DistanceMeters > 2010 {
		t.Fatalf("interested = %+v, err = %v", interested, err)
	}
}

func TestClaimExpiring_BRNED03_ClaimsEachNeedOnlyOnce(t *testing.T) {
	needs := newNeeds(t)
	must(t, needs.CreateNeed(t.Context(), posting(0)))
	reminderDay := base().Add(domain.ShowingPeriod - 2*24*time.Hour)

	first, firstErr := needs.ClaimExpiring(t.Context(), domain.ReminderWindowAt(reminderDay))
	second, secondErr := needs.ClaimExpiring(t.Context(), domain.ReminderWindowAt(reminderDay.Add(time.Hour)))
	early, earlyErr := needs.ClaimExpiring(t.Context(), domain.ReminderWindowAt(base()))

	if firstErr != nil || secondErr != nil || earlyErr != nil || len(first) != 1 || len(second) != 0 ||
		len(early) != 0 || first[0].Title != "Butuh camilan kantin" {
		t.Fatalf("first = %+v, second = %+v, early = %+v", first, second, early)
	}
}
