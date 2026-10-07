export interface LegalSection {
	heading: string;
	points: string[];
}

export interface LegalDocument {
	title: string;
	effective: string;
	intro: string;
	sections: LegalSection[];
}

const privacy: LegalDocument = {
	title: 'Kebijakan Privasi',
	effective: 'Berlaku sejak 7 Oktober 2026.',
	intro:
		'Kebijakan ini menjelaskan data apa yang dikumpulkan Trustip, untuk apa data itu dipakai, siapa yang bisa melihatnya, dan pilihan yang Anda miliki.',
	sections: [
		{
			heading: 'Data yang dikumpulkan',
			points: [
				'Data akun: nama lengkap, email, nomor WhatsApp, dan kata sandi. Kata sandi tidak pernah disimpan dalam bentuk asli, hanya hasil hash yang tidak bisa dikembalikan menjadi kata sandi.',
				'Data usaha yang Anda isi sendiri: nama usaha, tipe usaha, deskripsi, alamat, lokasi, foto, produk, dan informasi pembayaran.',
				'Catatan dagang yang dibuat bersama mitra: kesepakatan, titipan masuk, hitung sisa, setoran, dan pembayaran.',
				'Data teknis: alamat IP, waktu permintaan, dan nomor permintaan, untuk keamanan dan penelusuran galat.'
			]
		},
		{
			heading: 'Untuk apa data dipakai',
			points: [
				'Menjalankan layanan: masuk, mencatat titipan dan hitungan, menerbitkan setoran, dan menampilkan laporan.',
				'Mempertemukan penitip dan toko di Cari Mitra, hanya bila Anda menyalakan sakelar tampil.',
				'Menjaga keamanan akun: membatasi percobaan masuk, mengunci akun sementara setelah gagal berulang, dan menghentikan sesi yang dipakai secara mencurigakan.',
				'Trustip tidak menjual data dan tidak memakainya untuk iklan.'
			]
		},
		{
			heading: 'Siapa yang bisa melihat data Anda',
			points: [
				'Catatan dagang dan nilai uang hanya terlihat oleh kedua pihak dalam kemitraan.',
				'Halaman publik hanya menampilkan data untuk umum, seperti nama usaha, deskripsi, dan produk. Nomor WhatsApp, alamat lengkap penitip, dan angka penjualan tidak pernah tampil di halaman publik.',
				'Lokasi penitip yang tampil untuk umum dibulatkan, sehingga letak rumah tidak bisa ditentukan.',
				'Data lokasi di dalam foto dibuang saat foto diunggah.'
			]
		},
		{
			heading: 'Cookie',
			points: [
				'Trustip memakai dua cookie sesi: token akses yang berlaku 15 menit dan token penyegar yang berlaku 30 hari sejak terakhir dipakai. Keduanya tidak bisa dibaca skrip di halaman dan hanya dikirim ke Trustip.',
				'Trustip tidak memakai cookie pelacak atau cookie iklan.'
			]
		},
		{
			heading: 'Verifikasi keamanan',
			points: [
				'Halaman daftar dan masuk memakai Cloudflare Turnstile untuk membedakan manusia dari program otomatis. Widget ini dimuat dari server Cloudflare hanya di halaman yang memakainya dan dapat memproses data teknis peramban, seperti alamat IP, menurut kebijakan privasi Cloudflare.'
			]
		},
		{
			heading: 'Pihak yang membantu menjalankan layanan',
			points: [
				'Aplikasi disajikan lewat Vercel, basis data disimpan di Neon, dan API berjalan di server sewaan. Pihak tersebut memproses data hanya untuk menjalankan Trustip.'
			]
		},
		{
			heading: 'Segel dokumen',
			points: [
				'Setiap catatan dagang disegel dengan tanda tangan digital supaya tidak bisa diubah diam-diam.',
				'Halaman verifikasi segel bisa dibuka siapa saja, tetapi hanya menampilkan jenis dan nomor dokumen, nama kedua pihak, jenis peristiwa, waktu, dan status segel. Nilai uang tidak pernah tampil.'
			]
		},
		{
			heading: 'Penyimpanan dan hak Anda',
			points: [
				'Catatan dagang yang sudah disegel tidak bisa diubah atau dihapus, karena juga menjadi catatan milik mitra Anda.',
				'Log teknis diputar otomatis dan tidak memuat kata sandi, token, nomor telepon, atau alamat.',
				'Anda bisa meminta salinan atau penghapusan data pribadi kepada tim pengelola Trustip. Catatan dagang yang sudah disegel tetap disimpan karena menyangkut hak mitra Anda.'
			]
		}
	]
};

const terms: LegalDocument = {
	title: 'Ketentuan Layanan',
	effective: 'Berlaku sejak 7 Oktober 2026.',
	intro:
		'Dengan mendaftar dan memakai Trustip, Anda menyetujui ketentuan berikut. Bacalah bersama Kebijakan Privasi.',
	sections: [
		{
			heading: 'Tentang Trustip',
			points: [
				'Trustip membantu penitip dan toko mencatat titipan, hitung sisa, dan setoran secara bersama.',
				'Trustip bukan pihak dalam kemitraan dan tidak memegang uang pengguna. Pembayaran setoran dilakukan langsung antara toko dan penitip, sedangkan Trustip mencatat pembayaran dan konfirmasinya.'
			]
		},
		{
			heading: 'Akun',
			points: [
				'Satu akun untuk satu orang. Gunakan data yang benar saat mendaftar.',
				'Jaga kerahasiaan kata sandi. Anda bertanggung jawab atas aktivitas yang terjadi lewat akun Anda. Bila curiga akun dipakai orang lain, pilih keluar dari semua perangkat.'
			]
		},
		{
			heading: 'Kesepakatan dan catatan dagang',
			points: [
				'Kesepakatan dibuat dan disetujui oleh penitip dan toko sendiri. Trustip mencatat dan menghitung berdasarkan isian kedua pihak dan syarat kesepakatan yang berlaku saat catatan dibuat.',
				'Catatan yang sudah disegel tidak bisa diubah. Koreksi dilakukan lewat fitur yang tersedia, seperti pembatalan dalam batas waktunya atau catatan baru.'
			]
		},
		{
			heading: 'Batas segel',
			points: [
				'Segel membuktikan bahwa catatan tidak berubah sejak dicatat Trustip dan akun mana yang menyetujuinya.',
				'Segel bukan tanda tangan elektronik tersertifikasi dari penyelenggara sertifikasi elektronik.'
			]
		},
		{
			heading: 'Larangan',
			points: [
				'Mencatat transaksi palsu atau menyesatkan.',
				'Mengunggah konten yang melanggar hukum atau hak orang lain.',
				'Mencoba masuk ke akun atau data organisasi lain.',
				'Mengganggu layanan, termasuk dengan permintaan otomatis dalam jumlah besar.'
			]
		},
		{
			heading: 'Laporan dan penangguhan',
			points: [
				'Profil, produk, dan Kebutuhan Toko yang dilaporkan dapat ditinjau tim pengelola.',
				'Organisasi yang melanggar ketentuan dapat ditangguhkan: tidak bisa mengirim Ajuan baru dan tidak tampil di Cari Mitra, tetapi tetap bisa menyelesaikan kewajiban yang sedang berjalan.'
			]
		},
		{
			heading: 'Ketersediaan layanan',
			points: [
				'Titipan Masuk dan Hitung Sisa bisa dicatat tanpa koneksi dan dikirim saat sinyal kembali. Catatan yang belum terkirim tersimpan di perangkat Anda.',
				'Tim pengelola berusaha menjaga layanan tetap berjalan, tetapi tidak menjamin layanan selalu tersedia tanpa gangguan.'
			]
		},
		{
			heading: 'Perubahan ketentuan',
			points: [
				'Ketentuan ini dapat diperbarui. Tanggal berlaku di bagian atas halaman menunjukkan versi terbaru.'
			]
		}
	]
};

export const legal = {
	privacy,
	terms,
	seeAlso: 'Lihat juga'
};
