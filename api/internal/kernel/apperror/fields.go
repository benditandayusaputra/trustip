package apperror

import "strings"

type FieldCode string

const (
	FieldTooShort          FieldCode = "TOO_SHORT"
	FieldTooLong           FieldCode = "TOO_LONG"
	FieldTooCommon         FieldCode = "TOO_COMMON"
	FieldContainsIdentity  FieldCode = "CONTAINS_IDENTITY"
	FieldInvalidEmail      FieldCode = "INVALID_EMAIL"
	FieldInvalidWhatsApp   FieldCode = "INVALID_WHATSAPP"
	FieldInvalidCharacters FieldCode = "INVALID_CHARACTERS"
	FieldConsentRequired   FieldCode = "CONSENT_REQUIRED"
	FieldEmailTaken        FieldCode = "EMAIL_TAKEN"
)

type Param struct {
	Name  string
	Value string
}

func fieldMessages() map[FieldCode]string {
	return map[FieldCode]string{
		FieldTooShort:          "Minimal {min} karakter.",
		FieldTooLong:           "Maksimal {max} karakter.",
		FieldTooCommon:         "Kata sandi ini terlalu umum. Pilih yang lebih sulit ditebak.",
		FieldContainsIdentity:  "Kata sandi tidak boleh memuat nama atau email Anda.",
		FieldInvalidEmail:      "Format email tidak sah.",
		FieldInvalidWhatsApp:   "Nomor WhatsApp harus 10 sampai 15 angka, misalnya 0812 3456 7890.",
		FieldInvalidCharacters: "Isian memuat karakter yang tidak diizinkan.",
		FieldConsentRequired:   "Setujui kebijakan privasi dan ketentuan untuk mendaftar.",
		FieldEmailTaken:        "Email sudah terdaftar. Silakan masuk.",
	}
}

func FieldCodes() []FieldCode {
	codes := make([]FieldCode, 0, len(fieldMessages()))
	for code := range fieldMessages() {
		codes = append(codes, code)
	}
	return codes
}

func InvalidField(field string, code FieldCode, params ...Param) FieldError {
	message := fieldMessages()[code]
	for _, param := range params {
		message = strings.ReplaceAll(message, "{"+param.Name+"}", param.Value)
	}
	return FieldError{Field: field, Code: string(code), Message: message}
}
