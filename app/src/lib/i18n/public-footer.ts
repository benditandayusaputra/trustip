import { common } from './common.js';

export const publicFooter = {
	description:
		'Platform titip jual untuk pemilik produk dan toko lokal. Temukan mitra, sepakati harga, dan kelola catatan bersama.',
	note: 'Dari barang dititipkan sampai uang diterima.',
	copyright: (year: number) => `© ${String(year)} ${common.brand}`,
	groups: [
		{
			title: 'Jelajahi',
			label: 'Jelajahi Trustip',
			links: [
				{ label: common.nav.findStores, href: '/cari/toko' },
				{ label: 'Cari produk dan penitip', href: '/cari/penitip' },
				{ label: common.nav.needsBoard, href: '/kebutuhan' }
			]
		},
		{
			title: 'Mulai',
			label: 'Mulai di Trustip',
			links: [
				{ label: 'Cara kerja', href: '/#beranda-cara' },
				{ label: 'Buat akun usaha', href: '/daftar' },
				{ label: 'Jelajahi demo', href: '/demo' }
			]
		},
		{
			title: 'Informasi',
			label: common.layout.footerNavigation,
			links: [
				{ label: common.nav.privacy, href: '/privasi' },
				{ label: common.nav.terms, href: '/ketentuan' }
			]
		}
	]
};
