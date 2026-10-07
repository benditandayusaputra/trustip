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
		profileTitle: 'Profil usaha',
		typeLabel: 'Tipe usaha',
		logo: 'Logo',
		chooseLogo: 'Pilih gambar logo',
		logoHint: 'JPG, PNG, atau WebP. Dikecilkan otomatis sebelum diunggah.',
		logoAlt: (name: string) => `Logo ${name}`,
		imageInvalid: 'Berkas harus gambar JPG, PNG, atau WebP.',
		save: 'Simpan profil',
		saved: 'Profil usaha tersimpan.',
		loadFailed: 'Profil usaha tidak bisa dimuat.',
		ownersOnly: 'Hanya pemilik usaha yang bisa mengubah profil.',
		completeProfile: 'Lengkapi profil usaha'
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
		removePhoto: (number: number) => `Hapus foto ${String(number)}`,
		save: 'Simpan profil toko',
		saved: 'Profil toko tersimpan.'
	}
};
