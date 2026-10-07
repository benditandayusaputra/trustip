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

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
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
	return []domain.City{{
		Code: semarangCityCode, Name: "Kota Semarang", ProvinceName: "Jawa Tengah",
		Centroid: domain.Coordinates{Latitude: -6.99, Longitude: 110.42},
	}}, nil
}

const (
	semarangCityCode  = "33.74"
	organizationsPath = "/api/v1/organizations"
)

type memoryProfiles struct {
	mu       sync.Mutex
	settings map[string]domain.Settings
}

func (m *memoryProfiles) GetSettings(_ context.Context, organizationID string) (domain.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings, ok := m.settings[organizationID]
	if !ok {
		return domain.Settings{}, apperror.NotFound()
	}
	return settings, nil
}

func (m *memoryProfiles) UpdateProfile(_ context.Context, change domain.ProfileChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[change.OrganizationID]
	settings.Profile = change.Update.Profile
	m.settings[change.OrganizationID] = settings
	return nil
}

func (m *memoryProfiles) UpdateStoreProfile(_ context.Context, change domain.StoreProfileChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[change.OrganizationID]
	settings.Store.Profile = change.Profile
	m.settings[change.OrganizationID] = settings
	return nil
}

func (m *memoryProfiles) UpdateConsignorProfile(_ context.Context, change domain.ConsignorProfileChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[change.OrganizationID]
	settings.Type = actor.Consignor
	settings.Consignor = change.Profile
	m.settings[change.OrganizationID] = settings
	return nil
}

func (m *memoryProfiles) UpdatePaymentInfo(_ context.Context, change domain.PaymentInfoChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[change.OrganizationID]
	settings.Payment.Info = change.Payment.Info
	m.settings[change.OrganizationID] = settings
	return nil
}

func (m *memoryProfiles) GetListingFacts(_ context.Context, organizationID string) (domain.ListingFacts, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[organizationID]
	return domain.ListingFacts{
		Type: settings.Type, HasName: settings.Profile.Name != "", HasCity: settings.Profile.CityCode != "",
		HasLocation: settings.Profile.HasLocation,
	}, nil
}

func (m *memoryProfiles) SetListed(_ context.Context, change domain.ListingChange) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings := m.settings[change.OrganizationID]
	settings.Listing.IsListed = change.IsListed
	m.settings[change.OrganizationID] = settings
	return nil
}

func (memoryRegions) ListCategories(context.Context) ([]domain.Category, error) {
	return []domain.Category{{ID: breadCategoryID, Name: "Roti dan kue basah", Slug: "roti-dan-kue-basah"}}, nil
}

const breadCategoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"

func testOrgComponents(now *clock.Manual) orgComponents {
	store := domain.Settings{
		ID: memberOrganizationID, Type: actor.Store, Slug: "warung-bu-tini", CityName: "Kota Semarang",
	}
	profiles := &memoryProfiles{settings: map[string]domain.Settings{memberOrganizationID: store}}
	return orgComponents{
		founder: &memoryFounder{}, regions: memoryRegions{}, categories: memoryRegions{}, tx: inlineTx{},
		clock: now, ids: idgen.NewSequence(), profiles: profiles, sections: profiles, listings: profiles,
		audit: &memoryAudit{},
	}
}

func (b *browser) patch(path, body string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPatch, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

func profileUpdate(extra string) string {
	return `{` + extra + `"name":"Warung Bu Tini Jaya","description":"","whatsapp":"0812 3456 7890",` +
		`"city_code":"33.74","district":"Tembalang","address":"","location":null,"logo_file_id":null}`
}

func TestUpdateProfile_FRORG01_OwnerSavesTheBasicProfile(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID, profileUpdate(""))

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"name":"Warung Bu Tini Jaya"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestUpdateProfile_BRORG01_TypeCannotBeChanged(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID, profileUpdate(`"type":"CONSIGNOR",`))

	requireCode(t, got, "UNKNOWN_FIELD")
	loadContract(t).requireConforms(t, got)
}

func TestReadSettings_BRAUTHZ02_AnotherOrganizationIsNotFound(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.get(organizationsPath + "/0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a99")

	requireCode(t, got, "NOT_FOUND")
	loadContract(t).requireConforms(t, got)
}

func TestReadSettings_MemberReadsItsOwnOrganization(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.get(organizationsPath + "/" + memberOrganizationID)

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
}

func onboarding(cityCode string) string {
	return `{"type":"STORE","name":"Warung Bu Tini","description":"Dekat SD.\nBuka pagi.",` +
		`"whatsapp":"0812-3456-7890","city_code":"` + cityCode + `",` +
		`"district":"Tembalang","address":"Jl. Sirojudin 12","location":null}`
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
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
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

func TestCreateOrganization_FRORG03_PinnedLocationIsSavedAndEchoedToTheOwner(t *testing.T) {
	visitor := signedInBrowser(t)
	pinned := strings.Replace(onboarding(semarangCityCode), `"location":null`,
		`"location":{"latitude":-7.0546,"longitude":110.4386}`, 1)

	got := visitor.postOrganization(pinned)

	if got.status != http.StatusCreated || !strings.Contains(string(got.body), `"location":{"latitude":-7.0546`) {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestCreateOrganization_FRORG03_LocationOutOfRangeIsAFieldError(t *testing.T) {
	visitor := signedInBrowser(t)
	wrong := strings.Replace(onboarding(semarangCityCode), `"location":null`,
		`"location":{"latitude":-97,"longitude":110.4386}`, 1)

	got := visitor.postOrganization(wrong)

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"location","code":"OUT_OF_RANGE"`) {
		t.Fatalf("body = %s", got.body)
	}
}

func storeProfileBody(photos string) string {
	return `{"store_type":"BAKERY","accepted_category_ids":["` + breadCategoryID + `"],` +
		`"accepts_fixed_deposit":false,"accepts_commission":true,"commission_min_bps":1000,` +
		`"commission_max_bps":2000,"requires_distribution_permit":true,"requires_halal":false,` +
		`"requires_sample":false,"other_requirements":"","opening_hours":"07.00-21.00","photo_file_ids":[` +
		photos + `]}`
}

func TestUpdateStoreProfile_FRORG04_StoreSavesItsProfile(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/store-profile", storeProfileBody(""))

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"store_type":"BAKERY"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestUpdateStoreProfile_ConsignorGetsNotFound(t *testing.T) {
	visitor := memberBrowser(t, actor.Membership{
		OrganizationID: memberOrganizationID, OrganizationType: actor.Consignor, Role: actor.Owner,
	})

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/store-profile", storeProfileBody(""))

	requireCode(t, got, "NOT_FOUND")
}

func TestListCategories_IsPublicAndFollowsTheContract(t *testing.T) {
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}

	got := visitor.get("/api/v1/categories")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
}

func consignorOwner() actor.Membership {
	return actor.Membership{OrganizationID: memberOrganizationID, OrganizationType: actor.Consignor, Role: actor.Owner}
}

const consignorProfileBody = `{"delivery_radius_km":15,"delivery_days":["MON","THU"],` +
	`"preferred_store_types":["CAFE"],"business_id_no":"1234567890123","distribution_permit_no":"",` +
	`"halal_cert_no":"","can_provide_sample":true}`

func TestUpdateConsignorProfile_FRORG05_ConsignorSavesItsProfile(t *testing.T) {
	visitor := memberBrowser(t, consignorOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/consignor-profile", consignorProfileBody)

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"delivery_radius_km":15`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestUpdatePaymentInfo_FRORG07_StoreGetsNotFound(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/payment-info",
		`{"payment_info":"BCA 123","payment_qr_file_id":null}`)

	requireCode(t, got, "NOT_FOUND")
}

func TestSetListing_BRORG05_IncompleteStoreIsRefusedWithTheMissingList(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/listing", `{"is_listed":true}`)

	requireCode(t, got, "PROFILE_INCOMPLETE")
	if !strings.Contains(string(got.body), "Lengkapi profil dulu: nama usaha, kota, lokasi") {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestSetListing_FRORG06_HidingIsAlwaysAllowed(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.patch(organizationsPath+"/"+memberOrganizationID+"/listing", `{"is_listed":false}`)

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"listing":{"is_listed":false,"percent":`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}
