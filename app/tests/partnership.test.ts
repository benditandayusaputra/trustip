import { devices, expect, test, type Browser, type Page } from '@playwright/test';
import { networkTimeoutMs, onboard, registerThenSignIn, uniqueEmail } from './session';

const appOrigin = 'http://localhost:5173';
const flowTimeoutMs = 300_000;
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
	await page.getByLabel('Jadwal setor').selectOption('ON_COUNT');
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

async function storeRecordsTwentyLoaves(page: Page): Promise<void> {
	await page.getByRole('link', { name: 'Catat titipan' }).first().click();
	await page.getByRole('textbox', { name: /^Jumlah/u }).fill('20');
	await page.keyboard.press('Tab');
	await expect(page.getByText('Rp35.000')).toBeVisible();
	await page.getByRole('button', { name: 'Simpan Titipan Masuk' }).click();
	await page.waitForURL(/\/app\/titipan\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
	await expect(page.getByText('Segel sah')).toBeVisible({ timeout: networkTimeoutMs });
}

async function storeCountsTwoLeftAndReturned(page: Page, partnerPage: string): Promise<string> {
	await page.goto(partnerPage);
	await page.getByRole('link', { name: 'Hitung sisa' }).click();
	await page.getByRole('textbox', { name: /^Sisa fisik/u }).fill('2');
	await page.keyboard.press('Tab');
	await page.getByRole('textbox', { name: /^Dikembalikan/u }).fill('2');
	await page.keyboard.press('Tab');
	await page.getByRole('combobox', { name: 'Alasan pengembalian' }).selectOption('UNSOLD');
	await page.getByRole('button', { name: 'Simpan Hitung Sisa' }).click();
	await page.waitForURL(/\/app\/hitung\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
	await expect(page.getByText('Segel sah')).toBeVisible({ timeout: networkTimeoutMs });
	return page.url();
}

async function expectRecap(page: Page): Promise<void> {
	for (const amount of ['Rp36.000', 'Rp4.500', 'Rp31.500']) {
		await expect(page.getByText(amount).filter({ visible: true }).first()).toBeVisible();
	}
	await expect(
		page.getByText('18', { exact: true }).filter({ visible: true }).first()
	).toBeVisible();
}

async function storePaysCash(page: Page, recapPage: string): Promise<string> {
	await page.goto(new URL(recapPage).pathname);
	await page.getByRole('link', { name: /^Catat bayar ST-/u }).click();
	await page.waitForURL(/\/app\/setoran\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
	await expect(page.getByText('Rp31.500').filter({ visible: true }).first()).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.getByRole('combobox', { name: 'Metode' }).selectOption('CASH');
	await page.getByRole('button', { name: 'Simpan pembayaran' }).click();
	await expect(page.getByText('Pembayaran tercatat. Menunggu penitip.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	return page.url();
}

async function consignorReceivesTheMoney(page: Page, settlementPage: string): Promise<void> {
	await page.goto(new URL(settlementPage).pathname);
	await page.getByRole('button', { name: 'Uang diterima' }).click();
	await page.getByRole('dialog').getByRole('button', { name: 'Uang diterima' }).click();
	await expect(page.getByText('Pembayaran diterima. Setoran lunas.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await expect(page.getByText('Lunas', { exact: true }).first()).toBeVisible();
}

test('FR-PRP-07 BR-PRP-03 FR-AGR-01 FR-DRP-01 FR-CNT-05 FR-STL-06 two businesses partner, then 20 loaves are dropped, 2 left and returned, the recap is confirmed, and the cash payment is received', async ({
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

	await store.goto(partnerPage);
	await storeRecordsTwentyLoaves(store);
	const recapPage = await storeCountsTwoLeftAndReturned(store, partnerPage);
	await expectRecap(store);
	await consignor.goto(new URL(recapPage).pathname);
	await expectRecap(consignor);
	await consignor.getByRole('button', { name: 'Sesuai' }).click();
	await consignor.getByRole('dialog').getByRole('button', { name: 'Sesuai' }).click();
	await expect(consignor.getByText('Hitung Sisa dinyatakan sesuai.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	const settlementPage = await storePaysCash(store, recapPage);
	await consignorReceivesTheMoney(consignor, settlementPage);
});
