import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import HomePage from './components/HomePage.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

describe('HomePage', () => {
	it('FR-PUB-01 explains the problem and the four steps and links to sign up', async () => {
		const { container } = await render(HomePage);

		await expect
			.element(page.getByRole('heading', { name: 'Masalah titip jual hari ini' }))
			.toBeVisible();
		await expect.element(page.getByRole('heading', { name: 'Kesepakatan Dagang' })).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Daftar gratis' }).first())
			.toHaveAttribute('href', '/daftar');
		await expectAccessible(container);
	});
});
