export const organization = {
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
		mapUnavailable: 'Peta tidak bisa dimuat. Gunakan lokasi perangkat atau coba lagi nanti.',
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
	settings: {
		title: 'Pengaturan',
		profileTitle: 'Profil',
		general: 'Informasi profil usaha',
		edit: 'Edit profil',
		cancel: 'Batal',
		typeLabel: 'Tipe usaha',
		logo: 'Logo',
		chooseLogo: 'Pilih gambar logo',
		logoHint: 'JPG, PNG, atau WebP. Dikecilkan otomatis sebelum diunggah.',
		logoAlt: (name: string) => `Logo ${name}`,
		imageInvalid: 'Berkas harus gambar JPG, PNG, atau WebP.',
		save: 'Simpan profil',
		saved: 'Profil usaha tersimpan.',
		loadFailed: 'Profil usaha tidak bisa dimuat.',
		ownersOnly: 'Hanya pemilik usaha yang bisa mengubah profil.'
	},
	storeProfile: {
		title: 'Profil toko',
		storeType: 'Jenis toko',
		storeTypes: {
			GROCERY_STALL: 'Warung atau toko kelontong',
			LOCAL_MINIMARKET: 'Minimarket lokal',
			SOUVENIR_SHOP: 'Toko oleh-oleh',
			CANTEEN: 'Kantin (sekolah, kampus, kantor)',
			CAFE: 'Kafe atau kedai kopi',
			BAKERY: 'Toko kue dan roti',
			COOPERATIVE: 'Koperasi',
			MARKET_KIOSK: 'Kios pasar',
			OTHER: 'Lainnya'
		},
		categories: 'Kategori produk yang diterima',
		acceptsFixedDeposit: 'Menerima model Harga Setor tetap',
		acceptsCommission: 'Menerima model Komisi',
		commissionMin: 'Komisi terendah (%)',
		commissionMax: 'Komisi tertinggi (%)',
		commissionOrder: 'Komisi terendah tidak boleh melebihi yang tertinggi.',
		requirements: 'Syarat penitip',
		requiresDistributionPermit: 'Wajib izin edar (PIRT atau MD)',
		requiresHalal: 'Wajib sertifikat halal',
		requiresSample: 'Wajib kirim sampel',
		otherRequirements: 'Syarat lain',
		openingHours: 'Jam buka',
		openingHoursHint: 'Contoh: Setiap hari 07.00 sampai 21.00',
		photos: 'Foto toko',
		photosHint: 'Sampai 5 foto. JPG, PNG, atau WebP.',
		addPhoto: 'Tambah foto toko',
		photoAlt: (number: number) => `Foto toko ${String(number)}`,
		removePhoto: (number: number) => `Hapus foto ${String(number)}`
	},
	consignorProfile: {
		title: 'Profil penitip',
		deliveryRadius: 'Jangkauan antar (km)',
		deliveryRadiusHint: 'Isi 0 bila belum ditentukan.',
		deliveryDays: 'Hari antar',
		days: {
			MON: 'Senin',
			TUE: 'Selasa',
			WED: 'Rabu',
			THU: 'Kamis',
			FRI: 'Jumat',
			SAT: 'Sabtu',
			SUN: 'Minggu'
		},
		preferredStoreTypes: 'Jenis toko yang diincar',
		legalTitle: 'Legalitas',
		declaredHint: 'Dinyatakan oleh pemilik usaha. Trustip tidak memverifikasinya.',
		businessIdNo: 'NIB',
		distributionPermitNo: 'Nomor izin edar (PIRT atau MD)',
		halalCertNo: 'Nomor sertifikat halal',
		canProvideSample: 'Bisa mengirim sampel'
	},
	payment: {
		title: 'Cara pembayaran',
		intro:
			'Pilih cara toko membayar setoran. Hanya terlihat oleh toko yang bermitra aktif dengan Anda.',
		methods: 'Cara bayar yang diterima',
		methodsHint:
			'Toko hanya bisa memilih cara yang Anda centang, dan wajib mengunggah bukti bayar.',
		methodLabels: { CASH: 'Tunai', TRANSFER: 'Transfer bank', QRIS: 'QRIS' },
		bankName: 'Nama bank',
		bankNameHint: 'Contoh: BRI, BCA, Mandiri, atau BSI',
		bankAccountNumber: 'Nomor rekening',
		bankAccountHolder: 'Atas nama',
		info: 'Catatan untuk toko',
		infoHint:
			'Opsional. Contoh: transfer sebelum pukul 17.00, atau bayar tunai saat barang diantar.',
		qr: 'Gambar QRIS',
		chooseQr: 'Pilih gambar QRIS',
		qrAlt: 'QRIS pembayaran',
		edit: 'Edit',
		save: 'Simpan info pembayaran',
		saved: 'Info pembayaran tersimpan.'
	},
	listing: {
		title: 'Tampil di Cari Mitra',
		switchLabel: 'Tampilkan usaha saya di Cari Mitra',
		description: 'Usaha baru bisa tampil setelah semua butir wajib terisi.',
		meter: (percent: number) => `Kelengkapan profil ${String(percent)}%`,
		required: 'Wajib untuk tampil',
		listed: 'Usaha Anda tampil di Cari Mitra.',
		hidden: 'Usaha Anda tidak tampil di Cari Mitra.',
		requirements: {
			NAME: 'Nama usaha',
			CITY: 'Kota',
			LOCATION: 'Lokasi di peta',
			STORE_TYPE: 'Jenis toko',
			ACCEPTED_CATEGORY: 'Minimal satu kategori yang diterima',
			ACTIVE_PRODUCT: 'Minimal satu produk aktif',
			DESCRIPTION: 'Deskripsi singkat',
			LOGO: 'Logo',
			STORE_PHOTO: 'Foto toko',
			PAYMENT_INFO: 'Info pembayaran'
		}
	}
};
