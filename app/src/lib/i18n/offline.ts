export const offline = {
	offline: {
		title: 'Tanpa koneksi',
		body: 'Halaman ini belum tersimpan di perangkat. Sambungkan internet lalu coba lagi.',
		retry: 'Coba lagi',
		ribbon: 'Anda sedang luring. Yang tampil adalah data terakhir yang tersimpan di perangkat.',
		staleRibbon:
			'Server belum terjangkau. Yang tampil adalah data terakhir yang tersimpan di perangkat.',
		storedAt: (when: string) => `Data terakhir: ${when}.`
	}
};
