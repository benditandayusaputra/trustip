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
	auth: {
		register: {
			title: 'Daftar akun',
			intro: 'Satu akun untuk mencatat titipan, hitungan, dan setoran bersama mitra.',
			fullName: 'Nama lengkap',
			email: 'Email',
			password: 'Kata sandi',
			passwordHint: 'Minimal 10 karakter. Jangan memuat nama atau email Anda.',
			whatsapp: 'Nomor WhatsApp',
			whatsappHint: 'Contoh: 0812 3456 7890',
			consentPrefix: 'Saya menyetujui',
			consentJoin: 'dan',
			submit: 'Daftar',
			hasAccount: 'Sudah punya akun?',
			signIn: 'Masuk',
			created: 'Akun dibuat. Silakan masuk.',
			offline: 'Daftar butuh koneksi internet. Sambungkan lalu coba lagi.'
		},
		validation: {
			required: 'Wajib diisi.',
			minLength: (count: number) => `Minimal ${String(count)} karakter.`,
			consent: 'Setujui kebijakan privasi dan ketentuan untuk mendaftar.',
			captcha: 'Selesaikan verifikasi keamanan dulu.'
		},
		retry: 'Coba lagi',
		opensInNewTab: '(tab baru)'
	},
	captcha: {
		label: 'Verifikasi keamanan',
		loadFailed: 'Verifikasi keamanan tidak bisa dimuat. Periksa koneksi lalu muat ulang halaman.'
	},
	form: {
		saving: 'Menyimpan',
		decrease: (label: string) => `Kurangi ${label}`,
		increase: (label: string) => `Tambah ${label}`,
		currencyPrefix: 'Rp'
	}
};
