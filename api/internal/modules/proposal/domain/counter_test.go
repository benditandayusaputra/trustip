package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func storeCounter(t *testing.T, items ...ItemInput) RevisionDraft {
	t.Helper()
	draft, err := NewRevisionDraft(RevisionInput{
		Terms: terms.Input{SettlementSchedule: "WEEKLY", PaymentTermsDays: 14, CountRhythm: "DAILY"},
		Items: items, Message: "Bisa harga setor 1.700?",
	})
	if err != nil {
		t.Fatalf("draft: %v", err)
	}
	return draft
}

func TestCounter_FRPRP04_ReceiverRevisesAndTheTurnMovesBack(t *testing.T) {
	proposal := claimedProposal()
	countedAt := sentAt().Add(48 * time.Hour)

	draft := storeCounter(t, ItemInput{ProductID: breadID, Terms: depositTerms(2000, 1700)})

	offer, err := proposal.Counter(draft, CounterParams{RevisionID: "r2", Counterer: storeDecider(), At: countedAt})

	if err != nil || offer.RevisionNo != 2 || offer.PreviousNo != 1 || offer.AwaitingOrgID != consignorID ||
		offer.AuthorOrgID != storeID || offer.ConsignorID != consignorID {
		t.Fatalf("offer = %+v, err = %v", offer, err)
	}
	if !offer.ExpiresAt.Equal(countedAt.Add(14 * 24 * time.Hour)) {
		t.Fatalf("BR-PRP-08 expires at %v", offer.ExpiresAt)
	}
}

func TestCounter_BRPRP02_AuthorOfTheLastRevisionWaits(t *testing.T) {
	proposal := claimedProposal()
	consignor := Decider{UserID: consignorUser, OrganizationID: consignorID}

	_, err := proposal.Counter(storeCounter(t), CounterParams{Counterer: consignor, At: sentAt()})

	if apperror.From(err).Code() != apperror.CodeNotYourTurn {
		t.Fatalf("err = %v", err)
	}
}

func TestCounter_DecidedOrExpiredProposalCannotBeCountered(t *testing.T) {
	declined := claimedProposal()
	declined.Status = Declined
	open := claimedProposal()

	_, decidedErr := declined.Counter(storeCounter(t), CounterParams{Counterer: storeDecider(), At: sentAt()})
	_, expiredErr := open.Counter(storeCounter(t), CounterParams{Counterer: storeDecider(), At: open.ExpiresAt})

	if apperror.From(decidedErr).Code() != apperror.CodeProposalDecided ||
		apperror.From(expiredErr).Code() != apperror.CodeProposalExpired {
		t.Fatalf("decided = %v, expired = %v", decidedErr, expiredErr)
	}
}

func TestCounter_FRPRP08_AmendmentCounterIsCheckedAgainstTheAgreement(t *testing.T) {
	proposal := claimedProposal()
	proposal.Kind, proposal.PartnershipID = Amendment, partnershipID

	draft := storeCounter(t, ItemInput{ProductID: jamID, Action: "UPDATE", Terms: depositTerms(1, 1)})
	params := CounterParams{Counterer: storeDecider(), Current: currentAgreement(), At: sentAt()}

	_, err := proposal.Counter(draft, params)

	if fieldCodes(t, err)["items[0].product_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("err = %v", err)
	}
}

func TestCounteroffer_FRNTF01_TellsTheOtherPartyAboutTheCounter(t *testing.T) {
	offer, _ := claimedProposal().Counter(storeCounter(t), CounterParams{
		RevisionID: "r2", Counterer: storeDecider(), At: sentAt(),
	})

	notice := offer.Notice()

	if notice.Type != "PROPOSAL_COUNTERED" || notice.OrganizationID != consignorID ||
		notice.Body != "Warung Bu Sri menawar ajuan AJ-202610-0001." || !notice.OwnersOnly {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestHistory_FRPRP04_MarksWhatChangedInEachRevision(t *testing.T) {
	proposal := claimedProposal()
	first := proposal.Revision
	second := first
	second.No, second.AuthorOrgID = 2, storeID
	second.Terms.PaymentTermsDays = 14
	changed := first.Items[0]
	changed.Terms.DepositPrice = 1700
	added := Item{ProductID: chipsID, Action: Add}
	second.Items = []Item{changed, added}

	history := proposal.History([]Revision{first, second})

	changes := history[1].Changes
	isFirstAdded := history[0].Changes.Items[0].Change == ItemAdded
	if len(history) != 2 || !isFirstAdded || history[1].AuthorName != "Warung Bu Sri" ||
		len(changes.TermFields) != 1 || changes.TermFields[0] != "payment_terms_days" ||
		changes.Items[0].Change != ItemChanged || changes.Items[1].Change != ItemAdded {
		t.Fatalf("history = %+v", history)
	}
}

func TestCompare_FRPRP04_RemovedAndUnchangedRows(t *testing.T) {
	previous := claimedProposal().Revision
	current := previous
	current.Items = []Item{}

	same := Compare(previous, previous)
	removed := Compare(previous, current)

	if same.Items[0].Change != ItemKept || len(same.TermFields) != 0 || removed.Items[0].Change != ItemRemoved {
		t.Fatalf("same = %+v, removed = %+v", same, removed)
	}
}
