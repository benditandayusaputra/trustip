import { describe, expect, it } from 'vitest';
import {
	consignorAddress,
	consignorQueryFrom,
	consignorSearchParams,
	defaultConsignorQuery
} from './consignor-query';

describe('consignor directory query', () => {
	it('FR-DIR-02 opens on the products tab with an empty address', () => {
		expect(consignorQueryFrom(new URLSearchParams())).toEqual(defaultConsignorQuery());
		expect(consignorAddress(defaultConsignorQuery()).toString()).toBe('');
	});

	it('FR-DIR-07 FR-DIR-12 keeps the tab and filters in the address', () => {
		const query = {
			...defaultConsignorQuery(),
			tab: 'penitip' as const,
			maxPrice: '20000',
			minShelfLifeDays: '7',
			hasHalal: true,
			deliversToMe: true
		};

		const address = consignorAddress(query);

		expect(address.toString()).toBe(
			'max_price=20000&min_shelf_life_days=7&has_halal=true&delivers_to_me=true&tab=penitip'
		);
		expect(consignorQueryFrom(address)).toEqual(query);
	});

	it('does not send the tab to the API', () => {
		const query = { ...defaultConsignorQuery(), tab: 'penitip' as const, canSample: true };

		expect(consignorSearchParams(query).toString()).toBe('can_sample=true');
	});
});
