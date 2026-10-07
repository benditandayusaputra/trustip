package domain

import (
	"strconv"
	"strings"
	"testing"
)

const categoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"

func validStoreForm() StoreProfileForm {
	return StoreProfileForm{
		StoreType:   "GROCERY_STALL",
		CategoryIDs: []string{categoryID},
		Acceptance:  Acceptance{AcceptsCommission: true},
		Commission:  CommissionRange{Given: true, MinBPS: 1000, MaxBPS: 2500},
		Conditions:  StoreConditions{RequiresHalal: true, OtherRequirements: "Kemasan berlabel."},
	}
}

func TestNewStoreProfile_FRORG04_ValidFormIsKept(t *testing.T) {
	profile, err := NewStoreProfile(validStoreForm())

	if err != nil || profile.StoreType != "GROCERY_STALL" || profile.Commission.MaxBPS != 2500 {
		t.Fatalf("profile = %+v, err = %v", profile, err)
	}
}

func TestNewStoreProfile_StoreTypeMayStayUnchosenButMustBeKnown(t *testing.T) {
	unchosen := validStoreForm()
	unchosen.StoreType = ""
	unknown := validStoreForm()
	unknown.StoreType = "SUPERMARKET"

	if _, err := NewStoreProfile(unchosen); err != nil {
		t.Fatalf("unchosen: %v", err)
	}
	if got := fieldCodes(t, mustFail(NewStoreProfile(unknown)))["store_type"]; got != "UNKNOWN_CHOICE" {
		t.Fatalf("store_type = %q", got)
	}
}

func TestNewStoreProfile_BRAGR03_CommissionRangeStaysWithinFiftyPercentAndInOrder(t *testing.T) {
	cases := map[CommissionRange]string{
		{Given: true, MinBPS: 0, MaxBPS: 5000}:    "",
		{Given: true, MinBPS: 0, MaxBPS: 5001}:    "OUT_OF_RANGE",
		{Given: true, MinBPS: -1, MaxBPS: 100}:    "OUT_OF_RANGE",
		{Given: true, MinBPS: 2000, MaxBPS: 1000}: "OUT_OF_RANGE",
		{}: "",
	}
	for commission, want := range cases {
		form := validStoreForm()
		form.Commission = commission

		_, err := NewStoreProfile(form)

		if got := codeOrEmpty(t, err, "commission_max_bps"); got != want {
			t.Errorf("%+v: %q, want %q", commission, got, want)
		}
	}
}

func TestNewStoreProfile_FRORG04_AtMostFivePhotosWithoutRepeats(t *testing.T) {
	six := validStoreForm()
	six.PhotoFileIDs = []string{fileID(1), fileID(2), fileID(3), fileID(4), fileID(5), fileID(6)}
	repeated := validStoreForm()
	repeated.PhotoFileIDs = []string{fileID(1), fileID(1)}
	five := validStoreForm()
	five.PhotoFileIDs = six.PhotoFileIDs[:5]

	if got := fieldCodes(t, mustFail(NewStoreProfile(six)))["photo_file_ids"]; got != "TOO_MANY" {
		t.Errorf("six = %q", got)
	}
	if got := fieldCodes(t, mustFail(NewStoreProfile(repeated)))["photo_file_ids"]; got != "UNKNOWN_CHOICE" {
		t.Errorf("repeated = %q", got)
	}
	if _, err := NewStoreProfile(five); err != nil {
		t.Errorf("five: %v", err)
	}
}

func TestNewStoreProfile_CategoriesMustBeIDsWithoutRepeats(t *testing.T) {
	form := validStoreForm()
	form.CategoryIDs = []string{categoryID, categoryID}

	if got := fieldCodes(t, mustFail(NewStoreProfile(form)))["accepted_category_ids"]; got != "UNKNOWN_CHOICE" {
		t.Fatalf("accepted_category_ids = %q", got)
	}
}

func TestNewStoreProfile_RSEC46_TextLimits(t *testing.T) {
	form := validStoreForm()
	form.Conditions.OtherRequirements = strings.Repeat("a", 501)
	form.OpeningHours = strings.Repeat("a", 201)

	codes := fieldCodes(t, mustFail(NewStoreProfile(form)))

	if codes["other_requirements"] != "TOO_LONG" || codes["opening_hours"] != "TOO_LONG" {
		t.Fatalf("codes = %v", codes)
	}
}

func fileID(n int) string {
	return "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a9" + strconv.Itoa(n)
}

func mustFail(_ StoreProfile, err error) error {
	return err
}
