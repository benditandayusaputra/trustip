export const proposals = {
	proposals: {
		newTitle: 'Undang mitra',
		newIntro:
			'Buat Ajuan Kerja Sama lalu bagikan tautannya ke mitra lama lewat WhatsApp. Mitra yang membuka tautan bisa menerima atau menolak.',
		detailTitle: 'Ajuan Kerja Sama',
		products: 'Produk yang ditawarkan',
		productsHint: 'Centang produk yang ingin dititipkan, lalu atur syaratnya.',
		noProducts: 'Belum ada produk aktif. Tambahkan produk di katalog dulu.',
		storeIntro:
			'Undangan dari toko memuat syarat umum saja. Penitip menambahkan produknya setelah bermitra.',
		model: 'Model harga',
		models: { FIXED_DEPOSIT: 'Harga Setor', COMMISSION: 'Komisi' },
		retailPrice: 'Harga jual',
		depositPrice: 'Harga setor',
		commission: 'Komisi (%)',
		commissionHint: 'Contoh: 20 atau 17,5. Maksimal 50.',
		example: (quantity: number) => `Bila terjual ${String(quantity)}`,
		gross: 'Omzet',
		storeShare: 'Untung toko',
		consignorShare: 'Hak penitip',
		perUnitConsignor: 'Hak penitip per unit',
		perUnitStore: 'Untung toko per unit',
		generalTerms: 'Syarat umum',
		schedule: 'Jadwal setor',
		schedules: {
			ON_COUNT: 'Setiap Hitung Sisa',
			WEEKLY: 'Mingguan, Senin sampai Minggu',
			MONTHLY: 'Bulanan, per bulan kalender'
		},
		paymentTerms: 'Tempo bayar (hari)',
		paymentTermsHint: 'Hari setelah Setoran terbit. 0 sampai 30.',
		rhythm: 'Irama hitung',
		rhythms: {
			DAILY: 'Setiap tutup toko',
			EACH_VISIT: 'Setiap penitip datang',
			WEEKLY: 'Mingguan'
		},
		notes: 'Catatan',
		notesHint: 'Opsional, misalnya letak pajang. Maksimal 500 karakter.',
		fixedTerms: 'Ketentuan tetap',
		fixedTermsItems: [
			'Barang tak laku dan kedaluwarsa dikembalikan ke penitip dan tidak dibayar.',
			'Selisih antara stok tercatat dan sisa fisik dianggap terjual.'
		],
		message: 'Pesan pengantar',
		messageHint: 'Opsional. Maksimal 1.000 karakter.',
		send: 'Buat tautan undangan',
		sent: 'Tautan undangan siap dibagikan.',
		shareTitle: 'Bagikan undangan',
		shareBody:
			'Kirim tautan ini ke mitra. Tautan berlaku 14 hari dan hanya bisa dipakai satu usaha.',
		copyLink: 'Salin tautan',
		copied: 'Tautan disalin.',
		shareWhatsApp: 'Bagikan lewat WhatsApp',
		whatsAppText: (sender: string, link: string) =>
			`Halo, ${sender} mengajak bermitra lewat Trustip. Buka tautan ini untuk melihat syaratnya: ${link}`,
		openProposal: 'Lihat Ajuan',
		number: 'Nomor',
		status: 'Status',
		statuses: {
			OPEN: 'Terbuka',
			ACCEPTED: 'Diterima',
			DECLINED: 'Ditolak',
			WITHDRAWN: 'Ditarik',
			EXPIRED: 'Kedaluwarsa'
		},
		consignor: 'Penitip',
		store: 'Toko',
		notJoined: 'Belum bergabung',
		awaiting: (name: string) => `Menunggu tanggapan dari ${name}.`,
		awaitingInvitee: 'Menunggu mitra membuka tautan undangan.',
		expiresAt: (date: string) => `Berlaku sampai ${date}`,
		loadFailed: 'Ajuan tidak bisa dimuat.',
		invalidPercent: 'Isi angka 0 sampai 50.'
	},
	invitation: {
		title: 'Undangan kerja sama',
		signedOutBody:
			'Anda diundang bermitra di Trustip. Masuk atau daftar untuk melihat undangannya.',
		signIn: 'Masuk',
		register: 'Daftar',
		newcomerBody: 'Daftarkan usaha Anda dulu untuk membuka undangan ini.',
		onboard: 'Daftarkan usaha',
		summary: (sender: string) => `${sender} mengajak Anda bermitra.`,
		productCount: (count: number) => `${String(count)} produk ditawarkan.`,
		open: 'Buka Ajuan',
		loadFailed: 'Undangan tidak bisa dibuka.'
	}
};
