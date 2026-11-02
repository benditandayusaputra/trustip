import { expect, test } from '@playwright/test';
import { networkTimeoutMs } from './session';

const widths = [320, 768, 1440];
const themes = ['light', 'dark'] as const;
const publicPaths = [
	'/',
	'/masuk',
	'/daftar',
	'/privasi',
	'/panduan',
	'/faq',
	'/cari/toko',
	'/halaman-tidak-ada'
];

for (const width of widths) {
	for (const theme of themes) {
		test(`public pages fit ${String(width)} pixels in ${theme} theme`, async ({ page }) => {
			await page.setViewportSize({ width, height: 900 });
			await page.addInitScript((chosen) => {
				localStorage.setItem('trustip-appearance', chosen);
			}, theme);
			await page.emulateMedia({ colorScheme: theme });
			for (const path of publicPaths) {
				await page.goto(path, { waitUntil: 'load' });
				if (path === '/masuk' || path === '/daftar') {
					await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
						timeout: networkTimeoutMs
					});
				}
				await expect(page.getByRole('status', { name: 'Memuat' })).toHaveCount(0, {
					timeout: networkTimeoutMs
				});
				await expect(page.locator('html')).toHaveAttribute('data-appearance', theme);
				const dimensions = await page.evaluate(() => ({
					viewport: window.innerWidth,
					content: document.documentElement.scrollWidth
				}));
				expect(dimensions.content, path).toBeLessThanOrEqual(dimensions.viewport);
			}
		});
	}
}
