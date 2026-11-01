import { devices, expect, test, type Browser, type Page } from '@playwright/test';
import { networkTimeoutMs, registerThenSignIn, uniqueEmail } from './session';

const appOrigin = 'http://localhost:5173';
const knownHosts = new Set([
	'localhost:5173',
	'localhost:8080',
	'127.0.0.1:8080',
	'challenges.cloudflare.com',
	'tile.openstreetmap.org'
]);
const payload = '<img src=x onerror=alert(1)><script>alert(2)</script>';
const flowTimeoutMs = 300_000;

interface Watch {
	problems: string[];
}

async function watchedPage(browser: Browser, watch: Watch): Promise<Page> {
	const context = await browser.newContext({ ...devices['Desktop Chrome'], baseURL: appOrigin });
	const page = await context.newPage();
	page.on('dialog', (dialog) => {
		watch.problems.push(`dialog: ${dialog.message()}`);
		void dialog.dismiss();
	});
	page.on('console', (message) => {
		if (message.text().includes('Content Security Policy')) {
			watch.problems.push(`csp: ${message.text()}`);
		}
	});
	page.on('request', (request) => {
		const url = new URL(request.url());
		if (url.protocol.startsWith('http') && !knownHosts.has(url.host)) {
			watch.problems.push(`request: ${request.url()}`);
		}
	});
	return page;
}

async function onboardWithPayload(page: Page, choice: RegExp, name: string): Promise<void> {
	await page.goto('/daftar?lanjut=%2Fapp');
	await registerThenSignIn(page, uniqueEmail('xss'));
	await page.getByRole('radio', { name: choice }).check();
	await page.getByRole('button', { name: 'Lanjut' }).click();
	await page.getByLabel('Nama usaha').fill(name);
	await page.getByLabel('Deskripsi singkat').fill(payload);
	await page.getByLabel('Nomor WhatsApp usaha').fill('0812 3456 7890');
	await page.getByRole('button', { name: 'Lanjut' }).click();
	await page.getByLabel('Provinsi').selectOption({ label: 'Jawa Tengah' });
	await expect(page.getByRole('option', { name: 'Kota Semarang' })).toBeAttached();
	await page.getByLabel('Kabupaten atau kota').selectOption({ label: 'Kota Semarang' });
	await page.getByLabel('Alamat').fill(payload);
	await page.context().grantPermissions(['geolocation']);
	await page.context().setGeolocation({ latitude: -7.0546, longitude: 110.4386 });
	await page.getByRole('button', { name: 'Gunakan lokasi saya' }).click();
	await expect(page.getByText('Lokasi ditandai.')).toBeVisible();
	await page.getByRole('button', { name: 'Simpan dan mulai' }).click();
	await expect(page).toHaveURL(/\/app\/(penitip|toko)$/u, { timeout: networkTimeoutMs });
}

async function productWithPayload(page: Page): Promise<void> {
	await page.goto('/app/produk/baru');
	await page.getByLabel('Nama produk').fill(`Roti ${payload}`);
	await expect(page.getByRole('option', { name: 'Roti dan kue basah' })).toBeAttached();
	await page.getByLabel('Kategori').selectOption({ label: 'Roti dan kue basah' });
	await page.getByLabel('Deskripsi').fill(payload);
	await page.getByLabel('Komposisi').fill(payload);
	await page.getByLabel('Harga jual saran').fill('2000');
	await page.getByLabel('Harga setor saran').fill('1750');
	await page.getByRole('button', { name: 'Simpan produk' }).click();
	await expect(page.getByText('Produk tersimpan.')).toBeVisible({ timeout: networkTimeoutMs });
}

async function listStore(page: Page): Promise<void> {
	await page.goto('/app/profil');
	await page.getByRole('button', { name: 'Edit profil' }).click();
	await page.getByLabel('Jenis toko').selectOption('BAKERY');
	await page.getByRole('checkbox', { name: 'Roti dan kue basah' }).check();
	await page.getByLabel('Syarat lain').fill(payload);
	await page.getByRole('button', { name: 'Simpan profil', exact: true }).click();
	await expect(page.getByText('Profil usaha tersimpan.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.goto('/app/pengaturan');
	await page.getByRole('switch', { name: 'Tampilkan usaha saya di Cari Mitra' }).click();
	await expect(page.getByText('Usaha Anda tampil di Cari Mitra.')).toBeVisible({
		timeout: networkTimeoutMs
	});
}

async function needWithPayload(page: Page, title: string): Promise<string> {
	await page.goto('/app/kebutuhan/baru');
	await page.getByLabel('Judul').fill(title);
	await expect(page.getByRole('option', { name: 'Roti dan kue basah' })).toBeAttached();
	await page.getByLabel('Kategori').selectOption({ label: 'Roti dan kue basah' });
	await page.getByLabel('Deskripsi').fill(payload);
	await page.getByRole('button', { name: 'Pasang Kebutuhan' }).click();
	await page.waitForURL(/\/app\/kebutuhan\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
	return new URL(page.url()).pathname.split('/').pop() ?? '';
}

async function proposeWithPayload(page: Page, storeStamp: string): Promise<string> {
	await page.goto('/cari/toko');
	await page.getByLabel('Cari nama atau deskripsi toko').fill(storeStamp);
	await page
		.getByRole('article')
		.filter({ hasText: storeStamp })
		.getByRole('link', { name: 'Ajukan kerja sama' })
		.click({ timeout: networkTimeoutMs });
	await page.getByLabel(/^Roti /u).check();
	await page.getByLabel('Harga setor').fill('1750');
	await page.getByLabel('Jadwal setor').selectOption('ON_COUNT');
	await page.getByRole('textbox', { name: 'Catatan' }).fill(payload);
	await page.getByLabel('Pesan pengantar').fill(payload);
	await page.getByRole('button', { name: 'Kirim ajuan' }).click();
	await page.waitForURL(/\/app\/ajuan\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
	return new URL(page.url()).pathname;
}

async function visitAll(page: Page, paths: string[]): Promise<void> {
	for (const path of paths) {
		await page.goto(path);
		await page.waitForLoadState('networkidle');
	}
}

test.describe.configure({ timeout: flowTimeoutMs });

test(
	'R-SEC-69 script payloads saved in free text fields never run on the pages that show them',
	{ tag: '@security' },
	async ({ browser }) => {
		const watch: Watch = { problems: [] };
		const consignor = await watchedPage(browser, watch);
		await onboardWithPayload(consignor, /Penitip/u, `Roti ${payload}`);
		await productWithPayload(consignor);
		await visitAll(consignor, ['/app/penitip', '/app/produk', '/app/pengaturan', '/app/profil']);

		const store = await watchedPage(browser, watch);
		const storeStamp = `Toko ${Date.now().toString(36)}`;
		await onboardWithPayload(store, /Toko/u, `${storeStamp} ${payload}`);
		await listStore(store);
		const needId = await needWithPayload(store, `Butuh roti ${Date.now().toString(36)}`);
		await visitAll(store, [
			'/app/toko',
			`/app/kebutuhan/${needId}`,
			'/app/pengaturan',
			'/app/profil'
		]);
		const proposalPath = await proposeWithPayload(consignor, storeStamp);
		await visitAll(store, [proposalPath, '/app/ajuan']);
		await visitAll(consignor, ['/kebutuhan', `/kebutuhan/${needId}`]);
		await consignor.getByRole('link', { name: 'Lihat profil toko' }).click();
		await consignor.waitForLoadState('networkidle');
		await expect(consignor.getByText(payload).first()).toBeVisible();

		expect(watch.problems).toEqual([]);
	}
);

test(
	'R-SEC-21 a write sent from another site is refused',
	{ tag: '@security' },
	async ({ request }) => {
		const response = await request.post('/api/v1/products', {
			headers: { Origin: 'https://situs-lain.example', 'X-Trustip-Client': 'web' },
			data: {}
		});

		expect(response.status()).toBe(403);
	}
);
