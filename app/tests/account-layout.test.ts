import { expect, test } from '@playwright/test';
import { networkTimeoutMs } from './session';

test('directory input survives account loading and More navigation reaches products', async ({
	page
}) => {
	await page.goto('/masuk');
	await page.getByLabel('Email').fill('penitip@demo.trustip.id');
	await page.getByLabel('Kata sandi').fill('Coba-Demo-2026');
	await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
		timeout: networkTimeoutMs
	});
	await page.getByRole('button', { name: 'Masuk', exact: true }).click();
	await page.waitForURL(/\/app\/penitip$/u, { timeout: networkTimeoutMs });

	const account = Promise.withResolvers<undefined>();
	await page.route('**/api/v1/me', async (route) => {
		await account.promise;
		await route.continue();
	});
	const search = page.getByLabel('Cari nama atau deskripsi toko');
	try {
		await page.goto('/cari/toko');
		await search.fill('Warung Bu Sri');
		await expect(search).toHaveValue('Warung Bu Sri');
	} finally {
		account.resolve(undefined);
	}
	await expect(page.locator('summary')).toHaveText('Akun', { timeout: networkTimeoutMs });
	await expect(search).toHaveValue('Warung Bu Sri');
	await expect(page.getByRole('heading', { name: 'Warung Bu Sri', exact: true })).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.setViewportSize({ width: 390, height: 844 });
	await page.goto('/app/penitip');
	await page.getByRole('button', { name: 'Lainnya' }).click({ timeout: networkTimeoutMs });
	const more = page.getByRole('dialog', { name: 'Lainnya' });
	await more.getByRole('link', { name: 'Produk', exact: true }).click();
	await page.waitForURL(/\/app\/produk$/u, { timeout: networkTimeoutMs });
	await expect(more).not.toBeVisible();
});
