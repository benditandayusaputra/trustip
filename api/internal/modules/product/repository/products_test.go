package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	consignorID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	strangerID      = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	userID          = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	breadCategoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"
	chipsCategoryID = "01a1161e-0a62-7a64-ab6f-024c2a8cc433"
	storedHash      = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func base() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

type world struct {
	t        *testing.T
	pool     *database.Pool
	products *repository.Products
}

func newWorld(t *testing.T) world {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	if err := queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "darmi@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Darmi", Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}); err != nil {
		t.Fatalf("user: %v", err)
	}
	for id, slug := range map[string]string{consignorID: "kripik-mbok-darmi", strangerID: "roti-lain"} {
		if _, err := queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
			ID: id, Type: "CONSIGNOR", Name: slug, Slug: slug, Whatsapp: "6281234567890", CityCode: "33.74",
			CreatedAt: base(),
		}); err != nil {
			t.Fatalf("organization: %v", err)
		}
	}
	return world{t: t, pool: pool, products: repository.NewProducts(pool)}
}

func productID(index int) string {
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index)
}

func change(index int, sku, name string) domain.Change {
	return domain.Change{
		ProductID: productID(index), ConsignorID: consignorID, SKU: sku,
		Details: domain.Details{
			Name: name, CategoryID: breadCategoryID, Unit: "pcs",
			Prices:    domain.Prices{Retail: money.Rupiah(2000), Deposit: money.Rupiah(1750)},
			ShelfLife: domain.ShelfLife{Given: true, Days: 3},
		},
		At: base().Add(time.Duration(index) * time.Minute),
	}
}

func (w world) save(changes ...domain.Change) {
	w.t.Helper()
	for _, item := range changes {
		if err := w.products.SaveProduct(w.t.Context(), item); err != nil {
			w.t.Fatalf("save %s: %v", item.SKU, err)
		}
	}
}

func TestSaveProduct_FRPRD01_StoredProductReadsBackWithItsCategory(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"))

	got, err := scene.products.GetProduct(t.Context(), ownKey(1))

	if err != nil || got.SKU != "ROT-0001" || got.CategoryName != "Roti dan kue basah" ||
		got.Details.ShelfLife.Days != 3 || got.Details.Prices.Deposit != 1750 || got.Status() != domain.Active {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestSaveProduct_FRPRD03_SecondSaveUpdatesTheSameProduct(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"))
	edited := change(1, "ROT-0001", "Roti Sobek Keju")
	edited.Details.ShelfLife = domain.ShelfLife{}
	scene.save(edited)

	got, err := scene.products.GetProduct(t.Context(), ownKey(1))

	if err != nil || got.Details.Name != "Roti Sobek Keju" || got.Details.ShelfLife.Given {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestSaveProduct_BRPRD01_SKUIsUniquePerConsignor(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"))
	elsewhere := change(3, "ROT-0001", "Roti Lain")
	elsewhere.ConsignorID = strangerID
	scene.save(elsewhere)

	err := scene.products.SaveProduct(t.Context(), change(2, "ROT-0001", "Roti Kedua"))

	if !errors.Is(err, domain.ErrSKUTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveProduct_UnknownCategoryAndForeignPhotoAreRejected(t *testing.T) {
	scene := newWorld(t)
	unknownCategory := change(1, "ROT-0001", "Roti Sobek")
	unknownCategory.Details.CategoryID = productID(99)
	foreignPhoto := change(2, "ROT-0002", "Roti Sobek")
	foreignPhoto.Details.PhotoFileID = productID(98)

	categoryErr := scene.products.SaveProduct(t.Context(), unknownCategory)
	photoErr := scene.products.SaveProduct(t.Context(), foreignPhoto)

	if !errors.Is(categoryErr, domain.ErrUnknownCategory) || !errors.Is(photoErr, domain.ErrUnknownPhoto) {
		t.Fatalf("category = %v, photo = %v", categoryErr, photoErr)
	}
}

func TestNextSKUNumber_BRPRD01_FollowsTheHighestNumberOfThePrefix(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti"), change(2, "ROT-0007", "Roti"), change(3, "RO-0099", "Ro"),
		change(4, "KER-0050", "Keripik"))

	next, err := scene.products.NextSKUNumber(t.Context(), domain.SKUSeries{ConsignorID: consignorID, Prefix: "ROT"})

	if err != nil || next != 8 {
		t.Fatalf("next = %d, err = %v", next, err)
	}
}

func TestNextSKUNumber_BRPRD01_ConcurrentAutoSKUsDoNotCollide(t *testing.T) {
	scene := newWorld(t)
	var group sync.WaitGroup
	failures := make(chan error, 2)
	for index := range 2 {
		group.Go(func() {
			failures <- scene.pool.WithinTx(t.Context(), func(ctx context.Context) error {
				series := domain.SKUSeries{ConsignorID: consignorID, Prefix: "ROT"}
				next, err := scene.products.NextSKUNumber(ctx, series)
				if err != nil {
					return fmt.Errorf("next: %w", err)
				}
				return scene.products.SaveProduct(ctx, change(index+1, domain.AutoSKU("ROT", next), "Roti"))
			})
		})
	}
	group.Wait()
	close(failures)

	for err := range failures {
		if err != nil {
			t.Fatalf("err = %v", err)
		}
	}
}

func TestGetProduct_RSEC38_ProductOfAnotherConsignorIsNotFound(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"))

	_, err := scene.products.GetProduct(t.Context(), domain.ProductKey{
		ConsignorID: strangerID, ProductID: productID(1),
	})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestSetArchived_FRPRD03_ArchivedProductLeavesTheActiveList(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"), change(2, "ROT-0002", "Roti Tawar"))

	err := scene.products.SetArchived(t.Context(), domain.StatusChange{
		ProductID: productID(1), ConsignorID: consignorID, IsArchived: true, At: base(),
	})
	active := scene.list(domain.Search{Status: domain.Active})
	archived := scene.list(domain.Search{Status: domain.Archived})

	if err != nil || len(active) != 1 || active[0].ID != productID(2) || len(archived) != 1 {
		t.Fatalf("active = %+v, archived = %+v, err = %v", active, archived, err)
	}
}

func (w world) list(search domain.Search) []domain.Product {
	w.t.Helper()
	found, err := w.products.ListProducts(w.t.Context(), domain.PageQuery{
		ConsignorID: consignorID, Search: search, Limit: pagination.MaxLimit,
	})
	if err != nil {
		w.t.Fatalf("list: %v", err)
	}
	return found
}

func TestListProducts_FRPRD05_SearchesNameAndSKUAndFiltersCategory(t *testing.T) {
	scene := newWorld(t)
	chips := change(3, "KER-0001", "Keripik Singkong")
	chips.Details.CategoryID = chipsCategoryID
	scene.save(change(1, "ROT-0001", "Roti Sobek"), change(2, "ROT-0002", "Roti Tawar"), chips)

	byName := scene.list(domain.Search{Text: "tawar", Status: domain.Active})
	bySKU := scene.list(domain.Search{Text: "ker-0", Status: domain.Active})
	byCategory := scene.list(domain.Search{CategoryID: chipsCategoryID, Status: domain.Active})

	if len(byName) != 1 || len(bySKU) != 1 || len(byCategory) != 1 || byCategory[0].ID != productID(3) {
		t.Fatalf("byName = %d, bySKU = %d, byCategory = %+v", len(byName), len(bySKU), byCategory)
	}
}

func TestListProducts_RSEC59_WildcardsInTheSearchAreLiteral(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti Sobek"), change(2, "DIS-0001", "Diskon 50%"))

	got := scene.list(domain.Search{Text: "%", Status: domain.Active})
	underscore := scene.list(domain.Search{Text: "_", Status: domain.Active})

	if len(got) != 1 || got[0].ID != productID(2) || len(underscore) != 0 {
		t.Fatalf("percent = %+v, underscore = %+v", got, underscore)
	}
}

func TestListProducts_RRES04_NewestFirstAcrossPages(t *testing.T) {
	scene := newWorld(t)
	scene.save(change(1, "ROT-0001", "Roti"), change(2, "ROT-0002", "Roti"), change(3, "ROT-0003", "Roti"))

	first, err := scene.products.ListProducts(t.Context(), domain.PageQuery{
		ConsignorID: consignorID, Search: domain.Search{Status: domain.Active}, Limit: 2,
	})
	if err != nil || len(first) != 2 || first[0].ID != productID(3) {
		t.Fatalf("first = %+v, err = %v", first, err)
	}
	second, err := scene.products.ListProducts(t.Context(), domain.PageQuery{
		ConsignorID: consignorID, Search: domain.Search{Status: domain.Active}, Limit: 2,
		After: pagination.Position{At: first[1].CreatedAt, ID: first[1].ID},
	})

	if err != nil || len(second) != 1 || second[0].ID != productID(1) {
		t.Fatalf("second = %+v, err = %v", second, err)
	}
}

func ownKey(index int) domain.ProductKey {
	return domain.ProductKey{ConsignorID: consignorID, ProductID: productID(index)}
}
