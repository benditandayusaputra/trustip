export interface GuidePhase {
	slug: string;
	title: string;
	actor: string;
	summary: string;
	steps: string[];
}

const phases: GuidePhase[] = [
	{
		slug: 'akun',
		title: 'Siapkan akun dan profil usaha',
		actor: 'Penitip dan toko',
		summary: 'Satu akun mewakili satu usaha. Profil yang lengkap memudahkan calon mitra percaya.',
		steps: [
			'Tekan Daftar, isi nama, email, nomor WhatsApp, dan kata sandi, lalu selesaikan verifikasi keamanan.',
			'Pilih tipe usaha: Penitip bila Anda membuat produk, atau Toko bila Anda menjual titipan. Tipe usaha tidak bisa diubah setelah disimpan.',
			'Lengkapi Pengaturan: nama usaha, lokasi di peta, foto, serta jam buka atau hari antar.',
			'Penitip menambah produk di menu Produk, lengkap dengan harga jual saran, harga setor, dan masa simpan.',
			'Nyalakan Tampilkan usaha saya di Cari Mitra di tab Publikasi.'
		]
	},
	{
		slug: 'cari-mitra',
		title: 'Temukan mitra',
		actor: 'Penitip dan toko',
		summary:
			'Penitip mencari toko yang mau menerima titipan. Toko mencari produk untuk mengisi rak.',
		steps: [
			'Penitip membuka Cari untuk melihat toko terdekat yang menerima kategorinya, lengkap dengan jarak dan syarat titip.',
			'Toko membuka Cari untuk menemukan produk dan penitip, atau memasang Kebutuhan bila belum menemukan yang cocok.',
			'Pakai Filter dan urutan untuk menyaring kategori, model harga, dan jarak. Pilih tampilan Peta untuk melihat lokasi.',
			'Buka profil mitra untuk membaca syarat, produk, dan kebutuhannya sebelum mengajak kerja sama.'
		]
	},
	{
		slug: 'kesepakatan',
		title: 'Sepakati aturan dagang lewat Ajuan',
		actor: 'Penitip dan toko',
		summary:
			'Ajuan berisi produk, harga, dan jadwal. Setelah diterima, aturannya berlaku sama untuk kedua pihak.',
		steps: [
			'Dari profil mitra, tekan Ajukan kerja sama, atau buka Ajuan lalu Undang mitra.',
			'Pilih produk dan model harga: Harga Setor tetap per unit, atau Komisi dalam persen dari harga jual.',
			'Atur jadwal setor (setiap Hitung Sisa, mingguan, atau bulanan), tempo bayar, dan irama hitung.',
			'Pihak lawan menekan Terima, Tawar dengan angka baru, atau Tolak.',
			'Begitu diterima, Kesepakatan Dagang terbentuk dan tampil sama di dua layar. Perubahan berikutnya lewat Ajukan perubahan.'
		]
	},
	{
		slug: 'titipan-masuk',
		title: 'Catat Titipan Masuk',
		actor: 'Toko mencatat, penitip memeriksa',
		summary: 'Setiap barang yang diantar dicatat saat itu juga, sehingga stok kedua pihak sama.',
		steps: [
			'Saat barang datang, toko membuka Titipan, menekan Catat titipan, lalu memilih penitip.',
			'Isi jumlah per produk dengan tombol tambah dan kurang. Tanggal kedaluwarsa dihitung dari masa simpan produk.',
			'Simpan. Stok bertambah dan penitip mendapat pemberitahuan untuk menekan Sesuai.',
			'Salah catat? Toko bisa membatalkan dalam 15 menit selama stoknya belum berubah.'
		]
	},
	{
		slug: 'hitung-sisa',
		title: 'Hitung Sisa di rak',
		actor: 'Toko mencatat, penitip memeriksa',
		summary: 'Toko menghitung barang yang tersisa. Selisih dengan stok tercatat dianggap terjual.',
		steps: [
			'Buka Hitung lalu Hitung sisa, atau tekan Hitung pada daftar Hitung cepat di Beranda.',
			'Isi sisa fisik per produk. Pilih Isi sisa atau Isi terjual, Trustip menghitung yang satunya.',
			'Isi berapa yang dikembalikan ke penitip. Selebihnya tetap dititip di rak.',
			'Periksa Pratinjau: terjual, omzet, untung toko, dan harus setor. Lalu tekan Simpan Hitung Sisa.',
			'Penitip menerima pemberitahuan dan menekan Sesuai.'
		]
	},
	{
		slug: 'setoran',
		title: 'Setoran dan pembayaran',
		actor: 'Toko membayar, penitip mengonfirmasi',
		summary: 'Setoran merangkum uang yang harus dibayar toko kepada penitip.',
		steps: [
			'Setoran terbit otomatis sesuai jadwal setor: setiap Hitung Sisa, mingguan, atau bulanan.',
			'Toko membuka Laporan, tab Setoran, memilih Setoran yang terbit, lalu mengisi Catat bayar: metode, tanggal, dan foto bukti untuk transfer atau QRIS.',
			'Uang dibayar langsung ke penitip memakai info rekening atau QRIS yang tampil di Setoran.',
			'Penitip memeriksa bukti lalu menekan Uang diterima, atau Tolak dengan alasan.'
		]
	},
	{
		slug: 'laporan',
		title: 'Pantau laporan dan pengingat',
		actor: 'Pemilik usaha',
		summary: 'Angka dan grafik membantu memutuskan produk mana yang ditambah atau dikurangi.',
		steps: [
			'Beranda menampilkan angka 30 hari, grafik penjualan harian, pembagian omzet, dan hal yang perlu ditindaklanjuti.',
			'Laporan menyajikan penjualan per mitra, per produk, atau per hari, lengkap dengan grafik dan unduhan CSV. Tab Setoran memuat ringkasan dan daftar Setoran.',
			'Pengingat di Beranda memberi tahu stok yang segera habis, barang mendekati kedaluwarsa, dan Setoran lewat tempo.',
			'Stok menunjukkan jumlah tercatat di setiap toko beserta kedaluwarsa terdekat.'
		]
	},
	{
		slug: 'luring',
		title: 'Tetap mencatat tanpa sinyal',
		actor: 'Toko',
		summary: 'Titipan Masuk dan Hitung Sisa tetap bisa dicatat saat sinyal hilang.',
		steps: [
			'Pasang Trustip di HP lewat menu peramban Tambahkan ke layar utama.',
			'Saat luring, catatan disimpan di perangkat dan masuk Antrean.',
			'Saat sinyal kembali, antrean terkirim sendiri. Catatan yang bentrok ditandai untuk ditinjau.'
		]
	}
];

export const guide = {
	title: 'Panduan pemakaian Trustip',
	eyebrow: 'Panduan',
	headline: 'Dari mencari mitra sampai setoran diterima',
	intro: 'Delapan tahap singkat untuk memakai Trustip. Buka tiap tahap untuk melihat langkahnya.',
	register: 'Buat akun usaha',
	stepsLabel: 'Langkah',
	stepsToggle: 'Lihat langkah',
	phases,
	faqTitle: 'Masih ada pertanyaan?',
	faqBody:
		'Jawaban singkat tentang uang, keamanan, kerja luring, dan akun demo ada di halaman FAQ.',
	faqLink: 'Buka FAQ'
};
