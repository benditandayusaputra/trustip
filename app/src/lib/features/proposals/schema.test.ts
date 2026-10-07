import { describe, expect, it } from 'vitest';
import {
	commissionBps,
	emptyInvitation,
	termsOf,
	toInvitationRequest,
	validateInvitation,
	type ItemValues
} from './schema';

const bread: ItemValues = {
	productId: '0192f0c4-0000-7000-8000-000000000101',
	name: 'Roti Manis',
	unit: 'pcs',
	model: 'FIXED_DEPOSIT',
	retailPrice: 2000,
	depositPrice: 1750,
	commissionPercent: '20'
};

describe('invitation schema', () => {
	it('BR-AGR-03 reads a commission percentage with a decimal comma as basis points', () => {
		expect(commissionBps('17,5')).toBe(1750);
		expect(commissionBps('20')).toBe(2000);
		expect(commissionBps('')).toBeUndefined();
		expect(commissionBps('abc')).toBeUndefined();
	});

	it('FR-PRP-02 sends only the fields of the chosen pricing model', () => {
		const chips: ItemValues = { ...bread, model: 'COMMISSION', commissionPercent: '17,5' };
		const values = { ...emptyInvitation(), items: [bread, chips] };

		expect(toInvitationRequest(values).items).toEqual([
			{
				product_id: bread.productId,
				pricing_model: 'FIXED_DEPOSIT',
				retail_price: 2000,
				commission_bps: 0,
				deposit_price: 1750
			},
			{
				product_id: bread.productId,
				pricing_model: 'COMMISSION',
				retail_price: 2000,
				commission_bps: 1750,
				deposit_price: 0
			}
		]);
	});

	it('BR-AGR-02 and BR-AGR-03 flag the row with the wrong terms', () => {
		const values = {
			...emptyInvitation(),
			items: [
				{ ...bread, depositPrice: 2500 },
				{ ...bread, model: 'COMMISSION' as const, commissionPercent: '51' }
			]
		};

		expect(Object.keys(validateInvitation(values))).toEqual([
			'items[0].deposit_price',
			'items[1].commission_bps'
		]);
	});

	it('maps empty money fields to zero terms', () => {
		expect(termsOf({ ...bread, retailPrice: undefined, depositPrice: undefined })).toEqual({
			model: 'FIXED_DEPOSIT',
			retailPrice: 0,
			commissionBps: 0,
			depositPrice: 0
		});
	});
});
