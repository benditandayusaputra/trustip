package domain

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000501"
	jamID         = "0192f0c4-0000-7000-8000-000000000103"
)

func currentAgreement() CurrentAgreement {
	breadTerms, _ := pricing.NewTerms(depositTerms(2000, 1750))
	chipsTerms, _ := pricing.NewTerms(commissionTerms(15000, 2000))
	return CurrentAgreement{
		PartnershipID: partnershipID, IsActive: true,
		Consignor: Party{OrganizationID: consignorID, Name: "Roti Pak Anton"},
		Store:     Party{OrganizationID: storeID, Name: "Warung Bu Sri"},
		Terms:     terms.General{Schedule: terms.ScheduleWeekly, PaymentTermsDays: 7, Rhythm: terms.RhythmDaily},
		Items: map[string]CurrentItem{
			breadID: {IsActive: true, Terms: breadTerms},
			chipsID: {IsActive: false, Terms: chipsTerms},
		},
	}
}

func draftOf(t *testing.T, items ...ItemInput) RevisionDraft {
	t.Helper()
	draft, err := NewRevisionDraft(RevisionInput{
		Terms: terms.Input{SettlementSchedule: "WEEKLY", PaymentTermsDays: 7, CountRhythm: "DAILY"}, Items: items,
	})
	if err != nil {
		t.Fatalf("draft: %v", err)
	}
	return draft
}

func amend(draft RevisionDraft, current CurrentAgreement) (Opening, error) {
	return draft.Amend(AmendmentParams{
		IDs: InvitationIDs{ProposalID: "p2", RevisionID: "r2"}, Current: current, Sequence: 3, At: sentAt(),
		Sender: Sender{UserID: "u", OrganizationID: consignorID, OrganizationType: actor.Consignor},
	})
}

func TestAmend_FRPRP08_AddUpdateAndStopGoToTheStore(t *testing.T) {
	draft := draftOf(t,
		ItemInput{ProductID: jamID, Action: "ADD", Terms: depositTerms(35000, 28000)},
		ItemInput{ProductID: breadID, Action: "STOP"},
		ItemInput{ProductID: chipsID, Action: "ADD", Terms: commissionTerms(15000, 1500)},
	)

	opening, err := amend(draft, currentAgreement())

	if err != nil || opening.Kind != Amendment || opening.AwaitingID != storeID ||
		opening.PartnershipID != partnershipID ||
		opening.Draft.Items[1].Terms.DepositPrice != 1750 || opening.Number != "AJ-202610-0003" {
		t.Fatalf("opening = %+v, err = %v", opening, err)
	}
}

func TestAmend_FRAGR06_UpdatingAProductOutsideTheAgreementIsRefused(t *testing.T) {
	draft := draftOf(t,
		ItemInput{ProductID: breadID, Action: "ADD", Terms: depositTerms(2000, 1700)},
		ItemInput{ProductID: jamID, Action: "UPDATE", Terms: depositTerms(35000, 28000)},
	)

	codes := fieldCodes(t, mustFail(amend(draft, currentAgreement())))

	if codes["items[0].product_id"] != "UNKNOWN_CHOICE" || codes["items[1].product_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestAmend_NothingChangedIsRefused(t *testing.T) {
	codes := fieldCodes(t, mustFail(amend(draftOf(t), currentAgreement())))

	if codes["items"] != "REQUIRED" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestAmend_EndedPartnershipCannotBeAmended(t *testing.T) {
	current := currentAgreement()
	current.IsActive = false

	_, err := amend(draftOf(t, ItemInput{ProductID: breadID, Action: "STOP"}), current)

	if apperror.From(err).Code() != apperror.CodePartnershipNotActive {
		t.Fatalf("err = %v", err)
	}
}

func TestInvite_FRPRP07_InvitationOnlyAddsProducts(t *testing.T) {
	draft := draftOf(t, ItemInput{ProductID: breadID, Action: "STOP"})

	_, err := draft.Invite(InvitationParams{IDs: InvitationIDs{Code: "ABCDEFGH23"}, Sender: consignorSender()})

	if fieldCodes(t, err)["items[0].action"] != "UNKNOWN_CHOICE" {
		t.Fatalf("err = %v", err)
	}
}

func TestAgreementNotices_FRNTF01_BothOwnersHearAboutTheNewVersion(t *testing.T) {
	proposal := claimedProposal()
	proposal.Kind, proposal.PartnershipID = Amendment, partnershipID

	notices := proposal.AgreementNotices(Decision{DecidedBy: storeUser}, 2)

	if len(notices) != 2 || notices[0].Body != "Kesepakatan dengan Warung Bu Sri kini versi 2." ||
		notices[1].OrganizationID != storeID || notices[1].Link != "/app/mitra/"+partnershipID+"/kesepakatan" {
		t.Fatalf("notices = %+v", notices)
	}
}
