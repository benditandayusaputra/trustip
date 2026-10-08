import { describe, expect, it } from 'vitest';
import type { Insight } from '#lib/api/insights.js';
import { cardOf } from './cards';

const match: Insight = {
	kind: 'MATCHING_PARTNER',
	level: 'LOW',
	action: 'OPEN_PROFILE',
	subject: {
		partnership_id: null,
		counterpart_name: '',
		product_id: null,
		product_name: '',
		batch_id: null,
		batch_code: '',
		settlement_id: null,
		settlement_number: '',
		organization_id: '0192f0c4-0000-7000-8000-000000000b01',
		organization_name: 'Kantin Sehat'
	},
	units: 0,
	days: 0,
	amount: 0,
	score: 82
};

describe('insight cards', () => {
	it('INS-10 opens the matching organization in Cari Mitra for each side', () => {
		expect(cardOf(match, 0, 'CONSIGNOR')).toMatchObject({
			title: 'Kantin Sehat cocok untuk Anda',
			body: 'Skor kecocokan 82 dari 100.',
			href: '/cari/toko?q=Kantin%20Sehat'
		});
		expect(cardOf(match, 0, 'STORE').href).toBe('/cari/penitip?q=Kantin%20Sehat');
	});

	it('INS-01 opens the count with the expired units filled in as returns', () => {
		const expired: Insight = {
			...match,
			kind: 'EXPIRED',
			level: 'HIGH',
			action: 'COUNT_WITH_RETURNS',
			subject: { ...match.subject, partnership_id: 'p1', product_id: 'roti', batch_code: 'B1' },
			units: 4
		};

		expect(cardOf(expired, 0, 'STORE').href).toBe(
			'/app/hitung/baru?mitra=p1&produk=roti&kembali=4'
		);
	});
});
