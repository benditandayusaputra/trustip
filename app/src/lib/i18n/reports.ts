export const reports = {
	reports: {
		title: 'Laporan',
		salesTitle: 'Laporan penjualan',
		filters: 'Saring laporan',
		range: 'Rentang',
		presets: {
			TODAY: 'Hari ini',
			LAST_7: '7 hari',
			LAST_30: '30 hari',
			THIS_MONTH: 'Bulan ini',
			LAST_MONTH: 'Bulan lalu',
			CUSTOM: 'Atur sendiri'
		},
		from: 'Dari',
		to: 'Sampai',
		groupBy: 'Kelompokkan',
		groupings: {
			STORE: { PARTNER: 'Per penitip', PRODUCT: 'Per produk', DAY: 'Per hari' },
			CONSIGNOR: { PARTNER: 'Per toko', PRODUCT: 'Per produk', DAY: 'Per hari' }
		},
		soldUnits: 'Unit terjual',
		gross: 'Omzet',
		storeShare: 'Untung toko',
		consignorShare: 'Hak penitip',
		mustPay: 'Harus setor',
		returnedUnits: 'Unit dikembalikan',
		returnedPercent: 'Persen dikembalikan',
		rowLabel: 'Kelompok',
		loadFailed: 'Laporan tidak bisa dimuat.',
		empty: 'Belum ada penjualan pada rentang ini.',
		toCounts: 'Lihat Hitung Sisa'
	}
};
