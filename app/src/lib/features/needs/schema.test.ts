import { describe, expect, it } from 'vitest';
import { emptyNeedValues, toNeedRequest, validateNeed } from './schema';

describe('need form', () => {
	it('BR-NED-02 asks for a title of 5 to 80 characters and a category', () => {
		const errors = validateNeed({ ...emptyNeedValues(), title: 'Roti' });

		expect(errors.title).toBe('Judul 5 sampai 80 karakter.');
		expect(errors.categoryId).toBe('Pilih kategori');
	});

	it('FR-NED-01 sends empty optional fields as null and commission in basis points', () => {
		const request = toNeedRequest({
			...emptyNeedValues(),
			title: ' Butuh camilan ',
			categoryId: 'c1',
			model: 'COMMISSION',
			commissionPercent: '17,5'
		});

		expect(request).toEqual({
			title: 'Butuh camilan',
			category_id: 'c1',
			description: '',
			est_qty_per_week: null,
			max_retail_price: null,
			preferred_model: 'COMMISSION',
			offered_commission_bps: 1750
		});
	});
});
