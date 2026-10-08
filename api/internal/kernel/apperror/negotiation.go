package apperror

import "net/http"

const (
	CodeProposalLimitReached  Code = "PROPOSAL_LIMIT_REACHED"
	CodeProposalCooldown      Code = "PROPOSAL_COOLDOWN"
	CodeNeedLimitReached      Code = "NEED_LIMIT_REACHED"
	CodeNeedNotOpen           Code = "NEED_NOT_OPEN"
	CodeOrganizationSuspended Code = "ORGANIZATION_SUSPENDED"
)

func negotiationEntries() map[Code]entry {
	return map[Code]entry{
		CodeProposalLimitReached: {
			KindConflict, http.StatusConflict, "Batas 10 ajuan baru per hari tercapai. Coba lagi besok.",
		},
		CodeProposalCooldown: {
			KindConflict, http.StatusConflict,
			"{organisasi} menolak ajuan Anda. Ajuan baru bisa dikirim lagi mulai {tanggal}.",
		},
		CodeNeedLimitReached: {
			KindConflict, http.StatusConflict,
			"Batas 5 Kebutuhan terbuka tercapai. Tutup atau tandai terpenuhi salah satunya dulu.",
		},
		CodeNeedNotOpen: {KindConflict, http.StatusConflict, "Kebutuhan ini sudah ditutup."},
		CodeOrganizationSuspended: {
			KindForbidden, http.StatusForbidden,
			"Usaha Anda sedang ditangguhkan sehingga belum bisa mengirim Ajuan. " +
				"Kewajiban yang berjalan tetap bisa diselesaikan.",
		},
	}
}
