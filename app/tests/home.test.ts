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
});
