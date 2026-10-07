import { describe, expect, it } from 'vitest';
import { isCurrent, navigationFor } from './navigation';

describe('navigationFor', () => {
	it('gives a consignor Beranda, Cari, Mitra, Setoran before Lainnya', () => {
		expect(navigationFor('CONSIGNOR').primary.map((item) => item.label)).toEqual([
			'Beranda',
			'Cari',
			'Mitra',
			'Setoran'
		]);
	});

	it('gives a store Beranda, Hitung, Titipan, Cari before Lainnya', () => {
		expect(navigationFor('STORE').primary.map((item) => item.label)).toEqual([
			'Beranda',
			'Hitung',
			'Titipan',
			'Cari'
		]);
	});

	it('keeps products only for consignors and needs only for stores', () => {
		const consignorMore = navigationFor('CONSIGNOR').more.map((item) => item.href);
		const storeMore = navigationFor('STORE').more.map((item) => item.href);
		expect(consignorMore).toContain('/app/produk');
		expect(storeMore).not.toContain('/app/produk');
		expect(storeMore).toContain('/app/kebutuhan');
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
