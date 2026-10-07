import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { themeColor } from './tokens';

describe('tokens.ts', () => {
	it('keeps the theme color equal to the page background token', () => {
		const css = readFileSync(new URL('./tokens.css', import.meta.url), 'utf8');
		expect(css).toContain(`--color-bg: ${themeColor};`);
	});
});
