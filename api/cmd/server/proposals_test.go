package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	agreementdomain "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	partnershipdomain "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

const invitationsPath = "/api/v1/invitations"

type emptyProposals struct{}

func (emptyProposals) ConsumeQuota(context.Context, domain.QuotaKey) error { return nil }

func (emptyProposals) GetDirectParties(context.Context, domain.DirectPair) (domain.DirectParties, error) {
	return domain.DirectParties{}, apperror.NotFound()
}

func (emptyProposals) NextNumber(context.Context, domain.SequenceKey) (int, error) { return 1, nil }

func (emptyProposals) Open(context.Context, domain.Opening) error {
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

func (emptyProposals) LockProposal(context.Context, domain.ProposalKey) (domain.Proposal, error) {
	return domain.Proposal{}, apperror.NotFound()
}

func (emptyProposals) SaveDecision(context.Context, domain.Decision) error {
	return domain.ErrAlreadyDecided
}

func (emptyProposals) ListProposals(context.Context, domain.ListQuery) ([]domain.Summary, error) {
	return []domain.Summary{}, nil
}

type refusingCollaborators struct{}

func (refusingCollaborators) Found(context.Context, domain.Founding) (string, error) {
	return "", domain.ErrAlreadyPartners
}

func (refusingCollaborators) Amend(context.Context, domain.Founding) (int, error) {
	return 0, domain.ErrAlreadyDecided
}

func (refusingCollaborators) CurrentAgreement(context.Context, domain.PartnershipKey) (domain.CurrentAgreement, error) {
	return domain.CurrentAgreement{}, apperror.NotFound()
}

func (refusingCollaborators) Announce(context.Context, domain.Notice) error { return nil }

type testSealKeys struct{}

func (testSealKeys) Active() (string, []byte) { return "v1", bytes.Repeat([]byte{9}, ed25519.SeedSize) }

func (testSealKeys) Lookup(id string) ([]byte, bool) {
	if id != "v1" {
		return nil, false
	}
	return bytes.Repeat([]byte{9}, ed25519.SeedSize), true
}

func (emptyProposals) HasActivePartnership(context.Context, domain.Pair) (bool, error) {
	return false, nil
}

func testProposalComponents(now *clock.Manual) proposalComponents {
	return proposalComponents{deps: proposalservice.ProposalDeps{
		Proposals: emptyProposals{}, Direct: emptyProposals{}, Audit: &memoryAudit{}, Tx: inlineTx{}, Clock: now,
		IDs:    idgen.NewSequence(),
		Codes:  inviteCodes(),
		Keeper: refusingCollaborators{}, Announcer: refusingCollaborators{}, Current: refusingCollaborators{},
	}}
}

func invitationBody(productID string) string {
	return invitationBodyWith(productID, "WEEKLY")
}

func invitationBodyWith(productID, schedule string) string {
	items := `[]`
	if productID != "" {
		items = `[{"product_id":"` + productID + `","pricing_model":"FIXED_DEPOSIT","retail_price":2000,` +
			`"commission_bps":0,"deposit_price":1750}]`
	}
	return `{"terms":{"settlement_schedule":"` + schedule + `","payment_terms_days":7,"count_rhythm":"DAILY",` +
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

type missingAgreements struct{}

func (missingAgreements) ListVersions(
	context.Context, agreementdomain.PartnershipKey,
) ([]agreementdomain.Agreement, error) {
	return nil, apperror.NotFound()
}

func (missingAgreements) GetAgreement(
	context.Context, agreementdomain.VersionKey,
) (agreementdomain.View, error) {
	return agreementdomain.View{}, apperror.NotFound()
}

func (missingAgreements) Check(context.Context, agreementdomain.SealCheck) (agreementdomain.SealVerdict, error) {
	return agreementdomain.SealVerdict{}, nil
}

func testAgreementComponents() agreementComponents {
	return agreementComponents{reader: missingAgreements{}, checker: missingAgreements{}}
}

type missingPartnerships struct{}

func (missingPartnerships) ListPartnerships(
	context.Context, partnershipdomain.ListQuery,
) ([]partnershipdomain.Summary, error) {
	return []partnershipdomain.Summary{}, nil
}

func (missingPartnerships) GetPartnership(
	context.Context, partnershipdomain.Key,
) (partnershipdomain.Partnership, error) {
	return partnershipdomain.Partnership{}, apperror.NotFound()
}

func (missingPartnerships) ListCatalog(
	context.Context, partnershipdomain.Key,
) ([]partnershipdomain.CatalogItem, error) {
	return []partnershipdomain.CatalogItem{}, nil
}
