package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const snacksID = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"

func postedAt() time.Time {
	return time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)
}

func validForm() Form {
	return Form{
		Title: "Butuh camilan untuk kantin", CategoryID: snacksID, Description: "Kemasan kecil, harga di bawah 5 ribu.",
		WeeklyQuantity: Amount{Value: 50, IsGiven: true}, MaxRetailPrice: Amount{Value: 5000, IsGiven: true},
		PreferredModel: "COMMISSION", OfferedCommissionBPS: Amount{Value: 2000, IsGiven: true},
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

func TestNewDraft_FRNED01_StorePostsWhatItIsLookingFor(t *testing.T) {
	draft, err := NewDraft(validForm())

	if err != nil || draft.Title != "Butuh camilan untuk kantin" || draft.OfferedCommissionBPS.Value != 2000 {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestNewDraft_BRNED02_TitleAndDescriptionLengths(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(*Form)
		field string
		code  string
	}{
		{"title of 4", func(f *Form) { f.Title = "Roti" }, "title", "TOO_SHORT"},
		{"title of 81", func(f *Form) { f.Title = strings.Repeat("a", MaxTitleLength+1) }, "title", "TOO_LONG"},
		{
			"description of 501", func(f *Form) { f.Description = strings.Repeat("é", MaxDescriptionLength+1) },
			"description", "TOO_LONG",
		},
		{"missing category", func(f *Form) { f.CategoryID = "" }, "category_id", "REQUIRED"},
		{"unknown model", func(f *Form) { f.PreferredModel = "RENT" }, "preferred_model", "UNKNOWN_CHOICE"},
		{
			"commission over 50%", func(f *Form) { f.OfferedCommissionBPS.Value = 5001 }, "offered_commission_bps",
			"OUT_OF_RANGE",
		},
		{"zero per week", func(f *Form) { f.WeeklyQuantity.Value = 0 }, "est_qty_per_week", "OUT_OF_RANGE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			form := validForm()
			tc.edit(&form)

			_, err := NewDraft(form)

			if fieldCodes(t, err)[tc.field] != tc.code {
				t.Fatalf("codes = %v", fieldCodes(t, err))
			}
		})
	}
}

func TestNewDraft_BRNED02_BoundaryLengthsAreAccepted(t *testing.T) {
	form := validForm()
	form.Title = strings.Repeat("a", MinTitleLength)
	form.Description = strings.Repeat("é", MaxDescriptionLength)

	if _, err := NewDraft(form); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestNewDraft_BRNED02_PhoneNumbersAndLinksAreRefused(t *testing.T) {
	texts := []string{
		"Hubungi 0812 3456 7890", "WA +62-812-3456-7890", "Lihat https://contoh.id", "cek www.toko.com",
		"wa.me/6281234567890", "order di tokoku.id",
	}
	for _, text := range texts {
		form := validForm()
		form.Description = text

		if fieldCodes(t, mustFail(NewDraft(form)))["description"] != "CONTACT_NOT_ALLOWED" {
			t.Errorf("%q accepted", text)
		}
	}
}

func TestNewDraft_BRNED02_OrdinaryNumbersStayAllowed(t *testing.T) {
	form := validForm()
	form.Description = "Sekitar 50 bungkus per minggu, harga Rp5.000, isi 100 gram."

	if _, err := NewDraft(form); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestPost_FRNED02_ShowsForThirtyDays(t *testing.T) {
	draft, _ := NewDraft(validForm())

	posting := draft.Post(Author{NeedID: "n1", StoreID: "s1", UserID: "u1"}, postedAt())

	if !posting.ExpiresAt.Equal(postedAt().Add(30 * 24 * time.Hour)) {
		t.Fatalf("expires at %v", posting.ExpiresAt)
	}
}

func openNeed() Need {
	draft, _ := NewDraft(validForm())
	return Need{ID: "n1", StoreID: "s1", Draft: draft, Status: Open, ExpiresAt: postedAt().Add(ShowingPeriod)}
}

func TestStatusAt_BRNED03_ExpiresWhenTheShowingPeriodEnds(t *testing.T) {
	need := openNeed()

	if need.StatusAt(need.ExpiresAt.Add(-time.Second)) != Open || need.StatusAt(need.ExpiresAt) != Expired {
		t.Fatal("unexpected status around the expiry")
	}
}

func TestExtend_FRNED02_OpenOrExpiredNeedRunsAnotherThirtyDays(t *testing.T) {
	need := openNeed()
	later := need.ExpiresAt.Add(48 * time.Hour)

	change, err := need.Extend(later)

	if err != nil || change.Status != Open || !change.ExpiresAt.Equal(later.Add(ShowingPeriod)) {
		t.Fatalf("change = %+v, err = %v", change, err)
	}
}

func TestFulfillAndClose_FRNED06_OnlyAnOpenNeed(t *testing.T) {
	need := openNeed()
	closed := openNeed()
	closed.Status = Closed

	fulfilled, fulfillErr := need.Fulfill(postedAt())
	_, closeErr := closed.Close(postedAt())
	_, extendErr := closed.Extend(postedAt())
	_, expiredErr := need.Close(need.ExpiresAt)

	if fulfillErr != nil || fulfilled.Status != Fulfilled || fulfilled.Previous != Open ||
		apperror.From(closeErr).Code() != apperror.CodeNeedNotOpen ||
		apperror.From(extendErr).Code() != apperror.CodeNeedNotOpen ||
		apperror.From(expiredErr).Code() != apperror.CodeNeedNotOpen {
		t.Fatalf("fulfill = %+v %v, close = %v, extend = %v, expired = %v",
			fulfilled, fulfillErr, closeErr, extendErr, expiredErr)
	}
}

func mustFail[T any](_ T, err error) error {
	return err
}

func TestParseStatus_KnownStatusesOnly(t *testing.T) {
	for _, raw := range []string{"", "OPEN", "FULFILLED", "CLOSED", "EXPIRED"} {
		if _, err := ParseStatus(raw); err != nil {
			t.Errorf("%q: %v", raw, err)
		}
	}
	if fieldCodes(t, mustFail(ParseStatus("LOST")))["status"] != "UNKNOWN_CHOICE" {
		t.Fatal("unknown status accepted")
	}
}

func TestAsOf_BRNED03_ReportsTheStatusAtThatTime(t *testing.T) {
	need := openNeed()

	if need.AsOf(need.ExpiresAt).Status != Expired || need.AsOf(postedAt()).Status != Open {
		t.Fatal("unexpected status")
	}
}
