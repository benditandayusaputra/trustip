package main

import (
	"context"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

const invitationsPath = "/api/v1/invitations"

type emptyProposals struct{}

func (emptyProposals) NextNumber(context.Context, domain.SequenceKey) (int, error) { return 1, nil }

func (emptyProposals) CreateInvitation(context.Context, domain.Invitation) error {
	return domain.ErrUnknownProduct
}

func (emptyProposals) ClaimInvitation(context.Context, domain.Claim) error {
	return domain.ErrClaimRaced
}

func (emptyProposals) GetProposal(context.Context, domain.ProposalKey) (domain.Proposal, error) {
	return domain.Proposal{}, apperror.NotFound()
}

func (emptyProposals) GetInvitation(context.Context, string) (domain.InvitationView, error) {
	return domain.InvitationView{}, apperror.NotFound()
}

func (emptyProposals) HasActivePartnership(context.Context, domain.Pair) (bool, error) {
	return false, nil
}

func testProposalComponents(now *clock.Manual) proposalComponents {
	return proposalComponents{deps: proposalservice.ProposalDeps{
		Proposals: emptyProposals{}, Audit: &memoryAudit{}, Tx: inlineTx{}, Clock: now, IDs: idgen.NewSequence(),
		Codes: inviteCodes(),
	}}
}

func invitationBody(productID string) string {
	items := `[]`
	if productID != "" {
		items = `[{"product_id":"` + productID + `","pricing_model":"FIXED_DEPOSIT","retail_price":2000,` +
			`"commission_bps":0,"deposit_price":1750}]`
	}
	return `{"terms":{"settlement_schedule":"WEEKLY","payment_terms_days":7,"count_rhythm":"DAILY",` +
		`"notes":"Rak depan kasir."},"items":` + items + `,"message":"Salam kenal, Bu."}`
}

func TestSendInvitation_BRAUTHZ01_StaffCannotSendProposals(t *testing.T) {
	staff := consignorOwner()
	staff.Role = actor.Staff
	visitor := memberBrowser(t, staff)

	got := visitor.postJSON(invitationsPath, invitationBody(""))

	requireCode(t, got, "FORBIDDEN")
	loadContract(t).requireConforms(t, got)
}

func TestSendInvitation_RSEC47_InvalidTermsNameTheirFields(t *testing.T) {
	visitor := memberBrowser(t, consignorOwner())
	body := `{"terms":{"settlement_schedule":"DAILY","payment_terms_days":31,"count_rhythm":"DAILY","notes":""},` +
		`"items":[{"product_id":"x","pricing_model":"COMMISSION","retail_price":2000,"commission_bps":5001,` +
		`"deposit_price":0}],"message":""}`

	got := visitor.postJSON(invitationsPath, body)

	requireCode(t, got, "VALIDATION_FAILED")
	for _, field := range []string{
		"terms.settlement_schedule", "terms.payment_terms_days", "items[0].product_id",
		"items[0].commission_bps",
	} {
		if !containsField(got, field) {
			t.Errorf("missing %s in %s", field, got.body)
		}
	}
	loadContract(t).requireConforms(t, got)
}

func containsField(got exchange, field string) bool {
	return strings.Contains(string(got.body), `"field":"`+field+`"`)
}

func TestViewInvitation_MalformedCodeIsNotFound(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.get(invitationsPath + "/bukan-kode")

	requireCode(t, got, "NOT_FOUND")
	loadContract(t).requireConforms(t, got)
}
