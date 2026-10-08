package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	consignorID   = "0192f0c4-0000-7000-8000-000000000402"
	ownerID       = "0192f0c4-0000-7000-8000-000000000501"
)

func anEndState() EndState {
	return EndState{
		PartnershipID: partnershipID, Status: Active, ConsignorID: consignorID, StoreID: storeID,
		StoreName: "Warung Bu Sri", ConsignorName: "Roti Pak Anton",
	}
}

func endedAt() time.Time {
	return time.Date(2026, 11, 20, 3, 0, 0, 0, time.UTC)
}

func TestEnd_FRPRT04_OwnerEndsAPartnershipWithoutBlockers(t *testing.T) {
	ending, err := anEndState().End(Ender{UserID: ownerID, OrganizationID: storeID}, "Toko tutup", endedAt())

	if err != nil || ending.EndedByOrgID != storeID || ending.Reason != "Toko tutup" {
		t.Fatalf("ending = %+v, err = %v", ending, err)
	}
	request := ending.SealRequest()
	if request.Event != "PARTNERSHIP_ENDED" || request.DocumentType != "PARTNERSHIP" ||
		request.Content["end_reason"] != "Toko tutup" {
		t.Fatalf("request = %+v", request)
	}
	notice := ending.Notice()
	if notice.OrganizationID != consignorID || !notice.OwnersOnly ||
		notice.Body != "Warung Bu Sri mengakhiri kemitraan." {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestEnd_BRPRT02_BlockersAreListedInTheRefusal(t *testing.T) {
	state := anEndState()
	state.Blockers = []Blocker{
		{Kind: StockAtStore, Qty: 12},
		{Kind: PendingCount, DocumentNumber: "HS-202611-0004"},
		{Kind: UnpaidSettlement, DocumentNumber: "ST-202611-0002", Amount: 31500},
	}

	_, err := state.End(Ender{UserID: ownerID, OrganizationID: consignorID}, "", endedAt())

	appErr, _ := errors.AsType[*apperror.Error](err)
	want := "Belum bisa diakhiri: stok 12 unit masih di toko, Hitung Sisa HS-202611-0004 menunggu konfirmasi, " +
		"Setoran ST-202611-0002 belum lunas."
	if appErr == nil || appErr.Code() != apperror.CodePartnershipBlocked || appErr.Message() != want {
		t.Fatalf("err = %v", err)
	}
}

func TestEnd_EndedPartnershipCannotEndAgain(t *testing.T) {
	state := anEndState()
	state.Status = Ended

	_, err := state.End(Ender{UserID: ownerID, OrganizationID: storeID}, "", endedAt())

	if appErr, _ := errors.AsType[*apperror.Error](err); appErr == nil ||
		appErr.Code() != apperror.CodePartnershipNotActive {
		t.Fatalf("err = %v", err)
	}
}

func TestEnd_ReasonIsLimited(t *testing.T) {
	long := make([]rune, MaxEndReasonLength+1)
	for index := range long {
		long[index] = 'a'
	}

	_, err := anEndState().End(Ender{UserID: ownerID, OrganizationID: storeID}, string(long), endedAt())

	appErr, _ := errors.AsType[*apperror.Error](err)
	if appErr == nil || len(appErr.Fields()) != 1 || appErr.Fields()[0].Field != "reason" {
		t.Fatalf("err = %v", err)
	}
}
