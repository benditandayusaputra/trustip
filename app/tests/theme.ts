import type { Page } from '@playwright/test';

export async function chooseTheme(page: Page, theme: string): Promise<void> {
	const picker = page.getByLabel('Tema tampilan').filter({ visible: true });
	if ((await picker.count()) === 0) {
		await page.getByLabel('Buka menu navigasi').click();
	}
	await picker.selectOption(theme);
	await page.keyboard.press('Escape');
}
