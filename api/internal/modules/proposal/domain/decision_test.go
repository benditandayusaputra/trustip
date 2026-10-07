package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	consignorUser = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d01"
	storeUser     = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d02"
)

func claimedProposal() Proposal {
	general := terms.General{Schedule: terms.ScheduleWeekly, PaymentTermsDays: 7, Rhythm: terms.RhythmDaily}
	itemTerms, _ := pricing.NewTerms(pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750})
	return Proposal{
		ID: "p1", Number: "AJ-202610-0001", Kind: New, Status: Open,
		Consignor: Party{OrganizationID: consignorID, Name: "Roti Pak Anton"},
		Store:     Party{OrganizationID: storeID, Name: "Warung Bu Sri"},
		FromOrgID: consignorID, AwaitingOrgID: storeID, ExpiresAt: sentAt().Add(OpenFor),
		Revision: Revision{
			No: 1, AuthorOrgID: consignorID, AuthorUserID: consignorUser, Terms: general,
			Items: []Item{{ProductID: breadID, Action: Add, Terms: itemTerms}},
		},
	}
}

func storeDecider() Decider {
	return Decider{UserID: storeUser, OrganizationID: storeID}
}

func TestAccept_BRPRP02_AwaitingPartyAccepts(t *testing.T) {
	decision, _, err := claimedProposal().Accept(storeDecider(), sentAt())

	if err != nil || decision.Status != Accepted || decision.DecidedBy != storeUser || decision.RevisionNo != 1 {
		t.Fatalf("decision = %+v, err = %v", decision, err)
	}
}

func TestAccept_BRPRP03_AcceptanceCarriesWhatTheAgreementNeeds(t *testing.T) {
	decidedAt := sentAt().Add(time.Hour)

	_, founding, _ := claimedProposal().Accept(storeDecider(), decidedAt)

	if founding.ConsignorID != consignorID || founding.StoreID != storeID || founding.ProposedBy != consignorUser ||
		founding.AcceptedBy != storeUser || len(founding.Items) != 1 || !founding.At.Equal(decidedAt) {
		t.Fatalf("founding = %+v", founding)
	}
}

func TestAccept_BRPRP02_AuthorOfTheLastRevisionCannotAccept(t *testing.T) {
	_, _, err := claimedProposal().Accept(Decider{UserID: consignorUser, OrganizationID: consignorID}, sentAt())

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeNotYourTurn || appErr.Message() != "Menunggu tanggapan dari Warung Bu Sri." {
		t.Fatalf("err = %v, message = %q", err, appErr.Message())
	}
}

func TestAccept_UnclaimedInvitationWaitsForTheInvitee(t *testing.T) {
	unclaimed := claimedProposal()
	unclaimed.Store, unclaimed.AwaitingOrgID = Party{}, ""

	_, _, err := unclaimed.Accept(Decider{UserID: consignorUser, OrganizationID: consignorID}, sentAt())

	if apperror.From(err).Message() != "Menunggu tanggapan dari mitra yang diundang." {
		t.Fatalf("err = %v", err)
	}
}

func TestAccept_DecidedProposalIsAlreadyDecided(t *testing.T) {
	declined := claimedProposal()
	declined.Status = Declined

	_, _, err := declined.Accept(storeDecider(), sentAt())

	if apperror.From(err).Code() != apperror.CodeProposalDecided {
		t.Fatalf("err = %v", err)
	}
}

func TestAccept_BRPRP08_ExpiredProposalCannotBeAccepted(t *testing.T) {
	proposal := claimedProposal()

	_, _, err := proposal.Accept(storeDecider(), proposal.ExpiresAt)

	if apperror.From(err).Code() != apperror.CodeProposalExpired {
		t.Fatalf("err = %v", err)
	}
}

func TestDecline_FRPRP03_ReceiverDeclinesWithAReason(t *testing.T) {
	decision, err := claimedProposal().Decline(storeDecider(), "Rak sudah penuh.", sentAt())

	if err != nil || decision.Status != Declined || decision.Reason != "Rak sudah penuh." {
		t.Fatalf("decision = %+v, err = %v", decision, err)
	}
}

func TestDecline_FRPRP03_ReasonIsRequiredAndLimited(t *testing.T) {
	_, empty := claimedProposal().Decline(storeDecider(), "", sentAt())
	_, long := claimedProposal().Decline(storeDecider(), strings.Repeat("a", MaxReasonLength+1), sentAt())

	if fieldCodes(t, empty)["reason"] != "REQUIRED" || fieldCodes(t, long)["reason"] != "TOO_LONG" {
		t.Fatalf("empty = %v, long = %v", empty, long)
	}
}

func TestWithdraw_FRPRP05_OnlyTheAuthorOfTheLastRevision(t *testing.T) {
	decision, err := claimedProposal().Withdraw(Decider{UserID: consignorUser, OrganizationID: consignorID}, sentAt())
	_, receiverErr := claimedProposal().Withdraw(storeDecider(), sentAt())

	if err != nil || decision.Status != Withdrawn || apperror.From(receiverErr).Code() != apperror.CodeForbidden {
		t.Fatalf("decision = %+v, err = %v, receiver = %v", decision, err, receiverErr)
	}
}

func TestNotice_FRNTF01_AcceptanceTellsTheAuthorsOwners(t *testing.T) {
	proposal := claimedProposal()
	decision, _, _ := proposal.Accept(storeDecider(), sentAt())

	notice := proposal.NoticeOf(decision)

	if notice.OrganizationID != consignorID || notice.Type != "PROPOSAL_ACCEPTED" || !notice.OwnersOnly ||
		notice.Body != "Warung Bu Sri menerima ajuan AJ-202610-0001." || notice.Link != "/app/ajuan/p1" {
		t.Fatalf("notice = %+v", notice)
	}
}
