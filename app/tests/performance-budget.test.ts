import { gzipSync } from 'node:zlib';
import { expect, test } from '@playwright/test';

const publicPageBudgetBytes = 90 * 1024;

test('initial JavaScript of a public page stays under 90 KB compressed', async ({ page }) => {
	const scripts: Promise<number>[] = [];
	page.on('response', (response) => {
		if (response.request().resourceType() === 'script') {
			scripts.push(response.body().then((body) => gzipSync(body).length));
		}
	});

	await page.goto('/', { waitUntil: 'networkidle' });
	const sizes = await Promise.all(scripts);
	const total = sizes.reduce((sum, size) => sum + size, 0);

	expect(sizes.length).toBeGreaterThan(0);
	expect(total).toBeLessThanOrEqual(publicPageBudgetBytes);
});
