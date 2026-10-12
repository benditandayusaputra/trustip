import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DashboardHarness from '#lib/testing/DashboardHarness.svelte';
import type { Dashboard } from '#lib/api/insights.js';
import { expectAccessible } from '#lib/testing/accessibility.js';

const storeId = '0192f0c4-0000-7000-8000-000000000401';
const partnershipId = '0192f0c4-0000-7000-8000-000000000201';
const settlementId = '0192f0c4-0000-7000-8000-000000000901';
const tally = (count: number, amount: number) => ({ count, amount });
const split = { gross: 36000, store_share: 4500, consignor_share: 31500 };
const figures = { sold_qty: 18, returned_qty: 2, received_qty: 20, returned_bps: 1000, split };

const dashboard: Dashboard = {
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
		proposals_awaiting: 0,
		needs_expiring: 0
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
		rows: [
			{
				key: storeId,
				label,
				...figures,
				...(groupBy === 'PRODUCT' && {
					product: { photo_url: '/product-photo.png', href: '/p/roti-anton/ROT-1' }
				})
			}
		]
	};
}

let served = dashboard;
let servedInsights = [overdueInsight];

const needId = '0192f0c4-0000-7000-8000-000000000a01';
const candidate = (name: string) => ({
	id: '0192f0c4-0000-7000-8000-000000000b01',
	name,
	is_partner: false,
	is_self: false,
	match: { score: 82, label: 'VERY_GOOD', reasons: [] }
});
const directoryData: Record<string, unknown[]> = {
	'/api/v1/directory/needs': [
		{
			id: needId,
			title: 'Butuh roti sarapan',
			store: { id: storeId, name: 'Kantin Sehat' },
			distance_m: 1200,
			is_own: false
		}
	],
	'/api/v1/directory/stores': [candidate('Kantin Sehat')],
	'/api/v1/directory/consignors': [candidate('Keripik Bu Ani')],
	'/api/v1/needs': [{ id: needId, title: 'Butuh roti sarapan', response_count: 2 }]
};

function dataFor(path: string): unknown {
	const url = new URL(path, 'https://trustip.test');
	const listed = directoryData[url.pathname];
	if (listed) {
		return listed;
	}
	if (url.pathname === '/api/v1/dashboard') {
		return served;
	}
	if (url.pathname === '/api/v1/insights') {
		return servedInsights;
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
	served = dashboard;
	servedInsights = [overdueInsight];
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
		await expect
			.element(page.getByRole('link', { name: /Butuh roti sarapan/u }))
			.toHaveAttribute('href', `/app/ajuan/baru?sasaran=${storeId}&kebutuhan=${needId}`);
		await expect
			.element(page.getByRole('link', { name: /Kantin Sehat Skor kecocokan 82/u }))
			.toHaveAttribute('href', '/cari/toko?q=Kantin+Sehat');
		await expectAccessible(container);
	});

	it('DSH narrows the daily chart to one store', async () => {
		serve();
		await render(DashboardHarness, {});

		await userEvent.click(page.getByRole('combobox', { name: 'Saring grafik' }));
		await userEvent.selectOptions(
			page.getByRole('combobox', { name: 'Saring grafik' }),
			`partner:${storeId}`
		);

		await expect.element(page.getByText(`chart ${storeId}`)).toBeVisible();
		await expect.element(page.getByRole('combobox', { name: 'Saring grafik' })).toHaveFocus();
	});

	it('keeps consignor suggestions compact without hiding their actions', async () => {
		servedInsights = Array.from({ length: 5 }, (_, index) => ({
			...overdueInsight,
			subject: { ...overdueInsight.subject, settlement_number: `ST-${String(index + 1)}` }
		}));
		serve();
		const { container } = await render(DashboardHarness, {});
		await expect.element(page.getByText('Setoran ST-3 lewat tempo')).toBeVisible();
		await expect.element(page.getByText('Setoran ST-4 lewat tempo')).not.toBeInTheDocument();
		const expand = page.getByRole('button', { name: 'Lihat semua saran (5)' });
		await expect.element(expand).toHaveAttribute('aria-expanded', 'false');
		await userEvent.click(expand);
		await expect.element(page.getByText('Setoran ST-5 lewat tempo')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Ingatkan bayar' }).last())
			.toHaveAttribute('href', `/app/setoran/${settlementId}`);
		await userEvent.click(page.getByRole('button', { name: 'Tampilkan lebih sedikit saran' }));
		await expect.element(page.getByText('Setoran ST-4 lewat tempo')).not.toBeInTheDocument();
		await expectAccessible(container);
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

describe('Store dashboard', () => {
	it('keeps many suggestions compact and lets the user reveal the rest', async () => {
		servedInsights = Array.from({ length: 7 }, (_, index) => ({
			...overdueInsight,
			subject: { ...overdueInsight.subject, settlement_number: `ST-${String(index + 1)}` }
		}));
		serve();
		const { container } = await render(DashboardHarness, { side: 'STORE' });
		await expect.element(page.getByText('Setoran ST-3 lewat tempo')).toBeVisible();
		await expect.element(page.getByText('Setoran ST-4 lewat tempo')).not.toBeInTheDocument();
		const expand = page.getByRole('button', { name: 'Lihat semua saran (7)' });
		await expect.element(expand).toHaveAttribute('aria-expanded', 'false');
		await userEvent.click(expand);
		await expect.element(page.getByText('Setoran ST-7 lewat tempo')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Tampilkan lebih sedikit saran' }));
		await expect.element(page.getByText('Setoran ST-4 lewat tempo')).not.toBeInTheDocument();
		await expectAccessible(container);
	});
	const storeBoard: Dashboard = {
		...dashboard,
		figures: { ...dashboard.figures, due_soon: tally(1, 14000) },
		tasks: { ...dashboard.tasks, drops_to_confirm: 0, payments_rejected: 1 },
		quick_count: [
			{ partnership_id: partnershipId, counterpart_name: 'Roti Pak Anton', stock_units: 12 }
		]
	};

	it('FR-CNT-11 lists partners not counted today with a Hitung button', async () => {
		served = storeBoard;
		serve();
		const { container } = await render(DashboardHarness, { side: 'STORE' });

		await expect.element(page.getByText('Roti Pak Anton')).toBeVisible();
		await expect.element(page.getByText('12 unit tercatat')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Hitung Roti Pak Anton' }))
			.toHaveAttribute('href', `/app/hitung/baru?mitra=${partnershipId}`);
		await expect.element(page.getByText('1 jatuh tempo 3 hari lagi, Rp14.000')).toBeVisible();
		await expect.element(page.getByText('Rp4.500', { exact: true })).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: '2 Setoran belum dibayar' }))
			.toHaveAttribute('href', '/app/setoran');
		await expect.element(page.getByRole('link', { name: '1 pembayaran ditolak' })).toBeVisible();
		await expect.element(page.getByText('1. Roti Sobek (ROT-1)')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: /Roti Sobek.*Lihat detail/u }))
			.toHaveAttribute('href', '/p/roti-anton/ROT-1');
		expect(container.querySelector('img')?.getAttribute('src')).toBe('/product-photo.png');
		await expect
			.element(page.getByRole('link', { name: /Butuh roti sarapan 2 tanggapan/u }))
			.toHaveAttribute('href', `/app/kebutuhan/${needId}`);
		await expect.element(page.getByRole('link', { name: /Keripik Bu Ani/u })).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-CNT-11 says everyone is counted when no partner is waiting', async () => {
		serve();
		await render(DashboardHarness, { side: 'STORE' });

		await expect
			.element(page.getByText('Semua mitra berstok sudah dihitung hari ini.'))
			.toBeVisible();
	});
});
