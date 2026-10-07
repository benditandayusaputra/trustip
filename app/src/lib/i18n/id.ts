import { legal } from './legal';

export const id = {
	brand: 'Trustip',
	legal,
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
		crossSite: 'Permintaan dari situs lain ditolak. Buka Trustip langsung lalu coba lagi.',
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
		login: {
			title: 'Masuk',
			intro: 'Masuk untuk melanjutkan pencatatan bersama mitra.',
			email: 'Email',
			password: 'Kata sandi',
			submit: 'Masuk',
			noAccount: 'Belum punya akun?',
			register: 'Daftar',
			offline: 'Masuk butuh koneksi internet. Sambungkan lalu coba lagi.'
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
	account: {
		greeting: (name: string) => `Halo, ${name}`,
		emptyDashboard: 'Dasbor akan terisi setelah Anda punya mitra dan mulai mencatat titipan.',
		menu: 'Akun',
		logOut: 'Keluar',
		logOutEverywhere: 'Keluar dari semua perangkat',
		confirmEverywhereTitle: 'Keluar dari semua perangkat?',
		confirmEverywhereBody:
			'Semua sesi di HP dan komputer lain akan berakhir. Anda perlu masuk lagi di setiap perangkat.',
		cancel: 'Batal',
		loadFailed: 'Akun tidak bisa dimuat.'
	},
	onboarding: {
		title: 'Daftarkan usaha Anda',
		progress: (step: number, total: number) => `Langkah ${String(step)} dari ${String(total)}`,
		steps: { type: 'Tipe usaha', profile: 'Profil usaha', location: 'Lokasi usaha' },
		types: {
			CONSIGNOR: { label: 'Penitip', body: 'Saya membuat produk dan menitipkannya ke toko.' },
			STORE: { label: 'Toko', body: 'Saya menerima titipan produk untuk dijual.' }
		},
		typeFixed: 'Tipe usaha tidak bisa diubah setelah disimpan.',
		name: 'Nama usaha',
		description: 'Deskripsi singkat',
		descriptionHint: 'Opsional. Ceritakan produk atau toko Anda.',
		whatsapp: 'Nomor WhatsApp usaha',
		whatsappHint: 'Contoh: 0812 3456 7890',
		province: 'Provinsi',
		city: 'Kabupaten atau kota',
		choose: 'Pilih',
		district: 'Kecamatan',
		address: 'Alamat',
		addressHint: 'Opsional. Alamat lengkap penitip hanya terlihat oleh toko mitra.',
		mapLabel: 'Peta lokasi usaha. Geser peta dengan tombol panah sampai pin tepat di lokasi.',
		mapHint: 'Opsional. Geser peta sampai pin tepat di lokasi usaha Anda.',
		useMyLocation: 'Gunakan lokasi saya',
		locationDenied: 'Lokasi perangkat tidak bisa dipakai. Geser peta untuk menandai lokasi.',
		pinned: 'Lokasi ditandai.',
		consignorPrivacy:
			'Untuk penitip, Cari Mitra hanya menampilkan titik yang dibulatkan sekitar 500 meter.',
		regionsFailed: {
			title: 'Daftar wilayah tidak bisa dimuat.',
			body: 'Periksa koneksi lalu coba lagi.'
		},
		back: 'Kembali',
		next: 'Lanjut',
		submit: 'Simpan dan mulai',
		founded: 'Usaha Anda terdaftar.',
		offline: 'Menyimpan profil butuh koneksi internet.',
		validation: {
			choose: 'Pilih salah satu.',
			maxLength: (count: number) => `Maksimal ${String(count)} karakter.`
		}
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
