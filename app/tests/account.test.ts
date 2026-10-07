import { expect, test, type Page } from '@playwright/test';

const password = 'Kp7#mQz2vX';
const captchaAnswer = 'input[name="cf-turnstile-response"]';

async function waitForCaptcha(page: Page): Promise<void> {
	await expect(page.locator(captchaAnswer)).not.toHaveValue('', { timeout: 15_000 });
}

async function register(page: Page, email: string): Promise<void> {
	await page.goto('/daftar?lanjut=/undangan/KODE123');
	await page.getByLabel('Nama lengkap').fill('Sari Wulandari');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill(password);
	await page.getByLabel('Nomor WhatsApp').fill('0812 3456 7890');
	await page.getByRole('checkbox').check();
	await waitForCaptcha(page);
	await page.getByRole('button', { name: 'Daftar' }).click();
}

async function signIn(page: Page, email: string): Promise<void> {
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill(password);
	await waitForCaptcha(page);
	await page.getByRole('button', { name: 'Masuk' }).click();
}

async function expectTokensHiddenFromScripts(page: Page): Promise<void> {
	const cookies = await page.context().cookies();
	const session = cookies.filter(
		(cookie) => cookie.name.endsWith('trustip_at') || cookie.name.endsWith('trustip_rt')
	);
	const visibleToScripts = await page.evaluate(() =>
		[
			document.cookie,
			JSON.stringify(Object.entries(localStorage)),
			JSON.stringify(Object.entries(sessionStorage))
		].join('|')
	);
	expect(session).toHaveLength(2);
	session.forEach((cookie) => {
		expect(cookie.httpOnly).toBe(true);
		expect(cookie.sameSite).toBe('Lax');
		expect(visibleToScripts).not.toContain(cookie.value);
	});
}

test('a visitor registers, signs in, and returns to where they came from', async ({
	page
}, testInfo) => {
	const email = `e2e-${testInfo.project.name}-${String(Date.now())}@contoh.id`;

	await register(page, email);
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fundangan%2FKODE123$/u);
	await expect(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible();
	await signIn(page, email);

	await expect(page).toHaveURL(/\/undangan\/KODE123$/u);
	await expectTokensHiddenFromScripts(page);
});
