import { expect, test, type Page } from '@playwright/test';

const password = 'Kp7#mQz2vX';
const captchaAnswer = 'input[name="cf-turnstile-response"]';
const networkTimeoutMs = 15_000;

async function waitForCaptcha(page: Page): Promise<void> {
	await expect(page.locator(captchaAnswer)).not.toHaveValue('', { timeout: networkTimeoutMs });
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

interface Business {
	choice: RegExp;
	name: string;
	dashboard: string;
}

const businessByProject: Record<string, Business> = {
	mobile: { choice: /Penitip/u, name: 'Kripik Uji', dashboard: '/app/penitip' },
	desktop: { choice: /Toko/u, name: 'Warung Uji', dashboard: '/app/toko' }
};

async function onboard(page: Page, business: Business): Promise<void> {
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

test('FR-AUTH-04 FR-AUTH-06 a consignor on a phone and a store on a desktop each reach an empty dashboard', async ({
	page
}, testInfo) => {
	const email = `e2e-${testInfo.project.name}-${String(Date.now())}@contoh.id`;
	const business = businessByProject[testInfo.project.name];
	if (!business) {
		throw new Error(`no business for project ${testInfo.project.name}`);
	}
	const dashboard = new RegExp(`${business.dashboard}$`, 'u');
	const greeting = page.getByRole('heading', { name: 'Halo, Sari Wulandari' });

	await page.goto('/app');
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp$/u);
	await page.getByRole('main').getByRole('link', { name: 'Daftar' }).click();
	await register(page, email);
	await expect(page).toHaveURL(/\/masuk\?lanjut=%2Fapp$/u);
	await expect(page.getByText('Akun dibuat. Silakan masuk.')).toBeVisible();
	await signIn(page, email);
	await expect(page).toHaveURL(/\/onboarding\?lanjut=%2Fapp$/u);
	await onboard(page, business);
	await expect(page).toHaveURL(dashboard, { timeout: networkTimeoutMs });
	await expect(greeting).toBeVisible();
	await expectTokensHiddenFromScripts(page);

	await page.getByRole('link', { name: 'Lengkapi profil usaha' }).click();
	await expect(page.getByLabel('Nama usaha')).toHaveValue(business.name);
	await page.getByLabel('Nama usaha').fill(`${business.name} Jaya`);
	await page.getByRole('button', { name: 'Simpan profil', exact: true }).click();
	await expect(page.getByText('Profil usaha tersimpan.')).toBeVisible();
	await page.getByRole('link', { name: 'Pemberitahuan', exact: true }).click();
	await expect(page.getByText(/Belum ada pemberitahuan/u)).toBeVisible();
	await page.goto(business.dashboard);

	await page.context().clearCookies({ name: 'trustip_at' });
	await page.reload();
	await expect(greeting).toBeVisible();
	await expectTokensHiddenFromScripts(page);

	await page.getByRole('button', { name: 'Keluar', exact: true }).click();
	await expect(page).toHaveURL(
		new RegExp(`/masuk\\?lanjut=${encodeURIComponent(business.dashboard)}$`, 'u')
	);
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
