import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import StockHarness from '#lib/testing/StockHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const breadId = '0192f0c4-0000-7000-8000-000000000101';
const dropId = '0192f0c4-0000-7000-8000-000000000701';

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: null, has_more: false }
	});
}

function serve(routes: Record<string, () => Response>) {
	vi.stubGlobal('fetch', (path: string, init: RequestInit) =>
		Promise.resolve(
			routes[`${init.method ?? 'GET'} ${path}`]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		)
	);
}

const holding = (name: string, store: string, balance: number) => ({
	partnership_id: `0192f0c4-0000-7000-8000-00000000050${String(balance % 10)}`,
	counterpart: {
		id: `0192f0c4-0000-7000-8000-00000000040${String(balance % 10)}`,
		name: store,
		type: 'STORE'
	},
	product_id: breadId,
	product_name: name,
	sku: 'ROT-0001',
	unit: 'pcs',
	balance,
	flag: balance > 10 ? 'SAFE' : 'NEAR_EXPIRY',
	nearest_expiry: '2026-11-13'
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Stock', () => {
	it('FR-STK-02 adds one product across every store', async () => {
		serve({
			'GET /api/v1/stock': () =>
				list([holding('Roti Manis', 'Warung Bu Sri', 12), holding('Roti Manis', 'Toko Maju', 5)])
		});
		const { container } = await render(StockHarness, { view: 'overview' });

		await expect.element(page.getByText('17 pcs')).toBeVisible();
		await expect.element(page.getByRole('link', { name: 'Toko Maju' })).toBeVisible();
		await expect.element(page.getByText('Mendekati kedaluwarsa')).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-STK-01 FR-STK-05 shows every batch with its age and flag', async () => {
		serve({
			[`GET /api/v1/partnerships/${partnershipId}/stock`]: () =>
				list([
					{
						product_id: breadId,
						product_name: 'Roti Manis',
						sku: 'ROT-0001',
						unit: 'pcs',
						balance: 17,
						flag: 'EXPIRED',
						nearest_expiry: '2026-11-09',
						batches: [
							{
								id: '0192f0c4-0000-7000-8000-000000000801',
								batch_code: 'TM-202611-0001/1',
								expiry_date: '2026-11-09',
								received_at: '2026-11-01T02:00:00Z',
								balance: 5,
								age_days: 9,
								flag: 'EXPIRED'
							}
						]
					}
				]),
			[`GET /api/v1/partnerships/${partnershipId}/drop-suggestions`]: () =>
				list([{ product_id: breadId, suggested_qty: 19, has_enough_data: true }])
		});
		const { container } = await render(StockHarness, { view: 'partner', partnershipId });

		await expect.element(page.getByText('17 pcs')).toBeVisible();
		await expect.element(page.getByText('Saran titip: 19 pcs.')).toBeVisible();
		await expect.element(page.getByText('9 hari').first()).toBeInTheDocument();
		await expect
			.element(page.getByRole('link', { name: 'Kartu stok Roti Manis' }))
			.toHaveAttribute('href', `/app/mitra/${partnershipId}/stok/${breadId}`);
		await expectAccessible(container);
	});

	it('FR-STK-04 links each movement on the stock card to its document', async () => {
		serve({
			[`GET /api/v1/partnerships/${partnershipId}/stock/${breadId}/card`]: () =>
				Response.json({
					code: 'OK',
					message: '',
					data: {
						product_id: breadId,
						product_name: 'Roti Manis',
						sku: 'ROT-0001',
						unit: 'pcs',
						entries: [
							{
								id: '0192f0c4-0000-7000-8000-000000000901',
								type: 'RECEIVE',
								qty: 20,
								occurred_at: '2026-11-10T02:00:00Z',
								batch_code: 'TM-202611-0001/1',
								source_type: 'DROP',
								source_id: dropId,
								source_number: 'TM-202611-0001',
								balance_after: 20
							}
						],
						next_cursor: null,
						has_more: false
					},
					meta: { request_id: 'r' }
				})
		});
		await render(StockHarness, { view: 'card', partnershipId, productId: breadId });

		await expect.element(page.getByText('+20')).toBeVisible();
		await expect.element(page.getByText('Saldo 20')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'TM-202611-0001' }))
			.toHaveAttribute('href', `/app/titipan/${dropId}`);
	});

	it('shows a call to record when there is no stock yet', async () => {
		serve({ 'GET /api/v1/stock': () => list([]) });
		await render(StockHarness, { view: 'overview' });

		await userEvent.tab();
		await expect.element(page.getByRole('link', { name: 'Catat titipan' })).toBeVisible();
	});
});
