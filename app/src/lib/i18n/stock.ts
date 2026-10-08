export const stock = {
	stock: {
		title: 'Stok',
		intro:
			'Stok tercatat adalah jumlah semua titipan masuk dikurangi yang terjual dan dikembalikan.',
		loadFailed: 'Stok tidak bisa dimuat.',
		empty: 'Belum ada stok. Stok bertambah setelah Titipan Masuk dicatat.',
		flags: { SAFE: 'Aman', NEAR_EXPIRY: 'Mendekati kedaluwarsa', EXPIRED: 'Kedaluwarsa' },
		balance: 'Stok tercatat',
		total: (units: string, unit: string) => `${units} ${unit}`,
		nearestExpiry: (date: string) => `Kedaluwarsa terdekat ${date}`,
		batches: (name: string) => `Batch ${name}`,
		batch: 'Batch',
		expiry: 'Kedaluwarsa',
		age: 'Umur',
		ageDays: (days: number) => `${String(days)} hari`,
		remaining: 'Sisa',
		flag: 'Tanda',
		noExpiry: 'Tanpa tanggal',
		openCard: (name: string) => `Kartu stok ${name}`,
		cardTitle: 'Kartu stok',
		cardFailed: 'Kartu stok tidak bisa dimuat.',
		cardEmpty: 'Belum ada mutasi stok untuk produk ini.',
		types: {
			RECEIVE: 'Titipan masuk',
			SALE: 'Terjual',
			RETURN: 'Dikembalikan',
			ADJUSTMENT: 'Penyesuaian'
		},
		balanceAfter: (units: string) => `Saldo ${units}`,
		loadMore: 'Muat lebih banyak',
		partner: 'Mitra',
		openPartner: 'Buka halaman mitra'
	}
};
