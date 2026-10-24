package repository_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

const (
	nearNeed = "0192f0c4-0000-7000-8000-000000000901"
	farNeed  = "0192f0c4-0000-7000-8000-000000000902"
	oldNeed  = "0192f0c4-0000-7000-8000-000000000903"
)

func (w world) need(id, storeID string, expiresAt time.Time) {
	w.t.Helper()
	created, err := w.queries.CreateNeed(w.t.Context(), sqlc.CreateNeedParams{
		ID: id, StoreOrgID: storeID, Title: "Butuh camilan " + id[len(id)-3:], CategoryID: snackCategoryID,
		MaxRetailPrice: pgtype.Int8{Int64: 5000, Valid: true}, ExpiresAt: expiresAt, CreatedBy: userID,
		CreatedAt: base(),
	})
	w.must(err)
	if created != 1 {
		w.t.Fatal("need not created")
	}
}

func needWorld(t *testing.T) world {
	t.Helper()
	scene := newWorld(t)
	scene.must(scene.queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "kantin@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Bu Tini", Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}))
	scene.need(nearNeed, nearCanteen, base().Add(30*24*time.Hour))
	scene.need(farNeed, twentyKmGrocery, base().Add(30*24*time.Hour))
	scene.need(oldNeed, threeKmCafe, base().Add(time.Hour))
	return scene
}

func needQuery(listing domain.Listing) domain.NeedQuery {
	listing.Limit, listing.ViewerID = 20, consignorID
	return domain.NeedQuery{Listing: listing, Now: base().Add(2 * time.Hour)}
}

func TestListNeeds_FRNED03_OpenNeedsNearestFirstWithoutExpiredOnes(t *testing.T) {
	scene := needWorld(t)

	cards, err := scene.directory.ListNeeds(t.Context(), needQuery(nearest()))

	if err != nil || len(cards) != 2 || cards[0].ID != nearNeed || cards[1].ID != farNeed ||
		cards[0].Store.Name != "Kantin Teknik 100%" || cards[0].MaxRetailPrice != 5000 ||
		cards[0].HasOfferedCommission {
		t.Fatalf("cards = %+v, err = %v", cards, err)
	}
}

func TestCountNeedsWithin_BRDIR03_CountsOpenNeedsPerRadius(t *testing.T) {
	scene := needWorld(t)

	counts, err := scene.directory.CountNeedsWithin(t.Context(), needQuery(nearest()))

	if err != nil || counts[2] != 1 || counts[25] != 2 {
		t.Fatalf("counts = %v, err = %v", counts, err)
	}
}

func TestListStores_FRNED04_CardCarriesTheOpenNeedCategories(t *testing.T) {
	scene := needWorld(t)

	cards := scene.list(domain.StoreQuery{
		Listing: nearest(), Filters: domain.StoreFilters{HasOpenNeed: true}, Now: base(),
	})

	if len(cards) != 3 || len(cards[0].NeedCategoryIDs) != 1 || cards[0].NeedCategoryIDs[0] != snackCategoryID {
		t.Fatalf("cards = %+v", cards)
	}
}
