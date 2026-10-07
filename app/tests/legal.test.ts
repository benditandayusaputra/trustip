import { expect, test } from '@playwright/test';

test('FR-PUB-09 the privacy policy and terms open from the footer', async ({ page }) => {
	await page.goto('/');
	const footer = page.getByRole('navigation', { name: 'Tautan kebijakan' });

	await footer.getByRole('link', { name: 'Kebijakan privasi' }).click();
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Kebijakan Privasi');
	await expect(page.getByText(/Cloudflare Turnstile/u)).toBeVisible();

	await footer.getByRole('link', { name: 'Ketentuan' }).click();
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('Ketentuan Layanan');
});
