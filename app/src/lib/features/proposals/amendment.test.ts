import { describe, expect, it } from 'vitest';
import { toAmendmentRequest, type AmendmentRow } from './amendment.svelte';
import { emptyInvitation, type ItemValues } from './schema';

const item = (productId: string): ItemValues => ({
	productId,
	name: productId,
	unit: 'pcs',
	model: 'FIXED_DEPOSIT',
	retailPrice: 2000,
	depositPrice: 1750,
	commissionPercent: '20'
});

describe('amendment request', () => {
	it('FR-PRP-08 sends only changed products with their action', () => {
		const rows: AmendmentRow[] = [
			{ mode: 'KEEP', item: item('kept') },
			{ mode: 'STOP', item: item('stopped') },
			{ mode: 'UPDATE', item: item('updated') },
			{ mode: 'ADD', item: item('added') }
		];

		const request = toAmendmentRequest(emptyInvitation(), rows);

		expect(request.items.map((line) => [line.product_id, line.action])).toEqual([
			['stopped', 'STOP'],
			['updated', 'UPDATE'],
			['added', 'ADD']
		]);
	});
});
