export const offline = {
	offline: {
		title: 'Tanpa koneksi',
		body: 'Halaman ini belum tersimpan di perangkat. Sambungkan internet lalu coba lagi.',
		retry: 'Coba lagi',
		ribbon: 'Anda sedang luring. Yang tampil adalah data terakhir yang tersimpan di perangkat.',
		staleRibbon:
			'Server belum terjangkau. Yang tampil adalah data terakhir yang tersimpan di perangkat.',
		storedAt: (when: string) => `Data terakhir: ${when}.`,
		queued: 'Tersimpan di perangkat, menunggu sinyal.',
		recordedOffline: 'Dicatat luring',
		bothTimes: (recordedAt: string, receivedAt: string) =>
			`Dicatat ${recordedAt} · Terkirim ${receivedAt}`
	},
	install: {
		title: 'Pasang Trustip',
		body: 'Pasang Trustip di layar utama supaya bisa dibuka cepat dan tetap mencatat tanpa sinyal.',
		iosHint:
			'Pasang Trustip di iPhone: ketuk tombol Bagikan di Safari, lalu pilih Tambahkan ke Layar Utama.',
		install: 'Pasang',
		later: 'Nanti'
	},
	outbox: {
		title: 'Antrean',
		intro:
			'Titipan Masuk dan Hitung Sisa yang disimpan tanpa sinyal menunggu di sini, lalu terkirim sendiri saat sinyal kembali.',
		oldWarning: (count: number) =>
			`${String(count)} catatan sudah menunggu lebih dari 24 jam. Sambungkan internet lalu tekan Kirim sekarang.`,
		iosNote:
			'Di iPhone, catatan terkirim saat app dibuka kembali karena iPhone tidak mengirim di latar belakang.',
		badge: (count: number) => `Antrean, ${String(count)} catatan menunggu`,
		empty: 'Tidak ada catatan yang menunggu.',
		backToApp: 'Kembali ke dasbor',
		loadFailed: 'Antrean tidak bisa dibaca.',
		sendNow: 'Kirim sekarang',
		kinds: { DROP: 'Titipan Masuk', COUNT: 'Hitung Sisa' },
		statuses: {
			PENDING: 'Menunggu sinyal',
			SENDING: 'Mengirim',
			SENT: 'Terkirim',
			HELD: 'Tertahan',
			NEEDS_REVIEW: 'Perlu ditinjau'
		},
		recordedAt: (when: string) => `Disimpan ${when}`,
		heldHint: 'Menunggu catatan sebelumnya untuk mitra yang sama ditinjau.',
		lastError: (message: string) => `Galat terakhir: ${message}`,
		openDocument: (number: string) => `Buka ${number}`,
		review: 'Tinjau',
		reviewTitle: 'Tinjau catatan',
		reviewIntro:
			'Server menolak catatan ini. Bandingkan isian Anda dengan keadaan terbaru, lalu perbaiki atau buang.',
		reviewMissing: 'Catatan ini sudah tidak ada di antrean perangkat ini.',
		reviewFailed: 'Keadaan terbaru tidak bisa dimuat.',
		backToOutbox: 'Kembali ke Antrean',
		yourInput: 'Isian Anda',
		fix: 'Perbaiki dan kirim ulang',
		product: 'Produk',
		seenStock: 'Stok yang Anda lihat',
		latestStock: 'Stok tercatat sekarang',
		countedQty: 'Sisa yang Anda isi',
		returnedQty: 'Dikembalikan',
		qty: 'Jumlah',
		expiry: 'Kedaluwarsa',
		agreed: 'Di Kesepakatan berlaku',
		agreedYes: 'Ya',
		agreedNo: 'Tidak',
		noStockNow: 'Produk tanpa stok sekarang',
		versions: (used: number, active: number) =>
			`Anda memakai Kesepakatan versi ${String(used)}. Versi yang berlaku sekarang ${String(active)}.`,
		revising:
			'Anda memperbaiki catatan dari antrean. Stok dan Kesepakatan di formulir ini sudah yang terbaru.',
		discard: 'Buang',
		discardTitle: 'Buang catatan ini?',
		discardBody: 'Catatan dihapus dari perangkat dan tidak akan dikirim.',
		cancel: 'Batal',
		discarded: 'Catatan dibuang.',
		sent: (count: number) =>
			count > 0 ? `${String(count)} catatan terkirim.` : 'Tidak ada catatan yang perlu dikirim.',
		retryLater: 'Belum terkirim. Trustip mencoba lagi sendiri saat sinyal ada.',
		signedOut: 'Sesi berakhir. Masuk lagi untuk mengirim catatan.',
		leaveTitle: (count: number) => `Masih ada ${String(count)} catatan belum terkirim`,
		leaveBody:
			'Pilih apa yang dilakukan dengan catatan itu. Data lain di perangkat ini tetap dihapus saat keluar.',
		sendFirst: 'Kirim dulu',
		keepOnDevice: 'Simpan di perangkat',
		discardAll: 'Buang catatan',
		stillWaiting: (count: number) =>
			`Masih ada ${String(count)} catatan belum terkirim. Coba lagi saat sinyal ada.`
	}
};
