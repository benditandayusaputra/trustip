package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func foundingInput() FoundingInput {
	bread, _ := pricing.NewTerms(pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750})
	chips, _ := pricing.NewTerms(pricing.TermsInput{Model: "COMMISSION", RetailPrice: 15000, CommissionBPS: 2000})
	return FoundingInput{
		ProposalID: "p1", ProposalNumber: "AJ-202610-0001", ConsignorID: "c1", StoreID: "s1",
		Terms:      terms.General{Schedule: terms.ScheduleWeekly, PaymentTermsDays: 7, Rhythm: terms.RhythmDaily},
		Items:      []Item{{ProductID: "b", Terms: chips}, {ProductID: "a", Terms: bread}},
		ProposedBy: "u1", AcceptedBy: "u2", At: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC),
	}
}

func TestFound_BRPRP03_ActivePartnershipAndVersionOne(t *testing.T) {
	founding := Found(foundingInput(), FoundingIDs{PartnershipID: "k1", AgreementID: "a1"})

	agreement := founding.Agreement
	if founding.Partnership.ConsignorID != "c1" || founding.Partnership.StoreID != "s1" || agreement.Version != 1 ||
		agreement.Status != Active || agreement.PartnershipID != "k1" || !agreement.Items[0].IsActive ||
		agreement.Number() != "AJ-202610-0001/V1" {
		t.Fatalf("founding = %+v", founding)
	}
}

func TestContent_BRSIG01_SameAgreementGivesTheSameHashWhateverTheItemOrder(t *testing.T) {
	first := Found(foundingInput(), FoundingIDs{PartnershipID: "k1", AgreementID: "a1"}).Agreement
	reordered := first
	reordered.Items = []Item{first.Items[1], first.Items[0]}

	firstHash, firstErr := canonical.Digest(first.Content())
	reorderedHash, reorderedErr := canonical.Digest(reordered.Content())

	if firstErr != nil || reorderedErr != nil || string(firstHash) != string(reorderedHash) {
		t.Fatalf("hashes differ: %x %x (%v %v)", firstHash, reorderedHash, firstErr, reorderedErr)
	}
}

func TestContent_RSEC33_ChangedPriceChangesTheHash(t *testing.T) {
	agreement := Found(foundingInput(), FoundingIDs{PartnershipID: "k1", AgreementID: "a1"}).Agreement
	before, _ := canonical.Digest(agreement.Content())
	agreement.Items[1].Terms.DepositPrice = 1700

	after, _ := canonical.Digest(agreement.Content())

	if string(before) == string(after) {
		t.Fatal("hash ignored the price")
	}
}
