package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func directParams() DirectParams {
	return DirectParams{
		IDs: InvitationIDs{ProposalID: "p9", RevisionID: "r9"}, Sender: consignorSender(), SenderName: "Roti Pak Anton",
		Target:   Target{OrganizationID: storeID, Name: "Warung Bu Sri", Type: actor.Store},
		Sequence: 4, At: sentAt(),
	}
}

func TestPropose_FRPRP01_ConsignorProposesStraightToAStore(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())

	opening, err := draft.Propose(directParams())

	if err != nil || opening.Kind != New || opening.ConsignorID != consignorID || opening.StoreID != storeID ||
		opening.AwaitingID != storeID || opening.Number != "AJ-202610-0004" || opening.IDs.Code != "" {
		t.Fatalf("opening = %+v, err = %v", opening, err)
	}
}

func TestPropose_BRPRP01_StoreProposalKeepsTheConsignorAsProductOwner(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())
	params := directParams()
	params.Sender = Sender{UserID: "u2", OrganizationID: storeID, OrganizationType: actor.Store}
	params.Target = Target{OrganizationID: consignorID, Name: "Roti Pak Anton", Type: actor.Consignor}

	opening, err := draft.Propose(params)

	if err != nil || opening.ConsignorID != consignorID || opening.StoreID != storeID ||
		opening.AwaitingID != consignorID {
		t.Fatalf("opening = %+v, err = %v", opening, err)
	}
}

func TestPropose_BRPRP01_TargetMustBeTheOppositeType(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())
	sameType := directParams()
	sameType.Target = Target{OrganizationID: "c2", Type: actor.Consignor}

	if fieldCodes(t, mustFail(draft.Propose(sameType)))["target_organization_id"] != "UNKNOWN_CHOICE" {
		t.Fatal("same type accepted")
	}
}

func TestPropose_BRPRP06_SevenDaysAfterADeclineTheSenderMayTryAgain(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())
	tooSoon := directParams()
	tooSoon.LastDeclinedAt = sentAt().Add(-DeclineCooldown + time.Minute)
	afterwards := directParams()
	afterwards.LastDeclinedAt = sentAt().Add(-DeclineCooldown)

	_, soonErr := draft.Propose(tooSoon)
	_, laterErr := draft.Propose(afterwards)

	appErr := apperror.From(soonErr)
	if appErr.Code() != apperror.CodeProposalCooldown || laterErr != nil ||
		appErr.Message() != "Warung Bu Sri menolak ajuan Anda. Ajuan baru bisa dikirim lagi mulai 19 Oktober 2026." {
		t.Fatalf("soon = %v (%q), later = %v", soonErr, appErr.Message(), laterErr)
	}
}

func TestPropose_FRPRP02_DirectProposalOnlyAddsProducts(t *testing.T) {
	draft := draftOf(t, ItemInput{ProductID: breadID, Action: "STOP"})

	if fieldCodes(t, mustFail(draft.Propose(directParams())))["items[0].action"] != "UNKNOWN_CHOICE" {
		t.Fatal("stop accepted")
	}
}

func TestReceivedNotice_FRNTF01_TargetOwnersHearAboutTheProposal(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())
	opening, _ := draft.Propose(directParams())

	notice := opening.ReceivedNotice()

	if notice.OrganizationID != storeID || notice.Type != "PROPOSAL_RECEIVED" || !notice.OwnersOnly ||
		notice.Body != "Roti Pak Anton mengajukan kerja sama." || notice.Link != "/app/ajuan/p9" {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestQuotaKeyOf_BRPRP06_CountsPerJakartaDay(t *testing.T) {
	key := QuotaKeyOf(consignorID, time.Date(2026, 10, 19, 17, 30, 0, 0, time.UTC))

	if key.Day != "2026-10-20" || key.OrganizationID != consignorID {
		t.Fatalf("key = %+v", key)
	}
}
