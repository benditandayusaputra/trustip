import { describe, expect, it } from 'vitest';
import { publicProductUrl } from './links';

describe('publicProductUrl', () => {
	it('FR-PRD-07 points the QR at the public product page of the consignor', () => {
		expect(
			publicProductUrl('https://trustip.id', { consignor_slug: 'roti-pak-anton', sku: 'ROT-0001' })
		).toBe('https://trustip.id/p/roti-pak-anton/ROT-0001');
	});
});
