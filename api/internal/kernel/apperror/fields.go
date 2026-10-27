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
	FieldRequired          FieldCode = "REQUIRED"
	FieldUnknownChoice     FieldCode = "UNKNOWN_CHOICE"
	FieldOutOfRange        FieldCode = "OUT_OF_RANGE"
	FieldFileTooLarge      FieldCode = "FILE_TOO_LARGE"
	FieldInvalidImage      FieldCode = "INVALID_IMAGE"
	FieldImageTooLarge     FieldCode = "IMAGE_TOO_LARGE"
	FieldTooMany           FieldCode = "TOO_MANY"
	FieldTaken             FieldCode = "TAKEN"
	FieldExceedsRetail     FieldCode = "EXCEEDS_RETAIL_PRICE"
	FieldCountExceedsStock FieldCode = "COUNT_EXCEEDS_STOCK"
	FieldNotInAgreement    FieldCode = "PRODUCT_NOT_IN_AGREEMENT"
	FieldExpiryBeforeDrop  FieldCode = "EXPIRY_BEFORE_DROP"
	FieldInvalidDate       FieldCode = "INVALID_DATE"
	FieldBeforeLastDoc     FieldCode = "BEFORE_LAST_DOCUMENT"
	FieldPeriodSettled     FieldCode = "PERIOD_ALREADY_SETTLED"
	FieldAmountMismatch    FieldCode = "PAYMENT_AMOUNT_MISMATCH"
	FieldProofRequired     FieldCode = "PROOF_REQUIRED"
	FieldContactInText     FieldCode = "CONTACT_NOT_ALLOWED"
	FieldWrongCurrent      FieldCode = "WRONG_PASSWORD"
	FieldSameAsCurrent     FieldCode = "SAME_AS_CURRENT"
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
		FieldRequired:          "Wajib diisi.",
		FieldUnknownChoice:     "Pilih salah satu dari daftar.",
		FieldOutOfRange:        "Nilai di luar rentang yang diizinkan.",
		FieldFileTooLarge:      "Ukuran berkas maksimal {max} MB.",
		FieldInvalidImage:      "Berkas harus gambar JPG, PNG, atau WebP.",
		FieldImageTooLarge:     "Gambar maksimal 4096 piksel di tiap sisi.",
		FieldTooMany:           "Maksimal {max} pilihan.",
		FieldTaken:             "Sudah dipakai. Pilih yang lain.",
		FieldExceedsRetail:     "Harga setor tidak boleh melebihi harga jual.",
		FieldCountExceedsStock: "Sisa melebihi stok tercatat ({n}). Catat dulu titipan masuk yang belum dicatat.",
		FieldNotInAgreement:    "Produk ini belum ada di kesepakatan. Usulkan dulu lewat Ajuan.",
		FieldExpiryBeforeDrop:  "Tanggal kedaluwarsa tidak boleh sebelum tanggal titip.",
		FieldInvalidDate:       "Tulis tanggal dengan format TTTT-BB-HH.",
		FieldBeforeLastDoc:     "Waktu ini lebih awal dari dokumen terakhir dengan mitra ini.",
		FieldPeriodSettled:     "Tanggal ini sudah masuk Setoran {no}.",
		FieldAmountMismatch:    "Jumlah harus sama dengan sisa Setoran ({Rp}).",
		FieldProofRequired:     "Unggah bukti untuk pembayaran transfer atau QRIS.",
		FieldWrongCurrent:      "Kata sandi saat ini salah.",
		FieldSameAsCurrent:     "Kata sandi baru harus berbeda dari yang sekarang.",
		FieldContactInText: "Nomor telepon dan tautan tidak boleh ada di teks. " +
			"Calon mitra menghubungi lewat tombol yang tersedia.",
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
