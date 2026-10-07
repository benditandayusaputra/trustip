package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

const productsPath = "/api/v1/products"

type memoryProducts struct {
	mu       sync.Mutex
	products map[string]domain.Product
}

func (m *memoryProducts) NextSKUNumber(context.Context, domain.SKUSeries) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.products) + 1, nil
}

func (m *memoryProducts) SaveProduct(_ context.Context, change domain.Change) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.products[change.ProductID] = domain.Product{
		ID: change.ProductID, ConsignorID: change.ConsignorID, SKU: change.SKU, Details: change.Details,
		CategoryName: "Roti dan kue basah", CreatedAt: change.At, UpdatedAt: change.At,
	}
	return nil
}

func (m *memoryProducts) SetArchived(_ context.Context, change domain.StatusChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	product := m.products[change.ProductID]
	product.IsArchived = change.IsArchived
	m.products[change.ProductID] = product
	return nil
}

func (m *memoryProducts) SetSeeking(_ context.Context, change domain.SeekingChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	product := m.products[change.ProductID]
	product.IsSeekingStores = change.IsSeeking
	m.products[change.ProductID] = product
	return nil
}

func (m *memoryProducts) GetProduct(_ context.Context, key domain.ProductKey) (domain.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	product, ok := m.products[key.ProductID]
	if !ok || product.ConsignorID != key.ConsignorID {
		return domain.Product{}, apperror.NotFound()
	}
	return product, nil
}

func (m *memoryProducts) ListProducts(context.Context, domain.PageQuery) ([]domain.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	found := make([]domain.Product, 0, len(m.products))
	for _, product := range m.products {
		found = append(found, product)
	}
	return found, nil
}

func testProductComponents(now *clock.Manual) productComponents {
	return productComponents{
		products: &memoryProducts{products: map[string]domain.Product{}}, audit: &memoryAudit{}, tx: inlineTx{},
		clock: now, ids: idgen.NewSequence(),
	}
}

func (b *browser) postJSON(path, body string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

func productBody(sku string) string {
	return `{"sku":"` + sku + `","name":"Roti Sobek Cokelat","category_id":"` + breadCategoryID + `",` +
		`"unit":"pcs","description":"Lembut.\nTanpa pengawet.","composition":"Tepung, gula","net_content":"60 g",` +
		`"photo_file_id":null,"suggested_retail_price":2000,"suggested_deposit_price":1750,"shelf_life_days":3}`
}

func TestCreateProduct_FRPRD02_EmptySKUIsGeneratedFromTheName(t *testing.T) {
	visitor := memberBrowser(t, consignorOwner())

	got := visitor.postJSON(productsPath, productBody(""))

	if got.status != http.StatusCreated || !strings.Contains(string(got.body), `"sku":"ROT-0001"`) ||
		!strings.HasPrefix(got.header.Get(fiber.HeaderLocation), productsPath+"/") {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestCreateProduct_BRAUTHZ01_StoreOwnerIsForbidden(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.postJSON(productsPath, productBody(""))

	requireCode(t, got, "FORBIDDEN")
	loadContract(t).requireConforms(t, got)
}

func TestCreateProduct_RSEC47_EveryInvalidFieldIsReported(t *testing.T) {
	visitor := memberBrowser(t, consignorOwner())
	body := strings.NewReplacer(`"name":"Roti Sobek Cokelat"`, `"name":"R"`,
		`"suggested_deposit_price":1750`, `"suggested_deposit_price":2500`).Replace(productBody(""))

	got := visitor.postJSON(productsPath, body)

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"name"`) ||
		!strings.Contains(string(got.body), `"code":"EXCEEDS_RETAIL_PRICE"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestCreateProduct_RRES05_UnknownFieldIsRejected(t *testing.T) {
	visitor := memberBrowser(t, consignorOwner())

	got := visitor.postJSON(productsPath, strings.Replace(productBody(""), `{`, `{"consignor_org_id":"x",`, 1))

	requireCode(t, got, "UNKNOWN_FIELD")
}
