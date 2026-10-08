package repository_test

import (
	"math"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	consignorID      = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a71"
	nearCanteen      = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b01"
	threeKmCafe      = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b03"
	sevenKmCanteen   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b07"
	twentyKmGrocery  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b20"
	unlistedStore    = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6b99"
	userID           = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	sessionID        = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d1e"
	membershipID     = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d2e"
	breadCategoryID  = "01a1161e-0a30-752d-ab9b-7397976f0dba"
	snackCategoryID  = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"
	semarangCode     = "33.74"
	originLatitude   = -6.99
	originLongitude  = 110.42
	kmInLatitude     = 1 / 111.195
	toleranceMeters  = 5
	storedHash       = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
	productSKUPrefix = "RT-000"
	sha256Size       = 32
)

func base() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

type storeSeed struct {
	id         string
	name       string
	km         float64
	storeType  string
	categories []string
	halal      bool
	listed     bool
}

type world struct {
	t         *testing.T
	queries   *sqlc.Queries
	directory *repository.Directory
}

func newWorld(t *testing.T) world {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	scene := world{t: t, queries: sqlc.New(pool.Querier(t.Context())), directory: repository.NewDirectory(pool)}
	scene.organization(storeSeed{id: consignorID, name: "Roti Pak Anton"}, "CONSIGNOR")
	scene.must(scene.queries.CreateConsignorProfile(t.Context(), sqlc.CreateConsignorProfileParams{
		OrganizationID: consignorID, CreatedAt: base(),
	}))
	for _, seed := range []storeSeed{
		{nearCanteen, "Kantin Teknik 100%", 1, "CANTEEN", []string{snackCategoryID}, false, true},
		{threeKmCafe, "Kedai Kopi Sore", 3, "CAFE", []string{breadCategoryID}, true, true},
		{sevenKmCanteen, "Kantin Kampus", 7, "CANTEEN", []string{snackCategoryID, breadCategoryID}, false, true},
		{twentyKmGrocery, "Warung Bu Sri", 20, "GROCERY_STALL", []string{breadCategoryID}, false, true},
		{unlistedStore, "Toko Belum Tampil", 0.5, "CANTEEN", []string{snackCategoryID}, false, false},
	} {
		scene.store(seed)
	}
	return scene
}

func (w world) must(err error) {
	w.t.Helper()
	if err != nil {
		w.t.Fatal(err)
	}
}

func (w world) organization(seed storeSeed, kind string) {
	w.t.Helper()
	latitude := originLatitude + seed.km*kmInLatitude
	_, err := w.queries.CreateOrganization(w.t.Context(), sqlc.CreateOrganizationParams{
		ID: seed.id, Type: kind, Name: seed.name, Slug: "org-" + seed.id[len(seed.id)-4:], Whatsapp: "6281234567890",
		CityCode: semarangCode, HasLocation: true, Latitude: latitude, Longitude: originLongitude,
		PublicLatitude: latitude, PublicLongitude: originLongitude, CreatedAt: base(),
	})
	w.must(err)
}

func (w world) store(seed storeSeed) {
	w.t.Helper()
	ctx := w.t.Context()
	w.organization(seed, "STORE")
	w.must(w.queries.CreateStoreProfile(ctx, sqlc.CreateStoreProfileParams{OrganizationID: seed.id, CreatedAt: base()}))
	_, err := w.queries.UpdateStoreProfile(ctx, sqlc.UpdateStoreProfileParams{
		OrganizationID: seed.id, StoreType: seed.storeType, AcceptsCommission: true, HasCommission: true,
		CommissionMinBps: 1000, CommissionMaxBps: 2000, RequiresHalal: seed.halal, UpdatedAt: base(),
	})
	w.must(err)
	_, err = w.queries.AddAcceptedCategories(ctx, sqlc.AddAcceptedCategoriesParams{
		OrganizationID: seed.id, CategoryIds: seed.categories,
	})
	w.must(err)
	w.must(w.queries.SetOrganizationListed(ctx, sqlc.SetOrganizationListedParams{
		ID: seed.id, IsListed: seed.listed, UpdatedAt: base(),
	}))
}

func nearOrigin() domain.Origin {
	return domain.Origin{Point: domain.Point{Latitude: originLatitude, Longitude: originLongitude}, IsKnown: true}
}

func nearest() domain.Listing {
	return domain.Listing{Origin: nearOrigin(), Sort: domain.Nearest}
}

func withinKm(radiusKm int) domain.Listing {
	listing := nearest()
	listing.RadiusKm = radiusKm
	return listing
}

func firstOf(limit int) domain.Listing {
	listing := nearest()
	listing.Limit = limit
	return listing
}

func after(position domain.Position) domain.Listing {
	listing := nearest()
	listing.After = position
	return listing
}

func (w world) list(query domain.StoreQuery) []domain.StoreCard {
	w.t.Helper()
	if query.Limit == 0 {
		query.Limit = 20
	}
	query.ViewerID = consignorID
	cards, err := w.directory.ListStores(w.t.Context(), query)
	w.must(err)
	return cards
}

func idsOf(cards []domain.StoreCard) []string {
	ids := make([]string, 0, len(cards))
	for _, card := range cards {
		ids = append(ids, card.ID)
	}
	return ids
}

func requireIDs(t *testing.T, cards []domain.StoreCard, want ...string) {
	t.Helper()
	got := idsOf(cards)
	if len(got) != len(want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ids = %v, want %v", got, want)
		}
	}
}

func TestListStores_RCC51_FRDIR03_NearestFirstWithStraightLineDistance(t *testing.T) {
	scene := newWorld(t)

	cards := scene.list(domain.StoreQuery{Listing: nearest()})

	requireIDs(t, cards, nearCanteen, threeKmCafe, sevenKmCanteen, twentyKmGrocery)
	if math.Abs(float64(cards[1].Distance.Meters-3000)) > toleranceMeters || !cards[1].Distance.IsKnown {
		t.Fatalf("distance = %+v", cards[1].Distance)
	}
	if len(cards[2].AcceptedCategories) != 2 || !cards[0].Terms.Commission.IsKnown {
		t.Fatalf("card = %+v", cards[2])
	}
}

func TestListStores_FRDIR05_RadiusLimitsTheResults(t *testing.T) {
	scene := newWorld(t)

	cards := scene.list(domain.StoreQuery{Listing: withinKm(5)})

	requireIDs(t, cards, nearCanteen, threeKmCafe)
}

func TestCountStoresWithin_BRDIR03_CountsEveryRadiusStep(t *testing.T) {
	scene := newWorld(t)

	counts, err := scene.directory.CountStoresWithin(t.Context(), domain.StoreQuery{Listing: nearest()})

	if err != nil || counts[2] != 1 || counts[5] != 2 || counts[10] != 3 || counts[25] != 4 {
		t.Fatalf("counts = %v, err = %v", counts, err)
	}
}

func TestListStores_FRDIR06_FiltersNarrowTheResults(t *testing.T) {
	scene := newWorld(t)
	cases := []struct {
		name    string
		filters domain.StoreFilters
		want    []string
	}{
		{"store type", domain.StoreFilters{StoreType: "CANTEEN"}, []string{nearCanteen, sevenKmCanteen}},
		{
			"category",
			domain.StoreFilters{CategoryIDs: []string{breadCategoryID}},
			[]string{threeKmCafe, sevenKmCanteen, twentyKmGrocery},
		},
		{
			"without halal",
			domain.StoreFilters{WithoutHalal: true},
			[]string{nearCanteen, sevenKmCanteen, twentyKmGrocery},
		},
		{"commission model", domain.StoreFilters{PricingModel: "FIXED_DEPOSIT"}, nil},
		{"max commission", domain.StoreFilters{HasMaxCommission: true, MaxCommissionBPS: 999}, nil},
		{"text", domain.StoreFilters{Text: "kantin"}, []string{nearCanteen, sevenKmCanteen}},
		{"RSEC59 percent is literal", domain.StoreFilters{Text: "100%"}, []string{nearCanteen}},
		{"RSEC59 underscore is literal", domain.StoreFilters{Text: "_"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cards := scene.list(domain.StoreQuery{Listing: nearest(), Filters: tc.filters})

			requireIDs(t, cards, tc.want...)
		})
	}
}

func TestListStores_BRDIR05_KeysetContinuesAfterTheLastCard(t *testing.T) {
	scene := newWorld(t)
	first := scene.list(domain.StoreQuery{Listing: firstOf(2)})

	rest := scene.list(domain.StoreQuery{Listing: after(first[1].Position)})

	requireIDs(t, first, nearCanteen, threeKmCafe)
	requireIDs(t, rest, sevenKmCanteen, twentyKmGrocery)
}

func TestListStores_BRDIR04_WithoutOriginDistanceIsUnknown(t *testing.T) {
	scene := newWorld(t)

	cards := scene.list(domain.StoreQuery{Listing: domain.Listing{Sort: domain.ByCity}})

	if len(cards) != 4 || cards[0].Distance.IsKnown || cards[0].Position.Text != "Kota Semarang" {
		t.Fatalf("cards = %+v", cards)
	}
}

func TestGetViewer_BRDIR06_CategoriesComeFromActiveProducts(t *testing.T) {
	scene := newWorld(t)
	_, err := scene.queries.SaveProduct(t.Context(), sqlc.SaveProductParams{
		ID: "0192f0c4-0000-7000-8000-000000000101", ConsignorOrgID: consignorID, Sku: productSKUPrefix + "1",
		Name: "Roti manis", CategoryID: breadCategoryID, Unit: "pcs", SuggestedRetailPrice: 2000,
		SuggestedDepositPrice: 1750, ChangedAt: base(),
	})
	scene.must(err)

	viewer, err := scene.directory.GetViewer(t.Context(), domain.ViewerKey{
		OrganizationID: consignorID, CityCode: semarangCode,
	})

	if err != nil || !viewer.HasLocation || !viewer.HasCity || !viewer.IsConsignor() ||
		len(viewer.Consignor.Categories) != 1 || viewer.Consignor.Categories[0].Name != "Roti dan kue basah" {
		t.Fatalf("viewer = %+v, err = %v", viewer, err)
	}
}

func TestSessions_FRDIR10_SigningInMarksTheOrganizationActive(t *testing.T) {
	scene := newWorld(t)
	ctx := t.Context()
	scene.must(scene.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID: userID, Email: "kantin@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Bu Tini", Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}))
	scene.must(scene.queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		ID: membershipID, OrganizationID: nearCanteen, UserID: userID, Role: "OWNER", CreatedAt: base(),
	}))
	signedInAt := base().Add(time.Hour)

	scene.must(scene.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID: sessionID, UserID: userID, FamilyID: sessionID, TokenHash: make([]byte, sha256Size),
		ExpiresAt: signedInAt.Add(time.Hour),
		CreatedAt: signedInAt,
	}))

	cards := scene.list(domain.StoreQuery{Listing: firstOf(1)})
	if !cards[0].LastActiveAt.Equal(signedInAt) {
		t.Fatalf("last active = %v", cards[0].LastActiveAt)
	}
}
