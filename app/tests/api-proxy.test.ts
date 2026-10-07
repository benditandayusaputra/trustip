import { expect, test } from '@playwright/test';

test('the public footer shows the API version fetched through the proxy', async ({ page }) => {
	const metaResponse = page.waitForResponse((response) => response.url().endsWith('/api/v1/meta'));

	await page.goto('/');

	const response = await metaResponse;
	expect(new URL(response.url()).origin).toBe(new URL(page.url()).origin);
	expect(response.status()).toBe(200);
	await expect(page.getByTestId('api-version')).toHaveText(/^Versi API \S+$/u);
});

test('the proxy rejects nothing the browser could forge', async ({ request }) => {
	const response = await request.get('/api/v1/meta', {
		headers: { 'x-edge-signature': 'forged', 'x-edge-client-ip': '6.6.6.6' }
	});

	expect(response.status()).toBe(200);
	expect(response.headers()['x-request-id']).toBeTruthy();
});
