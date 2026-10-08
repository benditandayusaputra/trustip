import { describe, expect, it } from 'vitest';
import type { Agreement } from '#lib/api/agreements.js';
import type { CountSheet } from '#lib/api/counts.js';
import type { OutboxEntry } from '#lib/offline/outbox.js';
import { countReviewRows, dropReviewRows, fixHrefOf } from './review-rows';

const terms = {
	model: 'FIXED_DEPOSIT' as const,
	retailPrice: 2000,
	commissionBps: 0,
	depositPrice: 1750
};

const base = {
	revision: 0,
	partnerName: 'Roti Pak Anton',
	recordedAt: '2026-11-10T02:00:00.000Z',
	savedOffline: true,
	userId: 'u1',
	organizationId: 's1',
	status: 'NEEDS_REVIEW' as const,
	attempts: 1,
	problem: null,
	document: null
};

const count: Extract<OutboxEntry, { kind: 'COUNT' }> = {
	...base,
	id: 'c1',
	idempotencyKey: 'c1',
	kind: 'COUNT',
	partnershipId: 'p1',
	request: {
		partnership_id: 'p1',
		agreement_version: 2,
		counted_at: null,
		occasion: 'STORE_CLOSING',
		note: '',
		lines: [
			{
				product_id: 'bread',
				recorded_qty: 18,
				counted_qty: 2,
				returned_qty: 0,
				return_reason: null
			}
		]
	}
};

const drop: Extract<OutboxEntry, { kind: 'DROP' }> = {
	...base,
	id: 'd1',
	idempotencyKey: 'd1',
	kind: 'DROP',
	partnershipId: 'p1',
	request: { partnership_id: 'p1', agreement_version: 1, occurred_at: null, items: [] },
	lines: [
		{ productId: 'bread', name: 'Roti Manis', unit: 'pcs', qty: 20, expiryDate: '', terms },
		{ productId: 'cake', name: 'Bolu', unit: 'pcs', qty: 4, expiryDate: '2026-11-12', terms }
	]
};

describe('review rows (R-OFF-09)', () => {
	it('puts the stock the user saw next to the stock recorded now', () => {
		const sheet = {
			products: [{ product_id: 'bread', product_name: 'Roti Manis', recorded_qty: 20 }]
		} as CountSheet;

		expect(countReviewRows(count, sheet)).toEqual([
			{
				productId: 'bread',
				product: 'Roti Manis',
				seen: '18',
				latest: '20',
				counted: '2',
				returned: '0'
			}
		]);
	});

	it('marks products that left the agreement in force', () => {
		const agreement = {
			version: 2,
			items: [
				{ product_id: 'bread', is_active: true },
				{ product_id: 'cake', is_active: false }
			]
		} as Agreement;

		expect(dropReviewRows(drop, agreement).map((row) => row.agreed)).toEqual(['Ya', 'Tidak']);
	});

	it('opens the matching form with the entry to fix', () => {
		expect(fixHrefOf(count)).toBe('/app/hitung/baru?mitra=p1&antrean=c1');
		expect(fixHrefOf(drop)).toBe('/app/titipan/baru?mitra=p1&antrean=d1');
	});
});
