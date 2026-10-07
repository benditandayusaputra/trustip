import { describe, expect, it } from 'vitest';
import {
	formatBusinessDate,
	formatDateTime,
	formatDistance,
	formatInteger,
	formatRupiah,
	parseDigits
} from './format';

describe('formatInteger', () => {
	it('groups thousands with dots as in Indonesian', () => {
		expect(formatInteger(1250000)).toBe('1.250.000');
	});

	it('keeps small numbers ungrouped', () => {
		expect(formatInteger(999)).toBe('999');
	});
});

describe('parseDigits', () => {
	it('reads digits from grouped text', () => {
		expect(parseDigits('Rp 31.500')).toBe(31500);
	});

	it('returns undefined when there are no digits', () => {
		expect(parseDigits('abc')).toBeUndefined();
	});
});

describe('formatRupiah', () => {
	it('writes rupiah without decimals and without a space, as in the specification', () => {
		expect(formatRupiah(36000)).toBe('Rp36.000');
		expect(formatRupiah(31500)).toBe('Rp31.500');
	});

	it('puts the minus sign before Rp', () => {
		expect(formatRupiah(-4500)).toBe('-Rp4.500');
	});

	it('shows zero as Rp0', () => {
		expect(formatRupiah(0)).toBe('Rp0');
	});
});

describe('formatBusinessDate', () => {
	it('keeps a Jakarta business date on the same day', () => {
		expect(formatBusinessDate('2026-10-19')).toBe('19 Okt 2026');
	});
});

describe('formatDateTime', () => {
	it('shows instants in Jakarta time', () => {
		expect(formatDateTime('2026-10-18T17:30:00Z')).toBe('19 Okt 2026, 00.30');
	});
});

describe('formatDistance', () => {
	it('uses meters below one kilometer', () => {
		expect(formatDistance(850)).toBe('850 m');
	});

	it('uses one decimal below ten kilometers', () => {
		expect(formatDistance(1234)).toBe('1,2 km');
	});

	it('drops decimals from ten kilometers up', () => {
		expect(formatDistance(12400)).toBe('12 km');
	});
});
