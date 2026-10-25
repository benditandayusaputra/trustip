import { readdirSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { pageSkeletonKind, pendingPagePath, type PageSkeletonKind } from './page-skeleton';

const routes: Record<PageSkeletonKind, string[]> = {
	home: ['/'],
	auth: ['/masuk', '/daftar', '/demo'],
	article: ['/privasi', '/ketentuan', '/panduan', '/faq'],
	profile: ['/t/sample', '/u/sample'],
	product: ['/p/sample/sample'],
	form: [
		'/onboarding',
		'/app/pengaturan',
		'/app/produk/baru',
		'/app/produk/sample',
		'/app/titipan/baru',
		'/app/hitung/baru',
		'/app/ajuan/baru',
		'/app/kebutuhan/baru',
		'/app/mitra/sample/perubahan'
	],
	report: ['/app/laporan'],
	catalog: ['/app/produk'],
	cards: ['/cari/toko', '/cari/penitip', '/kebutuhan'],
	'store-dashboard': ['/app', '/app/toko'],
	'consignor-dashboard': ['/app/penitip'],
	list: [
		'/app/mitra',
		'/app/titipan',
		'/app/hitung',
		'/app/setoran',
		'/app/stok',
		'/app/kebutuhan',
		'/app/ajuan',
		'/app/antrean',
		'/app/pemberitahuan'
	],
	detail: [
		'/luring',
		'/kebutuhan/sample',
		'/undangan/sample',
		'/app/titipan/sample',
		'/app/hitung/sample',
		'/app/setoran/sample',
		'/app/kebutuhan/sample',
		'/app/ajuan/sample',
		'/app/antrean/sample',
		'/app/mitra/sample',
		'/app/mitra/sample/stok/sample',
		'/app/mitra/sample/kesepakatan',
		'/app/mitra/sample/kesepakatan/sample'
	]
};

describe('Page loading patterns', () => {
	it('covers every existing page route', () => {
		const existing = readdirSync(new URL('../../../routes/', import.meta.url), { recursive: true })
			.map((file) => file.toString())
			.filter((file) => file.endsWith('+page.svelte'))
			.map(
				(file) =>
					'/' +
					file
						.replace(/\([^/]+\)\//gu, '')
						.replace(/\[[^\]]+\]/gu, 'sample')
						.replace(/\/?\+page.svelte$/u, '')
			);
		expect(Object.values(routes).flat().sort()).toEqual(existing.sort());
	});
	for (const [kind, paths] of Object.entries(routes)) {
		it.each(paths)(`uses ${kind} for %s`, (path) => {
			expect(pageSkeletonKind(path)).toBe(kind);
		});
	}
	it('handles query parameters without changing the page shape', () => {
		expect(pageSkeletonKind('/app/titipan/baru?mitra=sample#form')).toBe('form');
	});
	it('ignores navigation within the same page and exposes the destination of a new page', () => {
		const current = new URL('https://trustip.test/cari/toko');
		expect(pendingPagePath(undefined, current)).toBeUndefined();
		expect(
			pendingPagePath(new URL('https://trustip.test/cari/toko?q=roti'), current)
		).toBeUndefined();
		expect(pendingPagePath(new URL('https://trustip.test/masuk'), current)).toBe('/masuk');
	});
});
