interface FaqItem {
	question: string;
	answer: string;
}

interface FaqGroup {
	slug: string;
	title: string;
	items: FaqItem[];
}

const groups: FaqGroup[] = [
	{
		slug: 'umum',
		title: 'Umum',
		items: [
			{
				question: 'Apa itu Trustip?',
				answer:
					'Trustip adalah platform titip jual untuk UMKM. Penitip dan toko saling menemukan, menyepakati aturan dagang, lalu mencatat titipan, hitungan sisa, dan setoran di satu catatan yang sama.'
			},
			{
				question: 'Siapa yang cocok memakai Trustip?',
				answer:
					'Produsen rumahan yang menitipkan produk, misalnya roti, keripik, atau sambal, dan toko yang menerima titipan seperti warung, kantin, kedai kopi, koperasi, atau toko oleh-oleh.'
			},
			{
				question: 'Apakah Trustip berbayar?',
				answer: 'Tidak. Daftar dan semua fitur pencatatan bisa dipakai tanpa biaya.'
			},
			{
				question: 'Apakah perlu memasang aplikasi?',
				answer:
					'Tidak wajib. Trustip berjalan di peramban HP maupun komputer. Di HP, Trustip bisa dipasang ke layar utama supaya terbuka seperti aplikasi.'
			}
		]
	},
	{
		slug: 'kerja-sama',
		title: 'Kerja sama dan pencatatan',
		items: [
			{
				question: 'Apa beda Harga Setor dan Komisi?',
				answer:
					'Dengan Harga Setor, penitip menerima jumlah tetap untuk setiap unit terjual dan toko mendapat selisihnya dari harga jual. Dengan Komisi, toko mendapat persen dari harga jual dan sisanya menjadi hak penitip.'
			},
			{
				question: 'Bagaimana Trustip menghitung barang terjual?',
				answer:
					'Terjual adalah stok tercatat dikurangi sisa fisik yang dihitung toko. Dari sisa itu, toko mengisi berapa yang dikembalikan ke penitip, dan selebihnya tetap dititip.'
			},
			{
				question: 'Bagaimana bila penitip tidak sempat memeriksa catatan?',
				answer:
					'Titipan Masuk dan Hitung Sisa yang tidak ditanggapi dalam 48 jam dianggap sesuai, sehingga Setoran tetap berjalan.'
			},
			{
				question: 'Bisakah catatan yang salah diubah?',
				answer:
					'Toko bisa membatalkan Titipan Masuk dalam 15 menit selama stoknya belum berubah. Catatan yang sudah tersimpan tidak bisa diubah diam-diam oleh siapa pun, termasuk pengelola Trustip.'
			},
			{
				question: 'Apa itu Kebutuhan Toko?',
				answer:
					'Pengumuman dari toko tentang produk yang sedang dicari. Penitip menanggapinya dengan Ajuan. Satu toko bisa memasang sampai 5 Kebutuhan terbuka sekaligus.'
			},
			{
				question: 'Bisakah kemitraan diakhiri?',
				answer:
					'Bisa, dari halaman mitra. Trustip menampilkan hal yang perlu diselesaikan dulu, seperti stok yang masih ada, dokumen yang menunggu konfirmasi, atau Setoran yang belum lunas. Catatan lama tetap bisa dibaca.'
			}
		]
	},
	{
		slug: 'uang',
		title: 'Uang dan setoran',
		items: [
			{
				question: 'Apakah uang pembayaran lewat Trustip?',
				answer:
					'Tidak. Toko membayar langsung ke penitip lewat transfer, QRIS, atau tunai. Trustip mencatat pembayaran itu dan penitip mengonfirmasi bahwa uangnya sudah diterima.'
			},
			{
				question: 'Kapan Setoran terbit?',
				answer:
					'Sesuai jadwal setor di Kesepakatan Dagang: setiap Hitung Sisa, mingguan, atau bulanan. Setoran terbit otomatis dengan rincian per produk dan tanggal jatuh tempo.'
			},
			{
				question: 'Apakah bukti bayar wajib?',
				answer:
					'Wajib untuk transfer dan QRIS, tidak untuk tunai. Bukti bayar hanya bisa dilihat kedua pihak kemitraan.'
			}
		]
	},
	{
		slug: 'keamanan',
		title: 'Keamanan dan data',
		items: [
			{
				question: 'Siapa yang bisa melihat data usaha saya?',
				answer:
					'Profil, produk, dan Kebutuhan tampil publik tanpa harga setor, nomor WhatsApp, atau lokasi persis penitip. Catatan dagang, Setoran, dan bukti bayar hanya untuk kedua pihak kemitraan. Laporan hanya untuk pemilik usaha.'
			},
			{
				question: 'Bagaimana bila tidak ada sinyal?',
				answer:
					'Titipan Masuk dan Hitung Sisa tetap bisa disimpan di perangkat dan terkirim sendiri saat sinyal kembali. Isinya bisa dilihat di menu Antrean.'
			}
		]
	},
	{
		slug: 'demo',
		title: 'Akun demo',
		items: [
			{
				question: 'Akun demo mana yang sebaiknya dipakai?',
				answer:
					'Bendi untuk sisi penitip, Zara untuk toko yang sudah bermitra dengan Bendi, dan Jessica untuk toko yang sedang menunggu jawaban Ajuan dari Bendi.'
			},
			{
				question: 'Apakah perubahan di akun demo tersimpan?',
				answer:
					'Tersimpan sampai data demo disetel ulang setiap malam pukul 03.00 WIB. Akun demo hanya bisa bermitra dengan akun demo lain.'
			}
		]
	}
];

export const faq = {
	title: 'Pertanyaan yang sering diajukan',
	eyebrow: 'FAQ',
	intro: 'Jawaban singkat tentang cara kerja Trustip, uang, keamanan, dan akun demo.',
	contentsLabel: 'Topik pertanyaan',
	groups,
	guideTitle: 'Ingin langkah lengkapnya?',
	guideBody:
		'Panduan menjelaskan alur dari mencari mitra sampai setoran diterima, tahap demi tahap.',
	guideLink: 'Baca panduan'
};
