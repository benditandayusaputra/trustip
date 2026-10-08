import { expect, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { networkTimeoutMs, signIn } from './session';

const queuedToast = 'Tersimpan di perangkat, menunggu sinyal.';

interface DocumentCounts {
	drops: number;
	counts: number;
}

export function partnershipIdOf(partnerPage: string): string {
	return new URL(partnerPage).pathname.split('/').pop() ?? '';
}

function formPaths(partnershipId: string) {
	return {
		drop: `/app/titipan/baru?mitra=${partnershipId}`,
		count: `/app/hitung/baru?mitra=${partnershipId}`
	};
}

export async function documentCounts(page: Page, partnershipId: string): Promise<DocumentCounts> {
	const read = async (path: string) => {
		const response = await page.request.get(`${path}?partnership_id=${partnershipId}`);
		return ((await response.json()) as { data: unknown[] }).data.length;
	};
	return { drops: await read('/api/v1/drops'), counts: await read('/api/v1/counts') };
}

export async function prepareForOffline(page: Page, partnershipId: string): Promise<void> {
	const forms = formPaths(partnershipId);
	for (const path of ['/app/antrean', forms.count, forms.drop]) {
		await page.goto(path);
		await expect(page.locator('main h1')).toBeVisible({ timeout: networkTimeoutMs });
	}
	await page.evaluate(async () => {
		await navigator.serviceWorker.ready;
	});
	await expect(page.getByRole('textbox', { name: /^Jumlah/u })).toBeVisible({
		timeout: networkTimeoutMs
	});
}

export async function queueDrop(page: Page, partnershipId: string, qty: number): Promise<void> {
	await page.goto(formPaths(partnershipId).drop);
	await page.getByRole('textbox', { name: /^Jumlah/u }).fill(String(qty));
	await page.keyboard.press('Tab');
	await page.getByRole('button', { name: 'Simpan Titipan Masuk' }).click();
	await expect(page.getByText(queuedToast)).toBeVisible({ timeout: networkTimeoutMs });
	await page.waitForURL(/\/app\/antrean$/u);
}

export async function queueCount(page: Page, partnershipId: string, left: number): Promise<void> {
	await page.goto(formPaths(partnershipId).count);
	await page.getByRole('textbox', { name: /^Sisa fisik/u }).fill(String(left));
	await page.keyboard.press('Tab');
	await page.getByRole('button', { name: 'Simpan Hitung Sisa' }).click();
	await expect(page.getByText(queuedToast)).toBeVisible({ timeout: networkTimeoutMs });
	await page.waitForURL(/\/app\/antrean$/u);
}

export async function expectRecordedStock(page: Page, partnershipId: string, units: number) {
	await page.goto(formPaths(partnershipId).count);
	await expect(page.getByText(`Stok tercatat ${String(units)} pcs`)).toBeVisible({
		timeout: networkTimeoutMs
	});
}

export async function recordDropOnline(page: Page, partnershipId: string, qty: number) {
	await page.goto(formPaths(partnershipId).drop);
	await page.getByRole('textbox', { name: /^Jumlah/u }).fill(String(qty));
	await page.keyboard.press('Tab');
	await page.getByRole('button', { name: 'Simpan Titipan Masuk' }).click();
	await page.waitForURL(/\/app\/titipan\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
}

export async function secondDevice(browser: Browser, context: BrowserContext): Promise<Page> {
	const other = await browser.newContext({
		baseURL: 'http://localhost:5173',
		storageState: await context.storageState()
	});
	return other.newPage();
}

export async function fixReviewedCount(page: Page): Promise<void> {
	await page.goto('/app/antrean');
	await expect(page.getByText('Perlu ditinjau')).toBeVisible({ timeout: networkTimeoutMs });
	await page.getByRole('link', { name: 'Tinjau' }).click();
	await expect(
		page.getByText('Stok tercatat sekarang').filter({ visible: true }).first()
	).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.getByRole('link', { name: 'Perbaiki dan kirim ulang' }).click();
	await expect(page.getByText(/Anda memperbaiki catatan dari antrean/u)).toBeVisible({
		timeout: networkTimeoutMs
	});
	await page.getByRole('button', { name: 'Simpan Hitung Sisa' }).click();
	await page.waitForURL(/\/app\/hitung\/[0-9a-f-]{36}$/u, { timeout: networkTimeoutMs });
}

export async function signOutKeepingEntries(page: Page): Promise<void> {
	await page.goto('/app/antrean');
	await page.getByRole('button', { name: 'Keluar', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: /catatan belum terkirim$/u });
	await expect(dialog).toBeVisible({ timeout: networkTimeoutMs });
	await dialog.getByRole('button', { name: 'Simpan di perangkat' }).click();
	await page.waitForURL(/\/masuk/u, { timeout: networkTimeoutMs });
}

export async function signOut(page: Page): Promise<void> {
	await page.goto('/app/antrean');
	await page.getByRole('button', { name: 'Keluar', exact: true }).click();
	await page.waitForURL(/\/masuk/u, { timeout: networkTimeoutMs });
}

export async function signInAgain(page: Page, email: string): Promise<void> {
	await page.goto('/masuk?lanjut=%2Fapp%2Fantrean');
	await signIn(page, email);
	await page.waitForURL(/\/app\/antrean$/u, { timeout: networkTimeoutMs });
}
