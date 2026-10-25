import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { expect, test, type Browser, type Page } from '@playwright/test';
import { networkTimeoutMs } from './session';

declare global {
	interface Window {
		axe: typeof import('axe-core');
	}
}

const appOrigin = 'http://localhost:5173';
const axeSource = readFileSync(createRequire(import.meta.url).resolve('axe-core'), 'utf8');
const failedRequests: string[] = [];
const blockingImpacts = new Set(['serious', 'critical']);
const flowTimeoutMs = 1_800_000;
const demoPassword = 'Coba-Demo-2026';
const listTimeoutMs = 30_000;
const rateWindowMs = 60_000;
const loadAttempts = 3;

interface Finding {
	path: string;
	rule: string;
	impact: string;
	targets: string;
}

async function settle(page: Page, path: string): Promise<void> {
	for (let attempt = 0; attempt < loadAttempts; attempt++) {
		const failuresBefore = failedRequests.length;
		const response = await page.goto(path, { waitUntil: 'load' });
		await expect(page.getByRole('status', { name: 'Memuat' })).toHaveCount(0, {
			timeout: networkTimeoutMs
		});
		const serverFailed = (response?.status() ?? 0) >= 500;
		if (failedRequests.length === failuresBefore && !serverFailed) {
			await expect(page, `title of ${path}`).toHaveTitle(/\S/u, { timeout: listTimeoutMs });
			return;
		}
		await page.waitForTimeout(rateWindowMs);
	}
	throw new Error(`${path} remained unavailable after ${String(loadAttempts)} attempts`);
}

async function responsivePage(page: Page, path: string): Promise<void> {
	for (const width of [320, 768, 1024, 1440]) {
		await page.setViewportSize({ width, height: 900 });
		await expect
			.poll(() => page.evaluate(() => document.documentElement.scrollWidth), {
				message: `${path} at ${String(width)}px`
			})
			.toBeLessThanOrEqual(width);
	}
}

async function findingsOn(page: Page, path: string): Promise<Finding[]> {
	await settle(page, path);
	await responsivePage(page, path);
	await page.addScriptTag({ content: axeSource });
	const violations = await page.evaluate(async () => {
		const results = await window.axe.run(document, {
			runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']
		});
		return results.violations.map((violation) => ({
			rule: violation.id,
			impact: violation.impact ?? 'minor',
			targets: violation.nodes.map((node) => node.target.join(' ')).join(', ')
		}));
	});
	return violations.map((violation) => ({ path, ...violation }));
}

async function openPage(browser: Browser, colorScheme: 'light' | 'dark' = 'light'): Promise<Page> {
	const context = await browser.newContext({ baseURL: appOrigin, bypassCSP: true, colorScheme });
	const page = await context.newPage();
	page.on('response', (response) => {
		if (response.status() === 429 || response.status() >= 500) {
			failedRequests.push(response.url());
		}
	});
	return page;
}

async function signInAs(page: Page, email: string): Promise<void> {
	await page.goto('/masuk');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Kata sandi').fill(demoPassword);
	await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
		timeout: networkTimeoutMs
	});
	await page.getByRole('button', { name: 'Masuk' }).click();
	await page.waitForURL(/\/app\//u, { timeout: networkTimeoutMs });
}

async function firstLink(
	page: Page,
	listPath: string | { path: string; tab: string },
	pattern: RegExp
): Promise<string> {
	const { path, tab } =
		typeof listPath === 'string' ? { path: listPath, tab: undefined } : listPath;
	await settle(page, path);
	if (tab) {
		await page.getByRole('tab', { name: tab, exact: true }).click();
	}
	const matching = async () => {
		const hrefs = await page
			.locator('a[href]')
			.evaluateAll((anchors) => anchors.map((anchor) => anchor.getAttribute('href') ?? ''));
		return hrefs.find((href) => pattern.test(href)) ?? '';
	};
	await expect
		.poll(matching, { timeout: listTimeoutMs, message: `${String(pattern)} on ${path}` })
		.not.toBe('');
	return matching();
}

async function audit(page: Page, paths: string[]): Promise<Finding[]> {
	const findings: Finding[] = [];
	for (const path of paths) {
		findings.push(...(await findingsOn(page, path)));
	}
	return findings;
}

async function visitorPages(page: Page): Promise<string[]> {
	const store = await firstLink(page, '/cari/toko', /^\/t\//u);
	const need = await firstLink(page, '/kebutuhan', /^\/kebutuhan\/[0-9a-f-]{36}$/u);
	const product = await firstLink(page, '/u/roti-mas-bendi', /^\/p\//u);
	return [
		'/',
		'/masuk',
		'/daftar',
		'/demo',
		'/privasi',
		'/ketentuan',
		'/panduan',
		'/faq',
		'/luring',
		'/cari/toko',
		'/cari/penitip',
		'/kebutuhan',
		store,
		need,
		'/u/roti-mas-bendi',
		product,
		'/verifikasi/AAAA-AAAA-AAAA-AAAA',
		'/undangan/TIDAKADA'
	];
}

async function storePages(page: Page): Promise<string[]> {
	const id = /[0-9a-f-]{36}$/u.source;
	const partner = await firstLink(page, '/app/mitra', new RegExp(`^/app/mitra/${id}`, 'u'));
	const partnershipId = partner.split('/').pop() ?? '';
	return [
		'/app/toko',
		'/cari/penitip',
		'/kebutuhan',
		'/app/mitra',
		partner,
		`${partner}/kesepakatan`,
		await firstLink(page, `${partner}/kesepakatan`, /\/kesepakatan\/\d+$/u),
		await firstLink(page, { path: partner, tab: 'Stok' }, /\/stok\/[0-9a-f-]{36}$/u),
		'/app/titipan/baru',
		`/app/titipan/baru?mitra=${partnershipId}`,
		'/app/hitung/baru',
		`/app/hitung/baru?mitra=${partnershipId}`,
		'/app/ajuan',
		await firstLink(page, '/app/ajuan', new RegExp(`^/app/ajuan/${id}`, 'u')),
		'/app/titipan',
		await firstLink(page, '/app/titipan', new RegExp(`^/app/titipan/${id}`, 'u')),
		'/app/hitung',
		await firstLink(page, '/app/hitung', new RegExp(`^/app/hitung/${id}`, 'u')),
		'/app/setoran',
		await firstLink(page, '/app/setoran', new RegExp(`^/app/setoran/${id}`, 'u')),
		'/app/stok',
		'/app/kebutuhan',
		await firstLink(page, '/app/kebutuhan', new RegExp(`^/app/kebutuhan/${id}`, 'u')),
		'/app/kebutuhan/baru',
		'/app/laporan',
		'/app/pemberitahuan',
		'/app/pengaturan',
		'/app/antrean'
	];
}

async function consignorPages(page: Page): Promise<string[]> {
	const id = /[0-9a-f-]{36}$/u.source;
	const partner = await firstLink(page, '/app/mitra', new RegExp(`^/app/mitra/${id}`, 'u'));
	return [
		'/app/penitip',
		'/cari/toko',
		'/kebutuhan',
		'/app/produk',
		await firstLink(page, '/app/produk', new RegExp(`^/app/produk/${id}`, 'u')),
		'/app/produk/baru',
		'/app/ajuan/baru',
		`${partner}/perubahan`
	];
}

test.describe.configure({ timeout: flowTimeoutMs });

async function auditAs(
	browser: Browser,
	session: { email: string; theme: 'light' | 'dark' },
	pagesOf: (page: Page) => Promise<string[]>
): Promise<Finding[]> {
	const page = await openPage(browser, session.theme);
	await signInAs(page, session.email);
	const findings = await audit(page, await pagesOf(page));
	await page.context().close();
	return findings;
}

const themes = ['light', 'dark'] as const;

for (const theme of themes) {
	test(
		`NFR-11 every page has no serious accessibility violation in ${theme} theme`,
		{ tag: '@accessibility' },
		async ({ browser }) => {
			const visitor = await openPage(browser, theme);
			const findings = [
				...(await audit(visitor, await visitorPages(visitor))),
				...(await auditAs(browser, { email: 'toko@demo.trustip.id', theme }, storePages)),
				...(await auditAs(browser, { email: 'penitip@demo.trustip.id', theme }, consignorPages))
			];
			await visitor.context().close();
			test.info().annotations.push({ type: 'axe', description: JSON.stringify(findings) });
			expect(findings.filter((finding) => blockingImpacts.has(finding.impact))).toEqual([]);
		}
	);
}

async function focusedText(page: Page): Promise<string> {
	return page.evaluate(() => {
		const focused = document.activeElement;
		const visible = focused?.matches(':focus-visible') ?? false;
		const name = focused?.getAttribute('aria-label') ?? focused?.textContent ?? '';
		return `${visible ? 'visible' : 'hidden'}|${name.trim()}`;
	});
}

async function tabTo(page: Page, text: string | RegExp, key = 'Tab'): Promise<void> {
	const maximumStops = 40;
	for (let stop = 0; stop < maximumStops; stop++) {
		await page.keyboard.press(key);
		const focused = await focusedText(page);
		const name = focused.slice(focused.indexOf('|') + 1);
		if (typeof text === 'string' ? name === text : text.test(name)) {
			expect(focused.startsWith('visible')).toBe(true);
			return;
		}
	}
	throw new Error(`keyboard never reached ${String(text)}`);
}

test(
	'NFR-11 a store signs in and opens a quick count with the keyboard alone',
	{ tag: '@accessibility' },
	async ({ browser }) => {
		const page = await openPage(browser);
		await page.goto('/demo');
		await expect(page.locator('input[name="cf-turnstile-response"]')).not.toHaveValue('', {
			timeout: networkTimeoutMs
		});

		await tabTo(page, 'Lewati ke konten');
		await page.keyboard.press('Enter');
		await expect(page.locator('main#konten')).toBeFocused();
		await tabTo(page, 'Isi sebagai Zara (toko)');
		await page.keyboard.press('Enter');
		await expect(page.getByLabel('Email')).toHaveValue('toko@demo.trustip.id');
		await expect(page.getByLabel('Kata sandi')).toHaveValue(demoPassword);
		await tabTo(page, 'Masuk');
		await page.keyboard.press('Enter');
		await page.waitForURL(/\/app\/toko$/u, { timeout: networkTimeoutMs });
		await expect(page.getByRole('heading', { name: 'Hitung cepat' })).toBeVisible({
			timeout: networkTimeoutMs
		});
		await tabTo(page, /^Hitung .+/u);
		await page.keyboard.press('Enter');

		await page.waitForURL(/\/app\/hitung\/baru/u, { timeout: networkTimeoutMs });
	}
);
