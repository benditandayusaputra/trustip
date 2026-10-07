import { describe, expect, it } from 'vitest';
import { safeHref } from './safe-href';

describe('safeHref (R-SEC-64)', () => {
	it.each([
		['/app/titipan/7', '/app/titipan/7'],
		['https://wa.me/6281234567890', 'https://wa.me/6281234567890'],
		['javascript:alert(1)', undefined],
		['//evil.example/app', undefined],
		['/\\evil.example', undefined],
		['http://plain.example', undefined],
		['data:text/html,x', undefined],
		['', undefined]
	])('turns %s into %s', (raw, expected) => {
		expect(safeHref(raw)).toBe(expected);
	});
});
