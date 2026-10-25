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
	}
};
