import { describe, expect, it } from 'vitest';
import { reasonText } from './reasons';

describe('match reasons', () => {
	it('FR-DIR-10 reads like the card examples', () => {
		expect(reasonText({ code: 'DISTANCE', distance_m: 1200 })).toBe('1,2 km');
		expect(reasonText({ code: 'NEEDS_CATEGORY', category: 'Kue kering dan camilan' })).toBe(
			'Butuh Kue kering dan camilan'
		);
		expect(reasonText({ code: 'ON_TIME_PAYMENT', on_time_bps: 9600 })).toBe(
			'Bayar tepat waktu 96%'
		);
	});

	it('names every reason code', () => {
		const codes = [
			'ACCEPTS_CATEGORY',
			'ESTABLISHED',
			'PREFERRED_STORE_TYPE',
			'ACTIVE_RECENTLY'
		] as const;

		for (const code of codes) {
			expect(reasonText({ code, category: 'Minuman' })).not.toBe('');
		}
	});
});
