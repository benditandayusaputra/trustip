import { agreements } from './agreements';
import { counts } from './counts';
import { directory } from './directory';
import { drops } from './drops';
import { legal } from './legal';
import { needs } from './needs';
import { offline } from './offline';
import { organization } from './organization';
import { partnerships } from './partnerships';
import { products } from './products';
import { proposals } from './proposals';
import { reports } from './reports';
import { settlements } from './settlements';
import { stock } from './stock';

export const id = {
	brand: 'Trustip',
	legal,
	...agreements,
	...counts,
	...directory,
	...drops,
	...needs,
	...offline,
	...organization,
	...partnerships,
	...products,
	...proposals,
	...reports,
	...settlements,
	...stock,
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
	notifications: {
		title: 'Pemberitahuan',
		bell: (count: number) =>
			count > 0 ? `Pemberitahuan, ${String(count)} belum dibaca` : 'Pemberitahuan',
		empty: 'Belum ada pemberitahuan. Kabar dari mitra akan muncul di sini.',
		backToDashboard: 'Kembali ke dasbor',
		markAll: 'Tandai semua dibaca',
		loadMore: 'Muat lebih banyak',
		loadFailed: 'Pemberitahuan tidak bisa dimuat.',
		unread: 'Belum dibaca',
		markedAll: 'Semua pemberitahuan ditandai dibaca.'
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
