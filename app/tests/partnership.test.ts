import { devices, expect, test, type Browser, type Page } from '@playwright/test';
import { networkTimeoutMs, onboard, registerThenSignIn, uniqueEmail } from './session';

const appOrigin = 'http://localhost:5173';
const flowTimeoutMs = 180_000;
const partnerPagePattern = /\/app\/mitra\/[0-9a-f-]{36}$/u;

async function openAs(browser: Browser, device: string): Promise<Page> {
	const context = await browser.newContext({ ...devices[device], baseURL: appOrigin });
	return context.newPage();
}

async function consignorWithProduct(page: Page): Promise<void> {
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
}

async function sendInvitation(page: Page): Promise<string> {
	await page.goto('/app/ajuan/baru');
	await page.getByLabel(/Roti Manis Uji/u).check();
	await expect(page.getByText('Rp17.500')).toBeVisible();
	await page.getByRole('button', { name: 'Buat tautan undangan' }).click();
	const link = page.getByText(/\/undangan\/[A-Z0-9]{10}$/u);
	await expect(link).toBeVisible({ timeout: networkTimeoutMs });
	return new URL((await link.textContent()) ?? '').pathname;
}

async function storeAcceptsInvitation(page: Page, invitationPath: string): Promise<string> {
	const email = uniqueEmail('store');
	await page.goto(invitationPath);
	await page.getByRole('link', { name: 'Daftar' }).click();
	await registerThenSignIn(page, email);
	await page.getByRole('link', { name: 'Daftarkan usaha' }).click();
	await onboard(page, { choice: /Toko/u, name: 'Warung Uji', dashboard: '/app/toko' });
	await expect(page.getByText('Roti Uji mengajak Anda bermitra.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.getByRole('button', { name: 'Buka Ajuan' }).click();
	await page.getByRole('button', { name: 'Terima' }).click();
	await page.getByRole('dialog').getByRole('button', { name: 'Terima' }).click();
	await expect(page.getByText('Ajuan diterima. Kesepakatan berlaku.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.getByRole('link', { name: 'Lihat kemitraan' }).click();
	await page.waitForURL(partnerPagePattern);
	return page.url();
}

async function agreementSeenOn(page: Page): Promise<string> {
	await page.getByRole('tab', { name: 'Kesepakatan' }).click();
	await expect(page.getByText('Segel sah')).toBeVisible({ timeout: networkTimeoutMs });
	await expect(page.getByText('Rp1.750').filter({ visible: true }).first()).toBeVisible();
	return (await page.getByText(/\/V1 · Versi 1/u).textContent()) ?? '';
}

test('FR-PRP-07 BR-PRP-03 FR-AGR-01 two businesses partner through an invitation link and see the same sealed agreement', async ({
	browser
}) => {
	test.setTimeout(flowTimeoutMs);
	const consignor = await openAs(browser, 'Pixel 7');
	const store = await openAs(browser, 'Desktop Chrome');
	await consignorWithProduct(consignor);
	const invitationPath = await sendInvitation(consignor);

	const partnerPage = await storeAcceptsInvitation(store, invitationPath);
	const seenByStore = await agreementSeenOn(store);
	await store.goto('/app');
	await expect(store).toHaveURL(/\/app\/toko$/u);
	await consignor.goto('/app/mitra');
	await consignor.getByRole('link', { name: /Warung Uji/u }).click();
	const seenByConsignor = await agreementSeenOn(consignor);

	expect(partnerPage).toMatch(partnerPagePattern);
	expect(seenByConsignor).toBe(seenByStore);
});
