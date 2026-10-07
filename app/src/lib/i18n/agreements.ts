export const agreements = {
	agreement: {
		title: 'Kesepakatan Dagang',
		version: (version: number) => `Versi ${String(version)}`,
		loadFailed: 'Kesepakatan tidak bisa dimuat.',
		products: 'Produk dan harga',
		noProducts: 'Belum ada produk di Kesepakatan ini.',
		stopped: 'Dihentikan',
		trail: 'Jejak persetujuan',
		proposedBy: (name: string, organization: string, number: string) =>
			`Diusulkan ${name} (${organization}) lewat Ajuan ${number}.`,
		acceptedBy: (name: string, organization: string, date: string) =>
			`Disetujui ${name} (${organization}) pada ${date}.`,
		backToProposal: 'Lihat Ajuan sumber',
		history: 'Riwayat versi',
		historyFailed: 'Riwayat versi tidak bisa dimuat.',
		openVersion: (version: number) => `Buka versi ${String(version)}`,
		superseded: 'Versi lama',
		active: 'Berlaku',
		changeKinds: { ADDED: 'Ditambah', UPDATED: 'Diubah', STOPPED: 'Dihentikan' },
		termChange: (label: string, before: string, after: string) =>
			`${label}: ${before} menjadi ${after}`,
		amend: 'Ajukan perubahan',
		emptyNotes: '(kosong)'
	},
	seal: {
		valid: 'Segel sah',
		invalid: 'Segel tidak sah',
		validBody: 'Isi dokumen sama dengan yang disegel dan tanda tangannya cocok.',
		invalidBody:
			'Isi dokumen berbeda dari yang disegel atau tanda tangannya tidak cocok. Hubungi pengelola Trustip.',
		events: 'Jejak segel',
		code: 'Kode verifikasi',
		eventNames: {
			AGREEMENT_ACCEPTED: 'Kesepakatan disetujui',
			DROP_RECORDED: 'Titipan Masuk dicatat',
			DROP_CONFIRMED: 'Titipan Masuk dikonfirmasi',
			DROP_CANCELLED: 'Titipan Masuk dibatalkan',
			COUNT_RECORDED: 'Hitung Sisa dicatat',
			COUNT_CONFIRMED: 'Hitung Sisa dikonfirmasi',
			SETTLEMENT_ISSUED: 'Setoran terbit',
			PAYMENT_RECORDED: 'Pembayaran dicatat',
			PAYMENT_CONFIRMED: 'Pembayaran dikonfirmasi',
			PAYMENT_REJECTED: 'Pembayaran ditolak',
			PARTNERSHIP_ENDED: 'Kemitraan berakhir'
		} as Record<string, string | undefined>,
		by: (user: string, organization: string) => `${user}, ${organization}`
	}
};
