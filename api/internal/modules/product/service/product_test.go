package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/service"
)

const (
	consignorID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	breadCategoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"
)

type fakeProducts struct {
	stored      map[string]domain.Product
	nextNumber  int
	failWrite   error
	askedSeries []domain.SKUSeries
	asked       []domain.PageQuery
}

func newFakeProducts() *fakeProducts {
	return &fakeProducts{stored: map[string]domain.Product{}, nextNumber: 1}
}

func (f *fakeProducts) NextSKUNumber(_ context.Context, series domain.SKUSeries) (int, error) {
	f.askedSeries = append(f.askedSeries, series)
	return f.nextNumber, nil
}

func (f *fakeProducts) SaveProduct(_ context.Context, change domain.Change) error {
	if f.failWrite != nil {
		return f.failWrite
	}
	product, exists := f.stored[change.ProductID]
	if !exists {
		product = domain.Product{ID: change.ProductID, ConsignorID: change.ConsignorID, CreatedAt: change.At}
	}
	product.SKU, product.Details, product.UpdatedAt = change.SKU, change.Details, change.At
	f.stored[change.ProductID] = product
	return nil
}

func (f *fakeProducts) SetArchived(_ context.Context, change domain.StatusChange) error {
	product := f.stored[change.ProductID]
	product.IsArchived = change.IsArchived
	f.stored[change.ProductID] = product
	return nil
}

func (f *fakeProducts) GetProduct(_ context.Context, key domain.ProductKey) (domain.Product, error) {
	product, ok := f.stored[key.ProductID]
	if !ok || product.ConsignorID != key.ConsignorID {
		return domain.Product{}, apperror.NotFound()
	}
	return product, nil
}

func (f *fakeProducts) ListProducts(_ context.Context, query domain.PageQuery) ([]domain.Product, error) {
	f.asked = append(f.asked, query)
	products := []domain.Product{}
	for index := range query.Limit {
		products = append(products, domain.Product{
			ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index), CreatedAt: fixedClock{}.Now(),
		})
	}
	return products, nil
}

type fakeAudit struct{ entries []audit.Entry }

func (f *fakeAudit) Record(_ context.Context, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC) }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

type world struct {
	products *fakeProducts
	audit    *fakeAudit
	deps     service.ProductDeps
}

func newWorld() world {
	products, trail := newFakeProducts(), &fakeAudit{}
	return world{products: products, audit: trail, deps: service.ProductDeps{
		Products: products, Audit: trail, Tx: inlineTx{}, Clock: fixedClock{}, IDs: &sequenceIDs{},
	}}
}

func member(organizationType actor.OrganizationType, role actor.Role) actor.Actor {
	return actor.Actor{UserID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", Membership: actor.Membership{
		OrganizationID: consignorID, OrganizationType: organizationType, Role: role,
	}}
}

func consignorStaff() actor.Actor { return member(actor.Consignor, actor.Staff) }

func breadForm() domain.ProductForm {
	return domain.ProductForm{
		Name: "Roti Sobek", CategoryID: breadCategoryID, Unit: "pcs",
		Prices: domain.SuggestedPrices{Retail: 2000, Deposit: 1750},
	}
}

func TestCreateProduct_BRAUTHZ01_StoreCannotManageProducts(t *testing.T) {
	scene := newWorld()

	_, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), member(actor.Store, actor.Owner), breadForm())

	if apperror.From(err).Code() != apperror.CodeForbidden || len(scene.products.stored) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateProduct_InvalidFormSavesNothing(t *testing.T) {
	scene := newWorld()
	form := breadForm()
	form.Name = ""

	_, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), form)

	if apperror.From(err).Code() != apperror.CodeValidationFailed || len(scene.products.stored) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateProduct_BRPRD01_DuplicateSKUIsAFieldError(t *testing.T) {
	scene := newWorld()
	scene.products.failWrite = domain.ErrSKUTaken
	form := breadForm()
	form.SKU = "ROT-0001"

	_, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), form)

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "sku" || fields[0].Code != "TAKEN" {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateProduct_UnknownCategoryIsAFieldError(t *testing.T) {
	scene := newWorld()
	scene.products.failWrite = domain.ErrUnknownCategory

	_, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), breadForm())

	if fields := apperror.From(err).Fields(); len(fields) != 1 || fields[0].Field != "category_id" {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateProduct_FRPRD02_EmptySKUGetsTheNextNumberForTheNamePrefix(t *testing.T) {
	scene := newWorld()
	scene.products.nextNumber = 7

	product, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), breadForm())

	if err != nil || product.SKU != "ROT-0007" || scene.products.askedSeries[0] != (domain.SKUSeries{
		ConsignorID: consignorID, Prefix: "ROT",
	}) {
		t.Fatalf("product = %+v, err = %v", product, err)
	}
}

func TestCreateProduct_FRPRD01_StaffCreatesAProductAndItIsAudited(t *testing.T) {
	scene := newWorld()

	product, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), breadForm())

	if err != nil || product.ConsignorID != consignorID || len(scene.audit.entries) != 1 ||
		scene.audit.entries[0].Action != audit.ProductCreated || scene.audit.entries[0].EntityID != product.ID {
		t.Fatalf("product = %+v, audit = %+v, err = %v", product, scene.audit.entries, err)
	}
}

func createBread(t *testing.T, scene world) domain.Product {
	t.Helper()
	product, err := service.NewCreateProduct(scene.deps).Execute(t.Context(), consignorStaff(), breadForm())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return product
}

func TestUpdateProduct_FRPRD03_ChangesTheDetails(t *testing.T) {
	scene := newWorld()
	bread := createBread(t, scene)
	form := breadForm()
	form.SKU, form.Prices.Retail = bread.SKU, 2500

	updated, err := service.NewUpdateProduct(scene.deps).Execute(t.Context(), consignorStaff(),
		service.UpdateProductCommand{ProductID: bread.ID, Form: form})

	if err != nil || updated.Details.Prices.Retail.Int64() != 2500 || updated.SKU != bread.SKU {
		t.Fatalf("updated = %+v, err = %v", updated, err)
	}
}

func TestUpdateProduct_BRAUTHZ02_ProductOfAnotherConsignorIsNotFound(t *testing.T) {
	scene := newWorld()
	bread := createBread(t, scene)
	stranger := consignorStaff()
	stranger.Membership.OrganizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"

	_, err := service.NewUpdateProduct(scene.deps).Execute(t.Context(), stranger,
		service.UpdateProductCommand{ProductID: bread.ID, Form: breadForm()})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestArchiveProduct_FRPRD03_ArchiveThenReactivate(t *testing.T) {
	scene := newWorld()
	bread := createBread(t, scene)

	archived, err := service.NewArchiveProduct(scene.deps).Execute(t.Context(), consignorStaff(), bread.ID)
	if err != nil || archived.Status() != domain.Archived {
		t.Fatalf("archived = %+v, err = %v", archived, err)
	}
	active, err := service.NewReactivateProduct(scene.deps).Execute(t.Context(), consignorStaff(), bread.ID)

	if err != nil || active.Status() != domain.Active || scene.audit.entries[2].Action != audit.ProductReactivated {
		t.Fatalf("active = %+v, err = %v", active, err)
	}
}

func TestArchiveProduct_BRAUTHZ01_StoreIsForbidden(t *testing.T) {
	scene := newWorld()
	bread := createBread(t, scene)

	_, err := service.NewArchiveProduct(scene.deps).Execute(t.Context(), member(actor.Store, actor.Staff), bread.ID)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestGetProduct_FRPRD01_ConsignorReadsItsOwnProduct(t *testing.T) {
	scene := newWorld()
	bread := createBread(t, scene)

	got, err := service.NewGetProduct(scene.products).Execute(t.Context(), consignorStaff(), bread.ID)

	if err != nil || got.ID != bread.ID {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestListProducts_FRPRD05_SearchesWithinTheConsignorAndPages(t *testing.T) {
	scene := newWorld()

	page, err := service.NewListProducts(scene.products).Execute(t.Context(), consignorStaff(), service.ListRequest{
		Search: domain.SearchForm{Text: "roti", Status: "ARCHIVED"}, Limit: "2",
	})

	asked := scene.products.asked[0]
	if err != nil || len(page.Items) != 2 || !page.HasMore || asked.ConsignorID != consignorID ||
		asked.Limit != 3 || asked.Search.Status != domain.Archived || asked.Search.Text != "roti" {
		t.Fatalf("page = %+v, asked = %+v, err = %v", page, asked, err)
	}
}

func TestListProducts_UnknownStatusIsAValidationFailure(t *testing.T) {
	_, err := service.NewListProducts(newFakeProducts()).Execute(t.Context(), consignorStaff(),
		service.ListRequest{Search: domain.SearchForm{Status: "GONE"}})

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}
