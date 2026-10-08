package repository_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

const (
	nearBakery     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6c02"
	farSnackMaker  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6c15"
	breadProduct   = "0192f0c4-0000-7000-8000-000000000201"
	cookieProduct  = "0192f0c4-0000-7000-8000-000000000202"
	snackProduct   = "0192f0c4-0000-7000-8000-000000000203"
	shortShelfLife = 3
	longShelfLife  = 90
)

type productSeed struct {
	id        string
	owner     string
	name      string
	category  string
	price     int64
	shelfLife int64
	seeking   bool
}

func (w world) consignor(seed storeSeed, radiusKm int64, halalNumber string) {
	w.t.Helper()
	ctx := w.t.Context()
	w.organization(seed, "CONSIGNOR")
	w.must(w.queries.CreateConsignorProfile(ctx, sqlc.CreateConsignorProfileParams{
		OrganizationID: seed.id, CreatedAt: base(),
	}))
	_, err := w.queries.UpdateConsignorProfile(ctx, sqlc.UpdateConsignorProfileParams{
		OrganizationID: seed.id, HasDeliveryRadius: true, DeliveryRadiusKm: radiusKm, DeliveryDays: []string{},
		PreferredStoreTypes: []string{"CANTEEN"}, HalalCertNo: halalNumber, UpdatedAt: base(),
	})
	w.must(err)
	w.must(w.queries.SetOrganizationListed(ctx, sqlc.SetOrganizationListedParams{
		ID: seed.id, IsListed: true, UpdatedAt: base(),
	}))
}

func (w world) product(seed productSeed) {
	w.t.Helper()
	ctx := w.t.Context()
	_, err := w.queries.SaveProduct(ctx, sqlc.SaveProductParams{
		ID: seed.id, ConsignorOrgID: seed.owner, Sku: "SKU-" + seed.id[len(seed.id)-4:], Name: seed.name,
		CategoryID: seed.category, Unit: "pcs", SuggestedRetailPrice: seed.price, SuggestedDepositPrice: 0,
		ShelfLifeDays: shelfLifeOf(seed.shelfLife), ChangedAt: base(),
	})
	w.must(err)
	_, err = w.queries.SetProductSeeking(ctx, sqlc.SetProductSeekingParams{
		ID: seed.id, ConsignorOrgID: seed.owner, IsSeekingStores: seed.seeking, UpdatedAt: base(),
	})
	w.must(err)
}

func consignorWorld(t *testing.T) world {
	t.Helper()
	scene := newWorld(t)
	scene.consignor(storeSeed{id: nearBakery, name: "Roti Bu Ani", km: 2}, 5, "ID00110000123")
	scene.consignor(storeSeed{id: farSnackMaker, name: "Keripik Mbok Darmi", km: 15}, 30, "")
	scene.product(productSeed{breadProduct, nearBakery, "Roti sobek", breadCategoryID, 12000, shortShelfLife, true})
	scene.product(productSeed{cookieProduct, nearBakery, "Kue kacang", snackCategoryID, 25000, longShelfLife, false})
	scene.product(productSeed{
		snackProduct, farSnackMaker, "Keripik pedas", snackCategoryID, 15000, longShelfLife, true,
	})
	return scene
}

func consignorQuery(listing domain.Listing, filters domain.ConsignorFilters) domain.ConsignorQuery {
	if listing.Limit == 0 {
		listing.Limit = 20
	}
	listing.ViewerID = nearCanteen
	return domain.ConsignorQuery{
		Listing: listing, Filters: filters,
		ViewerLocation: domain.Point{Latitude: originLatitude + kmInLatitude, Longitude: originLongitude},
	}
}

func TestListConsignors_FRDIR02_NearestConsignorsWithTheirCategories(t *testing.T) {
	scene := consignorWorld(t)

	cards, err := scene.directory.ListConsignors(t.Context(), consignorQuery(nearest(), domain.ConsignorFilters{}))

	if err != nil || len(cards) != 2 || cards[0].ID != nearBakery || cards[0].ProductCount != 2 ||
		len(cards[0].Categories) != 2 || !cards[0].Traits.Legality.HasHalalCertificate ||
		cards[1].Traits.DeliveryRadiusKm != 30 {
		t.Fatalf("cards = %+v, err = %v", cards, err)
	}
}

func TestListConsignors_FRDIR07_FiltersNarrowTheConsignors(t *testing.T) {
	scene := consignorWorld(t)
	cases := []struct {
		name    string
		filters domain.ConsignorFilters
		want    []string
	}{
		{"halal", domain.ConsignorFilters{HasHalal: true}, []string{nearBakery}},
		{
			"delivers to the canteen",
			domain.ConsignorFilters{DeliversToViewer: true},
			[]string{nearBakery, farSnackMaker},
		},
		{
			"price range",
			domain.ConsignorFilters{MinPrice: domain.Bound{Value: 20000, IsGiven: true}},
			[]string{nearBakery},
		},
		{
			"shelf life",
			domain.ConsignorFilters{MinShelfLifeDays: domain.Bound{Value: 30, IsGiven: true}},
			[]string{nearBakery, farSnackMaker},
		},
		{"product name", domain.ConsignorFilters{Text: "keripik"}, []string{farSnackMaker}},
		{"category", domain.ConsignorFilters{CategoryIDs: []string{breadCategoryID}}, []string{nearBakery}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cards, err := scene.directory.ListConsignors(t.Context(), consignorQuery(nearest(), tc.filters))

			got := make([]string, 0, len(cards))
			for _, card := range cards {
				got = append(got, card.ID)
			}
			if err != nil || len(got) != len(tc.want) || (len(got) > 0 && got[0] != tc.want[0]) {
				t.Fatalf("got = %v, want %v, err = %v", got, tc.want, err)
			}
		})
	}
}

func TestListProducts_FRDIR02_OnlyProductsSeekingStores(t *testing.T) {
	scene := consignorWorld(t)

	cards, err := scene.directory.ListProducts(t.Context(), consignorQuery(nearest(), domain.ConsignorFilters{}))

	if err != nil || len(cards) != 2 || cards[0].ID != breadProduct || cards[0].Consignor.Name != "Roti Bu Ani" ||
		cards[0].ShelfLifeDays != shortShelfLife || cards[1].ID != snackProduct {
		t.Fatalf("cards = %+v, err = %v", cards, err)
	}
}

func TestCountProductsWithin_BRDIR03_CountsConsignorsWithSeekingProducts(t *testing.T) {
	scene := consignorWorld(t)
	query := consignorQuery(nearest(), domain.ConsignorFilters{CategoryIDs: []string{snackCategoryID}})

	products, productErr := scene.directory.CountProductsWithin(t.Context(), query)
	consignors, consignorErr := scene.directory.CountConsignorsWithin(t.Context(), query)

	if productErr != nil || consignorErr != nil || products[5] != 0 || products[25] != 1 || consignors[5] != 1 {
		t.Fatalf("products = %v, consignors = %v", products, consignors)
	}
}

func shelfLifeOf(days int64) pgtype.Int8 {
	return pgtype.Int8{Int64: days, Valid: days > 0}
}
