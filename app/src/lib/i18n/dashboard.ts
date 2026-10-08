const units = (count: number, unit: string) => `${count.toLocaleString('id-ID')} ${unit}`;

export const dashboard = {
	dashboard: {
		consignorTitle: 'Dasbor Penitip',
		storeTitle: 'Dasbor Toko',
		loadFailed: 'Dasbor tidak bisa dimuat.',
		figures: 'Angka kunci',
		unitsAtStores: 'Unit di semua toko',
		unitsConsigned: 'Unit titipan',
		activePartners: 'Mitra aktif',
		soldLast30: 'Terjual 30 hari',
		consignorShareLast30: 'Hak penitip 30 hari',
		grossLast30: 'Omzet titipan 30 hari',
		storeShareLast30: 'Untung toko 30 hari',
		unpaidSettlements: 'Setoran belum lunas',
		mustPayTotal: 'Total harus setor',
		overdueCaption: (count: number, amount: string) => `${String(count)} lewat tempo, ${amount}`,
		dueSoonCaption: (count: number, amount: string) =>
			`${String(count)} jatuh tempo 3 hari lagi, ${amount}`,
		units: (count: number) => units(count, 'unit'),
		tasksTitle: 'Perlu tindakan',
		noTasks: 'Tidak ada yang menunggu Anda saat ini.',
		tasks: {
			dropsToConfirm: (count: number) => `${String(count)} Titipan Masuk menunggu konfirmasi`,
			countsToConfirm: (count: number) => `${String(count)} Hitung Sisa menunggu konfirmasi`,
			paymentsToReview: (count: number) => `${String(count)} pembayaran menunggu konfirmasi`,
			unpaidSettlements: (count: number) => `${String(count)} Setoran belum dibayar`,
			paymentsRejected: (count: number) => `${String(count)} pembayaran ditolak`,
			proposalsAwaiting: (count: number) => `${String(count)} Ajuan menunggu giliran Anda`
		},
		chartFilter: 'Saring grafik',
		allSales: 'Semua penjualan',
		rankingStores: 'Toko terlaris',
		rankingProducts: 'Produk terlaris',
		rankingEmpty: 'Belum ada penjualan 30 hari terakhir.',
		insightsTitle: 'Saran',
		noInsights: 'Belum ada saran. Saran muncul dari stok, hitungan, dan Setoran Anda.',
		levels: { HIGH: 'Penting', MEDIUM: 'Perlu dilihat', LOW: 'Info' },
		quickCountTitle: 'Hitung cepat',
		quickCountEmpty: 'Semua mitra berstok sudah dihitung hari ini.',
		quickCountRow: (count: number) => `${units(count, 'unit')} tercatat`,
		count: 'Hitung',
		ownerOnly: 'Angka penjualan hanya tampil untuk pemilik usaha.'
	},
	insights: {
		titles: {
			EXPIRED: (product: string) => `${product} sudah kedaluwarsa`,
			NEAR_EXPIRY: (product: string) => `${product} mendekati kedaluwarsa`,
			LOW_STOCK: (product: string) => `Stok ${product} segera habis`,
			SETTLEMENT_OVERDUE: (number: string) => `Setoran ${number} lewat tempo`,
			NOT_COUNTED: (partner: string) => `${partner} belum dihitung`,
			MATCHING_PARTNER: (organization: string) => `${organization} cocok untuk Anda`
		},
		bodies: {
			EXPIRED: (partner: string, batch: string, count: number) =>
				`${units(count, 'unit')} dari batch ${batch} di ${partner} perlu ditarik.`,
			NEAR_EXPIRY: (partner: string, days: number, unsold: number) =>
				unsold > 0
					? `Di ${partner}, sekitar ${units(unsold, 'unit')} diperkirakan tidak laku dalam ${String(days)} hari.`
					: `Di ${partner}, kedaluwarsa dalam ${String(days)} hari. Penjualan saat ini cukup untuk menghabiskannya.`,
			LOW_STOCK: (partner: string, days: number) =>
				`Di ${partner}, stok cukup sekitar ${String(days)} hari lagi.`,
			SETTLEMENT_OVERDUE: (partner: string, days: number, amount: string) =>
				`${amount} dari ${partner} lewat ${String(days)} hari dari jatuh tempo.`,
			NOT_COUNTED: (days: number, count: number) =>
				`${units(count, 'unit')} belum dihitung selama ${String(days)} hari.`,
			MATCHING_PARTNER: (score: number) => `Skor kecocokan ${String(score)} dari 100.`
		},
		actions: {
			COUNT_WITH_RETURNS: 'Hitung dan kembalikan',
			COUNT: 'Hitung Sisa',
			REMIND_STORE: 'Ingatkan toko',
			VIEW_DROP_SUGGESTION: 'Lihat stok mitra',
			ASK_RESTOCK: 'Minta kirim',
			REMIND_PAYMENT: 'Ingatkan bayar',
			RECORD_PAYMENT: 'Catat bayar',
			REMIND_COUNT: 'Ingatkan hitung',
			OPEN_PROFILE: 'Buka profil'
		}
	}
};
