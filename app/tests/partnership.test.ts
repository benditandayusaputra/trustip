import { devices, expect, test, type Browser, type Page } from '@playwright/test';
import {
	documentCounts,
	expectRecordedStock,
	fixReviewedCount,
	partnershipIdOf,
	prepareForOffline,
	queueCount,
	queueDrop,
	recordDropOnline,
	secondDevice,
	signInAgain,
	signOut,
	signOutKeepingEntries
} from './offline';
import { consignorWithProduct } from './catalog';
import { networkTimeoutMs, onboard, registerThenSignIn, uniqueEmail } from './session';

const appOrigin = 'http://localhost:5173';
const flowTimeoutMs = 600_000;
const partnerPagePattern = /\/app\/mitra\/[0-9a-f-]{36}$/u;
const liveNoticeMs = 3000;

async function openAs(browser: Browser, device: string): Promise<Page> {
	const context = await browser.newContext({ ...devices[device], baseURL: appOrigin });
	return context.newPage();
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

interface StoreSide {
	partnerPage: string;
	email: string;
}

async function storeAcceptsInvitation(page: Page, invitationPath: string): Promise<StoreSide> {
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
	return { partnerPage: page.url(), email };
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

async function storeRecordsWhileConsignorWatches(store: Page, consignor: Page): Promise<void> {
	const dropSaved = store.waitForResponse(
		(response) => response.url().endsWith('/api/v1/drops') && response.request().method() === 'POST'
	);
	const recorded = storeRecordsTwentyLoaves(store);
	await dropSaved;
	await expect(consignor.getByText('Ada pemberitahuan baru dari mitra.')).toBeVisible({
		timeout: liveNoticeMs
	});
	await recorded;
}

interface ProductPageScene {
	store: Page;
	consignor: Page;
	partnerPage: string;
}

async function buyerFindsTheStoreOnTheProductPage(
	browser: Browser,
	scene: ProductPageScene
): Promise<void> {
	await scene.store.goto(scene.partnerPage);
	await scene.store
		.getByRole('switch', { name: 'Tampilkan toko saya sebagai tempat membeli produk mitra ini' })
		.click();
	await expect(scene.store.getByText('Pilihan tampil disimpan.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await scene.consignor.goto('/app/produk');
	await scene.consignor.getByRole('link', { name: /Roti Manis Uji/u }).click();
	const publicLink = scene.consignor.getByRole('link', { name: 'Lihat halaman publik produk' });
	const publicPage = new URL((await publicLink.getAttribute('href')) ?? '').pathname;
	const buyer = await openAs(browser, 'Pixel 7');
	await buyer.goto(publicPage);
	await expect(buyer.getByRole('heading', { level: 1, name: 'Roti Manis Uji' })).toBeVisible();
	await expect(buyer.getByText('Warung Uji', { exact: true })).toBeVisible();
	await expect(buyer.getByRole('link', { name: 'Buka Warung Uji di peta' })).toBeVisible();
	await buyer.context().close();
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

test('FR-PRP-07 BR-PRP-03 FR-AGR-01 FR-DRP-01 FR-NTF-02 FR-PUB-04 FR-PUB-05 BR-PUB-01 FR-CNT-05 FR-STL-06 FR-PWA-06 BR-PWA-03 BR-PWA-05 BR-PWA-06 two businesses partner, trade 20 loaves through a paid recap, then the store records offline, reviews a stock change, and never sends for another user', async ({
	browser
}) => {
	test.setTimeout(flowTimeoutMs);
	const consignor = await openAs(browser, 'Pixel 7');
	const store = await openAs(browser, 'Desktop Chrome');
	const consignorEmail = await consignorWithProduct(consignor);
	const invitationPath = await sendInvitation(consignor);

	const { partnerPage, email: storeEmail } = await storeAcceptsInvitation(store, invitationPath);
	const seenByStore = await agreementSeenOn(store);
	await store.goto('/app');
	await expect(store).toHaveURL(/\/app\/toko$/u);
	await consignor.goto('/app/mitra');
	await consignor.getByRole('link', { name: /Warung Uji/u }).click();
	const seenByConsignor = await agreementSeenOn(consignor);

	expect(partnerPage).toMatch(partnerPagePattern);
	expect(seenByConsignor).toBe(seenByStore);

	await store.goto(partnerPage);
	await storeRecordsWhileConsignorWatches(store, consignor);
	await buyerFindsTheStoreOnTheProductPage(browser, { store, consignor, partnerPage });
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

	const partnershipId = partnershipIdOf(partnerPage);
	await storeWorksOffline(store, partnershipId);
	await storeReviewsAStockChangedElsewhere(browser, store, partnershipId);
	await anotherUserNeverSendsTheEntries(
		{ store, consignor, partnershipId },
		{ storeEmail, consignorEmail }
	);
});

async function storeWorksOffline(store: Page, partnershipId: string): Promise<void> {
	await prepareForOffline(store, partnershipId);
	const before = await documentCounts(store, partnershipId);
	await store.context().setOffline(true);
	await queueDrop(store, partnershipId, 10);
	await expectRecordedStock(store, partnershipId, 10);
	await queueCount(store, partnershipId, 4);
	await expect(store.getByRole('link', { name: 'Antrean, 2 catatan menunggu' })).toBeVisible();
	const secondTab = await store.context().newPage();
	await secondTab.goto('/app/antrean');
	await expect(secondTab.getByText('Menunggu sinyal').first()).toBeVisible();

	await store.context().setOffline(false);

	const sentOnce = { drops: before.drops + 1, counts: before.counts + 1 };
	await expect
		.poll(() => documentCounts(store, partnershipId), { timeout: networkTimeoutMs })
		.toEqual(sentOnce);
	await expect(store.getByRole('link', { name: /^Buka HS-[0-9]{6}-0002$/u })).toBeVisible({
		timeout: networkTimeoutMs
	});
	await expect(secondTab.getByText('Menunggu sinyal')).toHaveCount(0);
	expect(await documentCounts(store, partnershipId)).toEqual(sentOnce);
	await secondTab.close();
}

async function storeReviewsAStockChangedElsewhere(
	browser: Browser,
	store: Page,
	partnershipId: string
): Promise<void> {
	const otherDevice = await secondDevice(browser, store.context());
	await prepareForOffline(store, partnershipId);
	await store.context().setOffline(true);
	await queueCount(store, partnershipId, 1);
	await recordDropOnline(otherDevice, partnershipId, 3);
	await otherDevice.context().close();

	await store.context().setOffline(false);

	await fixReviewedCount(store);
	await expect(store.getByText('Segel sah')).toBeVisible({ timeout: networkTimeoutMs });
}

interface Devices {
	store: Page;
	consignor: Page;
	partnershipId: string;
}

interface Emails {
	storeEmail: string;
	consignorEmail: string;
}

async function anotherUserNeverSendsTheEntries(devices: Devices, emails: Emails): Promise<void> {
	const { store, consignor, partnershipId } = devices;
	const blockDrops = '**/api/v1/drops';
	await store.route(blockDrops, (route) =>
		route.request().method() === 'POST' ? route.abort() : route.continue()
	);
	await queueDrop(store, partnershipId, 2);
	await signOutKeepingEntries(store);
	await store.unroute(blockDrops);
	const kept = await documentCounts(consignor, partnershipId);

	await signInAgain(store, emails.consignorEmail);

	await expect(store.getByText('Tidak ada catatan yang menunggu.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	expect(await documentCounts(consignor, partnershipId)).toEqual(kept);
	await signOut(store);
	await signInAgain(store, emails.storeEmail);
	await expect
		.poll(() => documentCounts(consignor, partnershipId), { timeout: networkTimeoutMs })
		.toEqual({ ...kept, drops: kept.drops + 1 });
}
