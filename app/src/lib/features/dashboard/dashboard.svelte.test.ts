import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DashboardHarness from '#lib/testing/DashboardHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const storeId = '0192f0c4-0000-7000-8000-000000000401';
const partnershipId = '0192f0c4-0000-7000-8000-000000000201';
const settlementId = '0192f0c4-0000-7000-8000-000000000901';
const tally = (count: number, amount: number) => ({ count, amount });
const split = { gross: 36000, store_share: 4500, consignor_share: 31500 };
const figures = { sold_qty: 18, returned_qty: 2, received_qty: 20, returned_bps: 1000, split };

const dashboard = {
	figures: {
		active_partners: 2,
		stock_units: 42,
		unpaid: tally(2, 45500),
		overdue: tally(1, 31500),
		due_soon: tally(0, 0)
	},
	tasks: {
		drops_to_confirm: 1,
		counts_to_confirm: 0,
		payments_to_review: 2,
		payments_rejected: 0,
		proposals_awaiting: 0
	},
	quick_count: []
};

const overdueInsight = {
	kind: 'SETTLEMENT_OVERDUE',
	level: 'HIGH',
	action: 'REMIND_PAYMENT',
	subject: {
		partnership_id: partnershipId,
		counterpart_name: 'Warung Bu Tini',
		product_id: null,
		product_name: '',
		batch_id: null,
		batch_code: '',
		settlement_id: settlementId,
		settlement_number: 'ST-202611-0001',
		organization_id: null,
		organization_name: ''
	},
	units: 0,
	days: 3,
	amount: 31500,
	score: 0
};

function salesReport(groupBy: string) {
	const label = groupBy === 'PARTNER' ? 'Warung Bu Tini' : 'Roti Sobek (ROT-1)';
	return {
		from: '2026-10-22',
		to: '2026-11-20',
		group_by: groupBy,
		summary: figures,
		rows: [{ key: storeId, label, ...figures }]
	};
}

function dataFor(path: string): unknown {
	const url = new URL(path, 'https://trustip.test');
	if (url.pathname === '/api/v1/dashboard') {
		return dashboard;
	}
	if (url.pathname === '/api/v1/insights') {
		return [overdueInsight];
	}
	return salesReport(url.searchParams.get('group_by') ?? 'PARTNER');
}

function serve(status = 200) {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string) => {
		requested.push(path);
		const meta = { request_id: 'r', next_cursor: null, has_more: false };
		return Promise.resolve(
			status === 200
				? Response.json({ code: 'OK', message: '', data: dataFor(path), meta })
				: Response.json({ code: 'INTERNAL', message: 'Gangguan.', meta }, { status })
		);
	});
	return requested;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Consignor dashboard', () => {
	it('DSH shows key figures, tasks, insights, and rankings for the owner', async () => {
		serve();
		const { container } = await render(DashboardHarness, {});

		await expect.element(page.getByText('Unit di semua toko')).toBeVisible();
		await expect.element(page.getByText('42', { exact: true })).toBeVisible();
		await expect.element(page.getByText('Rp31.500').first()).toBeVisible();
		await expect.element(page.getByText('1 lewat tempo, Rp31.500')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: '1 Titipan Masuk menunggu konfirmasi' }))
			.toHaveAttribute('href', '/app/titipan');
		await expect.element(page.getByText('Setoran ST-202611-0001 lewat tempo')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Ingatkan bayar' }))
			.toHaveAttribute('href', `/app/setoran/${settlementId}`);
		await expect.element(page.getByText('1. Warung Bu Tini')).toBeVisible();
		await expect.element(page.getByText('chart all')).toBeVisible();
		await expectAccessible(container);
	});

	it('DSH narrows the daily chart to one store', async () => {
		serve();
		await render(DashboardHarness, {});

		await userEvent.selectOptions(
			page.getByRole('combobox', { name: 'Saring grafik' }),
			`partner:${storeId}`
		);

		await expect.element(page.getByText(`chart ${storeId}`)).toBeVisible();
	});

	it('BR-RPT-01 keeps sales figures away from staff', async () => {
		const requested = serve();
		await render(DashboardHarness, { isOwner: false });

		await expect.element(page.getByText('Unit di semua toko')).toBeVisible();
		await expect
			.element(page.getByText('Angka penjualan hanya tampil untuk pemilik usaha.'))
			.toBeVisible();
		expect(requested.some((path) => path.includes('/reports/'))).toBe(false);
	});

	it('shows an alert when the dashboard cannot load', async () => {
		serve(500);
		await render(DashboardHarness, {});

		await expect.element(page.getByText('Dasbor tidak bisa dimuat.')).toBeVisible();
	});
});
