export const listingReport = {
	listingReport: {
		open: (name: string) => `Laporkan ${name}`,
		title: (name: string) => `Laporkan ${name}`,
		intro: 'Laporan dibaca tim Trustip dan tidak dikirim ke pihak yang dilaporkan.',
		reason: 'Alasan',
		chooseReason: 'Pilih alasan',
		reasons: {
			MISLEADING: 'Informasi menyesatkan atau palsu',
			PROHIBITED_GOODS: 'Barang terlarang atau berbahaya',
			OFFENSIVE: 'Isi tidak pantas',
			SPAM: 'Spam atau promosi berulang',
			OTHER: 'Lainnya'
		},
		note: 'Keterangan',
		noteHint: 'Wajib diisi bila memilih Lainnya.',
		reasonRequired: 'Pilih alasan laporan.',
		noteRequired: 'Tulis keterangan untuk alasan Lainnya.',
		cancel: 'Batal',
		send: 'Kirim laporan',
		sent: 'Laporan terkirim. Tim kami akan meninjaunya.'
	}
};
