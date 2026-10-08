import { describe, expect, it } from 'vitest';
import type { CountSheet } from '#lib/api/counts.js';
import type { OutboxEntry } from '#lib/offline/outbox.js';
import { withQueuedDrops } from './queued';

const terms = {
	model: 'FIXED_DEPOSIT' as const,
	retailPrice: 2000,
	commissionBps: 0,
	depositPrice: 1750
};

const sheet: CountSheet = {
	partnership_id: 'p1',
	is_active: true,
	agreement_version: 2,
	store: { id: 's1', name: 'Warung Bu Sri' },
	consignor: { id: 'c1', name: 'Roti Pak Anton' },
	last_document_at: '2026-11-09T02:00:00Z',
	products: [
		{
			product_id: 'bread',
			product_name: 'Roti Manis',
			sku: 'ROT-0001',
			unit: 'pcs',
			recorded_qty: 5,
			batches: [
				{
					id: 'b1',
					expiry_date: '2026-11-12',
					received_at: '2026-11-09T02:00:00Z',
					balance: 5,
					terms: {
						pricing_model: 'FIXED_DEPOSIT',
						retail_price: 2000,
						commission_bps: 0,
						deposit_price: 1750
					}
				}
			]
		}
	]
};

function queuedDrop(changes: Partial<OutboxEntry> = {}): OutboxEntry {
	return {
		id: 'entry-1',
		revision: 0,
		idempotencyKey: 'entry-1',
		kind: 'DROP',
		partnershipId: 'p1',
		partnerName: 'Roti Pak Anton',
		request: { partnership_id: 'p1', agreement_version: 2, occurred_at: null, items: [] },
		lines: [
			{
				productId: 'bread',
				name: 'Roti Manis',
				unit: 'pcs',
				qty: 20,
				expiryDate: '2026-11-13',
				terms
			},
			{ productId: 'cake', name: 'Bolu', unit: 'pcs', qty: 4, expiryDate: '', terms }
		],
		recordedAt: '2026-11-10T02:00:00.000Z',
		savedOffline: true,
		userId: 'u1',
		organizationId: 's1',
		status: 'PENDING',
		attempts: 0,
		problem: null,
		document: null,
		...changes
	} as OutboxEntry;
}

describe('withQueuedDrops (R-OFF-04)', () => {
	it('adds a drop still in the queue to the recorded stock of the work packet', () => {
		const merged = withQueuedDrops(sheet, [queuedDrop()]);

		expect(merged.products.map((product) => [product.product_id, product.recorded_qty])).toEqual([
			['bread', 25],
			['cake', 4]
		]);
		expect(merged.products[0]?.batches.map((batch) => batch.id)).toEqual(['b1', 'entry-1/1']);
		expect(merged.last_document_at).toBe('2026-11-10T02:00:00.000Z');
	});

	it('leaves out drops already sent and drops for another partner', () => {
		const merged = withQueuedDrops(sheet, [
			queuedDrop({ status: 'SENT' }),
			queuedDrop({ partnershipId: 'p2' })
		]);

		expect(merged).toEqual(sheet);
	});
});
