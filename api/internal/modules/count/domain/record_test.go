package domain

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
	breadID       = "0192f0c4-0000-7000-8000-000000000101"
	chipsID       = "0192f0c4-0000-7000-8000-000000000102"
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	consignorID   = "0192f0c4-0000-7000-8000-000000000402"
	cashierID     = "0192f0c4-0000-7000-8000-000000000501"
)

func noon() time.Time {
	return time.Date(2026, 11, 10, 12, 0, 0, 0, time.FixedZone("WIB", 7*60*60))
}

func aSheet() Sheet {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	chips := pricing.Terms{Model: pricing.Commission, RetailPrice: 15000, CommissionBPS: 2000}
	return Sheet{
		PartnershipID: partnershipID, IsActive: true, AgreementVersion: 1, StoreID: storeID,
		StoreName: "Warung Bu Sri", ConsignorID: consignorID,
		LastDocument: LastDocument{At: noon().Add(-2 * time.Hour), Given: true},
		Products: []StockedProduct{
			{ProductID: breadID, Batches: []PricedBatch{
				{Stock: stock.Batch{ID: "bread-1", Balance: 20}, Terms: bread},
			}},
			{ProductID: chipsID, Batches: []PricedBatch{
				{Stock: stock.Batch{ID: "chips-1", Balance: 6}, Terms: chips},
			}},
		},
	}
}

func anInput() RecordInput {
	return RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 1, CountedAt: claimed.At(noon()),
		Occasion: string(StoreClosing),
		Lines: []RequestLine{{
			ProductID: breadID, RecordedQty: 20, CountedQty: 2, ReturnedQty: 2, ReturnReason: string(ReasonUnsold),
		}},
	}
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return appErr.Code()
	}
	return ""
}

func TestCheck_BRDRP01_EndedPartnershipCannotBeCounted(t *testing.T) {
	sheet := aSheet()
	sheet.IsActive = false

	_, err := sheet.Check(anInput(), noon())

	if codeOf(err) != apperror.CodePartnershipNotActive {
		t.Fatalf("err = %v", err)
	}
}

func TestCheck_BRPWA06_StockSeenByTheUserMustStillBeTheStock(t *testing.T) {
	input := anInput()
	input.Lines[0].RecordedQty = 21

	_, err := aSheet().Check(input, noon())

	appErr, _ := errors.AsType[*apperror.Error](err)
	if codeOf(err) != apperror.CodeStockChanged ||
		appErr.Message() != "Stok baru saja berubah. Muat ulang lalu hitung lagi." {
		t.Fatalf("err = %v", err)
	}
}

func TestCheck_BRCNT03_OnlyStockedProductsOnceEach(t *testing.T) {
	input := anInput()
	input.Lines = append(input.Lines, input.Lines[0], RequestLine{ProductID: "missing"})

	_, err := aSheet().Check(input, noon())

	codes := fieldCodes(t, err)
	if codes["lines[1].product_id"] != "UNKNOWN_CHOICE" || codes["lines[2].product_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestCheck_BRCNT06_CountTimeBoundaries(t *testing.T) {
	cases := []struct {
		name      string
		countedAt time.Time
		code      string
	}{
		{"now", noon(), ""},
		{"before the last document", noon().Add(-3 * time.Hour), "BEFORE_LAST_DOCUMENT"},
		{"six minutes ahead", noon().Add(6 * time.Minute), "OUT_OF_RANGE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := anInput()
			input.CountedAt = claimed.At(tc.countedAt)

			_, err := aSheet().Check(input, noon())

			if tc.code == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tc.code != "" && fieldCodes(t, err)["counted_at"] != tc.code {
				t.Fatalf("codes = %v", fieldCodes(t, err))
			}
		})
	}
}

func TestCheck_BRCNT06_ThreeCalendarDaysBackWithoutEarlierDocuments(t *testing.T) {
	sheet := aSheet()
	sheet.LastDocument = LastDocument{}
	input := anInput()
	input.CountedAt = claimed.At(time.Date(2026, 11, 6, 23, 59, 0, 0, noon().Location()))

	_, err := sheet.Check(input, noon())

	if fieldCodes(t, err)["counted_at"] != "OUT_OF_RANGE" {
		t.Fatalf("codes = %v", fieldCodes(t, err))
	}
}

func TestCheck_FRCNT03_OccasionIsRequiredFromTheList(t *testing.T) {
	for occasion, code := range map[string]string{"": "REQUIRED", "LUNCH": "UNKNOWN_CHOICE"} {
		input := anInput()
		input.Occasion = occasion

		_, err := aSheet().Check(input, noon())

		if fieldCodes(t, err)["occasion"] != code {
			t.Fatalf("%q: codes = %v", occasion, fieldCodes(t, err))
		}
	}
}

func TestCheck_BRCNT01_CountAboveTheStockIsAFieldError(t *testing.T) {
	input := anInput()
	input.Lines[0].CountedQty = 21
	input.Lines[0].ReturnedQty = 0

	_, err := aSheet().Check(input, noon())

	if fieldCodes(t, err)["lines[0].counted_qty"] != "COUNT_EXCEEDS_STOCK" {
		t.Fatalf("codes = %v", fieldCodes(t, err))
	}
}

func idsFor(draft Draft) []string {
	ids := make([]string, 0, draft.IDCount())
	for index := range draft.IDCount() {
		ids = append(ids, fmt.Sprintf("id-%d", index))
	}
	return ids
}

func aCount(t *testing.T) Count {
	t.Helper()
	draft, err := aSheet().Check(anInput(), noon())
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return draft.Record(Recording{
		IDs: idsFor(draft), Sequence: 3, RecordedBy: cashierID, IdempotencyKey: "key",
		Fingerprint: []byte{1}, At: noon().Add(time.Minute),
	})
}

func TestRecord_FRCNT05_SectionTwelveRecap(t *testing.T) {
	count := aCount(t)

	if count.Number != "HS-202611-0003" || count.SoldQty != 18 || count.ReturnedQty != 2 ||
		count.Split != (pricing.Split{Gross: 36000, StoreShare: 4500, ConsignorShare: 31500}) {
		t.Fatalf("count = %+v", count)
	}
}

func TestMovements_BRCNT04_SaleAndReturnLeaveTheirBatches(t *testing.T) {
	movements := aCount(t).Movements()

	if len(movements) != 2 || movements[0].Type != "SALE" || movements[0].Qty != -18 ||
		movements[1].Type != "RETURN" || movements[1].Qty != -2 || movements[1].DropItemID != "bread-1" {
		t.Fatalf("movements = %+v", movements)
	}
}

func TestNotice_FRCNT08_TellsTheConsignorSoldAndReturned(t *testing.T) {
	notice := aCount(t).Notice("Warung Bu Sri")

	if notice.OrganizationID != consignorID || notice.Type != "COUNT_RECORDED" ||
		notice.Body != "Warung Bu Sri menghitung sisa: 18 terjual, 2 dikembalikan (HS-202611-0003)." {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestContentDigest_RSEC28_ChangesWithASoldQuantity(t *testing.T) {
	count := aCount(t)
	tampered := count
	tampered.Lines = append([]CountLine(nil), count.Lines...)
	tampered.Lines[0].SoldQty = 17

	original, originalErr := count.ContentDigest()
	changed, changedErr := tampered.ContentDigest()

	if errors.Join(originalErr, changedErr) != nil || bytes.Equal(original, changed) {
		t.Fatalf("digests = %x %x", original, changed)
	}
}

func TestFingerprint_BRCNT07_FollowsTheRequest(t *testing.T) {
	first, firstErr := anInput().Fingerprint()
	changed := anInput()
	changed.Lines[0].CountedQty = 3
	second, secondErr := changed.Fingerprint()

	if errors.Join(firstErr, secondErr) != nil || bytes.Equal(first, second) {
		t.Fatalf("fingerprints = %x %x", first, second)
	}
}

func TestCheck_BRCNT06_CountCannotLandInASettledPeriod(t *testing.T) {
	sheet := aSheet()
	sheet.Settled = Settled{Until: noon().Add(-time.Hour), Number: "ST-202611-0004", Given: true}
	input := anInput()
	input.CountedAt = claimed.At(noon().Add(-90 * time.Minute))
	sheet.LastDocument = LastDocument{}

	_, err := sheet.Check(input, noon())

	appErr, _ := errors.AsType[*apperror.Error](err)
	if fieldCodes(t, err)["counted_at"] != "PERIOD_ALREADY_SETTLED" ||
		appErr.Fields()[0].Message != "Tanggal ini sudah masuk Setoran ST-202611-0004." {
		t.Fatalf("err = %v", err)
	}
	input.CountedAt = claimed.At(noon().Add(-time.Hour))
	if _, err := sheet.Check(input, noon()); err != nil {
		t.Fatalf("at the settled boundary: %v", err)
	}
}

func TestCheck_BRPWA07_OmittedCountTimeIsTheDeviceTime(t *testing.T) {
	input := anInput()
	input.CountedAt = claimed.Absent()
	input.ClientRecordedAt = claimed.At(noon().Add(-time.Hour))

	draft, err := aSheet().Check(input, noon())
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	if !draft.CountedAt.Equal(noon().Add(-time.Hour)) || !draft.ClientRecordedAt.Equal(noon().Add(-time.Hour)) {
		t.Fatalf("counted = %s, client = %s", draft.CountedAt, draft.ClientRecordedAt)
	}
}

func TestCheck_BRPWA07_CountDeviceTimeTooFarAheadBecomesTheReceipt(t *testing.T) {
	input := anInput()
	input.CountedAt = claimed.Absent()
	input.ClientRecordedAt = claimed.At(noon().Add(6 * time.Minute))

	draft, err := aSheet().Check(input, noon())
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	if !draft.CountedAt.Equal(noon()) || !draft.ClientRecordedAt.Equal(noon()) {
		t.Fatalf("counted = %s, client = %s", draft.CountedAt, draft.ClientRecordedAt)
	}
}

func TestRecord_FRPWA10_KeepsTheOfflineMarkOnTheCount(t *testing.T) {
	input := anInput()
	input.ClientRecordedAt = claimed.At(noon().Add(-time.Hour))
	input.RecordedOffline = true
	draft, err := aSheet().Check(input, noon())
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	count := draft.Record(Recording{IDs: idsFor(draft), At: noon()})

	if !count.RecordedOffline || !count.ClientRecordedAt.Equal(noon().Add(-time.Hour)) ||
		!count.ReceivedAt.Equal(noon()) {
		t.Fatalf("count = %+v", count)
	}
}

func TestFingerprint_BRPWA03_OfflineMarkLeavesTheCountFingerprintAlone(t *testing.T) {
	offline := anInput()
	offline.RecordedOffline = true

	first, firstErr := anInput().Fingerprint()
	second, secondErr := offline.Fingerprint()

	if errors.Join(firstErr, secondErr) != nil || !bytes.Equal(first, second) {
		t.Fatalf("fingerprints differ: %x %x", first, second)
	}
}
