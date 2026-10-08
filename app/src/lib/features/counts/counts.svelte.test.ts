import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { aQueuedDrop, emptyOutbox, storeCashier } from '#lib/testing/outbox.js';
import { outboxStore } from '#lib/offline/store.js';
import { render } from 'vitest-browser-svelte';
import CountsHarness from '#lib/testing/CountsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const partnershipId = '0192f0c4-0000-7000-8000-000000000501';
const breadId = '0192f0c4-0000-7000-8000-000000000101';
const countId = '0192f0c4-0000-7000-8000-000000000901';

const sheet = {
	partnership_id: partnershipId,
	is_active: true,
	agreement_version: 1,
	store: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d', name: 'Warung Bu Sri' },
	consignor: { id: '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c', name: 'Roti Pak Anton' },
	last_document_at: null,
	products: [
		{
			product_id: breadId,
			product_name: 'Roti Manis',
			sku: 'ROT-0001',
			unit: 'pcs',
			recorded_qty: 20,
			batches: [
				{
					id: '0192f0c4-0000-7000-8000-000000000801',
					expiry_date: '2026-11-13',
					received_at: '2026-11-10T02:00:00Z',
					balance: 20,
					terms: {
						pricing_model: 'FIXED_DEPOSIT',
						retail_price: 2000,
						commission_bps: 0,
						deposit_price: 1750
					}
				}
			]
		}
	]
};

const recap = {
	id: countId,
	number: 'HS-202611-0001',
	partnership_id: partnershipId,
	status: 'RECORDED',
	occasion: 'STORE_CLOSING',
	note: '',
	counted_at: '2026-11-10T12:00:00Z',
	client_recorded_at: '2026-11-10T12:00:00Z',
	received_at: '2026-11-10T12:00:00Z',
	recorded_offline: false,
	store: sheet.store,
	consignor: sheet.consignor,
	recorded_by_name: 'Sri',
	confirmation: null,
	lines: [
		{
			id: '0192f0c4-0000-7000-8000-000000000902',
			product_id: breadId,
			product_name: 'Roti Manis',
			sku: 'ROT-0001',
			unit: 'pcs',
			recorded_qty: 20,
			counted_qty: 2,
			sold_qty: 18,
			returned_qty: 2,
			kept_qty: 0,
			return_reason: 'UNSOLD',
			split: { gross: 36000, store_share: 4500, consignor_share: 31500 }
		}
	],
	sold_qty: 18,
	returned_qty: 2,
	split: { gross: 36000, store_share: 4500, consignor_share: 31500 },
	seal: { is_valid: true, events: [] }
};

function envelope(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

function serve(
	post: () => Response = () => envelope(recap, 201),
	extra: Record<string, () => Response> = {}
) {
	const sent: unknown[] = [];
	const requested: string[] = [];
	const routes: Record<string, () => Response> = {
		[`GET /api/v1/partnerships/${partnershipId}/count-sheet`]: () => envelope(sheet),
		[`GET /api/v1/counts/${countId}`]: () => envelope(recap),
		'POST /api/v1/counts': post,
		'GET /api/v1/me': () => envelope(storeCashier),
		...extra
	};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		requested.push(key);
		if (typeof init.body === 'string') {
			sent.push(JSON.parse(init.body));
		}
		return Promise.resolve(
			routes[key]?.() ??
				Response.json(
					{ code: 'NOT_FOUND', message: 'x', meta: { request_id: 'r' } },
					{ status: 404 }
				)
		);
	});
	return { sent, requested };
}

async function fill(label: string, value: string) {
	await userEvent.fill(page.getByRole('textbox', { name: label }), value);
	await userEvent.tab();
}

afterEach(async () => {
	await emptyOutbox();
	vi.unstubAllGlobals();
});

describe('Count form', () => {
	it('FR-CNT-04 previews the section 12 example while typing and saves it', async () => {
		const { sent } = serve();
		const ondone = vi.fn();
		const { container } = await render(CountsHarness, {
			view: 'form',
			targetId: partnershipId,
			ondone
		});

		await fill('Sisa fisik (pcs)', '2');
		await fill('Dikembalikan (pcs)', '2');
		await userEvent.selectOptions(
			page.getByRole('combobox', { name: 'Alasan pengembalian' }),
			'UNSOLD'
		);
		await expect.element(page.getByText('Rp36.000')).toBeVisible();
		await expect.element(page.getByText('Rp4.500')).toBeVisible();
		await expect.element(page.getByText('Rp31.500')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Simpan Hitung Sisa' }));

		await vi.waitFor(() => {
			expect(ondone).toHaveBeenCalledWith(countId);
		});
		expect(sent[0]).toMatchObject({
			partnership_id: partnershipId,
			occasion: 'STORE_CLOSING',
			lines: [
				{
					product_id: breadId,
					recorded_qty: 20,
					counted_qty: 2,
					returned_qty: 2,
					return_reason: 'UNSOLD'
				}
			]
		});
	});

	it('R-OFF-04 counts a drop still waiting in the queue as recorded stock', async () => {
		serve();
		await outboxStore.put(
			aQueuedDrop({
				partnershipId,
				lines: [
					{
						productId: breadId,
						name: 'Roti Manis',
						unit: 'pcs',
						qty: 5,
						expiryDate: '',
						terms: {
							model: 'FIXED_DEPOSIT',
							retailPrice: 2000,
							commissionBps: 0,
							depositPrice: 1750
						}
					}
				]
			})
		);

		await render(CountsHarness, { view: 'form', targetId: partnershipId });

		await expect.element(page.getByText('Stok tercatat 25 pcs')).toBeVisible();
	});

	it('FR-CNT-07 fills sold units and works out what is left', async () => {
		serve();
		await render(CountsHarness, { view: 'form', targetId: partnershipId });

		await userEvent.selectOptions(page.getByRole('combobox', { name: 'Cara isi' }), 'sold');
		await fill('Terjual (pcs)', '8');

		await expect.element(page.getByText('Terjual 8, tetap dititip 12')).toBeVisible();
	});

	it('FR-CNT-06 marks everything sold with one button', async () => {
		serve();
		await render(CountsHarness, { view: 'form', targetId: partnershipId });

		await userEvent.click(page.getByRole('button', { name: 'Habis semua' }));

		await expect.element(page.getByText('Terjual 20, tetap dititip 0')).toBeVisible();
		await expect.element(page.getByText('Rp40.000')).toBeVisible();
	});

	it('BR-CNT-02 asks for a reason before sending a return', async () => {
		const { sent } = serve();
		await render(CountsHarness, { view: 'form', targetId: partnershipId });

		await fill('Sisa fisik (pcs)', '5');
		await fill('Dikembalikan (pcs)', '1');
		await userEvent.click(page.getByRole('button', { name: 'Simpan Hitung Sisa' }));

		await expect.element(page.getByText('Pilih alasan pengembalian.')).toBeVisible();
		expect(sent).toHaveLength(0);
	});

	it('reloads the sheet when the stock changed on the server', async () => {
		const { requested } = serve(() =>
			Response.json(
				{ code: 'STOCK_CHANGED', message: 'Stok baru saja berubah.', meta: { request_id: 'r' } },
				{ status: 409 }
			)
		);
		await render(CountsHarness, { view: 'form', targetId: partnershipId });

		await fill('Sisa fisik (pcs)', '5');
		await userEvent.click(page.getByRole('button', { name: 'Simpan Hitung Sisa' }));

		await expect.element(page.getByText(/Lembar hitung dimuat ulang/u)).toBeVisible();
		expect(requested.filter((key) => key.endsWith('/count-sheet'))).toHaveLength(2);
	});
});

describe('Count recap', () => {
	it('FR-CNT-05 shows gross, store share, and what to pay, with the seal', async () => {
		serve();
		const { container } = await render(CountsHarness, { view: 'recap', targetId: countId });

		await expect.element(page.getByText('HS-202611-0001')).toBeVisible();
		await expect.element(page.getByText('Rp31.500').first()).toBeInTheDocument();
		await expect.element(page.getByText('Segel sah')).toBeVisible();
		await expect.element(page.getByRole('link', { name: 'Catat titipan baru' })).toBeVisible();
		await expectAccessible(container);
	});
});

describe('Count confirmation and history', () => {
	it('FR-CNT-08 lets the consignor confirm the recap', async () => {
		serve(undefined, {
			[`POST /api/v1/counts/${countId}/confirm`]: () =>
				envelope({
					...recap,
					status: 'CONFIRMED',
					confirmation: { by_name: 'Anton', at: '2026-11-10T13:00:00Z', is_automatic: false }
				})
		});
		await render(CountsHarness, {
			view: 'recap',
			targetId: countId,
			viewerOrganizationId: sheet.consignor.id
		});

		await userEvent.click(page.getByRole('button', { name: 'Sesuai' }));
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Sesuai' }));

		await expect.element(page.getByText('Hitung Sisa dinyatakan sesuai.')).toBeVisible();
		await expect.element(page.getByText(/Dinyatakan sesuai oleh Anton/u)).toBeVisible();
	});

	it('FR-CNT-08 shows no confirm button to the store', async () => {
		serve();
		await render(CountsHarness, { view: 'recap', targetId: countId });

		await expect.element(page.getByText('HS-202611-0001')).toBeVisible();
		await expect.element(page.getByRole('button', { name: 'Sesuai' })).not.toBeInTheDocument();
	});

	it('FR-CNT-10 lists the partner counts with what the store owes', async () => {
		const summary = {
			id: countId,
			number: 'HS-202611-0001',
			partnership_id: partnershipId,
			status: 'CONFIRMED',
			counted_at: '2026-11-10T12:00:00Z',
			recorded_offline: false,
			counterpart: { ...sheet.consignor, type: 'CONSIGNOR' },
			sold_qty: 18,
			returned_qty: 2,
			consignor_share: 31500
		};
		serve(undefined, {
			[`GET /api/v1/counts?partnership_id=${partnershipId}`]: () =>
				Response.json({
					code: 'OK',
					message: '',
					data: [summary],
					meta: { request_id: 'r', next_cursor: null, has_more: false }
				})
		});
		const { container } = await render(CountsHarness, { view: 'history', targetId: partnershipId });

		await expect
			.element(page.getByText('18 terjual, 2 dikembalikan', { exact: false }))
			.toBeVisible();
		await expect.element(page.getByText(/Rp31\.500/u)).toBeVisible();
		await expectAccessible(container);
	});
});
