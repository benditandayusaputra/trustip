import { expect, test, type Page } from '@playwright/test';
import { mockDashboard } from './dashboard-fixture';

function deferred() {
	let release: () => void = () => undefined;
	const promise = new Promise<void>((resolve) => {
		release = resolve;
	});
	return { promise, release };
}

test('loading honors both themes and reduced motion on a narrow screen', async ({ page }, info) => {
	await page.setViewportSize({ width: 320, height: 800 });
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await mockDashboard(page, 'STORE');
	await failData(page);
	const pending = deferred();
	await page.route('**/api/v1/me', async (route) => {
		await pending.promise;
		await route.fallback();
	});
	await page.goto('/app/laporan');
	for (const theme of ['light', 'dark']) {
		await page.getByLabel('Tema tampilan').selectOption(theme);
		await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth)).toBe(320);
		const animations = await page
			.locator('[data-loading-shape] > div')
			.evaluateAll((elements) =>
				elements.map((element) => getComputedStyle(element).animationName)
			);
		expect(animations.length).toBeGreaterThan(0);
		expect(new Set(animations)).toEqual(new Set(['none']));
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
		await page.screenshot({ path: info.outputPath(`report-${theme}-320.png`), fullPage: true });
	}
	pending.release();
});

async function failData(page: Page) {
	await page.route('**/api/v1/**', async (route) => {
		const path = new URL(route.request().url()).pathname;
		if (path === '/api/v1/me' || path.includes('/notifications/')) {
			await route.fallback();
			return;
		}
		await route.fulfill({
			status: 503,
			json: { code: 'UNAVAILABLE', message: 'Gagal memuat data uji.', meta: {} }
		});
	});
}

const appPages = [
	['/app/produk', 'catalog'],
	['/app/mitra', 'list'],
	['/app/stok', 'list'],
	['/app/hitung/sample', 'detail'],
	['/app/titipan/baru', 'form'],
	['/app/laporan', 'report'],
	['/app/pengaturan', 'form'],
	['/app/profil', 'form'],
	['/app/pemberitahuan', 'list'],
	['/app/antrean', 'list']
] as const;

for (const [path, kind] of appPages) {
	test(`${path} shows its page skeleton while the account is loading`, async ({ page }, info) => {
		await mockDashboard(page, 'CONSIGNOR');
		await failData(page);
		const pending = deferred();
		await page.route('**/api/v1/me', async (route) => {
			await pending.promise;
			await route.fallback();
		});
		await page.goto(path);
		const skeleton = page.locator(`[data-loading-skeleton="${kind}"]`);
		await expect(skeleton).toBeVisible();
		await expect(skeleton.locator('button, a, input, select')).toHaveCount(0);
		await expect(skeleton).toHaveAttribute('aria-busy', 'true');
		await page.screenshot({ path: info.outputPath(`${kind}-account-loading.png`), fullPage: true });
		pending.release();
		await expect(page.locator('[data-loading-skeleton]')).toHaveCount(0);
	});
}

const dataPages = [
	['/app/produk', 'catalog'],
	['/app/laporan', 'report'],
	['/app/titipan/baru?mitra=sample', 'form']
] as const;

for (const [path, kind] of dataPages) {
	test(`${path} keeps a content skeleton until its API request finishes`, async ({ page }) => {
		await mockDashboard(page, 'STORE');
		const pending = deferred();
		await page.route('**/api/v1/**', async (route) => {
			const pathname = new URL(route.request().url()).pathname;
			if (pathname === '/api/v1/me' || pathname.includes('/notifications/')) {
				await route.fallback();
				return;
			}
			await pending.promise;
			await route.fulfill({
				status: 503,
				json: { code: 'UNAVAILABLE', message: 'Gagal memuat data uji.', meta: {} }
			});
		});
		await page.goto(path);
		await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
		await expect(page.locator(`[data-loading-skeleton="${kind}"]`)).toBeVisible();
		pending.release();
		await expect(page.locator('[data-loading-skeleton]')).toHaveCount(0);
		await expect(page.getByText('Gagal memuat data uji.', { exact: true })).toBeVisible();
	});
}

test('new product keeps typed values while its category field is loading', async ({ page }) => {
	await mockDashboard(page, 'CONSIGNOR');
	const pending = deferred();
	await page.route('**/api/v1/categories', async (route) => {
		await pending.promise;
		await route.fulfill({ json: { code: 'OK', data: [], meta: {} } });
	});
	await page.goto('/app/produk/baru');
	await page.getByLabel('Nama produk').fill('Roti baru');
	await expect(page.locator('form [aria-busy="true"]')).toBeVisible();
	pending.release();
	await expect(page.getByLabel('Kategori', { exact: true })).toBeVisible();
	await expect(page.getByLabel('Nama produk')).toHaveValue('Roti baru');
	await expect(page.locator('form [aria-busy="true"]')).toHaveCount(0);
});

const publicPages = [
	['/t/skeleton-preview', 'profile'],
	['/u/skeleton-preview', 'profile'],
	['/p/skeleton-preview/SKU', 'product']
] as const;

for (const [path, kind] of publicPages) {
	test(`navigation to ${path} shows a skeleton while server data is loading`, async ({
		page
	}, info) => {
		await mockDashboard(page, 'STORE');
		await page.goto('/masuk');
		await page.waitForLoadState('networkidle');
		await page.getByLabel('Email').fill('tersimpan@contoh.id');
		const pending = deferred();
		await page.route(`**${path}/__data.json*`, async (route) => {
			await pending.promise;
			await route.continue();
		});
		await page.evaluate((href) => {
			const link = document.createElement('a');
			link.href = href;
			link.textContent = 'Buka halaman tujuan';
			document.querySelector('main')?.append(link);
		}, path);
		await page.getByRole('link', { name: 'Buka halaman tujuan' }).click();
		await expect(
			page.locator(`[data-navigation-loading] [data-loading-skeleton="${kind}"]`)
		).toBeVisible();
		await expect(page.getByLabel('Email')).not.toBeVisible();
		await page.screenshot({
			path: info.outputPath(`${kind}-navigation-loading.png`),
			fullPage: true
		});
		pending.release();
		await expect(page).toHaveURL(path);
		await expect(page.locator('[data-navigation-loading]')).toHaveCount(0);
	});
}
