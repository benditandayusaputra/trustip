import { expect, test } from '@playwright/test';

test('pages carry the security headers and a hash based policy', async ({ page }) => {
	const response = await page.goto('/');
	const headers = response?.headers() ?? {};

	expect(headers['x-frame-options']).toBe('DENY');
	expect(headers['strict-transport-security']).toBe('max-age=63072000; includeSubDomains');
	expect(headers['content-security-policy']).toMatch(
		/script-src 'self' https:\/\/challenges\.cloudflare\.com 'sha256-/u
	);
	expect(headers['content-security-policy']).not.toMatch(/script-src[^;]*unsafe/u);
	expect(headers['content-security-policy']).toMatch(/style-src 'self'[^;]*;/u);
});

test('loading public pages raises no policy violations', async ({ page }) => {
	const violations: string[] = [];
	page.on('console', (message) => {
		if (message.text().includes('Content Security Policy')) {
			violations.push(message.text());
		}
	});

	await page.goto('/');
	await page.goto('/daftar');
	await expect(page.locator('input[name="cf-turnstile-response"]')).toHaveCount(1, {
		timeout: 15_000
	});
	await page.goto('/privasi');
	await page.goto('/ketentuan');
	await page.goto('/alamat-yang-tidak-ada');

	expect(violations).toEqual([]);
});
