package apperror

import (
	"maps"
	"net/http"
)

type Code string

const (
	CodeValidationFailed     Code = "VALIDATION_FAILED"
	CodeUnknownField         Code = "UNKNOWN_FIELD"
	CodeMalformedRequest     Code = "MALFORMED_REQUEST"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodeUnauthenticated      Code = "UNAUTHENTICATED"
	CodeSessionExpired       Code = "SESSION_EXPIRED"
	CodeSessionRevoked       Code = "SESSION_REVOKED"
	CodeLoginFailed          Code = "LOGIN_FAILED"
	CodeEdgeSignatureInvalid Code = "EDGE_SIGNATURE_INVALID"
	CodeForbidden            Code = "FORBIDDEN"
	CodeCrossSiteRejected    Code = "CROSS_SITE_REJECTED"
	CodeNotFound             Code = "NOT_FOUND"
	CodeCaptchaFailed        Code = "CAPTCHA_FAILED"
	CodeOrganizationExists   Code = "ORGANIZATION_EXISTS"
	CodeProfileIncomplete    Code = "PROFILE_INCOMPLETE"
	CodeInviteWrongOrgType   Code = "INVITE_WRONG_ORG_TYPE"
	CodeProposalExpired      Code = "PROPOSAL_EXPIRED"
	CodeProposalAlreadyOpen  Code = "PROPOSAL_ALREADY_OPEN"
	CodeAlreadyPartners      Code = "ALREADY_PARTNERS"
	CodeNotYourTurn          Code = "NOT_YOUR_TURN"
	CodeProposalDecided      Code = "PROPOSAL_ALREADY_DECIDED"
	CodePartnershipNotActive Code = "PARTNERSHIP_NOT_ACTIVE"
	CodeAgreementChanged     Code = "AGREEMENT_CHANGED"
	CodeIdempotencyKeyReused Code = "IDEMPOTENCY_KEY_REUSED"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeAccountLocked        Code = "ACCOUNT_LOCKED"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
	CodeInternal             Code = "INTERNAL"
)

type Kind int

const (
	KindValidation Kind = iota + 1
	KindUnauthenticated
	KindForbidden
	KindNotFound
	KindConflict
	KindRateLimited
	KindUnavailable
	KindInternal
)

type entry struct {
	kind    Kind
	status  int
	message string
}

func catalog() map[Code]entry {
	entries := requestEntries()
	maps.Copy(entries, accessEntries())
	maps.Copy(entries, outcomeEntries())
	return entries
}

func requestEntries() map[Code]entry {
	return map[Code]entry{
		CodeValidationFailed: {KindValidation, http.StatusUnprocessableEntity, "Periksa kembali isian Anda."},
		CodeUnknownField: {
			KindValidation, http.StatusBadRequest,
			"Permintaan memuat isian yang tidak dikenal.",
		},
		CodeMalformedRequest: {KindValidation, http.StatusBadRequest, "Permintaan tidak bisa dibaca."},
		CodePayloadTooLarge:  {KindValidation, http.StatusRequestEntityTooLarge, "Isi permintaan terlalu besar."},
		CodeUnsupportedMediaType: {
			KindValidation, http.StatusUnsupportedMediaType,
			"Jenis isi permintaan tidak didukung.",
		},
		CodeCaptchaFailed: {KindValidation, http.StatusUnprocessableEntity, "Verifikasi keamanan gagal. Coba lagi."},
		CodeProfileIncomplete: {
			KindValidation, http.StatusUnprocessableEntity, "Lengkapi profil dulu: {daftar}.",
		},
		CodeInviteWrongOrgType: {KindValidation, http.StatusUnprocessableEntity, "Undangan ini untuk {jenis}."},
		CodeIdempotencyKeyReused: {
			KindValidation, http.StatusUnprocessableEntity,
			"Kunci permintaan ini sudah dipakai untuk isi yang berbeda. Muat ulang lalu simpan lagi.",
		},
	}
}

func accessEntries() map[Code]entry {
	return map[Code]entry{
		CodeUnauthenticated: {KindUnauthenticated, http.StatusUnauthorized, "Silakan masuk dulu."},
		CodeSessionExpired: {
			KindUnauthenticated, http.StatusUnauthorized,
			"Sesi Anda berakhir. Silakan masuk lagi.",
		},
		CodeSessionRevoked: {
			KindUnauthenticated, http.StatusUnauthorized,
			"Sesi Anda dicabut demi keamanan. Silakan masuk lagi.",
		},
		CodeEdgeSignatureInvalid: {KindUnauthenticated, http.StatusUnauthorized, "Permintaan tidak sah."},
		CodeLoginFailed:          {KindUnauthenticated, http.StatusUnauthorized, "Email atau kata sandi salah."},
		CodeAccountLocked: {
			KindRateLimited, http.StatusTooManyRequests,
			"Akun dikunci sementara karena terlalu banyak percobaan masuk. Coba lagi dalam {n} menit.",
		},
		CodeForbidden: {KindForbidden, http.StatusForbidden, "Peran Anda tidak diizinkan melakukan ini."},
		CodeCrossSiteRejected: {
			KindForbidden, http.StatusForbidden,
			"Permintaan dari situs lain ditolak. Buka Trustip langsung lalu coba lagi.",
		},
		CodeNotFound: {
			KindNotFound, http.StatusNotFound,
			"Data tidak ditemukan atau Anda tidak punya akses.",
		},
	}
}

func outcomeEntries() map[Code]entry {
	return map[Code]entry{
		CodeOrganizationExists: {
			KindConflict, http.StatusConflict,
			"Akun Anda sudah terhubung ke sebuah usaha.",
		},
		CodeProposalExpired:      {KindConflict, http.StatusConflict, "Ajuan ini sudah kedaluwarsa."},
		CodeNotYourTurn:          {KindConflict, http.StatusConflict, "Menunggu tanggapan dari {organisasi}."},
		CodeProposalDecided:      {KindConflict, http.StatusConflict, "Ajuan ini sudah diputuskan."},
		CodePartnershipNotActive: {KindConflict, http.StatusConflict, "Kemitraan ini sudah tidak aktif."},
		CodeAgreementChanged: {
			KindConflict, http.StatusConflict, "Kesepakatan sudah berubah ke versi {n}. Tinjau lagi catatan ini.",
		},
		CodeProposalAlreadyOpen: {
			KindConflict, http.StatusConflict, "Masih ada Ajuan yang terbuka dengan {organisasi}.",
		},
		CodeAlreadyPartners: {
			KindConflict, http.StatusConflict,
			"Anda sudah bermitra dengan {organisasi}. Ajukan perubahan Kesepakatan.",
		},
		CodeRateLimited: {
			KindRateLimited, http.StatusTooManyRequests,
			"Terlalu banyak percobaan. Coba lagi dalam {n} menit.",
		},
		CodeServiceUnavailable: {
			KindUnavailable, http.StatusServiceUnavailable,
			"Layanan sedang sibuk. Coba lagi sebentar lagi.",
		},
		CodeInternal: {
			KindInternal, http.StatusInternalServerError,
			"Terjadi kesalahan di sistem. Kode: {request_id}.",
		},
	}
}

func Codes() []Code {
	codes := make([]Code, 0, len(catalog()))
	for code := range catalog() {
		codes = append(codes, code)
	}
	return codes
}

func lookup(code Code) entry {
	if found, ok := catalog()[code]; ok {
		return found
	}
	return catalog()[CodeInternal]
}
