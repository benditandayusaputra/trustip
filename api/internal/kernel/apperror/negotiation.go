package apperror

import "net/http"

const (
	CodeProposalLimitReached Code = "PROPOSAL_LIMIT_REACHED"
	CodeProposalCooldown     Code = "PROPOSAL_COOLDOWN"
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
	}
}
