import { expect, test, type Page } from '@playwright/test';
import { networkTimeoutMs, onboard, register, signIn, type Business } from './session';

const businessByProject: Record<string, Business> = {
	mobile: { choice: /Penitip/u, name: 'Kripik Uji', dashboard: '/app/penitip' }
};

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

test('FR-AUTH-04 FR-AUTH-06 a new business reaches its dashboard and settings', async ({
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

	await page.goto('/app/profil');
	await expect(page.getByLabel('Nama usaha')).toHaveValue(business.name);
	await page.getByRole('button', { name: 'Edit profil' }).click();
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

	await page.getByRole('button', { name: 'Buka menu profil' }).click();
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
