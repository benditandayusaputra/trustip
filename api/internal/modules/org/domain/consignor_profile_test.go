package domain

import (
	"strings"
	"testing"
)

func validConsignorForm() ConsignorProfileForm {
	return ConsignorProfileForm{
		DeliveryRadius:      DeliveryRadius{Given: true, Kilometres: 10},
		DeliveryDays:        []string{"MON", "THU"},
		PreferredStoreTypes: []string{"CAFE", "CANTEEN"},
		Legal:               LegalNumbers{BusinessIDNo: "1234567890123", HalalCertNo: "ID33110000123"},
		CanProvideSample:    true,
	}
}

func TestNewConsignorProfile_FRORG05_ValidFormIsKept(t *testing.T) {
	profile, err := NewConsignorProfile(validConsignorForm())

	if err != nil || profile.DeliveryRadius.Kilometres != 10 || !profile.CanProvideSample {
		t.Fatalf("profile = %+v, err = %v", profile, err)
	}
}

func TestNewConsignorProfile_RSEC46_DeliveryRadiusBoundaries(t *testing.T) {
	cases := map[DeliveryRadius]string{
		{Given: true, Kilometres: 0}:   "OUT_OF_RANGE",
		{Given: true, Kilometres: 1}:   "",
		{Given: true, Kilometres: 500}: "",
		{Given: true, Kilometres: 501}: "OUT_OF_RANGE",
		{}:                             "",
	}
	for radius, want := range cases {
		form := validConsignorForm()
		form.DeliveryRadius = radius

		_, err := NewConsignorProfile(form)

		if got := codeOrEmpty(t, err, "delivery_radius_km"); got != want {
			t.Errorf("%+v: %q, want %q", radius, got, want)
		}
	}
}

func TestNewConsignorProfile_DaysAndStoreTypesComeFromTheirLists(t *testing.T) {
	form := validConsignorForm()
	form.DeliveryDays = []string{"MON", "MON"}
	form.PreferredStoreTypes = []string{"SUPERMARKET"}

	codes := fieldCodes(t, mustFailConsignor(NewConsignorProfile(form)))

	if codes["delivery_days"] != "UNKNOWN_CHOICE" || codes["preferred_store_types"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewConsignorProfile_BRORG07_LegalNumbersAreFreeTextWithALimit(t *testing.T) {
	form := validConsignorForm()
	form.Legal.DistributionPermitNo = strings.Repeat("9", 51)

	if got := fieldCodes(t, mustFailConsignor(NewConsignorProfile(form)))["distribution_permit_no"]; got != "TOO_LONG" {
		t.Fatalf("distribution_permit_no = %q", got)
	}
}

func TestNewPaymentInfo_FRORG07_InfoLimitAndQRFileID(t *testing.T) {
	cash := []string{"CASH"}
	_, longInfo := NewPaymentInfo(PaymentInfoForm{Info: strings.Repeat("a", 501), Methods: cash})
	_, badQR := NewPaymentInfo(PaymentInfoForm{Info: "BCA 123", QRFileID: "qris.png", Methods: cash})
	info, err := NewPaymentInfo(PaymentInfoForm{Info: "Tunai saat antar", Methods: cash})

	if fieldCodes(t, longInfo)["payment_info"] != "TOO_LONG" ||
		fieldCodes(t, badQR)["payment_qr_file_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("long = %v, qr = %v", longInfo, badQR)
	}
	if err != nil || info.Info != "Tunai saat antar" {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
}

func TestNewPaymentInfo_ConsignorChoosesAtLeastOneKnownMethod(t *testing.T) {
	_, none := NewPaymentInfo(PaymentInfoForm{})
	_, unknown := NewPaymentInfo(PaymentInfoForm{Methods: []string{"CASH", "GOPAY"}})
	_, twice := NewPaymentInfo(PaymentInfoForm{Methods: []string{"CASH", "CASH"}})

	if fieldCodes(t, none)["payment_methods"] != "REQUIRED" ||
		fieldCodes(t, unknown)["payment_methods"] != "UNKNOWN_CHOICE" ||
		fieldCodes(t, twice)["payment_methods"] != "UNKNOWN_CHOICE" {
		t.Fatalf("none = %v, unknown = %v, twice = %v", none, unknown, twice)
	}
}

func TestNewPaymentInfo_TransferNeedsTheBankAccount(t *testing.T) {
	transfer := []string{"TRANSFER"}
	_, empty := NewPaymentInfo(PaymentInfoForm{Methods: transfer})
	_, letters := NewPaymentInfo(PaymentInfoForm{Methods: transfer, Bank: BankAccount{
		Name: "BRI", Number: "0123-abc-789", Holder: "Bendi",
	}})
	info, err := NewPaymentInfo(PaymentInfoForm{Methods: transfer, Bank: BankAccount{
		Name: "BRI", Number: "0123 4567 8910", Holder: "Bendi",
	}})

	codes := fieldCodes(t, empty)
	if codes["bank_name"] != "REQUIRED" || codes["bank_account_number"] != "REQUIRED" ||
		codes["bank_account_holder"] != "REQUIRED" {
		t.Fatalf("empty = %v", codes)
	}
	if fieldCodes(t, letters)["bank_account_number"] != "INVALID_CHARACTERS" {
		t.Fatalf("letters = %v", letters)
	}
	if err != nil || info.Bank.Number != "0123 4567 8910" {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
}

func TestNewPaymentInfo_QRISNeedsTheQRImage(t *testing.T) {
	_, missing := NewPaymentInfo(PaymentInfoForm{Methods: []string{"QRIS"}})
	_, err := NewPaymentInfo(PaymentInfoForm{
		Methods: []string{"CASH", "QRIS"}, QRFileID: "0192f0c4-0000-7000-8000-000000000001",
	})

	if fieldCodes(t, missing)["payment_qr_file_id"] != "REQUIRED" || err != nil {
		t.Fatalf("missing = %v, err = %v", missing, err)
	}
}

func mustFailConsignor(_ ConsignorProfile, err error) error {
	return err
}
