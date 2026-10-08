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
	outbox: {
		title: 'Antrean',
		intro:
			'Titipan Masuk dan Hitung Sisa yang disimpan tanpa sinyal menunggu di sini, lalu terkirim sendiri saat sinyal kembali.',
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
