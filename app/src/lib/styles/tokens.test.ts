import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { themeColor } from './tokens';

describe('tokens.ts', () => {
	it('keeps the theme color equal to the page background token', () => {
		const css = readFileSync(new URL('./tokens.css', import.meta.url), 'utf8');
		expect(css).toContain(`--color-bg: ${themeColor};`);
	});

	it('keeps the install manifest colors equal to the theme color', () => {
		const manifest = JSON.parse(
			readFileSync(new URL('../../../static/manifest.webmanifest', import.meta.url), 'utf8')
		) as { background_color: string; theme_color: string };
		expect(manifest.theme_color.toLowerCase()).toBe(themeColor);
		expect(manifest.background_color.toLowerCase()).toBe(themeColor);
	});
});
