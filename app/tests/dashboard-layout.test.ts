import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { expect, test, type Page, type Locator } from '@playwright/test';
import { dashboardBoard, mockDashboard } from './dashboard-fixture';

const axeSource = readFileSync(createRequire(import.meta.url).resolve('axe-core'), 'utf8');

function section(page: Page, title: string) {
	return page
		.locator('section')
		.filter({ has: page.getByRole('heading', { name: title, exact: true }) })
		.last();
}

async function captureMobile(page: Page, width: number, path: string) {
	if (width === 320 || width === 390) {
		await page.screenshot({ path, fullPage: true });
	}
}

async function boundsOf(locator: Locator) {
	const bounds = await locator.boundingBox();
	if (!bounds) throw new Error('Section is not visible');
	return bounds;
}

async function inspectLayout(page: Page, side: 'STORE' | 'CONSIGNOR') {
	await expect(page.getByRole('link', { name: /Roti Sobek.*Lihat detail/u })).toBeVisible();
	const opportunities = await boundsOf(section(page, 'Peluang'));
	const products = await boundsOf(section(page, 'Produk terlaris'));
	expect(opportunities.width).toBe(products.width);
	expect(opportunities.y).toBeLessThan(products.y);
	const suggestions = section(page, 'Saran').locator('li');
	const heights = await suggestions.evaluateAll((items) =>
		items.map((item) => item.getBoundingClientRect().height)
	);
	expect(new Set(heights).size).toBe(1);
	if (side === 'STORE') {
		const quick = await boundsOf(section(page, 'Hitung cepat'));
		const tasks = await boundsOf(section(page, 'Perlu tindakan'));
		expect(quick.height).toBe(tasks.height);
	}
	await expect(section(page, 'Produk terlaris').locator('img')).toBeVisible();
	await expect
		.poll(() =>
			section(page, 'Produk terlaris')
				.locator('img')
				.evaluate((image: HTMLImageElement) => image.naturalWidth)
		)
		.toBeGreaterThan(0);
}

for (const side of ['STORE', 'CONSIGNOR'] as const) {
	test(`${side} dashboard aligns cards and full-width sections in both themes`, async ({
		page
	}, testInfo) => {
		await mockDashboard(page, side);
		await page.goto(side === 'STORE' ? '/app/toko' : '/app/penitip');
		await expect(page.getByRole('link', { name: /Roti Sobek.*Lihat detail/u })).toBeVisible();
		for (const theme of ['light', 'dark']) {
			await page.getByLabel('Tema tampilan').selectOption(theme);
			for (const width of [320, 390, 768, 1024, 1440]) {
				await page.setViewportSize({ width, height: 900 });
				await expect
					.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
					.toBeLessThanOrEqual(width);
				await captureMobile(
					page,
					width,
					testInfo.outputPath(`${side}-${theme}-${String(width)}.png`)
				);
			}
			await inspectLayout(page, side);
			await expect
				.poll(() =>
					page.evaluate(
						() =>
							document
								.getAnimations()
								.filter(
									(animation) =>
										animation instanceof CSSTransition && animation.playState === 'running'
								).length
					)
				)
				.toBe(0);
			await page.evaluate(axeSource);
			const violations = await page.evaluate(async () =>
				(await window.axe.run({ resultTypes: ['violations'] })).violations.map((violation) => ({
					id: violation.id,
					nodes: violation.nodes.map((node) => ({
						target: node.target,
						summary: node.failureSummary
					}))
				}))
			);
			expect(violations).toEqual([]);
			await page.screenshot({ path: testInfo.outputPath(`${side}-${theme}.png`), fullPage: true });
		}
		await page.getByRole('link', { name: /Roti Sobek.*Lihat detail/u }).click();
		await expect(page).toHaveURL(/\/produk\/roti-anton\/ROT-1$/u);
	});
}

test('logout from a store and login as a consignor redirects to the current dashboard', async ({
	page
}) => {
	await mockDashboard(page, 'STORE');
	await page.goto('/app/toko');
	await page.getByText('Akun', { exact: true }).click();
	await page.getByRole('button', { name: 'Keluar', exact: true }).click();
	await expect(page).toHaveURL(/\/masuk\?lanjut=/u);
	await page.getByLabel('Email').fill('penitip@contoh.id');
	await page.getByLabel('Kata sandi', { exact: true }).fill('Kp7#mQz2vX');
	await page.getByRole('button', { name: 'Masuk', exact: true }).click();
	await expect(page).toHaveURL(/\/app\/penitip$/u);
	await expect(page.getByText('Unit di semua toko')).toBeVisible();
	await expect(page.getByText('Mitra aktif', { exact: true })).toHaveCount(0);
	await page.goto('/app/toko');
	await expect(page).toHaveURL(/\/app\/penitip$/u);
	await page.goto('/app/%74oko');
	await expect(page).toHaveURL(/\/app\/penitip$/u);
	await expect(page.getByText('Mitra aktif', { exact: true })).toHaveCount(0);
});

test('a store cannot open the consignor dashboard', async ({ page }) => {
	await mockDashboard(page, 'STORE');
	await page.goto('/app/penitip');
	await expect(page).toHaveURL(/\/app\/toko$/u);
	await expect(page.getByText('Mitra aktif', { exact: true })).toBeVisible();
	await expect(page.getByText('Unit di semua toko')).toHaveCount(0);
});

test('dashboard uses a page skeleton until data arrives', async ({ page }, testInfo) => {
	await mockDashboard(page, 'STORE');
	let release: () => void = () => undefined;
	const pending = new Promise<void>((resolve) => {
		release = resolve;
	});
	await page.route('**/api/v1/dashboard', async (route) => {
		await pending;
		await route.fulfill({ json: { code: 'OK', data: dashboardBoard, meta: {} } });
	});
	await page.goto('/app/toko');
	await expect(page.locator('[aria-busy="true"]')).toBeVisible();
	await expect(page.getByText('Mitra aktif', { exact: true })).toHaveCount(0);
	await page.screenshot({ path: testInfo.outputPath('loading.png'), fullPage: true });
	release();
	await expect(page.getByText('Mitra aktif', { exact: true })).toBeVisible();
	await expect(page.locator('[aria-busy="true"]')).toHaveCount(0);
});
