import type { Page } from '@playwright/test';
import type { OrganizationType } from '../src/lib/api/organizations';

const partnershipId = '0192f0c4-0000-7000-8000-000000000201';
const productId = '0192f0c4-0000-7000-8000-000000000401';
const split = { gross: 36000, store_share: 4500, consignor_share: 31500 };
const sales = { sold_qty: 18, returned_qty: 2, received_qty: 20, returned_bps: 1000, split };
const tally = { count: 1, amount: 31500 };

export const dashboardBoard = {
	figures: { active_partners: 3, stock_units: 42, unpaid: tally, overdue: tally, due_soon: tally },
	tasks: {
		drops_to_confirm: 1,
		counts_to_confirm: 1,
		payments_to_review: 1,
		payments_rejected: 1,
		proposals_awaiting: 1,
		needs_expiring: 0
	},
	quick_count: ['Roti Pak Anton', 'Keripik Bu Ani', 'Kue Sari'].map((name) => ({
		partnership_id: partnershipId + name,
		counterpart_name: name,
		stock_units: 14
	}))
};
const insights = ['Roti Sobek', 'Kue rumahan dengan isian cokelat dan keju spesial', 'Keripik'].map(
	(name, index) => ({
		kind: 'NEAR_EXPIRY',
		level: 'HIGH',
		action: 'COUNT',
		units: 12,
		days: 2,
		amount: 0,
		score: 0,
		subject: {
			partnership_id: partnershipId,
			counterpart_name: 'Warung Bu Tini',
			product_id: productId,
			product_name: name,
			batch_id: null,
			batch_code: String(index),
			settlement_id: null,
			settlement_number: '',
			organization_id: null,
			organization_name: ''
		}
	})
);

function account(type: OrganizationType) {
	return {
		id: 'u-' + type,
		full_name: type === 'STORE' ? 'Pemilik Toko' : 'Penjual Penitip',
		email: 'akun@contoh.id',
		membership: { organization_id: 'org-' + type, organization_type: type, role: 'OWNER' }
	};
}

function salesReport(group: string) {
	return {
		from: '2026-09-12',
		to: '2026-10-11',
		group_by: group,
		summary: sales,
		rows: [
			{
				key: productId,
				label: group === 'PARTNER' ? 'Warung Bu Tini' : 'Roti Sobek (ROT-1)',
				...sales,
				...(group === 'PRODUCT' && {
					product: { photo_url: '/icons/icon-192.png', href: '/p/roti-anton/ROT-1' }
				})
			}
		]
	};
}

export async function mockDashboard(page: Page, initialType: OrganizationType) {
	let type = initialType;
	await page.context().addCookies([
		{
			name: 'trustip_at',
			value: 'fixture-session',
			url: 'http://localhost',
			httpOnly: true,
			sameSite: 'Lax'
		}
	]);
	await page.addInitScript(() => {
		window.turnstile = {
			render: (_container, options) => {
				queueMicrotask(() => {
					options.callback('test-token');
				});
				return 'captcha';
			},
			reset: () => undefined,
			remove: () => undefined
		};
	});
	await page.route('**/api/v1/**', async (route) => {
		const url = new URL(route.request().url());
		let data: unknown = [];
		if (url.pathname === '/api/v1/me') data = account(type);
		if (url.pathname === '/api/v1/sessions') {
			type = 'CONSIGNOR';
			data = account(type);
		}
		if (url.pathname === '/api/v1/dashboard') data = dashboardBoard;
		if (url.pathname === '/api/v1/insights') data = insights;
		if (url.pathname === '/api/v1/reports/sales')
			data = salesReport(url.searchParams.get('group_by') ?? 'DAY');
		if (url.pathname === '/api/v1/notifications/unread-count') data = { unread_count: 0 };
		await route.fulfill({
			json: {
				code: 'OK',
				message: '',
				data,
				meta: { request_id: 'test', next_cursor: null, has_more: false }
			}
		});
	});
}
