export const directory = {
	placePicker: {
		label: 'Titik asal',
		kinds: {
			ORGANIZATION: 'Lokasi usaha saya',
			HERE: 'Lokasi saya sekarang',
			CITY: 'Kota tertentu'
		},
		province: 'Provinsi',
		city: 'Kota atau kabupaten',
		choose: 'Pilih',
		locating: 'Membaca lokasi Anda...',
		locateFailed: 'Lokasi tidak bisa dibaca. Izinkan akses lokasi atau pilih kota.'
	},
	directory: {
		storesTitle: 'Cari toko',
		filters: 'Saring toko',
		search: 'Cari nama atau deskripsi toko',
		radius: 'Jarak',
		radii: { '2': '2 km', '5': '5 km', '10': '10 km', '25': '25 km', all: 'Semua jarak' },
		sort: 'Urutkan',
		sorts: {
			NEAREST: 'Terdekat',
			BEST_MATCH: 'Paling cocok',
			REPUTATION: 'Reputasi',
			NEWEST: 'Terbaru',
			CITY: 'Kota lalu terbaru'
		},
		storeType: 'Jenis toko',
		anyStoreType: 'Semua jenis toko',
		category: 'Kategori diterima',
		ownCategories: 'Kategori produk saya',
		anyCategory: 'Semua kategori',
		pricingModel: 'Model kesepakatan',
		anyModel: 'Semua model',
		models: { FIXED_DEPOSIT: 'Harga Setor', COMMISSION: 'Komisi' },
		maxCommission: 'Komisi maksimal (%)',
		withoutPermit: 'Tanpa syarat izin edar',
		withoutHalal: 'Tanpa syarat halal',
		expanded: (radiusKm: number | null) =>
			radiusKm === null ? 'Diperluas ke semua jarak' : `Diperluas ke ${String(radiusKm)} km`,
		noOrigin: 'Lokasi usaha belum diatur. Toko diurutkan menurut kota.',
		setLocation: 'Atur lokasi usaha',
		empty: 'Belum ada toko yang cocok dengan saringan ini.',
		clearFilters: 'Hapus saringan',
		failed: 'Daftar toko tidak bisa dimuat.',
		retry: 'Coba lagi',
		offline: 'Cari Mitra butuh koneksi internet.',
		loadMore: 'Muat lagi',
		commissionRange: (minimum: string, maximum: string) => `Komisi ${minimum} sampai ${maximum}`,
		newcomer: 'Baru',
		onTime: (percent: string) => `Bayar tepat waktu ${percent}`,
		labels: { VERY_GOOD: 'Sangat cocok', GOOD: 'Cocok' },
		score: (score: number) => `Skor kecocokan ${String(score)}`,
		reasons: {
			needsCategory: (category: string) => `Butuh ${category}`,
			acceptsCategory: (category: string) => `Terima ${category}`,
			established: 'Sudah bermitra dengan 3 toko',
			preferredStoreType: 'Jenis toko incaran',
			activeRecently: 'Aktif 2 minggu terakhir'
		},
		self: 'Usaha Anda',
		partner: 'Sudah bermitra'
	}
};
