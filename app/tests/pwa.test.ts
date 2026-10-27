import { expect, test } from '@playwright/test';
import { chooseTheme } from './theme';

interface WebManifest {
	name: string;
	start_url: string;
	display: string;
	icons: { sizes: string; purpose: string }[];
	shortcuts: { name: string; url: string }[];
}

test('the manifest makes the app installable with two shortcuts', async ({ page, request }) => {
	await page.goto('/');
	const href = await page.locator('link[rel="manifest"]').getAttribute('href');

	const manifest = (await (await request.get(href ?? '')).json()) as WebManifest;

	expect(manifest).toMatchObject({ name: 'Trustip', display: 'standalone', start_url: '/' });
	expect(manifest.icons.map((icon) => `${icon.sizes} ${icon.purpose}`)).toEqual([
		'192x192 any',
		'512x512 any',
		'512x512 maskable'
	]);
	expect(manifest.shortcuts.map((shortcut) => shortcut.name)).toEqual([
		'Hitung sisa',
		'Catat titipan'
	]);
});

test('the service worker takes control and serves the offline page without a network', async ({
	page,
	context
}) => {
	await page.goto('/');
	await page.evaluate(async () => {
		await navigator.serviceWorker.ready;
	});
	await page.reload();
	await expect
		.poll(() => page.evaluate(() => navigator.serviceWorker.controller !== null))
		.toBe(true);
	await chooseTheme(page, 'dark');

	await context.setOffline(true);
	await page.goto('/kebutuhan');

	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Tanpa koneksi');
	await expect(page.locator('html')).toHaveAttribute('data-appearance', 'dark');
	await chooseTheme(page, 'light');
	await expect(page.locator('html')).toHaveAttribute('data-appearance', 'light');
	await context.setOffline(false);
});
