import { describe, expect, it } from 'vitest';
import { registerHref, safeNextPath, signInHref, signedInPath } from './links';

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

describe('login landing uses the new account', () => {
	const account = {
		id: 'u',
		full_name: 'Sari',
		email: 'sari@contoh.id',
		membership: {
			organization_id: 'o',
			organization_type: 'CONSIGNOR' as const,
			role: 'OWNER' as const
		}
	};
	it('replaces the previous store dashboard with the consignor dashboard', () => {
		expect(signedInPath('/app/toko?old=1', account)).toBe('/app/penitip');
	});
	it('keeps a valid invitation destination', () => {
		expect(signedInPath('/undangan/ABC', account)).toBe('/undangan/ABC');
	});
	it('rejects external redirects before choosing the dashboard', () => {
		expect(signedInPath('//evil.example', account)).toBe('/app/penitip');
	});
	it('sends newcomers through onboarding', () => {
		expect(signedInPath('/undangan/ABC', { ...account, membership: null })).toBe(
			'/onboarding?lanjut=%2Fundangan%2FABC'
		);
	});
});
