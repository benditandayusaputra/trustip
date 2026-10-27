import { expect, test } from '@playwright/test';

test('FR-PUB-01 NFR-17 R-PERF-08 home page explains the problem, how it works, and invites to sign up', async ({
	page
}) => {
	const response = await page.goto('/');

	expect(response?.headers()['cache-control']).toBe(
		'public, s-maxage=60, stale-while-revalidate=300'
	);
	await expect(page.getByRole('heading', { level: 1 })).toContainText('Titip jual lebih mudah.');
	await expect(page.getByRole('heading', { name: 'Masalah titip jual hari ini' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Cara kerja' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Daftar gratis' }).first()).toHaveAttribute(
		'href',
		'/daftar'
	);
	await expect(page.getByRole('heading', { name: 'Saya penitip produk' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Saya pemilik toko' })).toBeVisible();
	const workflow = page.getByRole('region', { name: 'Cara kerja', exact: true });
	await expect(workflow.getByRole('listitem')).toHaveCount(5);
	await expect(workflow).toContainText('Pembayaran dilakukan langsung antar mitra.');
	const explain = page.getByRole('link', { name: 'Lihat cara kerja', exact: true });
	await explain.focus();
	await explain.press('Enter');
	await expect(workflow).toBeFocused();
	await expect(page).toHaveURL(/\/#beranda-cara$/u);
	const footer = page.getByRole('contentinfo');
	await expect(footer.getByRole('link', { name: 'Jelajahi demo' })).toHaveAttribute(
		'href',
		'/demo'
	);
	await expect(footer.getByRole('link', { name: 'Kebijakan privasi' })).toHaveAttribute(
		'href',
		'/privasi'
	);
	const browse = footer.getByRole('link', { name: 'Cari produk dan penitip' });
	await browse.focus();
	await browse.press('Enter');
	await expect(page).toHaveURL(/\/cari\/penitip$/u);
});
