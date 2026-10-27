import type { Page } from '@playwright/test';

const labels = { light: 'Terang', dark: 'Gelap' } as const;

export async function chooseTheme(page: Page, theme: keyof typeof labels): Promise<void> {
	const trigger = page.locator('summary[aria-label="Mode tampilan"]').filter({ visible: true });
	const phoneMenu = page.getByLabel('Buka menu navigasi').filter({ visible: true });
	await trigger.or(phoneMenu).first().waitFor();
	await ((await trigger.count()) > 0 ? trigger.first() : phoneMenu.first()).click();
	await page
		.getByRole('group', { name: 'Pilihan mode tampilan' })
		.filter({ visible: true })
		.getByRole('button', { name: labels[theme] })
		.click();
	await page.keyboard.press('Escape');
}

export async function useSavedTheme(page: Page, theme: keyof typeof labels): Promise<void> {
	await page.addInitScript((chosen) => {
		localStorage.setItem('trustip-appearance', chosen);
	}, theme);
}
