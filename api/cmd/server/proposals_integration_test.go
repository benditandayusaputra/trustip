package main

import (
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	sealdomain "github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	sealrepository "github.com/benditandayusaputra/trustip/api/internal/modules/seal/repository"
)

type invitationScene struct {
	consignor  *browser
	store      *browser
	productID  string
	inviteCode string
	proposalID string
}

func newInvitationScene(t *testing.T) invitationScene {
	t.Helper()
	return newInvitationSceneWith(t, "WEEKLY")
}

func newInvitationSceneWith(t *testing.T, schedule string) invitationScene {
	t.Helper()
	consignor := consignorBrowser(t)
	product := consignor.postJSON(productsPath, productBody(""))
	scene := invitationScene{consignor: consignor, productID: quotedAfter(string(product.body), `"id":"`)}
	sent := consignor.postJSON(invitationsPath, invitationBodyWith(scene.productID, schedule))
	if sent.status != http.StatusCreated {
		t.Fatalf("send = %d %s", sent.status, sent.body)
	}
	loadContract(t).requireConforms(t, sent)
	scene.inviteCode = quotedAfter(string(sent.body), `"invite_code":"`)
	scene.proposalID = quotedAfter(string(sent.body), `"id":"`)
	scene.store = consignor.anotherUserOnTheSameServer("tini@contoh.id")
	scene.store.found(onboarding(semarangCityCode), "LOGO")
	return scene
}

func TestInvitation_FRPRP07_StoreOpensAndClaimsTheConsignorInvitation(t *testing.T) {
	scene := newInvitationScene(t)
	contract := loadContract(t)

	viewed := scene.store.get(invitationsPath + "/" + scene.inviteCode)
	claimed := scene.store.post(invitationsPath + "/" + scene.inviteCode + "/claim")
	read := scene.store.get("/api/v1/proposals/" + scene.proposalID)

	for _, got := range []exchange{viewed, claimed, read} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(viewed.body), `"item_count":1`) ||
		!strings.Contains(string(read.body), `"per_unit":{"gross":2000,"store_share":250,"consignor_share":1750}`) ||
		!strings.Contains(string(read.body), `"number":"AJ-`) {
		t.Fatalf("viewed = %s\nread = %s", viewed.body, read.body)
	}
}

func TestInvitation_BRPRP07_ConsignorOpeningAConsignorInvitationIsToldItIsForStores(t *testing.T) {
	scene := newInvitationScene(t)
	other := scene.consignor.anotherUserOnTheSameServer("ani@contoh.id")
	other.found(strings.Replace(consignorOnboarding(), "Warung Bu Tini", "Penitip Lain", 1), "LOGO")

	got := other.get(invitationsPath + "/" + scene.inviteCode)

	requireCode(t, got, "INVITE_WRONG_ORG_TYPE")
	if !strings.Contains(string(got.body), "Undangan ini untuk toko.") {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func (s invitationScene) claimAsStore(t *testing.T) {
	t.Helper()
	if got := s.store.post(invitationsPath + "/" + s.inviteCode + "/claim"); got.status != http.StatusOK {
		t.Fatalf("claim = %d %s", got.status, got.body)
	}
}

func TestProposal_BRPRP03_AcceptingFormsASealedPartnershipAndTellsTheSender(t *testing.T) {
	scene := newInvitationScene(t)
	scene.claimAsStore(t)

	accepted := scene.store.post("/api/v1/proposals/" + scene.proposalID + "/accept")

	requireCode(t, accepted, "OK")
	loadContract(t).requireConforms(t, accepted)
	partnershipID := quotedAfter(string(accepted.body), `"partnership_id":"`)
	chain, err := sealrepository.NewSeals(scene.store.pool).ListChain(t.Context(), partnershipID)
	if err != nil || len(chain) != 1 || chain[0].Entry.Event != sealdomain.AgreementAccepted ||
		sealdomain.VerifyChain(chain) != nil {
		t.Fatalf("chain = %+v, err = %v", chain, err)
	}
	inbox := scene.consignor.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "menerima ajuan AJ-") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func TestProposal_BRPRP02_SenderCannotAcceptAndBothSeeTheListsWithTurns(t *testing.T) {
	scene := newInvitationScene(t)
	scene.claimAsStore(t)
	contract := loadContract(t)

	ownAccept := scene.consignor.post("/api/v1/proposals/" + scene.proposalID + "/accept")
	outgoing := scene.consignor.get("/api/v1/proposals?box=OUTGOING")
	incoming := scene.store.get("/api/v1/proposals?box=INCOMING")

	requireCode(t, ownAccept, "NOT_YOUR_TURN")
	for _, got := range []exchange{ownAccept, outgoing, incoming} {
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(outgoing.body), scene.proposalID) || !strings.Contains(string(incoming.body),
		scene.proposalID) {
		t.Fatalf("outgoing = %s\nincoming = %s", outgoing.body, incoming.body)
	}
}

func TestProposal_FRPRP03_DeclineWithAReasonEndsTheProposal(t *testing.T) {
	scene := newInvitationScene(t)
	scene.claimAsStore(t)

	declined := scene.store.postJSON("/api/v1/proposals/"+scene.proposalID+"/decline", `{"reason":"Rak penuh."}`)
	again := scene.store.post("/api/v1/proposals/" + scene.proposalID + "/accept")

	requireCode(t, declined, "OK")
	requireCode(t, again, "PROPOSAL_ALREADY_DECIDED")
	if !strings.Contains(string(declined.body), `"decision_reason":"Rak penuh."`) {
		t.Fatalf("declined = %s", declined.body)
	}
	loadContract(t).requireConforms(t, declined)
}

func TestProposal_RCC74_TwoDecisionsAtOnceOnlyOnePasses(t *testing.T) {
	scene := newInvitationScene(t)
	scene.claimAsStore(t)
	deps := buildTradeComponents(scene.store.pool, scene.store.now, testSealKeys{}).proposals.deps
	storeOwner := scene.store.actor(t)
	var group sync.WaitGroup
	outcomes := make(chan error, 2)
	group.Go(func() {
		_, err := proposalservice.NewAcceptProposal(deps).Execute(t.Context(), storeOwner, scene.proposalID)
		outcomes <- err
	})
	group.Go(func() {
		_, err := proposalservice.NewDeclineProposal(deps).Execute(t.Context(), storeOwner,
			proposalservice.DeclineCommand{ProposalID: scene.proposalID, Reason: "Batal."})
		outcomes <- err
	})
	group.Wait()
	close(outcomes)

	passed, refused := 0, 0
	for err := range outcomes {
		switch {
		case err == nil:
			passed++
		case apperror.From(err).Code() == apperror.CodeProposalDecided:
			refused++
		}
	}
	if passed != 1 || refused != 1 {
		t.Fatalf("passed = %d, refused = %d", passed, refused)
	}
}

func (b *browser) actor(t *testing.T) actor.Actor {
	t.Helper()
	account := string(b.getAccount().body)
	return actor.Actor{UserID: quotedAfter(account, `"id":"`), Membership: actor.Membership{
		OrganizationID:   quotedAfter(account, `"organization_id":"`),
		OrganizationType: actor.OrganizationType(quotedAfter(account, `"organization_type":"`)),
		Role:             actor.Role(quotedAfter(account, `"role":"`)),
	}}
}
