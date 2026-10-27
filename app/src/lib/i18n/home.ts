export const home = {
	home: {
		title: 'Trustip: titip jual untuk UMKM',
		headline: 'Titip jual lebih mudah.',
		headlineAccent: 'Dari mitra sampai setoran.',
		audience: 'Platform titip jual untuk UMKM',
		audienceDetail: 'Produsen rumahan, warung, kantin, dan toko oleh-oleh.',
		intro:
			'Trustip membantu pemilik produk menitipkan barang di toko, menyepakati harga, dan mencatat penjualan hingga pembayaran. Penitip dan toko melihat catatan yang sama.',
		seeWorkflow: 'Lihat cara kerja',
		highlights: ['Temukan mitra', 'Hitungan otomatis', 'Catatan bersama'],
		illustration: {
			eyebrow: 'Satu alur, dua mitra',
			caption: 'Produk dititipkan. Catatan dibagikan.',
			consignor: 'Penitip',
			consignorBody: 'Menyiapkan dan menitipkan produk',
			store: 'Toko',
			storeBody: 'Menjual dan mencatat sisanya'
		},
		rolesEyebrow: 'Untuk usaha Anda',
		rolesTitle: 'Punya produk atau punya toko?',
		rolesIntro: 'Dua peran usaha, satu tempat untuk bekerja sama.',
		roles: [
			{
				title: 'Saya penitip produk',
				body: 'Titipkan produk ke mitra dan pantau perkembangannya tanpa menanyakan catatan satu per satu.',
				benefits: [
					'Cari toko yang menerima titipan',
					'Pantau stok, penjualan, dan setoran',
					'Sepakati harga dan retur dengan jelas'
				],
				action: 'Temukan toko',
				href: '/cari/toko'
			},
			{
				title: 'Saya pemilik toko',
				body: 'Temukan produk untuk mengisi rak dan kelola titipan dari beberapa penitip dengan lebih rapi.',
				benefits: [
					'Cari produk atau pasang Kebutuhan',
					'Catat barang masuk dan hitung sisanya',
					'Lihat untung dan kewajiban setoran'
				],
				action: 'Temukan produk dan penitip',
				href: '/cari/penitip'
			}
		],
		register: 'Daftar gratis',
		signIn: 'Masuk',
		problemsTitle: 'Masalah titip jual hari ini',
		problemsEyebrow: 'Lebih mudah dikelola',
		problemsIntro:
			'Dari mencari tempat titip sampai memeriksa catatan, Trustip membantu di setiap tahap.',
		problems: [
			{
				title: 'Sulit menemukan toko',
				body: 'Produsen berkeliling mencari toko yang mau dititipi, dan syarat tiap toko berbeda.',
				solution: 'Cari toko dan baca syaratnya sebelum mengajukan kerja sama.'
			},
			{
				title: 'Harus datang untuk tahu laku berapa',
				body: 'Jumlah terjual baru diketahui setelah penitip datang atau menelepon toko.',
				solution: 'Pantau penjualan setelah toko mencatat Hitung Sisa.'
			},
			{
				title: 'Catatan dua pihak tidak sama',
				body: 'Nota kertas di dua tempat memicu selisih hitungan, telat bayar, sampai konflik.',
				solution: 'Periksa catatan bersama lalu konfirmasi angkanya.'
			},
			{
				title: 'Retur tidak tercatat',
				body: 'Barang tak laku dan kedaluwarsa menumpuk tanpa catatan, dan jumlah titip hanya kira-kira.',
				solution: 'Catat retur dan lihat peringatan kedaluwarsa pada dasbor.'
			}
		],
		stepsTitle: 'Cara kerja',
		stepsEyebrow: 'Dari mitra sampai uang diterima',
		stepsIntro:
			'Lima langkah yang menghubungkan penitip dan toko. Setiap pihak tahu apa yang perlu dilakukan.',
		paymentNote:
			'Pembayaran dilakukan langsung antar mitra. Trustip mencatat pembayaran dan konfirmasi penerimaannya.',
		steps: [
			{
				title: 'Cari mitra',
				actor: 'Penitip & toko',
				body: 'Temukan toko yang menerima titipan atau produk untuk mengisi rak.'
			},
			{
				title: 'Sepakati harga',
				actor: 'Penitip & toko',
				body: 'Setujui harga, bagian toko, jadwal setor, dan aturan barang yang kembali.'
			},
			{
				title: 'Catat titipan',
				actor: 'Toko & penitip',
				body: 'Toko mencatat produk dan jumlah yang masuk. Penitip memeriksa catatannya.'
			},
			{
				title: 'Hitung sisa',
				actor: 'Toko',
				body: 'Isi sisa dan retur. Jumlah terjual serta uang yang harus disetor dihitung otomatis.'
			},
			{
				title: 'Konfirmasi setoran',
				actor: 'Toko & penitip',
				body: 'Toko mencatat pembayaran. Penitip mengonfirmasi uang yang diterima.'
			}
		],
		closingTitle: 'Mulai bersama mitra Anda',
		closingBody: 'Satu akun untuk mencatat titipan, hitungan, dan setoran bersama mitra.',
		tryDemo: 'Coba akun demo',
		browseStores: 'Lihat toko yang menerima titipan',
		browseNeeds: 'Lihat Kebutuhan toko',
		exampleEyebrow: 'Begini hasilnya',
		exampleTitle: 'Cukup catat barang masuk dan sisanya.',
		exampleIntro:
			'Dari 20 roti yang dititipkan dan 2 yang tersisa, Trustip menghitung 18 terjual. Omzet, untung toko, dan setoran mengikuti harga yang sudah disepakati.',
		exampleHighlights: [
			'Penjualan dihitung dari sisa',
			'Pembagian sesuai kesepakatan',
			'Catatan bisa diperiksa kedua pihak'
		],
		example: {
			title: 'Satu catatan, dua pihak',
			partners: 'Roti Mas Bendi → Warung Mbak Zara',
			pricing: 'Contoh harga: jual Rp2.000 · setor Rp1.750 per roti.',
			product: 'Roti manis',
			units: '20 dititip · 2 tersisa · 18 terjual',
			gross: 'Rp36.000',
			storeShareLabel: 'Untung toko',
			storeShare: 'Rp4.500',
			mustPayLabel: 'Harus setor',
			mustPay: 'Rp31.500',
			sameRecord: 'Toko dan penitip melihat hitungan yang sama.',
			action: 'Lihat alurnya di akun demo',
			disclaimer: 'Contoh hitungan untuk menjelaskan cara kerja.'
		}
	}
};
