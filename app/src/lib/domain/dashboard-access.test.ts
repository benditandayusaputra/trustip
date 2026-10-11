import { describe, expect, it } from 'vitest';
import { dashboardDestination } from './dashboard-access';

describe('dashboard access follows the current organization', () => {
	it.each([
		['/app/toko', 'CONSIGNOR', '/app/penitip'],
		['/app/penitip?old=1', 'STORE', '/app/toko'],
		['/app/toko/nested', 'CONSIGNOR', '/app/penitip'],
		['/app/toko?period=30', 'STORE', '/app/toko?period=30'],
		['/app/stok', 'CONSIGNOR', '/app/stok'],
		['/undangan/ABC', 'STORE', '/undangan/ABC'],
		['/app', 'CONSIGNOR', '/app/penitip']
	])('resolves %s for %s', (path, type, expected) => {
		expect(dashboardDestination(path, type)).toBe(expected);
	});
});
