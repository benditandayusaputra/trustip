package main

import (
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
)

func (b *browser) markDemo(organizationID string) {
	b.t.Helper()
	err := sqlc.New(b.pool.Querier(b.t.Context())).SetOrganizationDemo(b.t.Context(),
		sqlc.SetOrganizationDemoParams{ID: organizationID, IsDemo: true})
	if err != nil {
		b.t.Fatalf("mark demo: %v", err)
	}
}

func TestDirectoryStores_DemoConsignorSeesOnlyDemoStores(t *testing.T) {
	consignor, store := listedStoreScene(t)
	consignor.markDemo(consignor.actor(t).Membership.OrganizationID)
	search := directoryStoresPath + "?origin=HERE&lat=-7.0456&lng=110.4386&radius_km=2"

	apart := consignor.get(search)
	store.owner.markDemo(store.id)
	together := consignor.get(search)

	requireCode(t, apart, "OK")
	if strings.Contains(string(apart.body), store.id) || !strings.Contains(string(together.body), store.id) {
		t.Fatalf("apart = %s\ntogether = %s", apart.body, together.body)
	}
}

func TestSendProposal_DemoConsignorCannotReachARealStore(t *testing.T) {
	consignor, store := listedStoreScene(t)
	consignor.markDemo(consignor.actor(t).Membership.OrganizationID)

	got := consignor.postJSON("/api/v1/proposals", directBody(store.id, ""))

	requireCode(t, got, "NOT_FOUND")
}

func TestClaimInvitation_RealStoreCannotClaimADemoInvitation(t *testing.T) {
	scene := newInvitationScene(t)
	scene.consignor.markDemo(scene.consignor.actor(t).Membership.OrganizationID)

	got := scene.store.post(invitationsPath + "/" + scene.inviteCode + "/claim")

	requireCode(t, got, "NOT_FOUND")
}
