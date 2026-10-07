export const partnerships = {
	partners: {
		title: 'Mitra',
		statuses: { ACTIVE: 'Aktif', ENDED: 'Berakhir' },
		filters: { ACTIVE: 'Aktif', ENDED: 'Berakhir', ALL: 'Semua' },
		filterLabel: 'Saring mitra',
		types: { CONSIGNOR: 'Penitip', STORE: 'Toko' },
		empty: 'Belum ada mitra. Undang mitra lama lewat tautan atau tunggu Ajuan masuk.',
		invite: 'Undang mitra',
		loadFailed: 'Daftar mitra tidak bisa dimuat.',
		loadMore: 'Muat lebih banyak',
		since: (date: string) => `Bermitra sejak ${date}`,
		endedOn: (date: string) => `Berakhir ${date}`,
		tabs: {
			summary: 'Ringkasan',
			agreement: 'Kesepakatan',
			stock: 'Stok',
			drops: 'Titipan Masuk',
			counts: 'Hitung Sisa',
			settlements: 'Setoran'
		},
		tabsLabel: 'Bagian kemitraan',
		comingStock: 'Stok di toko mitra akan tampil di sini setelah titipan pertama dicatat.',
		comingDrops: 'Titipan Masuk dengan mitra ini akan tampil di sini.',
		comingCounts: 'Hitung Sisa dengan mitra ini akan tampil di sini.',
		comingSettlements: 'Setoran dengan mitra ini akan tampil di sini.',
		contact: 'Hubungi lewat WhatsApp',
		address: 'Alamat',
		area: 'Wilayah',
		catalog: 'Katalog mitra',
		notAgreed: 'Belum ada di Kesepakatan',
		stoppedItem: 'Dihentikan di Kesepakatan'
	}
};
