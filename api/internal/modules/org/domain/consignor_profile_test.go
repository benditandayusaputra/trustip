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
	_, longInfo := NewPaymentInfo(PaymentInfoForm{Info: strings.Repeat("a", 501)})
	_, badQR := NewPaymentInfo(PaymentInfoForm{Info: "BCA 123", QRFileID: "qris.png"})
	info, err := NewPaymentInfo(PaymentInfoForm{Info: "BCA 1234567 a.n. Darmi"})

	if fieldCodes(t, longInfo)["payment_info"] != "TOO_LONG" ||
		fieldCodes(t, badQR)["payment_qr_file_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("long = %v, qr = %v", longInfo, badQR)
	}
	if err != nil || info.Info != "BCA 1234567 a.n. Darmi" {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
}

func mustFailConsignor(_ ConsignorProfile, err error) error {
	return err
}
