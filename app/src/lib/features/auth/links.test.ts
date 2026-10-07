import { describe, expect, it } from 'vitest';
import { registerHref, safeNextPath, signInHref } from './links';

describe('safeNextPath (R-SEC-49)', () => {
	it.each([
		['/undangan/KODE123', '/undangan/KODE123'],
		['/app/titipan?mitra=7#baris', '/app/titipan?mitra=7#baris'],
		['/app/../app/stok', '/app/stok']
	])('keeps the same-origin path %s', (raw, expected) => {
		expect(safeNextPath(raw)).toBe(expected);
	});

	it.each([
		null,
		'',
		'https://evil.example/app',
		'//evil.example/app',
		'/\\evil.example',
		'javascript:alert(1)',
		'app/stok',
		'\t/app'
	])('replaces %s with the dashboard', (raw) => {
		expect(safeNextPath(raw)).toBe('/app');
	});
});

describe('account links', () => {
	it('carry the destination along', () => {
		expect(signInHref('/undangan/A B')).toBe('/masuk?lanjut=%2Fundangan%2FA+B');
		expect(registerHref(null)).toBe('/daftar');
	});
});
