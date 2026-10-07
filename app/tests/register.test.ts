import { expect, test } from '@playwright/test';

test('a visitor creates an account and is pointed to sign in, keeping where they came from', async ({
	page
}, testInfo) => {
	const email = `e2e-${testInfo.project.name}-${String(Date.now())}@contoh.id`;

	await page.goto('/daftar?lanjut=/undangan/KODE123');
	await page.getByLabel('Nama lengkap').fill('Sari Wulandari');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill('Kp7#mQz2vX');
	await page.getByLabel('Nomor WhatsApp').fill('0812 3456 7890');
	await page.getByRole('checkbox').check();
	await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
		timeout: 15_000
	});
	await page.getByRole('button', { name: 'Daftar' }).click();

	await expect(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible();
	await expect(page.getByRole('main').getByRole('link', { name: 'Masuk' })).toHaveAttribute(
		'href',
		'/masuk?lanjut=%2Fundangan%2FKODE123'
	);
});
