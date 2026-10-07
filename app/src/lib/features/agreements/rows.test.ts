import { describe, expect, it } from 'vitest';
import type { Agreement } from '#lib/api/agreements.js';
import { productRows } from './rows';

const item = {
	product_id: 'p1',
	product_name: 'Keripik Pedas',
	sku: 'KER-0001',
	unit: 'bungkus',
	pricing_model: 'COMMISSION' as const,
	retail_price: 15000,
	commission_bps: 2000,
	deposit_price: 0,
	is_active: false,
	per_unit: { gross: 15000, store_share: 3000, consignor_share: 12000 }
};

describe('agreement product rows', () => {
	it('FR-AGR-01 shows retail price, consignor share, and store share per unit', () => {
		const rows = productRows({ items: [item] } as unknown as Agreement);

		expect(rows).toEqual([
			{
				id: 'p1',
				product: 'Keripik Pedas (KER-0001) · Dihentikan',
				model: 'Komisi 20%',
				retail: 'Rp15.000',
				consignor: 'Rp12.000',
				store: 'Rp3.000'
			}
		]);
	});
});
