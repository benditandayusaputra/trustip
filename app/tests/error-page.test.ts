import { expect, test } from '@playwright/test';

test('unknown address shows a friendly not found page', async ({ page }) => {
	const response = await page.goto('/alamat-yang-tidak-ada');

	expect(response?.status()).toBe(404);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Halaman tidak ditemukan');
	await expect(page.getByRole('link', { name: 'Kembali ke beranda' })).toBeVisible();
});
