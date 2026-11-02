package main

import (
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	notificationrepository "github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
)

func (b *browser) found(body, purpose string) organizationFixture {
	t := b.t
	t.Helper()
	founded := b.postOrganization(body)
	b.post("/api/v1/sessions/refresh")
	uploaded := b.upload(part{name: "purpose", data: []byte(purpose)},
		part{name: "file", filename: "image", data: photoWithLocation(t)})
	fixture := organizationFixture{
		owner: b, id: quotedAfter(string(founded.body), `"id":"`),
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
	world.consignor.notificationID = consignorOwner.receiveNotification()
	product := consignorOwner.postJSON(productsPath, productBody(""))
	world.consignor.productID = quotedAfter(string(product.body), `"id":"`)
	sent := consignorOwner.postJSON(invitationsPath, invitationBodyWith(world.consignor.productID, "ON_COUNT"))
	world.consignor.inviteCode = quotedAfter(string(sent.body), `"invite_code":"`)
	world.consignor.proposalID = quotedAfter(string(sent.body), `"id":"`)
	storeOwner.post(invitationsPath + "/" + world.consignor.inviteCode + "/claim")
	accepted := storeOwner.post("/api/v1/proposals/" + world.consignor.proposalID + "/accept")
	world.consignor.partnershipID = quotedAfter(string(accepted.body), `"partnership_id":"`)
	recorded := storeOwner.recordDrop(
		dropBody(world.consignor.partnershipID, world.consignor.productID, 20), dropKey)
	world.consignor.dropID = quotedAfter(string(recorded.body), `"id":"`)
	counted := storeOwner.recordCount(countBody(world.consignor.partnershipID, world.consignor.productID,
		countLine{recorded: 20, counted: 19}))
	world.consignor.countID = quotedAfter(string(counted.body), `"id":"`)
	world.consignor.settlementID = quotedAfter(string(counted.body), `"settlement":{"id":"`)
	if world.consignor.settlementID == "" {
		t.Fatalf("counted = %s", counted.body)
	}
	paid := storeOwner.recordPayment(world.consignor.settlementID,
		paymentBody(1750, "CASH", storeOwner.uploadProof()), paymentKey)
	world.consignor.paymentID = quotedAfter(string(paid.body), `"payments":[{"id":"`)
	posted := storeOwner.postJSON(needsPath, needBody("Butuh roti untuk sarapan"))
	world.store.needID = quotedAfter(string(posted.body), `"id":"`)
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

func (b *browser) receiveNotification() string {
	b.t.Helper()
	account := b.getAccount()
	notificationID := "0192f0c4-0000-7000-8000-000000000201"
	err := notificationrepository.NewInbox(b.pool).Deliver(b.t.Context(), []domain.Notification{{
		ID: notificationID, UserID: quotedAfter(string(account.body), `"id":"`), Type: "DROP_RECORDED",
		Title: "Titipan masuk", Body: "Warung mencatat titipan.", Link: "/app/titipan", CreatedAt: testNow(),
	}})
	if err != nil {
		b.t.Fatalf("deliver notification: %v", err)
	}
	return notificationID
}
