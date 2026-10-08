import { expect, type Page } from '@playwright/test';
import { networkTimeoutMs, onboard, registerThenSignIn, uniqueEmail } from './session';

export async function consignorWithProduct(page: Page): Promise<string> {
	const email = uniqueEmail('consignor');
	await page.goto('/daftar?lanjut=%2Fapp');
	await registerThenSignIn(page, email);
	await onboard(page, { choice: /Penitip/u, name: 'Roti Uji', dashboard: '/app/penitip' });
	await expect(page).toHaveURL(/\/app\/penitip$/u, { timeout: networkTimeoutMs });
	await page.goto('/app/produk/baru');
	await page.getByLabel('Nama produk').fill('Roti Manis Uji');
	await expect(page.getByRole('option', { name: 'Roti dan kue basah' })).toBeAttached();
	await page.getByLabel('Kategori').selectOption({ label: 'Roti dan kue basah' });
	await page.getByLabel('Harga jual saran').fill('2000');
	await page.getByLabel('Harga setor saran').fill('1750');
	await page.getByRole('button', { name: 'Simpan produk' }).click();
	await expect(page.getByText('Produk tersimpan.')).toBeVisible({ timeout: networkTimeoutMs });
	return email;
}

export async function listedStore(page: Page, name: string): Promise<void> {
	await page.goto('/daftar?lanjut=%2Fapp');
	await registerThenSignIn(page, uniqueEmail('store'));
	await onboard(page, { choice: /Toko/u, name, dashboard: '/app/toko' });
	await expect(page).toHaveURL(/\/app\/toko$/u, { timeout: networkTimeoutMs });
	await page.goto('/app/pengaturan');
	await page.getByLabel('Jenis toko').selectOption('BAKERY');
	await page.getByRole('checkbox', { name: 'Roti dan kue basah' }).check();
	await page.getByRole('button', { name: 'Simpan profil toko' }).click();
	await expect(page.getByText('Profil toko tersimpan.')).toBeVisible({ timeout: networkTimeoutMs });
	await page.getByRole('switch', { name: 'Tampilkan usaha saya di Cari Mitra' }).click();
	await expect(page.getByText('Usaha Anda tampil di Cari Mitra.')).toBeVisible({
		timeout: networkTimeoutMs
	});
}
