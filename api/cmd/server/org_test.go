package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type memoryFounder struct {
	mu        sync.Mutex
	foundings []domain.Founding
}

func (m *memoryFounder) TakenSlugs(context.Context, domain.Slug) ([]domain.Slug, error) {
	return []domain.Slug{}, nil
}

func (m *memoryFounder) CreateOrganization(_ context.Context, founding domain.Founding) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if founding.Organization.CityCode != semarangCityCode {
		return "", domain.ErrUnknownCity
	}
	m.foundings = append(m.foundings, founding)
	return "Kota Semarang", nil
}

type memoryRegions struct{}

func (memoryRegions) ListProvinces(context.Context) ([]domain.Province, error) {
	return []domain.Province{{Code: "33", Name: "Jawa Tengah"}}, nil
}

func (memoryRegions) ListCities(_ context.Context, provinceCode string) ([]domain.City, error) {
	if provinceCode != "33" {
		return []domain.City{}, nil
	}
	return []domain.City{{Code: semarangCityCode, Name: "Kota Semarang", ProvinceName: "Jawa Tengah"}}, nil
}

const (
	semarangCityCode  = "33.74"
	organizationsPath = "/api/v1/organizations"
)

func testOrgComponents(now *clock.Manual) orgComponents {
	return orgComponents{
		founder: &memoryFounder{}, regions: memoryRegions{}, tx: inlineTx{},
		clock: now, ids: idgen.NewSequence(),
	}
}

func onboarding(cityCode string) string {
	return `{"type":"STORE","name":"Warung Bu Tini","description":"Dekat SD.\nBuka pagi.",` +
		`"whatsapp":"0812-3456-7890","city_code":"` + cityCode + `",` +
		`"district":"Tembalang","address":"Jl. Sirojudin 12"}`
}

func (b *browser) postOrganization(body string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, organizationsPath, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: "198.51.100.200"})
	return b.send(req)
}

func (b *browser) get(path string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodGet, path, http.NoBody)
	b.attachCookies(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: "198.51.100.201"})
	return b.send(req)
}

func TestCreateOrganization_FRAUTH04_NewcomerFoundsAStore(t *testing.T) {
	visitor := signedInBrowser(t)

	got := visitor.postOrganization(onboarding(semarangCityCode))

	if got.status != http.StatusCreated || !strings.HasPrefix(got.header.Get("Location"), organizationsPath+"/") {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
	expected := []string{`"slug":"warung-bu-tini"`, `"whatsapp":"6281234567890"`, `"city_name":"Kota Semarang"`}
	for _, field := range expected {
		if !strings.Contains(string(got.body), field) {
			t.Errorf("body lacks %s: %s", field, got.body)
		}
	}
	loadContract(t).requireConforms(t, got)
}

func TestCreateOrganization_UnknownCityIsAValidationFailure(t *testing.T) {
	visitor := signedInBrowser(t)

	got := visitor.postOrganization(onboarding("99.99"))

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"city_code"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestCreateOrganization_RSEC51_HiddenCharactersInANameAreRejected(t *testing.T) {
	visitor := signedInBrowser(t)
	body := strings.Replace(onboarding(semarangCityCode), "Warung Bu Tini", `Warung\u200bBu Tini`, 1)

	got := visitor.postOrganization(body)

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"code":"INVALID_CHARACTERS"`) {
		t.Fatalf("body = %s", got.body)
	}
}

func TestCreateOrganization_BRAUTHZ05_OrganizationIDInTheBodyIsAnUnknownField(t *testing.T) {
	visitor := signedInBrowser(t)
	smuggled := `{"organization_id":"0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c",`
	body := strings.Replace(onboarding(semarangCityCode), `{`, smuggled, 1)

	got := visitor.postOrganization(body)

	requireCode(t, got, "UNKNOWN_FIELD")
}

func TestListProvinces_IsPublicAndFollowsTheContract(t *testing.T) {
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}

	got := visitor.get("/api/v1/provinces")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
}

func TestListCities_ListsOneProvinceAndRejectsAMalformedCode(t *testing.T) {
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}

	found := visitor.get("/api/v1/cities?province_code=33")
	malformed := visitor.get("/api/v1/cities?province_code=33.74")

	requireCode(t, found, "OK")
	if !strings.Contains(string(found.body), `"name":"Kota Semarang"`) {
		t.Fatalf("body = %s", found.body)
	}
	requireCode(t, malformed, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, found)
	loadContract(t).requireConforms(t, malformed)
}
