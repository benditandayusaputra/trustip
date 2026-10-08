import { describe, expect, it } from 'vitest';
import { expiredReturnOf, withExpiredReturn, type CountLineValues } from './schema';

function line(productId: string, countedQty: number): CountLineValues {
	const values = { countedQty, returnedQty: 0, returnReason: '' as const };
	return {
		productId,
		name: productId,
		unit: 'pcs',
		recordedQty: countedQty,
		...values,
		input: { productId, batches: [], ...values }
	};
}

describe('expired return', () => {
	it('INS-01 reads the product and quantity from the address', () => {
		const params = new URLSearchParams({ mitra: 'p', produk: 'roti', kembali: '4' });

		expect(expiredReturnOf(params)).toEqual({ productId: 'roti', qty: 4 });
	});

	it('ignores a missing product or a quantity that is not a positive whole number', () => {
		expect(expiredReturnOf(new URLSearchParams({ kembali: '4' }))).toBeUndefined();
		expect(expiredReturnOf(new URLSearchParams({ produk: 'roti', kembali: '0' }))).toBeUndefined();
		expect(expiredReturnOf(new URLSearchParams({ produk: 'roti', kembali: 'x' }))).toBeUndefined();
	});

	it('INS-01 fills the return as expired without exceeding what is left', () => {
		const lines = withExpiredReturn([line('roti', 3), line('keripik', 5)], {
			productId: 'roti',
			qty: 4
		});

		expect(lines[0]).toMatchObject({ returnedQty: 3, returnReason: 'EXPIRED' });
		expect(lines[1]).toMatchObject({ returnedQty: 0, returnReason: '' });
	});
});
