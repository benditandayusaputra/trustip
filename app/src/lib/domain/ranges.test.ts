import { describe, expect, it } from 'vitest';
import { rangeOf } from './ranges';

const custom = { from: '2026-01-01', to: '2026-01-31' };

describe('report ranges', () => {
	it.each([
		['TODAY', '2026-11-20', '2026-11-20'],
		['LAST_7', '2026-11-14', '2026-11-20'],
		['LAST_30', '2026-10-22', '2026-11-20'],
		['THIS_MONTH', '2026-11-01', '2026-11-20']
	] as const)('FR-RPT-07 %s ends today', (preset, from, to) => {
		expect(rangeOf(preset, '2026-11-20', custom)).toEqual({ from, to });
	});

	it('FR-RPT-07 last month covers the whole previous month, also across a year', () => {
		expect(rangeOf('LAST_MONTH', '2026-03-15', custom)).toEqual({
			from: '2026-02-01',
			to: '2026-02-28'
		});
		expect(rangeOf('LAST_MONTH', '2027-01-05', custom)).toEqual({
			from: '2026-12-01',
			to: '2026-12-31'
		});
	});

	it('FR-RPT-07 a custom range is used as chosen', () => {
		expect(rangeOf('CUSTOM', '2026-11-20', custom)).toEqual(custom);
	});
});
