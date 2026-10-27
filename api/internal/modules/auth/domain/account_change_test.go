package domain

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestNewProfileChange_NormalizesWhatsAppAndKeepsTheAvatar(t *testing.T) {
	form := ProfileForm{FullName: "Sari Wulandari", WhatsApp: "0812 3456 7890", AvatarFileID: "file-1"}

	change, err := NewProfileChange(form)
	if err != nil {
		t.Fatalf("new profile change: %v", err)
	}

	if change.FullName != "Sari Wulandari" || change.WhatsApp.String() != "6281234567890" ||
		change.AvatarFileID != "file-1" {
		t.Fatalf("change = %+v", change)
	}
}

func TestNewProfileChange_ReportsNameAndWhatsAppTogether(t *testing.T) {
	_, err := NewProfileChange(ProfileForm{FullName: "S", WhatsApp: "12"})

	codes := problemCodes(t, err)
	if codes["full_name"] != string(apperror.FieldTooShort) ||
		codes["whatsapp"] != string(apperror.FieldInvalidWhatsApp) {
		t.Fatalf("codes = %v", codes)
	}
}

func TestCheckNewPassword_AcceptsAStrongDifferentPassword(t *testing.T) {
	form := PasswordChangeForm{Current: "Lama-sandi-2026", New: acceptableInput}

	if err := CheckNewPassword(form, sari(), NewPasswordPolicy()); err != nil {
		t.Fatalf("check new password: %v", err)
	}
}

func TestCheckNewPassword_RequiresTheCurrentPassword(t *testing.T) {
	form := PasswordChangeForm{New: acceptableInput}

	codes := problemCodes(t, CheckNewPassword(form, sari(), NewPasswordPolicy()))
	if codes["current_password"] != string(apperror.FieldRequired) {
		t.Fatalf("codes = %v", codes)
	}
}

func TestCheckNewPassword_RefusesTheSamePasswordAgain(t *testing.T) {
	form := PasswordChangeForm{Current: acceptableInput, New: acceptableInput}

	codes := problemCodes(t, CheckNewPassword(form, sari(), NewPasswordPolicy()))
	if codes["new_password"] != string(apperror.FieldSameAsCurrent) {
		t.Fatalf("codes = %v", codes)
	}
}

func TestCheckNewPassword_AppliesTheRegistrationPolicyToTheNewPassword(t *testing.T) {
	cases := map[string]apperror.FieldCode{
		"pendek":               apperror.FieldTooShort,
		"password123":          apperror.FieldTooCommon,
		"wulandari-sari-2026!": apperror.FieldContainsIdentity,
	}
	for password, want := range cases {
		form := PasswordChangeForm{Current: "Lama-sandi-2026", New: password}

		codes := problemCodes(t, CheckNewPassword(form, sari(), NewPasswordPolicy()))
		if codes["new_password"] != string(want) {
			t.Errorf("%s: codes = %v, want %s", password, codes, want)
		}
	}
}

func TestWrongCurrentPassword_PointsAtTheCurrentPasswordField(t *testing.T) {
	codes := problemCodes(t, WrongCurrentPassword())

	if codes["current_password"] != string(apperror.FieldWrongCurrent) {
		t.Fatalf("codes = %v", codes)
	}
}

func sari() PasswordOwner {
	return PasswordOwner{FullName: "Sari Wulandari", Email: "sari@contoh.id"}
}
