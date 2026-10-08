package domain

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
	agreementID   = "0192f0c4-0000-7000-8000-000000000301"
	breadID       = "0192f0c4-0000-7000-8000-000000000101"
	sambalID      = "0192f0c4-0000-7000-8000-000000000102"
	stoppedID     = "0192f0c4-0000-7000-8000-000000000103"
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	consignorID   = "0192f0c4-0000-7000-8000-000000000402"
	cashierID     = "0192f0c4-0000-7000-8000-000000000501"
)

func nineInTheMorning() time.Time {
	return time.Date(2026, 11, 10, 9, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
}

func aSetting() Setting {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	sambal := pricing.Terms{Model: pricing.Commission, RetailPrice: 35000, CommissionBPS: 2000}
	return Setting{
		PartnershipID: partnershipID, IsActive: true, AgreementID: agreementID, AgreementVersion: 2,
		StoreID: storeID, StoreName: "Warung Bu Sri", ConsignorID: consignorID,
		Products: map[string]AgreedProduct{
			breadID:  {ProductID: breadID, ShelfLife: stock.ShelfLifeOf(3), Terms: bread},
			sambalID: {ProductID: sambalID, ShelfLife: stock.UnknownShelfLife(), Terms: sambal},
		},
	}
}

func anInput() RecordInput {
	return RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 2, OccurredAt: nineInTheMorning(),
		Items: []ItemInput{{ProductID: breadID, Qty: 20}, {ProductID: sambalID, Qty: 6, ExpiryDate: "2027-01-31"}},
	}
}

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok || appErr.Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v, want a validation error", err)
	}
	codes := map[string]string{}
	for _, problem := range appErr.Fields() {
		codes[problem.Field] = problem.Code
	}
	return codes
}

func codeOf(err error) apperror.Code {
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		return ""
	}
	return appErr.Code()
}

func TestCheck_BRDRP01_RefusesAnEndedPartnership(t *testing.T) {
	setting := aSetting()
	setting.IsActive = false

	_, err := setting.Check(anInput(), nineInTheMorning())

	if codeOf(err) != apperror.CodePartnershipNotActive {
		t.Fatalf("err = %v, want PARTNERSHIP_NOT_ACTIVE", err)
	}
}

func TestCheck_BRPWA06_RefusesAnOlderAgreementVersion(t *testing.T) {
	input := anInput()
	input.AgreementVersion = 1

	_, err := aSetting().Check(input, nineInTheMorning())

	appErr, _ := errors.AsType[*apperror.Error](err)
	if codeOf(err) != apperror.CodeAgreementChanged || appErr.Message() !=
		"Kesepakatan sudah berubah ke versi 2. Tinjau lagi catatan ini." {
		t.Fatalf("err = %v", err)
	}
}

func TestCheck_BRDRP01_BRAGR07_RefusesProductsOutsideTheActiveVersion(t *testing.T) {
	input := anInput()
	input.Items = append(input.Items, ItemInput{ProductID: stoppedID, Qty: 1})

	_, err := aSetting().Check(input, nineInTheMorning())

	if fieldCodes(t, err)["items[2].product_id"] != "PRODUCT_NOT_IN_AGREEMENT" {
		t.Fatalf("codes = %v", fieldCodes(t, err))
	}
}

func TestCheck_BRDRP02_QuantityBoundaries(t *testing.T) {
	cases := []struct {
		qty     int
		isValid bool
	}{{0, false}, {1, true}, {10000, true}, {10001, false}}
	for _, tc := range cases {
		input := anInput()
		input.Items = []ItemInput{{ProductID: breadID, Qty: tc.qty}}

		_, err := aSetting().Check(input, nineInTheMorning())

		if tc.isValid != (err == nil) {
			t.Fatalf("qty %d: err = %v", tc.qty, err)
		}
		if !tc.isValid && fieldCodes(t, err)["items[0].qty"] != "OUT_OF_RANGE" {
			t.Fatalf("qty %d: codes = %v", tc.qty, fieldCodes(t, err))
		}
	}
}

func TestCheck_BRDRP02_LineCountBoundaries(t *testing.T) {
	cases := []struct {
		lines int
		code  string
	}{{0, "REQUIRED"}, {1, ""}, {50, ""}, {51, "TOO_MANY"}}
	for _, tc := range cases {
		input := anInput()
		input.Items = make([]ItemInput, 0, tc.lines)
		for range tc.lines {
			input.Items = append(input.Items, ItemInput{ProductID: breadID, Qty: 1})
		}

		_, err := aSetting().Check(input, nineInTheMorning())

		if tc.code == "" && err != nil {
			t.Fatalf("%d lines: err = %v", tc.lines, err)
		}
		if tc.code != "" && fieldCodes(t, err)["items"] != tc.code {
			t.Fatalf("%d lines: codes = %v", tc.lines, fieldCodes(t, err))
		}
	}
}

func TestCheck_BRDRP06_DropDateGoesBackThreeCalendarDaysAtMost(t *testing.T) {
	now := nineInTheMorning()
	cases := []struct {
		name       string
		occurredAt time.Time
		isValid    bool
	}{
		{"start of the third day back", time.Date(2026, 11, 7, 0, 0, 0, 0, now.Location()), true},
		{"last minute of the fourth day back", time.Date(2026, 11, 6, 23, 59, 0, 0, now.Location()), false},
		{"five minutes ahead", now.Add(5 * time.Minute), true},
		{"six minutes ahead", now.Add(6 * time.Minute), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := anInput()
			input.OccurredAt = tc.occurredAt

			_, err := aSetting().Check(input, now)

			if tc.isValid != (err == nil) {
				t.Fatalf("err = %v", err)
			}
			if !tc.isValid && fieldCodes(t, err)["occurred_at"] != "OUT_OF_RANGE" {
				t.Fatalf("codes = %v", fieldCodes(t, err))
			}
		})
	}
}

func TestCheck_BRPRD02_FillsExpiryFromTheShelfLife(t *testing.T) {
	draft, err := aSetting().Check(anInput(), nineInTheMorning())
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	breadExpiry, hasBreadExpiry := draft.Items[0].Expiry.Date()
	sambalExpiry, hasSambalExpiry := draft.Items[1].Expiry.Date()
	if !hasBreadExpiry || breadExpiry.Format(time.DateOnly) != "2026-11-13" {
		t.Fatalf("bread expiry = %v %v", breadExpiry, hasBreadExpiry)
	}
	if !hasSambalExpiry || sambalExpiry.Format(time.DateOnly) != "2027-01-31" {
		t.Fatalf("sambal expiry = %v %v", sambalExpiry, hasSambalExpiry)
	}
}

func TestCheck_LeavesExpiryEmptyWithoutAShelfLife(t *testing.T) {
	input := anInput()
	input.Items = []ItemInput{{ProductID: sambalID, Qty: 2}}

	draft, err := aSetting().Check(input, nineInTheMorning())

	if _, hasExpiry := draft.Items[0].Expiry.Date(); err != nil || hasExpiry {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestCheck_BRDRP05_ExpiryMayNotPrecedeTheDropDate(t *testing.T) {
	cases := []struct {
		expiry string
		code   string
	}{{"2026-11-10", ""}, {"2026-11-09", "EXPIRY_BEFORE_DROP"}, {"10-11-2026", "INVALID_DATE"}}
	for _, tc := range cases {
		input := anInput()
		input.Items = []ItemInput{{ProductID: breadID, Qty: 1, ExpiryDate: tc.expiry}}

		_, err := aSetting().Check(input, nineInTheMorning())

		if tc.code == "" && err != nil {
			t.Fatalf("%s: err = %v", tc.expiry, err)
		}
		if tc.code != "" && fieldCodes(t, err)["items[0].expiry_date"] != tc.code {
			t.Fatalf("%s: codes = %v", tc.expiry, fieldCodes(t, err))
		}
	}
}

func TestCheck_BRAGR06_CopiesTheTermsOfTheActiveVersion(t *testing.T) {
	draft, err := aSetting().Check(anInput(), nineInTheMorning())

	if err != nil || draft.Items[1].Terms != aSetting().Products[sambalID].Terms {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func aDrop(t *testing.T) Drop {
	t.Helper()
	draft, err := aSetting().Check(anInput(), nineInTheMorning())
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return draft.Record(Recording{
		IDs: RecordIDs{
			DropID: "drop-1", ItemIDs: []string{"item-1", "item-2"}, MovementIDs: []string{"move-1", "move-2"},
		},
		Sequence:       7,
		RecordedBy:     cashierID,
		IdempotencyKey: "0192f0c4-0000-7000-8000-000000000999",
		Fingerprint:    []byte{1, 2, 3},
		At:             nineInTheMorning().Add(time.Minute),
	})
}

func TestRecord_FRDRP05_NumbersPerStoreAndMonthOfReceipt(t *testing.T) {
	drop := aDrop(t)

	if drop.Number != "TM-202611-0007" || drop.Items[1].BatchCode != "TM-202611-0007/2" {
		t.Fatalf("number = %s, batch = %s", drop.Number, drop.Items[1].BatchCode)
	}
}

func TestRecord_BRDRP03_StartsRecordedAndReceivedNow(t *testing.T) {
	drop := aDrop(t)

	if drop.Status != Recorded || !drop.ReceivedAt.Equal(nineInTheMorning().Add(time.Minute)) ||
		drop.StoreID != storeID || drop.ConsignorID != consignorID || drop.TotalQty() != 26 {
		t.Fatalf("drop = %+v", drop)
	}
}

func TestMovements_BRSTK03_ReceiveEachBatchWithItsSource(t *testing.T) {
	drop := aDrop(t)

	movements := drop.Movements()

	if len(movements) != 2 || movements[0].ID != "move-1" || movements[0].DropItemID != "item-1" ||
		movements[0].Qty != 20 ||
		movements[1].ProductID != sambalID || movements[1].SourceNumber != "TM-202611-0007" {
		t.Fatalf("movements = %+v", movements)
	}
}

func TestNotice_FRDRP06_TellsTheConsignorTheUnitsAndNumber(t *testing.T) {
	notice := aDrop(t).Notice("Warung Bu Sri")

	if notice.OrganizationID != consignorID || notice.Type != "DROP_RECORDED" || notice.OwnersOnly ||
		notice.Body != "Warung Bu Sri mencatat 26 unit titipan masuk (TM-202611-0007)." ||
		notice.Link != "/app/titipan/drop-1" || notice.ActorUserID != cashierID {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestSealRequest_RSEC28_ContentLeavesTheStatusOut(t *testing.T) {
	drop := aDrop(t)
	confirmed := drop
	confirmed.Status = Confirmed

	recorded, err := drop.ContentDigest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	later, err := confirmed.ContentDigest()

	if err != nil || !bytes.Equal(recorded, later) {
		t.Fatalf("digests differ: %x %x %v", recorded, later, err)
	}
}

func TestSealRequest_RSEC28_ContentChangesWithAQuantity(t *testing.T) {
	drop := aDrop(t)
	changed := drop
	changed.Items = append([]Item(nil), drop.Items...)
	changed.Items[0].Qty = 19

	original, originalErr := drop.ContentDigest()
	tampered, tamperedErr := changed.ContentDigest()

	if errors.Join(originalErr, tamperedErr) != nil || bytes.Equal(original, tampered) {
		t.Fatalf("digests = %x %x", original, tampered)
	}
}

func TestSealRequest_NamesTheRecordedEvent(t *testing.T) {
	request := aDrop(t).SealRequest()

	if request.Event != "DROP_RECORDED" || request.DocumentType != EntityType || request.DocumentNumber !=
		"TM-202611-0007" || request.ActorUserID != cashierID || request.ActorOrgID != storeID {
		t.Fatalf("request = %+v", request)
	}
}

func TestFingerprint_BRDRP07_SameRequestSameFingerprint(t *testing.T) {
	first, firstErr := anInput().Fingerprint()
	second, secondErr := anInput().Fingerprint()
	changed := anInput()
	changed.Items[0].Qty = 21
	third, thirdErr := changed.Fingerprint()

	if errors.Join(firstErr, secondErr, thirdErr) != nil || !bytes.Equal(first, second) || bytes.Equal(first, third) {
		t.Fatalf("fingerprints = %x %x %x", first, second, third)
	}
}
