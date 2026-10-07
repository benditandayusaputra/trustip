import { expect, type Page } from '@playwright/test';

const password = 'Kp7#mQz2vX';
const captchaAnswer = 'input[name="cf-turnstile-response"]';
export const networkTimeoutMs = 15_000;

async function waitForCaptcha(page: Page): Promise<void> {
	await expect(page.locator(captchaAnswer)).not.toHaveValue('', { timeout: networkTimeoutMs });
}

export async function register(page: Page, email: string): Promise<void> {
	await page.getByLabel('Nama lengkap').fill('Sari Wulandari');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill(password);
	await page.getByLabel('Nomor WhatsApp').fill('0812 3456 7890');
	await page.getByRole('checkbox').check();
	await waitForCaptcha(page);
	await page.getByRole('button', { name: 'Daftar' }).click();
}

export async function registerThenSignIn(page: Page, email: string): Promise<void> {
	await register(page, email);
	await expect(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible({
		timeout: networkTimeoutMs
	});
	await signIn(page, email);
}

export async function signIn(page: Page, email: string): Promise<void> {
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill(password);
	await waitForCaptcha(page);
	await page.getByRole('button', { name: 'Masuk' }).click();
}

export interface Business {
	choice: RegExp;
	name: string;
	dashboard: string;
}

export async function onboard(page: Page, business: Business): Promise<void> {
	await page.getByRole('radio', { name: business.choice }).check();
	await page.getByRole('button', { name: 'Lanjut' }).click();
	await page.getByLabel('Nama usaha').fill(business.name);
	await page.getByLabel('Nomor WhatsApp usaha').fill('0812 3456 7890');
	await page.getByRole('button', { name: 'Lanjut' }).click();
	await page.getByLabel('Provinsi').selectOption({ label: 'Jawa Tengah' });
	await expect(page.getByRole('option', { name: 'Kota Semarang' })).toBeAttached();
	await page.getByLabel('Kabupaten atau kota').selectOption({ label: 'Kota Semarang' });
	await page.context().grantPermissions(['geolocation']);
	await page.context().setGeolocation({ latitude: -7.0546, longitude: 110.4386 });
	await page.getByRole('button', { name: 'Gunakan lokasi saya' }).click();
	await expect(page.getByText('Lokasi ditandai.')).toBeVisible();
	await page.getByRole('button', { name: 'Simpan dan mulai' }).click();
}

export function uniqueEmail(role: string): string {
	return `e2e-${role}-${String(Date.now())}@contoh.id`;
}
