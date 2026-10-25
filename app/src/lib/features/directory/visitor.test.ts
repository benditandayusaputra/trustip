import { describe, expect, it } from 'vitest';
import { defaultSearchQuery } from './query';
import { forVisitor, isRandomOrder, positionOf, shuffled } from './visitor';

describe('visitor search', () => {
	it('shows a visitor every distance and leaves a member radius alone', () => {
		const query = defaultSearchQuery();

		expect(forVisitor(query, true).radius).toBe('all');
		expect(forVisitor(query, false).radius).toBe(query.radius);
	});

	it('orders a visitor without a place at random unless a sort without distance is chosen', () => {
		const query = defaultSearchQuery();

		expect(isRandomOrder(query, true)).toBe(true);
		expect(isRandomOrder({ ...query, sort: 'NEWEST' }, true)).toBe(false);
		expect(isRandomOrder({ ...query, origin: 'HERE' }, true)).toBe(false);
		expect(isRandomOrder(query, false)).toBe(false);
	});

	it('shuffles without losing or repeating items', () => {
		const items = ['a', 'b', 'c', 'd'];

		const got = shuffled(items, () => 0);

		expect(got).toEqual(['b', 'c', 'd', 'a']);
		expect(items).toEqual(['a', 'b', 'c', 'd']);
	});

	it('turns a device position into a search from here', () => {
		expect(positionOf({ latitude: -7.9440123, longitude: 112.6150987 })).toEqual({
			origin: 'HERE',
			lat: '-7.94401',
			lng: '112.61510'
		});
	});
});
