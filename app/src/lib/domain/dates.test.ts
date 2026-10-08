import { describe, expect, it } from 'vitest';
import { addDays, daysBetween } from './dates';

describe('business dates', () => {
	it('adds days across the end of a leap February', () => {
		expect(addDays('2028-02-27', 3)).toBe('2028-03-01');
	});

	it('goes back across a month', () => {
		expect(addDays('2026-11-01', -3)).toBe('2026-10-29');
	});

	it('counts whole days between two dates', () => {
		expect(daysBetween('2026-11-10', '2026-11-12')).toBe(2);
		expect(daysBetween('2026-11-10', '2026-11-09')).toBe(-1);
	});
});
