import { describe, expect, it } from 'vitest';
import { findViolations } from './check-design-tokens.js';

const rulesIn = (text, file = 'src/lib/components/ui/Card.svelte') =>
	findViolations(file, text).map((found) => found.rule);

describe('findViolations', () => {
	it('flags hex colors outside token files', () => {
		expect(rulesIn('<div style="color: #FF453A"></div>')).toContain('raw-hex-color');
	});

	it('allows hex colors inside the token files', () => {
		expect(rulesIn('--color-danger: #ff453a;', 'src/lib/styles/tokens.css')).toEqual([]);
	});

	it('flags color functions', () => {
		expect(rulesIn('background: rgb(0 0 0 / 60%);', 'src/lib/styles/app.css')).toContain(
			'raw-color-function'
		);
	});

	it('flags default Tailwind palette classes', () => {
		expect(rulesIn('<p class="text-red-500 bg-white">x</p>')).toEqual([
			'default-palette-class',
			'default-palette-class'
		]);
	});

	it('flags arbitrary size classes', () => {
		expect(rulesIn('<p class="text-[15px] p-[13px]">x</p>')).toEqual([
			'arbitrary-size-class',
			'arbitrary-size-class'
		]);
	});

	it('flags spacing steps that have no token', () => {
		expect(rulesIn('<p class="p-3 gap-7 mt-4">x</p>')).toEqual([
			'spacing-outside-tokens',
			'spacing-outside-tokens'
		]);
	});

	it('allows token classes and css variables', () => {
		const markup =
			'<p class="bg-surface text-text-muted p-4 md:p-8 rounded-card z-(--z-nav) w-full">x</p>';
		expect(rulesIn(markup)).toEqual([]);
	});

	it('flags raw font sizes in css', () => {
		expect(rulesIn('h1 { font-size: 24px; }', 'src/lib/styles/app.css')).toEqual(['raw-css-size']);
	});

	it('does not mistake anchors for colors', () => {
		expect(rulesIn('<a href="#konten">Lewati</a>')).toEqual([]);
	});

	it('reports the line of each violation', () => {
		expect(findViolations('src/x.svelte', '<p>\n<b class="p-3">x</b>')[0]?.line).toBe(2);
	});
});

describe('findViolations on font metrics', () => {
	it('does not mistake line-gap-override for a raw gap', () => {
		expect(rulesIn('line-gap-override: 0%;', 'src/lib/styles/fonts.css')).toEqual([]);
	});
});
