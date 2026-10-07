package main

import (
	"strings"
	"testing"
)

func (owner *browser) found(body, purpose string) organizationFixture {
	t := owner.t
	t.Helper()
	founded := owner.postOrganization(body)
	owner.post("/api/v1/sessions/refresh")
	uploaded := owner.upload(part{name: "purpose", data: []byte(purpose)},
		part{name: "file", filename: "image", data: photoWithLocation(t)})
	fixture := organizationFixture{
		owner: owner, id: quotedAfter(string(founded.body), `"id":"`),
		fileID: quotedAfter(string(uploaded.body), `"id":"`),
	}
	if fixture.id == "" || fixture.fileID == "" {
		t.Fatalf("founded = %s, uploaded = %s", founded.body, uploaded.body)
	}
	return fixture
}

func newIsolationWorld(t *testing.T) isolationWorld {
	t.Helper()
	storeOwner := databaseBrowser(t)
	world := isolationWorld{store: storeOwner.found(onboarding(semarangCityCode), "LOGO")}
	consignorOwner := storeOwner.anotherUserOnTheSameServer("darmi@contoh.id")
	world.consignor = consignorOwner.found(consignorOnboarding(), "PAYMENT_QR")
	world.storeStranger = storeOwner.anotherUserOnTheSameServer("tini@contoh.id")
	world.storeStranger.found(strings.Replace(onboarding(semarangCityCode), "Warung Bu Tini", "Toko Lain", 1), "LOGO")
	world.consignorStranger = storeOwner.anotherUserOnTheSameServer("ani@contoh.id")
	world.consignorStranger.found(strings.Replace(consignorOnboarding(), "Warung Bu Tini", "Penitip Lain", 1), "LOGO")
	return world
}

func TestIsolation_RSEC41_OtherOrganizationsGetNotFoundOnEveryRouteWithAnID(t *testing.T) {
	world := newIsolationWorld(t)
	contract := loadContract(t)

	for route, request := range isolationCases() {
		t.Run(route, func(t *testing.T) {
			got := request(world)

			requireCode(t, got, "NOT_FOUND")
			contract.requireConforms(t, got)
		})
	}
}
