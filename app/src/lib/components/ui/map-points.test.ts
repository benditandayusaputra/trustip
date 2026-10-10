import { describe, expect, it } from 'vitest';
import { markerGroups } from './map-points';

const first = { id: 'a', label: 'Toko A', latitude: -7, longitude: 110 };
const same = { ...first, id: 'b', label: 'Toko B' };
const elsewhere = { ...first, id: 'c', longitude: 111 };

describe('Map marker groups', () => {
	it('groups coincident coordinates while keeping every store selectable', () => {
		const groups = markerGroups([first, same, elsewhere]);
		expect(groups).toEqual([
			{ point: first, members: [first, same] },
			{ point: elsewhere, members: [elsewhere] }
		]);
	});
	it('leaves empty and single-location results intact', () => {
		expect(markerGroups([])).toEqual([]);
		expect(markerGroups([first])).toEqual([{ point: first, members: [first] }]);
	});
});
