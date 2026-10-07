package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func versionOne() Agreement {
	return Found(foundingInput(), FoundingIDs{PartnershipID: "k1", AgreementID: "a1"}).Agreement
}

func amendment(changes ...Change) AmendmentInput {
	general := foundingInput().Terms
	general.PaymentTermsDays = 14
	return AmendmentInput{
		PartnershipID: "k1", ProposalID: "p2", ProposalNumber: "AJ-202611-0004", Terms: general, Changes: changes,
		ProposedBy: "u2", AcceptedBy: "u1", At: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC),
	}
}

func TestAmend_BRAGR05_NextVersionKeepsUntouchedProducts(t *testing.T) {
	jam, _ := pricing.NewTerms(pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 35000, DepositPrice: 28000})

	next := versionOne().Amend(amendment(Change{ProductID: "c", Action: AddProduct, Terms: jam},
		Change{ProductID: "a", Action: StopProduct}), "a2")

	if next.Version != 2 || next.ID != "a2" || len(next.Items) != 3 || next.Items[0].IsActive ||
		next.Items[0].Terms.DepositPrice != 1750 || !next.Items[1].IsActive ||
		next.Items[2].Terms.RetailPrice != 35000 ||
		next.Number() != "AJ-202611-0004/V2" || next.Terms.PaymentTermsDays != 14 {
		t.Fatalf("next = %+v", next)
	}
}

func TestDiffOf_FRAGR04_ShowsTermsAndProductChanges(t *testing.T) {
	previous := versionOne()
	cheaper, _ := pricing.NewTerms(pricing.TermsInput{Model: "COMMISSION", RetailPrice: 15000, CommissionBPS: 1500})
	next := previous.Amend(amendment(Change{ProductID: "b", Action: UpdateProduct, Terms: cheaper},
		Change{ProductID: "a", Action: StopProduct}), "a2")

	diff := DiffOf(previous, next)

	if len(diff.Terms) != 1 || diff.Terms[0] != (TermChange{"payment_terms_days", "7", "14"}) || len(diff.Items) != 2 ||
		diff.Items[0].Kind != ItemStopped || diff.Items[1].Kind != ItemUpdated ||
		diff.Items[1].Before.CommissionBPS != 2000 || diff.Items[1].After.CommissionBPS != 1500 {
		t.Fatalf("diff = %+v", diff)
	}
}

func TestHistory_FRAGR04_FirstVersionListsEveryProductAsAdded(t *testing.T) {
	first := versionOne()
	second := first.Amend(amendment(), "a2")

	history := History([]Agreement{first, second})

	if len(history) != 2 || len(history[0].Diff.Items) != 2 || history[0].Diff.Items[0].Kind != ItemAdded ||
		len(history[1].Diff.Items) != 0 || len(history[1].Diff.Terms) != 1 {
		t.Fatalf("history = %+v", history)
	}
}

func TestAmend_StoppedProductCanBeAddedBack(t *testing.T) {
	stopped := versionOne().Amend(amendment(Change{ProductID: "a", Action: StopProduct}), "a2")
	again, _ := pricing.NewTerms(pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2500, DepositPrice: 2000})

	resumed := stopped.Amend(amendment(Change{ProductID: "a", Action: AddProduct, Terms: again}), "a3")

	if diff := DiffOf(stopped, resumed); len(diff.Items) != 1 || diff.Items[0].Kind != ItemAdded {
		t.Fatalf("diff = %+v", diff)
	}
	if resumed.Terms.Schedule != terms.ScheduleWeekly {
		t.Fatalf("resumed = %+v", resumed)
	}
}
