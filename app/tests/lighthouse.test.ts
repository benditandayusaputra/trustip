import { chromium, expect, test, type BrowserContext } from '@playwright/test';
import lighthouse from 'lighthouse';
import { listedStore } from './catalog';

const appOrigin = 'http://localhost:5173';
const debuggingPort = 9333;
const minimumScore = 0.9;
const largestPaintBudgetMs = 2500;
const layoutShiftBudget = 0.1;
const flowTimeoutMs = 600_000;

interface Audit {
	path: string;
	score: number;
	largestPaintMs: number;
	layoutShift: number;
	serverMs: number;
}

type Report = NonNullable<Awaited<ReturnType<typeof lighthouse>>>['lhr'];

function numberOf(report: Report, id: string): number {
	return report.audits[id]?.numericValue ?? Infinity;
}

async function cookieHeader(context: BrowserContext): Promise<string> {
	const cookies = await context.cookies(appOrigin);
	return cookies.map((cookie) => `${cookie.name}=${cookie.value}`).join('; ');
}

async function audit(path: string, cookie: string): Promise<Audit> {
	const result = await lighthouse(`${appOrigin}${path}`, {
		port: debuggingPort,
		output: 'json',
		logLevel: 'error',
		onlyCategories: ['performance'],
		throttlingMethod: 'devtools',
		extraHeaders: { Cookie: cookie }
	});
	if (!result) {
		throw new Error(`Lighthouse returned nothing for ${path}`);
	}
	const report = result.lhr;
	await test.info().attach(`lighthouse-${path.replaceAll('/', '-')}`, {
		body: JSON.stringify(report, (key, value: unknown) =>
			key === 'extraHeaders' ? undefined : value
		),
		contentType: 'application/json'
	});
	return {
		path,
		score: report.categories.performance?.score ?? 0,
		largestPaintMs: numberOf(report, 'largest-contentful-paint'),
		layoutShift: numberOf(report, 'cumulative-layout-shift'),
		serverMs: numberOf(report, 'server-response-time')
	};
}

test.describe.configure({ timeout: flowTimeoutMs });

test(
	'NFR-04 public pages meet the Lighthouse budget and the dashboard does not shift',
	{ tag: '@lighthouse' },
	async ({ browser }) => {
		const context = await browser.newContext({ baseURL: appOrigin });
		const store = await context.newPage();
		const storeName = `Toko Ukur ${Date.now().toString(36)}`;
		await listedStore(store, storeName);
		const cookie = await cookieHeader(context);
		const profile = `/t/${storeName.toLowerCase().replaceAll(' ', '-')}`;
		const auditor = await chromium.launch({
			args: [`--remote-debugging-port=${String(debuggingPort)}`]
		});

		const audits: Audit[] = [];
		for (const path of ['/', '/cari/toko', profile]) {
			audits.push(await audit(path, ''));
		}
		const dashboard = await audit('/app/toko', cookie);
		await auditor.close();
		await test.info().attach('lighthouse', {
			body: JSON.stringify([...audits, dashboard]),
			contentType: 'application/json'
		});
		test.info().annotations.push({
			type: 'lighthouse',
			description: JSON.stringify([...audits, dashboard])
		});

		for (const result of audits) {
			expect.soft(result.score, result.path).toBeGreaterThanOrEqual(minimumScore);
			expect.soft(result.largestPaintMs, result.path).toBeLessThan(largestPaintBudgetMs);
			expect.soft(result.layoutShift, result.path).toBeLessThan(layoutShiftBudget);
		}
		expect(dashboard.layoutShift).toBeLessThan(layoutShiftBudget);
	}
);
