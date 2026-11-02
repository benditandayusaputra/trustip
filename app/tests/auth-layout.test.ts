import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { expect, test, type Page } from '@playwright/test';
import { networkTimeoutMs } from './session';

const axeSource = readFileSync(createRequire(import.meta.url).resolve('axe-core'), 'utf8');
const widths = [320, 390, 768, 1024, 1440];
const themes = ['light', 'dark'] as const;
const next = '%2Fapp%2Fmitra';

async function checkPage(page: Page): Promise<void> {
	await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
	await expect(page.getByText('Coba akun demo')).toHaveCount(0);
	await expect(page.getByText(/@demo\.trustip\.id/u)).toHaveCount(0);
	await expect(page.getByLabel('Email')).toHaveValue('');
	await expect(page.getByLabel('Kata sandi')).toHaveValue('');
	await expect(page.locator('header img')).toHaveAttribute('src', '/icons/icon-192.png');
	for (const theme of themes) {
		await page.evaluate((chosen) => {
			localStorage.setItem('trustip-appearance', chosen);
		}, theme);
		await page.reload();
		await expect(page.locator('html')).toHaveAttribute('data-appearance', theme);
		for (const width of widths) {
			await page.setViewportSize({ width, height: 900 });
			await expect
				.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
				.toBeLessThanOrEqual(width);
		}
		await page.evaluate(async () => {
			await Promise.all(document.getAnimations().map((animation) => animation.finished));
		});
		await page.evaluate(axeSource);
		const violations = await page.evaluate(async () => {
			const result = await window.axe.run({ resultTypes: ['violations'] });
			return result.violations.map((item) => ({
				rule: item.id,
				targets: item.nodes.map((node) => node.target)
			}));
		});
		expect(violations).toEqual([]);
	}
}

test('auth pages keep demo accounts separate and preserve accessible form controls', async ({
	page
}) => {
	await page.goto(`/masuk?lanjut=${next}`);
	await checkPage(page);
	await page.getByRole('main').getByRole('link', { name: 'Daftar', exact: true }).click();
	await expect(page).toHaveURL(new RegExp(`/daftar\\?lanjut=${next}$`, 'u'));
	await checkPage(page);
	const input = page.getByLabel('Kata sandi');
	await input.fill('Sandi-uji-12345');
	const reveal = page.getByRole('button', { name: 'Tampilkan sandi' });
	await reveal.focus();
	await reveal.press('Enter');
	await expect(input).toHaveAttribute('type', 'text');
	const hide = page.getByRole('button', { name: 'Sembunyikan sandi' });
	await expect(hide).toBeFocused();
	await hide.press('Space');
	await expect(input).toHaveAttribute('type', 'password');
	await expect(input).toHaveValue('Sandi-uji-12345');
	await page.getByRole('main').getByRole('link', { name: 'Masuk', exact: true }).click();
	await expect(page).toHaveURL(new RegExp(`/masuk\\?lanjut=${next}$`, 'u'));
});

test('the demo account page is hidden from visitors', async ({ page }) => {
	await page.goto('/');
	await expect(page.locator('a[href="/demo"]')).toHaveCount(0);
	await page.goto('/demo');
	await expect(page.getByRole('heading', { name: 'Halaman tidak ditemukan' })).toBeVisible();
});

test('auth pages fit 320 pixels after the security widget has loaded', async ({ page }) => {
	await page.setViewportSize({ width: 320, height: 900 });
	for (const path of ['/masuk', '/daftar']) {
		await page.goto(path);
		await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
			timeout: networkTimeoutMs
		});
		await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth)).toBe(320);
	}
});
