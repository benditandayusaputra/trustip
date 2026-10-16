import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ContentSkeleton from '../ui/ContentSkeleton.svelte';
import type { SkeletonVariant } from '../ui/skeleton';
import NavigationLoadingHarness from '#lib/testing/NavigationLoadingHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const variants: SkeletonVariant[] = [
	'list',
	'cards',
	'catalog',
	'detail',
	'form',
	'report',
	'profile',
	'product',
	'home',
	'article',
	'auth'
];

afterEach(() => {
	delete document.documentElement.dataset.appearance;
});

describe('Content loading skeletons', () => {
	for (const theme of ['light', 'dark']) {
		it.each(variants)(
			`renders an accessible %s placeholder in ${theme} on a small screen`,
			async (variant) => {
				document.documentElement.dataset.appearance = theme;
				await page.viewport(320, 740);
				const { container } = await render(ContentSkeleton, { variant });
				await expect.element(page.getByRole('status', { name: 'Memuat' })).toHaveLength(1);
				expect(container.querySelector('[aria-busy="true"]')).not.toBeNull();
				expect(container.querySelector('button, a, input, select')).toBeNull();
				expect(container.scrollWidth).toBeLessThanOrEqual(320);
				await expectAccessible(container);
			}
		);
	}
});

describe('Pending navigation', () => {
	it('keeps previous fields mounted and preserves their values when loading is cancelled', async () => {
		const { container } = await render(NavigationLoadingHarness);
		const original = container.querySelector('input');
		await userEvent.type(page.getByLabelText('Isian sebelumnya'), 'Catatan titipan');
		await userEvent.click(page.getByRole('button', { name: 'Memuat' }));
		await expect.element(page.getByLabelText('Isian sebelumnya')).not.toBeVisible();
		expect(original?.closest('[inert]')).not.toBeNull();
		await expect.element(page.getByRole('status', { name: 'Memuat' })).toHaveLength(1);
		await userEvent.click(page.getByRole('button', { name: 'Batalkan' }));
		await expect.element(page.getByLabelText('Isian sebelumnya')).toHaveValue('Catatan titipan');
		expect(container.querySelector('input')).toBe(original);
		expect(original?.closest('[inert]')).toBeNull();
	});
});
