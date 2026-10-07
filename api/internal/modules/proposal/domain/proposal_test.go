package domain

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

const (
	breadID     = "0192f0c4-0000-7000-8000-000000000101"
	chipsID     = "0192f0c4-0000-7000-8000-000000000102"
	consignorID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	storeID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	otherStore  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7e"
)

func sentAt() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func validInput() RevisionInput {
	return RevisionInput{
		Terms: TermsInput{SettlementSchedule: "WEEKLY", PaymentTermsDays: 7, CountRhythm: "DAILY", Notes: "Rak depan."},
		Items: []ItemInput{
			{ProductID: breadID, Terms: depositTerms(2000, 1750)},
			{ProductID: chipsID, Terms: commissionTerms(15000, 2000)},
		},
		Message: "Salam kenal, Bu.",
	}
}

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		t.Fatalf("err = %v, want a validation error", err)
	}
	codes := map[string]string{}
	for _, problem := range appErr.Fields() {
		codes[problem.Field] = problem.Code
	}
	return codes
}

func TestNewRevisionDraft_FRPRP02_ProductsWithTermsAndGeneralTerms(t *testing.T) {
	draft, err := NewRevisionDraft(validInput())

	if err != nil || len(draft.Items) != 2 || draft.Items[0].Action != Add ||
		draft.Items[1].Terms.CommissionBPS != 2000 || draft.Terms.Schedule != ScheduleWeekly {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestNewRevisionDraft_RSEC47_ItemProblemsNameTheirRow(t *testing.T) {
	input := validInput()
	input.Items[1].Terms.CommissionBPS = 5001
	input.Items[0].Terms.DepositPrice = 2001

	codes := fieldCodes(t, mustFail(NewRevisionDraft(input)))

	if codes["items[1].commission_bps"] != "OUT_OF_RANGE" || codes["items[0].deposit_price"] != "EXCEEDS_RETAIL_PRICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewRevisionDraft_GeneralTermsBoundaries(t *testing.T) {
	cases := []struct {
		name  string
		terms TermsInput
		field string
		code  string
	}{
		{
			"payment terms 30 days",
			TermsInput{SettlementSchedule: "ON_COUNT", PaymentTermsDays: 30, CountRhythm: "WEEKLY"},
			"", "",
		},
		{
			"payment terms 31 days",
			TermsInput{SettlementSchedule: "ON_COUNT", PaymentTermsDays: 31, CountRhythm: "WEEKLY"},
			"terms.payment_terms_days", "OUT_OF_RANGE",
		},
		{
			"negative payment terms",
			TermsInput{SettlementSchedule: "MONTHLY", PaymentTermsDays: -1, CountRhythm: "DAILY"},
			"terms.payment_terms_days", "OUT_OF_RANGE",
		},
		{
			"unknown schedule",
			TermsInput{SettlementSchedule: "DAILY", CountRhythm: "DAILY"},
			"terms.settlement_schedule", "UNKNOWN_CHOICE",
		},
		{"missing rhythm", TermsInput{SettlementSchedule: "WEEKLY"}, "terms.count_rhythm", "REQUIRED"},
		{"notes 501 characters", TermsInput{
			SettlementSchedule: "WEEKLY", CountRhythm: "EACH_VISIT",
			Notes: strings.Repeat("a", MaxNotesLength+1),
		}, "terms.notes", "TOO_LONG"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := validInput()
			input.Terms = tc.terms

			_, err := NewRevisionDraft(input)

			if tc.field == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tc.field != "" && fieldCodes(t, err)[tc.field] != tc.code {
				t.Fatalf("codes = %v", fieldCodes(t, err))
			}
		})
	}
}

func TestNewRevisionDraft_BRPRP09_MessageUpToOneThousandCharacters(t *testing.T) {
	atLimit := validInput()
	atLimit.Message = strings.Repeat("é", MaxMessageLength)
	overLimit := validInput()
	overLimit.Message = strings.Repeat("é", MaxMessageLength+1)

	if _, err := NewRevisionDraft(atLimit); err != nil {
		t.Fatalf("at limit: %v", err)
	}
	if fieldCodes(t, mustFail(NewRevisionDraft(overLimit)))["message"] != "TOO_LONG" {
		t.Fatal("over limit accepted")
	}
}

func TestNewRevisionDraft_SameProductTwiceIsRejected(t *testing.T) {
	input := validInput()
	input.Items[1].ProductID = breadID

	if fieldCodes(t, mustFail(NewRevisionDraft(input)))["items[1].product_id"] != "UNKNOWN_CHOICE" {
		t.Fatal("duplicate product accepted")
	}
}

func TestNewRevisionDraft_TooManyProducts(t *testing.T) {
	input := validInput()
	input.Items = nil
	for index := range MaxItems + 1 {
		input.Items = append(input.Items, ItemInput{
			ProductID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index),
			Terms:     pricing.TermsInput{Model: "COMMISSION", RetailPrice: 1000},
		})
	}

	if fieldCodes(t, mustFail(NewRevisionDraft(input)))["items"] != "TOO_MANY" {
		t.Fatal("too many products accepted")
	}
}

func consignorSender() Sender {
	return Sender{UserID: "u1", OrganizationID: consignorID, OrganizationType: actor.Consignor}
}

func TestInvite_FRPRP07_ConsignorInvitationWaitsForAStore(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())

	invitation, err := draft.Invite(InvitationParams{
		IDs: InvitationIDs{ProposalID: "p1", RevisionID: "r1", Code: "ABCDEFGH23"}, Sender: consignorSender(),
		Sequence: 7, At: sentAt(),
	})

	if err != nil || invitation.ConsignorID() != consignorID || invitation.StoreID() != "" ||
		invitation.Number != "AJ-202610-0007" || !invitation.ExpiresAt.Equal(sentAt().Add(14*24*time.Hour)) {
		t.Fatalf("invitation = %+v, err = %v", invitation, err)
	}
}

func TestInvite_BRPRP01_StoreInvitationCannotNameProductsBeforeAConsignorJoins(t *testing.T) {
	draft, _ := NewRevisionDraft(validInput())
	store := Sender{UserID: "u2", OrganizationID: storeID, OrganizationType: actor.Store}

	_, err := draft.Invite(InvitationParams{IDs: InvitationIDs{Code: "ABCDEFGH23"}, Sender: store, At: sentAt()})

	if fieldCodes(t, err)["items"] != "UNKNOWN_CHOICE" {
		t.Fatalf("err = %v", err)
	}
}

func TestIsInviteCode_BRPRP07_TenCharactersFromTheUnambiguousAlphabet(t *testing.T) {
	cases := map[string]bool{
		"ABCDEFGH23": true, "ABCDEFGH2": false, "ABCDEFGH234": false, "ABCDEFGHI2": false,
		"abcdefgh23": false, "ABCDEFGH01": false,
	}
	for code, want := range cases {
		if got := IsInviteCode(code); got != want {
			t.Errorf("IsInviteCode(%q) = %v", code, got)
		}
	}
}

func openInvitation() InvitationView {
	return InvitationView{
		ProposalID: "p1", Status: Open, ExpiresAt: sentAt().Add(OpenFor), ConsignorID: consignorID,
		Sender: SenderView{OrganizationID: consignorID, Name: "Roti Pak Anton", Type: actor.Consignor},
	}
}

func storeMember(organizationID string) actor.Membership {
	return actor.Membership{OrganizationID: organizationID, OrganizationType: actor.Store, Role: actor.Owner}
}

func TestClaimBy_BRPRP07_StoreClaimsAConsignorInvitation(t *testing.T) {
	claim, err := openInvitation().ClaimBy(storeMember(storeID), sentAt())

	if err != nil || claim.ConsignorID != consignorID || claim.StoreID != storeID || claim.ClaimerID != storeID ||
		claim.IsRepeat {
		t.Fatalf("claim = %+v, err = %v", claim, err)
	}
}

func TestClaimBy_BRPRP07_SameTypeIsToldWhoTheInvitationIsFor(t *testing.T) {
	consignor := actor.Membership{OrganizationID: "x", OrganizationType: actor.Consignor, Role: actor.Owner}

	_, err := openInvitation().ClaimBy(consignor, sentAt())

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeInviteWrongOrgType || appErr.Message() != "Undangan ini untuk toko." {
		t.Fatalf("err = %v, message = %q", err, appErr.Message())
	}
}

func TestClaimBy_BRPRP07_OnlyOnce(t *testing.T) {
	claimed := openInvitation()
	claimed.StoreID = storeID

	_, otherErr := claimed.ClaimBy(storeMember(otherStore), sentAt())
	again, againErr := claimed.ClaimBy(storeMember(storeID), sentAt())

	if apperror.From(otherErr).Code() != apperror.CodeNotFound || againErr != nil || !again.IsRepeat {
		t.Fatalf("other = %v, again = %+v %v", otherErr, again, againErr)
	}
}

func TestClaimBy_BRPRP08_ExpiredInvitationCannotBeClaimed(t *testing.T) {
	invitation := openInvitation()

	_, err := invitation.ClaimBy(storeMember(storeID), invitation.ExpiresAt)

	if apperror.From(err).Code() != apperror.CodeProposalExpired {
		t.Fatalf("err = %v", err)
	}
}

func TestClaimBy_DecidedInvitationIsNotFoundForStrangers(t *testing.T) {
	invitation := openInvitation()
	invitation.Status = Withdrawn

	_, err := invitation.ClaimBy(storeMember(storeID), sentAt())

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func mustFail[T any](_ T, err error) error {
	return err
}

func depositTerms(retail, deposit int64) pricing.TermsInput {
	return pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: retail, DepositPrice: deposit}
}

func commissionTerms(retail int64, commissionBPS int) pricing.TermsInput {
	return pricing.TermsInput{Model: "COMMISSION", RetailPrice: retail, CommissionBPS: commissionBPS}
}
