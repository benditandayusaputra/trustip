import { describe, expect, it } from 'vitest';
import { bottomNavigationOf, isCurrent, navigationFor } from './navigation';

const labels = (items: { label: string }[]) => items.map((item) => item.label);

describe('navigationFor', () => {
	it('gives a consignor a workspace group and a product-focused management group', () => {
		const navigation = navigationFor('CONSIGNOR');
		expect(labels(navigation.workspace)).toEqual(['Beranda', 'Cari', 'Mitra', 'Laporan']);
		expect(labels(navigation.manage)).toEqual(['Produk', 'Ajuan', 'Stok', 'Pengaturan']);
	});

	it('gives a store Laporan in its workspace group', () => {
		expect(labels(navigationFor('STORE').workspace)).toEqual([
			'Beranda',
			'Hitung',
			'Titipan',
			'Cari',
			'Laporan'
		]);
	});

	it('leaves notifications to the bell and settlements to Laporan', () => {
		for (const type of ['CONSIGNOR', 'STORE'] as const) {
			const navigation = navigationFor(type);
			const hrefs = [...navigation.workspace, ...navigation.manage].map((item) => item.href);
			expect(hrefs).not.toContain('/app/pemberitahuan');
			expect(hrefs).not.toContain('/app/setoran');
		}
	});

	it('keeps four bottom tabs before Lainnya', () => {
		expect(labels(bottomNavigationOf(navigationFor('STORE')).primary)).toEqual([
			'Beranda',
			'Hitung',
			'Titipan',
			'Cari'
		]);
		expect(labels(bottomNavigationOf(navigationFor('CONSIGNOR')).more)).toContain('Produk');
	});
});

describe('isCurrent', () => {
	it('marks the section and its sub pages', () => {
		expect(isCurrent('/app/titipan', '/app/titipan')).toBe(true);
		expect(isCurrent('/app/titipan', '/app/titipan/0192f0c4')).toBe(true);
	});

	it('does not mark a different section sharing a prefix', () => {
		expect(isCurrent('/app/titipan', '/app/titipan-lama')).toBe(false);
	});
});
