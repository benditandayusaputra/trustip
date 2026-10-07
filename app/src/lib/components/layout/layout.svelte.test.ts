import { page, userEvent } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AppShell from './AppShell.svelte';
import PublicShell from './PublicShell.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { textSnippet } from '#lib/testing/snippet.js';

const mainNavigation = () => page.getByRole('navigation', { name: 'Navigasi utama' });

describe('AppShell on a phone', () => {
	it('shows five bottom items for a store and marks the current one', async () => {
		await page.viewport(360, 740);
		const { container } = await render(AppShell, {
			organizationType: 'STORE',
			currentPath: '/app/titipan/0192f0c4',
			children: textSnippet('Isi')
		});

		const nav = mainNavigation();
		await expect.element(nav.getByRole('link')).toHaveLength(4);
		await expect.element(nav.getByRole('button', { name: 'Lainnya' })).toBeVisible();
		await expect
			.element(nav.getByRole('link', { name: 'Titipan' }))
			.toHaveAttribute('aria-current', 'page');
		await expectAccessible(container);
	});

	it('opens the remaining sections from Lainnya', async () => {
		await page.viewport(360, 740);
		await render(AppShell, {
			organizationType: 'CONSIGNOR',
			currentPath: '/app/penitip',
			children: textSnippet('Isi')
		});

		await userEvent.click(page.getByRole('button', { name: 'Lainnya' }));

		const sheet = page.getByRole('dialog', { name: 'Lainnya' });
		await expect.element(sheet.getByRole('link', { name: 'Produk' })).toBeVisible();
	});
});

describe('AppShell on a laptop', () => {
	it('uses the sidebar and hides the bottom bar', async () => {
		await page.viewport(1280, 800);
		await render(AppShell, {
			organizationType: 'CONSIGNOR',
			currentPath: '/app/mitra',
			children: textSnippet('Isi')
		});

		const nav = mainNavigation();
		await expect
			.element(nav.getByRole('link', { name: 'Mitra' }))
			.toHaveAttribute('aria-current', 'page');
		await expect.element(nav.getByRole('link', { name: 'Pengaturan' })).toBeVisible();
		await expect.element(page.getByRole('button', { name: 'Lainnya' })).not.toBeInTheDocument();
	});
});

describe('PublicShell', () => {
	it('offers a skip link, site navigation, and policy links', async () => {
		const { container } = await render(PublicShell, { children: textSnippet('Isi') });

		await expect.element(page.getByRole('link', { name: 'Lewati ke konten' })).toBeInTheDocument();
		await expect.element(page.getByRole('link', { name: 'Masuk' })).toBeVisible();
		await expect.element(page.getByRole('link', { name: 'Kebijakan privasi' })).toBeVisible();
		await expectAccessible(container);
	});
});
