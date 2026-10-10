import { gzipSync } from 'node:zlib';
import type { Page, Response } from '@playwright/test';

const kilobyte = 1024;
export const publicPageBudgetBytes = 90 * kilobyte;
export const recordingScreenBudgetBytes = 120 * kilobyte;

export async function initialScriptBytes(page: Page, path: string): Promise<number> {
	const scripts = new Set<string>();
	const collect = (response: Response) => {
		if (response.request().resourceType() === 'script' && response.ok()) {
			scripts.add(response.url());
		}
	};
	page.on('response', collect);
	await page.goto(path, { waitUntil: 'load' });
	page.off('response', collect);
	const sizes = await Promise.all(
		[...scripts].map(async (url) => gzipSync(await (await page.request.get(url)).body()).length)
	);
	return sizes.reduce((sum, size) => sum + size, 0);
}
