import { expect, test, type Page } from '@playwright/test';

const password = 'Kp7#mQz2vX';
const captchaAnswer = 'input[name="cf-turnstile-response"]';

async function waitForCaptcha(page: Page): Promise<void> {
	await expect(page.locator(captchaAnswer)).not.toHaveValue('', { timeout: 15_000 });
}

async function register(page: Page, email: string): Promise<void> {
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

async function onboardStore(page: Page): Promise<void> {
	await page.getByRole('radio', { name: /Toko/u }).check();
	await page.getByRole('button', { name: 'Lanjut' }).click();
	await page.getByLabel('Nama usaha').fill('Warung Uji');
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

test('FR-AUTH-04 FR-AUTH-06 a visitor registers, signs in, onboards, and lands on the dashboard', async ({
	page
}, testInfo) => {
	const email = `e2e-${testInfo.project.name}-${String(Date.now())}@contoh.id`;
	const greeting = page.getByRole('heading', { name: 'Halo, Sari Wulandari' });

	await page.goto('/app');
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp$/u);
	await page.getByRole('main').getByRole('link', { name: 'Daftar' }).click();
	await register(page, email);
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp$/u);
	await expect(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible();
	await signIn(page, email);
	await expect(page).toHaveURL(/\/onboarding\?lanjut=%2Fapp$/u);
	await onboardStore(page);
	await expect(page).toHaveURL(/\/app\/toko$/u);
	await expect(greeting).toBeVisible();
	await expectTokensHiddenFromScripts(page);

	await page.getByRole('link', { name: 'Lengkapi profil usaha' }).click();
	await expect(page.getByLabel('Nama usaha')).toHaveValue('Warung Uji');
	await page.getByLabel('Nama usaha').fill('Warung Uji Jaya');
	await page.getByRole('button', { name: 'Simpan profil', exact: true }).click();
	await expect(page.getByText('Profil usaha tersimpan.')).toBeVisible();
	await page.getByRole('link', { name: 'Pemberitahuan', exact: true }).click();
	await expect(page.getByText(/Belum ada pemberitahuan/u)).toBeVisible();
	await page.goto('/app/toko');

	await page.context().clearCookies({ name: 'trustip_at' });
	await page.reload();
	await expect(greeting).toBeVisible();
	await expectTokensHiddenFromScripts(page);

	await page.getByRole('button', { name: 'Keluar', exact: true }).click();
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp%2Ftoko$/u);
	expect(await page.context().cookies()).toHaveLength(0);
	await page.goto('/app');
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp$/u);
});

test('R-SEC-21 a write sent from another site is rejected', async ({ request }) => {
	const response = await request.post('/api/v1/sessions/logout', {
		headers: { origin: 'https://evil.example', 'x-trustip-client': 'web' },
		data: {}
	});

	expect(response.status()).toBe(403);
	expect(((await response.json()) as { code: string }).code).toBe('CROSS_SITE_REJECTED');
});
