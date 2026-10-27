import { page, userEvent } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ThemePicker from './ThemePicker.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

describe('ThemePicker', () => {
	it('opens a solid menu with only light and dark for the page script to apply', async () => {
		const { container } = await render(ThemePicker, {});

		await userEvent.click(page.getByLabelText('Mode tampilan', { exact: true }));
		const menu = page.getByRole('group', { name: 'Pilihan mode tampilan' });

		await expect.element(menu).toBeVisible();
		expect(container.querySelector('details')?.open).toBe(true);
		await expect.element(menu.getByRole('button', { name: 'Terang' })).toBeVisible();
		await expect.element(menu.getByRole('button', { name: 'Gelap' })).toBeVisible();
		await expect.element(menu.getByRole('button', { name: 'Sistem' })).not.toBeInTheDocument();
		expect(
			[...container.querySelectorAll('[data-appearance-choice]')].map((button) =>
				button.getAttribute('data-appearance-choice')
			)
		).toEqual(['light', 'dark']);
		await expectAccessible(container);
	});

	it('shows both choices side by side in the phone menu', async () => {
		const { container } = await render(ThemePicker, { expanded: true });

		await expect.element(page.getByRole('group', { name: 'Pilihan mode tampilan' })).toBeVisible();
		expect(container.querySelector('details')).toBeNull();
	});
});
