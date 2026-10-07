import { describe, expect, it } from 'vitest';
import { formatInteger, parseDigits } from './format';

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
