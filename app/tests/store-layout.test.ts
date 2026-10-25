import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { expect, test, type Page } from '@playwright/test';
import { networkTimeoutMs } from './session';

const axeSource = readFileSync(createRequire(import.meta.url).resolve('axe-core'), 'utf8');
const themes = ['light', 'dark'];
const widths = [320, 768, 1024, 1440];

async function accessible(page: Page): Promise<void> {
	await page.evaluate(axeSource);
	const violations = await page.evaluate(async () => {
		const results = await window.axe.run({ resultTypes: ['violations'] });
		return results.violations.map((item) => ({
			rule: item.id,
			targets: item.nodes.map((node) => node.target)
		}));
	});
	expect(violations).toEqual([]);
}

async function checkWidths(page: Page): Promise<void> {
	for (const width of widths) {
		await page.setViewportSize({ width, height: 900 });
		await expect
			.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
			.toBeLessThanOrEqual(width);
	}
}

async function chooseStoreFromMap(page: Page): Promise<void> {
	const markers = page.getByRole('button', { name: /^Lihat detail /u });
	await expect(markers.first()).toBeVisible({ timeout: networkTimeoutMs });
	await markers.first().dispatchEvent('click');
	await expect(markers.first()).toHaveAttribute('aria-pressed', 'true');
	const details = page.getByRole('region', { name: 'Detail toko terpilih' });
	await expect(details.getByRole('heading')).toBeVisible();
	const selected = await markers.first().getAttribute('title');
	await expect(details.getByRole('heading')).toHaveText(
		selected?.replace('Lihat detail ', '') ?? ''
	);
	const next = (await markers.count()) > 1 ? markers.nth(1) : markers.first();
	await next.focus();
	await next.press('Space');
	await expect(next).toHaveAttribute('aria-pressed', 'true');
	await expect(next).toBeFocused();
	await next.press('Enter');
	await expect(next).toHaveAttribute('aria-pressed', 'true');
	await page
		.locator('details')
		.filter({ has: page.locator('summary', { hasText: 'Pilih toko pada peta' }) })
		.locator('summary')
		.click();
	const choices = page.getByRole('list', { name: 'Pilih toko pada peta' }).getByRole('button');
	if ((await choices.count()) > 1) {
		const name = await choices.nth(1).innerText();
		await choices.nth(1).click();
		await expect(details.getByRole('heading')).toHaveText(name.trim());
	}
}

test('store directory keeps filters optional and map markers open accessible details', async ({
	page
}) => {
	test.setTimeout(120_000);
	await page.goto('/cari/toko');
	await expect(page.getByRole('heading', { level: 2 }).first()).toBeVisible({
		timeout: networkTimeoutMs
	});
	await expect(page.getByLabel('Jenis toko')).toHaveCount(0);
	await expect(page.getByLabel('Cari nama atau deskripsi toko')).toBeVisible();
	const filters = page.getByRole('button', { name: /^Filter/u });
	await filters.focus();
	await filters.press('Enter');
	await expect(filters).toHaveAttribute('aria-expanded', 'true');
	await expect(page.getByLabel('Jenis toko')).toBeVisible();
	await checkWidths(page);
	await filters.press('Space');
	await expect(filters).toHaveAttribute('aria-expanded', 'false');
	await page.getByRole('button', { name: 'Peta', exact: true }).click();
	await chooseStoreFromMap(page);
	for (const theme of themes) {
		await page.evaluate((value) => {
			document.documentElement.dataset.appearance = value;
		}, theme);
		await checkWidths(page);
		await accessible(page);
	}
	await page.getByRole('button', { name: 'Daftar', exact: true }).click();
	await expect(page.getByRole('application', { name: 'Peta lokasi publik toko' })).toHaveCount(0);
	await expect(page.getByRole('heading', { level: 2 }).first()).toBeVisible();
});
