export const id = {
	brand: 'Trustip',
	layout: {
		skipToContent: 'Lewati ke konten',
		mainNavigation: 'Navigasi utama',
		publicNavigation: 'Navigasi situs',
		footerNavigation: 'Tautan kebijakan',
		more: 'Lainnya'
	},
	nav: {
		home: 'Beranda',
		search: 'Cari',
		partners: 'Mitra',
		settlements: 'Setoran',
		count: 'Hitung',
		drops: 'Titipan',
		products: 'Produk',
		proposals: 'Ajuan',
		stock: 'Stok',
		reports: 'Laporan',
		needs: 'Kebutuhan',
		notifications: 'Pemberitahuan',
		outbox: 'Antrean',
		settings: 'Pengaturan',
		findStores: 'Cari toko',
		needsBoard: 'Kebutuhan toko',
		signIn: 'Masuk',
		signUp: 'Daftar',
		privacy: 'Kebijakan privasi',
		terms: 'Ketentuan'
	},
	errorPage: {
		notFoundTitle: 'Halaman tidak ditemukan',
		notFoundBody: 'Alamat ini tidak ada atau sudah dipindah.',
		failedTitle: 'Terjadi kesalahan',
		failedBody: 'Muat ulang halaman. Bila tetap gagal, coba lagi beberapa saat lagi.',
		backHome: 'Kembali ke beranda'
	},
	apiErrors: {
		serviceUnavailable: 'Layanan sedang sibuk. Coba lagi sebentar lagi.',
		network: 'Tidak ada koneksi. Periksa sinyal lalu coba lagi.',
		unreadable: 'Jawaban server tidak bisa dibaca. Coba lagi.'
	},
	system: {
		apiVersion: (version: string) => `Versi API ${version}`,
		updateAvailable: 'Versi baru Trustip tersedia.',
		reload: 'Muat ulang'
	},
	offline: {
		title: 'Tanpa koneksi',
		body: 'Halaman ini belum tersimpan di perangkat. Sambungkan internet lalu coba lagi.',
		retry: 'Coba lagi'
	},
	feedback: {
		loading: 'Memuat',
		close: 'Tutup',
		dismissToast: 'Tutup pemberitahuan',
		notifications: 'Pemberitahuan'
	},
	form: {
		saving: 'Menyimpan',
		decrease: (label: string) => `Kurangi ${label}`,
		increase: (label: string) => `Tambah ${label}`,
		currencyPrefix: 'Rp'
	}
};
