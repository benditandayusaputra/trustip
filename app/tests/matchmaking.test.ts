import { devices, expect, test, type Browser, type Locator, type Page } from '@playwright/test';
import { consignorWithProduct, listedStore } from './catalog';
import { networkTimeoutMs } from './session';

const appOrigin = 'http://localhost:5173';
const flowTimeoutMs = 300_000;
const ownServer = { tag: '@own-server' };
const proposalPagePattern = /\/app\/ajuan\/[0-9a-f-]{36}$/u;

async function openAs(browser: Browser, device: string): Promise<Page> {
	const context = await browser.newContext({ ...devices[device], baseURL: appOrigin });
	return context.newPage();
}

function stamp(): string {
	return Date.now().toString(36);
}

async function sendOpenProposal(page: Page, depositPrice: string): Promise<string> {
	await page.getByLabel(/Roti Manis Uji/u).check();
	await page.getByLabel('Harga setor').fill(depositPrice);
	await page.getByLabel('Jadwal setor').selectOption('ON_COUNT');
	await page.getByRole('button', { name: 'Kirim ajuan' }).click();
	await page.waitForURL(proposalPagePattern, { timeout: networkTimeoutMs });
	return new URL(page.url()).pathname;
}

function cardTitled(page: Page, title: string): Locator {
	return page.getByRole('article').filter({ has: page.getByRole('heading', { name: title }) });
}

async function accept(page: Page): Promise<void> {
	await page.getByRole('button', { name: 'Terima' }).click();
	await page.getByRole('dialog').getByRole('button', { name: 'Terima' }).click();
	await expect(page.getByText('Ajuan diterima. Kesepakatan berlaku.')).toBeVisible({
		timeout: networkTimeoutMs
	});
}

test.describe.configure({ mode: 'serial', timeout: flowTimeoutMs });

let consignor: Page;

test.beforeAll(async ({ browser }) => {
	consignor = await openAs(browser, 'Pixel 7');
	await consignorWithProduct(consignor);
});

test(
	'FR-DIR-01 FR-PRP-01 FR-PRP-03 FR-PRP-04 Alur A: the consignor finds a store, proposes, the store counters, and the consignor accepts',
	ownServer,
	async ({ browser }) => {
		const store = await openAs(browser, 'Desktop Chrome');
		const storeName = `Toko Cari ${stamp()}`;
		await listedStore(store, storeName);

		await consignor.goto('/cari/toko');
		await consignor.getByLabel('Cari nama atau deskripsi toko').fill(storeName);
		await cardTitled(consignor, storeName)
			.getByRole('link', { name: 'Ajukan kerja sama' })
			.click({ timeout: networkTimeoutMs });
		await expect(
			consignor.getByRole('heading', { name: `Ajukan kerja sama ke ${storeName}` })
		).toBeVisible({
			timeout: networkTimeoutMs
		});
		const proposalPath = await sendOpenProposal(consignor, '1800');

		await store.goto(proposalPath);
		await store.getByRole('button', { name: 'Tawar' }).click();
		await store.getByLabel('Harga setor').fill('1750');
		await store.getByRole('button', { name: 'Kirim tawaran' }).click();
		await expect(store.getByText('Tawaran terkirim.')).toBeVisible({ timeout: networkTimeoutMs });

		await consignor.goto(proposalPath);
		await expect(consignor.getByRole('heading', { name: 'Riwayat tawar-menawar' })).toBeVisible({
			timeout: networkTimeoutMs
		});
		await accept(consignor);
		await consignor.getByRole('link', { name: 'Lihat kemitraan' }).click();
		await consignor.getByRole('tab', { name: 'Kesepakatan' }).click();
		await expect(consignor.getByText('Segel sah')).toBeVisible({ timeout: networkTimeoutMs });
		await expect(consignor.getByText('Rp1.750').filter({ visible: true }).first()).toBeVisible();
	}
);

test(
	'FR-NED-01 FR-NED-03 FR-NED-05 FR-NED-06 Alur B: the store posts a Need, a consignor responds, and the store accepts',
	ownServer,
	async ({ browser }) => {
		const store = await openAs(browser, 'Desktop Chrome');
		const needTitle = `Butuh roti manis ${stamp()}`;
		await listedStore(store, `Toko Butuh ${stamp()}`);

		await store.goto('/app/kebutuhan/baru');
		await store.getByLabel('Judul').fill(needTitle);
		await expect(store.getByRole('option', { name: 'Roti dan kue basah' })).toBeAttached();
		await store.getByLabel('Kategori').selectOption({ label: 'Roti dan kue basah' });
		await store.getByRole('button', { name: 'Pasang Kebutuhan' }).click();
		await store.waitForURL(/\/app\/kebutuhan\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
		const needPath = new URL(store.url()).pathname;

		await consignor.goto('/kebutuhan');
		await consignor.getByLabel('Cari judul atau deskripsi Kebutuhan').fill(needTitle);
		await cardTitled(consignor, needTitle)
			.getByRole('link', { name: 'Tanggapi dengan Ajuan' })
			.click({ timeout: networkTimeoutMs });
		await expect(
			consignor.getByText('Ajuan ini menanggapi Kebutuhan toko.', { exact: false })
		).toBeVisible({
			timeout: networkTimeoutMs
		});
		await sendOpenProposal(consignor, '1750');

		await store.goto(needPath);
		await store.getByRole('link', { name: /Roti Uji/u }).click();
		await store.waitForURL(proposalPagePattern);
		await accept(store);
	}
);
