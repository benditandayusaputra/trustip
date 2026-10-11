import { expect, test } from '@playwright/test';
import { initialScriptBytes, publicPageBudgetBytes } from './budget';

for (const path of [
	'/',
	'/cari/toko',
	'/cari/penitip',
	'/kebutuhan',
	'/masuk',
	'/daftar',
	'/demo'
]) {
	test(`R-PERF-02 initial JavaScript of ${path} stays under 90 KB compressed`, async ({ page }) => {
		const bytes = await initialScriptBytes(page, path);
		test.info().annotations.push({ type: 'initial-js-bytes', description: String(bytes) });

		expect(bytes).toBeGreaterThan(0);
		expect(bytes).toBeLessThanOrEqual(publicPageBudgetBytes);
	});
}
