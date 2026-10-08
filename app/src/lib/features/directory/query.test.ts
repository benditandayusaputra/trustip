import { describe, expect, it } from 'vitest';
import { defaultStoreQuery, storeQueryFrom, storeSearchParams } from './query';

describe('store directory query', () => {
	it('FR-DIR-03 FR-DIR-05 starts nearest within 10 km with an empty address', () => {
		expect(storeQueryFrom(new URLSearchParams())).toEqual(defaultStoreQuery());
		expect(storeSearchParams(defaultStoreQuery()).toString()).toBe('');
	});

	it('FR-DIR-12 keeps every filter in the address and reads it back', () => {
		const query = {
			...defaultStoreQuery(),
			origin: 'CITY' as const,
			cityCode: '33.74',
			radius: '25' as const,
			sort: 'BEST_MATCH' as const,
			storeType: 'CANTEEN',
			categoryIds: ['a', 'b'],
			withoutHalal: true,
			hasOpenNeed: true,
			text: 'kopi'
		};

		const address = storeSearchParams(query);

		expect(address.toString()).toBe(
			'origin=CITY&city_code=33.74&radius_km=25&sort=BEST_MATCH&q=kopi' +
				'&store_type=CANTEEN&without_halal=true&has_open_need=true&category_id=a&category_id=b'
		);
		expect(storeQueryFrom(address)).toEqual(query);
	});

	it('BR-DIR-06 tells the default categories apart from every category', () => {
		const every = storeQueryFrom(new URLSearchParams('category_id=any'));

		expect(storeQueryFrom(new URLSearchParams()).categoryIds).toBeNull();
		expect(every.categoryIds).toEqual(['any']);
	});

	it('ignores unknown choices in a shared address', () => {
		const query = storeQueryFrom(new URLSearchParams('radius_km=3&sort=CHEAPEST&origin=MOON'));

		expect(query.radius).toBe('10');
		expect(query.sort).toBe('NEAREST');
		expect(query.origin).toBe('ORGANIZATION');
	});

	it('FR-DIR-04 sends the current position only for the current position origin', () => {
		const here = { ...defaultStoreQuery(), origin: 'HERE' as const, lat: '-7.05', lng: '110.43' };

		expect(storeSearchParams(here).toString()).toBe('origin=HERE&lat=-7.05&lng=110.43');
	});
});
