import { expect, test, type Page } from '@playwright/test';
import { chooseTheme } from './theme';

const storageKey = 'trustip-appearance';

async function expectAppearance(page: Page, theme: string): Promise<void> {
	await expect(page.locator('html')).toHaveAttribute('data-appearance', theme);
	const colors = await page.evaluate(() => {
		const root = document.documentElement;
		return {
			background: getComputedStyle(root).getPropertyValue('--color-bg').trim(),
			meta: document.querySelector('meta[name="theme-color"]')?.getAttribute('content')?.trim()
		};
	});
	expect(colors.meta).toBe(colors.background);
}

test('explicit theme persists across navigation, reload and browser tabs', async ({ page }) => {
	await page.goto('/');
	await chooseTheme(page, 'dark');
	await expectAppearance(page, 'dark');
	await page.getByRole('link', { name: 'Kebijakan privasi', exact: true }).click();
	await expectAppearance(page, 'dark');
	await page.reload();
	await expectAppearance(page, 'dark');
	const second = await page.context().newPage();
	await second.goto('/ketentuan');
	await expectAppearance(second, 'dark');
	await chooseTheme(second, 'light');
	await expectAppearance(second, 'light');
	await expectAppearance(page, 'light');
});

test('saved theme is applied before the first page body is parsed', async ({ page }) => {
	await page.addInitScript((key) => {
		localStorage.setItem(key, 'dark');
		const firstBody = new MutationObserver(() => {
			if (document.querySelector('body')) {
				document.documentElement.dataset.firstAppearance =
					document.documentElement.dataset.appearance;
				firstBody.disconnect();
			}
		});
		firstBody.observe(document, { childList: true, subtree: true });
	}, storageKey);
	await page.goto('/');
	await expect(page.locator('html')).toHaveAttribute('data-first-appearance', 'dark');
	await expectAppearance(page, 'dark');
});

test('theme controls work when browser storage is unavailable', async ({ page }) => {
	await page.addInitScript(() => {
		Object.defineProperty(window, 'localStorage', {
			get() {
				throw new DOMException('Storage unavailable', 'SecurityError');
			}
		});
	});
	await page.goto('/');
	await chooseTheme(page, 'dark');
	await expectAppearance(page, 'dark');
	await chooseTheme(page, 'light');
	await expectAppearance(page, 'light');
});

test.describe('a device that prefers dark', () => {
	test('still opens in light until the user chooses dark', async ({ page }) => {
		await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
		await page.goto('/');
		await expectAppearance(page, 'light');
		await chooseTheme(page, 'dark');
		await page.emulateMedia({ colorScheme: 'light' });
		await expectAppearance(page, 'dark');
	});
	test('an old system or unknown saved preference falls back to light', async ({ page }) => {
		for (const saved of ['system', 'unknown']) {
			await page.addInitScript(
				(stored) => {
					localStorage.setItem(stored.key, stored.value);
				},
				{ key: storageKey, value: saved }
			);
			await page.emulateMedia({ colorScheme: 'dark' });
			await page.goto('/');
			await expectAppearance(page, 'light');
		}
	});
});
