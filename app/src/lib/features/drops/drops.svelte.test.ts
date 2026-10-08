import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DropsHarness from '#lib/testing/DropsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';
import { addDays } from '#lib/domain/dates.js';
import { businessDateOf } from '#lib/i18n/format.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const breadId = '0192f0c4-0000-7000-8000-000000000101';
const dropId = '0192f0c4-0000-7000-8000-000000000701';
const today = businessDateOf(new Date());

const agreement = {
	version: 2,
	consignor: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c', name: 'Roti Pak Anton' },
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
};

const drop = {
	id: dropId,
	number: 'TM-202611-0007',
	partnership_id: partnershipId,
	agreement_version: 2,
	status: 'RECORDED',
	occurred_at: '2026-11-10T02:00:00Z',
	received_at: '2026-11-10T02:00:00Z',
	recorded_offline: false,
	store: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d', name: 'Warung Bu Sri' },
	consignor: agreement.consignor,
	recorded_by_name: 'Sri',
	confirmation: null,
	cancellation: null,
	cancel_deadline: '2999-01-01T00:00:00Z',
	items: [
		{
			id: '0192f0c4-0000-7000-8000-000000000801',
			product_id: breadId,
			product_name: 'Roti Manis',
			sku: 'ROT-0001',
			unit: 'pcs',
			batch_code: 'TM-202611-0007/1',
			expiry_date: '2026-11-13',
			qty: 20,
			terms: {
				pricing_model: 'FIXED_DEPOSIT',
				retail_price: 2000,
				commission_bps: 0,
				deposit_price: 1750
			}
		}
	],
	total_qty: 20,
	value_if_sold: { gross: 40000, store_share: 5000, consignor_share: 35000 },
	seal: { is_valid: true, events: [] }
};

function envelope(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

interface Sent {
	body: unknown;
	key: string | null;
}

function serve(routes: Record<string, () => Response>) {
	const sent: Sent[] = [];
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		if (init.method === 'POST') {
			const headers = new Headers(init.headers);
			sent.push({
				body: typeof init.body === 'string' ? JSON.parse(init.body) : init.body,
				key: headers.get('idempotency-key')
			});
		}
		return Promise.resolve(
			routes[key]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return sent;
}

const formRoutes = (post: () => Response) => ({
	[`GET /api/v1/partnerships/${partnershipId}/agreement`]: () => envelope(agreement),
	[`GET /api/v1/partnerships/${partnershipId}/catalog`]: () =>
		Response.json({
			code: 'OK',
			message: '',
			data: [{ product_id: breadId, shelf_life_days: 3 }],
			meta: { request_id: 'r', next_cursor: null, has_more: false }
		}),
	'POST /api/v1/drops': post
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Drop form', () => {
	it('FR-DRP-02 FR-DRP-04 fills the expiry and previews the value before saving', async () => {
		const sent = serve(formRoutes(() => envelope(drop, 201)));
		const ondone = vi.fn();
		const { container } = await render(DropsHarness, {
			view: 'form',
			targetId: partnershipId,
			ondone
		});

		await expect.element(page.getByLabelText('Kedaluwarsa')).toHaveValue(addDays(today, 3));
		await userEvent.fill(page.getByRole('textbox', { name: 'Jumlah (pcs)' }), '20');
		await userEvent.tab();
		await expect.element(page.getByText('Rp40.000')).toBeVisible();
		await expect.element(page.getByText('Rp35.000')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Simpan Titipan Masuk' }));

		await vi.waitFor(() => {
			expect(ondone).toHaveBeenCalledWith(expect.objectContaining({ number: 'TM-202611-0007' }));
		});
		expect(sent[0]?.body).toEqual({
			partnership_id: partnershipId,
			agreement_version: 2,
			occurred_at: null,
			items: [{ product_id: breadId, qty: 20, expiry_date: addDays(today, 3) }]
		});
		expect(sent[0]?.key).toMatch(/^[0-9a-f-]{36}$/u);
	});

	it('BR-DRP-07 retries a failed save with the same idempotency key', async () => {
		let attempts = 0;
		const sent = serve(
			formRoutes(() => {
				attempts += 1;
				return attempts === 1
					? Response.json(
							{ code: 'SERVICE_UNAVAILABLE', message: 'Sibuk.', meta: { request_id: 'r' } },
							{ status: 503, headers: { 'retry-after': '1' } }
						)
					: envelope(drop, 201);
			})
		);
		const ondone = vi.fn();
		await render(DropsHarness, { view: 'form', targetId: partnershipId, ondone });

		await userEvent.fill(page.getByRole('textbox', { name: 'Jumlah (pcs)' }), '20');
		await userEvent.tab();
		await userEvent.click(page.getByRole('button', { name: 'Simpan Titipan Masuk' }));
		await expect.element(page.getByText('Sibuk.')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Simpan Titipan Masuk' }));

		await vi.waitFor(() => {
			expect(ondone).toHaveBeenCalled();
		});
		expect(sent).toHaveLength(2);
		expect(sent[1]?.key).toBe(sent[0]?.key);
	});

	it('asks for at least one quantity before sending anything', async () => {
		const sent = serve(formRoutes(() => envelope(drop, 201)));
		await render(DropsHarness, { view: 'form', targetId: partnershipId });

		await userEvent.click(page.getByRole('button', { name: 'Simpan Titipan Masuk' }));

		await expect.element(page.getByText('Isi jumlah minimal satu produk.')).toBeVisible();
		expect(sent).toHaveLength(0);
	});

	it('BR-DRP-01 shows the server field error under the product it belongs to', async () => {
		serve(
			formRoutes(() =>
				Response.json(
					{
						code: 'VALIDATION_FAILED',
						message: 'Periksa kembali isian Anda.',
						errors: [
							{
								field: 'items[0].product_id',
								code: 'PRODUCT_NOT_IN_AGREEMENT',
								message: 'Produk ini belum ada di kesepakatan. Usulkan dulu lewat Ajuan.'
							}
						],
						meta: { request_id: 'r' }
					},
					{ status: 422 }
				)
			)
		);
		await render(DropsHarness, { view: 'form', targetId: partnershipId });

		await userEvent.fill(page.getByRole('textbox', { name: 'Jumlah (pcs)' }), '5');
		await userEvent.tab();
		await userEvent.click(page.getByRole('button', { name: 'Simpan Titipan Masuk' }));

		await expect
			.element(page.getByText('Produk ini belum ada di kesepakatan. Usulkan dulu lewat Ajuan.'))
			.toBeVisible();
		await expect.element(page.getByRole('textbox', { name: 'Jumlah (pcs)' })).toHaveFocus();
	});
});

describe('Drop detail', () => {
	it('FR-DRP-05 shows the number, the batch, the totals, and the seal', async () => {
		serve({ [`GET /api/v1/drops/${dropId}`]: () => envelope(drop) });
		const { container } = await render(DropsHarness, { view: 'detail', targetId: dropId });

		await expect.element(page.getByText('TM-202611-0007', { exact: true })).toBeVisible();
		await expect.element(page.getByText('Menunggu konfirmasi')).toBeVisible();
		await expect.element(page.getByText('Segel sah')).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Drop decisions', () => {
	const consignorId = drop.consignor.id;
	const storeId = drop.store.id;

	it('FR-DRP-06 lets the consignor confirm after a dialog that names the effect', async () => {
		serve({
			[`GET /api/v1/drops/${dropId}`]: () => envelope(drop),
			[`POST /api/v1/drops/${dropId}/confirm`]: () =>
				envelope({
					...drop,
					status: 'CONFIRMED',
					confirmation: { by_name: 'Anton', at: '2026-11-10T03:00:00Z', is_automatic: false }
				})
		});
		const { container } = await render(DropsHarness, {
			view: 'detail',
			targetId: dropId,
			viewerOrganizationId: consignorId
		});

		await userEvent.click(page.getByRole('button', { name: 'Sesuai' }));
		await expect.element(page.getByText('Nyatakan titipan ini sesuai?')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Sesuai' }));

		await expect.element(page.getByText('Titipan dinyatakan sesuai.')).toBeVisible();
		await expect.element(page.getByText(/Dinyatakan sesuai oleh Anton/u)).toBeVisible();
		await expect
			.element(page.getByRole('button', { name: 'Batalkan titipan' }))
			.not.toBeInTheDocument();
	});

	it('FR-DRP-10 shows the store a cancel button only before the deadline', async () => {
		serve({
			[`GET /api/v1/drops/${dropId}`]: () =>
				envelope({ ...drop, cancel_deadline: '2000-01-01T00:00:00Z' })
		});
		await render(DropsHarness, { view: 'detail', targetId: dropId, viewerOrganizationId: storeId });

		await expect.element(page.getByText('Menunggu konfirmasi')).toBeVisible();
		await expect
			.element(page.getByRole('button', { name: 'Batalkan titipan' }))
			.not.toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Sesuai' })).not.toBeInTheDocument();
	});

	it('FR-DRP-10 cancels within the window and shows who cancelled', async () => {
		serve({
			[`GET /api/v1/drops/${dropId}`]: () => envelope(drop),
			[`POST /api/v1/drops/${dropId}/cancel`]: () =>
				envelope({
					...drop,
					status: 'CANCELLED',
					cancellation: { by_name: 'Sri', at: '2026-11-10T02:10:00Z' }
				})
		});
		await render(DropsHarness, { view: 'detail', targetId: dropId, viewerOrganizationId: storeId });

		await userEvent.click(page.getByRole('button', { name: 'Batalkan titipan' }));
		await userEvent.click(
			page.getByRole('dialog').getByRole('button', { name: 'Batalkan titipan' })
		);

		await expect.element(page.getByText('Dibatalkan', { exact: true })).toBeVisible();
		await expect.element(page.getByText(/Dibatalkan oleh Sri/u)).toBeVisible();
	});
});
