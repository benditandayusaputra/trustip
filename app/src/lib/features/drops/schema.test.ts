import { describe, expect, it } from 'vitest';
import { errorKeyOf, isExpirySoon, toDropRequest, validateDrop, type DropLine } from './schema';

const line = (overrides: Partial<DropLine> = {}): DropLine => ({
	productId: 'bread',
	name: 'Roti Manis',
	unit: 'pcs',
	qty: 20,
	expiryDate: '2026-11-13',
	shelfLifeDays: 3,
	terms: { model: 'FIXED_DEPOSIT', retailPrice: 2000, commissionBps: 0, depositPrice: 1750 },
	...overrides
});

describe('drop schema', () => {
	it('BR-DRP-05 warns only for long shelf lives expiring within two days', () => {
		const longLife = line({ shelfLifeDays: 30, expiryDate: '2026-11-11' });

		expect(isExpirySoon(longLife, '2026-11-10')).toBe(true);
		expect(isExpirySoon({ ...longLife, expiryDate: '2026-11-12' }, '2026-11-10')).toBe(false);
		expect(isExpirySoon({ ...longLife, shelfLifeDays: 7 }, '2026-11-10')).toBe(false);
	});

	it('BR-DRP-05 refuses an expiry before the drop date', () => {
		expect(validateDrop([line({ expiryDate: '2026-11-09' })], '2026-11-10')).toEqual({
			'bread.expiry': 'Tanggal kedaluwarsa tidak boleh sebelum tanggal titip.'
		});
	});

	it('BR-DRP-06 sends a backdated drop at noon in Jakarta and leaves out empty lines', () => {
		const request = toDropRequest({
			partnershipId: 'p',
			agreementVersion: 1,
			dropDate: '2026-11-08',
			today: '2026-11-10',
			lines: [line({ expiryDate: '' }), line({ productId: 'chips', qty: 0 })]
		});

		expect(request).toEqual({
			partnership_id: 'p',
			agreement_version: 1,
			occurred_at: '2026-11-08T12:00:00+07:00',
			items: [{ product_id: 'bread', qty: 20, expiry_date: null }]
		});
	});

	it('maps server fields back to the product that was sent at that index', () => {
		const sent = [line(), line({ productId: 'chips' })];

		expect(errorKeyOf('items[1].expiry_date', sent)).toBe('chips.expiry');
		expect(errorKeyOf('items[0].qty', sent)).toBe('bread.qty');
		expect(errorKeyOf('occurred_at', sent)).toBe('dropDate');
		expect(errorKeyOf('items', sent)).toBe('items');
	});
});
