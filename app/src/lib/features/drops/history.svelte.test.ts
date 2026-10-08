import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DropsHarness from '#lib/testing/DropsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const breadId = '0192f0c4-0000-7000-8000-000000000101';

const summary = (number: string, status: string) => ({
	id: `0192f0c4-0000-7000-8000-00000000070${number.slice(-1)}`,
	number,
	partnership_id: partnershipId,
	status,
	occurred_at: '2026-11-10T02:00:00Z',
	recorded_offline: false,
	counterpart: {
		id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c',
		name: 'Roti Pak Anton',
		type: 'CONSIGNOR'
	},
	total_qty: 20,
	line_count: 1
});

function list(data: unknown[], hasMore = false): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r', next_cursor: hasMore ? 'next' : null, has_more: hasMore }
	});
}

function envelope(data: unknown): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } });
}

function serve(routes: Record<string, () => Response>) {
	const requested: string[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		requested.push(key);
		return Promise.resolve(
			routes[key]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return requested;
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Drop history', () => {
	it('FR-DRP-08 lists the partner drops and filters by status', async () => {
		const requested = serve({
			[`GET /api/v1/drops?partnership_id=${partnershipId}`]: () =>
				list([summary('TM-202611-0002', 'RECORDED'), summary('TM-202611-0001', 'CONFIRMED')], true),
			[`GET /api/v1/drops?partnership_id=${partnershipId}&cursor=next`]: () =>
				list([summary('TM-202611-0003', 'CANCELLED')]),
			[`GET /api/v1/drops?partnership_id=${partnershipId}&status=CONFIRMED`]: () =>
				list([summary('TM-202611-0001', 'CONFIRMED')])
		});
		const { container } = await render(DropsHarness, { view: 'history', targetId: partnershipId });

		await expect.element(page.getByText('TM-202611-0002')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Muat lebih banyak' }));
		await expect.element(page.getByText('TM-202611-0003')).toBeVisible();
		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Status' }), 'CONFIRMED');

		await expect.element(page.getByText('TM-202611-0002')).not.toBeInTheDocument();
		expect(requested).toContain(
			`GET /api/v1/drops?partnership_id=${partnershipId}&status=CONFIRMED`
		);
	});

	it('shows a call to record when the history is empty', async () => {
		serve({ [`GET /api/v1/drops?partnership_id=${partnershipId}`]: () => list([]) });
		await render(DropsHarness, { view: 'history', targetId: partnershipId });

		await expect.element(page.getByRole('link', { name: 'Catat titipan' })).toBeVisible();
	});
});

describe('Repeat last drop', () => {
	it('FR-DRP-09 fills the quantities of the last drop that was not cancelled', async () => {
		const previous = summary('TM-202611-0001', 'CONFIRMED');
		serve({
			[`GET /api/v1/partnerships/${partnershipId}/agreement`]: () =>
				envelope({
					version: 1,
					consignor: { id: 'c', name: 'Roti Pak Anton' },
					items: [
						{
							product_id: breadId,
							product_name: 'Roti Manis',
							sku: 'ROT-0001',
							unit: 'pcs',
							pricing_model: 'FIXED_DEPOSIT',
							retail_price: 2000,
							commission_bps: 0,
							deposit_price: 1750,
							is_active: true,
							per_unit: { gross: 2000, store_share: 250, consignor_share: 1750 }
						}
					]
				}),
			[`GET /api/v1/partnerships/${partnershipId}/catalog`]: () => list([]),
			[`GET /api/v1/drops?partnership_id=${partnershipId}`]: () =>
				list([summary('TM-202611-0002', 'CANCELLED'), previous]),
			[`GET /api/v1/drops/${previous.id}`]: () =>
				envelope({ number: previous.number, items: [{ product_id: breadId, qty: 18 }] })
		});
		await render(DropsHarness, { view: 'form', targetId: partnershipId });

		await userEvent.click(page.getByRole('button', { name: 'Ulangi titipan terakhir' }));

		await expect.element(page.getByRole('textbox', { name: 'Jumlah (pcs)' })).toHaveValue('18');
		await expect.element(page.getByText(/Jumlah diisi dari TM-202611-0001/u)).toBeVisible();
	});
});
