import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { expect, test, type Page, type Locator } from '@playwright/test';
import { dashboardBoard, mockDashboard } from './dashboard-fixture';
import { chooseTheme } from './theme';

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
	await expect(page.getByRole('heading', { name: 'Peluang', exact: true })).toHaveCount(0);
	const reminders = await boundsOf(section(page, 'Pengingat'));
	const products = await boundsOf(section(page, 'Produk terlaris'));
	expect(reminders.width).toBe(products.width);
	expect(reminders.y).toBeLessThan(products.y);
	const suggestions = section(page, 'Pengingat').locator('li');
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
		for (const theme of ['light', 'dark'] as const) {
			await chooseTheme(page, theme);
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
		await expect(page).toHaveURL(/\/p\/roti-anton\/ROT-1$/u);
	});
}

test('logout from a store and login as a consignor redirects to the current dashboard', async ({
	page
}) => {
	await mockDashboard(page, 'STORE');
	await page.goto('/app/toko');
	await page.getByLabel('Buka menu profil').click();
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

test('the bell is marked current on the notifications page', async ({ page }) => {
	await mockDashboard(page, 'STORE');
	await page.goto('/app/toko');
	const bell = page.locator('a[aria-label^="Pemberitahuan"]');
	await expect(bell).toBeVisible();
	await expect(bell).not.toHaveAttribute('aria-current', 'page');

	await bell.click();

	await expect(page).toHaveURL(/\/app\/pemberitahuan$/u);
	await expect(bell).toHaveAttribute('aria-current', 'page');
	await expect(bell).toHaveClass(/bg-primary/u);
});

test('a settlement opened from a menu page keeps that menu in the breadcrumb and sidebar', async ({
	page
}, testInfo) => {
	test.skip(testInfo.project.name !== 'desktop', 'The sidebar shows on wide screens');
	await mockDashboard(page, 'STORE');
	await page.route('**/api/v1/settlements/**', (route) =>
		route.fulfill({
			status: 404,
			json: { code: 'NOT_FOUND', message: 'Tidak ditemukan.', meta: { request_id: 'test' } }
		})
	);
	const sidebar = page.getByRole('complementary').getByRole('navigation');
	const back = page.getByRole('banner').getByRole('link', { name: 'Kembali' });
	const settlement = '/app/setoran/0192f0c4-0000-7000-8000-000000000301';

	await page.goto(settlement);
	await expect(back).toHaveAttribute('href', '/app/laporan');
	await expect(sidebar.getByRole('link', { name: 'Laporan' })).toHaveAttribute(
		'aria-current',
		'page'
	);

	await page.goto('/app/mitra?status=ACTIVE');
	await expect(sidebar.getByRole('link', { name: 'Mitra' })).toHaveAttribute(
		'aria-current',
		'page'
	);
	await page.evaluate((href) => {
		const link = document.createElement('a');
		link.href = href;
		link.textContent = 'Buka setoran';
		document.querySelector('main')?.append(link);
	}, settlement);
	await page.evaluate(() => {
		Object.assign(window, { sameDocument: true });
	});
	await page.getByRole('link', { name: 'Buka setoran' }).click();
	await expect(page).toHaveURL(settlement);
	expect(await page.evaluate(() => 'sameDocument' in window)).toBe(true);

	await expect(page).toHaveURL(settlement);
	await expect(back).toHaveAttribute('href', '/app/mitra?status=ACTIVE');
	await expect(page.getByRole('banner').getByRole('link', { name: 'Mitra' })).toBeVisible();
	await expect(sidebar.getByRole('link', { name: 'Mitra' })).toHaveAttribute(
		'aria-current',
		'page'
	);
	await expect(sidebar.getByRole('link', { name: 'Laporan' })).not.toHaveAttribute(
		'aria-current',
		'page'
	);
});
