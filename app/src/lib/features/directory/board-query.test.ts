import { describe, expect, it } from 'vitest';
import { boardQueryFrom, boardSearchParams, defaultBoardQuery } from './board-query';

describe('need board query', () => {
	it('FR-NED-03 starts nearest within 10 km', () => {
		expect(boardSearchParams(defaultBoardQuery()).toString()).toBe('');
	});

	it('FR-DIR-12 keeps the board filters in the address and only allows nearest or newest', () => {
		const address = new URLSearchParams('sort=BEST_MATCH&category_id=c1&radius_km=25');

		const query = boardQueryFrom(address);

		expect(query.sort).toBe('NEAREST');
		expect(boardSearchParams({ ...query, sort: 'NEWEST' }).toString()).toBe(
			'radius_km=25&sort=NEWEST&category_id=c1'
		);
	});
});
